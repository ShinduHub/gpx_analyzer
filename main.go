package main

import (
	"flag"
	"fmt"
	"github.com/ptrv/go-gpx"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var target gpx.Wpt
var dist *float64
var wg sync.WaitGroup
var globMinDist = math.MaxFloat64
var globAbsFileName string

func scan(file string) {
	defer wg.Done()
	minDist := math.MaxFloat64
	var gpxFile, err = gpx.ParseFile(file)
	absFileName, _ := filepath.Abs(file)
	if err != nil {
		fmt.Println("Error while parsing the file: ", absFileName)
		return
	}

	for _, track := range gpxFile.Tracks {
		for _, segment := range track.Segments {
			for _, waypoint := range segment.Waypoints {

				currDist := waypoint.Distance2D(&target)
				if minDist > currDist {
					minDist = currDist
				}

			}
		}
	}

	if minDist <= *dist {
		fmt.Printf("%8.0f m, %s\n", minDist, absFileName)
	} else if globMinDist > minDist {
		globMinDist = minDist
		globAbsFileName = absFileName
	}
}

func main() {

	lat := flag.Float64("lat", 0, "latitude of target (North to South)")
	lon := flag.Float64("lon", 0, "longitude of target (East to West)")
	dist = flag.Float64("dist", 1000, "distance between target and waypoint in meters")
	root := flag.String("path", ".", "path containing the gpx files (from the executable)")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		return
	}

	target = gpx.Wpt{Lat: *lat, Lon: *lon}
	var files []string
	abs, err := filepath.Abs(*root)
	fmt.Println("searching in", abs, "...")
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(*root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".gpx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	wg.Add(len(files))

	for _, file := range files {
		go scan(file)
	}

	wg.Wait()

	fmt.Println("\nNearest out of Range was:")

	if globMinDist != math.MaxFloat64 {
		fmt.Printf("%8.0f m, %s", globMinDist, globAbsFileName)
	}

}
