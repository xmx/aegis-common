package muxproto

import "net/url"

const (
	AgentHostSuffix  = ".agent.aegis.internal"
	BrokerHost       = "broker.aegis.internal"
	BrokerHostSuffix = "." + BrokerHost
	ServerHost       = "server.aegis.internal"
	// AgentBrokerHostSuffix = ".agent" + BrokerHostSuffix
)

// ServerToBrokerURL server -> broker
//
// <broker_id>.broker.aegis.internal
func ServerToBrokerURL(brokerID string, path string, ws ...bool) *url.URL {
	return buildURL(brokerID+BrokerHostSuffix, path, ws)
}

// ToServerURL broker -> server, agent -> server
//
// server.aegis.internal
func ToServerURL(path string, ws ...bool) *url.URL {
	return buildURL(ServerHost, path, ws)
}

// AgentToBrokerURL agent -> broker
//
// broker.aegis.internal
func AgentToBrokerURL(path string, ws ...bool) *url.URL {
	return buildURL(BrokerHost, path, ws)
}

// BrokerToAgentURL broker -> agent
//
// <agentID>.agent.aegis.internal
func BrokerToAgentURL(agentID string, pth string, ws ...bool) *url.URL {
	return buildURL(agentID+AgentHostSuffix, pth, ws)
}

func buildURL(host, path string, ws []bool) *url.URL {
	scheme := "http"
	if len(ws) > 0 && ws[0] {
		scheme = "ws"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
}
