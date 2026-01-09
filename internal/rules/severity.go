package rules

type Severity uint8

const (
	SeverityInfo  Severity = iota
	SeverityWarn  Severity = iota
	SeverityError Severity = iota
)

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
