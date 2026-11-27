package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBandRangeOfIsTheConnectScreensRange(t *testing.T) {
	offers := []Offer{
		{NodeID: "a", Model: "m", PriceIn: 0.1, PriceOut: 0.6, Online: true},
		{NodeID: "b", Model: "m", PriceIn: 0.2, PriceOut: 0.9, Online: true},
		{NodeID: "c", Model: "m", PriceOut: 0.1, Online: false},
		{NodeID: "d", Model: "other", PriceOut: 0.1, Online: true},
	}
	for _, tc := range []struct {
		model    string
		ok       bool
		min, max float64
		stations int
		cheap    string
	}{
		{"m", true, 0.6, 0.9, 2, "a"},
		{"other", true, 0.1, 0.1, 1, "d"},
		{"nobody", false, 0, 0, 0, ""},
	} {
		br, ok := BandRangeOf(offers, tc.model)
		require.Equal(t, tc.ok, ok, tc.model)
		require.Equal(t, tc.min, br.Min, tc.model)
		require.Equal(t, tc.max, br.Max, tc.model)
		require.Equal(t, tc.stations, br.Stations, tc.model)
		require.Equal(t, tc.cheap, br.CheapNode, tc.model)
	}
}
