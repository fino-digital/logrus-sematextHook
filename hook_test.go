package sematextHook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Test_SendingWithoutError(t *testing.T) {

	port, _ := freeport.GetFreePort()

	var intercepted string

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)
		intercepted = string(bytes)
	})
	go http.ListenAndServe(":"+strconv.Itoa(port), nil)

	hook, e := NewSematextHook(resty.New(), "http://localhost:"+strconv.Itoa(port), "test", "test", "test")
	if e != nil {
		t.Fatal(e)
	}

	logrus.AddHook(hook)

	logrus.WithError(nil).Error("something went wrong, but there is no error")

	// wait until we intercept the message, but no longer than 1 second
	start := time.Now()
	for intercepted == "" || time.Now().Sub(start) > 1*time.Second {
		time.Sleep(5 * time.Millisecond)
	}

	if strings.Contains(intercepted, `"error":null`) {
		t.Error("message should not contain a null error ")
	}

	fmt.Println(intercepted)
}

func Test_SendingWithoutEmptyObject(t *testing.T) {

	port, _ := freeport.GetFreePort()

	var intercepted string

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)
		intercepted = string(bytes)
	})
	go http.ListenAndServe(":"+strconv.Itoa(port), nil)

	client := resty.New()
	client.JSONMarshal = func(v interface{}) (bytes []byte, e error) {
		return json.Marshal(v)
	}

	hook, e := NewSematextHook(client, "http://localhost:"+strconv.Itoa(port), "test", "test", "test")
	if e != nil {
		t.Fatal(e)
	}

	logrus.AddHook(hook)

	logrus.WithError(errors.Wrap(errors.New("the cause"), "test")).Error("something went wrong, and there is some error")

	// wait until we intercept the message, but no longer than 1 second
	start := time.Now()
	for intercepted == "" || time.Now().Sub(start) > 1*time.Second {
		time.Sleep(5 * time.Millisecond)
	}

	if strings.Contains(intercepted, `"error":{}`) {
		t.Error("message should not contain an empty object error ")
	}

	fmt.Println(intercepted)
}
