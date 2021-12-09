/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmdrivers

import (
    "encoding/json"
    "errors"
    "strconv"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
)

const (
    senso8TempDriverSchemaId        pgschema.UUID   = "a4ecf01b-5c98-4d61-8891-d5eac816472b"
    senso8TempDriverSchemaVersion   string          = "1.22"

    senso8TempPropertyTemperatureName   string          = "Temperature"
    senso8TempPropertyHumidityName      string          = "Humidity"

    senso8TempPropertyLoTempThName      string          = "LoTempTh"       
    senso8TempPropertyHiTempThName      string          = "HiTempTh"
    senso8TempPropertyLoHumiThName      string          = "LoHumiTh"
    senso8TempPropertyHiHumiThName      string          = "HiHumiTh"

    //senso8TempPropertyLoTempThDefaultValue      string          = "10"       
    //senso8TempPropertyHiTempThDefaultValue      string          = "80"
    //senso8TempPropertyLoHumiThDefaultValue      string          = "10"
    //senso8TempPropertyHiHumiThDefaultValue      string          = "90"
)

type SENSO8TempDriver struct {
    BasicDriver
}

func NewSENSO8TempDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *SENSO8TempDriver {
    var driver SENSO8TempDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = senso8TempDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "SENSO8 Temperature BLE Sensor"
    metadata.Enabled              = true
    metadata.MExternalId          = senso8TempDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = senso8TempDriverSchemaVersion
    metadata.Name                 = "SENSO8 Temperatue BLE Sensor"
    metadata.Type                 = pgschema.MetadataTypeDevice
    metadata.MManufacturer        = "Arwin"
    metadata.MPicture             = driver.GetMediaId()

    schema := pgschema.NewSchema()
    schema.Metadata     = metadata

    schema.Controls     = append(schema.Controls, driver.newControlDecodePayload())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgTopicName())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgPayload())

    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddr())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgBLEMAC())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgLoTempTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgHiTempTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgLoHumiTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgHiHumiTh())

    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddr())
    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddrArgBLEMAC())

    schema.Properties   = append(schema.Properties, driver.newBLEAddressProperty())
    schema.Properties   = append(schema.Properties, driver.newTemperatureProperty())
    schema.Properties   = append(schema.Properties, driver.newHumidityProperty())

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newDCPowerProperty())

    schema.Properties   = append(schema.Properties, driver.newBLEGatewayProperty())

    schema.Properties   = append(schema.Properties, driver.newLoTempThProperty())
    schema.Properties   = append(schema.Properties, driver.newHiTempThProperty())
    schema.Properties   = append(schema.Properties, driver.newLoHumiThProperty())
    schema.Properties   = append(schema.Properties, driver.newHiHumiThProperty())

    driver.schema           = schema
    driver.schemaOwnerId    = schemaOwnerId

    //driver.qualifiers = make(map[pgcore.UUID]Qualifier)

    return &driver
}
//
//
func (this *SENSO8TempDriver) newControlDecodePayload() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = BasicControlDecodePayloadName
    control.Type            = pgschema.StringType
    control.Description     = "Decode payload"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8TempDriver) newControlDecodePayloadArgTopicName() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic name"
    control.Argument        = "topicName"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8TempDriver) newControlDecodePayloadArgPayload() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic payload"
    control.Argument        = "payload"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8TempDriver) newControlAddBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlAddBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Add BLE address to gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8TempDriver) newControlAddBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAc"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8TempDriver) newControlAddBLEAddrArgLoTempTh() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "Low temperatue threshold"
    control.Argument        = "loTempTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8TempDriver) newControlAddBLEAddrArgHiTempTh() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "Hi temperature threshold"
    control.Argument        = "hiTempTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8TempDriver) newControlAddBLEAddrArgLoHumiTh() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "Low humidity threshold"
    control.Argument        = "loHumiTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8TempDriver) newControlAddBLEAddrArgHiHumiTh() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "Higth humidity threshold"
    control.Argument        = "hiHumiTh"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8TempDriver) newControlDeleteBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlDeleteBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Delete BLE address from gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8TempDriver) newControlDeleteBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlDeleteBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8TempDriver) newLoTempThProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyLoTempThName
    property.Type           = pgschema.IntType
    property.Description    = "Low Temperature threshold"
    property.GroupName      = BasicPropertyGroupSettingsName
    //property.DefaultValue   = senso8TempPropertyLoTempThDefaultValue
    return property
}
//
//
func (this *SENSO8TempDriver) newHiTempThProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyHiTempThName
    property.Type           = pgschema.IntType
    property.Description    = "Higth temperatue threshold"
    property.GroupName      = BasicPropertyGroupSettingsName
    //property.DefaultValue   = senso8TempPropertyHiTempThDefaultValue
    return property
}
//
//
func (this *SENSO8TempDriver) newLoHumiThProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyLoTempThName
    property.Type           = pgschema.IntType
    property.Description    = "Low humidity threshold"
    property.GroupName      = BasicPropertyGroupSettingsName
    //property.DefaultValue   = senso8TempPropertyLoHumiThDefaultValue
    return property
}
//
//
func (this *SENSO8TempDriver) newHiHumiThProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyHiHumiThName
    property.Type           = pgschema.IntType
    property.Description    = "Higth humidity threshold"
    property.GroupName      = BasicPropertyGroupSettingsName
    //property.DefaultValue   = senso8TempPropertyHiHumiThDefaultValue
    return property
}
//
//
func (this *SENSO8TempDriver) newBLEAddressProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyBLEAddressName
    property.Type           = pgschema.StringType
    property.Description    = "BLE address"
    property.GroupName      = BasicPropertyGroupSettingsName
    //property.DefaultValue   = BasicPropertyBLEAddressDefaultValue
    return property
}
//
//
func (this *SENSO8TempDriver) newTemperatureProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyTemperatureName
    property.Type           = pgschema.IntType
    property.Description    = "Temperature"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8TempDriver) newHumidityProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = senso8TempPropertyHumidityName
    property.Type           = pgschema.IntType
    property.Description    = "Humidity"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8TempDriver) newBLEGatewayProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertyObjectIdName
    property.Type           = pgschema.StringType
    property.Description    = "SENSO8 gateway ID"
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}
//
//
func (this *SENSO8TempDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    err = this.pg.UpdateControlExecutionAck(controlMessage.Id)
    if err != nil {
        return err
    }
    switch  controlMessage.Name {
        case SENSO8ControlAddBLEMACName:
            err = this.AddBLEMACController(controlMessage)

        case SENSO8ControlDeleteBLEMACName:
            err = this.DeleteBLEMACController(controlMessage)

        case BasicControlDecodePayloadName:
            err = this.DecodePayloadController(controlMessage)

        default:
            pmlog.LogInfo("driver", this.GetSchemaId(), "unable route message", controlMessage.Name)
            err = errors.New("driver unable route message")
    }
    if err != nil {
        return err
    }
    err = this.pg.CreateControlExecutionEmptyReport(controlMessage.Id, false, true)
    if err != nil {
        return err
    }
    return err
}
//
//
const senso8TempBatLowTrh int = 20  // %

