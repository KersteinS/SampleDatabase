package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const testDbName = "sampleTEST.db"

var sampleVolunteers = []volunteer{
	{
		VolunteerName: "Tim",
	},
	{
		VolunteerName: "Bill",
	},
	{
		VolunteerName: "Jack",
	},
	{
		VolunteerName: "George",
	},
	{
		VolunteerName: "Bob",
	},
	{
		VolunteerName: "Lance",
	},
	{
		VolunteerName: "Larry",
	},
}

func simulateCreatedSampleVolunteers(currentUser string) (result []volunteer) {
	for i, val := range sampleVolunteers {
		val.VolunteerID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	return
}

var updatedVolunteers = []volunteer{
	{VolunteerID: 1, VolunteerName: "Timmy"},
	{VolunteerID: 2, VolunteerName: "Bill"},
	{VolunteerID: 3, VolunteerName: "Jack"},
	{VolunteerID: 4, VolunteerName: "George"},
	{VolunteerID: 5, VolunteerName: "Bob"},
	{VolunteerID: 6, VolunteerName: "Lance"},
	{VolunteerID: 7, VolunteerName: "Larry"},
}

func simulateUpdatedSampleVolunteers(currentUser string) (result []volunteer) {
	for i, val := range updatedVolunteers {
		val.VolunteerID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	return
}

func generateSampleSchedules(sm SampleModel) (result []schedule) {
	result = append(result, schedule{
		ScheduleName:       "test0",
		ShiftsOff:          0,
		VolunteersPerShift: 1,
		StartDate:          Must(sm.RequestDate(date{Month: 8, Day: 1, Year: 2023})).DateID,
		EndDate:            Must(sm.RequestDate(date{Month: 9, Day: 1, Year: 2023})).DateID,
	})
	result = append(result, schedule{
		ScheduleName:       "test1",
		ShiftsOff:          3,
		VolunteersPerShift: 3,
		StartDate:          Must(sm.RequestDate(date{Month: 1, Day: 1, Year: 2024})).DateID,
		EndDate:            Must(sm.RequestDate(date{Month: 3, Day: 1, Year: 2024})).DateID,
	})
	result = append(result, schedule{
		ScheduleName:       "test2",
		ShiftsOff:          3,
		VolunteersPerShift: 3,
		StartDate:          Must(sm.RequestDate(date{Month: 3, Day: 1, Year: 2024})).DateID,
		EndDate:            Must(sm.RequestDate(date{Month: 6, Day: 1, Year: 2024})).DateID,
	})
	result = append(result, schedule{
		ScheduleName:       "test3",
		ShiftsOff:          3,
		VolunteersPerShift: 3,
		StartDate:          Must(sm.RequestDate(date{Month: 6, Day: 1, Year: 2024})).DateID,
		EndDate:            Must(sm.RequestDate(date{Month: 9, Day: 1, Year: 2024})).DateID,
	})
	return
}

func simulateCreatedSampleSchedules(currentUser string, generatedSchedules []schedule) (result []schedule) {
	for i, val := range generatedSchedules {
		val.ScheduleID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	return
}

func simulateUpdatedSampleSchedules(currentUser string, generatedSchedules []schedule) (result []schedule) {
	for i, val := range generatedSchedules {
		val.ScheduleID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	result[0].ScheduleName = "test1a"
	result[0].ShiftsOff = 10
	result[0].VolunteersPerShift = 10
	result[0].StartDate = 550
	result[0].EndDate = 556
	return
}

func generateSampleWFS(currentUser string, sm SampleModel) (result []weekdayForSchedule) {
	result = append(result, weekdayForSchedule{
		Weekday:  Must(sm.RequestWeekday(weekday{WeekdayID: 2})).WeekdayName,
		Schedule: Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test0"})).ScheduleID,
	})
	result = append(result, weekdayForSchedule{
		Weekday:  Must(sm.RequestWeekday(weekday{WeekdayName: "Sunday"})).WeekdayName,
		Schedule: Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test1"})).ScheduleID,
	})
	result = append(result, weekdayForSchedule{
		Weekday:  Must(sm.RequestWeekday(weekday{WeekdayName: "Wednesday"})).WeekdayName,
		Schedule: Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test2"})).ScheduleID,
	})
	result = append(result, weekdayForSchedule{
		Weekday:  Must(sm.RequestWeekday(weekday{WeekdayName: "Friday"})).WeekdayName,
		Schedule: Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test3"})).ScheduleID,
	})
	return
}

func simulateCreatedSampleWFS(currentUser string, generatedWFS []weekdayForSchedule) (result []weekdayForSchedule) {
	for i, val := range generatedWFS {
		val.WFSID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	return
}

func simulateUpdatedSampleWFS(currentUser string, generatedWFS []weekdayForSchedule) (result []weekdayForSchedule) {
	for i, val := range generatedWFS {
		val.WFSID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	result[0].Weekday = "Saturday"
	result[0].Schedule = 4
	return
}

func generateSampleVFS(currentUser string, sm SampleModel) (result []volunteerForSchedule) {
	result = append(result, []volunteerForSchedule{
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test1"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Tim"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test1"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Bill"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test1"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Jack"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test1"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "George"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test2"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Bob"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test2"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Lance"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test2"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Larry"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test2"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Tim"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test3"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Bill"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test3"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "Jack"})).VolunteerID,
		},
		{
			Schedule:  Must(sm.RequestSchedule(currentUser, schedule{ScheduleName: "test3"})).ScheduleID,
			Volunteer: Must(sm.RequestVolunteer(currentUser, volunteer{VolunteerName: "George"})).VolunteerID,
		}}...)
	return
}

func simulateCreatedSampleVFS(currentUser string, generatedVFS []volunteerForSchedule) (result []volunteerForSchedule) {
	for i, val := range generatedVFS {
		val.VFSID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	return
}

func simulateUpdatedSampleVFS(currentUser string, generatedVFS []volunteerForSchedule) (result []volunteerForSchedule) {
	for i, val := range generatedVFS {
		val.VFSID = i + 1
		val.User = currentUser
		result = append(result, val)
	}
	result[0].Schedule = 4
	result[0].Volunteer = 7
	return
}

func checkResultsSlice[Slice []Struct, Struct comparable](t *testing.T, ans Slice, want Slice, input Slice, err error) {
	if !slices.Equal(ans, want) {
		if err != nil {
			t.Errorf("got error: `%v`, want `%+v`", err, want)
		} else {
			t.Errorf("got %+v, want %+v", ans, want)
		}
	} else if err != nil {
		if strings.Contains(err.Error(), "sql.") {
			t.Errorf("got error: `%v` for input: `%+v`", err, input)
		} else {
			t.Logf("logged error: `%v` for input: `%+v`", err, input)
		}
	}
}

func checkResults[T comparable](t *testing.T, ans T, want T, input T, err error) {
	if ans != want {
		if err != nil {
			t.Errorf("got error: `%v`, want `%+v`", err, want)
		} else {
			t.Errorf("got %+v, want %+v", ans, want)
		}
	} else if err != nil {
		if strings.Contains(err.Error(), "sql.") {
			t.Errorf("got error: `%v` for input: `%+v`", err, input)
		} else {
			t.Logf("logged error: `%v` for input: `%+v`", err, input)
		}
	}
}

func checkResultsErrOnly[Slice []Struct, Struct comparable](t *testing.T, input any, err error, want Slice, checkFunc func(a string, b Slice) (Slice, error), a string, b Slice) {
	check, checkErr := checkFunc(a, b)
	if checkErr != nil {
		t.Errorf("got error while generating check: `%v`, want `%+v`", err, want)
	}
	//t.Logf("check: `%+v`", check)
	if !slices.Equal(check, want) {
		if err != nil {
			t.Errorf("got error: `%v`, want `%+v`", err, want)
		} else {
			t.Errorf("got %+v, want %+v", check, want)
		}
	} else if err != nil {
		if strings.Contains(err.Error(), "sql.") {
			t.Errorf("got error: `%v` for input: `%+v`", err, input)
		} else {
			t.Logf("logged error: `%v` for input: `%+v`", err, input)
		}
	}
}

func setUpEnvironment(t *testing.T) (*Env, func(t *testing.T)) {
	t.Log("running setUpEnvironment")
	sampleModel, teardown := setUpDatabaseModel(t)
	env := &Env{
		sample:       sampleModel,
		loggedInUser: "Seth",
	}
	return env, teardown
}

func setUpDatabaseModel(t *testing.T) (SampleModel, func(t *testing.T)) {
	t.Log("running setUpDatabaseModel")
	testDbPath := fmt.Sprintf("%s\\%s", t.TempDir(), testDbName)
	model, teardown := setUpDatabase(t, testDbPath)
	model.CreateDatabase()
	return model, teardown
}

func setUpDatabase(t *testing.T, testDbPath string) (SampleModel, func(t *testing.T)) {
	t.Log("running setUpDatabase")
	if _, err := os.Stat(testDbPath); err == nil { // if it finds the file, err will be nil
		suberr := os.Remove(testDbPath)
		if suberr != nil {
			t.Errorf("Error: existing testdb file was not deleted during setUpDatabase: %v", suberr)
		}
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=on", testDbPath))
	if err != nil {
		t.Errorf("Error opening testdb file using sql package: %v", err)
	}
	testSample := SampleModel{DB: db}
	return testSample, func(t *testing.T) {
		t.Log("running tearDownDatabase")
		if err = testSample.DB.Close(); err != nil {
			t.Errorf("Error: created testdb file was not closed during tearDownDatabase: %v", err)
		}
	}
}

func TestCreateDatabase(t *testing.T) {
	testDbPath := fmt.Sprintf("%s\\%s", t.TempDir(), testDbName)
	testSample, tearDownDatabaseModel := setUpDatabase(t, testDbPath)
	defer tearDownDatabaseModel(t)
	err := testSample.CreateDatabase()
	if err != nil {
		t.Errorf("Error when calling CreateDatabase: %v", err)
	}
	if err := testSample.DB.Close(); err != nil {
		t.Errorf("Error closing database after creation but before opening for hashing: %v", err)
	}
	if _, err := os.Stat(testDbPath); err != nil {
		t.Errorf("Error: testdb file was not created: %v", err)
	}
	f, err := os.Open(testDbPath)
	if err != nil {
		t.Errorf("Error opening testdb file to calculate its hash: %v", err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Errorf("Error while hashing testdb file %v", err)
	}
	if hex.EncodeToString(h.Sum(nil)) != "5c26b4eea141e9daa7ef3b814662e30ceadb4d505bf1f0ce17a5d80ef08c8ccf" {
		t.Errorf("Error: test testdb file does not match stored hash value. Computed hash: %x", h.Sum(nil))
	}
	if err = f.Close(); err != nil {
		t.Errorf("Error closing database file after hashing: %v", err)
	}
}

func TestRequestWeekday(t *testing.T) {
	testSample, tearDownDatabaseModel := setUpDatabaseModel(t)
	defer tearDownDatabaseModel(t)
	var tests = []struct {
		name  string
		input weekday
		want  weekday
	}{
		{"Get Sunday", weekday{WeekdayName: "Sunday"}, weekday{WeekdayID: 1, WeekdayName: "Sunday"}},
		{"Get Monday", weekday{WeekdayName: "Monday"}, weekday{WeekdayID: 2, WeekdayName: "Monday"}},
		{"Get Tuesday", weekday{WeekdayName: "Tuesday"}, weekday{WeekdayID: 3, WeekdayName: "Tuesday"}},
		{"Get Wednesday", weekday{WeekdayName: "Wednesday"}, weekday{WeekdayID: 4, WeekdayName: "Wednesday"}},
		{"Get Thursday", weekday{WeekdayName: "Thursday"}, weekday{WeekdayID: 5, WeekdayName: "Thursday"}},
		{"Get Friday", weekday{WeekdayName: "Friday"}, weekday{WeekdayID: 6, WeekdayName: "Friday"}},
		{"Get Saturday", weekday{WeekdayName: "Saturday"}, weekday{WeekdayID: 7, WeekdayName: "Saturday"}},
		{"Test Bad WeekdayID Error", weekday{WeekdayID: 8}, weekday{}},
		{"Test Bad WeekdayNameError", weekday{WeekdayName: "Thorsday"}, weekday{}},
		{"Test Disagreeing WeekdayID and WeekdayName", weekday{WeekdayID: 1, WeekdayName: "Monday"}, weekday{}},
		{"Test Empty Input Error", weekday{}, weekday{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := testSample.RequestWeekday(tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestDate(t *testing.T) {
	testSample, tearDownDatabaseModel := setUpDatabaseModel(t)
	defer tearDownDatabaseModel(t)
	var tests = []struct {
		name  string
		input date
		want  date
	}{
		{name: "Get 9/9/2024", input: date{Month: 9, Day: 9, Year: 2024}, want: date{DateID: 618, Month: 9, Day: 9, Year: 2024, Weekday: "Monday"}},
		{name: "Get DateID 618", input: date{DateID: 618}, want: date{DateID: 618, Month: 9, Day: 9, Year: 2024, Weekday: "Monday"}},
		{name: "Fail to get DateID 6180000", input: date{DateID: 6180000}, want: date{}},
		{name: "Get more than one Date", input: date{Month: 9}, want: date{}},
		{name: "Don't provide any date fields", input: date{}, want: date{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := testSample.RequestDate(tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestDates(t *testing.T) {
	testSample, tearDownDatabaseModel := setUpDatabaseModel(t)
	defer tearDownDatabaseModel(t)
	var tests = []struct {
		name  string
		input []date
		want  []date
	}{
		{name: "Get 9/9/2024", input: []date{{Month: 9, Day: 9, Year: 2024}}, want: []date{{DateID: 618, Month: 9, Day: 9, Year: 2024, Weekday: "Monday"}}},
		{name: "Get DateID 618", input: []date{{DateID: 618}}, want: []date{{DateID: 618, Month: 9, Day: 9, Year: 2024, Weekday: "Monday"}}},
		{name: "Fail to get DateID 6180000", input: []date{{DateID: 6180000}}, want: []date{}},
		{name: "Fail to get DateID 618 due to wrong day field", input: []date{{DateID: 618, Day: 1}}, want: []date{}},
		{name: "Provide empty slice", input: []date{}, want: []date{}},
		{name: "Get all Mondays in July 2024", input: []date{{Month: 7, Year: 2024, Weekday: "Monday"}}, want: []date{{DateID: 548, Month: 7, Day: 1, Year: 2024, Weekday: "Monday"}, {DateID: 555, Month: 7, Day: 8, Year: 2024, Weekday: "Monday"}, {DateID: 562, Month: 7, Day: 15, Year: 2024, Weekday: "Monday"}, {DateID: 569, Month: 7, Day: 22, Year: 2024, Weekday: "Monday"}, {DateID: 576, Month: 7, Day: 29, Year: 2024, Weekday: "Monday"}}},
		{name: "Get DateID 618 and DateID 619", input: []date{{DateID: 618}, {DateID: 619}}, want: []date{{DateID: 618, Month: 9, Day: 9, Year: 2024, Weekday: "Monday"}, {DateID: 619, Month: 9, Day: 10, Year: 2024, Weekday: "Tuesday"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := testSample.RequestDates(tt.input)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestCreateVolunteers(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	tests := []struct {
		name  string
		input []volunteer
		want  []volunteer
	}{
		{name: "Create volunteers from sampleVolunteers", input: sampleVolunteers, want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Create volunteers from sampleVolunteers and one duplicate volunteer", input: append(sampleVolunteers, volunteer{VolunteerName: "Tim"}), want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Create volunteers but provide no volunteers", input: []volunteer{}, want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Create volunteers from sampleVolunteers but provide one empty volunteer", input: append(sampleVolunteers, volunteer{}), want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Only provide one empty volunteer", input: []volunteer{{}}, want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail to provide VolunteerName", input: []volunteer{{User: "Anyone"}}, want: simulateCreatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail by providing duplicate input", input: []volunteer{{VolunteerName: "Anyone"}, {User: "Doesn'tMatter", VolunteerName: "Anyone"}}, want: simulateCreatedSampleVolunteers(env.loggedInUser)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CreateVolunteers(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVolunteers, env.loggedInUser, []volunteer{})
		})
	}
}

func TestRequestVolunteer(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	err := env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	tests := []struct {
		name  string
		input volunteer
		want  volunteer
	}{
		{name: "Request Tim", input: volunteer{VolunteerName: "Tim"}, want: volunteer{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}},
		{name: "Request VolunteerID 1", input: volunteer{VolunteerID: 1}, want: volunteer{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}},
		{name: "Request volunteer with empty struct", input: volunteer{}, want: volunteer{}},
		{name: "Request volunteer with invalid VolunteerID", input: volunteer{VolunteerID: 100}, want: volunteer{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestVolunteer(env.loggedInUser, tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestVolunteers(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	err := env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	tests := []struct {
		name  string
		input []volunteer
		want  []volunteer
	}{
		{name: "Get 1 volunteer by VolunteerName", input: []volunteer{{VolunteerName: "Tim"}}, want: []volunteer{{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}}},
		{name: "Get 1 volunteer by VolunteerID", input: []volunteer{{VolunteerID: 1}}, want: []volunteer{{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}}},
		{name: "Get 1 volunteer by VolunteerID and Volunteer Name and User", input: []volunteer{{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}}, want: []volunteer{{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser}}},
		{name: "Get all volunteers", input: []volunteer{}, want: []volunteer{
			{VolunteerID: 1, VolunteerName: "Tim", User: env.loggedInUser},
			{VolunteerID: 2, VolunteerName: "Bill", User: env.loggedInUser},
			{VolunteerID: 3, VolunteerName: "Jack", User: env.loggedInUser},
			{VolunteerID: 4, VolunteerName: "George", User: env.loggedInUser},
			{VolunteerID: 5, VolunteerName: "Bob", User: env.loggedInUser},
			{VolunteerID: 6, VolunteerName: "Lance", User: env.loggedInUser},
			{VolunteerID: 7, VolunteerName: "Larry", User: env.loggedInUser},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestVolunteers(env.loggedInUser, tt.input)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestUpdateVolunteers(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	err := env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	tests := []struct {
		name  string
		input []volunteer
		want  []volunteer
	}{
		{name: "Update 1 volunteer's VolunteerName by VolunteerID", input: []volunteer{{VolunteerID: 1, VolunteerName: "Timmy"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail to update 1 volunteer by providing only VolunteerID", input: []volunteer{{VolunteerID: 1}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail to update 1 volunteer by not providing VolunteerID", input: []volunteer{{VolunteerName: "Tim"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail to update because of an empty volunteer struct", input: []volunteer{{}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Update volunteers but don't provide any volunteers", input: []volunteer{}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Fail to create a duplicate volunteer (same User and VolunteerName, different VolunteerID)", input: []volunteer{{VolunteerID: 2, VolunteerName: "Timmy"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)},
		{name: "Try to update a nonexistent volunteer", input: []volunteer{{VolunteerID: 10, VolunteerName: "Timmy"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)}, // the query gets no matches; so no error, but also no output
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.UpdateVolunteers(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVolunteers, env.loggedInUser, []volunteer{})
		})
	}
}

func TestDeleteVolunteers(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	err := env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	tests := []struct {
		name  string
		input []volunteer
		want  []volunteer
	}{
		{name: "Delete 1 volunteer by VolunteerID", input: []volunteer{{VolunteerID: 1}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)[1:]},
		{name: "Fail due to empty input struct", input: []volunteer{{}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)[1:]},
		{name: "Fail due to empty input slice", input: []volunteer{}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)[1:]},
		{name: "Delete 2 volunteers by name", input: []volunteer{{VolunteerName: "Bill"}, {VolunteerName: "Jack"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)[3:]},
		{name: "Fail by not providing neither VolunteerID nor VolunteerName", input: []volunteer{{User: "Doesn'tMatter"}}, want: simulateUpdatedSampleVolunteers(env.loggedInUser)[3:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.DeleteVolunteers(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVolunteers, env.loggedInUser, []volunteer{})
		})
	}
}

func TestCleanOrphanedVolunteers(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	err := env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	tests := []struct {
		name  string
		input []volunteer
		want  []volunteer
	}{
		{name: "Clean Orphaned Volunteers", input: simulateCreatedSampleVolunteers(env.loggedInUser), want: []volunteer{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CleanOrphanedVolunteers(env.loggedInUser)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVolunteers, env.loggedInUser, []volunteer{})
		})
	}
}

func TestCreateSchedulesExtended(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Create schedules from sampleSchedules", input: generatedSampleSchedules, want: simulatedCreatedSampleSchedules},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CreateSchedulesExtended(env.loggedInUser, tt.input, true)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestSchedules, env.loggedInUser, []schedule{})
		})
	}
}

func TestCreateSchedules(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)[1:]
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Create schedules from sampleSchedules", input: generatedSampleSchedules, want: simulatedCreatedSampleSchedules},
		{name: "Fail to create schedules from duplicate schedule", input: []schedule{generatedSampleSchedules[0]}, want: simulatedCreatedSampleSchedules},
		{name: "Create schedules but provide no schedule structs", input: []schedule{}, want: simulatedCreatedSampleSchedules},
		{name: "Only provide one empty schedule struct", input: []schedule{{}}, want: simulatedCreatedSampleSchedules},
		{name: "Do not provide ScheduleName", input: []schedule{
			{
				ShiftsOff:          4,
				VolunteersPerShift: 4,
				StartDate:          Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
				EndDate:            Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Provide negative ShiftsOff", input: []schedule{
			{
				ScheduleName:       "test4",
				VolunteersPerShift: 4,
				StartDate:          Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
				EndDate:            Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Do not provide VolunteersPerShift", input: []schedule{
			{
				ScheduleName: "test4",
				ShiftsOff:    4,
				StartDate:    Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
				EndDate:      Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Do not provide StartDate", input: []schedule{
			{
				ScheduleName:       "test4",
				ShiftsOff:          4,
				VolunteersPerShift: 4,
				EndDate:            Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Do not provide EndDate", input: []schedule{
			{
				ScheduleName:       "test4",
				ShiftsOff:          4,
				VolunteersPerShift: 4,
				StartDate:          Must(env.sample.RequestDate(date{Month: 4, Day: 4, Year: 2024})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Provide ShiftsOff = 0", input: []schedule{{ScheduleName: "test01", ShiftsOff: 0, VolunteersPerShift: 1, User: "Seth", StartDate: 213, EndDate: 244}}, want: simulatedCreatedSampleSchedules},
		{name: "Fail by providing duplicate input", input: []schedule{
			{
				ScheduleName: "test01", ShiftsOff: 3, VolunteersPerShift: 1, User: "Seth", StartDate: 213, EndDate: 244,
			},
			{
				ScheduleName: "test01", ShiftsOff: 3, VolunteersPerShift: 1, User: "Doesn'tMatter", StartDate: 213, EndDate: 244,
			}},
			want: simulatedCreatedSampleSchedules},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CreateSchedules(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestSchedules, env.loggedInUser, []schedule{})
		})
	}
}

func TestRequestSchedulesExtended(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Request all schedules", input: []schedule{}, want: simulatedCreatedSampleSchedules},
		{name: "Request all schedules with User", input: []schedule{{ShiftsOff: -1, User: "Seth"}}, want: simulatedCreatedSampleSchedules},
		{name: "Request fully specified schedule", input: []schedule{{ScheduleID: 1, ScheduleName: "test0", ShiftsOff: 0, VolunteersPerShift: 1, User: "Seth", StartDate: 213, EndDate: 244}}, want: simulatedCreatedSampleSchedules[:1]},
		{name: "Fail due to empty/default struct (manually set ShiftsOff to -1)", input: []schedule{{ShiftsOff: -1}}, want: []schedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestSchedulesExtended(env.loggedInUser, tt.input, true)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestSchedules(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)[1:]
	err := env.sample.CreateSchedules(env.loggedInUser, generatedSampleSchedules)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedules failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Request all schedules", input: []schedule{}, want: simulatedCreatedSampleSchedules},
		{name: "Fail due to manually setting ShiftsOff to -1 when using RequestSchedule", input: []schedule{{ShiftsOff: -1}}, want: []schedule{}},
		{name: "Fail by providing empty input struct", input: []schedule{{}}, want: []schedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestSchedules(env.loggedInUser, tt.input)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestSchedule(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)[1:]
	err := env.sample.CreateSchedules(env.loggedInUser, generatedSampleSchedules)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedules failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input schedule
		want  schedule
	}{
		{name: "Request schedule 1", input: schedule{ScheduleID: 1}, want: simulatedCreatedSampleSchedules[0]},
		{name: "Fail by providing empty input struct", input: schedule{}, want: schedule{}},
		{name: "Fail due to manually setting ShiftsOff to -1 when using RequestSchedule", input: schedule{ShiftsOff: -1}, want: schedule{}},
		{name: "Fail by requesting more than one schedule", input: schedule{User: "Seth"}, want: schedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestSchedule(env.loggedInUser, tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestUpdateSchedulesExtended(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	simulatedUpdatedSampleSchedules := simulateUpdatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Update 1 schedule", input: []schedule{{ScheduleID: 1, ScheduleName: "test1a", ShiftsOff: 10, VolunteersPerShift: 10, StartDate: 550, EndDate: 556}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by providing an empty input struct", input: []schedule{{ShiftsOff: -1}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by not providing a ScheduleID", input: []schedule{{ScheduleName: "test1", ShiftsOff: 4}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by only providing a ScheduleID", input: []schedule{{ScheduleID: 1, ShiftsOff: -1}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule because it would create a duplicate schedule", input: []schedule{{ScheduleID: 2, ScheduleName: "test1a", ShiftsOff: 10, VolunteersPerShift: 10, StartDate: 550, EndDate: 556}}, want: simulatedUpdatedSampleSchedules},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.UpdateSchedulesExtended(env.loggedInUser, tt.input, true)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestSchedules, env.loggedInUser, []schedule{})
		})
	}
}

func TestUpdateSchedules(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)[1:]
	err := env.sample.CreateSchedules(env.loggedInUser, generatedSampleSchedules)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedules failed): %v", err)
		t.FailNow()
	}
	simulatedUpdatedSampleSchedules := simulateUpdatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Update 1 schedule", input: []schedule{{ScheduleID: 1, ScheduleName: "test1a", ShiftsOff: 10, VolunteersPerShift: 10, StartDate: 550, EndDate: 556}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by providing an empty input struct", input: []schedule{{}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by not providing a ScheduleID", input: []schedule{{ScheduleName: "test1", ShiftsOff: 4}}, want: simulatedUpdatedSampleSchedules},
		{name: "Fail to update a schedule by only providing a ScheduleID", input: []schedule{{ScheduleID: 1}}, want: simulatedUpdatedSampleSchedules},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.UpdateSchedules(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestSchedules, env.loggedInUser, []schedule{})
		})
	}
}

func TestDeleteSchedules(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleSchedules := simulateCreatedSampleSchedules(env.loggedInUser, generatedSampleSchedules)
	tests := []struct {
		name  string
		input []schedule
		want  []schedule
	}{
		{name: "Fail by providing an empty input struct", input: []schedule{{}}, want: simulatedCreatedSampleSchedules},
		{name: "Fail by providing an everything but ScheduleID and ScheduleName", input: []schedule{
			{
				ShiftsOff:          0,
				VolunteersPerShift: 1,
				StartDate:          Must(env.sample.RequestDate(date{Month: 8, Day: 1, Year: 2023})).DateID,
				EndDate:            Must(env.sample.RequestDate(date{Month: 9, Day: 1, Year: 2023})).DateID,
			}},
			want: simulatedCreatedSampleSchedules},
		{name: "Delete one schedule by ScheduleID", input: []schedule{{ScheduleID: 1}}, want: simulatedCreatedSampleSchedules[1:]},
		{name: "Delete one schedule by ScheduleName", input: []schedule{{ScheduleName: "test1"}}, want: simulatedCreatedSampleSchedules[2:]},
		{name: "Fail to delete all schedules", input: []schedule{}, want: simulatedCreatedSampleSchedules[2:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.DeleteSchedules(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestSchedules, env.loggedInUser, []schedule{})
		})
	}
}

func TestCreateWFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	simulatedCreatedSampleWFS := simulateCreatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input []weekdayForSchedule
		want  []weekdayForSchedule
	}{
		{name: "Create WFS from sampleWFS", input: generatedSampleWFS, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS from duplicate WFS", input: []weekdayForSchedule{generatedSampleWFS[0]}, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS by providing no wfs structs", input: []weekdayForSchedule{}, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS by providing one empty wfs struct", input: []weekdayForSchedule{{}}, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS by not providing a Weekday", input: []weekdayForSchedule{{Schedule: 5}}, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS by not providing a Schedule", input: []weekdayForSchedule{{Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Thursday"})).WeekdayName}}, want: simulatedCreatedSampleWFS},
		{name: "Fail to create WFS by providing a duplicate input", input: []weekdayForSchedule{{Weekday: "Thursday", Schedule: 5}, {User: "Doesn'tMatter", Weekday: "Thursday", Schedule: 5}}, want: simulatedCreatedSampleWFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CreateWFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestWFS, env.loggedInUser, []weekdayForSchedule{})
		})
	}
}

func TestRequestWFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	err = env.sample.CreateWFS(env.loggedInUser, generatedSampleWFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateWFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleWFS := simulateCreatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input []weekdayForSchedule
		want  []weekdayForSchedule
	}{
		{name: "Request all WFS", input: []weekdayForSchedule{}, want: simulatedCreatedSampleWFS},
		{name: "Request a fully specified WFS", input: simulatedCreatedSampleWFS[:1], want: simulatedCreatedSampleWFS[:1]},
		{name: "Fail by requesting an empty WFS", input: []weekdayForSchedule{{}}, want: []weekdayForSchedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestWFS(env.loggedInUser, tt.input)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestWFSSingle(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	err = env.sample.CreateWFS(env.loggedInUser, generatedSampleWFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateWFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleWFS := simulateCreatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input weekdayForSchedule
		want  weekdayForSchedule
	}{
		{name: "Request a fully specified WFS", input: simulatedCreatedSampleWFS[1], want: simulatedCreatedSampleWFS[1]},
		{name: "Fail by requesting an empty WFS", input: weekdayForSchedule{}, want: weekdayForSchedule{}},
		{name: "Fail by requesting multiple WFS", input: weekdayForSchedule{User: env.loggedInUser}, want: weekdayForSchedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestWFSSingle(env.loggedInUser, tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestUpdateWFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	err = env.sample.CreateWFS(env.loggedInUser, generatedSampleWFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateWFS failed): %v", err)
		t.FailNow()
	}
	simulatedUpdatedSampleWFS := simulateUpdatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input []weekdayForSchedule
		want  []weekdayForSchedule
	}{
		{name: "Update 1 WFS", input: []weekdayForSchedule{{WFSID: 1, Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Saturday"})).WeekdayName, Schedule: 4}}, want: simulatedUpdatedSampleWFS},
		{name: "Update 1 WFS schedule", input: []weekdayForSchedule{{WFSID: 1, Schedule: 4}}, want: simulatedUpdatedSampleWFS},
		{name: "Update 1 WFS weekday", input: []weekdayForSchedule{{WFSID: 1, Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Saturday"})).WeekdayName}}, want: simulatedUpdatedSampleWFS},
		{name: "Fail to update by only providing one value in WFS", input: []weekdayForSchedule{{WFSID: 1}}, want: simulatedUpdatedSampleWFS},
		{name: "Fail to update by not providing WFSID", input: []weekdayForSchedule{{Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Saturday"})).WeekdayName}}, want: simulatedUpdatedSampleWFS},
		{name: "Fail to update by providing an empty WFS struct", input: []weekdayForSchedule{{}}, want: simulatedUpdatedSampleWFS},
		{name: "Fail to update by providing an empty WFS slice", input: []weekdayForSchedule{}, want: simulatedUpdatedSampleWFS},
		{name: "Fail to update because it would create a duplicate WFS", input: []weekdayForSchedule{{WFSID: 2, Schedule: 4, Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Saturday"})).WeekdayName}}, want: simulatedUpdatedSampleWFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.UpdateWFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestWFS, env.loggedInUser, []weekdayForSchedule{})
		})
	}
}

func TestDeleteWFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	err = env.sample.CreateWFS(env.loggedInUser, generatedSampleWFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateWFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleWFS := simulateCreatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input []weekdayForSchedule
		want  []weekdayForSchedule
	}{
		{name: "Delete one WFS by WFSID", input: []weekdayForSchedule{{WFSID: 1}}, want: simulatedCreatedSampleWFS[1:]},
		{name: "Delete one WFS by Schedule and Weekday", input: []weekdayForSchedule{{Schedule: 2, Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Sunday"})).WeekdayName}}, want: simulatedCreatedSampleWFS[2:]},
		{name: "Fail to delete one WFS by WFSID", input: []weekdayForSchedule{{WFSID: 1}}, want: simulatedCreatedSampleWFS[2:]},
		{name: "Fail to delete one WFS by providing only Schedule", input: []weekdayForSchedule{{Schedule: 3}}, want: simulatedCreatedSampleWFS[2:]},
		{name: "Fail to delete by not providing any WFS structs", input: []weekdayForSchedule{}, want: simulatedCreatedSampleWFS[2:]},
		{name: "Fail to delete by providing empty WFS struct", input: []weekdayForSchedule{{}}, want: simulatedCreatedSampleWFS[2:]},
		{name: "Fail to delete by not providing Schedule nor Weekday nor WFSID", input: []weekdayForSchedule{{User: "Doesn'tMatter"}}, want: simulatedCreatedSampleWFS[2:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.DeleteWFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestWFS, env.loggedInUser, []weekdayForSchedule{})
		})
	}
}

func TestCleanOrphanedWFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	generatedSampleWFS := generateSampleWFS(env.loggedInUser, env.sample)
	plusOrphanWFS := append(generatedSampleWFS, weekdayForSchedule{Schedule: Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test1"})).ScheduleID, Weekday: Must(env.sample.RequestWeekday(weekday{WeekdayName: "Friday"})).WeekdayName})
	err = env.sample.CreateWFS(env.loggedInUser, plusOrphanWFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateWFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleWFS := simulateCreatedSampleWFS(env.loggedInUser, generatedSampleWFS)
	tests := []struct {
		name  string
		input []map[schedule][]weekday
		want  []weekdayForSchedule
	}{
		{name: "Clean Orphaned WFS", input: []map[schedule][]weekday{
			{
				Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test1"})): []weekday{Must(env.sample.RequestWeekday(weekday{WeekdayName: "Sunday"}))},
			},
			{
				Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test2"})): []weekday{Must(env.sample.RequestWeekday(weekday{WeekdayName: "Wednesday"}))},
			}},
			want: simulatedCreatedSampleWFS},
		{name: "Fail by not providing a schedule with a ScheduleID", input: []map[schedule][]weekday{
			{
				schedule{ScheduleName: "test1"}: []weekday{Must(env.sample.RequestWeekday(weekday{WeekdayName: "Sunday"}))},
			}}, want: simulatedCreatedSampleWFS},
		{name: "Fail by not providing a weekday with a WeekdayName", input: []map[schedule][]weekday{
			{
				Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test2"})): []weekday{{WeekdayID: 3}},
			}}, want: simulatedCreatedSampleWFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CleanOrphanedWFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestWFS, env.loggedInUser, []weekdayForSchedule{})
		})
	}
}

func TestCreateVFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	simulatedCreatedSampleVFS := simulateCreatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input []volunteerForSchedule
		want  []volunteerForSchedule
	}{
		{name: "Create VFS", input: generatedSampleVFS, want: simulatedCreatedSampleVFS},
		{name: "Fail by trying to create an existing VFS", input: []volunteerForSchedule{generatedSampleVFS[0]}, want: simulatedCreatedSampleVFS},
		{name: "Fail by providing duplicate inputs", input: []volunteerForSchedule{{Schedule: 2, Volunteer: 6}, {User: "Doesn'tMatter", Schedule: 2, Volunteer: 6}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by not providing a Schedule", input: []volunteerForSchedule{{User: "Anybody", Volunteer: 6}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by not providing a Volunteer", input: []volunteerForSchedule{{User: "Anybody", Schedule: 2}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by providing an empty/default values VFS struct", input: []volunteerForSchedule{{}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by providing no input", input: []volunteerForSchedule{}, want: simulatedCreatedSampleVFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CreateVFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVFS, env.loggedInUser, []volunteerForSchedule{})
		})
	}
}

func TestRequestVFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	err = env.sample.CreateVFS(env.loggedInUser, generatedSampleVFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateVFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleVFS := simulateCreatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input []volunteerForSchedule
		want  []volunteerForSchedule
	}{
		{name: "Request all VFS", input: []volunteerForSchedule{}, want: simulatedCreatedSampleVFS},
		{name: "Request a fully specified VFS", input: simulatedCreatedSampleVFS[:1], want: simulatedCreatedSampleVFS[:1]},
		{name: "Fail by requesting an empty VFS", input: []volunteerForSchedule{{}}, want: []volunteerForSchedule{}},
		{name: "Request a nonexistent VFS", input: []volunteerForSchedule{{Schedule: 100}}, want: []volunteerForSchedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestVFS(env.loggedInUser, tt.input)
			checkResultsSlice(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestRequestVFSSingle(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	err = env.sample.CreateVFS(env.loggedInUser, generatedSampleVFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateVFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleVFS := simulateCreatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input volunteerForSchedule
		want  volunteerForSchedule
	}{
		{name: "Request a fully specified VFS", input: simulatedCreatedSampleVFS[1], want: simulatedCreatedSampleVFS[1]},
		{name: "Fail by requesting an empty VFS", input: volunteerForSchedule{}, want: volunteerForSchedule{}},
		{name: "Fail by requesting an multiple VFS", input: volunteerForSchedule{User: env.loggedInUser}, want: volunteerForSchedule{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := env.sample.RequestVFSSingle(env.loggedInUser, tt.input)
			checkResults(t, ans, tt.want, tt.input, err)
		})
	}
}

func TestUpdateVFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	err = env.sample.CreateVFS(env.loggedInUser, generatedSampleVFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateVFS failed): %v", err)
		t.FailNow()
	}
	simulatedUpdatedSampleVFS := simulateUpdatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input []volunteerForSchedule
		want  []volunteerForSchedule
	}{
		{name: "Update 1 VFS", input: []volunteerForSchedule{{VFSID: 1, Schedule: Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test3"})).ScheduleID, Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Larry"})).VolunteerID}}, want: simulatedUpdatedSampleVFS},
		{name: "Update 1 VFS schedule", input: []volunteerForSchedule{{VFSID: 1, Schedule: Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test3"})).ScheduleID}}, want: simulatedUpdatedSampleVFS},
		{name: "Update 1 VFS weekday", input: []volunteerForSchedule{{VFSID: 1, Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Larry"})).VolunteerID}}, want: simulatedUpdatedSampleVFS},
		{name: "Fail to update by only providing one value in VFS", input: []volunteerForSchedule{{VFSID: 1}}, want: simulatedUpdatedSampleVFS},
		{name: "Fail to update by not providing VFSID", input: []volunteerForSchedule{{Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Larry"})).VolunteerID}}, want: simulatedUpdatedSampleVFS},
		{name: "Fail to update by providing an empty VFS struct", input: []volunteerForSchedule{{}}, want: simulatedUpdatedSampleVFS},
		{name: "Fail to update by providing an empty VFS slice", input: []volunteerForSchedule{}, want: simulatedUpdatedSampleVFS},
		{name: "Fail to update because it would create a duplicate VFS", input: []volunteerForSchedule{{VFSID: 3, Schedule: Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test3"})).ScheduleID, Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Larry"})).VolunteerID}}, want: simulatedUpdatedSampleVFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.UpdateVFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVFS, env.loggedInUser, []volunteerForSchedule{})
		})
	}
}

func TestDeleteVFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	err = env.sample.CreateVFS(env.loggedInUser, generatedSampleVFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateVFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleVFS := simulateCreatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input []volunteerForSchedule
		want  []volunteerForSchedule
	}{
		{name: "Delete one VFS by VFSID", input: []volunteerForSchedule{{VFSID: 1}}, want: simulatedCreatedSampleVFS[1:]},
		{name: "Delete one VFS by Schedule and Volunteer", input: []volunteerForSchedule{{Schedule: 2, Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Bill"})).VolunteerID}}, want: simulatedCreatedSampleVFS[2:]},
		{name: "Fail to delete one VFS by VFSID", input: []volunteerForSchedule{{VFSID: 1}}, want: simulatedCreatedSampleVFS[2:]},
		{name: "Fail to delete one VFS by providing only Schedule", input: []volunteerForSchedule{{Schedule: 3}}, want: simulatedCreatedSampleVFS[2:]},
		{name: "Fail to delete by not providing any VFS structs", input: []volunteerForSchedule{}, want: simulatedCreatedSampleVFS[2:]},
		{name: "Fail to delete by providing empty VFS struct", input: []volunteerForSchedule{{}}, want: simulatedCreatedSampleVFS[2:]},
		{name: "Fail to delete by not providing Schedule nor Volunteer nor VFSID", input: []volunteerForSchedule{{User: "Doesn'tMatter"}}, want: simulatedCreatedSampleVFS[2:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.DeleteVFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVFS, env.loggedInUser, []volunteerForSchedule{})
		})
	}
}

func TestCleanOrphanedVFS(t *testing.T) {
	env, tearDownEnvironment := setUpEnvironment(t)
	defer tearDownEnvironment(t)
	generatedSampleSchedules := generateSampleSchedules(env.sample)
	err := env.sample.CreateSchedulesExtended(env.loggedInUser, generatedSampleSchedules, true)
	if err != nil {
		t.Errorf("Error setting up test (CreateSchedulesExtended failed): %v", err)
		t.FailNow()
	}
	err = env.sample.CreateVolunteers(env.loggedInUser, sampleVolunteers)
	if err != nil {
		t.Errorf("Error setting up test (CreateVolunteers failed): %v", err)
		t.FailNow()
	}
	generatedSampleVFS := generateSampleVFS(env.loggedInUser, env.sample)
	plusOrphanVFS := append(generatedSampleVFS, volunteerForSchedule{Schedule: Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test1"})).ScheduleID, Volunteer: Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Larry"})).VolunteerID})
	err = env.sample.CreateVFS(env.loggedInUser, plusOrphanVFS)
	if err != nil {
		t.Errorf("Error setting up test (CreateVFS failed): %v", err)
		t.FailNow()
	}
	simulatedCreatedSampleVFS := simulateCreatedSampleVFS(env.loggedInUser, generatedSampleVFS)
	tests := []struct {
		name  string
		input []map[schedule][]volunteer
		want  []volunteerForSchedule
	}{
		{name: "Clean Orphaned VFS", input: []map[schedule][]volunteer{
			{
				Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test1"})): []volunteer{
					Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Tim"})),
					Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Bill"})),
					Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Jack"})),
					Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "George"})),
				},
			}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by not providing a schedule with a ScheduleID", input: []map[schedule][]volunteer{
			{
				schedule{ScheduleName: "test1"}: []volunteer{Must(env.sample.RequestVolunteer(env.loggedInUser, volunteer{VolunteerName: "Tim"}))},
			}}, want: simulatedCreatedSampleVFS},
		{name: "Fail by not providing a Volunteer with a VolunteerName", input: []map[schedule][]volunteer{
			{
				Must(env.sample.RequestSchedule(env.loggedInUser, schedule{ScheduleName: "test2"})): []volunteer{{VolunteerID: 5}},
			}}, want: simulatedCreatedSampleVFS},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.sample.CleanOrphanedVFS(env.loggedInUser, tt.input)
			checkResultsErrOnly(t, tt.input, err, tt.want, env.sample.RequestVFS, env.loggedInUser, []volunteerForSchedule{})
		})
	}
}
