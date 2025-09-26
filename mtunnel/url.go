package mtunnel

import "net/url"

// server -> broker: <id>.broker.aegis.internal
// broker -> server: server.aegis.internal
//  broker -> agent: <id>.agent.aegis.internal
//  agent -> broker: broker.aegis.internal

const (
	AgentHostSuffix       = ".agent.aegis.internal"
	BrokerHost            = "broker.aegis.internal"
	BrokerHostSuffix      = "." + BrokerHost
	ServerHost            = "server.aegis.internal"
	AgentBrokerHostSuffix = ".agent" + BrokerHostSuffix
)

// NewServerURL broker/agent -> server
//
// server.aegis.internal
func NewServerURL(pth string, ws ...bool) *url.URL {
	return buildURL(ServerHost, pth, ws...)
}

// NewServerBrokerURL server -> broker
//
// <broker_id>.broker.aegis.internal
func NewServerBrokerURL(bid string, pth string, ws ...bool) *url.URL {
	return buildURL(bid+BrokerHostSuffix, pth, ws...)
}

// NewBrokerURL agent -> broker
//
// broker.aegis.internal
func NewBrokerURL(pth string, ws ...bool) *url.URL {
	return buildURL(BrokerHost, pth, ws...)
}

// NewBrokerAgentURL broker -> agent
//
// <agentID>.agent.aegis.internal
func NewBrokerAgentURL(agentID string, pth string, ws ...bool) *url.URL {
	return buildURL(agentID+AgentHostSuffix, pth, ws...)
}

func NewServerBrokerAgentURL(agentID string, pth string, ws ...bool) *url.URL {
	return buildURL(agentID+AgentBrokerHostSuffix, pth, ws...)
}

func buildURL(host, pth string, ws ...bool) *url.URL {
	scheme := "http"
	if len(ws) != 0 && ws[0] {
		scheme = "ws"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   pth,
	}
}
