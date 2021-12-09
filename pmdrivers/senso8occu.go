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
    SENSO8OccuDriverSchemaId        pgschema.UUID   = "566d45b8-604e-4e10-bcde-f49558871a65"
    SENSO8OccuDriverSchemaVersion   string          = "1.29"

    SENSO8OccuPropertyDistanseName   string         = "Distanse"
    SENSO8OccuPropertyOccupancyName  string         = "Occupancy state"
    SENSO8OccuPropertyDistThName     string         = "DistThreshold"
)

type SENSO8OccuDriver struct {
    BasicDriver
}

func NewSENSO8OccuDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *SENSO8OccuDriver {
    var driver SENSO8OccuDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = SENSO8OccuDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "SENSO8 Occupancy BLE Sensor"
    metadata.Enabled              = true
    metadata.MExternalId          = SENSO8OccuDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = SENSO8OccuDriverSchemaVersion
    metadata.Name                 = "SENSO8 Occupancy BLE Sensor"
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
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEAddrArgDistTh())

    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddr())
    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEAddrArgBLEMAC())
    
    schema.Properties   = append(schema.Properties, driver.newBLEAddressProperty())
    schema.Properties   = append(schema.Properties, driver.newDictanseProperty())
    schema.Properties   = append(schema.Properties, driver.newOccupancyProperty())

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newDCPowerProperty())

    schema.Properties   = append(schema.Properties, driver.newBLEGatewayProperty())
    schema.Properties   = append(schema.Properties, driver.newDistThresholdProperty())

    driver.schema           = schema
    driver.schemaOwnerId    = schemaOwnerId

    return &driver
}
//
//
func (this *SENSO8OccuDriver) newControlDecodePayload() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = BasicControlDecodePayloadName
    control.Type            = pgschema.StringType
    control.Description     = "Decode payload"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8OccuDriver) newControlDecodePayloadArgTopicName() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic name"
    control.Argument        = "topicName"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8OccuDriver) newControlDecodePayloadArgPayload() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic payload"
    control.Argument        = "payload"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8OccuDriver) newControlAddBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlAddBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Add BLE address to gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8OccuDriver) newControlAddBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "Distanse threshold"
    control.Argument        = "distTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8OccuDriver) newControlAddBLEAddrArgDistTh() *pgschema.Control {
    control := this.newControlAddBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAc"
    control.Type            = pgschema.StringType
    return control
}

//
//
func (this *SENSO8OccuDriver) newControlDeleteBLEAddr() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlDeleteBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Delete BLE address from gateway whitelist"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8OccuDriver) newControlDeleteBLEAddrArgBLEMAC() *pgschema.Control {
    control := this.newControlDeleteBLEAddr()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
//
// Device: newBLEAddressProperty()
//
func (this *SENSO8OccuDriver) newBLEAddressProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyBLEAddressName
    property.Type           = pgschema.StringType
    property.Description    = BasicPropertyBLEAddressName
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}
//
//
func (this *SENSO8OccuDriver) newDictanseProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8OccuPropertyDistanseName
    property.Type           = pgschema.IntType
    property.Description    = SENSO8OccuPropertyDistanseName
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8OccuDriver) newOccupancyProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8OccuPropertyOccupancyName
    property.Type           = pgschema.BoolType
    property.Description    = SENSO8OccuPropertyOccupancyName
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8OccuDriver) newBLEGatewayProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertyObjectIdName
    property.Type           = pgschema.StringType
    property.Description    = "SENSO8 gateway ID"
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}


func (this *SENSO8OccuDriver) newDistThresholdProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8OccuPropertyDistThName
    property.Type           = pgschema.StringType
    property.Description    = "Distanse threshold"
    property.GroupName      = BasicPropertyGroupSettingsName
    return property
}


//
// Driver: RouteControlMessage()
//
func (this *SENSO8OccuDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
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
const SENSO8OccuBatLowTrh int = 20  // %
//
//
func (this *SENSO8OccuDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackTopicArguments(controlMessage.Params)

    var senso8data SENSO8Type05
    err = json.Unmarshal(arguments.Payload, &senso8data)
    if err != nil {
        return err
    }

    distanseStr := strconv.Itoa(senso8data.Dist)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8OccuPropertyDistanseName, distanseStr)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }

    var occupancyBool bool = false
    if senso8data.State == 1 {
        occupancyBool = true
    }
    
    occupancyStr:= strconv.FormatBool(occupancyBool)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8OccuPropertyOccupancyName, occupancyStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }

    batLevelInt := senso8data.BLEBatt / 100
    batLevelStr := strconv.Itoa(batLevelInt)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLevelName, batLevelStr)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
    }

    batLowBool := true
    if batLevelInt > SENSO8OccuBatLowTrh {
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

func (this *SENSO8OccuDriver) AddBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
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
    params.DistTh   = arguments.DistTh

    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBLEAddressName, arguments.BLEMAC)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8OccuPropertyDistThName, arguments.DistTh)
    if err != nil {
        pmlog.LogError("*** driver", this.GetSchemaId(), "update params error:", err)
        return err
    }

    err = this.pg.CreateControlExecutionStealth(gwObjectId, SENSO8ControlAddBLEMACType3Name, params.Pack())
    if err != nil {
        return err
    }

    return err
}
//
//
func (this *SENSO8OccuDriver) DeleteBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
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
