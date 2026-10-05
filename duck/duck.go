// Package duck 判定语音闪避：某条总线上有人说话时，压低该总线上的其它音源。
package duck

// Event 是一次音频事件。
type Event struct {
	ID      string
	Bus     string
	Speaker string
	Action  string // start / end / gain
	Exempt  bool
	At      int64
}

// Result 是这条事件所在总线当前的增益。
type Result struct {
	Gain   float64
	Ducked bool
}

// busState 记录一条总线的说话人与释放窗口。
type busState struct {
	speakers map[string]bool
	lastEnd  int64
	hasEnd   bool
}

// Duck 是闪避台账。
type Duck struct {
	Threshold float64
	ReleaseMs int64

	buses   map[string]*busState
	seen    map[string]bool
	Scanned int
}

func (d *Duck) state(bus string) *busState {
	st, ok := d.buses[bus]
	if !ok {
		st = &busState{speakers: map[string]bool{}}
		d.buses[bus] = st
	}
	return st
}

// Handle 处理一条事件，返回该总线当前的增益。
func (d *Duck) Handle(ev Event) Result {
	if d.buses == nil {
		d.buses = map[string]*busState{}
	}
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	d.Scanned++

	if !d.seen[ev.ID] {
		d.seen[ev.ID] = true
		st := d.state(ev.Bus)
		switch ev.Action {
		case "start":
			st.speakers[ev.Speaker] = true
		case "end":
			delete(st.speakers, ev.Speaker)
			if len(st.speakers) == 0 {
				st.lastEnd = ev.At
				st.hasEnd = true
			}
		}
	}

	if ev.Exempt {
		return Result{Gain: 1, Ducked: false}
	}

	ducked := false
	if st, ok := d.buses[ev.Bus]; ok {
		switch {
		case len(st.speakers) > 0:
			ducked = true
		case st.hasEnd && ev.At-st.lastEnd <= d.ReleaseMs:
			ducked = true
		}
	}
	if !ducked {
		return Result{Gain: 1, Ducked: false}
	}
	return Result{Gain: d.Threshold, Ducked: true}
}
