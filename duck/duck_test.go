package duck

import "testing"

func newDuck() *Duck {
	return &Duck{Threshold: 0.35, ReleaseMs: 300}
}

func TestIdleBusIsLoud(t *testing.T) {
	got := newDuck().Handle(Event{ID: "e1", Bus: "music", Action: "gain", At: 0})
	if got.Gain != 1 || got.Ducked {
		t.Fatalf("没人说话时应满音量：%+v", got)
	}
}

func TestTalkingBusIsDucked(t *testing.T) {
	d := newDuck()
	d.Handle(Event{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0})
	got := d.Handle(Event{ID: "e2", Bus: "party", Action: "gain", At: 10})
	if !got.Ducked || got.Gain != 0.35 {
		t.Fatalf("有人说话时应压低：%+v", got)
	}
}

func TestDuckingUsesThreshold(t *testing.T) {
	d := &Duck{Threshold: 0.5, ReleaseMs: 300}
	d.Handle(Event{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0})
	got := d.Handle(Event{ID: "e2", Bus: "party", Action: "gain", At: 10})
	if got.Gain != 0.5 {
		t.Fatalf("增益应为阈值：%+v", got)
	}
}

func TestEndedTalkerStopsDucking(t *testing.T) {
	d := newDuck()
	d.Handle(Event{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0})
	d.Handle(Event{ID: "e2", Bus: "party", Speaker: "p1", Action: "end", At: 100})
	got := d.Handle(Event{ID: "e3", Bus: "party", Action: "gain", At: 5000})
	if got.Ducked {
		t.Fatalf("释放窗口过后应恢复：%+v", got)
	}
}
