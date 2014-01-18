package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type experiment struct {
	Header   []string
	Segments []segment
}

var numWorkers = flag.Int("n", 4, "the number of processors availble to the program")
var f = flag.String("f", "rharr-c453.head", "the path to the file to be parsed")

type segment struct {
	index int
	lines []string
	flags,
	tags []int
	data [][]float64
	I,
	sigmaI,
	iOverSigmaI float64
}

func (s segment) stdDev() float64 {
	total := 0.0
	for i := 0; i < len(s.data); i++ {
		total += s.data[i][3]
	}
	return total / float64(len(s.data))
}

// take in a scanner, and return a segment of the text file you are looking for
func chomp(scanner *bufio.Reader, lookFor string) ([]string, bool) {
	found := false
	var lines []string
	for !found {
		line, err := scanner.ReadString('\n')
		if err != nil {
			found = true
		}
		if strings.Contains(line, lookFor) {
			found = true
		}
		lines = append(lines, line)
		if line == "" {

			return lines, true
		}
	}

	return lines, false
}

// parse out
func parse(seg segment) [][]float64 {
	seg.data = make([][]float64, len(seg.lines))
	xBound := len(seg.lines)
	var tags []int
	for x := 0; x < xBound; x++ {

		line := seg.lines[x]
		if strings.Contains(line, " - ") {
			tags = append(tags, x)
			seg.data[x] = make([]float64, 5)
			b := strings.Split(line, " ")
			for y := 0; y < 5; y++ {

				hold, err := strconv.ParseFloat(b[y], 64)
				seg.data[x][y] = hold
				if err != nil {
					fmt.Println(err)
				}
			}
		}

	}
	return seg.data
}

// type listener creates an experiment, and reconsititues the santatized data
type listener struct {
	exp          *experiment
	fileLocation string
}

func (l listener) rebuild() {
	err := os.Rename(l.fileLocation, l.fileLocation+".old")
	if err != nil {
		fmt.Println("could not rename file")
	}
	file, err := os.Open(l.fileLocation)
	if err != nil {
		fmt.Println("could not open " + l.fileLocation)

	}
	writer := bufio.NewWriter(file)
	for i := 0; i < len(l.exp.Segments); i++ {
		for x := 0; x < len(l.exp.Segments[i].lines); x++ {
			writer.WriteString(l.exp.Segments[i].lines[x])
		}

	}
	file.Close()
}

// listen for segments and place them in order
func (l listener) listen(doneChan chan bool, segChan chan segment) {
	done := false
	for !done {
		select {
		case hold := <-segChan:
			l.exp.Segments[hold.index] = hold
		case done = <-doneChan:
			close(doneChan)
			close(segChan)
			l.rebuild()
		}
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*numWorkers)

	fileLocation := *f
	fmt.Println(fileLocation)
	file, err := os.Open(fileLocation)
	if err != nil {
		fmt.Println("File ", fileLocation, " does not exist.")
	}
	scanner := bufio.NewReader(file)
	exp := new(experiment)
	list := new(listener)
	list.exp = exp
	list.fileLocation = fileLocation
	exp.Header, _ = chomp(scanner, "Reflections measured after indexing")

	// collect lines from file and turn them into segments
	done := false
	i := 0
	var seg segment
	for !done {
		seg.lines, done = chomp(scanner, "Reflections measured after indexing")
		seg.index = i
		seg.data = parse(seg)
		fmt.Println(seg.data)
		fmt.Println(seg.stdDev())
		exp.Segments = append(exp.Segments, seg)
		i++

	}

}
