package namedparameter

import (
	"errors"
)

// convertArgsToMap converts a list of arguments into a key value map, where the keys are the parameter names.
// This method expects the parameter arguments to either being just a map[string]any or to come in pairs, in which case
// it processes them as key, value pairs, keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func convertArgsToMap(args []any) (map[string]any, error) {
	if len(args) == 0 {
		return nil, nil
	}
	if len(args) == 1 {
		if params, ok := args[0].(map[string]any); ok {
			return params, nil
		}
	}
	if len(args)%2 != 0 {
		return nil, errors.New("number of arguments passed to parameterized query is not correct, expected an even number of arguments")
	}
	params := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			return nil, errors.New("parameter representing a key needs to be a string")
		} else {
			params[keyStr] = val
		}
	}
	return params, nil
}

// parse automates all the process of processing the query and arguments in one place, in order to avoid
// doing this in every other method.
func parse(query string, args []any) (parsedQuery string, params []any, err error) {
	var mappedArgs map[string]any

	mappedArgs, err = convertArgsToMap(args)
	if err != nil {
		return
	}
	paramQuery := NewQuery(query)
	paramQuery.SetValuesFromMap(mappedArgs)

	parsedQuery = paramQuery.GetParsedQuery()
	params = paramQuery.GetParsedParameters()

	return
}
