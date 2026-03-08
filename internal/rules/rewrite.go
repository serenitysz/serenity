package rules

import (
	"go/ast"
	"reflect"
)

func ReplaceNode(parent ast.Node, oldNode, newNode ast.Node) bool {
	if parent == nil || oldNode == nil || newNode == nil {
		return false
	}

	value := reflect.ValueOf(parent)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return false
	}

	return replaceInValue(value.Elem(), oldNode, newNode)
}

func replaceInValue(value reflect.Value, oldNode, newNode ast.Node) bool {
	if !value.IsValid() {
		return false
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return false
		}

		node, ok := value.Interface().(ast.Node)
		if !ok || node != oldNode {
			return false
		}

		replacement := reflect.ValueOf(newNode)
		if !replacement.Type().AssignableTo(value.Type()) {
			return false
		}

		value.Set(replacement)
		return true
	case reflect.Slice:
		for i := range value.Len() {
			entry := value.Index(i)
			if entry.Kind() != reflect.Interface || entry.IsNil() {
				continue
			}

			node, ok := entry.Interface().(ast.Node)
			if !ok || node != oldNode {
				continue
			}

			replacement := reflect.ValueOf(newNode)
			if !replacement.Type().AssignableTo(entry.Type()) {
				return false
			}

			entry.Set(replacement)
			return true
		}
	case reflect.Struct:
		for i := range value.NumField() {
			field := value.Field(i)
			if !field.CanSet() {
				continue
			}

			if replaceInValue(field, oldNode, newNode) {
				return true
			}
		}
	}

	return false
}
