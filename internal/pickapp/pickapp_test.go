package pickapp

import "testing"

func TestPickBySeedStable(t *testing.T) {
	p1 := PickBySeed("56d1fdef-5d4d-480a-8929-e7402454b33f", nil)
	p2 := PickBySeed("56d1fdef-5d4d-480a-8929-e7402454b33f", nil)
	if p1.AppVersion != p2.AppVersion {
		t.Fatalf("same seed different app: %v vs %v", p1, p2)
	}
}
