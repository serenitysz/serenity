package linter

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unsafe"

	"github.com/serenitysz/serenity/internal/rules"
	"github.com/serenitysz/serenity/internal/version"
)

const (
	cacheSchemaVersion = "v5"
	cacheMagic         = "SRC5"
	cacheHashSize      = sha256.Size
	cacheSampleChunk   = 64
	cacheSampleWindows = 4
)

type cacheStore struct {
	enabled    bool
	dir        string
	configHash string
	mutating   bool
}

type packageInput struct {
	Path               string
	NormalizedPath     string
	Size               int64
	ModTimeUnixNano    int64
	ChangeTimeUnixNano int64
	Device             uint64
	Inode              uint64
	FastPathSupported  bool
	Src                []byte
	Hash               [cacheHashSize]byte
	SampleHash         [cacheHashSize]byte
}

type cacheHeader struct {
	IssueCount   int
	FixableCount int
	IssueStart   int
}

type cacheDecoder struct {
	data []byte
	off  int
}

func newCacheStore(cfg *rules.LinterOptions, mutating, unsafe bool) *cacheStore {
	if cfg == nil || unsafe {
		return &cacheStore{}
	}

	perf := cfg.Performance
	if perf == nil || !perf.Use || perf.Caching == nil || !*perf.Caching {
		return &cacheStore{}
	}

	dir, err := resolveCacheDir()
	if err != nil || dir == "" {
		return &cacheStore{}
	}

	return &cacheStore{
		enabled:    true,
		dir:        dir,
		configHash: cacheConfigHash(cfg),
		mutating:   mutating,
	}
}

func resolveCacheDir() (string, error) {
	if dir := os.Getenv("SERENITY_CACHE_DIR"); dir != "" {
		return dir, nil
	}

	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "serenity", "lint-cache"), nil
}

func cacheConfigHash(cfg *rules.LinterOptions) string {
	clone := *cfg
	clone.Performance = nil

	payload := struct {
		Schema  string
		Version string
		Config  rules.LinterOptions
	}{
		Schema:  cacheSchemaVersion,
		Version: version.Version,
		Config:  clone,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return cacheSchemaVersion + ":" + version.Version
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

func (c *cacheStore) enabledForRun() bool {
	return c != nil && c.enabled && c.dir != ""
}

func (c *cacheStore) entryPath(inputs []packageInput) (string, error) {
	key := sha256.New()

	_, _ = key.Write([]byte(c.configHash))
	_, _ = key.Write([]byte{0})

	for _, input := range inputs {
		_, _ = key.Write([]byte(input.NormalizedPath))
		_, _ = key.Write([]byte{0})
	}

	sum := hex.EncodeToString(key.Sum(nil))
	dir := filepath.Join(c.dir, sum[:2])

	return filepath.Join(dir, sum+".bin"), nil
}

func (c *cacheStore) load(inputs []packageInput, limit int) ([]rules.Issue, bool) {
	data, header, ok := c.loadValidated(inputs)
	if !ok {
		return nil, false
	}

	if !c.canReuse(header) {
		return nil, false
	}

	return decodeCachedIssues(data, header, inputs, limit)
}

func (c *cacheStore) loadRaw(inputs []packageInput) (*cachedBatch, bool) {
	data, header, ok := c.loadValidated(inputs)
	if !ok {
		return nil, false
	}

	if !c.canReuse(header) {
		return nil, false
	}

	return &cachedBatch{
		data:       data,
		issueCount: header.IssueCount,
		issueStart: header.IssueStart,
		inputs:     inputs,
	}, true
}

func (c *cacheStore) canReuse(header cacheHeader) bool {
	return !c.mutating || header.FixableCount == 0
}

func (c *cacheStore) loadValidated(inputs []packageInput) ([]byte, cacheHeader, bool) {
	if !c.enabledForRun() || len(inputs) == 0 {
		return nil, cacheHeader{}, false
	}

	path, err := c.entryPath(inputs)
	if err != nil {
		return nil, cacheHeader{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, cacheHeader{}, false
	}

	header, err := validateCacheHeader(data, inputs, c.configHash)
	if err != nil {
		return nil, cacheHeader{}, false
	}

	return data, header, true
}

func (c *cacheStore) save(inputs []packageInput, issues []rules.Issue) error {
	if !c.enabledForRun() || len(inputs) == 0 {
		return nil
	}

	path, err := c.entryPath(inputs)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := encodeCache(inputs, issues, c.configHash)

	tmp, err := os.CreateTemp(filepath.Dir(path), "*.tmp")
	if err != nil {
		return err
	}

	_, writeErr := tmp.Write(payload)
	closeErr := tmp.Close()
	if writeErr != nil {
		_ = os.Remove(tmp.Name())
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp.Name())
		return closeErr
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	return nil
}

func probePackageInputs(paths []string) ([]packageInput, error) {
	inputs := make([]packageInput, 0, len(paths))

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, packageProbeFromInfo(path, info))
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].NormalizedPath < inputs[j].NormalizedPath
	})

	return inputs, nil
}

