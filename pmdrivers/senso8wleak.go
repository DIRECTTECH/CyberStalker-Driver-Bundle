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
    SENSO8WaterLeakDriverSchemaId       pgschema.UUID   = "3a9fa67d-9701-42e4-b383-b7ea50ab28a6"
    SENSO8WaterLeakDriverSchemaVersion  string          = "1.8"
    SENSO8WaterLeakPropertyAlarmName    string          = "Alarm"
)
//
//
type SENSO8WaterLeakDriver struct {
    BasicDriver
}
//
//
func NewSENSO8WaterLeakDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *SENSO8WaterLeakDriver {
    var driver SENSO8WaterLeakDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = SENSO8WaterLeakDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "SENSO8 WaterLeak BLE Sensor"
    metadata.Enabled              = true
    metadata.MExternalId          = SENSO8WaterLeakDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = SENSO8WaterLeakDriverSchemaVersion
    metadata.Name                 = "SENSO8 WaterLeak BLE Sensor"
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

    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddr())
    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddrArgBLEMAC())
    
    schema.Properties   = append(schema.Properties, driver.newBLEAddressProperty())

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newDCPowerProperty())

    schema.Properties   = append(schema.Properties, driver.newBLEGatewayProperty())
    schema.Properties   = append(schema.Properties, driver.newAlarmProperty())

    driver.schema           = schema
    driver.schemaOwnerId    = schemaOwnerId

    return &driver
}
//
//
func (this *SENSO8WaterLeakDriver) newControlDecodePayload() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = BasicControlDecodePayloadName
    control.Type            = pgschema.StringType
    control.Description     = "Decode payload"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8WaterLeakDriver) newControlDecodePayloadArgTopicName() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic name"
    control.Argument        = "topicName"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8WaterLeakDriver) newControlDecodePayloadArgPayload() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic payload"
    control.Argument        = "payload"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8WaterLeakDriver) newControlAddBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlAddBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Add BLE address to gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8WaterLeakDriver) newControlAddBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8WaterLeakDriver) newControlDeleteBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlDeleteBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Delete BLE address from gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8WaterLeakDriver) newControlDeleteBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlDeleteBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8WaterLeakDriver) newBLEAddressProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyBLEAddressName
    property.Type           = pgschema.StringType
    property.Description    = "BLE MAC address"
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}
//
//
func (this *SENSO8WaterLeakDriver) newAlarmProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8WaterLeakPropertyAlarmName
    property.Type           = pgschema.BoolType
    property.Description    = "Sensor Alarm"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8WaterLeakDriver) newBLEGatewayProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertyObjectIdName
    property.Type           = pgschema.StringType
    property.Description    = "SENSO8 gateway Id"
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}
//
//
func (this *SENSO8WaterLeakDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
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
func (this *SENSO8WaterLeakDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackTopicArguments(controlMessage.Params)

    var senso8data SENSO8Type05
    err = json.Unmarshal(arguments.Payload, &senso8data)
    if err != nil {
        return err
    }

    manu, err := DecodeSENSO8BLEAdvManu(senso8data.BLEAdvManu)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "senso8 manu hex error:", err)
        return err
    }

    alarmStr:= strconv.FormatBool(manu.Alarm)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8WaterLeakPropertyAlarmName, alarmStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    batLowStr := strconv.FormatBool(manu.LowBattery)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLowName, batLowStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    batLevelInt := 100
    if manu.LowBattery {
        batLevelInt = 0
    }
    batLevelStr := strconv.Itoa(batLevelInt)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLevelName, batLevelStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    return err
}
//
//
func (this *SENSO8WaterLeakDriver) AddBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
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
        pmlog.LogError("update params error:", err)
    }
    err = this.pg.CreateControlExecutionStealth(gwObjectId, SENSO8ControlAddBLEMACType2Name, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *SENSO8WaterLeakDriver) DeleteBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
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
