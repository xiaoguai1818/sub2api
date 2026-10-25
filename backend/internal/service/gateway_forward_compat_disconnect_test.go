//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAnthropicCompatClientDisconnectDrainsFinalUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Final output usage arrives after message_start. Even when a client has
	// canceled the inbound HTTP request, the detached upstream stream must be
	// drained, and no fabricated terminal response sent to the old client.
	sse := "event: message_start\n" +
		"data: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"usage\":{\"input_tokens\":10}}}\n\n" +
		"event: message_delta\n" +
		"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":5}}\n\n" +
		"event: message_stop\n" +
		"data: {\"type\":\"message_stop\"}\n\n"

	for _, adapter := range []string{"chat", "responses"} {
		t.Run(adapter, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			ctx, cancel := context.WithCancel(context.Background())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+adapter, nil).WithContext(ctx)
			cancel()
			resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(sse))}
			svc := &GatewayService{}
			var result *ForwardResult
			var err error
			if adapter == "chat" {
				result, err = svc.handleCCStreamingFromAnthropic(resp, c, "claude-test", "claude-test", nil, time.Now(), false)
			} else {
				result, err = svc.handleResponsesStreamingResponse(resp, c, "claude-test", "claude-test", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Equal(t, 5, result.Usage.OutputTokens)
			require.NotContains(t, rec.Body.String(), "data: [DONE]")
			require.NotContains(t, rec.Body.String(), "response.completed")
		})
	}
}
