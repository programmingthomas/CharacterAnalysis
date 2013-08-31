package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"image"
	"image/color"
	"image/png"
	"math"
	"bufio"
	"regexp"
	"strconv"
)

func SaveImage(name string, img image.Image) {
	fo, err := os.Create(name)
	if err != nil {
		fmt.Print(err)
	}
	w := bufio.NewWriter(fo)
	png.Encode(w, img)
	w.Flush()
	defer fo.Close()
}

func CreateImage(words []string, searchTerm string, sectionSize int, quit chan int) {
	searchTerms := strings.Fields(searchTerm)
	
	imageHeight := 20
	
	//Setup the image
	totalWords := len(words)
	imageWidth := int(math.Ceil(float64(totalWords) / float64(sectionSize)))
	newImg := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	
	usedInSection := make([]bool, imageWidth, imageWidth)
	
	usedCount := 0
	
	for index, word := range words {
		section := int(math.Floor(float64(index) / float64(sectionSize)))
		for _, term := range searchTerms {
			if word == term {
				usedInSection[section] = true
				usedCount++
				break
			}
		}
	}
	
	fmt.Println("The terms", searchTerm, "were found", usedCount, "times")
	
	for x := 0; x < imageWidth; x++ {
		color := color.RGBA{255, 255, 255, 255}
		if usedInSection[x] == true {
			color.R = 0
			color.G = 0
			color.B = 0
		}
		for y := 0; y < imageHeight; y++ {
			newImg.SetRGBA(x, y, color)
		}
	}
	
	SaveImage(searchTerm + ".png", newImg)
	
	quit <- 1
}

func RemovePunctuation(original string) string {
	regex, _ := regexp.Compile("[^a-zA-Z0-9 -]")
	return regex.ReplaceAllString(original, "")
}

func CreateHTML(filename string, characters []string) {
	bookName := strings.Replace(filename, ".txt", "", -1)
	htmlDocumentName := bookName + ".html"
	
	//Should use HTML templating in Go, but this is so simple I couldn't be bothered
	html := "<!DOCTYPE html><html><head><style>body { font-family:sans-serif; }</style><title>" + bookName + "</title></head><body><h1>" + bookName + " analysis</h1><table><tr><th>Character</th><th>Mentioned</th></tr>"
	
	for _, character := range characters {
		html = html + "<tr><td>" + character + "</td><td><img src='" + character + ".png' /></td></tr>"
	}
	html = html + "</table></body></html>"
	
	ioutil.WriteFile(htmlDocumentName, []byte(html), 0777)
}

func main() {
	args := os.Args;
	//The first argument is the executable name
	if len(args) > 3 {
		filename := args[1]
		sectionSize, intParseErr := strconv.ParseInt(args[2], 10, 32)
		
		if intParseErr != nil {
			fmt.Println(intParseErr)
			sectionSize = 100
		}
		
		fmt.Println("Going to analyse", filename)
		//Read the file
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		//Lowercase the string
		contents := strings.ToLower(string(bytes))
		//Remove 's
		contents = strings.Replace(contents, "'s", "", -1)
		//Remove punctuation
		contents = RemovePunctuation(contents)
		//The Fields function splits on whitespace and removes all whitespace :)
		words := strings.Fields(contents)
		fmt.Println("There are", len(words), "words")
		
		channels := make([]chan int, len(args))
		
		for c := 3; c < len(args); c++ {
			channels[c] = make(chan int)
			character := strings.ToLower(args[c])
			fmt.Println("Searching for", character)
			//A cast from type int64 to int is required here
			go CreateImage(words, character, int(sectionSize), channels[c])
		}
		
		//Write the analysis HTML document
		CreateHTML(filename, args[3:len(args)])
		
		//Wait for all goroutines to exit
		for c := 3; c < len(channels); c++ {
			<- channels[c]
		}
	} else {
		fmt.Println("Use as characters <filenameofbook.txt> <sectionsize> <character>")
		fmt.Println("Section size indicates how many words one pixel column should represent")
		fmt.Println("<character> can be a single word (sherlock) or two words (\"sherlock holmes\") which will find all matches of 'sherlock' or 'holmes'")
	}
}