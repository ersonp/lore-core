package a

func bad(cache map[string]int, key string) {
	if cache[key] != 0 {
		process(cache[key]) // want "repeated map lookup"
	}
}

func badWithPointer(cache map[string]*int, key string) {
	if cache[key] != nil {
		use(cache[key]) // want "repeated map lookup"
	}
}

func good(cache map[string]int, key string) {
	if v := cache[key]; v != 0 {
		process(v)
	}
}

func goodCommaOk(cache map[string]int, key string) {
	if v, ok := cache[key]; ok {
		process(v)
	}
}

func goodDifferentKeys(cache map[string]int, key1, key2 string) {
	if cache[key1] != 0 {
		process(cache[key2]) // Different keys - OK
	}
}

func process(v int) {}
func use(v *int)    {}
