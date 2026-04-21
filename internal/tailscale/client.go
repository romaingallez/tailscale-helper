package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Client struct {
	binary string
}

func NewClient(binary string) Client {
	return Client{binary: binary}
}

func (c Client) Status(ctx context.Context) (Status, error) {
	cmd := exec.CommandContext(ctx, c.binary, "status", "--json")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Status{}, fmt.Errorf("tailscale status failed: %s", string(exitErr.Stderr))
		}
		return Status{}, fmt.Errorf("run tailscale status: %w", err)
	}

	var raw rawStatus
	if err := json.Unmarshal(out, &raw); err != nil {
		return Status{}, fmt.Errorf("decode tailscale status: %w", err)
	}

	return normalizeStatus(raw), nil
}

func (c Client) Ping(ctx context.Context, target string) (string, error) {
	cmd := exec.CommandContext(ctx, c.binary, "ping", "--c=1", target)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text == "" {
			return "", fmt.Errorf("tailscale ping failed: %w", err)
		}
		return "", fmt.Errorf("tailscale ping failed: %s", text)
	}

	return text, nil
}

func (s Status) Nodes() []Node {
	nodes := make([]Node, 0, len(s.Peers)+1)
	nodes = append(nodes, s.Self)
	nodes = append(nodes, s.Peers...)
	return nodes
}