func packageProbeFromInfo(path string, info os.FileInfo) packageInput {
	stamp := readFileStamp(info)

	return packageInput{
		Path:               path,
		NormalizedPath:     normalizeIssuePath(path),
		Size:               info.Size(),
		ModTimeUnixNano:    info.ModTime().UnixNano(),
		ChangeTimeUnixNano: stamp.ChangeTimeUnixNano,
		Device:             stamp.Device,
		Inode:              stamp.Inode,
		FastPathSupported:  stamp.FastPathSupported,
	}
}

func loadPackageInputs(paths []string) ([]packageInput, error) {
	inputs, err := probePackageInputs(paths)
	if err != nil {
		return nil, err
	}

	if err := loadPackageSources(inputs); err != nil {
		return nil, err
	}

	return inputs, nil
}

func loadPackageSources(inputs []packageInput) error {
	for i := range inputs {
		if err := loadSourceIntoInput(&inputs[i]); err != nil {
			return err
		}
	}

	return nil
}

func loadSourceIntoInput(input *packageInput) error {
	if input.Src != nil {
		return nil
	}

	src, err := os.ReadFile(input.Path)
	if err != nil {
		return err
	}

	input.Src = src
	input.Hash = sha256.Sum256(src)
	input.SampleHash = sampleHashFromBytes(src)

	return nil
}

func matchesFastPath(input packageInput, cached packageInput) bool {
	return input.Size == cached.Size &&
		input.ModTimeUnixNano == cached.ModTimeUnixNano &&
		input.ChangeTimeUnixNano == cached.ChangeTimeUnixNano &&
		input.Device == cached.Device &&
		input.Inode == cached.Inode
}

func encodeCache(inputs []packageInput, issues []rules.Issue, configHash string) []byte {
	buf := make([]byte, 0, estimateCacheSize(inputs, issues, configHash))
	pathIndex := make(map[string]uint64, len(inputs))

	buf = append(buf, cacheMagic...)
	buf = appendString(buf, configHash)
	buf = appendUvarint(buf, uint64(len(inputs)))

	for i, input := range inputs {
		pathIndex[input.NormalizedPath] = uint64(i + 1)
		buf = appendString(buf, input.NormalizedPath)
		buf = appendVarint(buf, input.Size)
		buf = appendVarint(buf, input.ModTimeUnixNano)
		buf = appendVarint(buf, input.ChangeTimeUnixNano)
		buf = appendUvarint(buf, input.Device)
		buf = appendUvarint(buf, input.Inode)
		buf = append(buf, input.Hash[:]...)
		buf = append(buf, input.SampleHash[:]...)
	}

	buf = appendUvarint(buf, uint64(len(issues)))
	buf = appendUvarint(buf, uint64(countFixableIssues(issues)))

	for _, issue := range issues {
		normalizedPath := normalizeIssuePath(issue.Filename())
		ref := pathIndex[normalizedPath]

		buf = appendUvarint(buf, ref)
		if ref == 0 {
			buf = appendString(buf, normalizedPath)
		}

		buf = appendUvarint(buf, uint64(issue.Line))
		buf = appendUvarint(buf, uint64(issue.Column))
		buf = appendUvarint(buf, uint64(issue.ID))
		buf = append(buf, issue.Flags, byte(issue.Severity))
		buf = appendUvarint(buf, uint64(issue.ArgInt1))
		buf = appendUvarint(buf, uint64(issue.ArgInt2))
		buf = appendString(buf, issue.ArgStr1)
	}

	return buf
}

