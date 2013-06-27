package gopu

import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"fmt"
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

func (pn *Pubnub) UUID() string {
	u4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return u4.String()
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
	req.UUID = pn.UUID()
	req.Channel = channel

	data, _ := pn.makeRequest(req)
	var sub_resp []interface{}
	json.Unmarshal([]byte(data), &sub_resp)

	timetoken := sub_resp[1]

	messages := sub_resp[0].([]interface{})
	if subscription.Callback != nil {
		for _, msg := range messages {
			subscription.Callback(msg)
		}
	}

	go pn.poll_loop(channel, subscription, timetoken.(string), req.UUID, restore)

}

func (pn *Pubnub) Presence(channel string, subscription *Subscription, restore bool) {
	req := NewPubnubRequest("subscribe")
	req.UUID = pn.UUID()
	req.Channel = channel + "-pnpres"

	data, _ := pn.makeRequest(req)
	var sub_resp []interface{}
	json.Unmarshal([]byte(data), &sub_resp)

	timetoken := sub_resp[1]

	messages := sub_resp[0].([]interface{})
	if subscription.Callback != nil {
		for _, msg := range messages {
			subscription.Callback(msg)
		}
	}

	go pn.poll_loop(channel, subscription, timetoken.(string), req.UUID, restore)

}

func (pn *Pubnub) poll_loop(channel string, subscription *Subscription, timetoken string, uuid string, restore bool) {
	tt := timetoken
	connected := true
	for {
		req := NewPubnubRequest("subscribe")
		req.Channel = channel
		req.Timetoken = tt
		req.UUID = uuid

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
	client := &http.Client{}
	hreq, _ := http.NewRequest("GET", req.Url(pn.publishKey, pn.subscribeKey, pn.secretKey), nil)
	hreq.Header.Set("V", "3.3")
	hreq.Header.Set("User-Agent", "Go-Google")
	hreq.Header.Set("Accept", "*/*")
	resp, err := client.Do(hreq)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(result))
	return string(result), nil

}
