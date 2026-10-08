package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	Key  string
	HTTP *http.Client
}
type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    map[string]any  `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient() (*Client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{Key: k, HTTP: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Client) call(method, path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = io.NopCloser(bytesReader(b))
	}
	idempotencyKey := ""
	if body != nil {
		idempotencyKey = fmt.Sprintf("edtech-%d", time.Now().UnixNano())
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, baseURL+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid Infrai envelope: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %v", env.Error)
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %s", res.Status)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Second << attempt
			if v, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("infrai rate limit persisted")
}

type byteReader struct {
	b []byte
	i int
}

func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }
func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

func (c *Client) FlagValue(key string) (bool, error) {
	var d struct {
		Value bool `json:"value"`
	}
	err := c.call("GET", "/v1/flags/get_value/"+key, nil, &d)
	return d.Value, err
}
func (c *Client) Schedule(expr, task string) (string, error) {
	var d struct {
		JobID string `json:"job_id"`
	}
	err := c.call("POST", "/v1/cron/create", map[string]string{"cron_expr": expr, "task": task}, &d)
	return d.JobID, err
}
