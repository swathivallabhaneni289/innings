package domain

import "testing"

func TestBucketsAndAverage(t *testing.T) {
	cases := []struct {
		ratings []float64
		avg     float64
		count   int
	}{
		{nil, 0, 0},
		{[]float64{5}, 5, 1},
		{[]float64{0.5, 5}, 2.75, 2},
		{[]float64{3, 3.5, 4}, 3.5, 3},
		{[]float64{7, 0}, 0, 0}, // out of range is ignored
	}
	for _, c := range cases {
		b := make([]int, 10)
		for _, r := range c.ratings {
			AddRating(b, r)
		}
		avg, n := BucketAverage(b)
		if avg != c.avg || n != c.count {
			t.Errorf("%v: got %v/%d want %v/%d", c.ratings, avg, n, c.avg, c.count)
		}
	}
}

func TestValidRating(t *testing.T) {
	for r, want := range map[float64]bool{0: false, 0.5: true, 1: true, 2.5: true, 5: true, 5.5: false, 1.25: false, -1: false} {
		if ValidRating(r) != want {
			t.Errorf("%v: want %v", r, want)
		}
	}
}

func TestPlayerAveragesAndPotm(t *testing.T) {
	players := PlayerAverages(map[string]Tally{
		"a": {Sum: 20, Count: 2}, // 10, too few for POTM
		"b": {Sum: 27, Count: 3}, // 9
		"c": {Sum: 36, Count: 4}, // 9, more ratings → ahead of b
		"d": {Sum: 0, Count: 0},  // dropped
		"e": {Sum: 15, Count: 3}, // 5
	})
	want := []string{"a", "c", "b", "e"}
	if len(players) != len(want) {
		t.Fatalf("got %v", players)
	}
	for i, id := range want {
		if players[i].PlayerID != id {
			t.Errorf("#%d: got %s want %s", i, players[i].PlayerID, id)
		}
	}
	if p := FansPotm(players); p == nil || *p != "c" {
		t.Errorf("potm: %v", p)
	}
	if FansPotm(PlayerAverages(map[string]Tally{"a": {10, 2}})) != nil {
		t.Error("potm needs 3 ratings")
	}
	if PlayerAverages(nil) == nil {
		t.Error("nil slice")
	}
}

func TestWatchShares(t *testing.T) {
	s := WatchShares(map[WatchPlace]int{WatchHome: 6, WatchStadium: 1, WatchFriends: 3, WatchCafe: 0})
	if len(s) != 3 || s[0].Place != WatchHome || s[0].Share != 0.6 || s[2].Place != WatchStadium {
		t.Errorf("got %v", s)
	}
	if WatchShares(nil) == nil || len(WatchShares(nil)) != 0 {
		t.Error("empty must be non-nil")
	}
	c := WatchCounts(map[WatchPlace]int{WatchOffice: 2, WatchHome: 2, WatchCafe: 5})
	if len(c) != 3 || c[0].Place != WatchCafe || c[1].Place != WatchHome {
		t.Errorf("counts %v", c)
	}
}

func TestShares(t *testing.T) {
	sh, total := Shares([2]string{"ind", "aus"}, map[string]int{"ind": 3, "aus": 1, "eng": 9})
	if total != 4 || sh["ind"] != 0.75 || sh["aus"] != 0.25 {
		t.Errorf("got %v %d", sh, total)
	}
	sh, total = Shares([2]string{"ind", "aus"}, nil)
	if total != 0 || len(sh) != 2 {
		t.Errorf("empty: %v", sh)
	}
}