func validateCacheHeader(data []byte, inputs []packageInput, configHash string) (cacheHeader, error) {
	dec := cacheDecoder{data: data}
	header := cacheHeader{}

	magic, err := dec.readBytes(len(cacheMagic))
	if err != nil {
		return header, err
	}
	if string(magic) != cacheMagic {
		return header, errors.New("invalid cache magic")
	}

	storedConfigHash, err := dec.readString()
	if err != nil {
		return header, err
	}
	if storedConfigHash != configHash {
		return header, errors.New("stale cache config")
	}

	fileCount64, err := dec.readUvarint()
	if err != nil {
		return header, err
	}
	if int(fileCount64) != len(inputs) {
		return header, errors.New("unexpected cache file count")
	}

	for i := range inputs {
		input := &inputs[i]

		path, err := dec.readString()
		if err != nil {
			return header, err
		}
		if path != input.NormalizedPath {
			return header, errors.New("unexpected cache path")
		}
		size, err := dec.readVarint()
		if err != nil {
			return header, err
		}
		modTime, err := dec.readVarint()
		if err != nil {
			return header, err
		}
		changeTime, err := dec.readVarint()
		if err != nil {
			return header, err
		}
		device, err := dec.readUvarint()
		if err != nil {
			return header, err
		}
		inode, err := dec.readUvarint()
		if err != nil {
			return header, err
		}
		hash, err := dec.readHash()
		if err != nil {
			return header, err
		}
		sampleHash, err := dec.readHash()
		if err != nil {
			return header, err
		}

		cached := packageInput{
			NormalizedPath:     path,
			Size:               size,
			ModTimeUnixNano:    modTime,
			ChangeTimeUnixNano: changeTime,
			Device:             device,
			Inode:              inode,
			Hash:               hash,
			SampleHash:         sampleHash,
		}

		if input.FastPathSupported && matchesFastPath(*input, cached) {
			currentSampleHash, err := readSampleHash(input.Path, input.Size)
			if err == nil && currentSampleHash == sampleHash {
				continue
			}
		}

		if err := loadSourceIntoInput(input); err != nil {
			return header, err
		}
		if input.Hash != hash {
			return header, errors.New("stale cache content")
		}
	}

	issueCount64, err := dec.readUvarint()
	if err != nil {
		return header, err
	}
	fixableCount64, err := dec.readUvarint()
	if err != nil {
		return header, err
	}

	header.IssueCount = int(issueCount64)
	header.FixableCount = int(fixableCount64)
	header.IssueStart = dec.off

	return header, nil
}

func decodeCachedIssues(data []byte, header cacheHeader, inputs []packageInput, limit int) ([]rules.Issue, bool) {
	dec := cacheDecoder{data: data, off: header.IssueStart}
	target := header.IssueCount
	if limit > 0 && limit < target {
		target = limit
	}
	issues := make([]rules.Issue, 0, target)

	for i := 0; i < header.IssueCount; i++ {
		pathRef, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}

		path := ""
		if pathRef > 0 {
			idx := int(pathRef - 1)
			if idx < 0 || idx >= len(inputs) {
				return nil, false
			}
			path = inputs[idx].Path
		} else {
			path, err = dec.readString()
			if err != nil {
				return nil, false
			}
		}

		line, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}
		column, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}
		id, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}
		flags, err := dec.readByte()
		if err != nil {
			return nil, false
		}
		severity, err := dec.readByte()
		if err != nil {
			return nil, false
		}
		argInt1, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}
		argInt2, err := dec.readUvarint()
		if err != nil {
			return nil, false
		}
		argStr1, err := dec.readString()
		if err != nil {
			return nil, false
		}

		if len(issues) < target {
			issues = append(issues, rules.Issue{
				Path:     path,
				Line:     uint32(line),
				Column:   uint32(column),
				ID:       uint16(id),
				Flags:    flags,
				Severity: rules.Severity(severity),
				ArgInt1:  uint32(argInt1),
				ArgInt2:  uint32(argInt2),
				ArgStr1:  argStr1,
			})
		}
	}

	return issues, true
}

