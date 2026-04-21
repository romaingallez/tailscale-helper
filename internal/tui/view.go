package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"tailscale_helper/internal/tailscale"
)

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	contentWidth := m.width
	contentHeight := m.height - 4
	if m.showHelp {
		contentHeight--
	}
	contentHeight -= 2
	if contentHeight < 3 {
		contentHeight = 3
	}

	leftWidth := max(24, contentWidth/3)
	rightWidth := max(24, contentWidth-leftWidth)
	if leftWidth+rightWidth > contentWidth {
		rightWidth = contentWidth - leftWidth
	}
	if rightWidth < 1 {
		rightWidth = 1
	}

	title := m.renderTitle(contentWidth)
	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderListPane(leftWidth, contentHeight),
		m.renderDetailsPane(rightWidth, contentHeight),
	)
	status := m.renderStatusBar(contentWidth)

	parts := []string{title, panes, status}
	if m.showHelp {
		parts = append(parts, m.renderHelp(contentWidth))
	}

	return strings.Join(parts, "\n")
}

func (m model) renderTitle(width int) string {
	tailnet := m.status.Tailnet
	if tailnet == "" {
		tailnet = m.tailnetLabel
	}
	stats := fmt.Sprintf("Tailnet: %s | Nodes: %d | Filter: %s | Sort: %s | Group: %s",
		tailnet,
		len(m.visible),
		displayFilter(m.filter),
		m.sortMode,
		m.groupMode,
	)
	return m.styles.Title.Width(width).Render(truncate(stats, width-2))
}

func (m model) renderListPane(width, height int) string {
	innerWidth := max(1, width-4)
	innerHeight := max(1, height-2)
	lines := []string{m.styles.PanelTitle.Render(truncate("Nodes", innerWidth))}
	currentGroup := ""
	remaining := innerHeight - 1

	if len(m.visible) == 0 {
		lines = append(lines, m.styles.Muted.Render(truncate("No nodes match the current filter.", innerWidth)))
		remaining--
	}

	for i, item := range m.visible {
		if remaining <= 0 {
			break
		}
		if item.group != currentGroup {
			currentGroup = item.group
			lines = append(lines, m.styles.Group.Render(truncate(currentGroup, innerWidth)))
			remaining--
			if remaining <= 0 {
				break
			}
		}
		lines = append(lines, truncate(m.listLine(i, item, innerWidth), innerWidth))
		remaining--
	}

	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	return m.styles.ListPanel.Width(width).Render(strings.Join(lines[:innerHeight], "\n"))
}

func (m model) listLine(i int, item nodeItem, width int) string {
	marker := "  "
	style := m.styles.ListItem
	if i == m.selected {
		marker = "> "
		style = m.styles.ListSelected
	}
	state := "off"
	if item.node.Online {
		state = "on "
	}
	name := item.node.Hostname
	if item.self {
		name += " (self)"
	}
	text := fmt.Sprintf("%s%-3s %-18s %s", marker, state, truncate(name, 18), item.node.IP)
	return style.Render(truncate(text, width))
}

func (m model) renderDetailsPane(width, height int) string {
	innerWidth := max(1, width-4)
	innerHeight := max(1, height-2)
	lines := []string{m.styles.PanelTitle.Render(truncate("Details", innerWidth))}
	node, ok := selectedNode(m.visible, m.selected)
	if ok {
		lines = append(lines, m.detailLines(node, innerWidth)...)
	} else {
		lines = append(lines, m.styles.Muted.Render("No node selected."))
	}

	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	return m.styles.DetailsPanel.Width(width).Render(strings.Join(lines[:innerHeight], "\n"))
}

func (m model) detailLines(node tailscale.Node, width int) []string {
	lines := []string{
		m.kv("Hostname", node.Hostname, width),
		m.kv("TS.NET", node.TSNet, width),
		m.kv("IP", node.IP, width),
		m.kv("Owner", node.Owner, width),
		m.kv("OS", node.OS, width),
		m.kv("Status", node.Status, width),
		m.kv("Last Seen", formatLastSeen(node), width),
		m.kv("Tags", displayTags(node.Tags), width),
		m.kv("Exit", displayExit(node), width),
		m.kv("Traffic", fmt.Sprintf("tx %s rx %s", humanBytes(node.TxBytes), humanBytes(node.RxBytes)), width),
		"",
		m.styles.Section.Render("Action Output"),
	}

	actionText := m.pingOutput
	if m.pinging {
		actionText = "Pinging selected node..."
	}
	for _, line := range wrapBlock(actionText, width) {
		lines = append(lines, truncate(line, width))
	}
	return lines
}

