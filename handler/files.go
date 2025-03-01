package handler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
        "path/filepath"
	"strings"
)

func RecursivelyScanFolder(folder string) []string {
	return gitScanFolder(make([]string, 0), folder)
}

func AddNewSlicePath(filePath string, newRepo []string) {
}

func GetDotfilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	dotFile := usr.HomeDir + "/.gogitlocalstatus"

	return dotFile

}


func getIgnorefile()[]string{
        currenDir,err:= os.Getwd() 
        if err != nil { 
                log.Fatal(err) 
        }
        filePath:= filepath.Join(currenDir,"handler/.ignore")
        filePtr,err := os.Open(filePath)
        if err != nil {
                log.Fatal(err)
        } 
         
        var ignoreList []string
        
        scanner := bufio.NewScanner(filePtr)
        for scanner.Scan() { 
                line:= strings.TrimSpace(scanner.Text())
                if line != "" {
                        ignoreList = append(ignoreList,line) 
                }

        }
        
        return ignoreList
}


func shouldIgnorefile(fileName string )bool{
        for _ , ignore := range getIgnorefile(){
                matched, err := filepath.Match(ignore, fileName)
                if err != nil {
                        log.Fatal(err)
                } 
                if matched { 
                        fmt.Println("... Ignoring file ",fileName)
                        return true
                } 
        }
        return false
}


func gitScanFolder(folders []string, folder string) []string {
	folder = strings.TrimSuffix(folder, "/")

	folderPtr, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}

	files, err := folderPtr.Readdir(-1)
	folderPtr.Close()

	if err != nil {
		log.Fatal(err)
	}

	path := ""
	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			if file.Name() == ".git" {
				fmt.Println("... FOUND GIT FOLDER")
				path = strings.TrimSuffix(path, "/.git")
				folders = append(folders, path)
				fmt.Printf("%s\n", folder)
				continue
			}
                        
                        if shouldIgnorefile(file.Name()){
                                continue 
                        }

			folders = gitScanFolder(folders, path)
                        
                
		}  

	}

	return folders

}
