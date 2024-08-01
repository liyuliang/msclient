package msclient

import "regexp"

func RegexGet(content string, expr string) string {
	result := ""
	rule, _ := regexp.Compile(expr)
	allMatch := rule.FindStringSubmatch(content)
	if 2 == len(allMatch) {
		result = allMatch[1]
	}
	return result
}
