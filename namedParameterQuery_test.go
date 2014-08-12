package namedParameterQuery

import (
	"testing"
)

/*
	Represents a single test of query parsing.
	Given an [Input] query, if the actual result of parsing
	does not match the [Expected] string, the test fails with the given [FailureMessage]
*/
type QueryParsingTest struct {
	Name string
	Input string
	Expected string
	ExpectedParameters int
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
