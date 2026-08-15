package services

import "context"

// SeedJobSource provides a fixed set of jobs for demo/hackathon purposes,
// implementing JobSource. Swap for a real API-backed source later without
// changing JobService at all.
type SeedJobSource struct{}

func NewSeedJobSource() *SeedJobSource {
	return &SeedJobSource{}
}

func (s *SeedJobSource) FetchJobs(ctx context.Context) ([]SourceJob, error) {
	return []SourceJob{
		{ExternalID: "seed-001", Title: "Backend Engineer", Company: "Acme Corp", Location: "Nairobi, Kenya", Remote: true, Description: "Build and scale backend services in Go."},
		{ExternalID: "seed-002", Title: "Frontend Engineer", Company: "Globex", Location: "Kisumu, Kenya", Remote: true, Description: "React/TypeScript product engineering."},
		{ExternalID: "seed-003", Title: "Data Engineer", Company: "Initech", Location: "Remote", Remote: true, Description: "Pipelines and data infrastructure."},
	}, nil
}
