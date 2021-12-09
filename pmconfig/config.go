/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmconfig

import (
    "io/ioutil"
    "encoding/json"
    "os"

    "app/pgschema"

    "github.com/go-yaml/yaml"
)

type Config struct {
    Version             string          `yaml:"-"           json:"-"`
    DataDir             string          `yanl:"-"           json:"-"`
    AppSchemaId         pgschema.UUID   `yaml:"schemaId"    json:"schemaId"`
    //AppObjectId         pgschema.UUID   `yaml:"objectId"    json:"objectId"`

    Core                Core            `yaml:"core"        json:"core"`
    Media               Media           `yaml:"media"       json:"media"`
}

type Media struct {
    URL     string                      `yaml:"url"         json:"url"`
}

type Core struct {
    URL         string  `yaml:"URL"         json:"URL"`
    Username    string  `yaml:"username"    json:"username"`
    Password    string  `yaml:"password"    json:"password"`
    JwtTTL      int     `yaml:"tokenttl"    json:"tokenttl"`
}

type Broker struct {
    Hostname    string  `yaml:"hostname"    json:"hostname"`
    Username    string  `yaml:"username"    json:"username"`
    Password    string  `yaml:"password"    json:"password"`
    Port        int     `yaml:"port"        json:"port"`
}

func (this *Config) GetJSON() string {
    jsonBytes, _ := json.MarshalIndent(this, "", "    ")
    return string(jsonBytes)
}

func (this *Config) GetYaml() string {
    jsonBytes, _ := yaml.Marshal(this)
    return string(jsonBytes)
}

func (this *Config) Write(fileName string) error {
    var data []byte
    var err error

    os.Rename(fileName, fileName + "~")

    data, err = yaml.Marshal(this)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(fileName, data, 0640)
}

func (this *Config) Read(fileName string) error {
    var data []byte
    var err error

    data, err = ioutil.ReadFile(fileName)
    if  err != nil {
        return err
    }
    return yaml.Unmarshal(data, &this)
}

func New() *Config {
    core := Core{
        URL:            "http://127.0.0.1:5000/graphql",

        Username:       "mqttdrivers",
        Password:       "weEfTgIeR8tn6Fx61WOc3nKiJEPfqieE",
        JwtTTL:         1, // min
    }
    media := Media{
        URL:            "http://127.0.0.1:5001",
    }

    return &Config{
        Version:            "1.0",
        DataDir:            "pmdata/",

        AppSchemaId:        "a9a42c43-b1f2-43db-b3aa-82c2b23c0072",     // may overlap from file config
        //AppObjectId:        "c0f8ceb4-6555-400f-87b2-4d3dfc7700b1",     // may overlap from file config

        Core:               core,
        Media:              media,
    }
}
