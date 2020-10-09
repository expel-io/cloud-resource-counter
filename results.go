/******************************************************************************
Cloud Resource Counter
File: results.go

Summary: Collects results (in the form of column names and column values) and
         writes to a CSV file
******************************************************************************/

package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

// Results is a struct that collects rows of data and writes them to the supplied
// file in CSV format.
type Results struct {
	Rows         [][]string
	StoreHeaders bool
	Writer       io.Writer
}

// Init performs one-time initialization on the results struct.
func (r *Results) Init() {
	// Are storing column names in our rows?
	if r.StoreHeaders {
		// Create a new row to hold them
		r.NewRow()
	}
}

// NewRow creates a new row to receive results
func (r *Results) NewRow() {
	r.Rows = append(r.Rows, []string{})
}

// Append the supplied column name and row value into our struct.
func (r *Results) Append(columnName string, rowValue interface{}) {
	// Are we storing column names?
	if r.StoreHeaders {
		r.Rows[0] = append(r.Rows[0], columnName)
	}

	// Append our value to the last row
	r.Rows[len(r.Rows)-1] = append(r.Rows[len(r.Rows)-1], fmt.Sprintf("%v", rowValue))
}

// Save the generated results to the supplied file
func (r *Results) Save(am ActivityMonitor) {
	// If we don't have a Writer, then get out now...
	if r.Writer == nil {
		panic("Caller did not supply a valid Writer")
	}

	// Indicate activity
	am.StartAction("Writing to file")

	// Get the CSV Writer
	writer := csv.NewWriter(r.Writer)

	// Write all of the contents at once
	err := writer.WriteAll(r.Rows)

	// Check for Error
	am.CheckError(err)

	// Indicate success
	am.EndAction("OK")
}
