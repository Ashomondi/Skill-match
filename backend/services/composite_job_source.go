package services

import (
	"context"
	"fmt"
	"log/slog"
)

type CompositeJobSource struct {
	sources []JobSource
}

func NewCompositeJobSource(sources ...JobSource) *CompositeJobSource {
	return &CompositeJobSource{sources: sources}
}

func (c *CompositeJobSource) FetchJobs(ctx context.Context) ([]SourceJob, error) {
	var allJobs []SourceJob
	var errs []error

	for _, src := range c.sources {
		jobs, err := src.FetchJobs(ctx)
		if err != nil {
			slog.Warn("job_source: fetch error, continuing to next source", "error", err)
			errs = append(errs, err)
			continue
		}
		allJobs = append(allJobs, jobs...)
	}

	if len(allJobs) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("composite_source: all job sources failed: %v", errs)
	}

	return allJobs, nil
}
