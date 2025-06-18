package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/corani/mcp-obsidian-go/internal/config"
	"github.com/corani/mcp-obsidian-go/internal/obsidian"
	"github.com/corani/mcp-obsidian-go/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed system-prompt.txt
var INSTRUCTIONS string

func main() {
	logfile, err := os.OpenFile("mcpserver.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	out := io.MultiWriter(logfile, os.Stdout)

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{})
	logger := slog.New(handler)

	conf := config.MustLoad(logger)

	hooks := new(server.Hooks)

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		logger.Error("Error in MCP method",
			slog.String("method", string(method)),
			slog.Any("id", id),
			slog.Any("message", message),
			slog.String("error", err.Error()),
		)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		logger.Info("Success in MCP method",
			slog.String("method", string(method)),
			slog.Any("id", id),
			slog.Any("message", message),
		)
	})

	obs := obsidian.New(conf)

	instructions := INSTRUCTIONS +
		fmt.Sprintf("\n\nThe current date is: %v", time.Now().Format("2006-01-02"))

	srv := server.NewMCPServer(
		"mcp-obsidian-go", "1.0.0",
		server.WithLogging(),
		server.WithInstructions(instructions),
		server.WithRecovery(),
		server.WithHooks(hooks),
	)

	tools.Register(srv, obs)

	// TODO(daniel): probably shouldn't use a lambda here, and we should check the request params.
	srv.AddPrompt(mcp.NewPrompt("instructions"),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult("instructions", []mcp.PromptMessage{
				mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(instructions)),
			}), nil
		})

	// TODO(daniel): probably shouldn't use a lambda here, and we should check the request params.
	srv.AddResource(mcp.NewResource("file:///mcpserver.log", "server log"),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			logfile.Sync()

			// TODO(daniel): reading a file that's open for writing is a bad idea, but this is just a demo.
			bs, err := os.ReadFile(logfile.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read log file: %w", err)
			}

			contents := mcp.TextResourceContents{
				Text:     string(bs),
				URI:      request.Params.URI,
				MIMEType: "text/plain",
			}

			return []mcp.ResourceContents{contents}, nil
		})

	sse := server.NewSSEServer(srv,
		server.WithSSEEndpoint("/mcp"),
	)
	defer sse.Shutdown(context.Background())

	go func() {
		logger.Info("Starting SSE server",
			slog.String("address", "http://localhost:8989/mcp"))

		if err := sse.Start("0.0.0.0:8989"); err != nil {
			logger.Error("Failed to start SSE server",
				slog.String("error", err.Error()),
			)
		}
	}()

	if err := server.ServeStdio(srv); err != nil {
		logger.Error("Failed to serve MCP server",
			slog.String("error", err.Error()),
		)
	}
}
