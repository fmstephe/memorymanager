package lds_csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// In the case that the CSV reader generates an error we generate a line
// ([]string) with two elements, the first is the errString constant and the
// second is the error itself
const errString = "error"

type RawCSVEntry struct {
	LineNum          int
	Plot             ParcelDimensions
	Id               int64
	Appellation      string
	AffectedSurveys  string
	ParcelIntent     string
	TopologyType     string
	StatutoryActions string
	LandDistrict     string
	Titles           string
	SurveyArea       float64
	CalcArea         float64
	// If Error is not nil, then all other fields must be zeroed
	Error error
}

type ParcelDimensions struct {
	Polygons []Polygon
	Box      BoundingBox
}

type Polygon struct {
	Points []Point
}

type BoundingBox struct {
	NorthWest Point
	SouthEast Point
}

type Point struct {
	Longitude float64
	Latitude  float64
}

func ReadAllCSVData(r io.Reader) ([]RawCSVEntry, error) {
	// Read all of the csv data from the reader
	csvR := csv.NewReader(r)

	// Consume the first line and ignore it
	// This line contains the column names and no data
	// TODO actually read and validate the column names
	_, err := csvR.Read()
	if err != nil {
		return nil, err
	}

	// Consume all data lines
	data, err := csvR.ReadAll()
	if err != nil {
		return nil, err
	}

	// Shift the read csv data into a nice struct
	entries := []RawCSVEntry{}
	lineNum := 0
	for _, line := range data {
		lineNum++
		entry := processCSVLine(line, lineNum)
		entries = append(entries, entry)
	}

	return entries, nil
}

// TODO right now errors just get printed and processing stops
// This is ok for a prototype, but isn't reasonable for a proper implementation
// Errors will likely need to be folded into the RawCSVEntry struct itself
func ReadCSVDataAsync(r io.Reader) (chan RawCSVEntry, error) {
	// Read all of the csv data from the reader
	csvR := csv.NewReader(r)
	lineChan, err := readCSVLinesAsync(csvR)
	if err != nil {
		return nil, err
	}

	entriesChan := processCSVLinesAsync(lineChan)

	return entriesChan, nil
}

func readCSVLinesAsync(csvR *csv.Reader) (chan []string, error) {
	lineChan := make(chan []string, 1024)

	// Consume the first line and ignore it
	// This line contains the column names and no data
	// TODO actually read and validate the column names
	_, err := csvR.Read()
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(lineChan)
		for {
			line, err := csvR.Read()
			if err == io.EOF {
				// No more csv data
				return
			}
			if err != nil {
				lineChan <- []string{errString, err.Error()}
			}

			lineChan <- line
		}
	}()
	return lineChan, nil
}

func processCSVLinesAsync(lineChan chan []string) chan RawCSVEntry {
	// Shift the read csv data into a nice struct
	entriesChan := make(chan RawCSVEntry, 1024)

	go func() {
		defer close(entriesChan)

		lineNum := 0
		for line := range lineChan {
			lineNum++
			entry := processCSVLine(line, lineNum)
			entriesChan <- entry
		}
	}()

	return entriesChan
}

func processCSVLine(line []string, lineNum int) RawCSVEntry {
	// First case is we may be consuming an error line
	if len(line) == 2 && line[0] == errString {
		return RawCSVEntry{LineNum: lineNum, Error: errors.New(line[1])}
	}

	// We expect exactly 11 data elements per line
	if len(line) != 11 {
		return RawCSVEntry{LineNum: lineNum, Error: fmt.Errorf("Error reading line %d, %d parts expect 11 in %v", lineNum, len(line), line)}
	}

	// Read the polygon(s) which define the physical dimensions of the parcel
	plot, err := parseParcelDimensions(line[0])
	if err != nil {
		return RawCSVEntry{LineNum: lineNum, Error: fmt.Errorf("Error reading line %d bad plot %q in %v %s", lineNum, line[0], line, err)}
	}

	// Read the ID of the parcel as an integer
	id, err := strconv.ParseInt(line[1], 10, 64)
	if err != nil {
		return RawCSVEntry{LineNum: lineNum, Error: fmt.Errorf("Error reading line %d bad Id %s in %v %s", lineNum, line[1], line, err)}
	}

	// Read the survey area of the parcel as a float
	surveyArea, err := parseFloat(line[9])
	if err != nil {
		return RawCSVEntry{LineNum: lineNum, Error: fmt.Errorf("Error reading line %d bad Survey Area %q in %v %s", lineNum, line[9], line, err)}
	}

	// Read the calculated area of the parcel as a float
	calcArea, err := parseFloat(line[10])
	if err != nil {
		return RawCSVEntry{LineNum: lineNum, Error: fmt.Errorf("Error reading line %d bad Calc Area %q in %v %s", lineNum, line[10], line, err)}
	}

	return RawCSVEntry{
		LineNum:          lineNum,
		Plot:             plot,
		Id:               id,
		Appellation:      line[2],
		AffectedSurveys:  line[3],
		ParcelIntent:     line[4],
		TopologyType:     line[5],
		StatutoryActions: line[6],
		LandDistrict:     line[7],
		Titles:           line[8],
		SurveyArea:       surveyArea,
		CalcArea:         calcArea,
	}
}

