# GoForARun

Simple package to help bootstrapping projects. 

## Getting started

Run to create a project:
```
go run github.com/davfer/goforarun/cmd/create@0.0.1 <project_name>
```

Use the package directly:
```
go get github.com/davfer/goforarun
```

## Usage

Your application will need to implement the `App` interface, like:

```go
// config.go
type ExampleConfig struct {
    FrameworkConfig *app.BaseAppConfig `yaml:"framework"` 
    // Extend your config here
}

func (c *ExampleConfig) Framework() *app.BaseAppConfig {
    return c.FrameworkConfig
}

// service.go
type ExampleService struct {
	cfg *ExampleConfig
	// Add your dependencies here
}

func (e *ExampleService) Init(cfg *ExampleConfig) ([]app.RunnableServer, error) {
	e.cfg = cfg
	// Add your servers and also initialize your dependencies here
	return []app.RunnableServer{}, nil
}

func (e *ExampleService) Run(ctx context.Context) error {
	// Run your application here if needed
	fmt.Printf("Hello, %s!\n", e.cfg.Framework().ServiceName)

	return nil
}

func (e *ExampleService) Shutdown(ctx context.Context) error {
	// Shutdown your application here if needed, no need to shut down your servers
	return nil
}
```

Then, create a `main.go` file like:

```go
// main.go
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	build := &app.BuildInfo{Version: version, Commit: commit, Date: date}
	if svc, err := app.NewService[*ExampleService, *ExampleConfig](&ExampleService{}, build); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}
```
