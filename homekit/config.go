package homekit

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"text/template"

	"github.com/gosimple/slug"
	"github.com/hashicorp/hcl"
)

// Config contains config for homeautomation homekit bridge
type Config struct {
	MQTTConfig  *MQTTConfig     `hcl:"mqtt"`
	Devices     []*DeviceConfig `hcl:"accessory"`
	Name        string          `hcl:"name"`
	Pin         string          `hcl:"pin"`
	Port        string          `hcl:"port"`
	StoragePath string          `hcl:"storagepath"`
}

// DeviceConfig represents a single HomeKit device
type DeviceConfig struct {
	Type         string `hcl:"type,key"`
	Name         string `hcl:"name"`
	SerialNumber string `hcl:"serialnumber"`
	Manufacturer string `hcl:"manufacturer"`
	Model        string `hcl:"model"`
	Firmware     string `hcl:"firmware"`

	Services []*ServiceConfig `hcl:"service"`
}

// ServiceConfig homekit device/accessory service
type ServiceConfig struct {
	Type            string                  `hcl:"type,key"`
	Characteristics []*CharacteristicConfig `hcl:"characteristic"`
}

// CharacteristicConfig homekit service characteristic
type CharacteristicConfig struct {
	Type        string `hcl:"type,key"`
	Description string `hcl:"description"`

	Get string `hcl:"get"`
	Set string `hcl:"set"`
}

// MQTTConfig config struct for mqtt client
type MQTTConfig struct {
	Servers  []string `hcl:"servers"`
	Username string   `hcl:"username"`
	Password string   `hcl:"password"`
	ClientID string   `hcl:"client_id"`
}

// ParseConfig returns a Config struct pointer parsed from a given reader
func ParseConfig(reader io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return Config{}, err
	}

	tmpl, err := template.New("config").Funcs(map[string]interface{}{
		"lowercase": strings.ToLower,
		"uppercase": strings.ToUpper,
		"slugify":   slug.Make,
		"env":       os.Getenv,
		"array": func(vals ...interface{}) []interface{} {
			return vals
		},
	}).Parse(string(data))
	if err != nil {
		return Config{}, err
	}

	b := &bytes.Buffer{}

	if err := tmpl.Execute(b, map[string]interface{}{}); err != nil {
		return Config{}, err
	}

	conf := &Config{
		StoragePath: "./db",
	}
	err = hcl.Decode(conf, string(b.Bytes()))

	return *conf, err
}
