/******************************************************************************
Cloud Resource Counter
File: results.go

Summary: Collects results (in the form of column names and column values) and
         writes to a CSV file
******************************************************************************/

package main

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"

	"github.com/expel-io/cloud-resource-counter/mock"
)

// Test the initialization of Results struct and invocation of NewRow
func TestResultsInitAndNewRow(t *testing.T) {
	// Create our test cases
	cases := []struct {
		StoreHeaders bool
		ExpectedRows int
	}{
		{
			StoreHeaders: true,
			ExpectedRows: 1,
		},
		{
			ExpectedRows: 0,
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create an instance of Results
		results := Results{
			StoreHeaders: c.StoreHeaders,
		}
		results.Init()

		// How many rows are there?
		actualRows := len(results.Rows)

		// Does it match?
		if actualRows != c.ExpectedRows {
			t.Errorf("Error: TestResultsInit (before NewRow()) returned %d; expected %d", actualRows, c.ExpectedRows)
		}

		// Let's add another row and try again...
		results.NewRow()

		// How many rows are there?
		actualRows = len(results.Rows)

		// Does it match?
		if actualRows != c.ExpectedRows+1 {
			t.Errorf("Error: TestResultsInit (after NewRow()) returned %d; expected %d", actualRows, c.ExpectedRows+1)
		}
	}
}

func TestResultsAppend(t *testing.T) {
	// Create a Builder to hold our generated results
	builder := strings.Builder{}

	// Create an instance of Results
	results := Results{
		StoreHeaders: true,
		Writer:       &builder,
	}
	results.Init()
	results.NewRow()

	// Add a bunch of column names and row values...
	results.Append("col1", "row1_col1")
	results.Append("col2", 123)
	results.Append("col3", 456)
	results.Append("col4", 789)

	// Create our mock activity monitor
	mon := mock.ActivityMonitorImpl{}

	// Save to our mock Writer
	results.Save(&mon)

	// Verify that no errors were encountered
	if mon.ErrorOccured {
		t.Errorf("Encountered an error during Results.Save: %s", mon.ErrorMessage)
	}

	// Read the generate results back into our CSV module...
	csvReader := csv.NewReader(strings.NewReader(builder.String()))

	// Expected rows...
	expected := [][]string{
		{
			"col1", "col2", "col3", "col4",
		}, {
			"row1_col1", "123", "456", "789",
		},
	}

	row := 0
	for {
		// Read a Row of data...
		record, err := csvReader.Read()

		// Have we read all of the rows?
		if err == io.EOF {
			break
		}
		// Do we have an error?
		if err != nil {
			t.Errorf("Unexpected error while reading row of data: %v", err)
		}

		// Are we asking for a row which we don't expect?
		if row == len(expected) {
			t.Errorf("We have encountered more stored rows than expected")
		}

		// Are the length the same?
		if len(record) != len(expected[row]) {
			t.Errorf("Unexpected number of columns: expected %d, got %d", len(expected[row]), len(record))
		}

		// Does the record match?
		for ix, actualVal := range record {
			if expected[row][ix] != actualVal {
				t.Errorf("Unexpected value for row %d, column %d: expected %v, got %v", row, ix, expected[row][ix], actualVal)
			}
		}

		// Increment our row counter
		row = row + 1
	}

	// Did we look at all of the expected rows?
	if row != len(expected) {
		t.Errorf("We have fewer stored rows than expected")
	} else if mon.ProgramExited {
		t.Errorf("Unexpected Exit: The program unexpected exited with status code=%d", mon.ExitCode)
	}
}

func TestResultsNoSave(t *testing.T) {
	// Create an instance of Results
	results := Results{}
	results.Init()
	results.NewRow()

	// Add a bunch of column names and row values...
	results.Append("col1", "row1_col1")
	results.Append("col2", 123)
	results.Append("col3", 456)
	results.Append("col4", 789)

	// Create our mock activity monitor
	mon := mock.ActivityMonitorImpl{}

	// Save to our mock Writer
	results.Save(&mon)

	// Verify that no errors were encountered
	if mon.ErrorOccured {
		t.Errorf("Encountered an error during Results.Save: %s", mon.ErrorMessage)
	}
}
