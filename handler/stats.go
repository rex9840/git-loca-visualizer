package handler

import (
	"time"
	"gopkg.in/src-d/go-git.v4"
)

func errorPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func GetFillCommits(email string, path string, commits map[int]int) map[int]int {

	repo, err := git.PlainOpen(path)
	errorPanic(err)
	hRef, err := repo.Head()
	errorPanic(err)
	commiIterattorFromHead, err := repo.Log(&git.LogOptions{From: hRef.Hash()})
	errorPanic(err)

	_ = commiIterattorFromHead
	return map[int]int{}

}

func CalculateOffset() (offset int) {
	// calculate the offset of the days and return the missing days
	weekdays := time.Now().Weekday()
	switch weekdays {
	case time.Sunday:
		offset = 7

	case time.Monday:
		offset = 6

	case time.Tuesday:
		offset = 5

	case time.Wednesday:
		offset = 4

	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}

	return
}
