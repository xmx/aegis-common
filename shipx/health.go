package shipx

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

type Health struct{}

func NewHealth() *Health {
	return new(Health)
}

func (hlt *Health) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/health/ping").GET(hlt.ping)
	return nil
}

func (hlt *Health) ping(c *ship.Context) error {
	return c.NoContent(http.StatusNoContent)
}
