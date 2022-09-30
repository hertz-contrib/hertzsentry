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
	F func(o *Options)
}

// Options defines the config for hertz sentry.
type Options struct {
	// RePanic configures whether Sentry should repanic after recovery.
	// Set to true, if Recover middleware is used.
	// Optional. Default: false
	RePanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// If Recover middleware is used, it's safe to either skip this option or set it to false.
	// Optional. Default: false
	WaitForDelivery bool

	// SendHead configures whether you want to add current request head when capturing sentry events.
	// Optional. Default: false
	SendRequest bool

	// SendHead configures whether you want to add current request body when capturing sentry events.
	// Optional. Default: false
	SendBody bool

	// Timeout for the event delivery requests.
	// Optional. Default: 2 Seconds
	Timeout time.Duration
}

func NewOptions(opts ...Option) Options {
	options := Options{
		RePanic:         false,
		WaitForDelivery: false,
		SendRequest:     false,
		SendBody:        false,
		Timeout:         2,
	}
	options.Apply(opts)
	return options
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func WithRePanic(rePanic bool) Option {
	return Option{F: func(o *Options) {
		o.RePanic = rePanic
	}}
}

func WithWaitForDelivery(waitForDelivery bool) Option {
	return Option{F: func(o *Options) {
		o.WaitForDelivery = waitForDelivery
	}}
}

func WithSendRequest(SendRequest bool) Option {
	return Option{F: func(o *Options) {
		o.SendRequest = SendRequest
	}}
}

func WithSendBody(sendBody bool) Option {
	return Option{F: func(o *Options) {
		o.SendBody = sendBody
	}}
}

func WithTimeout(timeout time.Duration) Option {
	return Option{F: func(o *Options) {
		if timeout == 0 {
			o.Timeout = 2 * time.Second
			return
		}
		o.Timeout = timeout
	}}
}
