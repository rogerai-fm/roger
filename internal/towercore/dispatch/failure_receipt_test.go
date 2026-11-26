package dispatch

// Contract: features/tower/tower_failed_attempt_release.feature.

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/towerobj"
)

func TestAFailureReceiptClaimsNothingAndParses(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	g := Grant{AttemptID: "att-1", StationID: "st-1"}
	rec, err := SignFailureReceipt(priv, "roger-public", g, []byte("req"), FailureThrottled)
	require.NoError(t, err)
	require.True(t, IsFailureReceipt(rec.Signed))

	got, err := ParseReceipt(rec.Signed, pub, "roger-public", "att-1", "st-1")
	require.NoError(t, err)
	require.Equal(t, FailureThrottled, got.Failure)
	require.Equal(t, Usage{}, got.Usage)
	require.Equal(t, Usage{}, got.TokUsage)
	require.Equal(t, digestOf(nil), got.ResponseDigest)

	_, err = SignFailureReceipt(priv, "roger-public", g, nil, "")
	require.Error(t, err, "a failure receipt names its class")
}

func TestAServedReceiptIsNotAFailureReceiptAndKeepsItsBytes(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	rec, err := SignReceipt(priv, "roger-public", Grant{AttemptID: "a", StationID: "s"}, []byte("q"), []byte("answer"),
		Usage{In: 1, Out: 6}, Usage{})
	require.NoError(t, err)
	require.False(t, IsFailureReceipt(rec.Signed))
	require.NotContains(t, string(rec.Signed), `"failure"`, "a served receipt's signed bytes are unchanged")
	require.False(t, IsFailureReceipt(nil))
	require.False(t, IsFailureReceipt([]byte("not json")))
}

func TestAFailureReceiptThatClaimsUsageOrBytesIsRefused(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sign := func(fields map[string]any) []byte {
		base := map[string]any{
			"network": "roger-public", "type": TypeReceipt, "version": towerobj.FormatInt(Version),
			"attempt_id": "att-1", "station_id": "st-1", "request_digest": digestOf([]byte("q")),
			"response_digest": digestOf(nil), "usage_in": "0", "usage_out": "0", "tok_in": "0", "tok_out": "0",
			"failure": FailureUpstreamError,
		}
		for k, v := range fields {
			base[k] = v
		}
		raw, err := json.Marshal(base)
		require.NoError(t, err)
		signed, err := towerobj.Sign(priv, "roger-public", TypeReceipt, Version, raw, "station_sig")
		require.NoError(t, err)
		return signed
	}
	for name, f := range map[string]map[string]any{
		"usage out":      {"usage_out": "5"},
		"usage in":       {"usage_in": "5"},
		"token out":      {"tok_out": "5"},
		"response bytes": {"response_digest": digestOf([]byte("x"))},
	} {
		raw := sign(f)
		_, err := ParseReceipt(raw, pub, "roger-public", "att-1", "st-1")
		require.Error(t, err, name)
		require.False(t, IsFailureReceipt(raw) && name != "response bytes", name)
	}
}

// Only the four classes the spec names may be signed into a failure receipt: an unknown or
// over-long class is a receipt nobody can act on, and refused.
func TestAnUnknownOrOverlongFailureClassIsRefused(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	for _, class := range []string{"the station had a bad day", string(make([]byte, 300)), "UPSTREAM-ERROR"} {
		raw, err := json.Marshal(map[string]any{
			"network": "roger-public", "type": TypeReceipt, "version": towerobj.FormatInt(Version),
			"attempt_id": "att-1", "station_id": "st-1", "request_digest": digestOf([]byte("q")),
			"response_digest": digestOf(nil), "usage_in": "0", "usage_out": "0", "tok_in": "0", "tok_out": "0",
			"failure": class,
		})
		require.NoError(t, err)
		signed, err := towerobj.Sign(priv, "roger-public", TypeReceipt, Version, raw, "station_sig")
		require.NoError(t, err)
		_, err = ParseReceipt(signed, pub, "roger-public", "att-1", "st-1")
		require.Error(t, err, "class %.40q", class)
	}
	for _, class := range []string{FailureUpstreamRefused, FailureThrottled, FailureUpstreamError, FailureTooLarge} {
		rec, err := SignFailureReceipt(priv, "roger-public", Grant{AttemptID: "att-1", StationID: "st-1"}, nil, class)
		require.NoError(t, err)
		_, err = ParseReceipt(rec.Signed, pub, "roger-public", "att-1", "st-1")
		require.NoError(t, err, class)
	}
}
