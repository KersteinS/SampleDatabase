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

type Env struct { //define in main module
	sample SampleModel //would need to reference submodule with ".", i.e. models.SampleModel.
}

type SampleModel struct { //define in submodule for db model
	DB *sql.DB
}

type thisDate struct { //test, not for real use
	Month       int
	Day         int
	Year        int
	MonthName   string
	WeekdayName string
}

func (sm SampleModel) BasicSelectQuery() ([]thisDate, error) { // test, not for real use
	basicQuery := `select Month, Day, Year, MonthName, Weekday from dates join Months on Dates.Month = Months.MonthID where Dates.Year = 2024 and Dates.Day = 1;`
	var result []thisDate
	rows, err := sm.DB.Query(basicQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		aDate := thisDate{}
		err = rows.Scan(&aDate.Month, &aDate.Day, &aDate.Year, &aDate.MonthName, &aDate.WeekdayName)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, aDate)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return result, nil
}

type weekday struct {
	WeekdayID   int
	WeekdayName string
}

type month struct {
	MonthID   int
	MonthName string
}

type date struct {
	DateID  int
	Month   int
	Day     int
	Year    int
	Weekday string
}

type user struct {
	UserName string
	Password []byte
}

type volunteer struct {
	VolunteerID   int
	VolunteerName string
	User          string
}

type schedule struct {
	ScheduleID         int
	ScheduleName       string
	ShiftsOff          int
	VolunteersPerShift int
	User               string
	StartDate          int
	EndDate            int
}

type weekdayForSchedule struct {
	WFSID    int
	User     string
	Weekday  string
	Schedule int
}

type volunteerForSchedule struct {
	VFSID     int
	User      string
	Schedule  int
	Volunteer int
}

type unavailabilityForSchedule struct {
	UFSID                int
	User                 string
	VolunteerForSchedule int
	Date                 int
}

type completedSchedule struct {
	CScheduleID  string
	ScheduleData string
	User         string
	Schedule     string
}

type SendReceiveDataStruct struct {
	User                      string
	ScheduleName              string
	VolunteerAvailabilityData []map[string][]string // this is where most of the fun is. Need to create two functions: one to encode the db values into this format, one to decode this format into structs for insertion into the db
	StartDate                 string
	EndDate                   string
	WeekdaysForSchedule       []string
	ShiftsOff                 int
	VolunteersPerShift        int
}

func execInTx(tx *sql.Tx, query string) {
	_, err := tx.Exec(query)
	if err != nil {
		log.Printf("%q: %s\n", err, query)
		log.Fatal(err)
	}
}

func (sm SampleModel) CreateDatabase() {
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
		Password blob(64)
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
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	execInTx(tx, initialTxQuery)
	execInTx(tx, fillWeekdaysTxQuery)
	execInTx(tx, fillMonthsTxQuery)
	execInTx(tx, `insert into Users (UserName) values ("Seth")`)
	fillDatesTableStmt, err := tx.Prepare(`insert into Dates (Month, Day, Year, Weekday) values (?, ?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer fillDatesTableStmt.Close()
	initDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 365.25*40; i++ {
		workingDate := initDate.AddDate(0, 0, i) // these should be a struct
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

func (sm SampleModel) SendScheduleNames(user string) []string {
	var result []string
	return result
}

func (sm SampleModel) FetchAndSendData(user string, schedule string) SendReceiveDataStruct {
	var result SendReceiveDataStruct
	return result
}

func (sm SampleModel) RecieveAndStoreData(data SendReceiveDataStruct) { // should this return a completed/failed value?
	// fill this in
}

func (sm SampleModel) RequestDeleteSchedule(user string, schedule string) { // figure out what to return as a completed/failed value
	// fill this in
}

func (sm SampleModel) RequestDeleteCompletedSchedule(user string, schedule string, toDelete string) { // how to identify what to delete? Figure out what to return as a completed/failed value
	// fill this in
}

/*
Implementation needs CRUD functions:
Create
Read
Update
Delete

Weekdays, Months, and Dates are readonly.
What data will be requested by the app?
	List of schedule names for a user;
	A schedule joined with its UFS (joined to its VFS (joined to its Volunteers)) and its WFS for a user;
	Completed schedules for a user;
What data will be sent by the app?
	A schedule struct including all the data needed to create/update rows on Schedules, Volunteers, WFS, VFS, and UFS
	A completed schedule struct to create a new row on CompletedSchedules
What Delete options are needed?
	Delete Schedule should delete a single row on Schedules and multiple rows on WFS, VFS, UFS, and CompletedSchedules
	Delete Completed Schedule should delete a single row on CompletedSchedules

This does not contemplate CRUDing users yet.
*/

func main() {
	dbExists := false
	if _, err := os.Stat(dbName); err == nil {
		dbExists = true
	}
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	env := &Env{
		sample: SampleModel{DB: db},
	}
	defer env.sample.DB.Close()
	if !dbExists {
		env.sample.CreateDatabase()
	}
	fmt.Println(env.sample.BasicSelectQuery())
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Done. Press enter to exit executable.")
	_, _ = reader.ReadString('\n')
	fmt.Print(weekday{}, month{}, date{}, user{}, volunteer{}, schedule{}, weekdayForSchedule{}, volunteerForSchedule{}, unavailabilityForSchedule{}, completedSchedule{})
}
