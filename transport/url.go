package transport

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
func NewServerURL(pth string) *url.URL {
	return buildURL(ServerHost, pth)
}

// NewServerBrokerURL server -> broker
//
// <broker_id>.broker.aegis.internal
func NewServerBrokerURL(bid string, pth string) *url.URL {
	return buildURL(bid+BrokerHostSuffix, pth)
}

// NewBrokerURL agent -> broker
//
// broker.aegis.internal
func NewBrokerURL(pth string) *url.URL {
	return buildURL(BrokerHost, pth)
}

// NewBrokerAgentURL broker -> agent
//
// <agentID>.agent.aegis.internal
func NewBrokerAgentURL(agentID string, pth string) *url.URL {
	return buildURL(agentID+AgentHostSuffix, pth)
}

func NewServerBrokerAgentURL(agentID string, pth string) *url.URL {
	return buildURL(agentID+AgentBrokerHostSuffix, pth)
}

func buildURL(host, pth string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   pth,
	}
}
