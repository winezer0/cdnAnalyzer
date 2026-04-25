package docheck

import (
	"testing"

	"github.com/winezer0/ipinfo/pkg/iplocate"
	"github.com/winezer0/ipinfo/pkg/queryip"
)

func TestConvertIPLocationsToMap(t *testing.T) {
	locations := []queryip.IPLocation{
		{
			IP:       "1.1.1.1",
			IPLocate: &iplocate.IPLocate{Location: "AU"},
		},
		{
			IP:       "8.8.8.8",
			IPLocate: nil,
		},
	}

	got := convertIPLocationsToMap(locations)
	if len(got) != 2 {
		t.Fatalf("convert result size mismatch, got=%d", len(got))
	}

	if v, ok := got[0]["1.1.1.1"]; !ok || v != "AU" {
		t.Fatalf("unexpected first map, got=%v", got[0])
	}

	if v, ok := got[1]["8.8.8.8"]; !ok || v != "" {
		t.Fatalf("unexpected second map, got=%v", got[1])
	}
}
