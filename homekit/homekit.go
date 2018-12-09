package homekit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/brutella/hc/accessory"

	"github.com/brutella/hc"

	"github.com/orktes/homeautomation/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/orktes/goja"
)

// HomeKit bridge for homeautomation mqtt platform
type HomeKit struct {
	c mqtt.Client
	t hc.Transport

	subscriptionID int
	subscriptions  map[string]map[int]mqtt.MessageHandler
	data           map[string]interface{}

	devices []*device

	runtime      *goja.Runtime
	runtimeMutex sync.Mutex

	sync.Mutex
}

// New returns a HomeKit instance for given devices
func New(conf Config) (*HomeKit, error) {

	hk := &HomeKit{
		runtime: goja.New(),

		subscriptions: map[string]map[int]mqtt.MessageHandler{},
		data:          map[string]interface{}{},
	}

	hk.runtime.Set("get", hk.get)
	hk.runtime.Set("set", hk.set)
	hk.runtime.Set("toRange", hk.convertFloatValueToRange)

	opts := mqtt.NewClientOptions()
	for _, server := range conf.MQTTConfig.Servers {

		opts = opts.AddBroker(server)
	}
	if conf.MQTTConfig.ClientID != "" {
		opts = opts.SetClientID(conf.MQTTConfig.ClientID)
	}
	if conf.MQTTConfig.Username != "" {
		opts = opts.SetUsername(conf.MQTTConfig.Username)
	}
	if conf.MQTTConfig.Password != "" {
		opts = opts.SetPassword(conf.MQTTConfig.Password)
	}

	opts = opts.SetDefaultPublishHandler(hk.handler)

	c := mqtt.NewClient(opts)

	hk.c = c

	err := hk.initDevices(conf.Devices)
	if err != nil {
		return nil, err
	}

	err = hk.initTransport(conf)

	return hk, err
}

func (hk *HomeKit) sync() error {
	var anyErr error
	for _, d := range hk.devices {
		err := d.sync()
		if err != nil {
			anyErr = err
		}
	}
	return anyErr
}

func (hk *HomeKit) initDevices(deviceConfigs []*DeviceConfig) error {
	if len(deviceConfigs) == 0 {
		return errors.New("no devices configured")
	}

	for _, devConfig := range deviceConfigs {
		device, err := newDevice(hk, devConfig)
		if err != nil {
			return err
		}

		hk.devices = append(hk.devices, device)
	}

	return nil
}

func (hk *HomeKit) initTransport(conf Config) error {
	accs := make([]*accessory.Accessory, len(hk.devices))

	for i, d := range hk.devices {
		accs[i] = d.getAccessory()
	}

	bridge := accessory.NewBridge(accessory.Info{
		Name: conf.Name,
	})

	t, err := hc.NewIPTransport(hc.Config{
		Pin:         conf.Pin,
		Port:        conf.Port,
		StoragePath: conf.StoragePath,
	}, bridge.Accessory, accs...)

	hk.t = t

	return err
}

func (hk *HomeKit) exec(str string, context map[string]interface{}) (val goja.Value, err error) {
	hk.runtimeMutex.Lock()
	defer hk.runtimeMutex.Unlock()

	contextData := []byte("{}")

	if context != nil {
		contextData, err = json.Marshal(context)
		if err != nil {
			return nil, err
		}
	}

	script := fmt.Sprintf(`
		with(%s) {
			%s
		}
	`, string(contextData), str)

	return hk.runtime.RunString(script)
}

func (hk *HomeKit) handler(client mqtt.Client, msg mqtt.Message) {
	hk.Lock()
	defer hk.Unlock()

	topic := msg.Topic()
	topicParts := strings.Split(topic, "/")
	if len(topicParts) >= 2 && topicParts[1] == "status" {
		valKey := strings.Join(append([]string{topicParts[0]}, topicParts[2:]...), "/")
		var val interface{}
		json.Unmarshal(msg.Payload(), &val)
		hk.data[valKey] = val
	}

	for subTopic, subs := range hk.subscriptions {
		if subTopic != topic {
			if strings.HasSuffix(subTopic, "#") {
				if !strings.HasPrefix(topic, subTopic[:len(subTopic)-1]) {
					continue
				}
			} else {
				continue
			}
		}
		for _, sub := range subs {
			go sub(client, msg)
		}
	}

	go hk.sync()
}

