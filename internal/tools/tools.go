package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/corani/mcp-obsidian-go/internal/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Register(srv *server.MCPServer, obs *obsidian.Obsidian) {
	tools := []Tool{
		newCalendarTool(),
		newListFilesInVaultTool(obs),
		newListFilesInDirTool(obs),
		newGetFileContentsTool(obs),
		newGetFileByNameTool(obs),
		newSimpleSearchTool(obs),
		newJsonlogicSearchTool(obs),
		newDataviewSearchTool(obs),
		newPeriodicNoteTool(obs),
		newPeriodicDateTool(obs),
		// newPeriodicRecentTool(obs),
	}

	for _, tool := range tools {
		srv.AddTool(tool.Schema(), tool.Handler)
	}
}

type Tool interface {
	Schema() mcp.Tool
	Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type calendarTool struct{}

func newCalendarTool() Tool {
	return &calendarTool{}
}

func (c *calendarTool) Schema() mcp.Tool {
	return mcp.NewTool("calendar",
		mcp.WithDescription("Returns the current date and time in the format YYYY-MM-DD HH:MM:SS. Use this to find out the current date and time."),
		mcp.WithString("ignore", mcp.Description("ignore this parameter")),
	)
}

func (c *calendarTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	return mcp.NewToolResultText(currentTime), nil
}

type listFilesInVault struct {
	obs *obsidian.Obsidian
}

func newListFilesInVaultTool(obs *obsidian.Obsidian) Tool {
	return &listFilesInVault{
		obs: obs,
	}
}

func (l *listFilesInVault) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_list_files_in_vault",
		mcp.WithDescription("Lists all files and directories in the root directory of your Obsidian vault."),
		mcp.WithString("ignore", mcp.Description("ignore this parameter")),
	)
}

func (l *listFilesInVault) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	files, err := l.obs.ListFilesInVault(ctx)
	if err != nil {
		return toError(err)
	}

	return toJSON(files)
}

type listFilesInDir struct {
	obs *obsidian.Obsidian
}

func newListFilesInDirTool(obs *obsidian.Obsidian) Tool {
	return &listFilesInDir{
		obs: obs,
	}
}

func (l *listFilesInDir) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_list_files_in_dir",
		mcp.WithDescription("Lists all files and directories in a specific directory of your Obsidian vault."),
		mcp.WithString("dirpath",
			mcp.Required(),
			mcp.Description("Path to list files from (relative to your vault root). Note that empty directories will not be returned."),
		),
	)
}

func (l *listFilesInDir) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dirpath := request.GetString("dirpath", "")
	if dirpath == "" {
		return toError(fmt.Errorf("dirpath is required"))
	}

	files, err := l.obs.ListFilesInDir(ctx, dirpath)
	if err != nil {
		return toError(err)
	}

	return toJSON(files)
}

type getFileContents struct {
	obs *obsidian.Obsidian
}

func newGetFileContentsTool(obs *obsidian.Obsidian) Tool {
	return &getFileContents{
		obs: obs,
	}
}

func (g *getFileContents) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_get_file_contents",
		mcp.WithDescription("Retrieves the contents of a file in your Obsidian vault."),
		mcp.WithString("filepath",
			mcp.Required(),
			mcp.Description("Path to the file (relative to your vault root)."),
		),
	)
}

func (g *getFileContents) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filepath := request.GetString("filepath", "")
	if filepath == "" {
		return toError(fmt.Errorf("filepath is required"))
	}

	content, err := g.obs.GetFileContents(ctx, filepath)
	if err != nil {
		return toError(err)
	}

	out, err := json.Marshal(content)
	if err != nil {
		return toError(err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

type getFileByName struct {
	obs *obsidian.Obsidian
}

func newGetFileByNameTool(obs *obsidian.Obsidian) Tool {
	return &getFileByName{
		obs: obs,
	}
}

func (g *getFileByName) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_get_file_by_name",
		mcp.WithDescription("Retrieves the contents of a file in your Obsidian vault by its name. Use this to e.g. resolve `[[filename]]` or `[[filename|alias]]` links in files."),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the file to retrieve (without path)."),
		),
		mcp.WithBoolean("include_content",
			mcp.Description("Whether to include the content of the file (default: false)"),
			mcp.DefaultBool(false),
		),
	)
}

