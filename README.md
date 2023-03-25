# NamedParameterQuery

Provides support for named parameters in SQL queries used by Go / golang programs and libraries.

SQL query parameters in go are positional. This means that
when writing a query, you'll need to do it like this:

    SELECT * FROM table
    WHERE col1 = ?
    AND col2 IN(?, ?, ?)
    AND col3 = ?

Where "?" is a parameter that you want to replace with an actual value at runtime.
Your code would need to look like this:

    sql.QueryRow(queryText, "foo", "bar", "baz", "woot", "bar")

As you can probably guess, this can lead to very unwieldy code in large queries.
You wind up needing to keep track not only of how many parameters you have, but in what
order the query expects them. Sometimes you want to reference the same variable in more
than one place in your query, which requires you to specify it more than once in your code!
Refactoring your queries even once can lead to disastrous
and annoying results.

The answer to this is to use named parameters, which would look like this:

    SELECT * FROM table
    WHERE col1 = :userName
    AND col2 IN(:firstName, :lastName, :middleName)
    AND col3 = :firstname

You would then add parameters to your query by name. This means you won't need to worry about what
order your parameters are specified, nor how many times they appear.

But golang doesn't support named parameters! That's what this library is for.

## Why doesn't Go support this normally?

Go needs to support every kind of SQL server - and not all SQL servers
support named parameters.
The servers that do support them do it with a variety of quirks and "gotchas".
But they all support positional parameters just fine.

I'm not sure why the Go authors didn't add this named parameter support
on their own, but this polyfill works fine anyway.

It's possible that someone else already implemented this, but I sure couldn't find
a pre-existing solution when I needed it.

## Isn't there a better way?


In short, not across all databases, and not without complicating your query.

There are other ways to achieve the same effect on some databases. You can
[register stored procedures which take positional parameters](http://www.mysqltutorial.org/stored-procedures-parameters.aspx), 
then call that procedure instead of writing a query. However that's a fairly specific use
case - you don't always want to store your query permanently on the server; that means you have to worry about query versioning on the server, and complicates updates to queries during deployment, and precludes you from easily deploying new queries without damaging processes relying on the old ones. For most cases, sending the entire query every time you want to use it is the better option.

Or, if your database supports it, you could [define user-local variables in your query](http://stackoverflow.com/questions/5154246/mysql-connector-j-allow-user-variables).
Usually this requires a change to your DB, connectionstring, and queries. The syntax also varies across
databases in unpredictable ways - meaning you're going to write less portable queries.

Personally I don't find those options attractive. To me, a query ought to support named parameters
without edits to your database. That's why this library exists.

## How do I use this?


Probably best to check out the API docs

But here are some quick examples of the main use cases.

### Using the parser directly

    query := namedparameter.NewQuery(`
        SELECT * FROM table
        WHERE col1 = :foo
        AND col2 IN(:firstName, :middleName, :lastName)
    `)

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

    query := namedparameter.NewQuery(`
        SELECT * FROM table
        WHERE col1 = :foo
        AND col2 IN(:firstName, :middleName, :lastName)
    `)

    var parameterMap = map[string]any {
        "foo":        "bar",
        "firstName":  "Alice",
        "lastName":   "Bob"
        "middleName": "Eve",
    }

    query.SetValuesFromMap(parameterMap)

    connection, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
    connection.QueryRow(query.GetParsedQuery(), (query.GetParsedParameters())...)

That example doesn't save any space because it defines the map immediately before using it,
but if you already have a map of parameters available, this is easier.

But maybe you know the benefits of strong typing, and want to add entire structs as parameters.
No problem.

    type QueryValues struct {
        Foo string          `sqlParameterName:"foo"`
        FirstName string    `sqlParameterName:"firstName"`
        MiddleName string   `sqlParameterName:"middleName"`
        LastName string     `sqlParameterName:"lastName"`
    }

    query := namedparameter.NewQuery(`
        SELECT * FROM table
        WHERE col1 = :foo
        AND col2 IN(:firstName, :middleName, :lastName)
    `)

    parameter = new(QueryValues)
    query.SetValuesFromStruct(parameter)

    connection, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
    connection.QueryRow(query.GetParsedQuery(), (query.GetParsedParameters())...)

When defining your struct, you don't *need* to add the "sqlParameterName" tags.
But if your query uses lowercase variable names (as mine did), your struct
will need to have exportable field names (as above) you can translate between the two
with a tag.

### Using the wrappers

    db, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
    defer db.Close()

    query := `SELECT * FROM table
              WHERE col1 = :foo
              AND col2 IN(:firstName, :middleName, :lastName)`

    rows, err := namedparameter.Using(db).Query(query, 
                     "foo", "bar",
                     "firstName", "Alice", 
                     "lastName", "Smith", 
                     "middleName", "Eve",
                 )

The order in which the parameters are passed doesn't matter, but they need to be passed in pairs key/value,
where the key has to be a string. The values can be any type supported by the driver in use.

The arguments can be passed using a `map[string]any`:

    db, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
    defer db.Close()

    query := `SELECT * FROM table
              WHERE col1 = :foo
              AND col2 IN(:firstName, :middleName, :lastName)`

    var parameterMap = map[string]any {
        "foo":        "bar",
        "firstName":  "Alice",
        "lastName":   "Bob"
        "middleName": "Eve",
    }

    rows, err := namedparameter.Using(db).Query(query, parameterMap)

`namedparameter.Using(..)` can wrap either a `*sql.DB` or a `*sql.Tx`, the methods supported by the wrapper are:

```go
Query(string, ...args) (*sql.Rows, error)
QueryContext(context.Context, string, ...args) (*sql.Rows, error)
QueryRow(string, ...args) (*sql.Row, error)
QueryRowContext(context.Context, string, ...args) (*sql.Row, error)
Exec(string, ...args) (sql.Result, error)
ExecContext(context.Context, string, ...args) (sql.Result, error)
```

Notice that `QueryRow` and `QueryRowContext` can return an error (unlike the equivalent methods in `sql.DB` and 
`sql.Tx`), this is because both the query parsing and the parameters processing can result in errors.

There is also support for queries directly from a `*sql.Conn` and the use of prepared statements. 

    db, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
    defer db.Close()

    query := `SELECT * FROM table
              WHERE col1 = :foo
              AND col2 IN(:firstName, :middleName, :lastName)`

    var parameterMap = map[string]any {
        "foo":        "bar",
        "firstName":  "Alice",
        "lastName":   "Bob"
        "middleName": "Eve",
    }

    conn, _ := db.Conn(context.Background())

    stmt, _ := namedparameter.UsingConnection(conn).PrepareContext(context.Background, query)

    rows, err := stmt.QueryContext(context.Background(), query, parameterMap)

`namedparameter.UsingConnection` supports all the context methods listed previously, and a wrapped
prepared statement will support all six methods.

## License

This implementation of Go named parameter queries is licensed under the MIT general use license.
You're free to integrate, fork, and play with this code as you feel fit without consulting the
author, as long as you provide proper credit to the author in your works. If you have questions,
issues, or patches, I'm completely open to pull requests, issues opened on github, or
emails from out of the blue.