func (m model) kv(label, value string, width int) string {
	text := fmt.Sprintf("%-10s %s", label+":", value)
	return truncate(text, width)
}

func (m model) renderStatusBar(width int) string {
	parts := []string{}
	if m.loading {
		parts = append(parts, "loading")
	}
	if m.errorText != "" {
		parts = append(parts, "error: "+m.errorText)
	} else if m.message != "" {
		parts = append(parts, m.message)
	}
	if !m.lastRefresh.IsZero() {
		parts = append(parts, "updated "+humanSince(m.lastRefresh)+" ago")
	}
	line := strings.Join(parts, " | ")
	if line == "" {
		line = "ready"
	}
	return m.styles.Status.Width(width).Render(truncate(line, width-2))
}

func (m model) renderHelp(width int) string {
	help := "/ filter  s sort  g group  r refresh  p ping  esc clear  q quit  ? help"
	if m.filterMode {
		help = "Type to filter, Enter to apply, Esc to exit filter mode"
	}
	return m.styles.Help.Width(width).Render(truncate(help, width-2))
}

func displayFilter(filter string) string {
	if strings.TrimSpace(filter) == "" {
		return "all"
	}
	return filter
}

func groupForNode(node tailscale.Node, mode groupMode) string {
	switch mode {
	case groupState:
		if node.Online {
			return "Online"
		}
		return "Offline"
	case groupOwner:
		return node.Owner
	case groupOS:
		return strings.ToUpper(node.OS)
	default:
		return "All Nodes"
	}
}

func sortNodes(nodes []tailscale.Node, mode sortMode) {
	sort.Slice(nodes, func(i, j int) bool {
		left := nodes[i]
		right := nodes[j]
		switch mode {
		case sortName:
			return strings.ToLower(left.Hostname) < strings.ToLower(right.Hostname)
		case sortIP:
			return left.IP < right.IP
		case sortSeen:
			if left.LastSeen.Equal(right.LastSeen) {
				return strings.ToLower(left.Hostname) < strings.ToLower(right.Hostname)
			}
			return left.LastSeen.After(right.LastSeen)
		default:
			if left.Online != right.Online {
				return left.Online
			}
			return strings.ToLower(left.Hostname) < strings.ToLower(right.Hostname)
		}
	})
}

func displayTags(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ", ")
}

func displayExit(node tailscale.Node) string {
	switch {
	case node.ExitNode:
		return "using exit node"
	case node.ExitNodeHint:
		return "offers exit node"
	default:
		return "-"
	}
}

func formatLastSeen(node tailscale.Node) string {
	if node.Online {
		return "online now"
	}
	if node.LastSeen.IsZero() {
		return "-"
	}
	return humanSince(node.LastSeen) + " ago"
}

func humanSince(t time.Time) string {
	if t.IsZero() {
		return "0s"
	}
	return humanDuration(time.Since(t))
}

func wrapBlock(text string, width int) []string {
	if strings.TrimSpace(text) == "" {
		return []string{"-"}
	}
	lines := []string{}
	for _, line := range strings.Split(text, "\n") {
		runes := []rune(line)
		for lipgloss.Width(string(runes)) > width && width > 1 {
			chunkLen := min(len(runes), width-1)
			lines = append(lines, string(runes[:chunkLen]))
			runes = runes[chunkLen:]
		}
		lines = append(lines, string(runes))
	}
	return lines
}

func humanDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d >= 24*time.Hour:
		return strconv.Itoa(int(d/(24*time.Hour))) + "d"
	case d >= time.Hour:
		return strconv.Itoa(int(d/time.Hour)) + "h"
	case d >= time.Minute:
		return strconv.Itoa(int(d/time.Minute)) + "m"
	default:
		return strconv.Itoa(int(d/time.Second)) + "s"
	}
}

func humanBytes(v int64) string {
	const unit = 1024
	if v < unit {
		return strconv.FormatInt(v, 10) + "B"
	}
	div, exp := int64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(v) / float64(div)
	text := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 1, 64), "0"), ".")
	return text + string("KMGTPE"[exp]) + "iB"
}
