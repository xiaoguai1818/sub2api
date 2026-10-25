//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Cancelled client reads may be settled using already observed usage; an
// upstream failure with a live client must never be treated as a cancellation.
func TestGeminiStreamAbortedWithBillableOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx, cancel := context.WithCancel(context.Background())
	c.Request = httptest.NewRequest("POST", "/v1beta/models/gemini:test", nil).WithContext(ctx)
	withUsage := &ClaudeUsage{InputTokens: 10, OutputTokens: 4}
	require.False(t, geminiStreamAbortedWithBillableOutput(c, io.ErrUnexpectedEOF, withUsage))
	cancel()
	require.True(t, geminiStreamAbortedWithBillableOutput(c, context.Canceled, withUsage))
	require.False(t, geminiStreamAbortedWithBillableOutput(c, io.EOF, withUsage))
	require.False(t, geminiStreamAbortedWithBillableOutput(c, context.Canceled, &ClaudeUsage{}))
	require.False(t, geminiStreamAbortedWithBillableOutput(nil, errors.New("upstream unavailable"), withUsage))
}
