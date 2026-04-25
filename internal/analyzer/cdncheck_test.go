package analyzer

import (
	"reflect"
	"testing"

	"github.com/winezer0/ipinfo/pkg/asninfo"
)

func TestGetUniqueOrgNumbers(t *testing.T) {
	input := []asninfo.ASNInfo{
		{FoundASN: false, OrganisationNumber: 100},
		{FoundASN: true, OrganisationNumber: 100},
		{FoundASN: true, OrganisationNumber: 100},
		{FoundASN: true, OrganisationNumber: 200},
		{FoundASN: false, OrganisationNumber: 300},
		{FoundASN: true, OrganisationNumber: 200},
	}

	want := []uint64{100, 200}
	got := getUniqueOrgNumbers(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result, got=%v want=%v", got, want)
	}
}

