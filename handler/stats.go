package handler

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"sort"
	"time"
)

const outOfRange int = -1

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
	offset := calculateOffset()

	err = commiIterattorFromHead.ForEach(func(c *object.Commit) error {
		daysAgo := countDaysSince(c.Author.When) + offset

		if c.Author.Email == email && daysAgo != outOfRange {
			commits[daysAgo]++
		}
		return nil
	})
	errorPanic(err)
	return commits

}

func PrintCommitsStats(commit map[int]int) {
        
        
}

func sortMapIntoSlice(m map[int]int) (sortedKeys []int) {
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Ints(sortedKeys)
	return
}

func daysInLastSixMonths() (days int) {
	now := time.Now()
	sixMonthAgo := now.AddDate(0, -6, 0)
	days = int(now.Sub(sixMonthAgo).Hours() / 24)
	return
}

func weeksInLastSixMonths() (weeks int) {
	days := daysInLastSixMonths()
	weeks = int(days / 7)
	return
}

func countDaysSince(date time.Time) (days int) {
	days = 0
	year, month, day := time.Now().Date()
	now := time.Date(year, month, day, 0, 0, 0, 0, &time.Location{})
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths() {
			return outOfRange
		}

	}
	return

}

func calculateOffset() (offset int) {
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
