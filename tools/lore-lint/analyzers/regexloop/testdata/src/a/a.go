package a

import "regexp"

func bad(texts []string) {
	for _, text := range texts {
		re := regexp.MustCompile(`\d+`) // want "regexp.MustCompile called inside loop"
		_ = re.FindAllString(text, -1)
	}
}

func badCompile(texts []string) {
	for _, text := range texts {
		re, _ := regexp.Compile(`\d+`) // want "regexp.Compile called inside loop"
		_ = re.FindAllString(text, -1)
	}
}

func good(texts []string) {
	re := regexp.MustCompile(`\d+`)
	for _, text := range texts {
		_ = re.FindAllString(text, -1)
	}
}

var globalRe = regexp.MustCompile(`\d+`)

func goodGlobal(texts []string) {
	for _, text := range texts {
		_ = globalRe.FindAllString(text, -1)
	}
}
