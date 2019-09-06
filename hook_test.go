package sematextHook

import (
	"encoding/json"
	stdliberr "errors"
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

var srv *http.Server
var port int
var intercepted string

func startLogInterceptor() {
	if srv != nil {
		// lets assume that the server has been started already, but clear the intercepted message
		intercepted = ""
		return
	}

	port, _ = freeport.GetFreePort()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)
		intercepted = string(bytes)
	})

	srv = &http.Server{Addr: ":" + strconv.Itoa(port), Handler: nil}
	go srv.ListenAndServe()
}

func Test_SendingWithoutError(t *testing.T) {

	startLogInterceptor()

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

	startLogInterceptor()

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

func Test_ErrorFromStdlibShouldNotBeEmptyObject(t *testing.T) {

	startLogInterceptor()

	client := resty.New()
	client.JSONMarshal = func(v interface{}) (bytes []byte, e error) {
		return json.Marshal(v)
	}

	hook, e := NewSematextHook(client, "http://localhost:"+strconv.Itoa(port), "test", "test", "test")
	if e != nil {
		t.Fatal(e)
	}

	logrus.AddHook(hook)

	logrus.WithError(stdliberr.New("test")).Error("something went wrong, and there is some error")

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
