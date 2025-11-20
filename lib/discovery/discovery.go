package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/snowmerak/renovates/lib/renovate"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type Discoverer interface {
	ListRepositories(ctx context.Context) ([]string, error)
}

func NewDiscoverer(cfg *renovate.Config) (Discoverer, error) {
	switch cfg.Platform {
	case "github":
		return NewGitHubDiscoverer(cfg), nil
	case "gitlab":
		return NewGitLabDiscoverer(cfg)
	default:
		return nil, fmt.Errorf("unsupported platform for discovery: %s", cfg.Platform)
	}
}

// --- GitHub ---

type GitHubDiscoverer struct {
	client *github.Client
	cfg    *renovate.Config
}

func NewGitHubDiscoverer(cfg *renovate.Config) *GitHubDiscoverer {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	var client *github.Client
	if cfg.Endpoint != "" && cfg.Endpoint != "https://api.github.com" {
		// Enterprise URL handling might need adjustment depending on the exact URL format
		client, _ = github.NewEnterpriseClient(cfg.Endpoint, cfg.Endpoint, tc)
	} else {
		client = github.NewClient(tc)
	}

	return &GitHubDiscoverer{client: client, cfg: cfg}
}

func (d *GitHubDiscoverer) ListRepositories(ctx context.Context) ([]string, error) {
	var allRepos []string
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Filter by visibility if needed, but for now let's fetch all and filter locally
	// or use API filters if possible.
	// GitHub API 'type' option: all, owner, public, private, member. Default: all

	user := ""
	if d.cfg.Discovery.Owner != "" {
		user = d.cfg.Discovery.Owner
	}

	for {
		repos, resp, err := d.client.Repositories.List(ctx, user, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if d.match(repo) {
				allRepos = append(allRepos, *repo.FullName)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (d *GitHubDiscoverer) match(repo *github.Repository) bool {
	// Topic Filter
	if len(d.cfg.Discovery.Topics) > 0 {
		matched := false
		for _, topic := range repo.Topics {
			for _, target := range d.cfg.Discovery.Topics {
				if topic == target {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	// Regex Include
	if len(d.cfg.Discovery.Includes) > 0 {
		matched := false
		for _, pattern := range d.cfg.Discovery.Includes {
			if matched, _ = regexp.MatchString(pattern, *repo.Name); matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Regex Exclude
	if len(d.cfg.Discovery.Excludes) > 0 {
		for _, pattern := range d.cfg.Discovery.Excludes {
			if matched, _ := regexp.MatchString(pattern, *repo.Name); matched {
				return false
			}
		}
	}

	return true
}

// --- GitLab ---

type GitLabDiscoverer struct {
	client *gitlab.Client
	cfg    *renovate.Config
}

func NewGitLabDiscoverer(cfg *renovate.Config) (*GitLabDiscoverer, error) {
	opts := []gitlab.ClientOptionFunc{}
	if cfg.Endpoint != "" {
		opts = append(opts, gitlab.WithBaseURL(cfg.Endpoint))
	}

	client, err := gitlab.NewClient(cfg.Token, opts...)
	if err != nil {
		return nil, err
	}

	return &GitLabDiscoverer{client: client, cfg: cfg}, nil
}

func (d *GitLabDiscoverer) ListRepositories(ctx context.Context) ([]string, error) {
	var allRepos []string
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
		Simple:      gitlab.Ptr(true), // Get simple details to save bandwidth
	}

	if d.cfg.Discovery.Owner != "" {
		// In GitLab, filtering by group/user is a bit different.
		// We can search by group, or just list all projects the user has access to.
		// For simplicity, let's assume we list all projects and filter.
		// Or we can use ListGroupProjects if the owner is a group.
		// But "Owner" in config is ambiguous.
		// Let's try to use the search/filter options.
	}

	// If topics are provided, use them
	if len(d.cfg.Discovery.Topics) > 0 {
		opt.Topic = gitlab.Ptr(strings.Join(d.cfg.Discovery.Topics, ","))
	}

	for {
		projects, resp, err := d.client.Projects.ListProjects(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for _, p := range projects {
			if d.match(p) {
				allRepos = append(allRepos, p.PathWithNamespace)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (d *GitLabDiscoverer) match(p *gitlab.Project) bool {
	// Owner check (if not handled by API)
	if d.cfg.Discovery.Owner != "" {
		if !strings.HasPrefix(p.PathWithNamespace, d.cfg.Discovery.Owner+"/") {
			return false
		}
	}

	// Regex Include
	if len(d.cfg.Discovery.Includes) > 0 {
		matched := false
		for _, pattern := range d.cfg.Discovery.Includes {
			if matched, _ = regexp.MatchString(pattern, p.Name); matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Regex Exclude
	if len(d.cfg.Discovery.Excludes) > 0 {
		for _, pattern := range d.cfg.Discovery.Excludes {
			if matched, _ := regexp.MatchString(pattern, p.Name); matched {
				return false
			}
		}
	}

	return true
}
