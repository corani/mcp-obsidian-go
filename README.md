
# mcp-obsidian-go 🚀

A Go implementation of an **MCP (Model Context Protocol)** server for Obsidian vaults. This project enables advanced AI-powered workflows and integrations with your Obsidian notes.

>[!note]
> Heavily inspired by the amazing [mcp-obsidian](https://github.com/MarkusPfundstein/mcp-obsidian) (Python-based) project by Markus Pfundstein.

## ✨ Features

- 📂 Access and query your Obsidian vault via MCP
- 🧠 AI-powered context and automation
- 🔒 Secure, local-first design
- ⚡ Fast and lightweight Go backend

## 🏁 Getting Started

### Prerequisites

- Go 1.24+
- An Obsidian vault
- Obsidian [Local REST API](https://github.com/coddingtonbear/obsidian-local-rest-api) plugin

### Build & Run

```sh
# Clone the repo
git clone https://github.com/corani/mcp-obsidian-go.git
cd mcp-obsidian-go

# Run
go run ./cmd/mcp-obsidian-go/
```

The server will start and listen for MCP connections on **Stdio**. By default, it also exposes an SSE endpoint at [`http://localhost:8989/mcp`](http://localhost:8989/mcp).

## 🛠️ Implemented Tools

This server implements the following MCP tools:

| Tool Name                      | Description                                                                 |
|---------------------------------|-----------------------------------------------------------------------------|
| `calendar`                     | Returns the current date and time in the format YYYY-MM-DD HH:MM:SS.        |
| `obsidian_list_files_in_vault` | Lists all files and directories in the root directory of your Obsidian vault.|
| `obsidian_list_files_in_dir`   | Lists all files and directories in a specific directory of your vault.       |
| `obsidian_get_file_contents`   | Retrieves the contents of a file in your Obsidian vault.                    |
| `obsidian_get_file_by_name`    | Retrieves the contents of a file by its name (e.g. to resolve `[[filename]]`).|
| `obsidian_simple_search`       | Simple search for documents matching a specified text query.                |
| `obsidian_jsonlogic_search`    | Complex search for documents using a JsonLogic query (advanced filters/tags).|
| `obsidian_dataview_search`     | Complex search for documents using a Dataview DQL query.                    |
| `obsidian_get_periodic_note`   | Get current periodic note for the specified period (daily, weekly, etc).    |
| `obsidian_get_periodic_date`   | Get the periodic note for the specified period on the given date.           |

## 🗂️ Project Structure

```text
cmd/mcp-obsidian-go/                  # Main entrypoint
cmd/mcp-obsidian-go/system-prompt.txt # System prompt for the AI
internal/config/                      # Configuration loading
internal/obsidian/                    # Obsidian integration logic
internal/tools/                       # MCP tool registration
```

## ⚙️ Configuration

Configuration is loaded from environment variables or a `.env` file. See `internal/config` for details.

### .env Example

Create a `.env` file in the project root with the following variables:

```env
OBSIDIAN_API_HOST="http://localhost:27123/"
OBSIDIAN_API_KEY="<your-obsidian-api-key>"
```

These are required for connecting to the Obsidian Local REST API plugin.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
