/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/go-chassis/client/rest"
	"github.com/ServiceComb/go-chassis/core/client"
	"github.com/ServiceComb/go-chassis/core/common"
	"net/http"
)

// Test is the function to call provider health check api and check the response
func Test(ctx context.Context, protocol, endpoint string, expected Reply) (err error) {
	switch protocol {
	case common.ProtocolRest:
		err = restTest(ctx, endpoint, expected)
	case common.ProtocolHighway:
		err = highwayTest(ctx, endpoint, expected)
	default:
		err = fmt.Errorf("Unsupport protocol %s", protocol)
	}
	return
}

func restTest(ctx context.Context, endpoint string, expected Reply) (err error) {
	c, err := client.GetClient(common.ProtocolRest, expected.ServiceName)
	if err != nil {
		return
	}

	arg, _ := rest.NewRequest(http.MethodGet, "cse://"+expected.ServiceName+"/healthz")
	req := client.NewRequest("", "", "", arg)
	rsp := rest.NewResponse()
	defer rsp.Close()
	err = c.Call(ctx, endpoint, req, rsp)
	if err != nil {
		return
	}
	if rsp.GetStatusCode() != http.StatusOK {
		return nil
	}
	var actual Reply
	err = json.Unmarshal(rsp.ReadBody(), &actual)
	if err != nil {
		return
	}
	if actual != expected {
		return fmt.Errorf("Endpoint is belong to %s:%s:%s",
			actual.ServiceName, actual.Version, actual.AppId)
	}
	return
}

func highwayTest(ctx context.Context, endpoint string, expected Reply) (err error) {
	c, err := client.GetClient(common.ProtocolHighway, expected.ServiceName)
	if err != nil {
		return
	}

	req := client.NewRequest(expected.ServiceName, "_chassis_highway_healthz", "HighwayCheck", &Request{})
	var actual Reply
	err = c.Call(ctx, endpoint, req, &actual)
	if err != nil {
		return
	}
	if actual != expected {
		return fmt.Errorf("Endpoint is belong to %s:%s:%s",
			actual.ServiceName, actual.Version, actual.AppId)
	}
	return
}
