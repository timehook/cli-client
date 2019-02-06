# Timehook CLI client

[![Build Status](https://travis-ci.com/timehook/cli-client.svg?branch=master)](https://travis-ci.com/timehook/cli-client)

Timehook CLI client is a client implementation to use with [https://timehook.io](https://timehook.io)

## Usage

##### Use executable (recommended)

Download at [releases](https://github.com/timehook/cli-client/releases)

##### Compile on your own

1. Download or clone the repo.
2. Build executable `go build -o bin/timehook cmd/main/timehook.go`

## Example

Set up the api key as environment variable

    export TIMEHOOK_KEY=__YOUR_KEY__


With defaults: 

    ./bin/timehook
    
With custom values:

    ./bin/timehook --sec 11 --url https://your-url.com body --body '{"bar" : "bar"}'
      
      
For further info:
 
    ./bin/timehook --help      
