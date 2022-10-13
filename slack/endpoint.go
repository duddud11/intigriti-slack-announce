package slack

import (
	"net/url"
)

type Endpoint struct {
	webHook   url.URL
	clientTag string
	Channel   string
}

type Message struct {
	Text string `json:"text"`
}

func NewEndpoint(webHook url.URL, channelName string, clientTag string) Endpoint {
	return Endpoint{
		webHook:   webHook,
		clientTag: clientTag,
		Channel:   channelName,
	}
}
