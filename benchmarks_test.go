package namedParameterQuery

import (
  "testing"
  "bytes"
  "fmt"
)

func BenchmarkSimpleParsing(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux"
  for i := 0; i < bench.N; i++ {

    NewNamedParameterQuery(query)
  }
}

func BenchmarkMultiOccurrenceParsing(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux " +
            "AND [something] = :quux " +
            "OR [otherStuff] NOT :quux"

  for i := 0; i < bench.N; i++ {

    NewNamedParameterQuery(query)
  }
}

func BenchmarkMultiParameterParsing(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux " +
            "AND [something] = :quux2 " +
            "OR [otherStuff] NOT :quux3"

  for i := 0; i < bench.N; i++ {

    NewNamedParameterQuery(query)
  }
}

/*
  Benchmarks returning a query which has no parameters
*/
func BenchmarkNoReplacement(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = quux"
  replacer := NewNamedParameterQuery(query)

  for i := 0; i < bench.N; i++ {

    replacer.GetParsedParameters()
  }
}

/*
  Benchmarks returning a query which uses exactly one parameter
*/
func BenchmarkSingleReplacement(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux"
  replacer := NewNamedParameterQuery(query)

  for i := 0; i < bench.N; i++ {

    replacer.SetValue("quux", bench.N)
    replacer.GetParsedParameters()
  }
}

/*
  Benchmarks returning a query which has multiple occurrences of one parameter
*/
func BenchmarkMultiOccurrenceReplacement(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux " +
            "AND [something] = :quux " +
            "OR [otherStuff] NOT :quux"
  replacer := NewNamedParameterQuery(query)

  for i := 0; i < bench.N; i++ {

    replacer.SetValue("quux", bench.N)
    replacer.GetParsedParameters()
  }
}

/*
  Benchmarks returning a query which has multiple parameters
*/
func BenchmarkMultiParameterReplacement(bench *testing.B) {

  query := "SELECT [foo] FROM bar WHERE [baz] = :quux " +
            "AND [something] = :quux2 " +
            "OR [otherStuff] NOT :quux3 "
  replacer := NewNamedParameterQuery(query)

  for i := 0; i < bench.N; i++ {

    replacer.SetValue("quux", bench.N)
    replacer.SetValue("quux2", bench.N)
    replacer.SetValue("quux3", bench.N)
    replacer.GetParsedParameters()
  }
}

func Benchmark16ParameterReplacement(bench *testing.B) {
    benchmarkMultiParameter(bench, 16)
}

func Benchmark32ParameterReplacement(bench *testing.B) {
    benchmarkMultiParameter(bench, 32)
}

func Benchmark64ParameterReplacement(bench *testing.B) {
    benchmarkMultiParameter(bench, 64)
}

func Benchmark128ParameterReplacement(bench *testing.B) {
    benchmarkMultiParameter(bench, 128)
}

/*
  Benchmarks returning a query which has multiple parameters
*/
func benchmarkMultiParameter(bench *testing.B, parameterCount int) {

  var queryBuffer bytes.Buffer
  var parameterName string

  queryBuffer.WriteString("SELECT [foo] FROM bar WHERE [baz] = :quux ")
  queryLine := "AND [something] = :quux%d "

  for i := 0; i < parameterCount; i++ {

    queryBuffer.WriteString(fmt.Sprintf(queryLine, i))
  }

  replacer := NewNamedParameterQuery(queryBuffer.String())

  for i := 0; i < bench.N; i++ {

    for n := 0; n < parameterCount; n++ {
      parameterName = fmt.Sprintf("quux%d", n)
      replacer.SetValue(parameterName, bench.N)
    }

    replacer.GetParsedParameters()
  }
}
