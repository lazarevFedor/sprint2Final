package entities

import (
	"pkg"
	"sync"
)

var (
	Wg          = &sync.WaitGroup{}
	ParserMutex = &sync.Mutex{}
	ParsersTree = pkg.NewRBTree()
	Tasks       = &pkg.Queue{}
	Expressions = pkg.NewSafeMap()
)
