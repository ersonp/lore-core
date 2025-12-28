package a

func bad(items []string) string {
	var result string
	for _, item := range items {
		result += item // want "O\\(n²\\) string concatenation in loop"
	}
	return result
}

func badWithConcat(items []string) string {
	var result string
	for _, item := range items {
		result += item + ", " // want "O\\(n²\\) string concatenation in loop"
	}
	return result
}

func good(items []string) string {
	// Integer addition is fine
	var count int
	for range items {
		count += 1
	}
	_ = count
	return ""
}

func goodForLoop() string {
	// Regular for loop with int
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	_ = sum
	return ""
}
