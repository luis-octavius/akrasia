package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// client has a timeout so a hanging request never blocks a command.
var client = &http.Client{Timeout: 15 * time.Second}

// Discord returns a Sender that posts to a channel webhook. Only the
// configured user can be pinged: a task named "@everyone" stays text.
func Discord(cfg Config) Sender {
	return func(ctx context.Context, content string) error {
		mentions := map[string][]string{"parse": {}}
		if cfg.Mention != "" {
			mentions = map[string][]string{"users": {cfg.Mention}}
		}

		body, err := json.Marshal(map[string]any{
			"content":          content,
			"allowed_mentions": mentions,
		})
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Webhook, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			// the error quotes the URL, whose last part is the webhook's secret
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return fmt.Errorf("posting to discord: %w", urlErr.Err)
			}
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return fmt.Errorf("discord answered %s: %s", resp.Status, bytes.TrimSpace(detail))
		}
		return nil
	}
}
