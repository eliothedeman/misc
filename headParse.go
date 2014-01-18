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
	Header,
	Footer,
	lines []string
	I,
	sigmaI,
	iOverSigmaI float32
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
	exp.Header, _ = chomp(scanner, "Reflections measured after indexing")
	fmt.Println(exp.Header)
}
