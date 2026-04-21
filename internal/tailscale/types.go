package tailscale

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

type rawStatus struct {
	Self           rawPeer            `json:"Self"`
	Peer           map[string]rawPeer `json:"Peer"`
	User           map[string]rawUser `json:"User"`
	CurrentTailnet rawTailnet         `json:"CurrentTailnet"`
}

type rawTailnet struct {
	Name           string `json:"Name"`
	MagicDNSSuffix string `json:"MagicDNSSuffix"`
}

type rawUser struct {
	ID          int64  `json:"ID"`
	LoginName   string `json:"LoginName"`
	DisplayName string `json:"DisplayName"`
}

type rawPeer struct {
	ID             string    `json:"ID"`
	HostName       string    `json:"HostName"`
	DNSName        string    `json:"DNSName"`
	OS             string    `json:"OS"`
	UserID         int64     `json:"UserID"`
	TailscaleIPs   []string  `json:"TailscaleIPs"`
	Tags           []string  `json:"Tags"`
	Online         bool      `json:"Online"`
	Active         bool      `json:"Active"`
	ExitNode       bool      `json:"ExitNode"`
	ExitNodeOption bool      `json:"ExitNodeOption"`
	LastSeen       time.Time `json:"LastSeen"`
	RxBytes        int64     `json:"RxBytes"`
	TxBytes        int64     `json:"TxBytes"`
	Expired        bool      `json:"Expired"`
}

type Status struct {
	Tailnet string `json:"tailnet"`
	Self    Node   `json:"self"`
	Peers   []Node `json:"peers"`
}

type Node struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	TSNet        string    `json:"ts_net"`
	IP           string    `json:"ip"`
	OS           string    `json:"os"`
	Owner        string    `json:"owner"`
	Online       bool      `json:"online"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
	ExitNode     bool      `json:"exit_node"`
	ExitNodeHint bool      `json:"exit_node_hint"`
	Active       bool      `json:"active"`
	RxBytes      int64     `json:"rx_bytes"`
	TxBytes      int64     `json:"tx_bytes"`
	Tags         []string  `json:"tags,omitempty"`
	Expired      bool      `json:"expired"`
	Status       string    `json:"status"`
	Self         bool      `json:"self"`
}

func normalizeStatus(raw rawStatus) Status {
	status := Status{
		Tailnet: raw.CurrentTailnet.MagicDNSSuffix,
		Self:    normalizeNode(raw.Self, raw.User, true),
	}

	status.Peers = make([]Node, 0, len(raw.Peer))
	for _, peer := range raw.Peer {
		status.Peers = append(status.Peers, normalizeNode(peer, raw.User, false))
	}

	sort.Slice(status.Peers, func(i, j int) bool {
		left := status.Peers[i]
		right := status.Peers[j]

		if left.Online != right.Online {
			return left.Online
		}

		return strings.ToLower(left.Hostname) < strings.ToLower(right.Hostname)
	})

	return status
}

func normalizeNode(peer rawPeer, users map[string]rawUser, self bool) Node {
	owner := ownerForPeer(peer, users)
	if owner == "" {
		owner = "-"
	}

	node := Node{
		ID:           peer.ID,
		Hostname:     displayHostname(peer),
		TSNet:        strings.TrimSuffix(peer.DNSName, "."),
		IP:           firstIP(peer.TailscaleIPs),
		OS:           peer.OS,
		Owner:        owner,
		Online:       peer.Online,
		LastSeen:     peer.LastSeen,
		ExitNode:     peer.ExitNode,
		ExitNodeHint: peer.ExitNodeOption,
		Active:       peer.Active,
		RxBytes:      peer.RxBytes,
		TxBytes:      peer.TxBytes,
		Tags:         peer.Tags,
		Expired:      peer.Expired,
		Self:         self,
	}

	node.Status = describeStatus(node)

	return node
}

func ownerForPeer(peer rawPeer, users map[string]rawUser) string {
	for _, user := range users {
		if user.ID != peer.UserID {
			continue
		}

		if user.LoginName != "" {
			return user.LoginName
		}

		return user.DisplayName
	}

	if len(peer.Tags) > 0 {
		return strings.Join(peer.Tags, ",")
	}

	return ""
}

func displayHostname(peer rawPeer) string {
	dnsLabel := dnsLabel(peer.DNSName)
	hostname := strings.ToLower(peer.HostName)

	if hostname != "" && hostname != "localhost" {
		return hostname
	}

	if dnsLabel == "" {
		return "-"
	}

	return dnsLabel
}

func dnsLabel(dnsName string) string {
	name := strings.TrimSuffix(dnsName, ".")
	if idx := strings.IndexByte(name, '.'); idx >= 0 {
		return name[:idx]
	}

	return name
}

func firstIP(ips []string) string {
	for _, ip := range ips {
		if strings.Contains(ip, ":") {
			continue
		}
		return ip
	}

	if len(ips) == 0 {
		return "-"
	}

	return ips[0]
}

func describeStatus(node Node) string {
	parts := make([]string, 0, 4)

	if node.Online {
		parts = append(parts, "online")
	} else if !node.LastSeen.IsZero() {
		parts = append(parts, "offline, last seen "+humanDuration(time.Since(node.LastSeen))+" ago")
	} else {
		parts = append(parts, "offline")
	}

	if node.Active && !node.Online {
		parts = append(parts, "active")
	}

	if node.TxBytes > 0 || node.RxBytes > 0 {
		parts = append(parts, "tx "+humanBytes(node.TxBytes)+" rx "+humanBytes(node.RxBytes))
	}

	if node.ExitNode {
		parts = append(parts, "using exit node")
	} else if node.ExitNodeHint {
		parts = append(parts, "offers exit node")
	}

	if node.Expired {
		parts = append(parts, "expired")
	}

	return strings.Join(parts, "; ")
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
