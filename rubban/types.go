package rubban

import (
	"context"
)

//Task A Rubban Task that run by the scheduler
type Task interface {
	Run(context.Context)
	Name() string
}
