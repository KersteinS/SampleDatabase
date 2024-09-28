The purpose of this project is to create a Go program that generates and pre-fills (for internally generated data) a sqlite3 database following the below schema and provides CRUD operations for each table in the database.

`TableName` (`PrimaryKey-ColumnName[data-type]`,`ColumnName[data-type]`, `ForeignKey-ColumnName-SourceTable(SourceColumnName)`) `data source`:

`Weekdays` (PK-`DayID`[`integer`], `DayName`[`unique-text`]) internally generated\
`Months` (PK-`MonthID`[`integer`], `MonthName`[`unique-text`]) internally generated\
`Dates` (PK-`DateID`[`integer`], `Month`[`integer`], `Day`[`integer`], `Year`[`integer`], `Weekday`[`text`], FK-`Month`-`Months(MonthID)`, FK-`Weekday`-`Weekdays(WeekdayID)`) internally generated\
`Users` (PK-`UserName`[`unique-text`], `Password`[`blob(64)`]) input\
`Volunteers` (PK-`VolunteerID`[`integer`], `VolunteerName`[`text`], `User`[`text`], FK-`User`-`Users(UserName)`) input\
`Schedules` (PK-`ScheduleID`[`integer`], `ScheduleName`-[`text`], `ShiftsOff`[`integer`], `VolunteersPerShift`[`integer`], `User`[`text`], `StartDate`[`integer`], `EndDate`[`integer`], FK-`User`-`Users(UserName)`, FK-`StartDate`-`Dates(DateID)`, FK-`EndDate`-`Dates(DateID)`) input\
`WeekdaysForSchedule` (PK-`WFSID`[`integer`], `User`[`text`], `Weekday`[`integer`], `Schedule`[`integer`], FK-`User`-`Users(UserName)`, FK-`Weekday`-`Weekdays(WeekdayID)`, FK-`Schedule`-`Schedules(ScheduleID)`) input\
`VolunteersForSchedule` (PK-`VFSID`[`integer`], `User`[`text`], `Schedule`[`integer`], `Volunteer`[`integer`], FK-`User`-`Users(UserName)`, FK-`Schedule`-`Schedules(ScheduleID)`, FK-`Volunteer`-`Volunteers(VolunteerID)`) input\
`UnavailabilitiesForSchedule` (PK-`UFSID`[`integer`], `User`[`text`], `VolunteerForSchedule`[`integer`], `Date`[`integer`], FK-`User`-`Users(UserName)`, FK-`VolunteerForSchedule`-`VolunteersForSchedule(VFSID)`, FK-`Date`-`Dates(DateID)`) input\
`CompletedSchedules` (PK-`CScheduleID`[`integer`], `ScheduleData`[`json-text`], `User`[`text`], `Schedule`[`integer`], FK-`User`-`Users(UserName)`, FK-`Schedule`-`Schedules(ScheduleID)`) output\
