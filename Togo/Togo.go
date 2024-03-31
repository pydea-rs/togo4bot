package Togo

import (
	// chrono "github.com/gochrono/chrono"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq" // postgres
)

var lastUsedId uint64 = 0

// ---------------------- Date/Time Struct & Date Receivers --------------------------------
type Date struct{ time.Time }

func (d *Date) Get() string {

	return fmt.Sprintf("%d-%d-%d\t%d:%d", d.Year(), d.Month(), d.Day(), d.Hour(), d.Minute())
}
func (d *Date) Short() string {

	return fmt.Sprintf("%d-%d-%d", d.Year(), d.Month(), d.Day())
}

func Now() time.Time {
	if timeZone, err := time.LoadLocation("Asia/Tehran"); err == nil {
		return time.Now().In(timeZone)
	}
	return time.Now()
}

func Today() Date {
	return Date{Now()}
}

// ---------------------- Togo Struct & Togo Receivers--------------------------------
type Togo struct {
	Id          uint64
	Title       string
	Description string
	Weight      uint16
	Progress    uint8
	Extra       bool
	Date        Date
	Duration    time.Duration
	OwnerId     int64 // telegram id
}

func (togo *Togo) Save() uint64 {
	/*const CREATE_TABLE_QUERY string = `CREATE TABLE IF NOT EXISTS togos (id SERIAL PRIMARY KEY, owner_id BIGINT NOT NULL,
	title VARCHAR(64) NOT NULL, description VARCHAR(256), weight INTEGER, extra INTEGER,
	progress INTEGER, date timestamp with time zone, duration INTEGER)`*/

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}
	defer db.Close()
	/*if _, err := db.Exec(CREATE_TABLE_QUERY); err != nil {
		panic(err)
	}*/
	extra := 0
	if togo.Extra {
		extra = 1
	}
	if res, err := db.Exec("INSERT INTO togos (owner_id, title, description, weight, extra, progress, date, duration) VALUES ($1, $2::varchar, $3::varchar, $4, $5, $6, $7, $8)",
		togo.OwnerId, togo.Title, togo.Description, togo.Weight, extra, togo.Progress,
		togo.Date.Time, togo.Duration.Minutes()); err != nil {
		panic(err)
	} else if id, e := res.LastInsertId(); e == nil {
		return uint64(id)
	}
	return 0
}

func isCommand(term string) bool {
	return term == "+" || term == "%" || term == "#" || term == "$"
}

func (togo *Togo) setFields(terms []string) {
	numOfTerms := len(terms)
	for i := 1; i < numOfTerms && !isCommand(terms[i]); i++ {
		switch terms[i] {
		case "=", "+w":
			i++

			if _, err := fmt.Sscan(terms[i], &togo.Weight); err != nil {
				panic(err)
			}

		case ":", "+d":
			i++
			togo.Description = terms[i]
		case "+x":
			togo.Extra = true
		case "-x":
			togo.Extra = false
		case "+p":
			i++

			if _, err := fmt.Sscan(terms[i], &togo.Progress); err != nil {
				panic(err)
			} else if togo.Progress > 100 {
				togo.Progress = 100
			}
		case "@":
			// im++
			i++
			today := Now()
			var delta int
			if _, err := fmt.Sscan(terms[i], &delta); err != nil {
				panic(err)
			}
			today = today.AddDate(0, 0, delta)
			i++
			temp := strings.Split(terms[i], ":")
			var hour, min int
			if _, err := fmt.Sscan(temp[0], &hour); err != nil {
				panic(err)
			} else if hour >= 24 || hour < 0 {
				panic("Hour part must be between 0 and 23!")
			}
			if _, err := fmt.Sscan(temp[1], &min); err != nil {
				panic(err)
			} else if min >= 60 || min < 0 {
				panic("Minute part must be between 0 and 59!")
			}
			if locale, err := time.LoadLocation("Asia/Tehran"); err == nil {

				togo.Date = Date{time.Date(today.Year(), today.Month(), today.Day(), hour, min, 0, 0, locale)}
			} else {
				togo.Date = Date{time.Date(today.Year(), today.Month(), today.Day(), hour, min, 0, 0, time.Local)}

			}
			// get the actual date here
		case "->":
			i++
			if _, err := fmt.Sscan(terms[i], &togo.Duration); err != nil {
				panic(err)
			} else if togo.Duration > 0 {
				togo.Duration *= time.Minute
			} else {
				panic("Duration must be positive integer!")
			}
		}

	}

}

func (togo *Togo) Update(ownerID int64) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}
	defer db.Close()

	extra := 0
	if togo.Extra {
		extra = 1
	}
	if _, err := db.Exec("UPDATE togos SET description=$1, weight=$2, extra=$3, progress=$4, date=$5, duration=$6 WHERE id=$7 AND owner_id=$8", // TODO: check ownerId? (no need)
		togo.Description, togo.Weight, extra, togo.Progress, togo.Date.Time, togo.Duration.Minutes(), togo.Id, ownerID); err != nil {
		panic(err)
	}
}

