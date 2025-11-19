package tunconst

import "net/url"

// server -> broker: <id>.broker.aegis.internal
// broker -> server: server.aegis.internal
// broker -> agent:  <id>.agent.aegis.internal
// agent -> broker:  broker.aegis.internal

const (
	AgentHostSuffix       = ".agent.aegis.internal"
	BrokerHost            = "broker.aegis.internal"
	BrokerHostSuffix      = "." + BrokerHost
	ServerHost            = "server.aegis.internal"
	AgentBrokerHostSuffix = ".agent" + BrokerHostSuffix
)

// ToServer broker/agent -> server
//
// server.aegis.internal
func ToServer(pth string, ws ...bool) *url.URL {
	return buildURL(ServerHost, pth, ws...)
}

// ServerToBroker server -> broker
//
// <broker_id>.broker.aegis.internal
func ServerToBroker(bid string, pth string, ws ...bool) *url.URL {
	return buildURL(bid+BrokerHostSuffix, pth, ws...)
}

// AgentToBroker agent -> broker
//
// broker.aegis.internal
func AgentToBroker(pth string, ws ...bool) *url.URL {
	return buildURL(BrokerHost, pth, ws...)
}

// BrokerToAgent broker -> agent
//
// <agentID>.agent.aegis.internal
func BrokerToAgent(agentID string, pth string, ws ...bool) *url.URL {
	return buildURL(agentID+AgentHostSuffix, pth, ws...)
}

func ServerToAgent(agentID string, pth string, ws ...bool) *url.URL {
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
