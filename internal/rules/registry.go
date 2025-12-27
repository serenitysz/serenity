package rules

var GlobalRegistry []Rule

func Register(r Rule) {
	GlobalRegistry = append(GlobalRegistry, r)
}
