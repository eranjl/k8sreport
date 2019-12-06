package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

func createCSVReport(reportName string) {
	csvFile, err := os.Create(reportName + ".csv")

	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
		log.Println("File cannot be created for report")
		os.Exit(2)
	}

	var recs [][]string

	// write headers to the report
	csvWriter := csv.NewWriter(csvFile)

	header := []string{
		"Cluster Name",
		"Namespace Name",
		"Workload Kind",
		"Workload Name",
		"Container Name",
		"Attribute Name",
		"Attribute Value",
		"Attribute Defaulted",
		"Attribute Set",
		"Attribute Flagged",
	}

	recs = append(recs, header)

	// preparing data for writing to csv
	for _, r := range rawData {
		for _, a := range r.Attributes {
			attr := []string{
				r.Cluster,
				r.Ns,
				r.Kind,
				r.KindName,
				a.Container,
				a.Name,
				fmt.Sprintf("%v", a.Value),
				strconv.FormatBool(a.Default),
				strconv.FormatBool(a.Set),
				strconv.FormatBool(a.Flagged),
			}

			//			fmt.Printf("Type: %T, Value: %v\n", a.Value, a.Value)

			recs = append(recs, attr)
		}
	}

	for _, r := range recs {
		csvWriter.Write(r)
	}
	csvWriter.Flush()
}
