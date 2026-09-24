package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Load reads one KV v2 secret. In deployment VAULT_TOKEN_FILE is provided by Vault Agent.
func Load(ctx context.Context, path string) (map[string]string, error) {
	addr := strings.TrimRight(os.Getenv("VAULT_ADDR"), "/")
	token := os.Getenv("VAULT_TOKEN")
	if file := os.Getenv("VAULT_TOKEN_FILE"); file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		token = strings.TrimSpace(string(b))
	}
	if addr == "" || token == "" {
		return nil, fmt.Errorf("VAULT_ADDR and Vault token are required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr+"/v1/secret/data/"+strings.TrimPrefix(path, "/"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault read %s: HTTP %d", path, resp.StatusCode)
	}
	var body struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.Data.Data, nil
}
