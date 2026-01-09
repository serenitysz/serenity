package rules

var GlobalRegistry []Rule

func Register(rule Rule) {
	GlobalRegistry = append(GlobalRegistry, rule)
}
