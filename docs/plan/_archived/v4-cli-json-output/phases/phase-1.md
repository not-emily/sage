# Phase 1: JSON Streaming

> **Depends on:** None
> **Enables:** Phase 2 (CLI JSON Coverage)
>
> See: [Full Plan](../plan.md)

## Goal

Add NDJSON streaming support to the `complete` command so hub-core can receive structured data while maintaining real-time streaming UX.

## Key Deliverables

- `--stream` flag for `sage complete`
- When combined with `--json`, outputs NDJSON (one JSON object per line)
- Each chunk includes: content, done flag, and final chunk includes model/usage

## Files to Modify

- `internal/cli/complete.go` — Add --stream flag, implement NDJSON output

## Dependencies

**Internal:** None

**External:** None

## Implementation Notes

**NDJSON output format:**

```jsonl
{"content":"Hello","done":false}
{"content":"!","done":false}
{"content":" How","done":false}
{"content":" can","done":false}
{"content":" I","done":false}
{"content":" help","done":false}
{"content":"?","done":false}
{"content":"","done":true,"model":"gpt-4o","usage":{"prompt_tokens":5,"completion_tokens":7}}
```

**Flag behavior:**
- `sage complete "hi"` — streams plain text (unchanged)
- `sage complete "hi" --json` — buffered JSON (unchanged)
- `sage complete "hi" --json --stream` — NDJSON streaming (new)
- `sage complete "hi" --stream` — same as plain text streaming (--stream without --json is a no-op)

**Implementation approach:**

```go
// In complete.go
stream := fs.Bool("stream", false, "stream output (use with --json for NDJSON)")

// Later in command logic
if *jsonOutput && *stream {
    return completeStreamJSON(client, *profile, req)
} else if *jsonOutput {
    return completeJSON(client, *profile, req)
} else {
    return completeStream(client, *profile, req)
}
```

**completeStreamJSON function:**

```go
func completeStreamJSON(client *sage.Client, profile string, req sage.Request) error {
    chunks, err := client.CompleteStream(profile, req)
    if err != nil {
        return err
    }

    enc := json.NewEncoder(os.Stdout)
    for chunk := range chunks {
        if chunk.Error != nil {
            return chunk.Error
        }

        output := map[string]interface{}{
            "content": chunk.Content,
            "done":    chunk.Done,
        }

        // Include metadata on final chunk
        if chunk.Done {
            output["model"] = chunk.Model
            output["usage"] = map[string]int{
                "prompt_tokens":     chunk.Usage.PromptTokens,
                "completion_tokens": chunk.Usage.CompletionTokens,
            }
        }

        enc.Encode(output)  // Writes with newline
    }

    return nil
}
```

**Note:** Check if `StreamChunk` struct has model/usage fields. If not, may need to track and include on final chunk differently.

## Validation

- [ ] `sage complete "hi" --json --stream` outputs valid NDJSON
- [ ] Each line is valid JSON parseable independently
- [ ] Final chunk has `"done":true` and includes model/usage
- [ ] Plain text streaming still works unchanged
- [ ] Buffered `--json` still works unchanged
