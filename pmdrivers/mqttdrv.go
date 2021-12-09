/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmdrivers

import (
    "errors"
    "encoding/json"
    "strings"
    "strconv"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
)

const (
    MQTTDriverSchemaId                      pgschema.UUID   = "5b282915-64cb-4fcf-b5d5-5ddc6a41c679"
    MQTTDriverSchemaVersion                 string          = "1.16"
    MQTTDriverSchemaTag                     string          = "mqtt device"    

    //MQTTControlPublishName                  string          = "SendDownlink"

    MQTTPropertyResultName                  string          = "Message"
    MQTTPropertyTopicBaseName               string          = "TopicBase"
    MQTTPropertyBridgeObjectIdName          string          = "BRIDGE"

    MQTTControlSetTopicBaseName             string          = "SetTopicName"
    MQTTControlSetMQTTBridgeIdName          string          = "SetMQTTBridgeId"

    MQTTControlSetAutoProvisionName       string            = "SetAutoProvision"
    MQTTPropertyAutoProvisionName         string            = "AutoProvision"
    MQTTPropertyAutoProvisionDefaultValue string            = "true"
)

type MQTTDriver struct {
    BasicDriver
}

func NewAbstractMQTTDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *MQTTDriver {
    var driver MQTTDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = MQTTDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "MQTT Abstract Driver Schema"
    metadata.Enabled              = true
    metadata.MExternalId          = MQTTDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = MQTTDriverSchemaVersion
    metadata.Name                 = "MQTT Abstract Driver Schema"
    metadata.Type                 = pgschema.MetadataTypeDevice

    schema := pgschema.NewSchema()
    schema.Metadata     = metadata

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())

    schema.Properties   = append(schema.Properties, driver.newMQTTBridgeIdProperty())

    driver.schema = schema
    driver.schemaOwnerId = schemaOwnerId

    return &driver
}
//
//
func (this *MQTTDriver) newSetAutoProvisionControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set auto provision"
    control.Hidden          = true
    control.RPC             = MQTTControlSetAutoProvisionName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *MQTTDriver) newSetAutoProvisionControlArgEnable() *pgschema.Control {
    control := this.newSetAutoProvisionControl()
    control.Description     = "Enable"
    control.Type            = pgschema.BoolType
    control.Argument        = "enable"
    return control
}
//
//
func (this *MQTTDriver) newSetTopicBaseControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set topic base"
    control.Hidden          = true
    control.RPC             = MQTTControlSetTopicBaseName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *MQTTDriver) newSetTopicBaseControlArgTopicBase() *pgschema.Control {
    control := this.newSetTopicBaseControl()
    control.Description     = "Topic base"
    control.Type            = pgschema.StringType
    control.Argument        = "topicBase"
    return control
}
//
//
func (this *MQTTDriver) newSetMQTTBridgeIdControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set MQTT Bridge Id"
    control.Hidden          = true
    control.RPC             = MQTTControlSetMQTTBridgeIdName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *MQTTDriver) newSetMQTTBridgeIdArgBridgeId() *pgschema.Control {
    control := this.newSetMQTTBridgeIdControl()
    control.Description     = "Bridge Id"
    control.Type            = pgschema.StringType
    control.Argument        = "bridgeId"
    return control
}
//
//
func (this *MQTTDriver) newControlDecodePayload() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = BasicControlDecodePayloadName
    control.Type            = pgschema.StringType
    control.Description     = "Decode payload"
    control.Argument        = control.RPC
    return control
}
func (this *MQTTDriver) newControlDecodePayloadArgTopicName() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic name"
    control.Argument        = "topicName"
    control.Type            = pgschema.StringType
    return control
}
func (this *MQTTDriver) newControlDecodePayloadArgPayload() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic payload"
    control.Argument        = "payload"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *MQTTDriver) newResultProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = MQTTPropertyResultName
    property.Type           = pgschema.StringType
    property.Description    = MQTTPropertyResultName
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *MQTTDriver) newMQTTBridgeIdProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = MQTTPropertyBridgeObjectIdName
    property.Type           = pgschema.StringType
    property.Description    = "MQTT bridge Object Id"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *MQTTDriver) newTopicBaseProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = MQTTPropertyTopicBaseName
    property.Type           = pgschema.StringType
    property.Description    = "Topic base"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *MQTTDriver) newAutoProvisionProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = MQTTPropertyAutoProvisionName
    property.Type           = pgschema.BoolType
    property.Description    = "Auto Provision"
    property.GroupName      = BasicPropertyGroupSettingsName
    property.DefaultValue   = MQTTPropertyAutoProvisionDefaultValue
    return property
}
//
//

