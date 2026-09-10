package openaiadapter

import (
	"encoding/json"
	"testing"

	"google.golang.org/genai"
)

// The wire format is the contract: openai-go tags call_id as omitzero, so an unset
// param.Opt drops the field from the payload entirely and still compiles.
func TestGetRequestContentItemsSerializesFunctionResponseCallID(t *testing.T) {
	content := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{{
			FunctionResponse: &genai.FunctionResponse{
				ID:       "call_abc123",
				Name:     "list_repos",
				Response: map[string]any{"ok": true},
			},
		}},
	}

	items := getRequestContentItems(content)
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	out := items[0].OfFunctionCallOutput
	if out == nil {
		t.Fatal("function response did not map to OfFunctionCallOutput")
	}

	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := payload["call_id"]; got != "call_abc123" {
		t.Errorf("call_id = %v, want %q (payload: %s)", got, "call_abc123", raw)
	}
	if payload["output"] == nil {
		t.Errorf("output missing from payload: %s", raw)
	}
}

func TestGetRequestContentItemsMapsFunctionCallAndText(t *testing.T) {
	content := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{Text: "calling a tool"},
			{FunctionCall: &genai.FunctionCall{
				ID:   "call_xyz789",
				Name: "list_repos",
				Args: map[string]any{"org": "katesclau"},
			}},
		},
	}

	items := getRequestContentItems(content)
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].OfOutputMessage == nil {
		t.Error("model text did not map to OfOutputMessage")
	}
	call := items[1].OfFunctionCall
	if call == nil {
		t.Fatal("function call did not map to OfFunctionCall")
	}
	if call.CallID != "call_xyz789" {
		t.Errorf("CallID = %q, want %q", call.CallID, "call_xyz789")
	}
	if call.Arguments != `{"org":"katesclau"}` {
		t.Errorf("Arguments = %q", call.Arguments)
	}
}