func (g *getFileByName) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filename := request.GetString("filename", "")
	if filename == "" {
		return toError(fmt.Errorf("filename is required"))
	}

	// only use the last part of file and remove the extension
	filename = filepath.Base(filename)
	if ext := filepath.Ext(filename); ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	includeContent := request.GetBool("include_content", false)

	content, err := g.obs.GetFileByName(ctx, filename, includeContent)
	if err != nil {
		return toError(err)
	}

	out, err := json.Marshal(content)
	if err != nil {
		return toError(err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

type simpleSearchTool struct {
	obs *obsidian.Obsidian
}

func newSimpleSearchTool(obs *obsidian.Obsidian) Tool {
	return &simpleSearchTool{
		obs: obs,
	}
}

func (s *simpleSearchTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_simple_search",
		mcp.WithDescription("Simple search for documents matching a specified text query across all files in the vault. Use this tool when you want to do a simple text search"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The text to search for in your vault."),
		),
		mcp.WithNumber("content_length",
			mcp.DefaultNumber(100),
			mcp.Description("How much context to return around the matching string (default: 100)"),
		),
	)
}

func (s *simpleSearchTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	if query == "" {
		return toError(fmt.Errorf("query is required"))
	}

	contentLength := request.GetInt("content_length", 100)
	if contentLength <= 0 {
		return toError(fmt.Errorf("content_length must be greater than 0"))
	}

	results, err := s.obs.SimpleSearch(ctx, query, contentLength)
	if err != nil {
		return toError(err)
	}

	return toJSON(results)
}

type jsonlogicSearchTool struct {
	obs *obsidian.Obsidian
}

func newJsonlogicSearchTool(obs *obsidian.Obsidian) Tool {
	return &jsonlogicSearchTool{
		obs: obs,
	}
}

func (s *jsonlogicSearchTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_jsonlogic_search",
		mcp.WithDescription("Complex search for documents using a JsonLogic query. Supports standard JsonLogic operators plus 'glob' and 'regexp' for pattern matching. Results must be non-falsy. Use this tool when you want to do a complex search, e.g. for all documents with certain tags etc."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("JsonLogic query object. Example: {\"glob\": [\"*.md\", {\"var\": \"path\"}]} matches all markdown files"),
		),
	)
}

func (s *jsonlogicSearchTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	if query == "" {
		return toError(fmt.Errorf("query is required"))
	}

	results, err := s.obs.ComplexSearch(ctx, query, "application/vnd.olrapi.jsonlogic+json")
	if err != nil {
		return toError(err)
	}

	return toJSON(results)
}

type dataviewSearchTool struct {
	obs *obsidian.Obsidian
}

func newDataviewSearchTool(obs *obsidian.Obsidian) Tool {
	return &dataviewSearchTool{
		obs: obs,
	}
}

func (s *dataviewSearchTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_dataview_search",
		mcp.WithDescription("Complex search for documents using a Dataview DQL query. Use this tool when you want to do a complex search, e.g. for all documents with certain tags etc."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Dataview query string. Example: 'table name, path from #tag'"),
		),
	)
}

func (s *dataviewSearchTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	if query == "" {
		return toError(fmt.Errorf("query is required"))
	}

	results, err := s.obs.ComplexSearch(ctx, query, "application/vnd.olrapi.dataview.dql+txt")
	if err != nil {
		return toError(err)
	}

	return toJSON(results)
}

type periodicNoteTool struct {
	obs *obsidian.Obsidian
}

func newPeriodicNoteTool(obs *obsidian.Obsidian) Tool {
	return &periodicNoteTool{
		obs: obs,
	}
}

