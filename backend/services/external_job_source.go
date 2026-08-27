package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	remotiveAPIURL = "https://remotive.com/api/remote-jobs"
	remoteOKAPIURL = "https://remoteok.com/api"
)

type ExternalJobSource struct {
	Client   *http.Client
	Fallback JobSource
}

func NewExternalJobSource(fallback JobSource) *ExternalJobSource {
	return &ExternalJobSource{Client: &http.Client{Timeout: 15 * time.Second}, Fallback: fallback}
}
func (s *ExternalJobSource) FetchJobs(ctx context.Context) ([]SourceJob, error) {
	fetchers := []func(context.Context) ([]SourceJob, error){s.fetchArbeitnow, s.fetchRemotive, s.fetchRemoteOK}
	var jobs []SourceJob
	failures := 0
	for _, fetch := range fetchers {
		items, err := fetch(ctx)
		if err != nil {
			failures++
			continue
		}
		jobs = append(jobs, items...)
	}
	if failures < len(fetchers) {
		return jobs, nil
	}
	if s.Fallback != nil {
		return s.Fallback.FetchJobs(ctx)
	}
	return nil, fmt.Errorf("all external job sources failed")
}
func (s *ExternalJobSource) get(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "SkillMatch/1.0")
	res, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s returned HTTP %d", url, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(target)
}
func (s *ExternalJobSource) fetchArbeitnow(ctx context.Context) ([]SourceJob, error) {
	var p arbeitnowSource
	if err := s.get(ctx, arbeitnowAPIURL, &p); err != nil {
		return nil, err
	}
	out := []SourceJob{}
	for _, x := range p.Data {
		if strings.TrimSpace(x.Slug) == "" || strings.TrimSpace(x.Title) == "" {
			continue
		}
		out = append(out, SourceJob{ExternalID: "arbeitnow:" + x.Slug, Title: x.Title, Company: x.CompanyName, Location: x.Location, Description: x.Description, Remote: x.Remote, SourceURL: x.URL})
	}
	return out, nil
}
func (s *ExternalJobSource) fetchRemotive(ctx context.Context) ([]SourceJob, error) {
	var p struct {
		Jobs []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Company     string `json:"company_name"`
			Location    string `json:"candidate_required_location"`
			Description string `json:"description"`
			Salary      string `json:"salary"`
			URL         string `json:"url"`
		} `json:"jobs"`
	}
	if err := s.get(ctx, remotiveAPIURL, &p); err != nil {
		return nil, err
	}
	out := []SourceJob{}
	for _, x := range p.Jobs {
		if x.ID == 0 || strings.TrimSpace(x.Title) == "" {
			continue
		}
		out = append(out, SourceJob{ExternalID: fmt.Sprintf("remotive:%d", x.ID), Title: x.Title, Company: x.Company, Location: x.Location, Description: x.Description, Salary: x.Salary, Remote: true, SourceURL: x.URL})
	}
	return out, nil
}
func (s *ExternalJobSource) fetchRemoteOK(ctx context.Context) ([]SourceJob, error) {
	var p []struct {
		ID          string `json:"id"`
		Slug        string `json:"slug"`
		Title       string `json:"position"`
		Company     string `json:"company"`
		Location    string `json:"location"`
		Description string `json:"description"`
		Salary      string `json:"salary"`
		URL         string `json:"url"`
	}
	if err := s.get(ctx, remoteOKAPIURL, &p); err != nil {
		return nil, err
	}
	out := []SourceJob{}
	for _, x := range p {
		id := x.ID
		if id == "" {
			id = x.Slug
		}
		if id == "" || strings.TrimSpace(x.Title) == "" {
			continue
		}
		out = append(out, SourceJob{ExternalID: "remoteok:" + id, Title: x.Title, Company: x.Company, Location: x.Location, Description: x.Description, Salary: x.Salary, Remote: true, SourceURL: x.URL})
	}
	return out, nil
}
