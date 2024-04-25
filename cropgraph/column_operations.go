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
	newColumnValues = MultiplyColumnValues(newColumnValues, operationDefinition.Multiply)

	return newColumnValues
}

func sumOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement sum operation
	// write a loop to iterate over the columnValues
	// and sum up the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily sums
	newColumnValues := make([]interface{}, len(columnValues[0]))

	for j := 0; j < len(columnValues[0]); j++ {
		sum := 0.0
		for i := 0; i < len(columnValues); i++ {
			sum = sum + AsFloat(columnValues[i][j])
		}
		newColumnValues[j] = sum
	}
	return newColumnValues
}

func diffOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement diff operation
	// write a loop to iterate over the columnValues
	// and calculate the difference between the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily differences between the values of each column
	// e.g. columnValues[0][0] - columnValues[1][0] = newColumnValues[0]
	newColumnValues := make([]interface{}, len(columnValues[0])) // create a new slice of columnValues []interfaceParseFloat()
	for j := 0; j < len(columnValues[0]); j++ {
		newColumnValues[j] = AsFloat(columnValues[0][j])
		for i := 1; i < len(columnValues); i++ {
			newColumnValues[j] = newColumnValues[j].(float64) - AsFloat(columnValues[i][j])
		}
	}

	return newColumnValues
}

func avgOperation(columnValues [][]interface{}) []interface{} {
	// TODO: Implement avg operation
	// write a loop to iterate over the columnValues
	// and calculate the average of the values of each days entry into a new slice newColumnValues
	// as a result, you should have one new slice of daily averages between the values of each column
	// the formula for the average is the sum of all values divided by the number of values
	newColumnValues := make([]interface{}, len(columnValues[0])) // create a new slice of columnValues []interface{}
	numColumnValues := float64(len(columnValues))

	for j := 0; j < len(columnValues[0]); j++ {
		sum := 0.0
		for i := 0; i < len(columnValues); i++ {
			sum = sum + AsFloat(columnValues[i][j])
		}
		newColumnValues[j] = sum / numColumnValues
	}

	return newColumnValues
}

func dailyDifferenceOperation(columnValues []interface{}) []interface{} {
	// TODO: Implement daily difference operation
	// please note that the columnValues are now a slice of interface{} instead of a slice of []interface{}
	// write a loop to iterate over the columnValues
	// and calculate the difference between two consecutive days into a new slice newColumnValues
	// as a result, you should have one new slice of daily differences between the values of each column
	// e.g. columnValues[1] - columnValues[0] = newColumnValues[0]
	// the first value of the newColumnValues should be 0, as there is no previous value to calculate the difference from
	newColumnValues := make([]interface{}, len(columnValues)) // create a new slice of columnValues []interface{}

	newColumnValues[0] = 0.0
	for i := 1; i < len(columnValues); i++ {
		newColumnValues[i] = AsFloat(columnValues[i]) - AsFloat(columnValues[i-1])
	}

	return newColumnValues
}

func AsFloat(value interface{}) float64 {
	// check if the value is a string and convert it to a float64
	// if the value is already a float64, return it
	// if the value is not a string or a float64, panic
	if _, ok := value.(float64); ok {
		return value.(float64)
	}

	if _, ok := value.(string); ok {
		parsedValue, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			panic(err)
		}
		return parsedValue
	}
	panic("value is not a float64 or a string")
}

func MultiplyColumnValues(columnValues []interface{}, factor float64) []interface{} {
	// check if factor is 0 and return the columnValues as they are
	// if factor is not 0, multiply each value in the columnValues with the factor
	if factor == 0 || factor == 1 {
		return columnValues
	}

	for i := 0; i < len(columnValues); i++ {
		columnValues[i] = AsFloat(columnValues[i]) * factor
	}
	return columnValues
}
