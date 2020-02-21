package rubban

import (
	"context"
)

type Task interface {
	Run(context.Context)
	Name() string
}

