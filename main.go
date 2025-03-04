package main

import (
	"flag"
	"fmt"
	"rex9840/git-contribution/handler"
)

func scan(path string) {
	fmt.Println("... FOUND FOLDER")
	repo := handler.RecursivelyScanFolder(path)
	fmt.Printf("%s\n", repo)
        filePath:=handler.GetDotfilePath()
        fmt.Printf("%s\n",filePath)
        handler.AddNewSliceToFile(filePath,repo)
        fmt.Printf("... Added to %s file\n",filePath)        
}

func stats(email string) {
	println("stats")
        fmt.Println(handler.CalculateOffset())
	println(email)

}



func main() {
          
	var folder string
	var email string
	flag.StringVar(&folder, "add", "", "Add a folder to the scan list of your git repository")
	flag.StringVar(&email, "stats", "your@email.com", "Get the stats of the email")
	flag.Parse()

	if folder != "" {
		scan(folder)
	}

}
