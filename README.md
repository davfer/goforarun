# GoForARun

Simple package to help bootstrapping projects. 

## Getting started

Run to create a project:
```
go run github.com/davfer/goforarun/cmd/create <project_name>
```

Use the package directly:
```
go get github.com/davfer/goforarun
```

## Usage

Your application will need to implement the `Application` interface, like:

```go
