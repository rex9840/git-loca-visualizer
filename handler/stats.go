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
	limit := daysInLastSixMonths() // cutoff

	err = iter.ForEach(func(c *object.Commit) error {
		// Normalize to local midnight (same location as now)
		d := getBeginningOfDay(c.Author.When.In(now.Location()))
		daysAgo := int(now.Sub(d).Hours() / 24) // <0 => future; >limit => too old

		if daysAgo < 0 || daysAgo > limit {
			return nil
		}
		if c.Author.Email == email {
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
	// Find the oldest (max) week index present
	lastWeek := 0
	for w := range cols {
		if w > lastWeek {
			lastWeek = w
		}
	}

	printMonths(lastWeek)

	// Rows: Sun(0) .. Sat(6); Columns: oldest → current (left → right)
	for row := 0; row <= 6; row++ {
		for w := lastWeek; w >= 0; w-- {
			if w == lastWeek {
				printDayCol(row)
			}

			if col, ok := cols[w]; ok && row < len(col) {
				isToday := (w == 0 && row == calculateOffset())
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

func printMonths(lastWeek int) {
	now := getBeginningOfDay(time.Now())
	startOfThisWeek := now.AddDate(0, 0, -int(time.Now().Weekday()))
	oldestSunday := startOfThisWeek.AddDate(0, 0, -7*lastWeek)

	fmt.Printf("     ") // left padding under day labels

	curMonth := oldestSunday.Month()
	weekSunday := oldestSunday

	for w := lastWeek; w >= 0; w-- {
		if weekSunday.Month() != curMonth || w == lastWeek {
			fmt.Printf("%-4s", weekSunday.Month().String()[:3])
			curMonth = weekSunday.Month()
		} else {
			fmt.Printf("    ")
		}
		weekSunday = weekSunday.AddDate(0, 0, 7)
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
func ProcessRepositories(email string) map[int]int {
	filePath := GetDotfilePath()
	repos := parseFileLinesToString(filePath)

	days := daysInLastSixMonths()
	commits := make(map[int]int, days+14)
	for i := 0; i <= days+14; i++ {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = GetFillCommits(email, path, commits)
	}

	return commits
}
