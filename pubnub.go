package gopu

import (
	"net/http"
	"io/ioutil"
)

var (
	OriginHost = "pubsub.pubnub.com"
)

type Pubnub struct {
	publishKey string
	subscribeKey string
	secretKey string
}


func NewPubnub(publishKey string, subscribeKey string, secretKey string) (*Pubnub) {
	pubnub := new(Pubnub)
	pubnub.publishKey = publishKey
	pubnub.subscribeKey = subscribeKey
	pubnub.secretKey = secretKey

	return pubnub
}

func (pn *Pubnub) Publish(channel string, message interface{}, callback func(string)) {
	req := NewPubnubRequest("publish")
	req.Channel = channel
	req.Message = message

	data := pn.makeRequest(req)

	callback(data)
}

func (pn *Pubnub) Time(callback func(string)) {
	req := NewPubnubRequest("time")
	data := pn.makeRequest(req)

	callback(data)
}


func (pn *Pubnub) makeRequest(req *PubnubRequest) (string) {
	resp, err := http.Get(req.Url(pn.publishKey, pn.subscribeKey, pn.secretKey))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result)

}