func parseFloat(raw string) (float64, error) {
	if raw == "" {
		return 0, nil
	}

	return strconv.ParseFloat(raw, 64)
}

// Expects a string with "POLYGON"/"MULTIPOLYGON"
// followed by one or more polygon strings which look like
// "((11.123 -22.123,33.123 -44.123))"
//
// Example a POLYGON would look like
// "POLYGON ((11.123 -22.123,33.123 -44.123))"
//
// Example a POLYGON would look like
// "MULTIPOLYGON (((11.123 -22.123,33.123 -44.123)),((55.123 -66.123, 77.123 -88.123)))"
func parseParcelDimensions(raw string) (ParcelDimensions, error) {
	polys := []Polygon{}

	// Remove leading POLYGON/MULTIPOLYGON string
	_, body, _ := strings.Cut(raw, "(")

	// Split into parts using '('
	parts := strings.Split(body, "(")

	for _, part := range parts {
		if part == "" {
			continue
		}
		poly, err := parsePolygon(part)
		if err != nil {
			return ParcelDimensions{}, err
		}
		polys = append(polys, poly)
	}

	boundingBox := getBoundingBox(polys)

	return ParcelDimensions{
		Polygons: polys,
		Box:      boundingBox,
	}, nil
}

// Expects a string with "((" + "long-lat pairs separated by comma" + "))"
// Sometimes there will be a trailing comma at the end of the string
//
// In the case of a POLYGON  "((11.123 -22.123,33.123 -44.123))"
//
// In the case of a MULTIPOLYGON "((11.123 -22.123,33.123 -44.123)),"
func parsePolygon(raw string) (Polygon, error) {
	poly := Polygon{}

	rawLongLats, _, _ := strings.Cut(raw, ")")
	longLats := strings.Split(rawLongLats, ",")

	for _, longLat := range longLats {
		point, err := parseLongLatPair(longLat)
		if err != nil {
			return Polygon{}, err
		}
		poly.Points = append(poly.Points, point)
	}

	return poly, nil
}

// Expects a string with two floats separated by a single space
// Example "11.123 -22.123"
func parseLongLatPair(raw string) (Point, error) {
	// The long lat pair are separated by a single space
	longStr, latStr, _ := strings.Cut(raw, " ")

	long, err := parseFloat(longStr)
	if err != nil {
		return Point{}, err
	}

	lat, err := parseFloat(latStr)
	if err != nil {
		return Point{}, err
	}

	return Point{
		Longitude: long,
		Latitude:  lat,
	}, nil
}

func getBoundingBox(polygons []Polygon) BoundingBox {
	// This will be the lowest longitude value
	westLongitude := math.MaxFloat64
	// This will be the largest longitude value
	eastLongitude := -westLongitude
	// This will be the lowest latitude value
	southLatitude := math.MaxFloat64
	// This will be the largest latitude value
	northLatitude := -southLatitude

	for _, polygon := range polygons {
		for _, point := range polygon.Points {
			if westLongitude > point.Longitude {
				westLongitude = point.Longitude
			}
			if eastLongitude < point.Longitude {
				eastLongitude = point.Longitude
			}
			if southLatitude > point.Latitude {
				southLatitude = point.Latitude
			}
			if northLatitude < point.Latitude {
				northLatitude = point.Latitude
			}
		}
	}

	return BoundingBox{
		NorthWest: Point{
			Longitude: westLongitude,
			Latitude:  northLatitude,
		},
		SouthEast: Point{
			Longitude: eastLongitude,
			Latitude:  southLatitude,
		},
	}
}
