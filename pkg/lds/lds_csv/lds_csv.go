package lds_csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type RawCSVEntry struct {
	Plot             []Polygon
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
}

type Polygon struct {
	Points []Point
}

type Point struct {
	Longitude float64
	Lattitude float64
}

func ReadCSVData(r io.Reader) ([]RawCSVEntry, error) {
	// Read all of the csv data from the reader
	csvR := csv.NewReader(r)
	data, err := csvR.ReadAll()
	if err != nil {
		return nil, err
	}

	// Shift the read csv data into a nice struct
	entries := []RawCSVEntry{}
	for i, line := range data {
		if i == 0 {
			// the first line simply names all the csv columns
			// Ignore this
			continue
		}

		plot, err := parsePlot(line[0])
		if err != nil {
			return nil, fmt.Errorf("Error reading line %d bad plot %q in %v %s", i, line[0], line, err)
		}

		if len(line) != 11 {
			return nil, fmt.Errorf("Bad line %v has %d parts, expect 11 %s", line, len(line), err)
		}

		id, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error reading line %d bad Id %s in %v %s", i, line[1], line, err)
		}

		surveyArea, err := parseFloat(line[9])
		if err != nil {
			return nil, fmt.Errorf("Error reading line %d bad Survey Area %q in %v %s", i, line[9], line, err)
		}

		calcArea, err := parseFloat(line[10])
		if err != nil {
			return nil, fmt.Errorf("Error reading line %d bad Calc Area %q in %v %s", i, line[10], line, err)
		}

		entry := RawCSVEntry{
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
		entries = append(entries, entry)
	}

	return entries, nil
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
func parsePlot(raw string) ([]Polygon, error) {
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
			return nil, err
		}
		polys = append(polys, poly)
	}
	return polys, nil
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
		Lattitude: lat,
	}, nil
}
