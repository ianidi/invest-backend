package graph

import (
	"sync"

	"github.com/ianidi/exchange-server/graph/model"
	"github.com/ianidi/exchange-server/internal/jwt"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	alert          []*model.Alert
	mu             sync.Mutex // nolint: structcheck
	ProfileHandler *jwt.ProfileHandler
}
