package mochi

import "testing"

func TestJoinSides(t *testing.T) {
	tests := []struct {
		name        string
		front, back string
		want        string
	}{
		{"both sides", "Front", "Back", "Front\n---\nBack"},
		{"empty back", "Front", "", "Front"},
		{"multiline", "Q line 1\nQ line 2", "A", "Q line 1\nQ line 2\n---\nA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinSides(tt.front, tt.back); got != tt.want {
				t.Errorf("JoinSides(%q, %q) = %q, want %q", tt.front, tt.back, got, tt.want)
			}
		})
	}
}

func TestSplitSides(t *testing.T) {
	tests := []struct {
		name                string
		content             string
		wantFront, wantBack string
	}{
		{"both sides", "Front\n---\nBack", "Front", "Back"},
		{"no separator", "Just front", "Just front", ""},
		{"separator only at first occurrence", "F\n---\nB\n---\nC", "F", "B\n---\nC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			front, back := SplitSides(tt.content)
			if front != tt.wantFront || back != tt.wantBack {
				t.Errorf("SplitSides(%q) = (%q, %q), want (%q, %q)", tt.content, front, back, tt.wantFront, tt.wantBack)
			}
		})
	}
}

func TestJoinSplitRoundTrip(t *testing.T) {
	front, back := "Question", "Answer"
	gotFront, gotBack := SplitSides(JoinSides(front, back))
	if gotFront != front || gotBack != back {
		t.Errorf("round trip = (%q, %q), want (%q, %q)", gotFront, gotBack, front, back)
	}
}