func (s *periodicNoteTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_get_periodic_note",
		mcp.WithDescription("Get current periodic note for the specified period. Use this to e.g. find out the tasks or calendar for today."),
		mcp.WithString("period",
			mcp.Required(),
			mcp.Description("The period type (daily, weekly, monthly, quarterly, yearly)"),
			mcp.Enum("daily", "weekly", "monthly", "quarterly", "yearly"),
		),
	)
}

func (s *periodicNoteTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := request.GetString("period", "daily")
	if period == "" {
		return toError(fmt.Errorf("period is required"))
	}

	// validate period
	if !slices.Contains([]string{"daily", "weekly", "monthly", "quarterly", "yearly"}, period) {
		return toError(fmt.Errorf("invalid period: %s, must be one of daily, weekly, monthly, quarterly, yearly", period))
	}

	note, err := s.obs.GetPeriodicNote(ctx, period)
	if err != nil {
		return toError(err)
	}

	out, err := json.Marshal(note)
	if err != nil {
		return toError(err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

type periodicDateTool struct {
	obs *obsidian.Obsidian
}

func newPeriodicDateTool(obs *obsidian.Obsidian) Tool {
	return &periodicDateTool{
		obs: obs,
	}
}

func (s *periodicDateTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_get_periodic_date",
		mcp.WithDescription("Get the periodic note for the specified period on the given date."),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("The date for which to get the periodic note (format: YYYY-MM-DD)"),
		),
		mcp.WithString("period",
			mcp.Required(),
			mcp.Description("The period type (daily, weekly, monthly, quarterly, yearly)"),
			mcp.Enum("daily", "weekly", "monthly", "quarterly", "yearly"),
		),
	)
}

func (s *periodicDateTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := request.GetString("period", "daily")
	if period == "" {
		return toError(fmt.Errorf("period is required"))
	}

	date := request.GetString("date", time.Now().Format("2006-01-02"))

	// validate period
	if !slices.Contains([]string{"daily", "weekly", "monthly", "quarterly", "yearly"}, period) {
		return toError(fmt.Errorf("invalid period: %s, must be one of daily, weekly, monthly, quarterly, yearly", period))
	}

	note, err := s.obs.GetPeriodicNoteByDate(ctx, period, date)
	if err != nil {
		return toError(err)
	}

	out, err := json.Marshal(note)
	if err != nil {
		return toError(err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

type periodicRecentTool struct {
	obs *obsidian.Obsidian
}

func newPeriodicRecentTool(obs *obsidian.Obsidian) Tool {
	return &periodicRecentTool{
		obs: obs,
	}
}

func (s *periodicRecentTool) Schema() mcp.Tool {
	return mcp.NewTool("obsidian_get_recent_periodic_note",
		mcp.WithDescription("Get the most recent periodic notes for the specified period."),
		mcp.WithString("period",
			mcp.Required(),
			mcp.Description("The period type (daily, weekly, monthly, quarterly, yearly)"),
			mcp.Enum("daily", "weekly", "monthly", "quarterly", "yearly"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 5)"),
			mcp.DefaultNumber(5),
		),
		mcp.WithBoolean("include_content",
			mcp.Description("Whether to include the content of the periodic note (default: false)"),
			mcp.DefaultBool(false),
		),
	)
}

func (s *periodicRecentTool) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := request.GetString("period", "daily")
	if period == "" {
		return toError(fmt.Errorf("period is required"))
	}

	// validate period
	if !slices.Contains([]string{"daily", "weekly", "monthly", "quarterly", "yearly"}, period) {
		return toError(fmt.Errorf("invalid period: %s, must be one of daily, weekly, monthly, quarterly, yearly", period))
	}

	limit := request.GetInt("limit", 5)

	// validate limit
	if limit <= 0 {
		return toError(fmt.Errorf("limit must be greater than 0"))
	}

	content := request.GetBool("include_content", false)

	note, err := s.obs.GetPeriodicNoteRecent(ctx, period, limit, content)
	if err != nil {
		return toError(err)
	}

	out, err := json.Marshal(note)
	if err != nil {
		return toError(err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

func toError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

func toJSON(v any) (*mcp.CallToolResult, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(out)), nil
}
