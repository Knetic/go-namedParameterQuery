package namedParameterQuery

import (
	"testing"
)

/*
	Represents a single test of query parsing.
	Given an [Input] query, if the actual result of parsing
	does not match the [Expected] string, the test fails
*/
type QueryParsingTest struct {
	Name string
	Input string
	Expected string
	ExpectedParameters int
}

/*
	Represents a single test of parameter parsing.
	Given the [Query] and a set of [Parameters], if the actual parameter output
	from GetParsedParameters() matches the given [ExpectedParameters].
	These tests specifically check type of output parameters, too.
*/
type ParameterParsingTest struct {

	Name string
	Query string
	Parameters []TestQueryParameter
	ExpectedParameters []interface{}
}

type TestQueryParameter struct {
	Name string
	Value interface{}
}

func TestQueryParsing(test *testing.T) {

	var query *NamedParameterQuery

	// Each of these represents a single test.
	queryParsingTests := []QueryParsingTest {
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = 1",
			Expected: "SELECT * FROM table WHERE col1 = 1",
			Name: "NoParameter",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :name",
			Expected: "SELECT * FROM table WHERE col1 = ?",
			ExpectedParameters: 1,
			Name: "SingleParameter",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :name AND col2 = :occupation",
			Expected: "SELECT * FROM table WHERE col1 = ? AND col2 = ?",
			ExpectedParameters: 2,
			Name: "TwoParameters",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :name AND col2 = :occupation",
			Expected: "SELECT * FROM table WHERE col1 = ? AND col2 = ?",
			ExpectedParameters: 2,
			Name: "OneParameterMultipleTimes",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 IN (:something, :else)",
			Expected: "SELECT * FROM table WHERE col1 IN (?, ?)",
			ExpectedParameters: 2,
			Name: "ParametersInParenthesis",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = ':literal' AND col2 LIKE ':literal'",
			Expected: "SELECT * FROM table WHERE col1 = ':literal' AND col2 LIKE ':literal'",
			Name: "ParametersInQuotes",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = ':literal' AND col2 = :literal AND col3 LIKE ':literal'",
			Expected: "SELECT * FROM table WHERE col1 = ':literal' AND col2 = ? AND col3 LIKE ':literal'",
			ExpectedParameters: 1,
			Name: "ParametersInQuotes2",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :foo AND col2 IN (SELECT id FROM tabl2 WHERE col10 = :bar)",
			Expected: "SELECT * FROM table WHERE col1 = ? AND col2 IN (SELECT id FROM tabl2 WHERE col10 = ?)",
			ExpectedParameters: 2,
			Name: "ParametersInSubclause",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :1234567890 AND col2 = :0987654321",
			Expected: "SELECT * FROM table WHERE col1 = ? AND col2 = ?",
			ExpectedParameters: 2,
			Name: "NumericParameters",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			Expected: "SELECT * FROM table WHERE col1 = ?",
			ExpectedParameters: 1,
			Name: "CapsParameters",
		},
		QueryParsingTest {
			Input: "SELECT * FROM table WHERE col1 = :abc123ABC098",
			Expected: "SELECT * FROM table WHERE col1 = ?",
			ExpectedParameters: 1,
			Name: "AltcapsParameters",
		},
	}

	// Run each test.
	for _, parsingTest := range queryParsingTests {

		query = NewNamedParameterQuery(parsingTest.Input)

		// test query texts
		if(query.GetParsedQuery() != parsingTest.Expected) {
			test.Log("Test '", parsingTest.Name, "': Expected query text did not match actual parsed output")
			test.Log("Actual: ", query.GetParsedQuery())
			test.Fail()
		}

		// test parameters
		if(len(query.GetParsedParameters()) != parsingTest.ExpectedParameters) {
			test.Log("Test '", parsingTest.Name, "': Expected parameters did not match actual parsed parameter count")
			test.Fail()
		}
	}
}

/*
	Tests to ensure that setting parameter values turns out correct when using GetParsedParameters().
	These tests ensure correct positioning and type.
*/
func TestParameterReplacement(test *testing.T) {

	var query *NamedParameterQuery

	queryVariableTests := []ParameterParsingTest {
		ParameterParsingTest {

			Name: "SingleStringParameter",
			Query: "SELECT * FROM table WHERE col1 = :foo",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "foo",
					Value: "bar",
				},
			},
			ExpectedParameters: []interface{} {
				"bar",
			},
		},
		ParameterParsingTest {

			Name: "TwoStringParameter",
			Query: "SELECT * FROM table WHERE col1 = :foo AND col2 = :foo2",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "foo",
					Value: "bar",
				},
				TestQueryParameter {
					Name: "foo2",
					Value: "bart",
				},
			},
			ExpectedParameters: []interface{} {
				"bar", "bart",
			},
		},
		ParameterParsingTest {

			Name: "TwiceOccurringParameter",
			Query: "SELECT * FROM table WHERE col1 = :foo AND col2 = :foo",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "foo",
					Value: "bar",
				},
			},
			ExpectedParameters: []interface{} {
				"bar", "bar",
			},
		},
		ParameterParsingTest {

			Name: "ParameterTyping",
			Query: "SELECT * FROM table WHERE col1 = :str AND col2 = :int AND col3 = :pi",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "str",
					Value: "foo",
				},
				TestQueryParameter {
					Name: "int",
					Value: 1,
				},
				TestQueryParameter {
					Name: "pi",
					Value: 3.14,
				},
			},
			ExpectedParameters: []interface{} {
				"foo", 1, 3.14,
			},
		},
		ParameterParsingTest {

			Name: "ParameterOrdering",
			Query: "SELECT * FROM table WHERE col1 = :foo AND col2 = :bar AND col3 = :foo AND col4 = :foo AND col5 = :bar",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "foo",
					Value: "something",
				},
				TestQueryParameter {
					Name: "bar",
					Value: "else",
				},
			},
			ExpectedParameters: []interface{} {
				"something", "else", "something", "something", "else",
			},
		},
		ParameterParsingTest {

			Name: "ParameterCaseSensitivity",
			Query: "SELECT * FROM table WHERE col1 = :foo AND col2 = :FOO",
			Parameters: []TestQueryParameter {
				TestQueryParameter {
					Name: "foo",
					Value: "baz",
				},
				TestQueryParameter {
					Name: "FOO",
					Value: "quux",
				},
			},
			ExpectedParameters: []interface{} {
				"baz", "quux",
			},
		},
	}

	// run variable tests.
	for _, variableTest := range queryVariableTests {

		// parse query and set values.
		query = NewNamedParameterQuery(variableTest.Query)
		for _, queryVariable := range variableTest.Parameters {
			query.SetValue(queryVariable.Name, queryVariable.Value)
		}

		// Test outputs
		for index, queryVariable := range query.GetParsedParameters() {

			if(queryVariable != variableTest.ExpectedParameters[index]) {
				test.Log("Test '", variableTest.Name, "' did not produce the expected parameter output. Actual: '", queryVariable, "', Expected: '", variableTest.ExpectedParameters[index], "'")
				test.Fail()
			}
		}
	}
}
