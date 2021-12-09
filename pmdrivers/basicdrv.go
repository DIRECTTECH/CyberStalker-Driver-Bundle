/*
 * Copyright: Pixel Networks <support@pixel-networks.com>
 */

package pmdrivers

import (
    "errors"
    "fmt"
    "io/ioutil"
    "strings"
    "os"
    "path/filepath"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
)

const (
    basicDriverSchemaId                     pgschema.UUID   = "9e4c046c-9ea8-4251-a444-31175a121306"
    basicDriverSchemaVersion                string          = "1.3"
    baseMediaId                             pgschema.UUID   = "352ddd71-2bc1-f4f8-1549-6fa2b6a87f16"

    // Controls
    BasicControlDecodePayloadName           string  = "DecodePayload"

    // Property groups
    BasicPropertyGroupMeasurementName       string = "Measurements"
    BasicPropertyGroupCredentialName        string = "Credentials"
    BasicPropertyGroupTopicsName            string = "Tpics"
    BasicPropertyGroupHealthCheckName       string = "HealthCheck"
    BasicPropertyGroupValueName             string = "Value"
    BasicPropertyGroupSettingsName          string = "Settings"

    // Properties
    BasicPropertyStatusName                 string = "Status"
    BasicPropertyMessageName                string = "Message"
    BasicPropertyTimeoutName                string = "Timeout"

    BasicPropertyDCPowerName                string = "DC_POWER"
    BasicPropertyDCPowerDefaultValue        string = "true"

    BasicPropertyResponseStatusName         string = "RESPONSE_STATUS"
    BasicPropertyResponseTimeoutName        string = "RESPONSE_TIMEOUT"
    BasicPropertyBatteryLowName             string = "BATTERY_LOW"
    BasicPropertyBatteryLevelName           string = "BATTERY_LEVEL"

    BasicPropertyBLEAddressName             string = "BLE address"       
    //BasicPropertyBLEAddressDefaultValue     string = "FFFFFFFFFFFF"

    BasicPropertyIMEAAddressName            string = "IMEA address"       
    //BasicPropertyIMEAAddressDefaultValue    string = "FFFFFFFFFFFF"

    // Property defaults
    BasicPropertyResponseStatusDefaultValue string = "true"
    BasicPropertyTimeoutDefaultValue        string = "120"
    BasicPropertyBatteryLevelDefaultValue   string = "100"
    BasicPropertyBatteryLowDefaultValue     string = "false"

    PropertyBoolDefaultTrue                 string = "true"
)

//type Schemer interface {
//    SetupSchema() error
//    GetSchemaId() pgschema.UUID
//}

type Driver interface {
    Init() error
    SetupSchema() error
    WriteSchema() error
 

    GetSchemaId() pgschema.UUID
    SetSchemaId(pgschema.UUID) 
    GetMediaId() pgschema.UUID
    SetSchema(pgschema.Schema)

    RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error
    Clone() interface{}
}
//
//
type BasicDriver struct {
    schema          *pgschema.Schema
    pg              *pgcore.Pixcore
    schemaOwnerId   pgschema.UUID
}
//
//
func (this *BasicDriver) Clone() interface{} {
    clone := *this
    return &clone
}
//
//
func (this *BasicDriver) GetMediaId() pgcore.UUID {
    return baseMediaId
}
//
//
func (this *BasicDriver) GetSchemaId() pgcore.UUID {
    return this.schema.Metadata.Id
}

func (this *BasicDriver) SetSchemaId(schemaId pgschema.UUID) {
    //this.schema.Metadata.Id = schemaId
}

func (this *BasicDriver) SetSchema(schema pgschema.Schema) {
    //this.schema = &schema
}
//
//
func (this *BasicDriver) Init() error {
    var err error
    pmlog.LogInfo("init driver", this.GetSchemaId())
    return err
}
//
//
const tmpPath = "./pmdata/tmp"
//
func (this *BasicDriver) WriteSchema() error {
    var err error
    err = os.MkdirAll(tmpPath, 0755)
    if err != nil {
        return err
    }
    schemaJson := this.schema.GetJSON()
    filename := strings.ToLower(this.schema.Metadata.Name + ".json")
    
    filename = strings.ReplaceAll(filename, ` `, `_`)
    filename = filepath.Join(tmpPath, filename)
    return ioutil.WriteFile(filename, []byte(schemaJson), 0644)
}
//
//
func (this *BasicDriver) SetupSchema() error {
    var err error
    schemaId := this.GetSchemaId()

    //pmlog.LogInfo("trying to create driver schema", schemaId)
    schemaJson := this.schema.GetJSON()
    uuid, err := this.pg.ImportSchema(schemaJson)
    if err != nil {
        return err
    }
    if len(uuid) > 0 {
    //    pmlog.LogInfo("driver schema", schemaId, "updated")
    }
    exists, err := this.pg.CheckSchemaExists(schemaId)
    if err != nil {
        return err
    }
    if !exists {
        return errors.New(fmt.Sprintf("unable create driver schema %s", schemaId))
    }
    //pmlog.LogInfo("driver schema", schemaId, "exists")
    return err
}
//
//
func (this *BasicDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    pmlog.LogInfo("*** basic driver", this.GetSchemaId(), "control message:", controlMessage.GetJSON())
    return err
}
//
//
func (this *BasicDriver) newStatusProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyStatusName
    property.Type           = pgschema.BoolType
    property.Description    = "Application online"
    property.GroupName      = BasicPropertyGroupHealthCheckName
    return property
}
//
//
func (this *BasicDriver) newMessageProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyMessageName
    property.Type           = pgschema.StringType
    property.Description    = "Status message"
    property.GroupName      = BasicPropertyGroupHealthCheckName
    return property
}
//
//
func (this *BasicDriver) newTimeoutProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyTimeoutName
    property.Type           = pgschema.IntType
    property.Description    = "Timeout for offline status"
    property.GroupName      = BasicPropertyGroupHealthCheckName
    property.DefaultValue   = BasicPropertyTimeoutDefaultValue
    return property
}
//
//
func (this *BasicDriver) newResponseStatusProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyResponseStatusName
    property.Type           = pgschema.BoolType
    property.Description    = "Response status"
    property.GroupName      = BasicPropertyGroupMeasurementName
    property.DefaultValue   = PropertyBoolDefaultTrue
    return property
}

//
//
func (this *BasicDriver) newBatteryLowProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyBatteryLowName
    property.Type           = pgschema.BoolType
    property.Description    = "Battery low"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *BasicDriver) newBatteryLevelProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyBatteryLevelName
    property.Type           = pgschema.IntType
    property.Description    = "Battery level"
    property.Type           = "double"
    property.Units          = "%"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *BasicDriver) newDCPowerProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyDCPowerName
    property.Type           = pgschema.BoolType
    property.Description    = "DC Power"
    property.GroupName      = BasicPropertyGroupMeasurementName
    property.DefaultValue   = PropertyBoolDefaultTrue 
    return property
}
//
//
func (this *BasicDriver) newResponseTimeoutProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyResponseTimeoutName
    property.Type           = pgschema.IntType
    property.Description    = "Response timeout"
    property.GroupName      = BasicPropertyGroupSettingsName
    property.DefaultValue   = PropertyBoolDefaultTrue
    return property
}
//EOF
