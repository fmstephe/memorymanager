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

	data, err := lds_csv.ReadCSVData(f)
	if err != nil {
		fmt.Printf("Error reading csv data %s\n", err)
		return
	}

	fmt.Printf("Successfully read %d lds data-points\n", len(data))

	for i, line := range data {
		fmt.Printf("%d: %v - %d\n", i, line.Plot, line.Id)
	}
}
