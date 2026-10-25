package service

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// A compatibility stream uses a detached upstream context so the provider can
// deliver final usage after the downstream client leaves. Close an idle drain
// after a finite interval rather than indefinitely holding upstream resources.
const anthropicCompatDisconnectDrainInterval = 180 * time.Second

func watchAnthropicCompatDisconnect(c *gin.Context, body io.Closer) func() {
	if c == nil || c.Request == nil || c.Request.Context().Done() == nil || body == nil {
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-c.Request.Context().Done():
			timer := time.NewTimer(anthropicCompatDisconnectDrainInterval)
			defer timer.Stop()
			select {
			case <-timer.C:
				_ = body.Close()
			case <-done:
			}
		case <-done:
		}
	}()
	return func() { close(done) }
}
