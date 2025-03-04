package handler

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RecursivelyScanFolder(folder string) []string {
	return gitScanFolder(make([]string, 0), folder)
}

func AddNewSliceToFile(filePath string, newRepo []string) {
	existingRepos := parseFileLinesToString(filePath)
	repos := joinRepos(newRepo, existingRepos)
	dunpStringToFile(repos, filePath)
}

func dunpStringToFile(stringSlice []string, filePath string) {
	fmt.Println("... Writing to file")
	fmt.Println(stringSlice)
	content := strings.Join(stringSlice, " \n")
	fmt.Println(content)
	ioutil.WriteFile(filePath, []byte(content), 0755)

}

func joinRepos(fromR []string, toR []string) []string {
	return JoinSlice(fromR, toR)
}

func JoinSlice(fromS []string, toS []string) []string {
	println(len(fromS))
	println(len(toS))
	for _, value := range toS {
		for _, sliceValue := range fromS {
			if sliceValue != value {
				toS = append(fromS, sliceValue)

			}
		}
	}
	return fromS
}

func GetDotfilePath() string {
	// usr, err := user.Current()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// dotFile := usr.HomeDir + "/.gogitlocalstatus"

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dotFile := filepath.Join(currentDir, ".gogitlocalstatus")
	return dotFile

}

func parseFileLinesToString(file string) []string {
	filePtr, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("... Creating file ", file)
			filePtr, err = os.Create(file)
			return []string{}
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	defer filePtr.Close()
	var lines []string
	scanner := bufio.NewScanner(filePtr)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

func getIgnorefile() []string {
	currenDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	filePath := filepath.Join(currenDir, "handler/.ignore")
	return parseFileLinesToString(filePath)
}

func shouldIgnorefile(fileName string) bool {
	for _, ignore := range getIgnorefile() {
		matched, err := filepath.Match(ignore, fileName)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			fmt.Println("... Ignoring file ", fileName)
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

			if shouldIgnorefile(file.Name()) {
				continue
			}

			folders = gitScanFolder(folders, path)

		}

	}

	return folders

}
