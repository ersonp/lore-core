package integration

import (
	"context"
	"os"
	"testing"

	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

const (
	testQdrantHost = "localhost"
	testQdrantPort = 6334
	testCollection = "lore_integration_test"
)

var testRepo *qdrant.Repository

func TestMain(m *testing.M) {
	// Skip if INTEGRATION_TEST is not set
	if os.Getenv("INTEGRATION_TEST") != "1" {
		os.Exit(0)
	}

	// Setup
	cfg := config.QdrantConfig{
		Host:       testQdrantHost,
		Port:       testQdrantPort,
		Collection: testCollection,
	}

	var err error
	testRepo, err = qdrant.NewRepository(cfg)
	if err != nil {
		panic("failed to create repository: " + err.Error())
	}

	// Ensure clean collection
	ctx := context.Background()
	_ = testRepo.DeleteCollection(ctx) // Ignore error if collection doesn't exist
	if err := testRepo.EnsureCollection(ctx, uint64(embedder.VectorSize)); err != nil {
		panic("failed to create collection: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = testRepo.DeleteCollection(ctx)
	testRepo.Close()

	os.Exit(code)
}

// cleanupFacts removes all facts between tests.
func cleanupFacts(t *testing.T) {
	t.Helper()
	if err := testRepo.DeleteAll(t.Context()); err != nil {
		t.Fatalf("failed to cleanup facts: %v", err)
	}
}
