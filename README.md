NamedParameterQuery
--

SQL query parameters in go are positional. This means that
when writing a query, you'll need to do it like this:

	SELECT * FROM table
	WHERE col1 = ?
	AND col2 IN(?, ?, ?)

Where "?" is a parameter that you want to replace with an actual value at runtime.
Your code would need to look like this:

	sql.QueryRow(queryText, "foo", "bar", "baz", "woot")

As you can probably guess, this can lead to very unwieldy code in large queries.
You wind up needing to keep track not only of how many parameters you have, but in what
order the query expects them. Refactoring your code even once can lead to disastrous
and annoying results.

The answer to this is to use named parameters, which would look like this:

	SELECT * FROM table
	WHERE col1 = :userName
	AND col2 IN(:firstName, :lastName, :middleName)

But golang doesn't support named parameters! That's what this library is for.

Why doesn't Go support this normally?
--

Go needs to support every kind of SQL server - and not all SQL servers
support named parameters.
The servers that do support them do it with a variety of quirks and "gotchas".
But they all support positional parameters just fine.

I'm not sure why the Go authors didn't add this named parameter support
on their own, but this polyfill works fine anyway.

It's possible that someone else already implemented this, but I sure couldn't find
a pre-existing solution when I needed it.


How do I use this?
--

Probably best to check out the API docs

http://godoc.org/github.com/knetic/go-namedParameterQuery

But here are some quick examples of the main use cases.

	query := NewNamedParameterQuery("
		SELECT * FROM table
		WHERE col1 = :foo
		AND col2 IN(:firstName, :middleName, :lastName)
	")

	query.SetValue("foo", "bar")
	query.SetValue("firstName", "Alice")
	query.SetValue("lastName", "Bob")
	query.SetValue("middleName", "Eve")

	connection, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
	connection.QueryRow(query.GetParsedQuery(), (query.GetParsedParameters())...)

It doesn't matter what order you specify the parameters, or how many times they appear in the query,
they're replaced as expected.

That looks a little tedious, and feels a lot like JDBC, where each parameter is given one line.
But you can also add groups of parameters with a map:

	query := NewNamedParameterQuery("
		SELECT * FROM table
		WHERE col1 = :foo
		AND col2 IN(:firstName, :middleName, :lastName)
	")

	var parameterMap = map[string]interface{} {
		"foo": 		"bar",
		"firstName": 	"Alice",
		"lastName": 	"Bob"
		"middleName": 	"Eve",
	}

	query.SetValuesFromMap(parameterMap)

	connection, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
	connection.QueryRow(query.GetParsedQuery(), (query.GetParsedParameters())...)

That example doesn't save any space because it defines the map immediately before using it,
but if you already have a map of parameters available, this is easier.

But maybe you know the benefits of strong typing, and want to add entire structs as parameters.
No problem.

	type QueryValues struct {
		Foo string		`sqlParameterName:"foo"`
		FirstName string 	`sqlParameterName:"firstName"`
		MiddleName string `sqlParameterName:"middleName"`
		LastName string 	`sqlParameterName:"lirstName"`
	}

	query := NewNamedParameterQuery("
		SELECT * FROM table
		WHERE col1 = :foo
		AND col2 IN(:firstName, :middleName, :lastName)
	")

	parameter = new(QueryValues)
	query.SetValuesFromStruct(parameter)

	connection, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
	connection.QueryRow(query.GetParsedQuery(), (query.GetParsedParameters())...)

When defining your struct, you don't *need* to add the "sqlParameterName" tags.
But if your query uses lowercase variable names (as mine did), your struct
will need to have exportable field names (as above) you can translate between the two
with a tag.
