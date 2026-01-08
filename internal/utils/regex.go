package utils

import "regexp"

func RegMatch(input, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(input)
}

func RegFind(input, pattern string) bool {
	return RegMatch(input, pattern)
}

func RegGetMatch(input, pattern string, group int) (string, bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", false
	}
	matches := re.FindStringSubmatch(input)
	if matches == nil || group >= len(matches) {
		return "", false
	}
	return matches[group], true
}

func RegReplace(input, pattern, replacement string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}
	return re.ReplaceAllString(input, replacement)
}

func RegValid(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}
