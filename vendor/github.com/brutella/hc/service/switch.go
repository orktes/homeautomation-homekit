// THIS FILE IS AUTO-GENERATED
package service

import (
	"github.com/brutella/hc/characteristic"
)

const TypeSwitch = "00000049-0000-1000-8000-0026BB765291"

type Switch struct {
	*Service

	On *characteristic.On
}

func NewSwitch() *Switch {
	svc := Switch{}
	svc.Service = New(TypeSwitch)

	svc.On = characteristic.NewOn()
	svc.AddCharacteristic(svc.On.Characteristic)

	return &svc
}
