package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tiktoken-go/tokenizer"
	toon "github.com/toon-format/toon-go"
)

// GPT-4o pricing (approx)
const (
	gpt4oInputUSDPer1MTokens  = 5.0  // $5 / 1M input tokens
	gpt4oOutputUSDPer1MTokens = 15.0 // $15 / 1M output tokens
)

// Return approximate input cost in USD
func approxInputCostUSD(inputTokens int) float64 {
	return float64(inputTokens) / 1_000_000 * gpt4oInputUSDPer1MTokens
}

// Count tokens using GPT-4o tokenizer
func countTokensForMessages(enc tokenizer.Codec, systemMsg, userMsg string) (int, error) {
	sys, err := enc.Count(systemMsg)
	if err != nil {
		return 0, err
	}
	usr, err := enc.Count(userMsg)
	if err != nil {
		return 0, err
	}
	return sys + usr, nil
}

// Fetch TOON via MCP server tool
func fetchUsersViaMCP(ctx context.Context) (UsersPayload, string, error) {
	var empty UsersPayload

	// Start MCP server binary located in ../server
	cmd := exec.Command("../server/server")

	transport := &mcp.CommandTransport{
		Command: cmd,
	}

	// Create MCP client
	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "toon-client",
			Version: "0.3.0",
		},
		nil,
	)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return empty, "", fmt.Errorf("connect MCP: %w", err)
	}
	defer session.Close()

	// Call the MCP tool
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_users",
		Arguments: map[string]any{},
	})
	if err != nil {
		return empty, "", fmt.Errorf("CallTool: %w", err)
	}
	if len(res.Content) == 0 || res.IsError {
		return empty, "", fmt.Errorf("no tool content or error returned")
	}

	// We expect TOON text content
	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		return empty, "", fmt.Errorf("unexpected content type %T", res.Content[0])
	}

	toonText := textContent.Text

	// Decode TOON to struct
	var payload UsersPayload
	if err := toon.Unmarshal([]byte(toonText), &payload); err != nil {
		return empty, "", fmt.Errorf("TOON decode failed: %w", err)
	}
	return payload, toonText, nil
}

// ------------------------------------------

func main() {
	ctx := context.Background()

	payload, toonText, err := fetchUsersViaMCP(ctx)
	if err != nil {
		log.Fatalf("MCP fetch failed: %v", err)
	}

	// Also create JSON form
	jsonBytes, _ := json.Marshal(payload)
	jsonText := string(jsonBytes)

	// Prompt
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter your question:")
	question, _ := reader.ReadString('\n')
	question = strings.TrimSpace(question)

	// Prompt variants
	systemMsg := "You answer only based on provided user data."

	userMsgTOON := fmt.Sprintf(
		"Users in TOON:\n%s\n\nQuestion: %s", toonText, question,
	)

	userMsgJSON := fmt.Sprintf(
		"Users in JSON:\n%s\n\nQuestion: %s", jsonText, question,
	)

	fmt.Println("JSON Payload length ", len(jsonText))
	fmt.Println("TOON Payload length ", len(toonText))

	// GPT-4o tokenization
	enc, err := tokenizer.ForModel(tokenizer.GPT4o)
	if err != nil {
		log.Fatalf("Tokenizer load failed: %v", err)
	}

	toonTokens, _ := countTokensForMessages(enc, systemMsg, userMsgTOON)
	jsonTokens, _ := countTokensForMessages(enc, systemMsg, userMsgJSON)

	toonCost := approxInputCostUSD(toonTokens)
	jsonCost := approxInputCostUSD(jsonTokens)

	fmt.Println("\n--- Token & Cost Analysis (GPT-4o) ---")
	fmt.Printf("TOON: %d tokens  →  $%.6f\n", toonTokens, toonCost)
	fmt.Printf("JSON: %d tokens  →  $%.6f\n", jsonTokens, jsonCost)
	if jsonTokens > 0 {
		fmt.Printf("Savings: %d tokens (%.2f%%)\n",
			jsonTokens-toonTokens,
			(float64(jsonTokens-toonTokens)/float64(jsonTokens))*100.0,
		)
	}
	fmt.Println("--------------------------------------")

	// 🔥 Real GPT-4o call using **TOON**
	oa := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	resp, err := oa.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMsg),
			openai.UserMessage(userMsgTOON),
		},
	})
	if err != nil {
		log.Fatalf("OpenAI API error: %v", err)
	}

	answer := resp.Choices[0].Message.Content

	fmt.Println("\n=== GPT-4o Answer ===")
	fmt.Println(answer)
	fmt.Println("=====================")
}
