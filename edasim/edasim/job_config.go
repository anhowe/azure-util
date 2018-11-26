package edasim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type JobConfig struct {
	Name string
	BatchName string
	PaddedString string
}

func InitializeJobConfig(name string, batchName string) *JobConfig {
	return &JobConfig{
		Name: name,
		BatchName: batchName,
	}
}

func ReadJobConfigFile(filename string) *JobConfig {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	check(err)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result JobConfig
	json.Unmarshal([]byte(byteValue), &result)
	result.PaddedString = ""
	return &result
}

func (j *JobConfig) WriteFile(filename string, fileSize int) {
	// read once
	data, err := json.Marshal(j)
	check(err)

	// pad and re-martial to match the bytes
	padLength := (KB * 384)-len(data)
	if padLength > 0 {
		j.PaddedString = RandStringRunes(padLength)
		data, err = json.Marshal(j)
		check(err)
	}
	
	// write the file
	f, err := os.Create(filename)
	check(err)
	defer f.Close()
	_, err = f.Write([]byte(data))
	check(err)
}

func (j *JobConfig) GetJobConfigName() string {
	return fmt.Sprintf("%s.job", j.Name)
}

func (j *JobConfig) GetJobConfigCompleteName() string {
	return fmt.Sprintf("%s.complete", j.GetJobConfigName())
}