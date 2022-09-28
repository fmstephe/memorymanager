package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/metrics"
	"strconv"

	"github.com/fmstephe/location-system/pkg/lds/lds_csv"
	"github.com/fmstephe/location-system/pkg/lowgc_quadtree"
)

type ParcelHandler struct {
	tree lowgc_quadtree.T[lds_csv.ParcelData]
}

func (s *ParcelHandler) Handle(w http.ResponseWriter, r *http.Request) {
	printHeapAllocs("start")

	err := r.ParseForm()
	if err != nil {
		// TODO handle error
		return
	}

	lxStr := r.Form.Get("lx")
	lx, err := strconv.ParseFloat(lxStr, 64)
	if err != nil {
		fmt.Printf("lx %s", err)
		return
	}

	rxStr := r.Form.Get("rx")
	rx, err := strconv.ParseFloat(rxStr, 64)
	if err != nil {
		fmt.Printf("rx %s", err)
		return
	}

	tyStr := r.Form.Get("ty")
	ty, err := strconv.ParseFloat(tyStr, 64)
	if err != nil {
		fmt.Printf("ty %s", err)
		return
	}

	byStr := r.Form.Get("by")
	by, err := strconv.ParseFloat(byStr, 64)
	if err != nil {
		fmt.Printf("by %s", err)
		return
	}

	view := lowgc_quadtree.NewView(lx, rx, ty, by)
	w.Write(startArray)
	s.tree.Survey(view, surveyFunc(w))
	w.Write(endArray)

	printHeapAllocs("finish")
}

var startArray = []byte("[")
var endArray = []byte("nil]")
var comma = []byte(",")

func surveyFunc(w http.ResponseWriter) func(_, _ float64, e lds_csv.ParcelData) {
	return func(_, _ float64, e lds_csv.ParcelData) {
		bytes, err := json.Marshal(e)
		if err != nil {
			// TODO handle error
		}

		_, err = w.Write(bytes)
		if err != nil {
			// TODO handle error
		}

		_, err = w.Write(comma)
		if err != nil {
			// TODO handle error
		}

	}
}

func printHeapAllocs(prefix string) {
	// Name of the metric we want to read.
	const myMetric = "/gc/heap/objects:objects"

	// Create a sample for the metric.
	sample := make([]metrics.Sample, 1)
	sample[0].Name = myMetric

	// Sample the metric.
	metrics.Read(sample)

	objects := sample[0].Value.Uint64()
	fmt.Printf("%s objects %d\n", prefix, objects)
}
