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

// Config defines the config for hertz sentry.
type Config struct {
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
	SendHead bool

	// SendHead configures whether you want to add current request body when capturing sentry events.
	// Optional. Default: false
	SendBody bool

	// Timeout for the event delivery requests.
	// Optional. Default: 2 Seconds
	Timeout time.Duration
}

// configDefault is the default config
var configDefault = Config{
	RePanic:         false,
	WaitForDelivery: false,
	SendHead:        false,
	SendBody:        false,
	Timeout:         time.Second * 2,
}

// setSentryConfig set the config for sentry
func setSentryConfig(config ...Config) {
	if len(config) < 1 {
		return
	}

	// Overwrite useless value of config
	configDefault = config[0]

	if configDefault.Timeout == 0 {
		configDefault.Timeout = time.Second * 2
	}
}