func (togo *Togo) ToString() string {
	return fmt.Sprintf("Togo #%d) %s:\t%s\nWeight: %d\nExtra: %t\nProgress: %d\nAt: %s, about %.1f minutes",
		togo.Id, togo.Title, togo.Description, togo.Weight, togo.Extra, togo.Progress, togo.Date.Get(), togo.Duration.Minutes())
}

// ---------------------- TogoList Type & Togo Receivers--------------------------------
type TogoList []Togo

func (togos TogoList) ToString() (result []string) {
	//result = "- - - - - - - - - - - - - - - - - - - - - -"
	for i := range togos {
		result = append(result, togos[i].ToString())
	}
	return
}

func (togos TogoList) Add(newTogo *Togo) TogoList {
	return append(togos, *newTogo)
}

func (togos TogoList) ProgressMade() (progress float64, completedInPercent float64, completed uint64, extra uint64, total uint64) {
	totalInPercent := uint64(0)
	for i := range togos {
		progress += float64(togos[i].Progress) * float64(togos[i].Weight)
		if togos[i].Progress == 100 {
			completed++
			completedInPercent += float64(togos[i].Progress) * float64(togos[i].Weight)
		}
		if !togos[i].Extra {
			totalInPercent += uint64(100 * togos[i].Weight)
			total++
		} else {
			extra++
		}
	}
	if totalInPercent > 0 {
		progress *= 100 / float64(totalInPercent) // CHECK IF IT CALCULAFES DECIMAL PART OR NOT
		completedInPercent *= 100 / float64(totalInPercent)
	}
	return
}

func (togos TogoList) Update(chatID int64, terms []string) string {
	var id uint64
	if _, err := fmt.Sscan(terms[0], &id); err != nil {
		panic(err)
	}
	targetIdx := -1
	// TODO: use simple version of FOR
	for i := range togos {
		if togos[i].Id == id {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return "There is no togo with this Id!"
	}
	if len(terms) > 1 && !isCommand(terms[1]) {

		togos[targetIdx].setFields(terms)
		togos[targetIdx].Update(chatID)
	}

	return togos[targetIdx].ToString()
}
func (togos *TogoList) RemoveIndex(index int) {
	list := *togos
	if len(list)-1 > index {
		list = append(list[:index], list[index+1])
	} else {
		list = append(list[:index])
	}
	*togos = list
}

func (togos TogoList) Remove(ownerID int64, togoID uint64) error {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		return err
	}

	defer db.Close()

	if _, err := db.Exec("DELETE FROM togos WHERE id=$1 AND owner_id=$2", togoID, ownerID); err != nil {
		return err
	}
	for i := range togos {
		if togos[i].Id == togoID && togos[i].OwnerId == ownerID {
			togos.RemoveIndex(i)
			break
		}
	}
	return nil
}

func (togos TogoList) Get(togoID uint64) (*Togo, error) {
	// TODO: use simple version of FOR
	for i := range togos {
		if togos[i].Id == togoID {
			return &togos[i], nil
		}
	}
	return nil, errors.New("Can not find this togo!")
}

// ---------------------- Shared Functions --------------------------------
func Load(ownerId int64, justToday bool) (togos TogoList, err error) {

	togos = make(TogoList, 0)
	err = nil
	if db, e := sql.Open("postgres", os.Getenv("POSTGRES_URL")); e == nil {
		defer db.Close()
		// ***** BETTER ALGORITHM
		// FIRST GET THE COUNT OF ROWS, then create a slice of that size and then load into that.
		const SELECT_QUERY string = "SELECT id, owner_id, title, description, weight, extra, progress, date, duration FROM togos WHERE owner_id=$1 ORDER BY date"
		/* if justToday {
			today := Date{time.Now()}
			next := Date{today.AddDate(0, 0, 1)}
			fmt.Println(next.Short())
			SELECT_QUERY = fmt.Sprintf("%s WHERE date >= DATETIME(%s)", SELECT_QUERY, today.Short())//, next.Short())
			fmt.Println(SELECT_QUERY)
		}*/
		rows, e := db.Query(SELECT_QUERY, ownerId)
		if e != nil {
			err = e
			return
		}

		now := Today()
		for rows.Next() {
			var togo Togo
			var date time.Time

			err = rows.Scan(&togo.Id, &togo.OwnerId, &togo.Title, &togo.Description, &togo.Weight, &togo.Extra, &togo.Progress, &date, &togo.Duration)
			if lastUsedId < togo.Id {
				lastUsedId = togo.Id
			}
			if timeZone, err := time.LoadLocation("Asia/Tehran"); err == nil {
				togo.Date = Date{date.In(timeZone)}
			} else {
				togo.Date = Date{date}
			}
			togo.Duration *= time.Minute
			if err != nil {
				panic(err)
			}
			if togo.Date.Short() == now.Short() {
				togos = togos.Add(&togo)
			} else if !justToday {
				togos = togos.Add(&togo)
			}
		}

	} else {
		err = e
	}
	return
}

func Extract(ownerId int64, terms []string) (togo Togo) {
	// setting default values
	if togo.Title = terms[0]; togo.Title == "" {
		togo.Title = "Untitled"
	}
	togo.OwnerId = ownerId
	togo.Weight = 1
	togo.Date = Today()
	(&togo).setFields(terms)
	return
}
