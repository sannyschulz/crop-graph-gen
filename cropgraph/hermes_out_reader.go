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

func SefaultCsvToGraph(inputFile string, configFile string, outputFile string) error {

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

	existingListOfColumns := map[string]int{}

	// read number of header, as defined in the config file
	for i := 0; i < config.NumHeader; i++ {

		col, err := reader.Read()
		if err != nil {
			return err
		}
		if i == 0 {
			for colIndex, colName := range col {
				if _, ok := configColumns[colName]; ok {
					existingListOfColumns[colName] = colIndex
				}
			}
		}
	}

	rowData := map[string][]interface{}{}
	// read data rows
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		for colName, colIndex := range existingListOfColumns {
			// read the value of the selected column
			// and store it in the crop graph
			rowData[colName] = append(rowData[colName], row[colIndex])
		}
	}

	for _, graph := range config.ColumnToGraph {
		// 4. generate the crop graph as defined in the config file
		// list of values for the graph
		values := make([][]interface{}, len(graph.Columns))
		for i, column := range graph.Columns {
			if _, ok := existingListOfColumns[column]; !ok {
				return fmt.Errorf("column %s not found in the input file", column)
			}
			values[i] = rowData[column]
		}

		GenerateGraph(outputFile, graph, values)
	}
	return nil
}

// graph generation
func GenerateGraph(outfile string, graphType GraphDefinition, values [][]interface{}) {
	// generate the graph
	page := components.NewPage()
	if len(values) == 0 {
		return
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

	page.AddCharts(
		lineMultiData(keys, dates, graphType.Title, columns, values),
	)

	// check if output file location exists
	// if not create it
	// get path from outfile
	outpath := filepath.Dir(outfile)
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		os.MkdirAll(outpath, os.ModePerm)
	}

	f, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))

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
