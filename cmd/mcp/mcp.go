package mcp

import (
	"fmt"

	op "github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"

	"github.com/mark3labs/mcp-go/server"
)

const (
	mcpDocShort = "Start an MCP server to control Hue lights from LLMs"
	mcpDocLong  = `
Start a Model Context Protocol (MCP) server that exposes tools for controlling
your Philips Hue lighting system. This allows LLMs to interact with your lights,
rooms, and scenes through standardized MCP tools.

The server communicates via stdio and is designed to be used with MCP-compatible
LLM clients like Claude Desktop, Cursor, or other AI assistants.
`
	mcpDocExample = `
# Start the MCP server
openhue mcp

# Configure in Claude Desktop (claude_desktop_config.json):
{
  "mcpServers": {
    "openhue": {
      "command": "openhue",
      "args": ["mcp"]
    }
  }
}`
)

// NewCmdMcp returns an initialized Command instance for the 'mcp' sub command
func NewCmdMcp(config *openhue.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mcp",
		Short:   mcpDocShort,
		Long:    mcpDocLong,
		Example: mcpDocExample,
		GroupID: "mcp",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMcpServer(config)
		},
	}

	return cmd
}

func runMcpServer(config *openhue.Config) error {
	// Create Hue API client
	h, err := op.NewHome(config.Bridge, config.Key)
	if err != nil {
		return fmt.Errorf("failed to connect to Hue bridge: %w", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"OpenHue",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools
	registerTools(s, h)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
