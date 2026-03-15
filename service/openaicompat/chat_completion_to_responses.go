package openaicompat

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func ChatCompletionsResponseToResponsesResponse(resp *dto.OpenAITextResponse, fallbackID string) (*dto.OpenAIResponsesResponse, *dto.Usage, error) {
	if resp == nil {
		return nil, nil, errors.New("response is nil")
	}
	if len(resp.Choices) == 0 {
		return nil, nil, errors.New("chat completion response has no choices")
	}

	createdAt, err := normalizeCreatedAt(resp.Created)
	if err != nil {
		return nil, nil, err
	}

	responseID := normalizeResponsesID(strings.TrimSpace(resp.Id), fallbackID)
	messageID := buildResponsesMessageID(responseID)
	usage := normalizeResponsesUsage(resp.Usage)

	choice := resp.Choices[0]
	output := make([]dto.ResponsesOutput, 0, 1)
	text := choice.Message.StringContent()
	if text != "" {
		output = append(output, dto.ResponsesOutput{
			ID:     messageID,
			Type:   "message",
			Status: "completed",
			Role:   "assistant",
			Content: []dto.ResponsesOutputContent{
				{
					Type:        "output_text",
					Text:        text,
					Annotations: []interface{}{},
				},
			},
		})
	}

	for _, toolCall := range choice.Message.ParseToolCalls() {
		callID := strings.TrimSpace(toolCall.ID)
		if callID == "" {
			callID = buildResponsesCallID(responseID, len(output))
		}
		output = append(output, dto.ResponsesOutput{
			ID:        callID,
			Type:      "function_call",
			Status:    "completed",
			CallId:    callID,
			Name:      toolCall.Function.Name,
			Arguments: common.Interface2String(toolCall.Function.Arguments),
		})
	}

	status := json.RawMessage(`"completed"`)
	toolChoice := json.RawMessage(`"auto"`)
	previousResponseID := json.RawMessage(`null`)
	truncation := json.RawMessage(`"disabled"`)
	user := json.RawMessage(`null`)
	metadata := json.RawMessage(`null`)

	out := &dto.OpenAIResponsesResponse{
		ID:                 responseID,
		Object:             "response",
		CreatedAt:          createdAt,
		Status:             status,
		Model:              resp.Model,
		Output:             output,
		ParallelToolCalls:  true,
		PreviousResponseID: previousResponseID,
		Store:              false,
		ToolChoice:         toolChoice,
		Tools:              []map[string]any{},
		Truncation:         truncation,
		Usage:              usage,
		User:               user,
		Metadata:           metadata,
	}

	return out, usage, nil
}

func normalizeCreatedAt(created any) (int, error) {
	switch v := created.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case json.Number:
		i, err := strconv.Atoi(v.String())
		if err != nil {
			return 0, fmt.Errorf("invalid created value: %w", err)
		}
		return i, nil
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("invalid created value: %w", err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported created value type: %T", created)
	}
}

func normalizeResponsesID(responseID string, fallbackID string) string {
	if responseID != "" && strings.HasPrefix(responseID, "resp_") {
		return responseID
	}
	if responseID != "" {
		return "resp_" + strings.TrimPrefix(responseID, "chatcmpl-")
	}
	if fallbackID != "" {
		if strings.HasPrefix(fallbackID, "resp_") {
			return fallbackID
		}
		return "resp_" + strings.TrimPrefix(fallbackID, "chatcmpl-")
	}
	return "resp_local"
}

func buildResponsesMessageID(responseID string) string {
	suffix := strings.TrimPrefix(responseID, "resp_")
	if suffix == "" {
		suffix = "local"
	}
	return "msg_" + suffix
}

func buildResponsesCallID(responseID string, index int) string {
	suffix := strings.TrimPrefix(responseID, "resp_")
	if suffix == "" {
		suffix = "local"
	}
	return fmt.Sprintf("call_%s_%d", suffix, index)
}

func normalizeResponsesUsage(usage dto.Usage) *dto.Usage {
	out := usage
	if out.InputTokens == 0 {
		out.InputTokens = out.PromptTokens
	}
	if out.OutputTokens == 0 {
		out.OutputTokens = out.CompletionTokens
	}
	if out.TotalTokens == 0 {
		out.TotalTokens = out.InputTokens + out.OutputTokens
	}
	if out.InputTokensDetails == nil {
		out.InputTokensDetails = &dto.InputTokenDetails{
			CachedTokens: out.PromptTokensDetails.CachedTokens,
			TextTokens:   out.InputTokens,
		}
	}
	if out.CompletionTokenDetails.TextTokens == 0 {
		out.CompletionTokenDetails.TextTokens = out.OutputTokens
	}
	return &out
}
