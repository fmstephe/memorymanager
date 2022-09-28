package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fmstephe/location-system/pkg/lds/lds_csv"
	"github.com/fmstephe/location-system/pkg/lowgc_quadtree"
)

type ParcelHandler struct {
	tree lowgc_quadtree.T[lds_csv.ParcelData]
}

func (s *ParcelHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
