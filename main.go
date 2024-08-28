package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "./sample.db?_foreign_keys=on"

func execInTx(tx *sql.Tx, query string) {
	_, err := tx.Exec(query)
	if err != nil {
		log.Printf("%q: %s\n", err, query)
		log.Fatal(err)
	}
}

func createDatabase(db *sql.DB) {
	initialTxQuery := `
	create table Weekdays (
		WeekdayID integer primary key autoincrement,
		WeekdayName text not null unique
	);
	create table Months (
		MonthID integer primary key autoincrement,
		MonthName text not null unique
	);
	create table Dates (
		DateID integer primary key autoincrement,
		Month integer not null,
		Day integer not null,
		Year integer not null,
		Weekday text not null,
		foreign key (Month) references Months(MonthID),
		foreign key (Weekday) references Weekdays(WeekdayName)
	);
	create table Users (
		UserName text primary key,
		Password blob(64) not null
	) without rowid;
	create table Volunteers (
		VolunteerID integer primary key autoincrement,
		VolunteerName text not null,
		User text,
		foreign key (User) references Users(UserName)
	);
	create table Schedules (
		ScheduleID integer primary key autoincrement,
		ScheduleName text not null,
		ShiftsOff integer not null,
		VolunteersPerShift integer not null,
		User text,
		StartDate integer,
		EndDate integer,
		foreign key (User) references Users(UserName),
		foreign key (StartDate) references Dates(DateID),
		foreign key (EndDate) references Dates(DateID)
	);
	create table WeekdaysForSchedule (
		WFSID integer primary key autoincrement,
		User text,
		Weekday text,
		Schedule integer,
		foreign key (User) references Users(UserName),
		foreign key (Weekday) references Weekdays(WeekdayName),
		foreign key (Schedule) references Schedules(ScheduleID)
	);
	create table VolunteersForSchedule (
		VFSID integer primary key autoincrement,
		User text,
		Schedule integer,
		Volunteer integer,
		foreign key (User) references Users(UserName),
		foreign key (Schedule) references Schedules(ScheduleID),
		foreign key (Volunteer) references Volunteers(VolunteerID)
	);
	create table UnavailabilitiesForSchedule (
		UFSID integer primary key autoincrement,
		User text,
		VolunteerForSchedule integer,
		Date integer,
		foreign key (User) references Users(UserName),
		foreign key (VolunteerForSchedule) references VolunteersForSchedule(VFSID),
		foreign key (Date) references Dates(DateID)
	);
	create table CompletedSchedules (
		CScheduleID integer primary key autoincrement,
		ScheduleData text not null,
		User text,
		Schedule integer,
		foreign key (User) references Users(UserName),
		foreign key (Schedule) references Schedules(ScheduleID)
	);
	`
	fillWeekdaysTxQuery := `insert into Weekdays (WeekdayName) values ("Sunday"), ("Monday"), ("Tuesday"), ("Wednesday"), ("Thursday"), ("Friday"), ("Saturday");`
	fillMonthsTxQuery := `insert into Months (MonthName) values ("January"), ("February"), ("March"), ("April"), ("May"), ("June"), ("July"), ("August"), ("September"), ("October"), ("November"), ("December");`
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	execInTx(tx, initialTxQuery)
	execInTx(tx, fillWeekdaysTxQuery)
	execInTx(tx, fillMonthsTxQuery)
	fillDatesTableStmt, err := tx.Prepare(`insert into Dates (Month, Day, Year, Weekday) values (?, ?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer fillDatesTableStmt.Close()
	initDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 365.25*40; i++ {
		workingDate := initDate.AddDate(0, 0, i)
		workingMonth := int(workingDate.Month())
		workingDay := workingDate.Day()
		workingYear := workingDate.Year()
		workingWeekday := fmt.Sprint(workingDate.Weekday())
		//log.Println(workingMonth, workingDay, workingYear, workingWeekday)
		_, err = fillDatesTableStmt.Exec(workingMonth, workingDay, workingYear, workingWeekday)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func initDatabase() (*sql.DB, error) { // https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
	dbExists := false
	if _, err := os.Stat(dbName); err == nil {
		dbExists = true
	}
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if !dbExists {
		createDatabase(db)
	}
	basicQuery := `select Month, Day, Year, MonthName, Weekday from dates join Months on Dates.Month = Months.MonthID where Dates.Year = 2024 and Dates.Day = 1;`
	rows, err := db.Query(basicQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var Month int
		var Day int
		var Year int
		var MonthName string
		var WeekdayName string
		err = rows.Scan(&Month, &Day, &Year, &MonthName, &WeekdayName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%d) %d, %d - %s\n", MonthName, Month, Day, Year, WeekdayName)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func main() {
	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Done. Press enter to exit executable.")
	_, _ = reader.ReadString('\n')
}
