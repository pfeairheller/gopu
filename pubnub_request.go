package gopu

import (
	"net/url"
	"encoding/json"
)

type PubnubRequest struct {
	operation string
	ssl bool
	url string
	Channel string
	Message interface {}
}

func NewPubnubRequest(operation string) (*PubnubRequest) {
	req := new(PubnubRequest)
	req.operation = operation
	req.ssl = false

	return req
}

func (req *PubnubRequest) Url(publishKey string, subscribeKey string, secretKey string) (string) {
	if req.url != "" {
		return req.url
	}
	
	var Url *url.URL
	if req.ssl {
		Url, _ = url.Parse("https://" + OriginHost)
	} else {
		Url, _ = url.Parse("http://" + OriginHost)
	}

	switch req.operation {
	case "time":
		Url.Path += "/" + req.operation + "/0" 
	case "publish":
		messageBytes, err := json.Marshal(req.Message)
		if err != nil {
			panic(err)
		}
		Url.Path += "/" + req.operation + "/" + publishKey + "/" + subscribeKey + "/" + secretKey + "/" + req.Channel + "/0/" + string(messageBytes)
	}

	req.url = Url.String()
	return req.url
}