func (hk *HomeKit) subscribe(topic string, handler mqtt.MessageHandler) int {
	hk.Lock()
	defer hk.Unlock()

	id := hk.subscriptionID
	hk.subscriptionID++

	if len(hk.subscriptions[topic]) == 0 {
		if token := hk.c.Subscribe(topic, 1, hk.handler); token.Wait() && token.Error() != nil {
			// TODO handle in a proper way
			panic(token.Error())
		}
		hk.subscriptions[topic] = map[int]mqtt.MessageHandler{}
	}

	hk.subscriptions[topic][id] = handler

	return id
}

func (hk *HomeKit) unsubscribe(topic string, id int) {
	hk.Lock()
	defer hk.Unlock()

	if subs, ok := hk.subscriptions[topic]; ok {
		delete(subs, id)
		if len(subs) == 0 {
			// TODO figure out a proper way to unsubscribe
		}
	}
}

func (hk *HomeKit) get(call goja.FunctionCall) goja.Value {
	key := call.Argument(0).String()

getVal:
	hk.Lock()
	val, ok := hk.data[key]
	hk.Unlock()

	if !ok {
		ch := make(chan struct{})

		statusTopic := util.ConvertValueToTopic(key, "status")
		id := hk.subscribe(statusTopic, func(client mqtt.Client, msg mqtt.Message) {
			ch <- struct{}{}
		})
		defer hk.unsubscribe(statusTopic, id)

		// TODO figure out right qos and retain
		if token := hk.c.Publish(util.ConvertValueToTopic(key, "get"), 0, false, []byte{}); token.Wait() && token.Error() != nil {
			// TODO handle error
		}

		<-ch // TODO timeout etc

		goto getVal
	}

	return hk.runtime.ToValue(val)

}

func (hk *HomeKit) convertFloatValueToRange(call goja.FunctionCall) goja.Value {
	val := call.Argument(0).ToFloat()
	inputRangeJS := call.Argument(1).Export().([]interface{})
	outputRangeJS := call.Argument(2).Export().([]interface{})

	inputRange := make([]float64, len(inputRangeJS))
	outputRange := make([]float64, len(outputRangeJS))

	for i, v := range inputRangeJS {
		inputRange[i] = convertInterfaceToFloat64(v)
	}

	for i, v := range outputRangeJS {
		outputRange[i] = convertInterfaceToFloat64(v)
	}

	fittedValue, err := util.ConvertFloatValueToRange(inputRange, outputRange, val)
	if err != nil {
		panic(err)
	}

	return hk.runtime.ToValue(fittedValue)
}

func (hk *HomeKit) set(call goja.FunctionCall) goja.Value {
	key := call.Argument(0).String()
	val := call.Argument(1).Export()

	topic := util.ConvertValueToTopic(key, "set")

	hk.Lock()
	hk.data[key] = val
	hk.Unlock()

	if b, err := json.Marshal(val); err == nil {
		if token := hk.c.Publish(topic, 1, false, b); token.Wait() && token.Error() != nil {
			// TODO handle error
		}
	}

	return goja.Undefined()
}

// Close closes HomeKit bridge and underlaying mqtt connection
func (hk *HomeKit) Close(ctx context.Context) error {
	hk.c.Disconnect(uint(0))
	hk.t.Stop()
	return nil
}

// Start starts homekit bridge (blocks until stopped)
func (hk *HomeKit) Start() error {
	if token := hk.c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	err := hk.sync()
	if err != nil {
		return err
	}

	hk.t.Start()

	return nil
}
