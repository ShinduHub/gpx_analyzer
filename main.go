package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ptrv/go-gpx"
)

var version = "2.1.0"
var target gpx.Wpt
var dist *float64
var debug *bool
var wg sync.WaitGroup
var globInnerMinDist = math.MaxFloat64
var globOuterMinDist = math.MaxFloat64
var globInnerMinDistAbsFileName string
var globOuterMinDistAbsFileName string
var parallel *bool
var finished int
var total int

// scans a file given as a string
func scan(file string) {
	if *parallel {
		defer wg.Done()
	}
	minDist := math.MaxFloat64
	var gpxFile, err = gpx.ParseFile(file)
	absFileName, _ := filepath.Abs(file)
	if err != nil {
		fmt.Println("\r[Error] while parsing the file: ", absFileName)
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

	if minDist == math.MaxFloat64 {
		fmt.Println("\r[Warning] file: ", absFileName, " has no waypoints")
		return
	}

	if *debug {
		fmt.Printf("\r%8.0f m, %s\n", minDist, absFileName)
	}

	if minDist <= *dist {
		if globInnerMinDist > minDist {
			globInnerMinDist = minDist
			globInnerMinDistAbsFileName = absFileName
		}
		fmt.Printf("\r%8.0f m, %s\n", minDist, absFileName)
	} else if globOuterMinDist > minDist {
		globOuterMinDist = minDist
		globOuterMinDistAbsFileName = absFileName
	}

	if *parallel {
		finished++
		printState(finished, total)
	}
}

func printState(index int, total int) {
	fmt.Print("\r")
	fmt.Print(strconv.Itoa(int(math.Round(float64(index)/float64(total)*100.0))), "% scanning...")
}

func main() {

	lat := flag.Float64("lat", 0, "latitude of target (North to South)")
	lon := flag.Float64("lon", 0, "longitude of target (East to West)")
	dist = flag.Float64("dist", 1000, "distance between target and waypoint in meters")
	root := flag.String("path", ".", "path containing the gpx files (from the executable)")
	parallel = flag.Bool("parallel", false, "parallel mode (faster, but this eats up your hardware!)")
	debug = flag.Bool("debug", false, "debug mode (print out all file distances)")

	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Print("gpx_analyzer version ", version, "\n\n")
		flag.PrintDefaults()
		return
	}

	target = gpx.Wpt{Lat: *lat, Lon: *lon}

	var files []string
	fileCount := 0
	err := filepath.Walk(*root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".gpx") {
			files = append(files, path)
			fileCount++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	abs, err := filepath.Abs(*root)
	fmt.Println("\rsearching in", abs, strings.Join([]string{"(", strconv.Itoa(fileCount), " GPX-Files)"}, ""), " parallel =", *parallel, "...")
	if err != nil {
		panic(err)
	}

	total = len(files)

	if *parallel {

		wg.Add(total)

		for _, file := range files {
			go scan(file)
		}

		for finished < total {
			printState(finished, total)
		}
		wg.Wait()

	} else {

		for i, file := range files {
			scan(file)
			printState(i, total)
		}

	}

	fmt.Print("\r")

	if globInnerMinDist <= math.MaxFloat64 {
		fmt.Println("\n\nNearest in dist was:")
		fmt.Printf("%8.0f m, %s\n", globInnerMinDist, globInnerMinDistAbsFileName)
	}

	if globOuterMinDist <= math.MaxFloat64 {
		fmt.Println("\n\nNearest out of dist was:")
		fmt.Printf("%8.0f m, %s\n", globOuterMinDist, globOuterMinDistAbsFileName)
	}

	fmt.Println("\nDone.")

}