func decodeCachedIssuesInto(dst []rules.Issue, batch *cachedBatch) (int, bool) {
	if batch == nil {
		return 0, true
	}

	dec := cacheDecoder{data: batch.data, off: batch.issueStart}
	count := 0

	for i := 0; i < batch.issueCount; i++ {
		pathRef, err := dec.readUvarint()
		if err != nil {
			return count, false
		}

		path := ""
		if pathRef > 0 {
			idx := int(pathRef - 1)
			if idx < 0 || idx >= len(batch.inputs) {
				return count, false
			}
			path = batch.inputs[idx].Path
		} else {
			path, err = dec.readString()
			if err != nil {
				return count, false
			}
		}

		line, err := dec.readUvarint()
		if err != nil {
			return count, false
		}
		column, err := dec.readUvarint()
		if err != nil {
			return count, false
		}
		id, err := dec.readUvarint()
		if err != nil {
			return count, false
		}
		flags, err := dec.readByte()
		if err != nil {
			return count, false
		}
		severity, err := dec.readByte()
		if err != nil {
			return count, false
		}
		argInt1, err := dec.readUvarint()
		if err != nil {
			return count, false
		}
		argInt2, err := dec.readUvarint()
		if err != nil {
			return count, false
		}
		argStr1, err := dec.readString()
		if err != nil {
			return count, false
		}

		dst[count] = rules.Issue{
			Path:     path,
			Line:     uint32(line),
			Column:   uint32(column),
			ID:       uint16(id),
			Flags:    flags,
			Severity: rules.Severity(severity),
			ArgInt1:  uint32(argInt1),
			ArgInt2:  uint32(argInt2),
			ArgStr1:  argStr1,
		}
		count++
	}

	return count, true
}

func estimateCacheSize(inputs []packageInput, issues []rules.Issue, configHash string) int {
	size := len(cacheMagic) + len(configHash) + len(inputs)*((cacheHashSize*2)+64) + len(issues)*48

	for _, input := range inputs {
		size += len(input.NormalizedPath)
	}

	for _, issue := range issues {
		size += len(issue.ArgStr1)
	}

	return size
}

func countFixableIssues(issues []rules.Issue) int {
	count := 0

	for _, issue := range issues {
		if issue.IsFixable() {
			count++
		}
	}

	return count
}

func normalizeIssuePath(path string) string {
	if path == "" {
		return ""
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	return abs
}

func truncateIssues(issues []rules.Issue, max int) []rules.Issue {
	if max <= 0 {
		return issues[:0]
	}

	if len(issues) <= max {
		return issues
	}

	return issues[:max]
}

func appendUvarint(dst []byte, value uint64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], value)
	return append(dst, buf[:n]...)
}

func appendVarint(dst []byte, value int64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], value)
	return append(dst, buf[:n]...)
}

func appendString(dst []byte, value string) []byte {
	dst = appendUvarint(dst, uint64(len(value)))
	return append(dst, value...)
}

func (d *cacheDecoder) readUvarint() (uint64, error) {
	if d.off >= len(d.data) {
		return 0, errors.New("unexpected end of cache")
	}

	value, n := binary.Uvarint(d.data[d.off:])
	if n <= 0 {
		return 0, errors.New("invalid cache uvarint")
	}

	d.off += n

	return value, nil
}

