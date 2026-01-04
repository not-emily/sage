package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/not-emily/sage/pkg/sage"
)

func runComplete(args []string) error {
	fs := flag.NewFlagSet("complete", flag.ExitOnError)

	profile := fs.String("profile", "", "profile to use (default: use default profile)")
	system := fs.String("system", "", "system message")
	maxTokens := fs.Int("max-tokens", 0, "maximum tokens to generate")
	jsonOutput := fs.Bool("json", false, "output JSON instead of streaming")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage complete [flags] [prompt]

Send a completion request to an LLM.

If no prompt is provided, reads from stdin.

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage complete "Hello, world!"
  sage complete --profile=big_brain "Explain quantum computing"
  sage complete --json "What is 2+2?"
  echo "Summarize this" | sage complete
`)
	}

	fs.Parse(args)

	// Get prompt from args or stdin
	prompt := getPrompt(fs.Args())
	if prompt == "" {
		return fmt.Errorf("no prompt provided")
	}

	// Create client
	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	req := sage.Request{
		Prompt:    prompt,
		System:    *system,
		MaxTokens: *maxTokens,
	}

	if *jsonOutput {
		return completeJSON(client, *profile, req)
	}

	return completeStream(client, *profile, req)
}

func completeJSON(client *sage.Client, profile string, req sage.Request) error {
	resp, err := client.Complete(profile, req)
	if err != nil {
		return err
	}

	output := map[string]interface{}{
		"content": resp.Content,
		"model":   resp.Model,
		"usage": map[string]int{
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func completeStream(client *sage.Client, profile string, req sage.Request) error {
	chunks, err := client.CompleteStream(profile, req)
	if err != nil {
		return err
	}

	for chunk := range chunks {
		if chunk.Error != nil {
			return chunk.Error
		}
		if chunk.Done {
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println() // Final newline

	return nil
}

func getPrompt(args []string) string {
	if len(args) > 0 {
		return strings.Join(args, " ")
	}

	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}

	return ""
}
