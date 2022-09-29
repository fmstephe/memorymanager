package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fmstephe/location-system/pkg/lds/lds_csv"
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

	for line := range entriesChan {
		if line.Error != nil {
			fmt.Printf("%d: %s\n", line.LineNum, line.Error)
		} else {
			fmt.Printf("%d: %v - %d\n", line.LineNum, line.Parcel.Plot.Box, line.Parcel.Id)
		}
	}
}
