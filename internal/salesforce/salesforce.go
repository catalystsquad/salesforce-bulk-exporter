package salesforce

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	sfutils "github.com/catalystsquad/salesforce-utils/pkg"
	"github.com/joomcode/errorx"
)

var componentTypes = []string{"address", "location"}
var sfClient *sfutils.SalesforceUtils

func InitSFClient(url, apiVersion, clientID, clientSecret, username, password, grantType string) (err error) {
	sfClient, err = sfutils.NewSalesforceUtils(true, sfutils.Config{
		BaseUrl:      url,
		ApiVersion:   apiVersion,
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		GrantType:    grantType,
	})
	if err != nil {
		err = errorx.Decorate(err, "failed to create new salesforce utils")
	}
	return
}

func GenerateQueryWithAllFields(object string) (string, error) {
	resp, err := sfClient.DescribeObject(object)
	if err != nil {
		err = errorx.Decorate(err, "failed to describe object")
		return "", err
	}
	var fieldsBuilder strings.Builder
	for i, field := range resp.Fields {
		// don't query for component types, because they aren't supported in bulk queries
		if !isComponentType(field.Type) {
			if i != 0 {
				fieldsBuilder.WriteString(", ")
			}
			fieldsBuilder.WriteString(field.Name)
		}
	}
	query := fmt.Sprintf("SELECT %s FROM %s", fieldsBuilder.String(), object)
	return query, nil
}

func isComponentType(sfType string) bool {
	for i := range componentTypes {
		if componentTypes[i] == sfType {
			return true
		}
	}
	return false
}

func SubmitBulkQueryJob(query string) (string, error) {
	resp, err := sfClient.CreateBulkQueryJob(query)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func WaitUntilJobComplete(jobID string, interval time.Duration) error {
	failures := 0
	for {
		// sleep at the beginning, so that the first iteration always waits
		fmt.Printf("job in progress, sleeping for %s...\n", interval.String())
		time.Sleep(interval)
		// get the job
		resp, err := sfClient.GetBulkQueryJob(jobID)
		if err != nil {
			err = errorx.Decorate(err, "failed to get job information")
			return err
		}

		switch resp.State {
		case "JobComplete":
			return nil
		case "InProgress":
			continue
		case "Aborted":
			return errors.New("Salesforce bulk query job aborted")
		case "Failed":
			return errors.New("Salesforce bulk query job failed")
		default:
			// sometimes salesforce responds with a weird job state, allow
			// salesforce to be dumb once, fail if it happens more than once
			fmt.Printf("unexpected job state: \"%s\"\n", resp.State)
			if failures > 0 {
				return fmt.Errorf("Salesforce bulk query job in unexpected state: %s", resp.State)
			}
			failures++
		}
	}
}

func CheckIfJobComplete(jobID string) (complete bool, state string, err error) {
	resp, err := sfClient.GetBulkQueryJob(jobID)
	if err != nil {
		err = errorx.Decorate(err, "failed to get job information")
		return
	}
	state = resp.State
	if state == "JobComplete" {
		complete = true
	}
	return
}

func SaveAllResults(jobID, filenamePrefix, filenameExtension string) ([]string, error) {
	locator := ""
	fileIterator := 0
	var filenames []string
	for {
		resultResp, err := sfClient.GetBulkQueryJobResults(jobID, locator)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s.%d.%s", filenamePrefix, fileIterator, filenameExtension)
		filenames = append(filenames, filename)
		err = writeBytesToFile(filename, resultResp.Body)
		if err != nil {
			return nil, err
		}

		if resultResp.Locator == "" {
			return filenames, nil
		}
		locator = resultResp.Locator
		fileIterator++
	}
}

func writeBytesToFile(filename string, bytes []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return errorx.Decorate(err, fmt.Sprintf("failed create file: %s", filename))
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return errorx.Decorate(err, fmt.Sprintf("failed to write to file: %s", filename))
	}
	return nil
}

func GetAllBulkJobs() ([]sfutils.BulkJobRecord, error) {
	var records []sfutils.BulkJobRecord
	resp, err := sfClient.ListBulkJobs()
	if err != nil {
		return nil, err
	}
	for _, v := range resp.Records {
		records = append(records, v)
	}

	nextRecordsURL := resp.NextRecordsUrl
	for {
		if nextRecordsURL == "" {
			return records, nil
		}

		nextRecordResp, err := sfClient.GetNextRecords(nextRecordsURL)
		if err != nil {
			return nil, err
		}

		for _, v := range nextRecordResp.Records {
			record, ok := v.(sfutils.BulkJobRecord)
			if !ok {
				return nil, errors.New("error asserting bulk job record result to BulkJobRecord type")
			}
			records = append(records, record)
		}
		nextRecordsURL = nextRecordResp.NextRecordsUrl
	}
}
