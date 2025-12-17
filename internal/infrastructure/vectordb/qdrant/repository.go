// Package qdrant provides a VectorDB implementation using Qdrant.
package qdrant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

// Repository implements the VectorDB interface using Qdrant.
type Repository struct {
	client     pb.CollectionsClient
	points     pb.PointsClient
	collection string
	conn       *grpc.ClientConn
}

// NewRepository creates a new Qdrant repository.
func NewRepository(cfg config.QdrantConfig) (*Repository, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connecting to qdrant: %w", err)
	}

	return &Repository{
		client:     pb.NewCollectionsClient(conn),
		points:     pb.NewPointsClient(conn),
		collection: cfg.Collection,
		conn:       conn,
	}, nil
}

// Close closes the gRPC connection.
func (r *Repository) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// EnsureCollection creates the collection if it doesn't exist.
func (r *Repository) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	_, err := r.client.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: r.collection,
	})
	if err == nil {
		return nil
	}

	_, err = r.client.Create(ctx, &pb.CreateCollection{
		CollectionName: r.collection,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     vectorSize,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}

	return nil
}

// Save stores a fact with its embedding.
func (r *Repository) Save(ctx context.Context, fact entities.Fact) error {
	return r.SaveBatch(ctx, []entities.Fact{fact})
}

// SaveBatch stores multiple facts.
func (r *Repository) SaveBatch(ctx context.Context, facts []entities.Fact) error {
	points := make([]*pb.PointStruct, 0, len(facts))

	for _, fact := range facts {
		pointID := fact.ID
		if pointID == "" {
			pointID = uuid.New().String()
		}

		point := &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{
					Uuid: pointID,
				},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{
						Data: fact.Embedding,
					},
				},
			},
			Payload: map[string]*pb.Value{
				"type":        {Kind: &pb.Value_StringValue{StringValue: string(fact.Type)}},
				"subject":     {Kind: &pb.Value_StringValue{StringValue: fact.Subject}},
				"predicate":   {Kind: &pb.Value_StringValue{StringValue: fact.Predicate}},
				"object":      {Kind: &pb.Value_StringValue{StringValue: fact.Object}},
				"context":     {Kind: &pb.Value_StringValue{StringValue: fact.Context}},
				"source_file": {Kind: &pb.Value_StringValue{StringValue: fact.SourceFile}},
				"source_line": {Kind: &pb.Value_IntegerValue{IntegerValue: int64(fact.SourceLine)}},
				"confidence":  {Kind: &pb.Value_DoubleValue{DoubleValue: fact.Confidence}},
				"created_at":  {Kind: &pb.Value_StringValue{StringValue: fact.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}},
				"updated_at":  {Kind: &pb.Value_StringValue{StringValue: fact.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}},
			},
		}
		points = append(points, point)
	}

	_, err := r.points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: r.collection,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("upserting points: %w", err)
	}

	return nil
}

// FindByID retrieves a fact by its ID.
func (r *Repository) FindByID(ctx context.Context, id string) (entities.Fact, error) {
	resp, err := r.points.Get(ctx, &pb.GetPoints{
		CollectionName: r.collection,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Uuid{Uuid: id}},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return entities.Fact{}, fmt.Errorf("getting point: %w", err)
	}

	if len(resp.Result) == 0 {
		return entities.Fact{}, fmt.Errorf("fact not found: %s", id)
	}

	return pointToFact(resp.Result[0])
}

// Search performs a semantic search and returns similar facts.
func (r *Repository) Search(ctx context.Context, embedding []float32, limit int) ([]entities.Fact, error) {
	resp, err := r.points.Search(ctx, &pb.SearchPoints{
		CollectionName: r.collection,
		Vector:         embedding,
		Limit:          uint64(limit),
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("searching points: %w", err)
	}

	return scoredPointsToFacts(resp.Result)
}

// SearchByType performs a semantic search filtered by fact type.
func (r *Repository) SearchByType(ctx context.Context, embedding []float32, factType entities.FactType, limit int) ([]entities.Fact, error) {
	resp, err := r.points.Search(ctx, &pb.SearchPoints{
		CollectionName: r.collection,
		Vector:         embedding,
		Limit:          uint64(limit),
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "type",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{
									Keyword: string(factType),
								},
							},
						},
					},
				},
			},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("searching points by type: %w", err)
	}

	return scoredPointsToFacts(resp.Result)
}

// Delete removes a fact by its ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.points.Delete(ctx, &pb.DeletePoints{
		CollectionName: r.collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Points{
				Points: &pb.PointsIdsList{
					Ids: []*pb.PointId{
						{PointIdOptions: &pb.PointId_Uuid{Uuid: id}},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting point: %w", err)
	}

	return nil
}

// List returns all facts with pagination.
func (r *Repository) List(ctx context.Context, limit int, offset uint64) ([]entities.Fact, error) {
	var offsetPtr *pb.PointId
	if offset > 0 {
		offsetPtr = &pb.PointId{
			PointIdOptions: &pb.PointId_Num{Num: offset},
		}
	}

	resp, err := r.points.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: r.collection,
		Limit:          pb.PtrOf(uint32(limit)),
		Offset:         offsetPtr,
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: false},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scrolling points: %w", err)
	}

	return retrievedPointsToFacts(resp.Result)
}

