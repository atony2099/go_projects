package img

//go:generate go run main.go

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

func Image(cha map[string]time.Duration, average float64) []byte {

	var cList []chart.Value
	loc, _ := time.LoadLocation("Asia/Shanghai")
	keys := make([]string, len(cha))
	i := 0
	for key := range cha {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		now := time.Now().In(loc)
		t, _ := time.Parse(time.DateOnly, key)
		// fmt.Println(t, key, "t")
		nows := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		date := nows.Format("01-02")
		keys := fmt.Sprintf("%s(%s)", date, nows.Weekday().String()[0:3])
		h := cha[key].Hours()
		hh := math.Round(h * 10)
		cList = append(cList, chart.Value{Value: hh / 10, Label: keys})
		fmt.Println(hh, keys, "keys")

	}

	s := fmt.Sprintf("%s--%s", keys[0], keys[len(keys)-1])
	cList = append(cList, chart.Value{Value: average, Label: "average"})
	cList = append(cList, chart.Value{Value: 0, Label: ""})

	graph := chart.BarChart{
		Title: s,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		// Height: 512,
		// BarWidth: 60,
		Bars:      cList,
		BaseValue: 0,
		// UseBaseValue: true,
	}

	// f, _ := os.Create("a.png")
	buffer := bytes.NewBuffer([]byte{})
	graph.Render(chart.PNG, buffer)
	// graph.Render(chart.PNG, f)
	return buffer.Bytes()
}

func Pip(values []chart.Value) ([]byte, error) {

	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: values,
	}

	// f, _ := os.Create("a.png")
	buffer := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	// graph.Render(chart.PNG, f)
	return buffer.Bytes(), nil
}
