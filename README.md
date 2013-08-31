#CharacterAnalysis

Analyses when characters appear in books

##Compilation
* Either use `go get` to download the contents of this repository or download characters.go using Git.
* `go build characters.go`

##Usage
* Download a text file for a book (Project Gutenberg is a good place to find these)
* If on Windows: `characters <filenameofbook> 100 character1 character2 character3 etc`
* If on Mac/Unix/Linux: `./characters <filenameofbook> 100 character1 character2 character3 etc`
* If you need to search for a character that may be referred to by two names remember you can put the name in quotes
* The second argument (a number) represents how many words each pixel represents (i.e. each black column in the final output represents that the character was mentioned at least once in those X words)
* An HTML file is created that shows all of the characters and their charts