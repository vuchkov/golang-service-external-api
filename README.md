# Task

## Golang service consumes an external API users filtered data

Design and implement a Go service that fetches, processes, filters, and 
persists user data from an external API.

Retrieve all users from the public endpoint: 
https://jsonplaceholder.typicode.com/users

Parse all results & Display the results in this format:
- Name: [Name]
- Email: [Email]
- Address: [Street], [Suite], [City], [Zipcode]
- Company: [Company Name], [Catch Phrase]

While displaying the results in parallel filter objects by company catch-phrase 
which could contain "task-force". In case of a match in parallel to the 
filtering operation persist the filtered objects in YAML file

## Run

Execute:
```bash
go get gopkg.in/yaml.v2 # install YAML package
go build main.go #build

go run main.go #run
```

## Tests

unit tests for individual functions and integration tests for the main workflow

Run:
```bash
go get gopkg.in/yaml.v2 # install YAML package
go test -v # run the tests

go test -cover # see the test coverage

# generate an HTML coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```
