package tui

// endpoint_hint_test.go: the endpoint plate masks the session key and tells the operator how
// to get the whole one. That hint has to name a command that does it, from wherever the
// plate is drawn (under the band list and in the channel), and the help has to list it.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const hintKey = "sess-0123456789abcdef"

func connectedForHint(md mode) model {
	m := seedFor(120, md, false)
	m.connected = &offer{NodeID: "nyx", Model: "qwen3-4b", Online: true}
	m.endpoint = "http://127.0.0.1:4141/v1"
	m.apikey = hintKey
	return m
}

func TestEndpointPlateHintNamesACommandThatCopiesTheKey(t *testing.T) {
	for _, tc := range []struct {
		name     string
		width    int
		wantHint bool
	}{
		{"wide plate carries the hint", 100, true},
		{"exactly 70 columns carries the hint", 70, true},
		{"69 columns keeps the masked key alone", 69, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := stripANSI(connectedForHint(modeBrowse).endpointBlock(tc.width))
			require.Contains(t, out, maskKey(hintKey))
			require.NotContains(t, out, hintKey, "the full key never sits on the plate")
			// `roger keys` manages account keys (rog-key_...). It never prints this session key.
			require.NotContains(t, out, "roger keys")
			if tc.wantHint {
				require.Contains(t, out, "/endpoint copies the full key")
			} else {
				require.NotContains(t, out, "/endpoint")
			}
		})
	}
}

func TestSlashEndpointCopiesTheExportBlockEverywhereThePlateShows(t *testing.T) {
	t.Run("under the band list", func(t *testing.T) {
		out, cmd := connectedForHint(modeBrowse).run("endpoint")
		status := stripANSI(asModel(out).status)
		require.NotNil(t, cmd, "/endpoint with a channel open writes the clipboard")
		require.Contains(t, status, "Copied to clipboard")
		require.NotContains(t, status, hintKey, "the status line never shows the full key")
	})
	t.Run("in the channel", func(t *testing.T) {
		out, cmd := connectedForHint(modeChat).runSession("/endpoint")
		body := stripANSI(strings.Join(asModel(out).transcript, "\n"))
		require.NotNil(t, cmd, "/endpoint in the channel writes the clipboard")
		require.Contains(t, body, "http://127.0.0.1:4141/v1")
		require.Contains(t, body, "copied to your clipboard")
	})
	t.Run("no channel: nothing to copy", func(t *testing.T) {
		m := seedFor(120, modeBrowse, false)
		m.connected, m.endpoint, m.apikey = nil, "", ""
		out, cmd := m.run("endpoint")
		require.Nil(t, cmd)
		require.Contains(t, stripANSI(asModel(out).status), "tune in first")
	})
}

func TestHelpListsTheEndpointAndKeyPaths(t *testing.T) {
	help := stripANSI(model{}.helpView())
	// The in-channel command line names /connect, the command that prints the snippets.
	var channelLine string
	for _, l := range strings.Split(help, "\n") {
		if strings.Contains(l, "in CHANNEL:") {
			channelLine = l
		}
	}
	require.Contains(t, channelLine, "/connect")
	require.Contains(t, channelLine, "/endpoint")
	// Account keys have their own panel now (keys.go), so the help names the command.
	require.Contains(t, help, "/keys")
	require.NotContains(t, help, "CLI only")
}
