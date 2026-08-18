package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const arbeitnowAPIURL = "https://www.arbeitnow.com/api/job-board-api"

type arbeitnowSource struct {
	Data []struct {
		Slug        string   `json:"slug"`
		CompanyName string   `json:"company_name"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Remote      bool     `json:"remote"`
		URL         string   `json:"url"`
		Location    string   `json:"location"`
		JobTypes    []string `json:"job_types"`
	} `json:"data"`
}

type ArbeitnowJobSource struct {
	Client   *http.Client
	Fallback JobSource
}

func NewArbeitnowJobSource(fallback JobSource) *ArbeitnowJobSource {
	return &ArbeitnowJobSource{Client: &http.Client{Timeout: 15 * time.Second}, Fallback: fallback}
}

func (s *ArbeitnowJobSource) FetchJobs(ctx context.Context) ([]SourceJob, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, arbeitnowAPIURL, nil)
	if err != nil {
		return s.fallback(ctx, err)
	}
	response, err := s.Client.Do(request)
	if err != nil {
		return s.fallback(ctx, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return s.fallback(ctx, fmt.Errorf("arbeitnow returned HTTP %d", response.StatusCode))
	}
	var payload arbeitnowSource
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return s.fallback(ctx, fmt.Errorf("decode arbeitnow response: %w", err))
	}
	jobs := make([]SourceJob, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.Slug) == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		jobs = append(jobs, SourceJob{ExternalID: item.Slug, Title: item.Title, Company: item.CompanyName, Location: item.Location, Description: item.Description, Remote: item.Remote, SourceURL: item.URL, Salary: strings.Join(item.JobTypes, ", ")})
	}
	return jobs, nil
}

func (s *ArbeitnowJobSource) fallback(ctx context.Context, err error) ([]SourceJob, error) {
	if s.Fallback == nil {
		return nil, err
	}
	return s.Fallback.FetchJobs(ctx)
}
