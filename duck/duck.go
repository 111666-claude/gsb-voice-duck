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

// Duck 是闪避台账。
type Duck struct {
	Threshold float64
	ReleaseMs int64

	buses   map[string]bool
	seen    map[string]bool
	Scanned int
}

// Handle 处理一条事件，返回该总线当前的增益。
func (d *Duck) Handle(ev Event) Result {
	if d.buses == nil {
		d.buses = map[string]bool{}
	}
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	d.seen[ev.ID] = true

	switch ev.Action {
	case "start":
		d.buses[ev.Bus] = true
	case "end":
		d.buses[ev.Bus] = false
	}

	ducked := false
	for _, on := range d.buses {
		d.Scanned++
		if on {
			ducked = true
		}
	}
	if !ducked {
		return Result{Gain: 1, Ducked: false}
	}
	return Result{Gain: d.Threshold, Ducked: true}
}
