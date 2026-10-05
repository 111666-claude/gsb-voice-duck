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

// busState 记录一条总线的说话状态。
type busState struct {
	speakers map[string]bool // 当前在说话的人
	released bool            // 是否经历过「最后一人停止说话」
	lastEnd  int64           // 最后一人停止说话的时刻
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

// ducked 判定某条总线在 at 时刻是否应被压低。
func (d *Duck) ducked(st *busState, at int64) bool {
	if len(st.speakers) > 0 {
		return true
	}
	if st.released && at-st.lastEnd <= d.ReleaseMs {
		return true
	}
	return false
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

	st := d.state(ev.Bus)

	// 幂等：同一个 eventID 只处理一次。
	if !d.seen[ev.ID] {
		d.seen[ev.ID] = true
		switch ev.Action {
		case "start":
			st.speakers[ev.Speaker] = true
		case "end":
			if _, ok := st.speakers[ev.Speaker]; ok {
				delete(st.speakers, ev.Speaker)
				if len(st.speakers) == 0 {
					st.released = true
					st.lastEnd = ev.At
				}
			}
		}
	}

	// 豁免覆盖层：exempt 音源不受闪避影响。
	if ev.Exempt || !d.ducked(st, ev.At) {
		return Result{Gain: 1, Ducked: false}
	}
	return Result{Gain: d.Threshold, Ducked: true}
}
