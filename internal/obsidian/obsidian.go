package obsidian

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/corani/mcp-obsidian-go/internal/config"
)

type Obsidian struct {
	conf   *config.Config
	logger *slog.Logger
	client *http.Client
}

func New(conf *config.Config) *Obsidian {
	// TODO(daniel): instead of storing the whole conf, maybe only store the necessary fields?
	return &Obsidian{
		conf:   conf,
		logger: conf.Logger,
		client: &http.Client{
			Transport: newTransport(conf),
		},
	}
}

func (o *Obsidian) ListFilesInVault(ctx context.Context) ([]string, error) {
	path := o.conf.ObsidianAPIHost + "/vault/"

	o.logger.Info("Listing files in vault",
		slog.String("path", path))

	var result struct {
		Files []string `json:"files"`
	}

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return nil, err
	}

	o.logger.Info("Successfully listed files in vault",
		slog.String("path", path),
		slog.Any("result", result))

	return result.Files, nil
}

func (o *Obsidian) ListFilesInDir(ctx context.Context, dir string) ([]string, error) {
	dir = strings.TrimPrefix(dir, "/")
	dir = strings.ReplaceAll(dir, " ", "%20")

	path := o.conf.ObsidianAPIHost + "/vault/" + dir

	o.logger.Info("Listing files in directory",
		slog.String("path", path))

	var result struct {
		Files []string `json:"files"`
	}

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return nil, err
	}

	o.logger.Info("Successfully listed files in directory",
		slog.String("path", path),
		slog.Any("result", result))

	return result.Files, nil
}

type FileContents struct {
	Content     string         `json:"content"`
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
	Path        string         `json:"path,omitempty"`
	Stat        struct {
		CTime int `json:"ctime"`
		MTime int `json:"mtime"`
		Size  int `json:"size"`
	} `json:"stat"`
	Tags []string `json:"tags,omitempty"`
}

func (f FileContents) String() string {
	f.Content = fmt.Sprintf("(%d bytes)", len(f.Content))

	out, err := json.Marshal(f)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

func (o *Obsidian) GetFileContents(ctx context.Context, filepath string) (FileContents, error) {
	filepath = strings.TrimPrefix(filepath, "/")
	filepath = strings.ReplaceAll(filepath, " ", "%20")

	path := o.conf.ObsidianAPIHost + "/vault/" + filepath

	o.logger.Info("Getting file contents",
		slog.String("path", path))

	var result FileContents

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return FileContents{Path: filepath}, err
	}

	o.logger.Info("Successfully retrieved file contents",
		slog.String("path", path),
		slog.String("result", result.String()))

	return result, nil
}

func (o *Obsidian) GetFileByName(ctx context.Context, filename string, includeContent bool) ([]FileContents, error) {
	files, err := o.ComplexSearch(ctx,
		fmt.Sprintf("TABLE WHERE file.name=%q", filename),
		"application/vnd.olrapi.dataview.dql+txt")
	if err != nil {
		return nil, err
	}

	var results []FileContents

	for _, file := range files {
		contents, err := o.GetFileContents(ctx, file.Filename)
		if err != nil {
			return nil, err
		}

		if !includeContent {
			contents.Content = ""
		}

		results = append(results, contents)
	}

	return results, nil
}

type SearchResult struct {
	Filename string  `json:"filename"`
	Score    float64 `json:"score"`
	Matches  []struct {
		Match struct {
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"match"`
		Context string `json:"context"`
	} `json:"matches"`
}

func (o *Obsidian) SimpleSearch(ctx context.Context, query string, length int) ([]SearchResult, error) {
	path := fmt.Sprintf("%s/search/simple/?query=%s&contextLength=%d",
		o.conf.ObsidianAPIHost, url.QueryEscape(query), length)

	o.logger.Info("Searching in vault",
		slog.String("path", path))

	var result []SearchResult

	if err := o.call(ctx, http.MethodPost, path, nil, "", &result); err != nil {
		o.logger.Error("Failed to search in vault",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return nil, err
	}

	o.logger.Info("Successfully searched in vault",
		slog.String("path", path),
		slog.Any("result", result))

	return result, nil
}

type ComplexResult struct {
	Filename string `json:"filename"`
	Result   any
}

func (o *Obsidian) ComplexSearch(ctx context.Context, query string, queryType string) ([]ComplexResult, error) {
	path := fmt.Sprintf("%s/search/", o.conf.ObsidianAPIHost)
	body := strings.NewReader(query)

	o.logger.Info("Searching in vault",
		slog.String("path", path))

	var result []ComplexResult

	if err := o.call(ctx, http.MethodPost, path, body, queryType, &result); err != nil {
		o.logger.Error("Failed to search in vault",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return nil, err
	}

	o.logger.Info("Successfully searched in vault",
		slog.String("path", path),
		slog.Any("result", result))

	return result, nil
}

func (o *Obsidian) GetPeriodicNote(ctx context.Context, period string) (FileContents, error) {
	path := o.conf.ObsidianAPIHost + "/periodic/" + period

	o.logger.Info("Getting periodic note",
		slog.String("path", path))

	var result FileContents

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return FileContents{}, err
	}

	o.logger.Info("Successfully retrieved periodic note",
		slog.String("path", path),
		slog.String("result", result.String()))

	return result, nil
}

func (o *Obsidian) GetPeriodicNoteByDate(ctx context.Context, period, date string) (FileContents, error) {
	path := fmt.Sprintf("%s/periodic/%s/%s", o.conf.ObsidianAPIHost,
		period, strings.ReplaceAll(date, "-", "/"))

	o.logger.Info("Getting periodic note by date",
		slog.String("path", path))

	var result FileContents

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return FileContents{}, err
	}

	o.logger.Info("Successfully retrieved periodic note by date",
		slog.String("path", path),
		slog.String("result", result.String()))

	return result, nil
}

func (o *Obsidian) GetPeriodicNoteRecent(ctx context.Context, period string, limit int, content bool) ([]FileContents, error) {
	path := fmt.Sprintf("%s/periodic/%s/recent?limit=%d&includeContent=%v",
		o.conf.ObsidianAPIHost, period, limit, content)

	o.logger.Info("Getting periodic note",
		slog.String("path", path))

	var result []FileContents

	if err := o.call(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return nil, err
	}

	o.logger.Info("Successfully retrieved periodic note",
		slog.String("path", path),
		slog.Int("results", len(result)))

	return result, nil
}

func (o *Obsidian) call(ctx context.Context, method string, path string, body io.Reader, contentType string, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		o.logger.Error("Failed to create request",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := o.client.Do(req)
	if err != nil {
		o.logger.Error("Failed to execute request",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return err
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		o.logger.Error("Failed to decode response",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return err
	}
	defer res.Body.Close()

	return nil
}
