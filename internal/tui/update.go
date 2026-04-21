package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case statusLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			m.message = "refresh failed"
			return m, nil
		}
		m.status = msg.status
		m.errorText = ""
		m.message = "status refreshed"
		m.lastRefresh = time.Now()
		m.rebuildVisible()
		return m, nil
	case pingFinishedMsg:
		m.pinging = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			m.message = "ping failed"
			return m, nil
		}
		m.errorText = ""
		m.message = "ping completed"
		m.pingOutput = msg.output
		return m, nil
	case tea.KeyMsg:
		if m.filterMode {
			return m.updateFilter(msg)
		}
		return m.updateKeys(msg)
	}

	return m, nil
}

func (m model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filterMode = false
		m.message = "filter mode off"
		return m, nil
	case tea.KeyEnter:
		m.filterMode = false
		m.message = "filter applied"
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.filter) > 0 {
			runes := []rune(m.filter)
			m.filter = string(runes[:len(runes)-1])
			m.rebuildVisible()
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.filter += msg.String()
			m.rebuildVisible()
		}
		return m, nil
	}
}

func (m model) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "j", "down":
		m.selected = min(m.selected+1, max(0, len(m.visible)-1))
		return m, nil
	case "k", "up":
		m.selected = max(m.selected-1, 0)
		return m, nil
	case "g", "G":
		if msg.String() == "g" {
			m.groupMode = nextGroupMode(m.groupMode)
			m.rebuildVisible()
			m.message = "group: " + string(m.groupMode)
			return m, nil
		}
		m.selected = max(0, len(m.visible)-1)
		return m, nil
	case "s":
		m.sortMode = nextSortMode(m.sortMode)
		m.rebuildVisible()
		m.message = "sort: " + string(m.sortMode)
		return m, nil
	case "/":
		m.filterMode = true
		m.message = "filter mode"
		return m, nil
	case "r":
		m.loading = true
		m.message = "refreshing"
		return m, fetchStatusCmd(m.client)
	case "p":
		node, ok := selectedNode(m.visible, m.selected)
		if !ok || m.pinging {
			return m, nil
		}
		m.pinging = true
		m.message = "pinging " + node.Hostname
		m.pingOutput = ""
		return m, pingCmd(m.client, node)
	case "h":
		m.showHelp = !m.showHelp
		return m, nil
	case "esc":
		if m.filter != "" {
			m.filter = ""
			m.rebuildVisible()
			m.message = "filter cleared"
		}
		return m, nil
	case "home":
		m.selected = 0
		return m, nil
	case "end":
		m.selected = max(0, len(m.visible)-1)
		return m, nil
	case "pgdown":
		m.selected = min(m.selected+10, max(0, len(m.visible)-1))
		return m, nil
	case "pgup":
		m.selected = max(m.selected-10, 0)
		return m, nil
	}

	if strings.EqualFold(msg.String(), "?") {
		m.showHelp = !m.showHelp
	}

	return m, nil
}

func nextSortMode(mode sortMode) sortMode {
	order := []sortMode{sortOnline, sortName, sortIP, sortSeen}
	for i, candidate := range order {
		if candidate == mode {
			return order[(i+1)%len(order)]
		}
	}
	return sortOnline
}

func nextGroupMode(mode groupMode) groupMode {
	order := []groupMode{groupNone, groupState, groupOwner, groupOS}
	for i, candidate := range order {
		if candidate == mode {
			return order[(i+1)%len(order)]
		}
	}
	return groupNone
}
