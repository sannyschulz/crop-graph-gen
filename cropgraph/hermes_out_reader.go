package cropgraph

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
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
func HermesCsvToGraph(inputFile string, config *Config, outputFile string) error {

	rowData, mappingColumnToIndex, err := ReadFileData(inputFile, *config)
	if err != nil {
		return err
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

func BatchFileToGraph(batchFile string, configFile string) error {
	// read config file
	config, err := ReadConfigFile(configFile)
	if err != nil {
		return err
	}

	// Read the batch file
	file, err := os.Open(batchFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// go csv reader
	// https://golang.org/pkg/encoding/csv/
	reader := csv.NewReader(file)
	reader.Comma = rune(config.Delimiter[0])
	if config.MultiFiles {
		outToInputFile := map[string][]string{}
		currentOUtputFile := ""
		for {
			// read the row
			row, err := reader.Read()
			if err != nil {
				break
			}
			// read the input and output file from the batch file
			if len(row) < 1 {
				return fmt.Errorf("batch file must have at least an input file")
			}

			inputFile := row[0]
			outputfile := currentOUtputFile
			if row[1] != "" {
				outputfile = row[1]
				currentOUtputFile = outputfile
			}
			if outputfile == "" {
				return fmt.Errorf("output file must be given")
			}

			if _, ok := outToInputFile[outputfile]; !ok {
				outToInputFile[outputfile] = []string{}
			}
			outToInputFile[outputfile] = append(outToInputFile[outputfile], inputFile)
		}

		for outputFile, inputFiles := range outToInputFile {
			// make graphs from multiple input files
			err = MultiFileToGraph(inputFiles, config, outputFile)
			if err != nil {
				return err
			}

		}

	} else {
		for {
			// read the row
			row, err := reader.Read()
			if err != nil {
				break
			}
			// read the input and output file from the batch file
			if len(row) < 2 {
				return fmt.Errorf("batch file must have at least two columns")
			}
			inputFile := row[0]
			outputFile := row[1]
			// generate the graph
			err = HermesCsvToGraph(inputFile, config, outputFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func MultiFileToGraph(inputFiles []string, config *Config, outputFile string) error {

	// sorted by the order in the config file
	graphNames := make([]string, 0, len(config.ColumnToGraph))
	for graphName := range config.ColumnToGraph {
		graphNames = append(graphNames, graphName)
	}
	//sort.Strings(graphNames)
	slices.Sort(graphNames)

	// read all input files
	rowDataList := make([]map[string][]interface{}, 0, len(inputFiles))
	mappingColumnToIndexList := make([]map[string]int, 0, len(inputFiles))
	for _, inputFile := range inputFiles {
		rowData, mappingColumnToIndex, err := ReadFileData(inputFile, *config)
		if err != nil {
			return err
		}
		rowDataList = append(rowDataList, rowData)
		mappingColumnToIndexList = append(mappingColumnToIndexList, mappingColumnToIndex)
	}

	// genrate a web page for the graph
	page := MakePage()
	// for each graph in the config file generate the graph

	type klineEntry struct {
		open  float64
		close float64
		low   float64
		high  float64
	}

	for _, graphName := range graphNames {
		graph := config.ColumnToGraph[graphName]
		if graph.GraphType == "kline" {
			// merge data for kline graph
			// requires a list of (date, open, close, low, high) values
			// number of columns must be 1 + date column
			dates := []string{}
			if graph.DateColumn != "" {
				if col, ok := rowDataList[0][graph.DateColumn]; ok {
					for _, date := range col {
						dates = append(dates, date.(string))
					}
				}
			}
			if (len(dates) == 0 && len(graph.Columns) != 1) || (len(dates) > 0 && len(graph.Columns) != 2) {
				return fmt.Errorf("kline graph requires one data column")
			}
			// number of entries
			numEntries := len(dates)

			// get data column
			columnName := graph.Columns[0]
			for _, colname := range graph.Columns {
				if colname != graph.DateColumn {
					columnName = colname
					break
				}
			}
			byDate := make([][]float64, numEntries)
			for i := range byDate {
				byDate[i] = make([]float64, len(rowDataList))
			}

			for i, rowData := range rowDataList {
				if _, ok := mappingColumnToIndexList[i][columnName]; !ok {
					return fmt.Errorf("column %s not found in the input file", columnName)
				}
				for j, value := range rowData[columnName] {
					byDate[j][i] = AsFloat(value)
				}
			}
			// calculate standard deviation and average
			averages := make([]float64, numEntries)
			low := make([]float64, numEntries)
			// init low with max value
			for i := range low {
				low[i] = math.MaxFloat64
			}
			high := make([]float64, numEntries)
			numFiles := float64(len(rowDataList))
			for i := range byDate {
				if i == 309 {
					fmt.Println("here")
				}
				for _, value := range byDate[i] {
					if value < low[i] {
						low[i] = value
					}
					if value > high[i] {
						high[i] = value
					}
					averages[i] += value
				}
				averages[i] /= numFiles
			}
			// calculate standard deviation
			stdDevs := make([]float64, numEntries)
			for i := range byDate {
				for _, value := range byDate[i] {
					stdDevs[i] += (value - averages[i]) * (value - averages[i])
				}
				stdDevs[i] = stdDevs[i] / numFiles
				stdDevs[i] = math.Sqrt(stdDevs[i])
			}

			klineEntries := make([]klineEntry, 0, numEntries)
			for i := range averages {
				klineEntries = append(klineEntries, klineEntry{
					open:  averages[i] - stdDevs[i],
					close: averages[i] + stdDevs[i],
					low:   low[i],
					high:  high[i]})
			}
			kline := makeKline(graphStyle{title: graph.Title, theme: config.Theme})
			klineEntriesOpt := make([]opts.KlineData, 0, len(klineEntries))
			for i := 0; i < len(klineEntries); i++ {
				// entry to [4]float
				val := []float64{klineEntries[i].open, klineEntries[i].close, klineEntries[i].low, klineEntries[i].high}
				klineEntriesOpt = append(klineEntriesOpt, opts.KlineData{Value: val})
			}

			kline.SetXAxis(dates).AddSeries("kline", klineEntriesOpt)
			page = page.AddCharts(kline)

		}

	}
	// save the page to the output file
	err := SavePage(page, outputFile)
	return err
}

func ReadFileData(inputFile string, config Config) (map[string][]interface{}, map[string]int, error) {
	// Read the hermes simulation output file
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
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

	return rowData, mappingColumnToIndex, nil
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
		fmt.Println("Warnung", graphType.GraphType, "is kind of buggy. It will temper with the theme and rendering.")
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
	bar3d := makebar3DShading(graphStyle)

	if dates == nil {
		dates = make([]string, len(keys))
		for i, key := range keys {
			dates[i] = strconv.Itoa(key)
		}
	}

	bar3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Data: dates}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Data: columns}),
	)
	bar3d.AddSeries("bar3d", generateItemBar3D(dates, values, columns), charts.WithBar3DChartOpts(opts.Bar3DChart{Shading: "lambert"}))
	return bar3d
}
func generateItemBar3D(dates []string, values [][]interface{}, columns []string) []opts.Chart3DData {

	items := make([]opts.Chart3DData, 0, len(dates)*len(columns))

	for i := range columns {
		for j := range dates {
			items = append(items, opts.Chart3DData{
				Value: []interface{}{j, i, values[i][j]}, // {x, y, z}
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

func makebar3DShading(graphStyle graphStyle) *charts.Bar3D {
	bar3d := charts.NewBar3D()
	bar3DRangeColor := []string{
		"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
		"#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
	}

	bar3d.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: graphStyle.title,
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: true,
			Max:        30,
			Range:      []float32{0, 30},
			InRange:    &opts.VisualMapInRange{Color: bar3DRangeColor},
		}),
		charts.WithGrid3DOpts(opts.Grid3D{
			BoxWidth: 200,
			BoxDepth: 80,
		}),
	)
	return bar3d
}

func makeKline(graphStyle graphStyle) *charts.Kline {
	kline := charts.NewKLine()
	kline.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: graphStyle.theme,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: graphStyle.title,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
	)
	return kline
}
