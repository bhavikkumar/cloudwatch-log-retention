[![Build Status](https://travis-ci.com/bhavikkumar/cloudwatch-log-retention.svg?branch=master)](https://travis-ci.com/bhavikkumar/cloudwatch-log-retention)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=cloudwatch-log-retention&metric=coverage)](https://sonarcloud.io/dashboard?id=cloudwatch-log-retention)
[![Go Report Card](https://goreportcard.com/badge/github.com/bhavikkumar/cloudwatch-log-retention)](https://goreportcard.com/report/github.com/bhavikkumar/cloudwatch-log-retention)
![GitHub](https://img.shields.io/github/license/bhavikkumar/cloudwatch-log-retention.svg)
![GitHub release](https://img.shields.io/github/release/bhavikkumar/cloudwatch-log-retention.svg)
# cloudwatch-log-retention

Lambda function which sets the retention period on cloudwatch log groups when the log group is create or if the retention period is modified.

## Building the function

Preparing a binary to deploy to AWS Lambda requires that it is compiled for Linux and placed into a .zip file.

## For developers on Linux and macOS
``` shell
# Remember to build your handler executable for Linux!
GOOS=linux GOARCH=amd64 go build -o main main.go
zip main.zip main
```

## For developers on Windows

Windows developers may have trouble producing a zip file that marks the binary as executable on Linux. To create a .zip that will work on AWS Lambda, the `build-lambda-zip` tool may be helpful.

Get the tool
``` shell
go.exe get -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip
```

Use the tool from your `GOPATH`. If you have a default installation of Go, the tool will be in `%USERPROFILE%\Go\bin`. 

in cmd.exe:
``` bat
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o main main.go
%USERPROFILE%\Go\bin\build-lambda-zip.exe -o main.zip main
```

in Powershell:
``` posh
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o main main.go
~\Go\Bin\build-lambda-zip.exe -o main.zip main
```