func (this *SENSO8TempDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackTopicArguments(controlMessage.Params)

    var senso8data SENSO8Type05
    err = json.Unmarshal(arguments.Payload, &senso8data)
    if err != nil {
        return err
    }

    tempStr := strconv.Itoa(senso8data.Temp)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyTemperatureName, tempStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    humidityStr := strconv.Itoa(senso8data.RH)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyHumidityName, humidityStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    batLevelInt := senso8data.BLEBatt / 100
    batLevelStr := strconv.Itoa(batLevelInt)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLevelName, batLevelStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    batLowBool := true
    if batLevelInt > senso8TempBatLowTrh {
        batLowBool = false
    }
    batLowStr := strconv.FormatBool(batLowBool)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLowName, batLowStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }
    return err
}
//
//
func (this *SENSO8TempDriver) AddBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, _ := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    gwObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, SENSO8BLEGWPropertyObjectIdName)
    if err != nil {
        return err
    }
    params := pgcore.NewBLEMACArguments()
    params.BLEMAC   = arguments.BLEMAC
    params.LoTempTh = arguments.LoTempTh
    params.HiTempTh = arguments.HiTempTh
    params.LoHumiTh = arguments.LoHumiTh
    params.HiHumiTh = arguments.HiHumiTh

    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBLEAddressName, arguments.BLEMAC)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyLoTempThName, arguments.LoHumiTh)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyHiTempThName, arguments.HiTempTh)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyLoHumiThName, arguments.LoHumiTh)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, senso8TempPropertyHiHumiThName, arguments.HiHumiTh)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    err = this.pg.CreateControlExecutionStealth(gwObjectId, SENSO8ControlAddBLEMACType1Name, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *SENSO8TempDriver) DeleteBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, _ := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    gwObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, SENSO8BLEGWPropertyObjectIdName)
    if err != nil {
        return err
    }
    params := pgcore.NewBLEMACArguments()
    params.BLEMAC   = arguments.BLEMAC

    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBLEAddressName, arguments.BLEMAC)
    if err != nil {
        return err
    }
    err = this.pg.CreateControlExecutionStealth(gwObjectId, SENSO8ControlDeleteBLEMACName, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//EOF
