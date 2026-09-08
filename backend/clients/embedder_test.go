package clients

import (
	"context"
	"testing"

	"skill-match/backend/models"
)

func TestLocalEmbedderDimension(t *testing.T) {
	e := NewLocalEmbedder()
	vec, err := e.Generate(context.Background(), "senior Go engineer with PostgreSQL and Kubernetes")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(vec) != models.EmbeddingDim {
		t.Fatalf("dimension = %d, want %d", len(vec), models.EmbeddingDim)
	}
}

func TestLocalEmbedderDeterministic(t *testing.T) {
	e := NewLocalEmbedder()
	a, err := e.Generate(context.Background(), "senior Go engineer with PostgreSQL")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, err := e.Generate(context.Background(), "senior Go engineer with PostgreSQL")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("embedding not deterministic at index %d", i)
		}
	}
}

func TestLocalEmbedderRejectsEmpty(t *testing.T) {
	e := NewLocalEmbedder()
	if _, err := e.Generate(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty text")
	}
}

// Cosine similarity of two unit vectors; identical wording must be ~1.0.
func TestLocalEmbedderSimilarTextClose(t *testing.T) {
	e := NewLocalEmbedder()
	ctx := context.Background()
	a, _ := e.Generate(ctx, "Go developer PostgreSQL Kubernetes")
	b, _ := e.Generate(ctx, "Go developer PostgreSQL")
	if dot(a, b) < 0.5 {
		t.Fatalf("expected overlapping skill vectors to be similar, dot=%f", dot(a, b))
	}
	// Dissimilar text should be noticeably farther.
	c, _ := e.Generate(ctx, "pastry chef patisserie baking desserts")
	if dot(a, c) > dot(a, b) {
		t.Fatalf("expected dissimilar text to be no closer than overlapping text")
	}
}

func dot(a, b []float32) float64 {
	var s float64
	for i := range a {
		s += float64(a[i]) * float64(b[i])
	}
	return s
}
