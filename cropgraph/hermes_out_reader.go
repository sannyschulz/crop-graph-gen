package cropgraph

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// HermesCsvToGraph reads the hermes simulation output file and generates graphs as defined in the config file
func HermesCsvToGraph(inputFile string, configFile string, outputFile string) error {

	// read config file
	config, err := ReadConfigFile(configFile)
	if err != nil {
		return err
	}

	// Read the hermes simulation output file
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// go csv reader
	// https://golang.org/pkg/encoding/csv/
	reader := csv.NewReader(file)
	reader.Comma = rune(config.Delimiter[0])

	// list all required columns from the config file
	configColumns := map[string]int{}
	for _, graph := range config.ColumnToGraph {
		for _, column := range graph.Columns {
			configColumns[column] = -1
		}
	}
	// map column name to index in the csv file
	mappingColumnToIndex := map[string]int{}

	// read number of header, as defined in the config file
	for i := 0; i < config.NumHeader; i++ {

		col, err := reader.Read()
		if err != nil {
			return err
		}
		// for the first header line
		if i == 0 {
			for colIndex, colName := range col {
				// if column is listed in the config file
				if _, ok := configColumns[colName]; ok {
					// store the index of the column
					mappingColumnToIndex[colName] = colIndex
				}
			}
		}
	}

	rowData := map[string][]interface{}{}
	// read data from rows after the header
	for {
		// read the row
		row, err := reader.Read()
		if err != nil {
			break
		}
		// for each column in the row check if it is listed in the config file
		// if yes, store the value to later generate a graph
		for colName, colIndex := range mappingColumnToIndex {
			// read the value of the selected column
			// and store it in the crop graph
			rowData[colName] = append(rowData[colName], row[colIndex])
		}
	}
	// genrate a web page for the graph
	page := MakePage()
	// for each graph in the config file generate the graph
	for _, graph := range config.ColumnToGraph {
		// get all lists of values for the graph
		values := make([][]interface{}, len(graph.Columns))
		for i, column := range graph.Columns {
			if _, ok := mappingColumnToIndex[column]; !ok {
				return fmt.Errorf("column %s not found in the input file", column)
			}
			values[i] = rowData[column]
		}
		// add the graph to the page
		page = GenerateGraph(page, graph, values)

	}
	// save the page to the output file
	err = SavePage(page, outputFile)
	return err
}

// MakePage creates a new web page
func MakePage() *components.Page {
	// create a new web page
	page := components.NewPage()
	return page
}

// SavePage saves the web page to a file
func SavePage(page *components.Page, outfile string) error {
	// check if output file location exists
	// if not create it
	outpath := filepath.Dir(outfile)
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		err := os.MkdirAll(outpath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	// save the page to the output file
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	err = page.Render(io.MultiWriter(f))
	return err
}

// graph generation
func GenerateGraph(page *components.Page, graphType GraphDefinition, values [][]interface{}) *components.Page {
	outPage := page
	// generate the graph
	if len(values) == 0 {
		return outPage
	}
	// extract keys from the first column
	keys := extractKeys(values[0])
	var dates []string = nil
	columns := graphType.Columns
	if graphType.DateColumn != "" {
		for i, column := range columns {
			if column == graphType.DateColumn {
				// convert dates to string
				dates = make([]string, len(values[i]))
				for j, date := range values[i] {
					dates[j] = date.(string)
				}
				// remove date column from columns
				values = append(values[:i], values[i+1:]...)
				columns = append(columns[:i], columns[i+1:]...)
				break
			}
		}
	}

	outPage = page.AddCharts(
		lineMultiData(keys, dates, graphType.Title, columns, values),
	)
	return outPage
}

func extractKeys(valueList []interface{}) []int {
	keys := make([]int, 0, len(valueList))
	for i := range valueList {
		keys = append(keys, i)
	}
	return keys
}

func lineMultiData(keys []int, dates []string, graphName string, columns []string, values [][]interface{}) *charts.Line {

	line := makeMultiLine(graphName)

	if dates == nil {
		dates = make([]string, len(keys))
		for i, key := range keys {
			dates[i] = strconv.Itoa(key)
		}
	}
	graph := line.SetXAxis(dates)
	for i, column := range columns {
		graph = graph.AddSeries(column, generateItems(keys, values[i]))
	}
	return line
}

func generateItems(keys []int, values []interface{}) []opts.LineData {

	items := make([]opts.LineData, 0, len(keys))

	for _, key := range keys {
		val := values[key]
		items = append(items, opts.LineData{Value: val})
	}
	return items
}

func makeMultiLine(title string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "shine",
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)
	return line
}
