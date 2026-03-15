package openai

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func OaiResponsesHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}

	var responsesResponse dto.OpenAIResponsesResponse
	if err = common.Unmarshal(responseBody, &responsesResponse); err == nil && responsesResponse.Object == "response" {
		if oaiError := responsesResponse.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
			return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
		}

		if responsesResponse.HasImageGenerationCall() {
			c.Set("image_generation_call", true)
			c.Set("image_generation_call_quality", responsesResponse.GetQuality())
			c.Set("image_generation_call_size", responsesResponse.GetSize())
		}

		service.IOCopyBytesGracefully(c, resp, responseBody)
		usage := dto.Usage{}
		if responsesResponse.Usage != nil {
			usage.PromptTokens = responsesResponse.Usage.InputTokens
			usage.CompletionTokens = responsesResponse.Usage.OutputTokens
			usage.TotalTokens = responsesResponse.Usage.TotalTokens
			if responsesResponse.Usage.InputTokensDetails != nil {
				usage.PromptTokensDetails.CachedTokens = responsesResponse.Usage.InputTokensDetails.CachedTokens
			}
		}
		if info == nil || info.ResponsesUsageInfo == nil || info.ResponsesUsageInfo.BuiltInTools == nil {
			return &usage, nil
		}
		for _, tool := range responsesResponse.Tools {
			buildToolinfo, ok := info.ResponsesUsageInfo.BuiltInTools[common.Interface2String(tool["type"])]
			if !ok || buildToolinfo == nil {
				logger.LogError(c, fmt.Sprintf("BuiltInTools not found for tool type: %v", tool["type"]))
				continue
			}
			buildToolinfo.CallCount++
		}
		return &usage, nil
	}

	var chatResponse dto.OpenAITextResponse
	if err = common.Unmarshal(responseBody, &chatResponse); err == nil && chatResponse.Object == "chat.completion" {
		if oaiError := chatResponse.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
			return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
		}
		responsesCompat, usage, compatErr := service.ChatCompletionsResponseToResponsesResponse(&chatResponse, helper.GetResponseID(c))
		if compatErr != nil {
			return nil, types.NewOpenAIError(compatErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
		convertedBody, marshalErr := common.Marshal(responsesCompat)
		if marshalErr != nil {
			return nil, types.NewOpenAIError(marshalErr, types.ErrorCodeJsonMarshalFailed, http.StatusInternalServerError)
		}
		service.IOCopyBytesGracefully(c, resp, convertedBody)
		return usage, nil
	}

	return nil, types.NewOpenAIError(fmt.Errorf("unsupported /responses response schema"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
}

func OaiResponsesStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		logger.LogError(c, "invalid response or response body")
		return nil, types.NewError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse)
	}

	defer service.CloseResponseBodyGracefully(resp)

	var usage = &dto.Usage{}
	var responseTextBuilder strings.Builder
	responseID := "resp_" + strings.TrimPrefix(helper.GetResponseID(c), "chatcmpl-")
	createAt := time.Now().Unix()
	model := info.UpstreamModelName
	messageID := ""
	responseCreatedSent := false
	responseCompletedSent := false

	sendResponsesEvent := func(eventType string, payload any) bool {
		data, err := common.Marshal(payload)
		if err != nil {
			logger.LogError(c, "failed to marshal responses event: "+err.Error())
			return false
		}
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: eventType}, string(data))
		return true
	}

	sendCreatedIfNeeded := func() bool {
		if responseCreatedSent {
			return true
		}
		if messageID == "" {
			messageID = "msg_" + strings.TrimPrefix(responseID, "resp_")
		}
		payload := map[string]any{
			"type": "response.created",
			"response": map[string]any{
				"id":         responseID,
				"object":     "response",
				"created_at": createAt,
				"status":     "in_progress",
				"model":      model,
				"output":     []any{},
			},
		}
		if !sendResponsesEvent("response.created", payload) {
			return false
		}
		responseCreatedSent = true
		return true
	}

	sendCompleted := func() bool {
		if responseCompletedSent {
			return true
		}
		if !sendCreatedIfNeeded() {
			return false
		}
		fullText := responseTextBuilder.String()
		if fullText != "" {
			if !sendResponsesEvent("response.output_text.done", map[string]any{
				"type":          "response.output_text.done",
				"item_id":       messageID,
				"output_index":  0,
				"content_index": 0,
				"text":          fullText,
			}) {
				return false
			}
		}
		usage.InputTokens = usage.PromptTokens
		usage.OutputTokens = usage.CompletionTokens
		if usage.TotalTokens == 0 {
			usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		}
		payload := map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":         responseID,
				"object":     "response",
				"created_at": createAt,
				"status":     "completed",
				"model":      model,
				"output": []any{
					map[string]any{
						"id":     messageID,
						"type":   "message",
						"status": "completed",
						"role":   "assistant",
						"content": []any{
							map[string]any{
								"type":        "output_text",
								"text":        fullText,
								"annotations": []any{},
							},
						},
					},
				},
				"usage": usage,
			},
		}
		if !sendResponsesEvent("response.completed", payload) {
			return false
		}
		responseCompletedSent = true
		return true
	}

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		var streamResponse dto.ResponsesStreamResponse
		if err := common.UnmarshalJsonStr(data, &streamResponse); err == nil && streamResponse.Type != "" {
			sendResponsesStreamData(c, streamResponse, data)
			switch streamResponse.Type {
			case "response.completed":
				if streamResponse.Response != nil {
					if streamResponse.Response.Usage != nil {
						if streamResponse.Response.Usage.InputTokens != 0 {
							usage.PromptTokens = streamResponse.Response.Usage.InputTokens
						}
						if streamResponse.Response.Usage.OutputTokens != 0 {
							usage.CompletionTokens = streamResponse.Response.Usage.OutputTokens
						}
						if streamResponse.Response.Usage.TotalTokens != 0 {
							usage.TotalTokens = streamResponse.Response.Usage.TotalTokens
						}
						if streamResponse.Response.Usage.InputTokensDetails != nil {
							usage.PromptTokensDetails.CachedTokens = streamResponse.Response.Usage.InputTokensDetails.CachedTokens
						}
					}
					if streamResponse.Response.HasImageGenerationCall() {
						c.Set("image_generation_call", true)
						c.Set("image_generation_call_quality", streamResponse.Response.GetQuality())
						c.Set("image_generation_call_size", streamResponse.Response.GetSize())
					}
				}
			case "response.output_text.delta":
				responseTextBuilder.WriteString(streamResponse.Delta)
			case dto.ResponsesOutputTypeItemDone:
				if streamResponse.Item != nil {
					switch streamResponse.Item.Type {
					case dto.BuildInCallWebSearchCall:
						if info != nil && info.ResponsesUsageInfo != nil && info.ResponsesUsageInfo.BuiltInTools != nil {
							if webSearchTool, exists := info.ResponsesUsageInfo.BuiltInTools[dto.BuildInToolWebSearchPreview]; exists && webSearchTool != nil {
								webSearchTool.CallCount++
							}
						}
					}
				}
			}
			return true
		}

		var chatChunk dto.ChatCompletionsStreamResponse
		if err := common.UnmarshalJsonStr(data, &chatChunk); err != nil || chatChunk.Object != "chat.completion.chunk" {
			if err != nil {
				logger.LogError(c, "failed to unmarshal stream response: "+err.Error())
			}
			return true
		}

		if chatChunk.Id != "" {
			if strings.HasPrefix(chatChunk.Id, "resp_") {
				responseID = chatChunk.Id
			} else {
				responseID = "resp_" + strings.TrimPrefix(chatChunk.Id, "chatcmpl-")
			}
		}
		if chatChunk.Created != 0 {
			createAt = chatChunk.Created
		}
		if chatChunk.Model != "" {
			model = chatChunk.Model
		}
		if messageID == "" {
			messageID = "msg_" + strings.TrimPrefix(responseID, "resp_")
		}
		if !sendCreatedIfNeeded() {
			return false
		}

		for _, choice := range chatChunk.Choices {
			deltaText := choice.Delta.GetContentString()
			if deltaText != "" {
				responseTextBuilder.WriteString(deltaText)
				if !sendResponsesEvent("response.output_text.delta", map[string]any{
					"type":          "response.output_text.delta",
					"item_id":       messageID,
					"output_index":  0,
					"content_index": 0,
					"delta":         deltaText,
				}) {
					return false
				}
			}
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				if chatChunk.Usage != nil {
					usage.PromptTokens = chatChunk.Usage.PromptTokens
					usage.CompletionTokens = chatChunk.Usage.CompletionTokens
					usage.TotalTokens = chatChunk.Usage.TotalTokens
					usage.PromptTokensDetails = chatChunk.Usage.PromptTokensDetails
					usage.CompletionTokenDetails = chatChunk.Usage.CompletionTokenDetails
					usage.InputTokens = chatChunk.Usage.InputTokens
					usage.OutputTokens = chatChunk.Usage.OutputTokens
					usage.InputTokensDetails = chatChunk.Usage.InputTokensDetails
				}
				return sendCompleted()
			}
		}
		return true
	})

	if usage.CompletionTokens == 0 {
		// 计算输出文本的 token 数量
		tempStr := responseTextBuilder.String()
		if len(tempStr) > 0 {
			// 非正常结束，使用输出文本的 token 数量
			completionTokens := service.CountTextToken(tempStr, info.UpstreamModelName)
			usage.CompletionTokens = completionTokens
		}
	}

	if usage.PromptTokens == 0 && usage.CompletionTokens != 0 {
		usage.PromptTokens = info.GetEstimatePromptTokens()
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	if !responseCompletedSent && responseTextBuilder.Len() > 0 {
		_ = sendCompleted()
	}

	return usage, nil
}
