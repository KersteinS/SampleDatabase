package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "./sample.db?_foreign_keys=on"

type Env struct { //define in main module
	sample       SampleModel //would need to reference submodule with ".", i.e. models.SampleModel.
	loggedInUser string
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
	CompletedSchedules        []string
}

func CsvSlice(stringSlice []string) string {
	jsonEncodedSlice, err := json.Marshal(stringSlice)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Trim(string(jsonEncodedSlice), "[]")
}

func countGTZero(intSlice []int) int {
	count := 0
	for _, val := range intSlice {
		if val > 0 {
			count++
		}
	}
	return count
}

func testEmpty[T comparable](sliceOfT []T, emptyT T) bool {
	for _, val := range sliceOfT {
		if val == emptyT {
			return true
		}
	}
	return false
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
	execInTx(tx, `insert into Users (UserName) values ("Seth")`) // for testing only
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

func (sm SampleModel) SendScheduleNames(currentUser string) []string {
	var result []string
	scheduleStructs := sm.RequestSchedules(currentUser, []schedule{})
	for i := 0; i < len(scheduleStructs); i++ {
		result = append(result, scheduleStructs[i].ScheduleName)
	}
	return result
}

func (sm SampleModel) FetchAndSendData(currentUser string, currentSchedule string) SendReceiveDataStruct {
	var result SendReceiveDataStruct
	scheduleQuery := fmt.Sprintf(`select StartDate, EndDate, ShiftsOff, VolunteersPerShift from Schedules where User = "%s" and ScheduleName = "%s"`, currentUser, currentSchedule)
	fmt.Println(scheduleQuery)
	result.User = currentUser
	result.ScheduleName = currentSchedule
	return result
}

func (sm SampleModel) RecieveAndStoreData(data SendReceiveDataStruct) { // should this return a completed/failed value?
	// fill this in
}

func (sm SampleModel) RequestDates(dates []date) []date {
	dateQuery := `select * from Dates`
	if len(dates) > 0 {
		if testEmpty(dates, date{}) {
			log.Fatal("RequestDates failed because one of the values in dates had an empty/default values date struct")
		}
		dateQuery = fmt.Sprintf(`%s where (`, dateQuery)
	} else {
		log.Fatal("RequestDates failed because the dates argument was an empty slice. At least one date must be requested. ")
	}
	for i := 0; i < len(dates); i++ {
		count := countGTZero([]int{dates[i].DateID, dates[i].Month, dates[i].Day, dates[i].Year, len(dates[i].Weekday)})
		// count must be at least 1 because the testEmpty check passed
		//fmt.Println(count)
		dateQuery = fmt.Sprintf(`%s(`, dateQuery)
		if dates[i].DateID > 0 {
			dateQuery = fmt.Sprintf(`%sDateID = %d`, dateQuery, dates[i].DateID)
			count--
			if count > 0 {
				dateQuery = fmt.Sprintf(`%s and `, dateQuery)
			}
		}
		if dates[i].Month > 0 {
			dateQuery = fmt.Sprintf(`%sMonth = %d`, dateQuery, dates[i].Month)
			count--
			if count > 0 {
				dateQuery = fmt.Sprintf(`%s and `, dateQuery)
			}
		}
		if dates[i].Day > 0 {
			dateQuery = fmt.Sprintf(`%sDay = %d`, dateQuery, dates[i].Day)
			count--
			if count > 0 {
				dateQuery = fmt.Sprintf(`%s and `, dateQuery)
			}
		}
		if dates[i].Year > 0 {
			dateQuery = fmt.Sprintf(`%sYear = %d`, dateQuery, dates[i].Year)
			count--
			if count > 0 {
				dateQuery = fmt.Sprintf(`%s and `, dateQuery)
			}
		}
		if len(dates[i].Weekday) > 0 {
			dateQuery = fmt.Sprintf(`%sWeekday = "%s"`, dateQuery, dates[i].Weekday)
		}
		dateQuery = fmt.Sprintf(`%s)`, dateQuery)
		if i+1 < len(dates) {
			dateQuery = fmt.Sprintf(`%s or `, dateQuery)
		}
		//fmt.Println(count)
		//fmt.Println(dateQuery)
	}
	if len(dates) > 0 {
		dateQuery = fmt.Sprintf(`%s)`, dateQuery)
	}
	//fmt.Println(dateQuery)
	var result []date
	rows, err := sm.DB.Query(dateQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var dateStruct date
		err = rows.Scan(&dateStruct.DateID, &dateStruct.Month, &dateStruct.Day, &dateStruct.Year, &dateStruct.Weekday)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, dateStruct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func (sm SampleModel) CreateVolunteers(currentUser string, toCreate []volunteer) {
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fillVolunteersTableString := `insert into Volunteers (VolunteerName, User) values (?, ?)`
	fillVolunteersTableStmt, err := tx.Prepare(fillVolunteersTableString)
	if err != nil {
		log.Fatalf("Error in Create Volunteers statement: %v\n%s", err, fillVolunteersTableString)
	}
	defer fillVolunteersTableStmt.Close()
	for i := 0; i < len(toCreate); i++ {
		_, err = fillVolunteersTableStmt.Exec(toCreate[i].VolunteerName, toCreate[i].User)
		if err != nil {
			log.Fatalf("Error in Create Volunteers loop: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (sm SampleModel) RequestVolunteers(currentUser string, volunteers []volunteer) []volunteer {
	volunteersQuery := fmt.Sprintf(`select * from Volunteers where User = "%s"`, currentUser)
	if len(volunteers) > 0 {
		if testEmpty(volunteers, volunteer{}) {
			log.Fatal("RequestVolunteers failed because one of the values in volunteers had an empty/default values volunteer struct")
		}
		volunteersQuery = fmt.Sprintf(`%s and (`, volunteersQuery)
	}
	for i := 0; i < len(volunteers); i++ {
		count := countGTZero([]int{volunteers[i].VolunteerID, len(volunteers[i].VolunteerName), len(volunteers[i].User)})
		// count must be at least 1 because the testEmpty check passed
		//fmt.Println(count)
		volunteersQuery = fmt.Sprintf(`%s(`, volunteersQuery)
		if volunteers[i].VolunteerID > 0 {
			volunteersQuery = fmt.Sprintf(`%sVolunteerID = %d`, volunteersQuery, volunteers[i].VolunteerID)
			count--
			if count > 0 {
				volunteersQuery = fmt.Sprintf(`%s and `, volunteersQuery)
			}
		}
		if len(volunteers[i].VolunteerName) > 0 {
			volunteersQuery = fmt.Sprintf(`%sVolunteerName = "%s"`, volunteersQuery, volunteers[i].VolunteerName)
			count--
			if count > 0 {
				volunteersQuery = fmt.Sprintf(`%s and `, volunteersQuery)
			}
		}
		if len(volunteers[i].User) > 0 {
			volunteersQuery = fmt.Sprintf(`%sUser = "%s"`, volunteersQuery, volunteers[i].User)
		}
		volunteersQuery = fmt.Sprintf(`%s)`, volunteersQuery)
		if i+1 < len(volunteers) {
			volunteersQuery = fmt.Sprintf(`%s or `, volunteersQuery)
		}
		//fmt.Println(count)
		//fmt.Println(volunteersQuery)
	}
	if len(volunteers) > 0 {
		volunteersQuery = fmt.Sprintf(`%s)`, volunteersQuery)
	}
	//fmt.Println(volunteersQuery)
	var result []volunteer
	rows, err := sm.DB.Query(volunteersQuery)
	if err != nil {
		log.Fatalf("Error in Request Volunteers query: %v\n%s", err, volunteersQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var volunteerStruct volunteer
		err = rows.Scan(&volunteerStruct.VolunteerID, &volunteerStruct.VolunteerName, &volunteerStruct.User)
		if err != nil {
			log.Fatalf("Error in Request Volunteers loop: %v", err)
		}
		result = append(result, volunteerStruct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("Error in Request Volunteers rows.Err(): %v", err)
	}
	return result
}

func (sm SampleModel) UpdateVolunteers(currentUser string, existingVolunteers []volunteer) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteVolunteers(currentUser string, existingVolunteers []volunteer) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) CleanOrphanedVolunteers(currentUser string) { // Need to delete all volunteers who are not referenced by a VFS/UFS. Figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in. If it's not in VFS or UFS but has the same UserName, delete it. I shouldn't need to put any sql in here. I should use the CRUD functions.
}

func (sm SampleModel) CreateSchedules(user string, toCreate []schedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fillSchedulesTableString := `insert into Schedules (ScheduleName, ShiftsOff, VolunteersPerShift, User, StartDate, EndDate) values (?, ?, ?, ?, ?, ?)`
	fillSchedulesTableStmt, err := tx.Prepare(fillSchedulesTableString)
	if err != nil {
		log.Fatalf("Error in Create Schedules statement: %v\n%s", err, fillSchedulesTableString)
	}
	defer fillSchedulesTableStmt.Close()
	for i := 0; i < len(toCreate); i++ {
		_, err = fillSchedulesTableStmt.Exec(toCreate[i].ScheduleName, toCreate[i].ShiftsOff, toCreate[i].VolunteersPerShift, toCreate[i].User, toCreate[i].StartDate, toCreate[i].EndDate)
		if err != nil {
			log.Fatalf("Error in Create Schedules loop: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

// This function is the simple version of RequestSchedulesExtended and does not allow ShiftsOff = 0 to be queried.
func (sm SampleModel) RequestSchedules(currentUser string, schedules []schedule) []schedule {
	return sm.RequestSchedulesExtended(currentUser, schedules, false)
}

// This version of RequestSchedules allows ShiftsOff = 0 to be queried, but any default schedule structs will have ShiftsOff: 0 implicitly, so ShiftsOff must be set to a desired value or to -1 to be ignored.
func (sm SampleModel) RequestSchedulesExtended(currentUser string, schedules []schedule, includeShiftsOff0 bool) []schedule {
	schedulesQuery := fmt.Sprintf(`select * from Schedules where User = "%s"`, currentUser)
	if len(schedules) > 0 {
		var checkAgainst schedule
		if includeShiftsOff0 {
			checkAgainst = schedule{ShiftsOff: -1}
		} else {
			checkAgainst = schedule{}
		}
		if testEmpty(schedules, checkAgainst) {
			log.Fatal("RequestSchedules failed because one of the values in schedules had an empty/default values schedule struct")
		}
		schedulesQuery = fmt.Sprintf(`%s and (`, schedulesQuery)
	}
	for i := 0; i < len(schedules); i++ {
		var count int
		if includeShiftsOff0 {
			count = countGTZero([]int{schedules[i].ScheduleID, len(schedules[i].ScheduleName), schedules[i].ShiftsOff + 1, schedules[i].VolunteersPerShift, len(schedules[i].User), schedules[i].StartDate, schedules[i].EndDate})
		} else {
			count = countGTZero([]int{schedules[i].ScheduleID, len(schedules[i].ScheduleName), schedules[i].ShiftsOff, schedules[i].VolunteersPerShift, len(schedules[i].User), schedules[i].StartDate, schedules[i].EndDate})
		}
		// count must be at least 1 because the testEmpty check passed
		//fmt.Println(count)
		schedulesQuery = fmt.Sprintf(`%s(`, schedulesQuery)
		if schedules[i].ScheduleID > 0 {
			schedulesQuery = fmt.Sprintf(`%sScheduleID = %d`, schedulesQuery, schedules[i].ScheduleID)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if len(schedules[i].ScheduleName) > 0 {
			schedulesQuery = fmt.Sprintf(`%sScheduleName = "%s"`, schedulesQuery, schedules[i].ScheduleName)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if schedules[i].ShiftsOff > 0 || includeShiftsOff0 && schedules[i].ShiftsOff > -1 {
			schedulesQuery = fmt.Sprintf(`%sShiftsOff = %d`, schedulesQuery, schedules[i].ShiftsOff)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if schedules[i].VolunteersPerShift > 0 {
			schedulesQuery = fmt.Sprintf(`%sVolunteersPerShift = %d`, schedulesQuery, schedules[i].VolunteersPerShift)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if len(schedules[i].User) > 0 {
			schedulesQuery = fmt.Sprintf(`%sUser = "%s"`, schedulesQuery, schedules[i].User)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if schedules[i].StartDate > 0 {
			schedulesQuery = fmt.Sprintf(`%sStartDate = %d`, schedulesQuery, schedules[i].StartDate)
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		if schedules[i].EndDate > 0 {
			count--
			if count > 0 {
				schedulesQuery = fmt.Sprintf(`%s and `, schedulesQuery)
			}
		}
		schedulesQuery = fmt.Sprintf(`%s)`, schedulesQuery)
		if i+1 < len(schedules) {
			schedulesQuery = fmt.Sprintf(`%s or `, schedulesQuery)
		}
		//fmt.Println(count)
		//fmt.Println(schedulesQuery)
	}
	if len(schedules) > 0 {
		schedulesQuery = fmt.Sprintf(`%s)`, schedulesQuery)
	}
	//fmt.Println(schedulesQuery)
	var result []schedule
	rows, err := sm.DB.Query(schedulesQuery)
	if err != nil {
		log.Fatalf("Error in Request Schedules query: %v\n%s", err, schedulesQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var userSchedule schedule
		err = rows.Scan(&userSchedule.ScheduleID, &userSchedule.ScheduleName, &userSchedule.ShiftsOff, &userSchedule.VolunteersPerShift, &userSchedule.User, &userSchedule.StartDate, &userSchedule.EndDate)
		if err != nil {
			log.Fatalf("Error in Request Schedules loop: %v", err)
		}
		result = append(result, userSchedule)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("Error in Request Schedules rows.Err(): %v", err)
	}
	return result
}

func (sm SampleModel) UpdateSchedule(currentUser string, existingSchedules []schedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteSchedule(currentUser string, existingSchedules []schedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) CreateWFS(currentUser string, toCreate []weekdayForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fillWFSTableString := `insert into WeekdaysForSchedule (User, Weekday, Schedule) values (?, ?, ?)`
	fillWFSTableStmt, err := tx.Prepare(fillWFSTableString)
	if err != nil {
		log.Fatalf("Error in Create WFS statement: %v\n%s", err, fillWFSTableString)
	}
	defer fillWFSTableStmt.Close()
	for i := 0; i < len(toCreate); i++ {
		_, err = fillWFSTableStmt.Exec(toCreate[i].User, toCreate[i].Weekday, toCreate[i].Schedule)
		if err != nil {
			log.Fatalf("Error in Create WFS loop: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (sm SampleModel) RequestWFS(currentUser string, weekdaysForSchedule []weekdayForSchedule) []weekdayForSchedule {
	weekdaysForScheduleQuery := fmt.Sprintf(`select * from WeekdaysForSchedule where User = "%s"`, currentUser)
	if len(weekdaysForSchedule) > 0 {
		if testEmpty(weekdaysForSchedule, weekdayForSchedule{}) {
			log.Fatal("RequestWFS failed because one of the values in weekdaysForSchedule had an empty/default values weekdayForSchedule struct")
		}
		weekdaysForScheduleQuery = fmt.Sprintf(`%s and (`, weekdaysForScheduleQuery)
	}
	for i := 0; i < len(weekdaysForSchedule); i++ {
		count := countGTZero([]int{weekdaysForSchedule[i].WFSID, len(weekdaysForSchedule[i].User), weekdaysForSchedule[i].Schedule, len(weekdaysForSchedule[i].Weekday)})
		// count must be at least 1 because the testEmpty check passed
		fmt.Println(count)
		weekdaysForScheduleQuery = fmt.Sprintf(`%s(`, weekdaysForScheduleQuery)
		if weekdaysForSchedule[i].WFSID > 0 {
			weekdaysForScheduleQuery = fmt.Sprintf(`%sWFSID = %d`, weekdaysForScheduleQuery, weekdaysForSchedule[i].WFSID)
			count--
			if count > 0 {
				weekdaysForScheduleQuery = fmt.Sprintf(`%s and `, weekdaysForScheduleQuery)
			}
		}
		if len(weekdaysForSchedule[i].User) > 0 {
			weekdaysForScheduleQuery = fmt.Sprintf(`%sUser = "%s"`, weekdaysForScheduleQuery, weekdaysForSchedule[i].User)
			count--
			if count > 0 {
				weekdaysForScheduleQuery = fmt.Sprintf(`%s and `, weekdaysForScheduleQuery)
			}
		}
		if weekdaysForSchedule[i].Schedule > 0 {
			weekdaysForScheduleQuery = fmt.Sprintf(`%sSchedule = %d`, weekdaysForScheduleQuery, weekdaysForSchedule[i].Schedule)
			count--
			if count > 0 {
				weekdaysForScheduleQuery = fmt.Sprintf(`%s and `, weekdaysForScheduleQuery)
			}
		}
		if len(weekdaysForSchedule[i].Weekday) > 0 {
			weekdaysForScheduleQuery = fmt.Sprintf(`%sWeekday = "%s"`, weekdaysForScheduleQuery, weekdaysForSchedule[i].Weekday)
		}
		weekdaysForScheduleQuery = fmt.Sprintf(`%s)`, weekdaysForScheduleQuery)
		if i+1 < len(weekdaysForSchedule) {
			weekdaysForScheduleQuery = fmt.Sprintf(`%s or `, weekdaysForScheduleQuery)
		}
		fmt.Println(count)
		fmt.Println(weekdaysForScheduleQuery)
	}
	if len(weekdaysForSchedule) > 0 {
		weekdaysForScheduleQuery = fmt.Sprintf(`%s)`, weekdaysForScheduleQuery)
	}
	fmt.Println(weekdaysForScheduleQuery)
	var result []weekdayForSchedule
	rows, err := sm.DB.Query(weekdaysForScheduleQuery)
	if err != nil {
		log.Fatalf("Error in Request WFS query: %v\n%s", err, weekdaysForScheduleQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var weekdayForScheduleStruct weekdayForSchedule
		err = rows.Scan(&weekdayForScheduleStruct.WFSID, &weekdayForScheduleStruct.User, &weekdayForScheduleStruct.Weekday, &weekdayForScheduleStruct.Schedule)
		if err != nil {
			log.Fatalf("Error in Request WFS loop: %v", err)
		}
		result = append(result, weekdayForScheduleStruct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("Error in Request WFS rows.Err(): %v", err)
	}
	return result
}

func (sm SampleModel) UpdateWFS(currentUser string, existingWeekdaysForSchedule []weekdayForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteWFS(currentUser string, existingWeekdaysForSchedule []weekdayForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) CleanOrphanedWFS(currentUser string, currentSchedule string, currentWeekdays []string) { // Figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in. If it's not in currentWeekdays but has the same UserName and ScheduleName, delete it. I shouldn't need to put any sql in here. I should use the CRUD functions.
}

func (sm SampleModel) CreateVFS(currentUser string, toCreate []volunteerForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fillVFSTableString := `insert into VolunteersForSchedule (User, Schedule, Volunteer) values (?, ?, ?)`
	fillVFSTableStmt, err := tx.Prepare(fillVFSTableString)
	if err != nil {
		log.Fatalf("Error in Create VFS statement: %v\n%s", err, fillVFSTableString)
	}
	defer fillVFSTableStmt.Close()
	for i := 0; i < len(toCreate); i++ {
		_, err = fillVFSTableStmt.Exec(toCreate[i].User, toCreate[i].Schedule, toCreate[i].Volunteer)
		if err != nil {
			log.Fatalf("Error in Create VFS loop: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (sm SampleModel) RequestVFS(currentUser string, volunteersForSchedule []volunteerForSchedule) []volunteerForSchedule {
	VFSQuery := fmt.Sprintf(`select VFSID, User, Schedule, Volunteer from VolunteersForSchedule where User = "%s"`, currentUser)
	if len(volunteersForSchedule) > 0 {
		if testEmpty(volunteersForSchedule, volunteerForSchedule{}) {
			log.Fatal("RequestVFS failed because one of the values in volunteersForSchedule had an empty/default values volunteerForSchedule struct")
		}
		VFSQuery = fmt.Sprintf(`%s and (`, VFSQuery)
	}
	for i := 0; i < len(volunteersForSchedule); i++ {
		count := countGTZero([]int{volunteersForSchedule[i].VFSID, len(volunteersForSchedule[i].User), volunteersForSchedule[i].Schedule, volunteersForSchedule[i].Volunteer})
		VFSQuery = fmt.Sprintf(`%s(`, VFSQuery)
		if volunteersForSchedule[i].VFSID > 0 {
			VFSQuery = fmt.Sprintf(`%sVFSID = %d`, VFSQuery, volunteersForSchedule[i].VFSID)
			count--
			if count > 0 {
				VFSQuery = fmt.Sprintf(`%s and `, VFSQuery)
			}
		}
		if len(volunteersForSchedule[i].User) > 0 {
			VFSQuery = fmt.Sprintf(`%sUser = "%s"`, VFSQuery, volunteersForSchedule[i].User)
			count--
			if count > 0 {
				VFSQuery = fmt.Sprintf(`%s and `, VFSQuery)
			}
		}
		if volunteersForSchedule[i].Schedule > 0 {
			VFSQuery = fmt.Sprintf(`%sSchedule = %d`, VFSQuery, volunteersForSchedule[i].Schedule)
			count--
			if count > 0 {
				VFSQuery = fmt.Sprintf(`%s and `, VFSQuery)
			}
		}
		if volunteersForSchedule[i].Volunteer > 0 {
			VFSQuery = fmt.Sprintf(`%sVolunteer = %d`, VFSQuery, volunteersForSchedule[i].Volunteer)
		}
		VFSQuery = fmt.Sprintf(`%s)`, VFSQuery)
		if i+1 < len(volunteersForSchedule) {
			VFSQuery = fmt.Sprintf(`%s or `, VFSQuery)
		}
		//fmt.Println(count)
		//fmt.Println(VFSQuery)
	}
	if len(volunteersForSchedule) > 0 {
		VFSQuery = fmt.Sprintf(`%s)`, VFSQuery)
	}
	//fmt.Println(VFSQuery)
	var result []volunteerForSchedule
	rows, err := sm.DB.Query(VFSQuery)
	if err != nil {
		log.Fatalf("Error in Request VFS query: %v\n%s", err, VFSQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var VFSStruct volunteerForSchedule
		err = rows.Scan(&VFSStruct.VFSID, &VFSStruct.User, &VFSStruct.Schedule, &VFSStruct.Volunteer)
		if err != nil {
			log.Fatalf("Error in Request VFS loop: %v", err)
		}
		result = append(result, VFSStruct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("Error in Request VFS rows.Err(): %v", err)
	}
	return result
}

func (sm SampleModel) UpdateVFS(currentUser string, existingVolunteersForSchedule []volunteerForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteVFS(currentUser string, existingVolunteersForSchedule []volunteerForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) CleanOrphanedVFS(currentUser string, currentSchedule string, currentVolunteers []string) { // Figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in. If it's not in currentVolunteers but has the same UserName and ScheduleName, delete it. I shouldn't need to put any sql in here. I should use the CRUD functions.
}

func (sm SampleModel) CreateUFS(currentUser string, toCreate []unavailabilityForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	tx, err := sm.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fillUFSTableString := `insert into UnavailabilitiesForSchedule (User, VolunteerForSchedule, Date) values (?, ?, ?)`
	fillUFSTableStmt, err := tx.Prepare(fillUFSTableString)
	if err != nil {
		log.Fatalf("Error in Create UFS statement: %v\n%s", err, fillUFSTableString)
	}
	defer fillUFSTableStmt.Close()
	for i := 0; i < len(toCreate); i++ {
		_, err = fillUFSTableStmt.Exec(toCreate[i].User, toCreate[i].VolunteerForSchedule, toCreate[i].Date)
		if err != nil {
			log.Fatalf("Error in Create UFS loop: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (sm SampleModel) RequestUFS(currentUser string, unavailabilitiesForSchedule []unavailabilityForSchedule) []unavailabilityForSchedule {
	UFSQuery := fmt.Sprintf(`select * from UnavailabilitiesForSchedule where User = "%s"`, currentUser)
	if len(unavailabilitiesForSchedule) > 0 {
		if testEmpty(unavailabilitiesForSchedule, unavailabilityForSchedule{}) {
			log.Fatal("RequestUFS failed because one of the values in unavailabilitiesForSchedule had an empty/default values unavailabilityForSchedule struct")
		}
		UFSQuery = fmt.Sprintf(`%s and (`, UFSQuery)
	}
	for i := 0; i < len(unavailabilitiesForSchedule); i++ {
		count := countGTZero([]int{unavailabilitiesForSchedule[i].UFSID, len(unavailabilitiesForSchedule[i].User), unavailabilitiesForSchedule[i].VolunteerForSchedule, unavailabilitiesForSchedule[i].Date})
		// count must be at least 1 because the testEmpty check passed
		//fmt.Println(count)
		UFSQuery = fmt.Sprintf(`%s(`, UFSQuery)
		if unavailabilitiesForSchedule[i].UFSID > 0 {
			UFSQuery = fmt.Sprintf(`%sUFSID = %d`, UFSQuery, unavailabilitiesForSchedule[i].UFSID)
			count--
			if count > 0 {
				UFSQuery = fmt.Sprintf(`%s and `, UFSQuery)
			}
		}
		if len(unavailabilitiesForSchedule[i].User) > 0 {
			UFSQuery = fmt.Sprintf(`%sUser = "%s"`, UFSQuery, unavailabilitiesForSchedule[i].User)
			count--
			if count > 0 {
				UFSQuery = fmt.Sprintf(`%s and `, UFSQuery)
			}
		}
		if unavailabilitiesForSchedule[i].VolunteerForSchedule > 0 {
			UFSQuery = fmt.Sprintf(`%sVolunteerForSchedule = %d`, UFSQuery, unavailabilitiesForSchedule[i].VolunteerForSchedule)
			count--
			if count > 0 {
				UFSQuery = fmt.Sprintf(`%s and `, UFSQuery)
			}
		}
		if unavailabilitiesForSchedule[i].Date > 0 {
			UFSQuery = fmt.Sprintf(`%sDate = %d`, UFSQuery, unavailabilitiesForSchedule[i].Date)
		}
		UFSQuery = fmt.Sprintf(`%s)`, UFSQuery)
		if i+1 < len(unavailabilitiesForSchedule) {
			UFSQuery = fmt.Sprintf(`%s or `, UFSQuery)
		}
		//fmt.Println(count)
		//fmt.Println(UFSQuery)
	}
	if len(unavailabilitiesForSchedule) > 0 {
		UFSQuery = fmt.Sprintf(`%s)`, UFSQuery)
	}
	//fmt.Println(UFSQuery)
	var result []unavailabilityForSchedule
	rows, err := sm.DB.Query(UFSQuery)
	if err != nil {
		log.Fatalf("Error in Request UFS query: %v\n%s", err, UFSQuery)
	}
	defer rows.Close()
	for rows.Next() {
		var UFSStruct unavailabilityForSchedule
		err = rows.Scan(&UFSStruct.UFSID, &UFSStruct.User, &UFSStruct.VolunteerForSchedule, &UFSStruct.Date)
		if err != nil {
			log.Fatalf("Error in Request UFS loop: %v", err)
		}
		result = append(result, UFSStruct)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("Error in Request UFS rows.Err(): %v", err)
	}
	return result
}

func (sm SampleModel) UpdateUFS(currentUser string, existingUnavailabilitiesForSchedule []unavailabilityForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteUFS(currentUser string, existingUnavailabilitiesForSchedule []unavailabilityForSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) CleanOrphanedUFS(currentUser string, currentSchedule string, currentVolunteer string, currentUnavailabilities []string) { // Need to delete all related UFS when deleting a VFS. If there is a volunteer with no VFS's, delete the volunteer (I think). Figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in. Must be done per volunteer. If it's not in currentUnavailabilities but has the same UserName, ScheduleName, and VolunteerName, delete it. I shouldn't need to put any sql in here. I should use the CRUD functions.
}

func (sm SampleModel) CreateCompletedSchedule(currentUser string, toCreate completedSchedule) { // figure out what to return as a completed/failed value, instead of just crashing the program
	// fill this in
}

func (sm SampleModel) DeleteCompletedSchedule(currentUser string, existingSchedule string, toDelete string) { // how to identify what to delete? Figure out what to return as a completed/failed value, instead of just crashing the program
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
		sample:       SampleModel{DB: db},
		loggedInUser: "Seth",
	}
	defer env.sample.DB.Close()
	if !dbExists {
		env.sample.CreateDatabase()
	}
	schedules := []schedule{
		{
			ScheduleName:       "test1",
			ShiftsOff:          3,
			VolunteersPerShift: 3,
			StartDate:          env.sample.RequestDates([]date{{Month: 1, Day: 1, Year: 2024}})[0].DateID,
			EndDate:            env.sample.RequestDates([]date{{Month: 3, Day: 1, Year: 2024}})[0].DateID,
			User:               env.loggedInUser,
		},
		{
			ScheduleName:       "test2",
			ShiftsOff:          3,
			VolunteersPerShift: 3,
			StartDate:          env.sample.RequestDates([]date{{Month: 3, Day: 1, Year: 2024}})[0].DateID,
			EndDate:            env.sample.RequestDates([]date{{Month: 6, Day: 1, Year: 2024}})[0].DateID,
			User:               env.loggedInUser,
		},
		{
			ScheduleName:       "test3",
			ShiftsOff:          3,
			VolunteersPerShift: 3,
			StartDate:          env.sample.RequestDates([]date{{Month: 6, Day: 1, Year: 2024}})[0].DateID,
			EndDate:            env.sample.RequestDates([]date{{Month: 9, Day: 1, Year: 2024}})[0].DateID,
			User:               env.loggedInUser,
		},
	}
	env.sample.CreateSchedules(env.loggedInUser, schedules)
	weekdaysForSchedule := []weekdayForSchedule{
		{
			User:    env.loggedInUser,
			Weekday: "Sunday",
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
		},
		{
			User:    env.loggedInUser,
			Weekday: "Wednesday",
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
		},
		{
			User:    env.loggedInUser,
			Weekday: "Friday",
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test3"},
			})[0].ScheduleID,
		},
	}
	env.sample.CreateWFS(env.loggedInUser, weekdaysForSchedule)
	volunteers := []volunteer{
		{
			VolunteerName: "Tim",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "Bill",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "Jack",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "George",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "Bob",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "Lance",
			User:          env.loggedInUser,
		},
		{
			VolunteerName: "Larry",
			User:          env.loggedInUser,
		},
	}
	env.sample.CreateVolunteers(env.loggedInUser, volunteers)
	volunteersForSchedule := []volunteerForSchedule{
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Tim"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bill"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Jack"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "George"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bob"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Lance"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Larry"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Tim"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test3"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bill"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test3"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Jack"}})[0].VolunteerID,
		},
		{
			User: env.loggedInUser,
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test3"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "George"}})[0].VolunteerID,
		},
	}
	env.sample.CreateVFS(env.loggedInUser, volunteersForSchedule)
	unavailabilitiesForSchedule := []unavailabilityForSchedule{
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test1"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Tim"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 1, Day: 14, Year: 2024}})[0].DateID,
		},
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test1"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bill"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 1, Day: 21, Year: 2024}})[0].DateID,
		},
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test2"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bob"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 5, Day: 12, Year: 2024}})[0].DateID,
		},
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test2"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Lance"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 5, Day: 19, Year: 2024}})[0].DateID,
		},
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test3"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Jack"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 8, Day: 11, Year: 2024}})[0].DateID,
		},
		{
			User: env.loggedInUser,
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test3"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "George"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 8, Day: 18, Year: 2024}})[0].DateID,
		},
	}
	env.sample.CreateUFS(env.loggedInUser, unavailabilitiesForSchedule)
	fmt.Println(env.sample.RequestWFS(env.loggedInUser, []weekdayForSchedule{
		{
			Weekday: env.sample.RequestDates([]date{
				{
					Month: 8,
					Day:   18,
					Year:  2024,
				},
			})[0].Weekday,
		},
		{
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
		},
		{
			Weekday: env.sample.RequestDates([]date{
				{Weekday: "Friday"},
			})[0].Weekday,
		},
	}))
	fmt.Println(env.sample.RequestSchedules(env.loggedInUser, []schedule{
		{ScheduleName: "test1",
			ShiftsOff: 3},
		{StartDate: env.sample.RequestDates([]date{{Month: 1, Day: 1, Year: 2024}})[0].DateID,
			ShiftsOff: 3,
		},
	}))
	fmt.Println(env.sample.RequestSchedulesExtended(env.loggedInUser, []schedule{{ShiftsOff: 0}}, true))
	fmt.Println(env.sample.RequestSchedules(env.loggedInUser, []schedule{}))
	fmt.Println(env.sample.RequestVolunteers(env.loggedInUser, []volunteer{
		{VolunteerID: 1, VolunteerName: "Tim"},
		{VolunteerName: "George"},
	}))
	fmt.Println(env.sample.RequestVolunteers(env.loggedInUser, []volunteer{}))
	fmt.Println(env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
		{
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test1"},
			})[0].ScheduleID,
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Tim"}})[0].VolunteerID,
		},
		{
			Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
				{ScheduleName: "test2"},
			})[0].ScheduleID,
		},
		{
			Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "George"}})[0].VolunteerID,
		},
	}))
	fmt.Println(env.sample.RequestUFS(env.loggedInUser, []unavailabilityForSchedule{
		{
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test1"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Tim"}})[0].VolunteerID,
				},
			})[0].VFSID,
			Date: env.sample.RequestDates([]date{{Month: 1, Day: 14, Year: 2024}})[0].DateID,
		},
		{
			UFSID: 2,
		},
		{
			VolunteerForSchedule: env.sample.RequestVFS(env.loggedInUser, []volunteerForSchedule{
				{
					Schedule: env.sample.RequestSchedules(env.loggedInUser, []schedule{
						{ScheduleName: "test2"},
					})[0].ScheduleID,
					Volunteer: env.sample.RequestVolunteers(env.loggedInUser, []volunteer{{VolunteerName: "Bob"}})[0].VolunteerID,
				},
			})[0].VFSID,
		},
		{
			Date: env.sample.RequestDates([]date{{Month: 8, Day: 18, Year: 2024}})[0].DateID,
		},
	}))
	fmt.Println(env.sample.RequestUFS(env.loggedInUser, []unavailabilityForSchedule{}))
	fmt.Println("Done. Press enter to exit executable.")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Print(weekday{}, month{}, user{})
}
