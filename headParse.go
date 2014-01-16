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
	Head,
	Foot segment
	Segments []segment
}

var numWorkers = flag.Int("n", 4, "the number of processors availble to the program")
var f = flag.String("f", "rharr-c453.head", "the path to the file to be parsed")

type segment struct {
	index int
	lines []string
	I,
	sigmaI,
	iOverSigmaI float32
}

// take in a scanner, and return a segment of the text file you are looking for
func chomp(scanner *bufio.Reader, seg *segment, lookFor string) (*segment, bool) {
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
			seg.lines = lines
			return seg, true
		}
	}
	seg.lines = lines
	return seg, false
}
func parse(line string) map[string]interface{} {
	b := strings.Split(line, " ")
	hold := make(map[string]interface{})

	hold["h"], _ = strconv.Atoi(b[0])
	hold["k"], _ = strconv.Atoi(b[1])
	hold["l"], _ = strconv.Atoi(b[2])
	hold["I"], _ = strconv.ParseFloat(b[3], 32)
	hold["sigmaI"], _ = strconv.ParseFloat(b[5], 32)

	return hold
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
	var seg segment
	seg.index = 0
	i := 0
	done := false
	for !done {
		seg := new(segment)
		seg.index = i
		seg, done = chomp(scanner, seg, "Reflections measured after indexing")
		fmt.Println(seg.index)
		i++
		for x := 0; x < len(seg.lines); x++ {
			fmt.Println(parse(seg.lines[x])["h"])
		}
	}

}
