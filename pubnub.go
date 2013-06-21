package gopu

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

var (
	OriginHost = "pubsub.pubnub.com"
)

type Pubnub struct {
	publishKey   string
	subscribeKey string
	secretKey    string
}

func NewPubnub(publishKey string, subscribeKey string, secretKey string) *Pubnub {
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

	data, _ := pn.makeRequest(req)

	callback(data)
}

func (pn *Pubnub) Time(callback func(string)) {
	req := NewPubnubRequest("time")
	data, _ := pn.makeRequest(req)

	callback(data)
}

func (pn *Pubnub) UUID() []byte {
	c := 10
	b := make([]byte, c)
	n, err := io.ReadFull(rand.Reader, b)
	if n != len(b) || err != nil {
		panic(err)
	}
	// The slice should now contain random bytes instead of only zeroes.
	return b
}

type Subscription struct {
	Callback   func(interface{})
	Connect    func()
	Disconnect func()
	Reconnect  func()
	Error      func()
	Presence   func(string, string, string)
}

func (pn *Pubnub) Subscribe(channel string, subscription *Subscription, restore bool) {
	req := NewPubnubRequest("subscribe")
	req.Channel = channel

	data, _ := pn.makeRequest(req)
	var sub_resp []interface{}
	json.Unmarshal([]byte(data), &sub_resp)

	timetoken := sub_resp[1]

	go pn.poll_loop(channel, subscription, timetoken.(string), restore)

}

func (pn *Pubnub) poll_loop(channel string, subscription *Subscription, timetoken string, restore bool) {
	tt := timetoken
	connected := true
	for {
		req := NewPubnubRequest("subscribe")
		req.Channel = channel
		req.Timetoken = tt

		var sub_resp []interface{}
		data, err := pn.makeRequest(req)
		if err != nil {
			if subscription.Disconnect != nil {
				subscription.Disconnect()
			}
			connected = false

			if restore {
				continue
			} else {
				return
			}
		}

		if !connected {
			if subscription.Reconnect != nil {
				subscription.Reconnect()
			}
			connected = true
		}

		json.Unmarshal([]byte(data), &sub_resp)

		tt = sub_resp[1].(string)
		messages := sub_resp[0].([]interface{})

		if subscription.Callback != nil {
			for _, msg := range messages {
				subscription.Callback(msg)
			}
		}
	}
}

func (pn *Pubnub) makeRequest(req *PubnubRequest) (string, error) {
	resp, err := http.Get(req.Url(pn.publishKey, pn.subscribeKey, pn.secretKey))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return string(result), nil

}
