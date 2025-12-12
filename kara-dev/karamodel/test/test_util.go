package karamodel_test

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func ReadConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{}
	configByte, err := ioutil.ReadFile("app.yml")
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(configByte, &config)
	return config, nil
}
