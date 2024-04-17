package cropgraph

import "strconv"

func HandleColumnViewOperation(operationDefinition OperationDefinition, columnValues [][]interface{}) []interface{} {

	// TODO: throw an error if the columnValues are empty

	var newColumnValues []interface{}
	switch operationDefinition.Operation {
	case "sum":
		newColumnValues = sumOperation(columnValues)
	case "diff":
		newColumnValues = diffOperation(columnValues)
	case "avg":
		newColumnValues = avgOperation(columnValues)
	case "dailydifference":
		newColumnValues = dailyDifferenceOperation(columnValues[0])
	case "none":
		newColumnValues = columnValues[0]
	}
	// TODO: Implement multiplication for each element in the result column

	return newColumnValues
}

func sumOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement sum operation
	// write a loop to iterate over the columnValues
	// and sum up the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily sums
	newColumnValues := make([]interface{}, len(columnValues[0]))

	return newColumnValues
}

func diffOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement diff operation
	// write a loop to iterate over the columnValues
	// and calculate the difference between the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily differences between the values of each column
	// e.g. columnValues[0][0] - columnValues[1][0] = newColumnValues[0]
	newColumnValues := make([]interface{}, len(columnValues[0])) // create a new slice of columnValues []interfaceParseFloat()

	return newColumnValues
}

func avgOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement avg operation
	// write a loop to iterate over the columnValues
	// and calculate the average of the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily averages between the values of each column
	// the formula for the average is the sum of all values divided by the number of values
	newColumnValues := make([]interface{}, len(columnValues[0])) // create a new slice of columnValues []interface{}

	return newColumnValues
}

func dailyDifferenceOperation(columnValues []interface{}) []interface{} {
	// TODO: Implement daily difference operation
	// please note that the columnValues are now a slice of interface{} instead of a slice of []interface{}
	// write a loop to iterate over the columnValues
	// and calculate the difference between two consecutive days into a new slice newColumnValues
	// as a result, you should have one new slice of daily differences between the values of each column
	// e.g. columnValues[0][0] - columnValues[0][1] = newColumnValues[0]
	// the first value of the newColumnValues should be 0, as there is no previous value to calculate the difference from
	newColumnValues := make([]interface{}, len(columnValues)) // create a new slice of columnValues []interface{}

	return newColumnValues
}

func AsFloat(value interface{}) float64 {
	parsedValue, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		panic(err)
	}
	return parsedValue
}
