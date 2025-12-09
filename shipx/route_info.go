package shipx

import "net/http"

type Allower interface {
	// Allowed 是否允许访问该路由。
	// 一些安全策略、IP 防火墙等需要。
	Allowed(*http.Request) (bool, error)
}

type AllowerFunc func(*http.Request) (bool, error)

func (f AllowerFunc) Allowed(r *http.Request) (bool, error) { return f(r) }

type DataRouter interface {
	Name() string
	Allower() Allower
}

func NewRouteData(name string) RouteData {
	return RouteData{name: name}
}

type RouteData struct {
	name    string
	allower Allower
}

func (r RouteData) Name() string {
	return r.name
}

func (r RouteData) Allower() Allower {
	return r.allower
}

func (r RouteData) SetAllower(allower Allower) RouteData {
	r.allower = allower
	return r
}
