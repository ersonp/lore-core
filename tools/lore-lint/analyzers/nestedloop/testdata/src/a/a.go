package a

type Item struct {
	ID string
}

func bad(items []Item) {
	for _, a := range items {
		for _, b := range items { // want "O\\(n²\\) pattern: nested loop over same collection"
			if a.ID != b.ID {
				_ = a.ID + b.ID
			}
		}
	}
}

func good(items []Item, others []Item) {
	// Different collections - OK
	for _, a := range items {
		for _, b := range others {
			_ = a.ID + b.ID
		}
	}
}

func goodSingleLoop(items []Item) {
	// Single loop - OK
	for _, item := range items {
		_ = item.ID
	}
}
