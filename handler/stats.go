package handler

import (
	"fmt"
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
	keys := sortMapIntoSlice(commit)
	columns := buildColumns(keys, commit)
	printCells(columns)

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

func buildColumns(key []int, commits map[int]int) map[int][]int {
	columns := make(map[int][]int)
	var column []int

	for _, k := range key {
		week := int(k / 7)
		dayinweek := k % 7
		if dayinweek == 0 {
			column = make([]int, 0)
		}
		column = append(column, commits[k])
		if dayinweek == 0 {
			columns[week] = column
		}

	}
	return columns
}


func printCell(val int, today bool) {
    escape := "\033[0;37;30m"
    switch {
    case val > 0 && val < 5:
        escape = "\033[1;30;47m"
    case val >= 5 && val < 10:
        escape = "\033[1;30;43m"
    case val >= 10:
        escape = "\033[1;30;42m"
    }

    if today {
        escape = "\033[1;37;45m"
    }

    if val == 0 {
        fmt.Printf(escape + "  - " + "\033[0m")
        return
    }

    str := "  %d "
    switch {
    case val >= 10:
        str = " %d "
    case val >= 100:
        str = "%d "
    }

    fmt.Printf(escape+str+"\033[0m", val)
}

func printCells(cols map[int][]int) {
	printMonths()
	for j := 6; j >= 0; j-- {
		for i := weeksInLastSixMonths() + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths()+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				//special case today
				if i == 0 && j == calculateOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}
func getBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}
func printMonths() {
	week := getBeginningOfDay(time.Now()).Add(-time.Duration(daysInLastSixMonths()*int(time.Hour)*24))
	month := week.Month()
	fmt.Printf("         ")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(7 * time.Hour * 24)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

// printDayCol given the day number (0 is Sunday) prints the day name,
// alternating the rows (prints just 2,4,6)
func printDayCol(day int) {
	out := "     "
	switch day {
	case 1:
		out = " Mon "
	case 3:
		out = " Wed "
	case 5:
		out = " Fri "
	}

	fmt.Printf(out)
}
func ProcessRepositories(email string) map[int]int {
    filePath := GetDotfilePath()
    repos := parseFileLinesToString(filePath)
    daysInMap := daysInLastSixMonths()

    commits := make(map[int]int, daysInMap)
    for i := daysInMap; i > 0; i-- {
        commits[i] = 0
    }

    for _, path := range repos {
        commits = GetFillCommits(email, path, commits)
    }

    return commits
}
