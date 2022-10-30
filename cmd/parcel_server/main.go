package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/fmstephe/location-system/pkg/lds/lds_csv"
	"github.com/fmstephe/location-system/pkg/quadtree"
	"github.com/fmstephe/location-system/pkg/store"
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

	tree, byteStore := fillTree(entriesChan)

	handler := ParcelHandler{
		byteStore: byteStore,
		tree:      tree,
	}

	http.HandleFunc("/survey", handler.Handle)
	http.Handle("/", http.FileServer(http.Dir("./html")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fillTree(parcelChan chan lds_csv.CSVParcelData) (quadtree.Tree[store.BytePointer], *store.ByteStore) {
	byteStore := store.NewByteStore()
	tree := quadtree.NewQuadTree[store.BytePointer](quadtree.NewLongLatView())

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

			// Json marshal the parcel data
			bytes, err := json.Marshal(line.Parcel)
			if err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			op, err := byteStore.New(uint32(len(bytes)))
			if err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			buffer := byteStore.Get(op)
			copy(buffer, bytes)

			if err := tree.Insert(northLong, westLat, op); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(northLong, eastLat, op); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(southLong, westLat, op); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}
			if err := tree.Insert(southLong, eastLat, op); err != nil {
				fmt.Printf("%d: %s\n", line.LineNum, err)
				errCount++
				continue
			}

			count++
		}
	}

	fmt.Printf("Inserted %d parcels into database\n", count)
	fmt.Printf("Failed to insert %d parcels into database\n", errCount)

	runtime.GC()

	return tree, byteStore
}
