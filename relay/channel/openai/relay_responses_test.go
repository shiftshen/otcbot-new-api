package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestOaiResponsesHandler_ConvertsChatCompletionToResponses(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	chatResponse := dto.OpenAITextResponse{
		Id:      "resp_test_chat_completion",
		Object:  "chat.completion",
		Created: int64(1773574521),
		Model:   "gpt-5.4",
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index: 0,
				Message: dto.Message{
					Role:    "assistant",
					Content: "OK",
				},
				FinishReason: "stop",
			},
		},
		Usage: dto.Usage{
			PromptTokens:     11,
			CompletionTokens: 17,
			TotalTokens:      28,
			CompletionTokenDetails: dto.OutputTokenDetails{
				ReasoningTokens: 10,
			},
		},
	}
	body, err := common.Marshal(chatResponse)
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(body))),
	}

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{},
	}

	usage, newAPIErr := OaiResponsesHandler(c, info, resp)
	require.Nil(t, newAPIErr)
	require.NotNil(t, usage)
	assert.Equal(t, 11, usage.PromptTokens)
	assert.Equal(t, 17, usage.CompletionTokens)

	var out dto.OpenAIResponsesResponse
	err = common.Unmarshal(recorder.Body.Bytes(), &out)
	require.NoError(t, err)
	assert.Equal(t, "response", out.Object)
	assert.Equal(t, "gpt-5.4", out.Model)
	require.Len(t, out.Output, 1)
	assert.Equal(t, "message", out.Output[0].Type)
	assert.Equal(t, "assistant", out.Output[0].Role)
	require.Len(t, out.Output[0].Content, 1)
	assert.Equal(t, "output_text", out.Output[0].Content[0].Type)
	assert.Equal(t, "OK", out.Output[0].Content[0].Text)
	assert.Equal(t, 11, out.Usage.InputTokens)
	assert.Equal(t, 17, out.Usage.OutputTokens)
}

func TestOaiResponsesStreamHandler_ConvertsChatStreamToResponsesEvents(t *testing.T) {
	t.Parallel()

	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		constant.StreamingTimeout = oldTimeout
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	streamBody := strings.Join([]string{
		"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1773574521,\"model\":\"gpt-5.4\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"OK\"},\"finish_reason\":null,\"native_finish_reason\":null}]}",
		"",
		"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1773574521,\"model\":\"gpt-5.4\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\",\"native_finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":11,\"completion_tokens\":16,\"total_tokens\":27,\"completion_tokens_details\":{\"reasoning_tokens\":9}}}",
		"",
		"data: [DONE]",
		"",
	}, "\n")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(streamBody)),
	}

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gpt-5.4",
		},
	}

	usage, newAPIErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, newAPIErr)
	require.NotNil(t, usage)
	assert.Equal(t, 11, usage.PromptTokens)
	assert.Equal(t, 16, usage.CompletionTokens)

	output := recorder.Body.String()
	assert.Contains(t, output, "event: response.created")
	assert.Contains(t, output, "event: response.output_text.delta")
	assert.Contains(t, output, "event: response.output_text.done")
	assert.Contains(t, output, "event: response.completed")
	assert.NotContains(t, output, "chat.completion.chunk")
	assert.Contains(t, output, "\"delta\":\"OK\"")
	assert.Contains(t, output, "\"text\":\"OK\"")
}
