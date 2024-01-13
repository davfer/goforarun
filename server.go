package app

import (
	"context"
)

type RunnableServer interface {
	Info() *InfoServer // TODO: Probably not needed
	Run(context.Context) error
	Shutdown(context.Context) error
}

// InfoServer contains the information of a server to be started.
type InfoServer struct {
	Net  string
	Host string
	Port string
	Name string
}
