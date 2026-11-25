package main

// features/tower/tower_upstream_key.feature, "A key too short to redact safely is refused at
// start": a short key is refused BEFORE anything is saved, so a refused key is never remembered
// and replayed into every later start.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAShortUpstreamKeyFlagIsRefusedAndNotSaved(t *testing.T) {
	s := &lanShareBDD{t: t}
	s.reset()
	t.Cleanup(s.after)
	t.Setenv("ROGER_LAN", "1")
	require.NoError(t, s.machineServes("qwen3-8b"))
	require.NoError(t, s.addTower("tower-a", true))
	before := *loadConfig().Share
	require.NoError(t, s.runShare("roger share --upstream-key short7k"))
	require.Error(t, s.err, "a 7-character --upstream-key was accepted:\n%s", s.out)
	require.Contains(t, s.err.Error(), "too short to keep out of replies")
	require.Equal(t, before, *loadConfig().Share, "the refused key was saved to config")
}

func TestTheTUIDoesNotSaveAShortUpstreamKey(t *testing.T) {
	useTempConfig(t)
	c := loadConfig()
	c.Share = &Share{Upstream: "http://127.0.0.1:9/v1", UpstreamKey: "a-long-enough-key"}
	require.NoError(t, saveConfig(c))
	tuiHooks(loadConfig()).SaveUpstream("http://127.0.0.1:10/v1", "short7k")
	got := loadConfig().Share
	require.Equal(t, "http://127.0.0.1:9/v1", got.Upstream, "a short key's endpoint was saved")
	require.Equal(t, "a-long-enough-key", got.UpstreamKey, "a short key was saved")

	tuiHooks(loadConfig()).SaveUpstream("http://127.0.0.1:10/v1", "a-new-long-key")
	require.Equal(t, "a-new-long-key", loadConfig().Share.UpstreamKey, "a long key is still saved")

}
