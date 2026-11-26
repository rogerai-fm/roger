package towerhub

// Contract: features/tower/tower_failed_attempt_release.feature.
//
// The node loop lets exactly one kind of receipt ride with a failure: a failure receipt, which
// claims nothing. A receipt that claims work is still dropped on a failed serve.

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/towercore/dispatch"
)

func TestOnlyAFailureReceiptRidesWithAFailure(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	g := dispatch.Grant{AttemptID: "att-1", StationID: "st-1"}
	fail, err := dispatch.SignFailureReceipt(priv, "roger-public", g, []byte("q"), dispatch.FailureUpstreamError)
	require.NoError(t, err)
	served, err := dispatch.SignReceipt(priv, "roger-public", g, []byte("q"), []byte("a"), dispatch.Usage{In: 1, Out: 1}, dispatch.Usage{})
	require.NoError(t, err)

	for _, tc := range []struct {
		name    string
		receipt []byte
		keep    bool
	}{
		{"a failure receipt", fail.Signed, true},
		{"a receipt that claims work", served.Signed, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := runOneFailedServe(t, tc.receipt)
			if tc.keep {
				require.Equal(t, tc.receipt, got)
			} else {
				require.Empty(t, got)
			}
		})
	}
}

// runOneFailedServe drives ServeLoop through one consumer job whose executor fails with receipt,
// and returns the receipt the consumer's submit came back with.
func runOneFailedServe(t *testing.T, receipt []byte) []byte {
	t.Helper()
	id := newTestNode(t)
	s, base := serveTestRig(t)
	s.RegisterNode("st-1", id.auth())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	node := id.client(base, 5*time.Second)
	go func() { _ = ServeLoop(ctx, node, "st-1", failingExec{receipt: receipt}, nil) }()
	consumer := &Client{BaseURL: base}
	sctx, scancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer scancel()
	res, err := consumer.SubmitJob(sctx, []byte("att-1|st-1"), []byte("sealed-req"))
	require.NoError(t, err)
	require.NotEmpty(t, res.Failure)
	return res.Receipt
}

type failingExec struct{ receipt []byte }

func (f failingExec) Serve(context.Context, []byte, []byte) ([]byte, []byte, string) {
	return nil, f.receipt, "the model did not answer (status 500)"
}
