package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/fmstephe/location-system/pkg/lds/lds_csv"
	"github.com/fmstephe/location-system/pkg/lowgc_quadtree"
)

var (
	filePathFlag = flag.String("path", "", "The path to an lds csv file to read")
)

func main() {
	flag.Parse()

	if *filePathFlag == "" {
		fmt.Printf("No -path flag provided. Nothing to read.\n")
		return
	}

	f, err := os.Open(*filePathFlag)
	if err != nil {
		fmt.Printf("Error opening csv data %s\n", err)
		return
	}

	entriesChan, err := lds_csv.ReadCSVDataAsync(f)
	if err != nil {
		fmt.Printf("Error reading csv data %s\n", err)
		return
	}

	tree := fillTree(entriesChan)

	handler := ParcelHandler{
		tree: tree,
	}

	http.HandleFunc("/survey", handler.Handle)
	http.ListenAndServe(":8080", nil)
}

func fillTree(parcelChan chan lds_csv.CSVParcelData) lowgc_quadtree.T[lds_csv.ParcelData] {
	// TODO convert this to use a byte store
	tree := lowgc_quadtree.NewQuadTree[lds_csv.ParcelData](lowgc_quadtree.NewLongLatView())

	count := 0
	errCount := 0
	for line := range parcelChan {
		if line.Error != nil {
			//fmt.Printf("%d: %s\n", line.LineNum, line.Error)

			errCount++
		} else {
			northLong := line.Parcel.Plot.Box.NorthWest.Longitude
			southLong := line.Parcel.Plot.Box.SouthEast.Longitude
			westLat := line.Parcel.Plot.Box.NorthWest.Latitude
			eastLat := line.Parcel.Plot.Box.SouthEast.Latitude

			if err := tree.Insert(northLong, westLat, line.Parcel); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(northLong, eastLat, line.Parcel); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(southLong, westLat, line.Parcel); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(southLong, eastLat, line.Parcel); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}

			count++
		}
	}

	fmt.Printf("Inserted %d parcels into database\n", count)
	fmt.Printf("Failed to insert %d parcels into database\n", errCount)

	return tree
}