func (this *MQTTDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    err = this.pg.UpdateControlExecutionAck(controlMessage.Id)
    if err != nil {
        return err
    }
    
    switch  controlMessage.Name {
        case BasicControlDecodePayloadName:
            err = this.DecodePayloadController(controlMessage)

        case MQTTControlSetTopicBaseName:
            err = this.SetTopicBaseController(controlMessage)

        case MQTTControlSetMQTTBridgeIdName:
            err = this.SetMQTTBridgeIdController(controlMessage)

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
func (this *MQTTDriver) LogController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    pmlog.LogInfo("*** driver", this.GetSchemaId(), "control message:", controlMessage.GetJSON())
    return err
}

func (this *MQTTDriver) SetAutoProvisionController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackBoolArguments(controlMessage.Params)
    enableFlag, err := strconv.ParseBool(arguments.Enable)
    if err != nil {
        return err
    }
    pmlog.LogInfo("driver", this.GetSchemaId(), "set auto provision:", enableFlag)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, MQTTPropertyAutoProvisionName, strconv.FormatBool(enableFlag))
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *MQTTDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackTopicArguments(controlMessage.Params)

    result := strings.Replace(string(arguments.Payload), `"`, `\"`, -1)

    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, MQTTPropertyResultName, result)
    if err != nil {
        pmlog.LogInfo("driver", this.GetSchemaId(), "error update message property:", err)
        return err
    }
    return err
}
//
//
func (this *MQTTDriver) SetTopicBaseController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackTopicBaseArguments(controlMessage.Params)
    pmlog.LogInfo("driver", this.GetSchemaId(), "set topic base:", arguments.TopicBase)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, MQTTPropertyTopicBaseName, arguments.TopicBase)
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *MQTTDriver) SetMQTTBridgeIdController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackMQTTBridgeIdArguments(controlMessage.Params)
    pmlog.LogInfo("driver", this.GetSchemaId(), "set mqtt bridge id:", arguments.MQTTBridgeId)
    _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, MQTTPropertyBridgeObjectIdName, arguments.MQTTBridgeId)
    if err != nil {
        return err
    }
    return err
}

//**********************************************************************//

type TopicBaseArguments struct {
    TopicBase   string      `json:"topicBase"`
}

func NewTopicBaseArguments() *TopicBaseArguments {
    var arguments TopicBaseArguments
    return &arguments
}

func UnpackTopicBaseArguments(jsonString string) (*TopicBaseArguments, error) {
    var err error
    var arguments TopicBaseArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *TopicBaseArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *TopicBaseArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//**********************************************************************//

type MQTTBridgeIdArguments struct {
    MQTTBridgeId   string      `json:"bridgeId"`
}

func NewMQTTBridgeIdArguments() *MQTTBridgeIdArguments {
    var arguments MQTTBridgeIdArguments
    return &arguments
}

func UnpackMQTTBridgeIdArguments(jsonString string) (*MQTTBridgeIdArguments, error) {
    var err error
    var arguments MQTTBridgeIdArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *MQTTBridgeIdArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *MQTTBridgeIdArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}
//EOF
