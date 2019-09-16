package handler_test

import (
	"github.com/go-chassis/go-chassis/control"
	"github.com/go-chassis/go-chassis/core/config"
	"github.com/go-chassis/go-chassis/core/config/model"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/core/invocation"
	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-chassis/go-chassis/examples/schemas/helloworld"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func init() {
	lager.Init(&lager.Options{
		LoggerLevel:   "INFO",
		RollingPolicy: "size",
	})
}
func TestConsumerRateLimiterDisable(t *testing.T) {
	t.Log("testing consumerratelimiter handler with qps enabled as false")
	gopath := os.Getenv("GOPATH")
	os.Setenv("CHASSIS_HOME", gopath+"/src/github.com/go-chassis/go-chassis/examples/discovery/server/")

	config.Init()
	opts := control.Options{
		Infra:   config.GlobalDefinition.Panel.Infra,
		Address: config.GlobalDefinition.Panel.Settings["address"],
	}
	err := control.Init(opts)
	assert.NoError(t, err)
	c := handler.Chain{}
	c.AddHandler(&handler.ConsumerRateLimiterHandler{})

	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.FlowControl.Consumer.QPS.Enabled = false
	i := &invocation.Invocation{
		SourceMicroService: "service1",
		SchemaID:           "schema1",
		OperationID:        "SayHello",
		Args:               &helloworld.HelloRequest{Name: "peter"},
	}
	c.Next(i, func(r *invocation.Response) error {
		assert.NoError(t, r.Err)
		log.Println(r.Result)
		return r.Err
	})

}
func init() {
	lager.Init(&lager.Options{
		LoggerLevel:   "INFO",
		RollingPolicy: "size",
	})
}
func TestConsumerRateLimiterHandler_Handle(t *testing.T) {
	t.Log("testing consumerratelimiter handler with qps enabled as true")

	config.Init()

	c := handler.Chain{}
	c.AddHandler(&handler.ConsumerRateLimiterHandler{})

	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.FlowControl.Consumer.QPS.Enabled = true
	i := &invocation.Invocation{
		MicroServiceName: "service1",
		SchemaID:         "schema1",
		OperationID:      "SayHello",
		Args:             &helloworld.HelloRequest{Name: "peter"},
	}

	c.Next(i, func(r *invocation.Response) error {
		assert.NoError(t, r.Err)
		log.Println(r.Result)
		return r.Err
	})
}

func TestConsumerRateLimiterHandler_Name(t *testing.T) {
	r1 := &handler.ConsumerRateLimiterHandler{}
	name := r1.Name()
	assert.Equal(t, "consumerratelimiter", name)

}