func (d *cacheDecoder) readVarint() (int64, error) {
	if d.off >= len(d.data) {
		return 0, errors.New("unexpected end of cache")
	}

	value, n := binary.Varint(d.data[d.off:])
	if n <= 0 {
		return 0, errors.New("invalid cache varint")
	}

	d.off += n

	return value, nil
}

func (d *cacheDecoder) readString() (string, error) {
	length, err := d.readUvarint()
	if err != nil {
		return "", err
	}

	bytes, err := d.readBytes(int(length))
	if err != nil {
		return "", err
	}

	if len(bytes) == 0 {
		return "", nil
	}

	return unsafe.String(unsafe.SliceData(bytes), len(bytes)), nil
}

func (d *cacheDecoder) readBytes(length int) ([]byte, error) {
	if length < 0 || d.off+length > len(d.data) {
		return nil, errors.New("unexpected end of cache")
	}

	bytes := d.data[d.off : d.off+length]
	d.off += length

	return bytes, nil
}

func (d *cacheDecoder) readByte() (byte, error) {
	if d.off >= len(d.data) {
		return 0, errors.New("unexpected end of cache")
	}

	value := d.data[d.off]
	d.off++

	return value, nil
}

func (d *cacheDecoder) readHash() ([cacheHashSize]byte, error) {
	var hash [cacheHashSize]byte

	bytes, err := d.readBytes(cacheHashSize)
	if err != nil {
		return hash, err
	}

	copy(hash[:], bytes)

	return hash, nil
}

func sampleHashFromBytes(src []byte) [cacheHashSize]byte {
	if len(src) <= cacheSampleChunk*cacheSampleWindows {
		return sha256.Sum256(src)
	}

	hash := sha256.New()

	for _, window := range sampleWindows(int64(len(src))) {
		if window.end <= window.start {
			continue
		}

		_, _ = hash.Write(src[window.start:window.end])
	}

	return sliceHashToArray(hash.Sum(nil))
}

func readSampleHash(path string, size int64) ([cacheHashSize]byte, error) {
	if size <= cacheSampleChunk*cacheSampleWindows {
		file, err := os.Open(path)
		if err != nil {
			return [cacheHashSize]byte{}, err
		}
		defer file.Close()

		var buffer [cacheSampleChunk * cacheSampleWindows]byte
		n, err := file.Read(buffer[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return [cacheHashSize]byte{}, err
		}

		return sha256.Sum256(buffer[:n]), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return [cacheHashSize]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	var buffer [cacheSampleChunk]byte

	for _, window := range sampleWindows(size) {
		if window.end <= window.start {
			continue
		}

		length := window.end - window.start
		n, err := file.ReadAt(buffer[:length], int64(window.start))
		if err != nil && !errors.Is(err, io.EOF) {
			return [cacheHashSize]byte{}, err
		}

		_, _ = hash.Write(buffer[:n])
	}

	return sliceHashToArray(hash.Sum(nil)), nil
}

type sampleWindow struct {
	start int
	end   int
}

func sampleWindows(size int64) []sampleWindow {
	if size <= cacheSampleChunk*cacheSampleWindows {
		return []sampleWindow{{start: 0, end: int(size)}}
	}

	offsets := [cacheSampleWindows]int64{
		0,
		size / 3,
		(2 * size) / 3,
		size - cacheSampleChunk,
	}
	windows := make([]sampleWindow, 0, cacheSampleWindows)
	lastEnd := -1

	for _, offset := range offsets {
		if offset < 0 {
			offset = 0
		}
		if offset > size-cacheSampleChunk {
			offset = size - cacheSampleChunk
		}

		start := int(offset)
		if start < lastEnd {
			start = lastEnd
		}

		end := start + cacheSampleChunk
		if end > int(size) {
			end = int(size)
		}
		if start >= end {
			continue
		}

		windows = append(windows, sampleWindow{start: start, end: end})
		lastEnd = end
	}

	return windows
}

func sliceHashToArray(sum []byte) [cacheHashSize]byte {
	var hash [cacheHashSize]byte
	copy(hash[:], sum)
	return hash
}
