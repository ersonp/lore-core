package a

import "context"

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type VectorDB interface {
	Save(ctx context.Context, id string) error
}

func bad(ctx context.Context, items []string, e Embedder, db VectorDB) {
	for _, item := range items {
		e.Embed(ctx, item) // want "potential N\\+1: Embed called inside loop"
		db.Save(ctx, item) // want "potential N\\+1: Save called inside loop"
	}
}

func good(ctx context.Context, items []string) {
	// No external calls - should not flag
	for _, item := range items {
		_ = len(item)
	}
}
