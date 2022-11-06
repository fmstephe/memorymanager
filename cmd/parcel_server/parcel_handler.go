package main

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/metrics"
	"strconv"

	"github.com/fmstephe/location-system/pkg/quadtree"
	"github.com/fmstephe/location-system/pkg/store/bytestore"
)

type ParcelHandler struct {
	byteStore *bytestore.Store
	tree      quadtree.Tree[bytestore.Pointer]
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

	view := quadtree.NewView(lx, rx, ty, by)

	// First we check if there are too many parcels to display effectively
	// If there are too many parcels, then we sample a single parcel from
	// 100 subdivisions of the view
	count := s.tree.Count(view)
	if s.tree.Count(view) > 4096 {
		w.Write(startIncomplete)
		w.Write(startArray)
		fmt.Printf("limited survey splitting view %v - %d\n", view, count)
		views := view.Split(32)
		for _, view := range views {
			count := s.tree.Count(view)
			fmt.Printf("split view %v - %d\n", view, count)
			s.tree.Survey(view, surveyFunc(w, s.byteStore, 1))
		}
	} else {
		w.Write(startComplete)
		w.Write(startArray)
		fmt.Printf("total survey - %d\n", count)
		s.tree.Survey(view, surveyFunc(w, s.byteStore, 0))
	}

	w.Write(endArray)
	w.Write(end)

	printHeapAllocs("finish")
	runtime.GC()
}

var startComplete = []byte(`{"complete": true, `)
var startIncomplete = []byte(`{"complete": false, `)
var startArray = []byte(`"parcels": [`)
var endArray = []byte(`null]`)
var comma = []byte(`,`)
var end = []byte(`}`)

func surveyFunc(w http.ResponseWriter, byteStore *bytestore.Store, limit int) func(_, _ float64, bp bytestore.Pointer) bool {
	pointerSet := map[bytestore.Pointer]struct{}{}
	return func(_, _ float64, bp bytestore.Pointer) bool {
		if _, ok := pointerSet[bp]; ok {
			// We've already seen this pointer, don't write it out again
			return true
		}

		pointerSet[bp] = struct{}{}

		// A limit of 0 or less means unlimited
		if limit > 0 && len(pointerSet) > limit {
			return false
		}

		bytes := byteStore.Get(bp)

		_, err := w.Write(bytes)
		if err != nil {
			// TODO handle error
		}

		_, err = w.Write(comma)
		if err != nil {
			// TODO handle error
		}

		return true
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
