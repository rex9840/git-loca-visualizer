package handler

import (
	"fmt"
	"sort"
	"time"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const outOfRange = -1

func errorPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func GetFillCommits(email, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	errorPanic(err)

	hRef, err := repo.Head()
	errorPanic(err)

	iter, err := repo.Log(&git.LogOptions{From: hRef.Hash()})
	errorPanic(err)

	offset := calculateOffset() // 0..6, Sunday=0

	now := getBeginningOfDay(time.Now())
	limit := daysInLastSixMonths()

	err = iter.ForEach(func(c *object.Commit) error {
		// Normalize commit time to beginning of day in the same location as now
		d := getBeginningOfDay(c.Author.When.In(now.Location()))
		daysAgo := int(now.Sub(d).Hours() / 24) // negative => future; positive => past

		// outside the window?
		if daysAgo < 0 || daysAgo > limit {
			return nil
		}

		if c.Author.Email == email {
			// Shift by weekday so rows/columns align like a contributions calendar
			key := daysAgo + offset
			commits[key]++
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

func daysInLastSixMonths() int {
	now := getBeginningOfDay(time.Now())
	sixMonthsAgo := getBeginningOfDay(now.AddDate(0, -6, 0))
	return int(now.Sub(sixMonthsAgo).Hours() / 24)
}

func weeksInLastSixMonths() int {
	return daysInLastSixMonths() / 7
}

// calculateOffset returns 0..6 with Sunday=0, Monday=1, ... Saturday=6
func calculateOffset() int {
	return int(time.Now().Weekday())
}

func buildColumns(keys []int, commits map[int]int) map[int][]int {
	columns := make(map[int][]int)

	for _, k := range keys {
		week := k / 7
		dayInWeek := k % 7

		if _, ok := columns[week]; !ok {
			columns[week] = make([]int, 7) // always 7 slots (Sun..Sat)
		}
		columns[week][dayInWeek] = commits[k]
	}
	return columns
}

func printCell(val int, today bool) {
	escape := "\033[0;37;30m" // default

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
	case val >= 100:
		str = "%d "
	case val >= 10:
		str = " %d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

func printCells(cols map[int][]int) {
	printMonths()

	// rows: Sun(0) .. Sat(6) rendered top→bottom as in your original
	for row := 0; row <= 6; row++ {
		// columns: from oldest week to current week (left→right)
		for w := 0; w <= weeksInLastSixMonths()+1; w++ {
			if w == 0 {
				printDayCol(row)
			}

			if col, ok := cols[w]; ok && row < len(col) {
				// Today highlight: current week is the last column,
				// and today's row equals calculateOffset()
				isToday := (w == 0 && row == calculateOffset()) // we render newest at w==0 below
				printCell(col[row], isToday)
			} else {
				printCell(0, false)
			}
		}
		fmt.Printf("\n")
	}
}

func getBeginningOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func printMonths() {
	start := getBeginningOfDay(time.Now()).
		Add(-time.Duration(daysInLastSixMonths()) * 24 * time.Hour)

	cur := start
	curMonth := cur.Month()

	fmt.Printf("         ")
	for {
		if cur.Month() != curMonth {
			fmt.Printf("%s ", cur.Month().String()[:3])
			curMonth = cur.Month()
		} else {
			fmt.Printf("    ")
		}

		cur = cur.Add(7 * 24 * time.Hour)
		if cur.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func printDayCol(day int) {
	out := "     "
	switch day {
	case 0:
		out = "Sun  "
	case 1:
		out = "Mon  "
	case 2:
		out = "Tue  "
	case 3:
		out = "Wed  "
	case 4:
		out = "Thu  "
	case 5:
		out = "Fri  "
	case 6:
		out = "Sat  "
	}
	fmt.Printf(out)
}

// ProcessRepositories reads repo paths (from your dotfile helpers) and aggregates commit counts.
// NOTE: GetDotfilePath and parseFileLinesToString are assumed to exist in your codebase.
func ProcessRepositories(email string) map[int]int {
	filePath := GetDotfilePath()
	repos := parseFileLinesToString(filePath)

	days := daysInLastSixMonths()

	// Pre-size with a little extra to hold offset shift + safety margin
	commits := make(map[int]int, days+14)
	for i := 0; i <= days+14; i++ {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = GetFillCommits(email, path, commits)
	}

	return commits
}
