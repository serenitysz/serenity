package rules

type Severity uint8

const (
	SeverityInfo  Severity = iota
	SeverityWarn  Severity = iota
	SeverityError Severity = iota
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarn:
		return "warn"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

func ParseSeverity(s string) Severity {
	switch s {
	case "error":
		return SeverityError
	case "warn":
		return SeverityWarn
	case "info":
		return SeverityInfo
	default:
		return SeverityWarn
	}
}
