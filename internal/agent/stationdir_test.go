package agent

// Contract: features/tower/station_mobility_idempotency.feature ("Two shares on one machine do
// not fight over one Station").

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEachShareGetsItsOwnStationDirAndTheFirstKeepsTheMachinesExistingOne(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "tower-station")
	require.Equal(t, legacy, stationDirFor(dir, "box-m-one"), "the first node claims the existing directory")
	require.Equal(t, legacy, stationDirFor(dir, "box-m-one"), "and keeps it")
	other := stationDirFor(dir, "box-m-two")
	require.NotEqual(t, legacy, other, "a second node gets a directory of its own")
	require.Equal(t, other, stationDirFor(dir, "box-m-two"), "stably")
	require.Equal(t, legacy, stationDirFor(dir, ""), "no node id is the legacy directory")
}

func TestTwoSharesStartingTogetherCannotBothClaimTheExistingDir(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	got := make([]string, 8)
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i] = stationDirFor(dir, "node-"+string(rune('a'+i)))
		}(i)
	}
	wg.Wait()
	legacy := filepath.Join(dir, "tower-station")
	n := 0
	for _, d := range got {
		if d == legacy {
			n++
		}
	}
	require.Equal(t, 1, n, "exactly one node claims the existing directory: %v", got)
	raw, err := os.ReadFile(legacy + ".node")
	require.NoError(t, err)
	require.NotEmpty(t, raw)
}
