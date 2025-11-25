package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	toon "github.com/toon-format/toon-go"
)

// MCP tool handler: get_users -> returns TOON-encoded UsersPayload
func GetUsersTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ struct{}, // no arguments
) (*mcp.CallToolResult, any, error) {
	_ = ctx
	_ = req

	// Generate 100 demo users across 5 roles
	roles := []string{"admin", "user", "viewer", "manager", "guest"}
	cities := []string{"bangalore", "dallas", "mumbai", "seattle", "pune"}

	users := make([]User, 1000)
	for i := 0; i < 1000; i++ {
		roleIdx := rand.Intn(len(roles))
		cityIdx := rand.Intn(len(cities))
		users[i] = User{
			ID:   i + 1,
			Name: fmt.Sprintf("User%d", i+1),
			// Role: roles[i%len(roles)],
			Role: roles[roleIdx],
			// City: cities[i%len(roles)],
			City: cities[cityIdx],
		}
	}

	payload := UsersPayload{
		Users: users,
	}

	// Encode to TOON
	encoded, err := toon.Marshal(payload, toon.WithLengthMarkers(true))
	if err != nil {
		return nil, nil, err
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(encoded), // raw TOON text
			},
		},
	}

	return result, nil, nil
}

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "toon-server",
			Version: "0.2.0",
		},
		nil,
	)

	// Register tool: get_users
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_users",
		Description: "Return a list of demo users encoded as TOON",
	}, GetUsersTool)

	// Expose MCP over stdio
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
