package bioparser

import (
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//LoadFile loads CSV file into an spec object
func LoadFile(fp string) (Specification, error) {
	var specification Specification

	log.Debug("Loading file ", fp)

	if strings.HasSuffix(fp, ".csv") {
		f, err := os.Open(fp)
		if err != nil {
			log.Panic(err)
		}
		defer f.Close()

		specification, err = ParseSpecificationCSV(csv.NewReader(f))
		if err != nil {
			return specification, err
		}

		return specification, nil
	}

	log.Debug("The file name doesn't have .csv extension")

	return specification, errors.New("Error that path doesn't looks like a CSV file")
}

//LoadURL loads CSV from url into an spec object
func LoadURL(u string) (Specification, error) {
	var specification Specification

	res, err := http.Get(u)
	if err != nil {
		return specification, err
	}
	defer res.Body.Close()

	specification, err = ParseSpecificationCSV(csv.NewReader(res.Body))
	if err != nil {
		return specification, err
	}

	return specification, nil
}

/*
ParseSpecificationCSV reads the CSV file
	parses it into an Specification struct. The CSV second row has the Specification Info,
	and from the fourth row the mapping.
*/
func ParseSpecificationCSV(r *csv.Reader) (Specification, error) {
	now := time.Now()

	var specification Specification
	log.Debug("START - CSV parsing")
	r.FieldsPerRecord = 10

	csv, err := r.ReadAll()
	if err != nil {
		log.Error("An error has occurred trying to parsing the CVS reader ", err)
		return specification, err
	}

	if len(csv) <= 1 {
		log.Error("Empty content, please check the CSV")
		return specification, err //TODO err is nil here it this correct?
	}
	log.Debug("CSV lenght ", len(csv))

	specParams := make([]SpecificationParam, 0)
	var specInfo SpecificationInfo

	for i, row := range csv {
		log.WithFields(log.Fields{
			"row":   row,
			"index": i,
		}).Debug("Parsing row ", i)
		if i == 1 {
			log.Debug("Row 1 Specifications Info")
			specInfo = SpecificationInfo{
				Title:        strings.TrimSpace(row[0]),
				Subtitle:     strings.TrimSpace(row[1]),
				Description:  strings.TrimSpace(row[2]),
				Version:      strings.TrimSpace(row[3]),
				VersionDate:  now.Format("20060102T150405"),
				OfficialType: strings.TrimSpace(row[4]),
				FullExample:  strings.TrimSpace(row[5]),
			}
		}

		if i <= 3 {
			log.Debug("The row doesn't contain Specification Parameters SKIPPING ")
			continue
		}

		xtypes := extractExpectedTypes(strings.TrimSpace(row[1]))
		log.Debug("Extracted Expected Types", xtypes)
		s := SpecificationParam{
			Property:             strings.TrimSpace(row[0]),
			ExpectedTypes:        xtypes,
			Description:          row[2],
			Type:                 strings.TrimSpace(row[3]),
			TypeURL:              strings.TrimSpace(row[4]),
			BscDescription:       row[5],
			Marginality:          strings.TrimSpace(row[6]),
			Cardinality:          strings.TrimSpace(row[7]),
			ControlledVocabulary: row[8],
			Example:              row[9],
		}
		specParams = append(specParams, s)
	}
	specification = Specification{specInfo, specParams}

	return specification, nil
}

//extractExpectedTypes returns a list of types. The types are in a string
//usually separated by the word "or" or a comma ",".
func extractExpectedTypes(s string) []string {
	re := regexp.MustCompile("(or|,)")
	ss := re.Split(s, -1)

	for i, et := range ss {
		ss[i] = strings.TrimSpace(et)
	}

	return ss
}
