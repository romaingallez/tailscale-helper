package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tailscale_helper/internal/tailscale"
)

type Client interface {
	Status(context.Context) (tailscale.Status, error)
	Ping(context.Context, string) (string, error)
}

type sortMode string

const (
	sortOnline sortMode = "online,name"
	sortName   sortMode = "name"
	sortIP     sortMode = "ip"
	sortSeen   sortMode = "last-seen"
)

type groupMode string

const (
	groupNone  groupMode = "none"
	groupState groupMode = "state"
	groupOwner groupMode = "owner"
	groupOS    groupMode = "os"
)

type nodeItem struct {
	node  tailscale.Node
	group string
	self  bool
}

type statusLoadedMsg struct {
	status tailscale.Status
	err    error
}

type pingFinishedMsg struct {
	output string
	err    error
}

type model struct {
	client       Client
	styles       styles
	width        int
	height       int
	status       tailscale.Status
	visible      []nodeItem
	selected     int
	filter       string
	filterMode   bool
	sortMode     sortMode
	groupMode    groupMode
	loading      bool
	pinging      bool
	errorText    string
	message      string
	pingOutput   string
	lastRefresh  time.Time
	quitting     bool
	showHelp     bool
	tailnetLabel string
}

func Run(ctx context.Context, client Client) error {
	p := tea.NewProgram(newModel(client), tea.WithContext(ctx), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(client Client) model {
	m := model{
		client:     client,
		styles:     defaultStyles(),
		sortMode:   sortOnline,
		groupMode:  groupNone,
		showHelp:   true,
		selected:   0,
		loading:    true,
		message:    "loading tailscale status",
		pingOutput: "Press p to ping the selected node.",
	}
	m.rebuildVisible()
	return m
}

func (m model) Init() tea.Cmd {
	return fetchStatusCmd(m.client)
}

func fetchStatusCmd(client Client) tea.Cmd {
	return func() tea.Msg {
		status, err := client.Status(context.Background())
		return statusLoadedMsg{status: status, err: err}
	}
}

func pingCmd(client Client, node tailscale.Node) tea.Cmd {
	target := node.TSNet
	if target == "" {
		target = node.IP
	}

	return func() tea.Msg {
		output, err := client.Ping(context.Background(), target)
		return pingFinishedMsg{output: output, err: err}
	}
}

func (m *model) rebuildVisible() {
	nodes := m.status.Nodes()
	filtered := make([]tailscale.Node, 0, len(nodes))
	query := strings.ToLower(strings.TrimSpace(m.filter))

	for _, node := range nodes {
		if query != "" && !matchesFilter(node, query) {
			continue
		}
		filtered = append(filtered, node)
	}

	sortNodes(filtered, m.sortMode)

	selectedID := ""
	if m.selected >= 0 && m.selected < len(m.visible) {
		selectedID = m.visible[m.selected].node.ID
	}

	m.visible = buildVisible(filtered, m.groupMode)
	m.selected = 0
	if selectedID != "" {
		for i, item := range m.visible {
			if item.node.ID == selectedID {
				m.selected = i
				break
			}
		}
	}
	if m.selected >= len(m.visible) {
		m.selected = max(0, len(m.visible)-1)
		if len(m.visible) == 0 {
			m.selected = 0
		}
	}
	if m.tailnetLabel == "" {
		m.tailnetLabel = m.status.Tailnet
	}
}

func buildVisible(nodes []tailscale.Node, mode groupMode) []nodeItem {
	items := make([]nodeItem, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, nodeItem{node: node, group: groupForNode(node, mode), self: node.Self})
	}
	return items
}

func selectedNode(items []nodeItem, idx int) (tailscale.Node, bool) {
	if idx < 0 || idx >= len(items) {
		return tailscale.Node{}, false
	}
	return items[idx].node, true
}

func matchesFilter(node tailscale.Node, query string) bool {
	haystacks := []string{
		node.Hostname,
		node.TSNet,
		node.IP,
		node.Owner,
		node.OS,
		node.Status,
		strings.Join(node.Tags, " "),
	}
	for _, text := range haystacks {
		if strings.Contains(strings.ToLower(text), query) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	cut := width - 1
	if cut > len(runes) {
		cut = len(runes)
	}
	return string(runes[:cut]) + "…"
}
