/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hertzsentry

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/getsentry/sentry-go"
)

// testCase basic info
type testCase struct {
	path    string
	method  string
	body    string
	handler app.HandlerFunc
	event   *sentry.Event
}

var yourDsn = ""

func Test_Sentry_Normal(t *testing.T) {
	// default host
	defaultHost := "localhost:6666"

	// set interval to 0 means using fs-watching mechanism.
	hertz := server.New(server.WithHostPorts(defaultHost))

	tc := testCase{
		path:   "/hello",
		method: "GET",
		handler: NewSentry(
			WithRePanic(true),
			WithSendRequest(true),
			WithWaitForDelivery(true),
			WithTimeout(3),
		),
		event: &sentry.Event{
			Level:   sentry.LevelDebug,
			Message: "test for normal",
			Request: &sentry.Request{
				URL:    "http://localhost:6666/hello",
				Method: "GET",
				Headers: map[string]string{
					"Host":            "localhost:6666",
					"User-Agent":      "hertz",
					"Content-Length":  "0",
					"Accept-Encoding": "gzip",
				},
			},
		},
	}
	// init handler func
	hertz.Use(recovery.Recovery())
	hertz.Use(tc.handler)

	// init sentry
	if err := sentry.Init(sentry.ClientOptions{
		// The DSN to use. If the DSN is not set, the client is effectively disabled.
		Dsn: yourDsn,
		// Before send callback.
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			assert.DeepEqual(t, tc.event.Request, event.Request)
			assert.DeepEqual(t, tc.event.Message, event.Message)
			assert.DeepEqual(t, tc.event.Level, event.Level)
			assert.DeepEqual(t, tc.event.Exception, event.Exception)
			fmt.Println(event)
			return event
		},
		// In debug mode, the debug information is printed to stdout to help you understand what
		// sentry is doing.
		Debug: true,
		// Configures whether SDK should generate and attach stacktraces to pure capture message calls.
		AttachStacktrace: true,
	}); err != nil {
		log.Fatal("sentry init failed")
	}

	hertz.Handle(tc.method, tc.path, func(c context.Context, ctx *app.RequestContext) {
		if hub := GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("hello", "CloudWeGo Hertz")
				scope.SetLevel(sentry.LevelDebug)
				hub.CaptureMessage("test for normal")
			})
		}
		ctx.SetStatusCode(0)
	})

	// set request head
	req, err := http.NewRequest(tc.method, "http://"+defaultHost+tc.path, strings.NewReader(tc.body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "hertz")
	testInt := uint32(0)
	hertz.Engine.OnShutdown = append(hertz.OnShutdown, func(ctx context.Context) {
		atomic.StoreUint32(&testInt, 1)
	})
	go hertz.Spin()
	time.Sleep(100 * time.Millisecond)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request %q failed: %s", tc.path, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status code = %d", resp.StatusCode)
	}
	hertz.Close()
}

func Test_Sentry_Abnormal(t *testing.T) {
	// default host
	defaultHost := "localhost:8088"

	// set interval to 0 means using fs-watching mechanism.
	hertz := server.New(server.WithHostPorts(defaultHost))

	tc := testCase{
		path:   "/hertz",
		method: "GET",
		handler: NewSentry(
			WithRePanic(true),
			WithSendBody(true),
		),
		event: &sentry.Event{
			Level:   sentry.LevelFatal,
			Message: "test for panic",
			Request: nil,
		},
	}
	// init handler func
	hertz.Use(recovery.Recovery())
	hertz.Use(tc.handler)

	// init sentry
	if err := sentry.Init(sentry.ClientOptions{
		// The DSN to use. If the DSN is not set, the client is effectively disabled.
		Dsn: yourDsn,
		// Before send callback.
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			println(event.Request)
			assert.DeepEqual(t, tc.event.Request, event.Request)
			assert.DeepEqual(t, tc.event.Message, event.Message)
			assert.DeepEqual(t, tc.event.Level, event.Level)
			assert.DeepEqual(t, tc.event.Exception, event.Exception)
			fmt.Println(event)
			return event
		},
		// In debug mode, the debug information is printed to stdout to help you understand what
		// sentry is doing.
		Debug: true,
		// Configures whether SDK should generate and attach stack traces to pure capture message calls.
		AttachStacktrace: true,
	}); err != nil {
		log.Fatal("sentry init failed")
	}

	hertz.Handle(tc.method, tc.path, func(c context.Context, ctx *app.RequestContext) {
		panic("test for panic")
	})

	// set request head
	req, err := http.NewRequest(tc.method, "http://"+defaultHost+tc.path, strings.NewReader(tc.body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "hertz")
	testInt := uint32(0)
	hertz.Engine.OnShutdown = append(hertz.OnShutdown, func(ctx context.Context) {
		atomic.StoreUint32(&testInt, 1)
	})
	go hertz.Spin()
	time.Sleep(100 * time.Millisecond)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request %q failed: %s", tc.path, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Status code = %d", resp.StatusCode)
	}
	hertz.Close()
}
