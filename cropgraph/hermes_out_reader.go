package cropgraph

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

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

	// sorted by the order in the config file
	graphNames := make([]string, 0, len(config.ColumnToGraph))
	for graphName := range config.ColumnToGraph {
		graphNames = append(graphNames, graphName)
	}
	//sort.Strings(graphNames)
	slices.Sort(graphNames)

	for _, graphName := range graphNames {
		graph := config.ColumnToGraph[graphName]
		// get all lists of values for the graph
		values := make([][]interface{}, len(graph.Columns))
		for i, column := range graph.Columns {
			if _, ok := mappingColumnToIndex[column]; !ok {
				return fmt.Errorf("column %s not found in the input file", column)
			}
			values[i] = rowData[column]
		}
		// add the graph to the page
		page = GenerateGraph(page, graph, config.Theme, values)

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
func GenerateGraph(page *components.Page, graphType GraphDefinition, theme string, values [][]interface{}) *components.Page {
	outPage := page
	// generate the graph
	if len(values) == 0 {
		return outPage
	}
	// extract keys from the first column
	keys := extractKeys(values[0])
	var dates []string = nil
	var columns []string
	var combinedColumnValues [][]interface{}
	if graphType.ColumnView != nil {
		combinedColumnValues = make([][]interface{}, 0, len(graphType.ColumnView))
		columns = make([]string, 0, len(graphType.ColumnView))
		// apply operations to the columns
		for _, operationDefinition := range graphType.ColumnView {
			// get the column values for the operation
			columnValues := make([][]interface{}, len(operationDefinition.Columns))
			for i, column := range operationDefinition.Columns {
				for j, col := range graphType.Columns {
					if col == column {
						columnValues[i] = values[j]
						break
					}
				}
			}
			// apply the operation to the column values
			newColumnValues := HandleColumnViewOperation(operationDefinition, columnValues)
			combinedColumnValues = append(combinedColumnValues, newColumnValues)
			columns = append(columns, operationDefinition.Name)
		}
	} else {
		columns = graphType.Columns
		combinedColumnValues = values
		if graphType.DateColumn != "" {
			for i, column := range columns {
				if column == graphType.DateColumn {
					// convert dates to string
					dates = make([]string, len(combinedColumnValues[i]))
					for j, date := range combinedColumnValues[i] {
						dates[j] = date.(string)
					}
					// remove date column from columns
					combinedColumnValues = append(combinedColumnValues[:i], combinedColumnValues[i+1:]...)
					columns = append(columns[:i], columns[i+1:]...)
					break
				}
			}
		}
	}
	// graph style
	graphStyle := graphStyle{
		title: graphType.Title,
		theme: theme,
	}

	switch graphType.GraphType {
	case "line":
		outPage = page.AddCharts(
			lineMultiData(keys, dates, graphStyle, columns, combinedColumnValues),
		)
	case "ThemeRiver":
		outPage = page.AddCharts(
			themeRiverMultiData(keys, dates, graphStyle, columns, combinedColumnValues),
		)
	case "bar3d":
		outPage = page.AddCharts(
			Bar3D(keys, dates, graphStyle, columns, combinedColumnValues),
		)
	default:
		fmt.Println("Graph type ", graphType.GraphType, " not supported")
	}

	return outPage
}

type graphStyle struct {
	title string
	theme string
}

func extractKeys(valueList []interface{}) []int {
	keys := make([]int, 0, len(valueList))
	for i := range valueList {
		keys = append(keys, i)
	}
	return keys
}

func lineMultiData(keys []int, dates []string, graphStyle graphStyle, columns []string, values [][]interface{}) *charts.Line {

	line := makeMultiLine(graphStyle)

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

func themeRiverMultiData(keys []int, dates []string, graphStyle graphStyle, columns []string, values [][]interface{}) *charts.ThemeRiver {
	themeRiver := makeThemeRiver(graphStyle)
	if dates == nil {
		dates = make([]string, len(keys))
		for i, key := range keys {
			dates[i] = strconv.Itoa(key)
		}
	}

	themeRiver.AddSeries("themeRiver", generateItemTripple(dates, values, columns))
	return themeRiver
}

func Bar3D(keys []int, dates []string, graphStyle graphStyle, columns []string, values [][]interface{}) *charts.Bar3D {
	bar3d := makeBar3D(graphStyle)
	if dates == nil {
		dates = make([]string, len(keys))
		for i, key := range keys {
			dates[i] = strconv.Itoa(key)
		}
	}
	bar3d.AddSeries("bar3d", generateItemBar3D(dates, values, columns))
	return bar3d
}
func generateItemBar3D(dates []string, values [][]interface{}, columns []string) []opts.Chart3DData {

	items := make([]opts.Chart3DData, 0, len(dates)*len(columns))

	for i, column := range columns {
		for j, date := range dates {
			// {"2015/11/28", 10, "DD"},
			// convert current date string to new date format
			dateTime, _ := time.Parse("02.01.2006", date)
			dateFormated := dateTime.Format("2006/01/02")

			valueAsFloat := AsFloat(values[i][j])
			items = append(items, opts.Chart3DData{
				Name:  column,
				Value: []interface{}{dateFormated, valueAsFloat},
			})
		}
	}
	return items

}

func generateItemTripple(dates []string, values [][]interface{}, columns []string) []opts.ThemeRiverData {

	items := make([]opts.ThemeRiverData, 0, len(dates)*len(columns))

	for i, column := range columns {
		for j, date := range dates {
			// {"2015/11/28", 10, "DD"},
			// convert current date string to new date format
			dateTime, _ := time.Parse("02.01.2006", date)
			dateFormated := dateTime.Format("2006/01/02")

			valueAsFloat := AsFloat(values[i][j])
			items = append(items, opts.ThemeRiverData{
				Date:  dateFormated,
				Value: valueAsFloat,
				Name:  column,
			})
		}
	}
	return items
}

func generateItems(keys []int, values []interface{}) []opts.LineData {

	items := make([]opts.LineData, 0, len(keys))

	for _, key := range keys {
		val := values[key]
		items = append(items, opts.LineData{Value: val})
	}
	return items
}

func makeThemeRiver(graphStyle graphStyle) *charts.ThemeRiver {
	themeRiver := charts.NewThemeRiver()
	themeRiver.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: graphStyle.title,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: graphStyle.theme,
		}),
		charts.WithLegendOpts(opts.Legend{Show: true,
			Right: "15%",
			Top:   "5%",
			Align: "left",
		}),
		charts.WithSingleAxisOpts(opts.SingleAxis{
			Type:   "time",
			Bottom: "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "inside",
			Start: 0,
			End:   100,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 50,
			End:   100,
		}),
	)
	return themeRiver
}

func makeMultiLine(graphStyle graphStyle) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: graphStyle.title,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: graphStyle.theme,
		}),
		charts.WithLegendOpts(opts.Legend{Show: true,
			Right: "15%",
			Top:   "5%",
			Align: "left",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)
	return line
}

func makeBar3D(graphStyle graphStyle) *charts.Bar3D {
	bar3d := charts.NewBar3D()
	bar3d.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: graphStyle.title,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: graphStyle.theme,
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: true,
			Max:        100,
			InRange: &opts.VisualMapInRange{
				Color: []string{"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8", "#ffffbf", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"},
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:      true,
			Trigger:   "item",
			Formatter: "{a} <br/>{b} <br/>{c}",
		}),
	)
	return bar3d
}
