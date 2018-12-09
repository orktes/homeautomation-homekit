package homekit

import (
	"fmt"
	"net"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

// device represents a single homekit controllable device
type device struct {
	config *DeviceConfig
	hk     *HomeKit
	acc    *accessory.Accessory

	syncs []func() error
}

// newDevice creates a new device with given device config
func newDevice(hk *HomeKit, config *DeviceConfig) (*device, error) {
	info := accessory.Info{
		Name:         config.Name,
		Manufacturer: config.Manufacturer,
		Model:        config.Model,
		SerialNumber: config.SerialNumber,
	}

	typ, ok := stringToAccessoryType[config.Type]
	if !ok {
		return nil, fmt.Errorf("accessory type %s is not valid", config.Type)
	}

	acc := accessory.New(info, typ)

	d := &device{
		hk:  hk,
		acc: acc,
	}

	for _, serviceConfig := range config.Services {
		serv := service.New(stringToServiceType[serviceConfig.Type])
		for _, charConfig := range serviceConfig.Characteristics {
			func(charConfig *CharacteristicConfig) {
				var char *characteristic.Characteristic
				switch charConfig.Type {
				case "brightness":
					briChar := characteristic.NewBrightness()
					char = briChar.Characteristic
				case "on":
					onChar := characteristic.NewOn()
					char = onChar.Characteristic
				}

				if char != nil {
					if charConfig.Get != "" {
						d.syncs = append(d.syncs, func() error {
							val, err := hk.exec(charConfig.Get, nil)
							if err != nil {
								return err
							}
							char.UpdateValue(val.Export())
							return nil
						})
					}

					if charConfig.Set != "" {
						char.OnValueUpdateFromConn(func(conn net.Conn, c *characteristic.Characteristic, new, old interface{}) {
							hk.exec(charConfig.Set, map[string]interface{}{
								"value": new,
							})
						})
					}

					serv.AddCharacteristic(char)
				} else {
					// TODO handle char that is not supported
				}
			}(charConfig)
		}
		acc.AddService(serv)
	}

	return d, nil
}

// sync device charasteristics from mqtt
func (d *device) sync() error {
	// TODO convert to multi error
	var syncErr error
	for _, sFn := range d.syncs {
		err := sFn()
		if err != nil {
			syncErr = err
		}
	}

	return syncErr
}

func (d *device) getAccessory() *accessory.Accessory {
	return d.acc
}

var stringToAccessoryType = map[string]accessory.AccessoryType{
	"lightbulb": accessory.TypeLightbulb,
	"switch":    accessory.TypeSwitch,
}

var stringToServiceType = map[string]string{
	"lightbulb": service.TypeLightbulb,
	"switch":    service.TypeSwitch,
}
