// Command duck 从标准输入读 JSONL 音频事件，逐行写该总线的增益。
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"example.com/voice-duck/duck"
)

type record struct {
	ID      string `json:"id"`
	Bus     string `json:"bus"`
	Speaker string `json:"speaker"`
	Action  string `json:"action"`
	Exempt  bool   `json:"exempt"`
	At      int64  `json:"at"`
}

type line struct {
	ID     string  `json:"id"`
	Gain   float64 `json:"gain"`
	Ducked bool    `json:"ducked"`
}

func feed(d *duck.Duck, recs []record) {
	for _, r := range recs {
		res := d.Handle(duck.Event{
			ID: r.ID, Bus: r.Bus, Speaker: r.Speaker,
			Action: r.Action, Exempt: r.Exempt, At: r.At,
		})
		out, _ := json.Marshal(line{ID: r.ID, Gain: res.Gain, Ducked: res.Ducked})
		fmt.Println(string(out))
	}
}

func stream(d *duck.Duck) {
	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		text := strings.TrimSpace(in.Text())
		if text == "" {
			continue
		}
		var r record
		if err := json.Unmarshal([]byte(text), &r); err != nil {
			fmt.Fprintln(os.Stderr, "跳过无法解析的行:", text)
			continue
		}
		feed(d, []record{r})
	}
}

func scaleRecs(buses, gains int) []record {
	out := make([]record, 0, buses+gains)
	for i := 0; i < buses; i++ {
		out = append(out, record{
			ID: fmt.Sprintf("s%d", i), Bus: fmt.Sprintf("b%d", i),
			Speaker: fmt.Sprintf("p%d", i), Action: "start", At: 0,
		})
	}
	for i := 0; i < gains; i++ {
		out = append(out, record{
			ID: fmt.Sprintf("g%d", i), Bus: fmt.Sprintf("b%d", i%buses), Action: "gain", At: int64(i + 1),
		})
	}
	return out
}

func scenario(name string) (*duck.Duck, []record) {
	switch name {
	case "cross-bus":
		return &duck.Duck{Threshold: 0.35, ReleaseMs: 300}, []record{
			{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0},
			{ID: "e2", Bus: "music", Action: "gain", At: 10},
		}
	case "two-talkers":
		return &duck.Duck{Threshold: 0.35, ReleaseMs: 300}, []record{
			{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0},
			{ID: "e2", Bus: "party", Speaker: "p2", Action: "start", At: 5},
			{ID: "e3", Bus: "party", Speaker: "p1", Action: "end", At: 10},
			{ID: "e4", Bus: "party", Action: "gain", At: 20},
		}
	case "exempt":
		return &duck.Duck{Threshold: 0.35, ReleaseMs: 300}, []record{
			{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0},
			{ID: "e2", Bus: "party", Action: "gain", Exempt: true, At: 10},
		}
	case "release":
		return &duck.Duck{Threshold: 0.35, ReleaseMs: 300}, []record{
			{ID: "e1", Bus: "party", Speaker: "p1", Action: "start", At: 0},
			{ID: "e2", Bus: "party", Speaker: "p1", Action: "end", At: 100},
			{ID: "e3", Bus: "party", Action: "gain", At: 150},
		}
	case "work":
		return &duck.Duck{Threshold: 0.35, ReleaseMs: 300}, scaleRecs(500, 3000)
	}
	fmt.Fprintf(os.Stderr, "未知场景 %q\n", name)
	os.Exit(2)
	return nil, nil
}

func main() {
	sample := flag.String("sample", "", "场景名；留空则从标准输入读 JSONL")
	flag.Parse()

	if *sample == "" {
		stream(&duck.Duck{Threshold: 0.35, ReleaseMs: 300})
		return
	}

	d, recs := scenario(*sample)
	if *sample == "work" {
		for _, r := range recs {
			d.Handle(duck.Event{
				ID: r.ID, Bus: r.Bus, Speaker: r.Speaker,
				Action: r.Action, Exempt: r.Exempt, At: r.At,
			})
		}
		out, _ := json.Marshal(map[string]int{"scanned": d.Scanned, "events": len(recs)})
		fmt.Println(string(out))
		return
	}
	feed(d, recs)
}
