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

import "time"

// Option set to config unique option
type Option struct {
	F func(o *options)
}

// Options defines the config for hertz sentry.
type options struct {
	// Optional. Default: false
	rePanic bool

	// Optional. Default: false
	waitForDelivery bool

	// Optional. Default: false
	sendRequest bool

	// Optional. Default: false
	sendBody bool

	// Optional. Default: 2 Seconds
	timeout time.Duration
}

func NewOptions(opts ...Option) options {
	cfg := options{
		rePanic:         false,
		waitForDelivery: false,
		sendRequest:     false,
		sendBody:        false,
		timeout:         2 * time.Second,
	}
	cfg.Apply(opts)
	return cfg
}

func (o *options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

// WithRePanic configures whether Sentry should repanic after recovery.
// Set to true, if Recover middleware is used.
func WithRePanic(rePanic bool) Option {
	return Option{F: func(o *options) {
		o.rePanic = rePanic
	}}
}

// WithWaitForDelivery configures whether you want to block the request before moving forward with the response.
// If Recover middleware is used, it's safe to either skip this option or set it to false.
func WithWaitForDelivery(waitForDelivery bool) Option {
	return Option{F: func(o *options) {
		o.waitForDelivery = waitForDelivery
	}}
}

// WithSendRequest configures whether you want to add current request head when capturing sentry events.
func WithSendRequest(sendRequest bool) Option {
	return Option{F: func(o *options) {
		o.sendRequest = sendRequest
	}}
}

// WithSendBody configures whether you want to add current request body when capturing sentry events.
func WithSendBody(sendBody bool) Option {
	return Option{F: func(o *options) {
		o.sendBody = sendBody
	}}
}

// WithTimeout configs timeout for the event delivery requests.
func WithTimeout(timeout time.Duration) Option {
	return Option{F: func(o *options) {
		if timeout != 0 {
			o.timeout = timeout
		}
	}}
}
