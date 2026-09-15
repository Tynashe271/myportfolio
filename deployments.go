package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type pageSite struct {
	HTMLURL string `json:"html_url"`
	Public  bool   `json:"public"`
}

type deploymentResult struct {
	url     string
	expires time.Time
}

type deploymentLookup struct {
	client  *http.Client
	apiBase string
	mu      sync.Mutex
	cache   map[string]deploymentResult
}

func newDeploymentLookup() *deploymentLookup {
	return &deploymentLookup{
		client:  &http.Client{Timeout: 3 * time.Second},
		apiBase: "https://api.github.com",
		cache:   make(map[string]deploymentResult),
	}
}

func (lookup *deploymentLookup) updateProjects(ctx context.Context, projects []project) {
	var pending sync.WaitGroup
	for i := range projects {
		if projects[i].Deployed || projects[i].Repo == "" {
			continue
		}
		pending.Add(1)
		go func(i int) {
			defer pending.Done()
			if siteURL := lookup.pagesURL(ctx, projects[i].Repo); siteURL != "" {
				projects[i].URL = siteURL
				projects[i].Status = "Live"
				projects[i].Deployed = true
			}
		}(i)
	}
	pending.Wait()
}

func (lookup *deploymentLookup) pagesURL(ctx context.Context, repo string) string {
	lookup.mu.Lock()
	if cached, ok := lookup.cache[repo]; ok && time.Now().Before(cached.expires) {
		lookup.mu.Unlock()
		return cached.url
	}
	lookup.mu.Unlock()

	endpoint := lookup.apiBase + "/repos/Tynashe271/" + url.PathEscape(repo) + "/pages"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "MyPortfolio")
	response, err := lookup.client.Do(request)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	result := deploymentResult{expires: time.Now().Add(10 * time.Minute)}
	if response.StatusCode == http.StatusOK {
		var site pageSite
		if json.NewDecoder(response.Body).Decode(&site) == nil && site.Public && strings.HasPrefix(site.HTMLURL, "https://") {
			result.url = site.HTMLURL
		}
	} else if response.StatusCode != http.StatusNotFound {
		result.expires = time.Now().Add(time.Minute)
	}
	lookup.mu.Lock()
	lookup.cache[repo] = result
	lookup.mu.Unlock()
	return result.url
}