// ListByType returns facts filtered by type.
func (r *Repository) ListByType(ctx context.Context, factType entities.FactType, limit int) ([]entities.Fact, error) {
	resp, err := r.points.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: r.collection,
		Limit:          pb.PtrOf(uint32(limit)),
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "type",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{
									Keyword: string(factType),
								},
							},
						},
					},
				},
			},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: false},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scrolling points by type: %w", err)
	}

	return retrievedPointsToFacts(resp.Result)
}

// ListBySource returns facts filtered by source file.
func (r *Repository) ListBySource(ctx context.Context, sourceFile string, limit int) ([]entities.Fact, error) {
	resp, err := r.points.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: r.collection,
		Limit:          pb.PtrOf(uint32(limit)),
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "source_file",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{
									Keyword: sourceFile,
								},
							},
						},
					},
				},
			},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: false},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scrolling points by source: %w", err)
	}

	return retrievedPointsToFacts(resp.Result)
}

// DeleteBySource removes all facts from a source file.
func (r *Repository) DeleteBySource(ctx context.Context, sourceFile string) error {
	_, err := r.points.Delete(ctx, &pb.DeletePoints{
		CollectionName: r.collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: &pb.Filter{
					Must: []*pb.Condition{
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "source_file",
									Match: &pb.Match{
										MatchValue: &pb.Match_Keyword{
											Keyword: sourceFile,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting points by source: %w", err)
	}

	return nil
}

// DeleteAll removes all facts.
func (r *Repository) DeleteAll(ctx context.Context) error {
	_, err := r.points.Delete(ctx, &pb.DeletePoints{
		CollectionName: r.collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: &pb.Filter{},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting all points: %w", err)
	}

	return nil
}

// Count returns the total number of facts.
func (r *Repository) Count(ctx context.Context) (uint64, error) {
	resp, err := r.client.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: r.collection,
	})
	if err != nil {
		return 0, fmt.Errorf("getting collection info: %w", err)
	}

	if resp.Result.PointsCount == nil {
		return 0, nil
	}

	return *resp.Result.PointsCount, nil
}

// retrievedPointsToFacts converts retrieved points to facts.
func retrievedPointsToFacts(points []*pb.RetrievedPoint) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0, len(points))

	for _, point := range points {
		fact, err := pointToFact(point)
		if err != nil {
			return nil, err
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

// pointToFact converts a Qdrant point to a Fact entity.
func pointToFact(point *pb.RetrievedPoint) (entities.Fact, error) {
	id := ""
	if uuid := point.Id.GetUuid(); uuid != "" {
		id = uuid
	}

	payload := point.Payload
	var embedding []float32
	if vec := point.Vectors.GetVector(); vec != nil {
		embedding = vec.Data
	}

	fact := entities.Fact{
		ID:         id,
		Type:       entities.FactType(getStringValue(payload, "type")),
		Subject:    getStringValue(payload, "subject"),
		Predicate:  getStringValue(payload, "predicate"),
		Object:     getStringValue(payload, "object"),
		Context:    getStringValue(payload, "context"),
		SourceFile: getStringValue(payload, "source_file"),
		SourceLine: int(getIntValue(payload, "source_line")),
		Confidence: getDoubleValue(payload, "confidence"),
		Embedding:  embedding,
	}

	return fact, nil
}

// scoredPointsToFacts converts scored points to facts.
func scoredPointsToFacts(points []*pb.ScoredPoint) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0, len(points))

	for _, point := range points {
		id := ""
		if uuid := point.Id.GetUuid(); uuid != "" {
			id = uuid
		}

		payload := point.Payload
		var embedding []float32
		if vec := point.Vectors.GetVector(); vec != nil {
			embedding = vec.Data
		}

		fact := entities.Fact{
			ID:         id,
			Type:       entities.FactType(getStringValue(payload, "type")),
			Subject:    getStringValue(payload, "subject"),
			Predicate:  getStringValue(payload, "predicate"),
			Object:     getStringValue(payload, "object"),
			Context:    getStringValue(payload, "context"),
			SourceFile: getStringValue(payload, "source_file"),
			SourceLine: int(getIntValue(payload, "source_line")),
			Confidence: getDoubleValue(payload, "confidence"),
			Embedding:  embedding,
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

// Helper functions for payload extraction.
func getStringValue(payload map[string]*pb.Value, key string) string {
	if v, ok := payload[key]; ok {
		return v.GetStringValue()
	}
	return ""
}

func getIntValue(payload map[string]*pb.Value, key string) int64 {
	if v, ok := payload[key]; ok {
		return v.GetIntegerValue()
	}
	return 0
}

func getDoubleValue(payload map[string]*pb.Value, key string) float64 {
	if v, ok := payload[key]; ok {
		return v.GetDoubleValue()
	}
	return 0
}
