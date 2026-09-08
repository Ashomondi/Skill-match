package clients

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"strings"

	"skill-match/backend/models"
)

// EmbeddingGenerator produces a fixed-dimension vector for a piece of text.
// Different backends can implement it (Amazon Bedrock Titan, a local
// deterministic embedder, or any future provider) and be swapped without
// changing callers.
type EmbeddingGenerator interface {
	Generate(ctx context.Context, text string) ([]float32, error)
}

// LocalEmbedder is a deterministic, dependency-free embedding provider that
// always satisfies models.EmbeddingDim. It hashes token n-grams into a
// signed 1024-dim vector and L2-normalizes the result so cosine similarity
// (what pgvector's vector_cosine_ops index uses) behaves correctly.
//
// It is a stand-in for a real semantic model: identical/similar wording
// yields nearby vectors, which is enough to keep the whole embedding
// pipeline (upsert, HNSW search, FindSimilarJobs) working with zero external
// services. Swap in a semantic model later without touching callers.
type LocalEmbedder struct{}

// NewLocalEmbedder constructs a LocalEmbedder.
func NewLocalEmbedder() *LocalEmbedder {
	return &LocalEmbedder{}
}

func (e *LocalEmbedder) Generate(_ context.Context, text string) ([]float32, error) {
	vec := make([]float32, models.EmbeddingDim)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is required to generate an embedding")
	}

	seen := make(map[string]bool)
	tokens := tokenize(text)
	for _, tok := range tokens {
		if seen[tok] {
			continue
		}
		seen[tok] = true

		h1 := fnv.New64a()
		h1.Write([]byte(tok))
		lo := h1.Sum64()

		h2 := fnv.New64a()
		h2.Write([]byte("dim:" + tok))
		hi := h2.Sum64()

		// Two hash-derived indices so the token energy spreads over the
		// vector instead of collapsing onto a single axis.
		idxA := int(lo % uint64(models.EmbeddingDim))
		idxB := int(hi % uint64(models.EmbeddingDim))
		sign := float32(1)
		if lo&1 == 0 {
			sign = -1
		}

		vec[idxA] += sign
		if idxB != idxA {
			vec[idxB] += sign * 0.5
		}
	}

	normalize(vec)
	return vec, nil
}

// tokenize lowercases text and splits on non-alphanumeric boundaries.
func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= 128)
	})
}

// normalize scales vec to unit length (in place) so cosine distance is the
// geometric angle rather than magnitude-dependent.
func normalize(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		return
	}
	norm := float32(math.Sqrt(sum))
	for i := range vec {
		vec[i] /= norm
	}
}

// BedrockEmbedder adapts BedrockClient.GenerateEmbedding to the
// EmbeddingGenerator interface by binding a fixed embed model id at
// construction.
type BedrockEmbedder struct {
	client  *BedrockClient
	modelID string
}

// NewBedrockEmbedder builds a Bedrock-backed EmbeddingGenerator.
func NewBedrockEmbedder(client *BedrockClient, modelID string) (*BedrockEmbedder, error) {
	if client == nil || strings.TrimSpace(modelID) == "" {
		return nil, fmt.Errorf("bedrock embedder requires a client and embed model id")
	}
	return &BedrockEmbedder{client: client, modelID: modelID}, nil
}

// Generate embeds text via Amazon Bedrock Titan.
func (b *BedrockEmbedder) Generate(ctx context.Context, text string) ([]float32, error) {
	return b.client.GenerateEmbedding(ctx, b.modelID, text)
}
