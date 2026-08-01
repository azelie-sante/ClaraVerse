package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HindsightClient talks to a self-hosted Hindsight instance
// (https://github.com/vectorize-io/hindsight) — an open-source,
// MIT-licensed agent memory system (91.4% on LongMemEval, first system
// past 90%) that models memory as retain/recall/reflect over typed facts,
// entities and relationships, retrieved via hybrid semantic+keyword search
// with cross-encoder reranking. This is an evaluated alternative to
// ClaraVerse's native memory pipeline (memory_extraction_service.go +
// memory_selection_service.go), NOT a replacement — see HINDSIGHT_URL
// below. Native memory keeps working unchanged when this isn't set.
//
// Each ClaraVerse user gets their own Hindsight "bank" (bank_id = userID),
// matching Hindsight's own multi-tenant model 1:1 — no extra namespacing
// needed on our side.
//
// Note on encryption: unlike the native pipeline (AES-256-GCM, per-user
// HKDF-derived keys, see crypto/encryption.go), Hindsight stores memory
// text in its own Postgres in plaintext. Since it's self-hosted alongside
// ClaraVerse (same trust boundary as MongoDB itself), this is likely
// acceptable for most deployments, but it IS a real reduction from the
// native system's security posture and should be called out to users
// before enabling.
type HindsightClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHindsightClient wires a client if HINDSIGHT_URL is configured.
// Returns nil (not an error) when unconfigured — callers should treat a
// nil *HindsightClient exactly like "feature not enabled" and skip it.
func NewHindsightClient(baseURL string) *HindsightClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &HindsightClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			// Retain calls run an LLM extraction pass server-side (Hindsight's
			// own retain() call took several seconds against our local model
			// in testing) — generous timeout, this runs off the hot path
			// (background extraction), same as the native pipeline's worker.
			Timeout: 120 * time.Second,
		},
	}
}

// Healthy does a cheap reachability check. Used at startup so a
// misconfigured HINDSIGHT_URL degrades to "feature disabled" with a clear
// log line instead of failing every chat turn silently.
func (h *HindsightClient) Healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Retain stores one piece of text in the user's bank. Hindsight does its
// own fact/entity extraction server-side (this is a single retain call,
// not a batch) — unlike the native pipeline, there's no separate
// extraction-service LLM call to make on our side.
func (h *HindsightClient) Retain(ctx context.Context, userID, content, context_ string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"items": []map[string]string{{"content": content, "context": context_}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/default/banks/%s/memories", h.baseURL, userID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hindsight retain: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("hindsight retain HTTP %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// HindsightMemory is one recall() hit.
type HindsightMemory struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Type     string            `json:"type"` // "world" | "experience" | ...
	Entities []string          `json:"entities"`
	Scores   map[string]float64 `json:"scores"`
}

// Recall runs a hybrid semantic+keyword+reranked search over the user's
// bank. limit is advisory — Hindsight has its own internal top-K before
// reranking; we slice client-side to be sure.
func (h *HindsightClient) Recall(ctx context.Context, userID, query string, limit int) ([]HindsightMemory, error) {
	body, _ := json.Marshal(map[string]interface{}{"query": query})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/default/banks/%s/memories/recall", h.baseURL, userID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hindsight recall: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("hindsight recall HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		Results []HindsightMemory `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("hindsight recall decode: %w", err)
	}
	if limit > 0 && len(parsed.Results) > limit {
		parsed.Results = parsed.Results[:limit]
	}
	return parsed.Results, nil
}
