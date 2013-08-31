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

//Takes a filename and an image and saves it to a file as a PNG (syncronhously)
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

//Takes a list of words (i.e. from a book), finds words that match the search term,
//and creates an image based on the usage of those words with each pixel representing
//the specified number of words (sectionSize). The channel will be sent a value of 1
//upon the successful completion of the task
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

//Removes any non alphanumeric (and dash + space) characters from a string
func RemovePunctuation(original string) string {
	regex, _ := regexp.Compile("[^a-zA-Z0-9 -]")
	return regex.ReplaceAllString(original, "")
}

//Takes a filename of a book (text file) that has been analysed along
//with a list of characters and outputs an HTML file with a list of those
//characters along with the images of the appropriate appearances
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

//This is the main function. It accepts a minimum of three command line arguments.
//The first argument should be the filename of the text file that you wish to analyse
//The second argument should be the number of words that each pixel should represent in
//the final image. A good number for a short book (like the first Harry Potter) might be 150 - 250
//however for a longer novel (like the fifth Harry Potter) 500 - 1000 may be more appropriate.
//If you give a value of 0 or less it will automatically use the total number of words, which could
//generate a VERY LARGE image. The default height for the output images is 20 pixels, you can change
//this in the CreateImage function.
//The following arguments should be the names of characters that you wish to find in the book.
//Remember that if you want to find a character that is referred to by the forename and surname
//you can format your command line arguments as "firstname surname". Any word that matches either of
//the names will be deemed as referring to the character. For example, if analysing Harry Potter it
//may be better to refer to the titular character as 'harry' rather than 'harry potter' because the
//search will find all references to his parents and children as well (whilst there are no other
//characters in the books called harry).
//And remember, you don't have to just stick to character's names!
func main() {
	args := os.Args;
	//The first argument is the executable name
	if len(args) > 3 {
		filename := args[1]
		//Parse the <sectionsize> argument as an int64 in base 10
		sectionSize, intParseErr := strconv.ParseInt(args[2], 10, 64)
		
		//Set the default to a value of 100 if parsing fails
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
		
		//
		if sectionSize <= 0 {
			sectionSize = len(words)
		}
		
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