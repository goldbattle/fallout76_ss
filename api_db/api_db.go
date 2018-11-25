package api_db

import (
	"database/sql"
	"fmt"
	"github.com/ararog/timeago"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Location of our database
var database_loc = "./dbs/testing.db"

// This should open the database connection
// It will also make sure that our database is in the correct format!
// If it is missing tables, it will create the tables needed...
func OpenDatabase() *sql.DB {

	// Make sure the directory is there
	dir, _ := filepath.Split(database_loc)
	os.MkdirAll(dir, os.ModePerm)

	// Open the database
	//start := time.Now()
	db, err := sql.Open("sqlite3", database_loc)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	// Create the table if needed
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS timing_1sec
	(
	 id INTEGER PRIMARY KEY AUTOINCREMENT,
	 apitype string,
	 status INTEGER,
	 rawstatus TEXT,
	 ms_dns FLOAT,
	 ms_tcp FLOAT,
	 ms_tls FLOAT,
	 ms_server FLOAT,
	 ms_content FLOAT,
	 ms_total FLOAT,
	 unixtime DATE
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Errorf("[database]: %v", err)
		fmt.Errorf("[database]: %v", sqlStmt)
	}

	// return this database
	//fmt.Printf("[database]: Opened database (took %s)\n", time.Since(start))
	return db

}

// This will take a api query result, and insert it into the database timing
// This always inserts it into the main timing database (ie. 1 second polling)
func InsertTiming(db *sql.DB,
	apitype string, status int, rawstatus string,
	ms_dns float64, ms_tcp float64, ms_tls float64,
	ms_server float64, ms_content float64, ms_total float64) {

	// beginning the commit message
	tx, err := db.Begin()
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	// Prepare the row insertion
	stmt, err := tx.Prepare("INSERT INTO timing_1sec(apitype, status, rawstatus, ms_dns, ms_tcp, ms_tls, ms_server, ms_content, ms_total, unixtime) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}
	defer stmt.Close()

	// Execute and commit the message
	_, err = stmt.Exec(apitype, status, rawstatus, ms_dns, ms_tcp, ms_tls, ms_server, ms_content, ms_total, time.Now())
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	// Commit this change to the database
	err = tx.Commit()
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

}

// This will get the newest entry in our database and report it to the website
// This will be what the big "online" or "offline" status on the homepage is
func GetCurrentStatus(db *sql.DB) (int, string) {

	// Query the current database, and get the newest
	//stmt, err := db.Query("SELECT status, unixtime FROM timing_1sec where id = (SELECT MAX(id) FROM timing_1sec)")
	stmt, err := db.Query("SELECT id, status, unixtime FROM timing_1sec ORDER BY unixtime DESC LIMIT 1")
	if err != nil {
		fmt.Errorf("%v", err)
		return 0, "Unknown"
	}
	defer stmt.Close()

	// Get the result
	var id int
	var status int
	var timestamp time.Time
	stmt.Next()
	err = stmt.Scan(&id, &status, &timestamp)
	if err != nil {
		fmt.Errorf("%v", err)
		return 0, "Unknown"
	}

	// Calculate duration
	//duration := time.Now().Sub(timestamp)
	//fmt.Printf("%v\n",duration)

	// Parse the time ago status
	str_timeago, err := timeago.TimeAgoWithTime(time.Now(), timestamp)
	if err != nil {
		fmt.Errorf("%v", err)
		return 0, "Unknown"
	}

	// Get the result and return
	return status, str_timeago
}

// This will get the status for the specified day
// Should loop through and see what the uptime/downtime is for this day
// If there is more then 15 minutes of downtime, then this is a BAD day so flag it
func GetDayStatus(db *sql.DB, day int, month time.Month, year int) StatusDay {

	// Get the current day, and the next day
	dt_cur := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	dt_nex := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)

	// Query the current database, and get the newest
	stmt, err := db.Query("SELECT id, status, unixtime FROM timing_1sec WHERE unixtime BETWEEN ? AND ? ORDER BY unixtime DESC", dt_cur, dt_nex)
	if err != nil {
		fmt.Errorf("%v", err)
		return StatusDay{}
	}
	defer stmt.Close()

	// Our total uptime and downtime times
	var ct_rows int
	var time_up, time_down float64

	// Get the first ever timestamp of this range
	var id, status int
	var time_prev time.Time
	stmt.Next()
	err = stmt.Scan(&id, &status, &time_prev)
	if err != nil {
		fmt.Errorf("%v", err)
	}

	// Loop through all entries for this day
	for stmt.Next() {

		// Get the result
		var id, status int
		var timestamp time.Time
		err = stmt.Scan(&id, &status, &timestamp)
		if err != nil {
			fmt.Errorf("%v", err)
			return StatusDay{}
		}

		// Add to uptime/downtime
		if status == 1 {
			time_up = time_up + time_prev.Sub(timestamp).Minutes()
		}
		if status == 2 {
			time_down = time_down + time_prev.Sub(timestamp).Minutes()
		}

		// Move forward in time
		time_prev = timestamp
		ct_rows = ct_rows + 1
	}

	// Debug print
	//fmt.Printf("%v - %.2f up | %.2f down | %d rows\n", dt_cur, time_up, time_down, ct_rows)

	// Create the status day
	var statusday StatusDay
	statusday.AmountMinuteDown = time_down
	statusday.AmountMinuteUp = time_up

	// If we do not have any rows, then we where not able to query the API...
	// If less then 15 minutes downtime, then we have a good day!
	statusday.StatusOnline = (ct_rows != 0 && time_down < 15)
	statusday.StatusOffline = (ct_rows != 0 && time_down >= 15)
	statusday.StatusUnknown = (ct_rows == 0)

	// return
	return statusday
}

// This will query our database and get all information needed for the homepage template
// So we need to get the current status of the
func GetHomepageData(db *sql.DB) HomepageData {

	// Create the data type we will return
	data := HomepageData{}

	status, strtime := GetCurrentStatus(db)

	// Set the current setting
	data.StatusOnline = (status == 1)
	data.StatusOffline = (status == 2)
	data.StatusUnknown = (status != 1 && status != 2)
	data.TimeAgoString = strtime

	// Get the current month
	year, month, _ := time.Now().Date()
	//fmt.Printf("Current month: [%v]\n", month)

	// Loop through the past 6 months
	for i := 0; i < 9; i++ {

		// Calculate weekday
		// Note we add one since the weekday starts on MONDAY and not SUNDAY
		t_cur := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
		weekday := (int(t_cur.Weekday()) + 1) % 7
		days_empty := make([]int, weekday)
		//fmt.Printf("Weekday %v [%v]\n", weekday, month)

		// Get the number of days of the current month
		t_new := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
		//fmt.Printf("Total number of days in [%v], [%v] is [%v]\n", month, year, t_new.Day())

		// Loop through each day and get the status
		var daystatus []StatusDay
		for day := 1; day <= t_new.Day(); day++ {
			daystatus = append(daystatus, GetDayStatus(db, day, month, year))
		}

		// Append to our data object this new month!
		datamonth := StatusMonth{
			Name:      month.String() + " " + strconv.Itoa(year),
			DaysEmpty: days_empty,
			Days:      daystatus,
		}
		data.Months = append(data.Months, datamonth)

		// Loop through and if we reach december, then we reset to january
		// If we reach January we should loop around to december (if we are going in reverse)
		month = month - 1
		if month > time.December {
			month = time.January
			year = year + 1
		}
		if month < time.January {
			month = time.December
			year = year - 1
		}

	}

	//// loop each day of the month
	//for day := 1; day <= t.Day(); day++ {
	//	// do whatever you want here ...
	//	fmt.Println(day, month, year)
	//}

	// return the data
	return data

}
