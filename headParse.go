package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
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
	data       []float64
	returnChan chan segment
}

func (s segment) stdDev() (float64, float64) {
	total := 0.0
	std := 0.0
	for i := 0; i < len(s.data); i++ {
		total += s.data[i]
	}
	avg := total / float64(len(s.data))

	for i := 0; i < len(s.data); i++ {
		std += math.Abs(s.data[i] - avg)
	}
	return std / float64(len(s.data)), avg
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
func (s segment) sanatize() {
	stdDev, avg := s.stdDev()
	x := 0
	offset := 0
	for i := 0; i < len(s.data); i++ {
		if (s.data[i] - avg) > 3*stdDev {
			s.flags = append(s.flags, i)
			x++
		}
	}
	for i := 0; i < len(s.flags); i++ {
		s.lines = append(s.lines[:s.tags[s.flags[i]]-offset], s.lines[s.tags[s.flags[i]]-offset+1:]...)
		offset++
	}
}

// parse out
func (seg segment) parse() ([]float64, []int) {
	seg.data = make([]float64, len(seg.lines))
	xBound := len(seg.lines)
	var tags []int
	for x := 0; x < xBound; x++ {
		r := regexp.MustCompile(`\W*\d+\W*\d*\s+\W*\d+\W*\d*\s+\W*\d+\W*\d*\s+(\W*\d+\W*\d*)\s+\W\s+\d+\W*\d*\s+\d\s+\d+\W*\d*\s+\d+\W*\d*`)
		line := seg.lines[x]
		if r.MatchString(line) {
			tags = append(tags, x)

			hold, err := strconv.ParseFloat(string(r.FindSubmatch([]byte(line))[1]), 64)
			seg.data[x] = hold
			if err != nil {
				fmt.Println(err)
			}

		}

	}
	return seg.data, tags

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
	file, err := os.Create(l.fileLocation)
	if err != nil {
		fmt.Println("could not open " + l.fileLocation)

	}
	writer := bufio.NewWriter(file)
	for i := 0; i < len(l.exp.Header); i++ {
		writer.WriteString(l.exp.Header[i])
	}

	for i := 0; i < len(l.exp.Segments); i++ {
		for x := 0; x < len(l.exp.Segments[i].lines); x++ {
			writer.WriteString(l.exp.Segments[i].lines[x])
		}

	}
	file.Close()
}

// listen for segments and place them in order
func (l listener) listen(length int, segChan chan segment, doneChan chan bool) {
	done := false
	i := 0
	for !done {
		select {
		case hold := <-segChan:
			l.exp.Segments[hold.index] = hold
			i++
			if i >= length {
				done = true
				l.rebuild()
				doneChan <- true
			}
		}
	}
}

func work(s segment) {
	s.data, s.tags = s.parse()
	s.sanatize()
	fmt.Printf("Segment: %d is done processing.\n", s.index)
	s.returnChan <- s
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
	doneChan := make(chan bool)
	list.exp = exp
	list.fileLocation = fileLocation
	exp.Header, _ = chomp(scanner, "Reflections measured after indexing")
	segChan := make(chan segment)
	// collect lines from file and turn them into segments
	done := false
	i := 0

	for !done {
		seg := *new(segment)
		seg.lines, done = chomp(scanner, "Reflections measured after indexing")
		seg.index = i
		seg.returnChan = segChan
		exp.Segments = append(exp.Segments, seg)
		i++

	}

	go list.listen(i, segChan, doneChan)
	for i := 0; i < len(list.exp.Segments); i++ {
		go work(list.exp.Segments[i])
	}
	<-doneChan
}
