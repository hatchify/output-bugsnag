module github.com/hatchify/output-bugsnag

go 1.13

require (
	github.com/bugsnag/bugsnag-go v1.5.3
	github.com/fatih/color v1.7.0 // indirect
	github.com/hatchify/output v0.0.0-20191023001025-e91d8413f743
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/sirupsen/logrus v1.4.2
)

replace github.com/bugsnag/bugsnag-go => ./hooks/bugsnag/bugsnag-go
