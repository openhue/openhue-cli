package mcp

import (
	"context"
	"fmt"
	"openhue-cli/openhue"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

type HiParams struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

func SayHi(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[HiParams]) (*mcp.CallToolResultFor[any], error) {
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "Hi " + params.Arguments.Name}},
	}, nil
}

func ListRooms(c context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[any]) (*mcp.CallToolResultFor[any], error) {

	ctx, ok := c.Value("openhue").(*openhue.Context)
	if !ok || ctx == nil {
		return nil, fmt.Errorf("could not retrieve openhue context")
	}

	rooms := openhue.SearchRooms(ctx.Home, nil)
	if len(rooms) == 0 {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "I couldn't find any rooms in your Hue home."}},
		}, nil
	}

	var roomNames []string
	for _, room := range rooms {
		roomNames = append(roomNames, room.Name)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Here are the rooms I found: %s.", strings.Join(roomNames, ", ")),
		}},
	}, nil
}

// NewCmdMcpServer returns an initialized Command instance for 'mcp-server' sub command
func NewCmdMcpServer(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "mcp-server",
		Aliases: []string{"mcp"},
		Short:   "Starts the MCP server",
		GroupID: "ai",
		Run: func(cmd *cobra.Command, args []string) {

			// Create a server instance
			server := mcp.NewServer(&mcp.Implementation{Name: "openhue", Version: ctx.BuildInfo.Version}, nil)

			// Add a tool to the server
			mcp.AddTool(server, &mcp.Tool{Name: "room_get", Description: "List all rooms from home"}, ListRooms)

			// Run the server over stdin/stdout, until the client disconnects
			if err := server.Run(context.WithValue(context.Background(), "openhue", ctx), mcp.NewStdioTransport()); err != nil {
				cobra.CheckErr(err)
			}
		},
	}

	return cmd
}
