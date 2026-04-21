package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

type StoryDiscoveryResult struct {
	Entries []StoryEntry `json:"entries"`
}

type StoryEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Name  string `json:"name"`
}

type storyIndex struct {
	Entries map[string]struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Name  string `json:"name"`
	} `json:"entries"`
}

func StoryDiscovery(ctx context.Context, cfg *config.Config) error {
	if !cfg.Output.WriteJSON && !cfg.Output.WriteMarkdown {
		return nil
	}

	indexURL, err := storyIndexURL(cfg.React.URL)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("storybook index.json request failed: %s", resp.Status)
	}

	var payload storyIndex
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}

	result := StoryDiscoveryResult{}
	for _, entry := range payload.Entries {
		result.Entries = append(result.Entries, StoryEntry{
			ID:    entry.ID,
			Title: entry.Title,
			Name:  entry.Name,
		})
	}

	if err := os.MkdirAll(cfg.Output.Dir, 0o755); err != nil {
		return err
	}

	if cfg.Output.WriteJSON {
		if err := writeJSON(filepath.Join(cfg.Output.Dir, "stories.json"), result); err != nil {
			return err
		}
	}
	if cfg.Output.WriteMarkdown {
		if err := writeStoriesMarkdown(filepath.Join(cfg.Output.Dir, "stories.md"), result); err != nil {
			return err
		}
	}

	return nil
}

func storyIndexURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Path = "/index.json"
	parsed.RawQuery = ""
	return parsed.String(), nil
}

func writeStoriesMarkdown(path string, result StoryDiscoveryResult) error {
	content := "# css-visual-diff Story Discovery Report\n\n"
	content += "| ID | Title | Name |\n"
	content += "| --- | --- | --- |\n"
	for _, e := range result.Entries {
		content += fmt.Sprintf("| %s | %s | %s |\n", e.ID, e.Title, e.Name)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
