package cropgraph

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// type of input file
	InputType string
	// number of header lines in the input file
	NumHeader int
	// delimiter of the input file (e.g. tab, comma, space)
	Delimiter string
	// theme of the graph
	Theme string `yaml:",omitempty"`
	// selected names of the columns
	ColumnToGraph map[string]GraphDefinition
}

type GraphDefinition struct {
	// type of the graph
	GraphType string
	// title of the graph
	Title string
	// names of the columns to be plotted
	Columns []string
	// name of Date column
	DateColumn string
	// operation to be applied to the columns
	ColumnView []OperationDefinition `yaml:",omitempty"`
}

type OperationDefinition struct {
	// type of the operation
	// OperationType string
	Operation string
	// name of the column to be operated
	Name string
	// multiply factor
	Multiply float64 `yaml:",omitempty"`
	// names of the columns to affected by the operation
	Columns []string
	// operation parameters
	Parameters map[string]interface{} `yaml:",omitempty"`
}

func ReadConfigFile(configFile string) (*Config, error) {
	// read the config file from yml file
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := Config{
		InputType:     "HermesCSVOut",
		NumHeader:     1,
		Delimiter:     ",",
		Theme:         "calk",
		ColumnToGraph: map[string]GraphDefinition{},
	}
	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// write default config file
func WriteDefaultConfigFile(configFile string) error {
	config := Config{
		InputType: "HermesCSVOut",
		NumHeader: 1,
		Delimiter: ",",
		ColumnToGraph: map[string]GraphDefinition{
			"Graph1": {
				GraphType:  "line",
				Title:      "Graph 1",
				Columns:    []string{"Column1", "Column2"},
				DateColumn: "Date",
				ColumnView: []OperationDefinition{
					{
						Operation:  "sum",
						Name:       "ColumnSum",
						Columns:    []string{"Column2", "Column1"},
						Parameters: map[string]interface{}{},
						Multiply:   1.0,
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// generate path
	err = os.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		return err
	}
	// write the config file
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
