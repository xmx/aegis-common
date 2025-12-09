package shipx

import "github.com/xgfone/ship/v5"

func NewFirewall(fallback Allower) ship.Middleware {
	fm := &firewallMiddle{fallback: fallback}
	return fm.call
}

type firewallMiddle struct {
	fallback Allower
}

func (fm *firewallMiddle) call(h ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		f := fm.checkout(c)
		if f == nil {
			return h(c)
		}

		if allowed, err := f.Allowed(c.Request()); err != nil {
			return err
		} else if !allowed {
			return ship.ErrForbidden
		}

		return h(c)
	}
}

func (fm *firewallMiddle) checkout(c *ship.Context) Allower {
	if dr, ok := c.Route.Data.(DataRouter); ok {
		if f := dr.Allower(); f != nil {
			return f
		}
	}

	return fm.fallback
}
