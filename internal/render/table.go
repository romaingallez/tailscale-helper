package render

import (
	"strings"
	"text/tabwriter"

	"github.com/romaingallez/tailscale-helper/internal/tailscale"
)

func StatusTable(status tailscale.Status) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	_, _ = w.Write([]byte("IP\tHOSTNAME\tTS.NET\tOWNER\tOS\tSTATUS\n"))
	writeNodeRow(w, status.Self)
	for _, peer := range status.Peers {
		writeNodeRow(w, peer)
	}
	_ = w.Flush()

	return b.String()
}

func writeNodeRow(w *tabwriter.Writer, node tailscale.Node) {
	_, _ = w.Write([]byte(strings.Join([]string{
		node.IP,
		node.Hostname,
		node.TSNet,
		node.Owner,
		node.OS,
		node.Status,
	}, "\t") + "\n"))
}
