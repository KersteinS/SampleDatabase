The purpose of this project is to create a Go program that generates and pre-fills a database following the below schema:

`TableName` (`PrimaryKey-ColumnName[data-type]`, `ForeignKey-SourceColumnName-ColumnName`, `ColumnName[data-type]`) `data source`:
	`Weekdays` (PK-`DayID`[`int`], `DayName`[`string`]) internally generated
	`Months` (PK-`MonthID`[`int`], `MonthName`[`string`]) internally generated
	`Dates` (PK-`DateID`[`int`], `Month`[`int`], `Day`[`int`], `Year`[`int`], FK-`MonthID`-`Month`, FK-`DayID`-`Day`) internally generated
	`Users` (PK-`UserID`[`unique-string`], `Password`[`hashed-string`]) input
	`Volunteers` (PK-`VolunteerID`[`int`], FK-`UserID`-`User`, `VolunteerName`[`string`]) input
	`Schedules` (PK-`ScheduleID`[`int`], FK-`UserID`-`User`, `ScheduleName`-[`string`], FK-`DateID`-`StartDate`, FK-`DateID`-`EndDate`, `ShiftsOff`[`int`], `VolunteersPerShift`[`int`]) input
	`WeekdaysForSchedule` (PK-`WFSID`[`int`], FK-`UserID`-`User`, FK-`DayID`-`Weekday`, FK-`ScheduleID`) input
	`VolunteersForSchedule` (PK-`VFSID`[`int`], FK-`UserID`-`User`, FK-`ScheduleID`-`Schedule`, FK-`VolunteerID`-`Volunteer`) input
	`UnavailabilitiesForSchedule` (PK-`UFSID`[`int`], FK-`UserID`-`User`, FK-`VFSID`-`VolunteerForSchedule`, FK-`DateID`-`Date`) input
	`CompletedSchedules` (PK-`CScheduleID`[`int`], FK-`UserID`-`User`, FK-`ScheduleID`-`Schedule`, `ScheduleData`[`json`]) output
