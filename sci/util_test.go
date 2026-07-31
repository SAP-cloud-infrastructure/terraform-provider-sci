package sci

import (
	"reflect"
	"testing"

	archermodels "github.com/sapcc/archer/v2/models"
)

func TestRemovePrefixIPAddress(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain ipv4",
			input: "192.0.2.10",
			want:  "192.0.2.10",
		},
		{
			name:  "plain ipv6",
			input: "2001:db8::10",
			want:  "2001:db8::10",
		},
		{
			name:  "ipv4 cidr",
			input: "192.0.2.10/24",
			want:  "192.0.2.10",
		},
		{
			name:  "ipv6 cidr",
			input: "2001:db8::10/64",
			want:  "2001:db8::10",
		},
		{
			name:  "invalid input",
			input: "not-an-ip",
			want:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := removePrefixIPAddress(tc.input)
			if got != tc.want {
				t.Fatalf("removePrefixIPAddress(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestArcherInetAddressConversions(t *testing.T) {
	input := []any{"192.0.2.10", "2001:db8::10"}

	gotExpanded := expandToArcherInetAddressSlice(input)
	wantExpanded := []archermodels.InetAddress{"192.0.2.10", "2001:db8::10"}
	if !reflect.DeepEqual(gotExpanded, wantExpanded) {
		t.Fatalf("expandToArcherInetAddressSlice() = %#v, want %#v", gotExpanded, wantExpanded)
	}

	gotFlattened := flattenToArcherInetAddressSlice(gotExpanded)
	wantFlattened := []string{"192.0.2.10", "2001:db8::10"}
	if !reflect.DeepEqual(gotFlattened, wantFlattened) {
		t.Fatalf("flattenToArcherInetAddressSlice() = %#v, want %#v", gotFlattened, wantFlattened)
	}
}
