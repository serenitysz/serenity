package rules

import (
	"fmt"
	"go/ast"
)

type Suppression struct {
	RuleName string
	Reason   string
	Line     int
}

func ProcessSuppressions(comments []*ast.CommentGroup) []Suppression {
	for _, comment := range comments {
		fmt.Println(comment.Text())
	}
	return []Suppression{}
}
