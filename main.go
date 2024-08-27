package main

import (
	"bufio"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "./sample.db?_foreign_keys=on"

func createDatabase(db *sql.DB) { //add not null to most things in sqlStmt
	sqlStmt := `
	create table Weekdays (
		DayID integer primary key autoincrement,
		DayName text
	);
	create table Months (
		MonthID integer primary key autoincrement,
		MonthName text not null
	);
	create table Dates (
		DateID integer primary key autoincrement,
		Month integer not null,
		Day integer not null,
		Year integer not null,
		foreign key (Month) references Months(MonthID),
		foreign key (Day) references Weekdays(DayID)
	);
	create table Users (
		UserID text primary key,
		Password blob(64) not null
	) without rowid;
	create table Volunteers (
		VolunteerID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		VolunteerName text not null
	);
	`
	/*create table Schedules (
		ScheduleID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		ScheduleName text not null,
		foreign key (StartDate) references Dates(DateID),
		foreign key (EndDate) references Dates(DateID),
		ShiftsOff integer not null,
		VolunteersPerShift integer not null
	);
	create table WeekdaysForSchedule (
		WFSID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		foreign key (Weekday) references Weekdays(DayID),
		foreign key (Schedule) references Schedules(ScheduleID)
	);
	create table VolunteersForSchedule (
		VFSID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		foreign key (Schedule) references Schedules(ScheduleID),
		foreign key (Volunteer) refernces Volunteers(VolunteerID)
	);
	create table UnavailabilitiesForSchedule (
		UFSID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		foreign key (VolunteerForSchedule) references VolunteersForSchedule(VFSID),
		foreign key (Date) refernces Dates(DateID)
	);
	create table CompletedSchedules (
		CScheduleID integer primary key autoincrement,
		foreign key (UserID) references Users(UserID),
		foreign key (Schedule) references Schedules(ScheduleID),
		ScheduleData text not null
	);
	`*/
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		log.Fatal(err)
	}
	/*stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("こんにちは世界%03d", i))
		if err != nil {
			log.Fatal(err)
		}
	}*/
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
	/*
		rows, err := db.Query("select id, name from foo")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(id, name)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err := db.Prepare("select name from foo where id = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		var name string
		err = stmt.QueryRow("3").Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(name)

		_, err = db.Exec("delete from foo")
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
		if err != nil {
			log.Fatal(err)
		}

		rows, err = db.Query("select id, name from foo")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(id, name)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}*/
	return db, nil
}

func main() {
	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}
