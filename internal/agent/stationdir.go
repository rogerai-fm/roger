package agent

// Contract: features/tower/station_mobility_idempotency.feature.
//
// stationDirFor gives each share on a machine its own Station identity ("Two shares on one
// machine do not fight over one Station"). A machine used to keep ONE tower-station directory,
// so every `roger share` on it presented the same assertion key: the second share's attach either
// read as an offer change to the first one's Station, or was refused. Now the first node to use
// the machine's existing directory claims it - so a Station already attached keeps its identity
// across this upgrade - and every other node gets a directory of its own.

import (
	"os"
	"path/filepath"
	"strings"
)

func stationDirFor(dir, nodeID string) string {
	legacy := filepath.Join(dir, "tower-station")
	if nodeID == "" {
		return legacy
	}
	claim := legacy + ".node"
	if raw, err := os.ReadFile(claim); err == nil {
		if strings.TrimSpace(string(raw)) == nodeID {
			return legacy
		}
		return filepath.Join(dir, "tower-station-"+slugify(nodeID))
	}
	// First claim wins, atomically: two shares starting together cannot both take the legacy id.
	if f, err := os.OpenFile(claim, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600); err == nil {
		_, werr := f.WriteString(nodeID + "\n")
		if cerr := f.Close(); werr == nil && cerr == nil {
			return legacy
		}
	} else if raw, rerr := os.ReadFile(claim); rerr == nil && strings.TrimSpace(string(raw)) == nodeID {
		return legacy
	}
	return filepath.Join(dir, "tower-station-"+slugify(nodeID))
}
