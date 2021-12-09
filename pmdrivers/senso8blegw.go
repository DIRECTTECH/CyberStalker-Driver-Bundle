/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmdrivers

import (
    "strconv"
    "encoding/json"
    "encoding/hex"
    "encoding/binary"
    "errors"
    "regexp"
    "sync"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
    "app/pmtools"
)

const (
    SENSO8BLEGWDriverSchemaId        pgschema.UUID   = "23035572-f87a-4d00-bfd3-ac68ddafb855"
    SENSO8BLEGWDriverSchemaVersion   string          = "1.46"

    SENSO8BLEGWPropertySNName               string      = "Serial number"
    SENSO8BLEGWPropertyBLEWLName            string      = "BLE Whitelist"
    SENSO8BLEGWPropertyReportName           string      = "Report"

    SENSO8BLEGWPropertyObjectIdName         string      = "SENSO8 GW ID"

    SENSO8ControlAddBLEMACType1Name         string      = "AddBLEMACType1"
    SENSO8ControlAddBLEMACType2Name         string      = "AddBLEMACType2"
    SENSO8ControlAddBLEMACType3Name         string      = "AddBLEMACType3"
    SENSO8ControlDeleteBLEMACName           string      = "DeleteBLEMAC"
    SENSO8ControlAddBLEMACName              string      = "AddBLEMAC"
)

type SENSO8BLEGWDriver struct {
    MQTTDriver

    autoProvision       map[pgschema.UUID]bool 
    autoProvisionMutex  sync.Mutex
}
//
func (this *SENSO8BLEGWDriver) SetAutoProvision(objectId pgschema.UUID, state bool) {
    this.autoProvisionMutex.Lock()
    defer this.autoProvisionMutex.Unlock()
    this.autoProvision[objectId] = state
}
//
//
func (this *SENSO8BLEGWDriver) GetAutoProvision(objectId pgschema.UUID) bool {
    this.autoProvisionMutex.Lock()
    defer this.autoProvisionMutex.Unlock()
    state, exists := this.autoProvision[objectId]
    if !exists {
        return false
    }
    return state
}


func (this *SENSO8BLEGWDriver) Init() error {
    var err error
    schemaId := this.GetSchemaId()
    listObjects, err := this.pg.ListObjectsBySchemaId(schemaId)
    if err != nil {
        return err
    }
    for _, object := range listObjects {
        autoProvisionStr, err := this.pg.GetObjectPropertyValue(object.Id, MQTTPropertyAutoProvisionName)
        if err != nil {
            return err
        }
        autoProvisionBool, err := strconv.ParseBool(autoProvisionStr)
        if err != nil {
            return err
        }
        this.SetAutoProvision(object.Id, autoProvisionBool)
    }
    return err
}

func NewSENSO8BLEGWDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *SENSO8BLEGWDriver {
    var driver SENSO8BLEGWDriver

    driver.pg = pg


    metadata := pgschema.NewMetadata()
    metadata.Id                   = SENSO8BLEGWDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "SENSO8 BLE Gateway"
    metadata.Enabled              = true
    metadata.MExternalId          = MQTTDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = SENSO8BLEGWDriverSchemaVersion
    metadata.Name                 = "SENSO8 BLE Gateway"
    metadata.Type                 = pgschema.MetadataTypeDevice
    metadata.MManufacturer        = "Arwin"
    metadata.MPicture             = driver.GetMediaId()

    schema := pgschema.NewSchema()
    schema.Metadata     = metadata
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayload())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgTopicName())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgPayload())

    schema.Controls     = append(schema.Controls, driver.newSetAutoProvisionControl())
    schema.Controls     = append(schema.Controls, driver.newSetAutoProvisionControlArgEnable())

    schema.Controls     = append(schema.Controls, driver.newSetTopicBaseControl())
    schema.Controls     = append(schema.Controls, driver.newSetTopicBaseControlArgTopicBase())

    schema.Controls     = append(schema.Controls, driver.newSetMQTTBridgeIdControl())
    schema.Controls     = append(schema.Controls, driver.newSetMQTTBridgeIdArgBridgeId())

    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2ArgBLEMAC())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2ArgLoTempTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2ArgHiTempTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2ArgLoHumiTh())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType2ArgHiHumiTh())

    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType3())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType3ArgBLEMAC())
    schema.Controls     = append(schema.Controls, driver.newControlAddBLEMACType3ArgDistTh())

    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEMAC())
    schema.Controls     = append(schema.Controls, driver.newControlDeleteBLEMACArgBLEMAC())

    schema.Properties   = append(schema.Properties, driver.newAutoProvisionProperty())

    schema.Properties   = append(schema.Properties, driver.newMQTTBridgeIdProperty())
    schema.Properties   = append(schema.Properties, driver.newTopicBaseProperty())

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newDCPowerProperty())

    schema.Properties   = append(schema.Properties, driver.newReportProperty())
    schema.Properties   = append(schema.Properties, driver.newSNProperty())
    schema.Properties   = append(schema.Properties, driver.newBLEWLProperty())

    driver.schema = schema
    driver.schemaOwnerId = schemaOwnerId

    driver.autoProvision = make(map[pgschema.UUID]bool)
    //
    return &driver
}
//
//

//
//
func (this *SENSO8BLEGWDriver) newControlDecodePayload() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = BasicControlDecodePayloadName
    control.Type            = pgschema.StringType
    control.Description     = "Decode payload"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8BLEGWDriver) newControlDecodePayloadArgTopicName() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic name"
    control.Argument        = "topicName"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8BLEGWDriver) newControlDecodePayloadArgPayload() *pgschema.Control {
    control := this.newControlDecodePayload()
    control.Description     = "Topic payload"
    control.Argument        = "payload"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType3() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlAddBLEMACType3Name
    control.Type            = pgschema.StringType
    control.Description     = "Adding BLE Type3 (nano S-D) address"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType3ArgBLEMAC() *pgschema.Control {
    control := this.newControlAddBLEMACType3()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType3ArgDistTh() *pgschema.Control {
    control := this.newControlAddBLEMACType3()
    control.Description     = "Distanse threshold"
    control.Argument        = "distTh"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlAddBLEMACType2Name
    control.Type            = pgschema.StringType
    control.Description     = "Adding BLE Type2 address"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2ArgBLEMAC() *pgschema.Control {
    control := this.newControlAddBLEMACType2()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}

func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2ArgLoTempTh() *pgschema.Control {
    control := this.newControlAddBLEMACType2()
    control.Description     = "Low temperatue threshold"
    control.Argument        = "loTempTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2ArgHiTempTh() *pgschema.Control {
    control := this.newControlAddBLEMACType2()
    control.Description     = "Hi temperature threshold"
    control.Argument        = "hiTempTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2ArgLoHumiTh() *pgschema.Control {
    control := this.newControlAddBLEMACType2()
    control.Description     = "Low humidity threshold"
    control.Argument        = "loHumiTh"
    control.Type            = pgschema.StringType
    return control
}
func (this *SENSO8BLEGWDriver) newControlAddBLEMACType2ArgHiHumiTh() *pgschema.Control {
    control := this.newControlAddBLEMACType2()
    control.Description     = "Higth humidity threshold"
    control.Argument        = "hiHumiTh"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8BLEGWDriver) newControlDeleteBLEMAC() *pgschema.Control {
    control := pgschema.NewControl()
    control.RPC             = SENSO8ControlDeleteBLEMACName
    control.Type            = pgschema.StringType
    control.Description     = "Deleting BLE MAC"
    control.Argument        = control.RPC
    return control
}
func (this *SENSO8BLEGWDriver) newControlDeleteBLEMACArgBLEMAC() *pgschema.Control {
    control := this.newControlDeleteBLEMAC()
    control.Description     = "BLE MAC"
    control.Argument        = "bleMAC"
    control.Type            = pgschema.StringType
    return control
}
//
//
func (this *SENSO8BLEGWDriver) newTopicBaseProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = MQTTPropertyTopicBaseName
    property.Type           = pgschema.StringType
    property.Description    = "Topic base"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *SENSO8BLEGWDriver) newIMEAAddressProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = BasicPropertyIMEAAddressName
    property.Type           = pgschema.StringType
    property.Description    = "IMEA address"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *SENSO8BLEGWDriver) newSNProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertySNName
    property.Type           = pgschema.StringType
    property.Description    = "Serial number"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *SENSO8BLEGWDriver) newReportProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertyReportName
    property.Type           = pgschema.JSONType
    property.Description    = "Raw Report"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *SENSO8BLEGWDriver) newBLEWLProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = SENSO8BLEGWPropertyBLEWLName
    property.Type           = pgschema.JSONType
    property.Description    = "BLE occupancy sensor whitelist"
    property.GroupName      = BasicPropertyGroupCredentialName
    return property
}
//
//
func (this *SENSO8BLEGWDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    err = this.pg.UpdateControlExecutionAck(controlMessage.Id)
    if err != nil {
        return err
    }
    switch  controlMessage.Name {
        case BasicControlDecodePayloadName:
            err = this.DecodePayloadController(controlMessage)

        case MQTTControlSetAutoProvisionName:
            err = this.SetAutoProvisionController(controlMessage)

        case SENSO8ControlAddBLEMACType1Name:
            err = this.AddBLEMACType1Controller(controlMessage)

        case SENSO8ControlAddBLEMACType2Name:
            err = this.AddBLEMACType2Controller(controlMessage)

        case SENSO8ControlAddBLEMACType3Name:
            err = this.AddBLEMACType3Controller(controlMessage)

        case SENSO8ControlDeleteBLEMACName:
            err = this.DeleteBLEMACController(controlMessage)

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
const (
    SENSO8DeviceType3   int = 3
    SENSO8DeviceType2   int = 2
    SENSO8DeviceType1   int = 1

    SENSO8DoorSensor        uint8 = 0x46
    SENSO8PIRSensor         uint8 = 0x02 
    SENSO8WaterLeakSensor   uint8 = 0x49

    SENSO8BLEGWBatMax int = 3800
    //senso8Pattern    string  = "^(SENSO8/nbiot/data/)(.*)$"  // For IMEA
)

func (this *SENSO8BLEGWDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, err := pgcore.UnpackTopicArguments(controlMessage.Params)
    if err != nil {
        pmlog.LogError("error UnpackTopicArguments:", err)
        return err
    }

    var senso8data SENSO8Type05
    err = json.Unmarshal(arguments.Payload, &senso8data)
    if err != nil {
        pmlog.LogError("error unmarshal senso8data", err)
        return err
    }

    if len(senso8data.BLEAddr) == len("DC9D4003E584") {

        bleAddress := senso8data.BLEAddr

        // Repack or change arguments
        bleArguments := pgcore.NewTopicArguments()
        bleArguments.TopicName  = arguments.TopicName
        bleArguments.Payload    = arguments.Payload

        controlName := BasicControlDecodePayloadName

        // 1 – nano S BLE temperature/humidity sensor
        // 2 – BLE Door Magnet, PIR and water leak probe
        // 3 – nano S-D occupancy sensor

        //autoProvisionStr, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyAutoProvisionName)
        //if err != nil {
        //    pmlog.LogError("*** driver", this.GetSchemaId(), "error:", err)
        //    return err
        //}

        //autoProvisionBool, err := strconv.ParseBool(autoProvisionStr)
        //if err != nil {
        //    pmlog.LogError("*** driver", this.GetSchemaId(), "error:", err)
        //    return err
        //}

        autoProvisionBool := this.GetAutoProvision(controlMessage.ObjectId)

        if autoProvisionBool == true {
            switch senso8data.BLEType {
                case SENSO8DeviceType1:
                    err = this.CheckOrCreateSENSO8Device(arguments.Payload, controlMessage.ObjectId,
                                            "Temperature Sensor", senso8TempDriverSchemaId)
                case SENSO8DeviceType2:
                    manu, err := DecodeSENSO8BLEAdvManu(senso8data.BLEAdvManu)
                    if err == nil {
                        switch manu.DeviceType {
                            case SENSO8DoorSensor:
                                err = this.CheckOrCreateSENSO8Device(arguments.Payload, controlMessage.ObjectId,
                                            "Door Sensor", SENSO8DoorDriverSchemaId)
                            case SENSO8PIRSensor:
                                err = this.CheckOrCreateSENSO8Device(arguments.Payload, controlMessage.ObjectId,
                                            "PIR Sensor", SENSO8PIRDriverSchemaId)
                            case SENSO8WaterLeakSensor:
                                err = this.CheckOrCreateSENSO8Device(arguments.Payload, controlMessage.ObjectId,
                                            "Water Leak Sensor", SENSO8WaterLeakDriverSchemaId)
                        }
                    }              
                case SENSO8DeviceType3:
                    err = this.CheckOrCreateSENSO8Device(arguments.Payload, controlMessage.ObjectId,
                                            "Occupacy Sensor", SENSO8OccuDriverSchemaId)
            }
            if err != nil {
                pmlog.LogError("*** driver", this.GetSchemaId(), "error create device:", err)
                return err
            }
        }

        err = this.pg.CreateControlExecutionStealthByPropertyValue(controlName, bleArguments.Pack(),
                            BasicPropertyGroupSettingsName, BasicPropertyBLEAddressName, bleAddress)
        if err != nil {
            if err != nil {
                pmlog.LogError("*** driver", this.GetSchemaId(), "unable control call ", controlName, "with error:", err.Error())
                return err
            }
        }
    }
    
    if senso8data.Bat > 0 {
        batLevelInt := 100
        if senso8data.BattPow == 1 { 
           batLevelInt = (senso8data.Bat * 100) / SENSO8BLEGWBatMax
        }
        batLevelStr := strconv.Itoa(batLevelInt)
        _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLevelName, batLevelStr)
        if err != nil {
            pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
            return err
        }

        batLowBool := true
        if batLevelInt > 10 {
            batLowBool = false
        }
        batLowStr := strconv.FormatBool(batLowBool)
        _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, BasicPropertyBatteryLowName, batLowStr)
        if err != nil {
            pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
            return err
        }
    }

    if senso8data.WlParam != nil && len(senso8data.WlParam) > 0 {
        jWhiteList, _ := json.Marshal(senso8data.WlParam)
        whiteList := pmtools.EscapeJSON(string(jWhiteList))

        _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8BLEGWPropertyBLEWLName, string(whiteList))
        if err != nil {
            pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
            return err
        }
    }

    if len(senso8data.SN) > 0 {
        serialNum := senso8data.SN
        _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8BLEGWPropertySNName, serialNum)
        if err != nil {
            pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
            return err
        }
    }

    if len(senso8data.Firmware) > 0 {
        jReport, _ := json.Marshal(senso8data)
        report := pmtools.EscapeJSON(string(jReport))
        _, err = this.pg.UpdateObjectPropertyByName(controlMessage.ObjectId, SENSO8BLEGWPropertyReportName, report)
        if err != nil {
            pmlog.LogError("*** driver", this.GetSchemaId(), "error update message property:", err)
            return err
        }
    }

    return err
}

func (this *SENSO8BLEGWDriver) CheckOrCreateSENSO8Device(payload pgcore.Bytes, downlinkObjectId pgschema.UUID, nameHint string, schemaId pgschema.UUID) error {
    var err error
    var senso8data SENSO8Type05
    err = json.Unmarshal(payload, &senso8data)
    if err != nil {
        return err
    }

    bleAddress := senso8data.BLEAddr

    objects, err := this.pg.ListObjectsBySchemaId(schemaId)
    if err != nil {
        return err
    }
    for i := range objects {
        storedBLEAddress, _ := this.pg.GetObjectPropertyValue(objects[i].Id, BasicPropertyBLEAddressName)
        if storedBLEAddress == bleAddress {
            return err
        }
    }
    pmlog.LogInfo("application trying to create new senso8 " + nameHint + "device for ", bleAddress)

    object := pgschema.NewObject()
    object.Id               = pmtools.GetNewUUID()
    object.SchemaId         = schemaId
    object.Name             = "SENSO8 " + nameHint + " #" + bleAddress
    object.Description      = "SENSO8 " + nameHint + " #" + bleAddress

    objectId, err := this.pg.CreateObject(object)
    if err != nil {
        return err
    }

    _, err = this.pg.UpdateObjectPropertyByName(objectId, BasicPropertyBLEAddressName, bleAddress)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }

    _, err = this.pg.UpdateObjectPropertyByName(objectId, SENSO8BLEGWPropertyObjectIdName, downlinkObjectId)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }

    pmlog.LogInfo("application created new senso8 occupancy device", objectId)
    return err
}







const (
    mqttBridgeControlPublishName    string  = "SendDownlink"
    senso8Pattern                   string  = "^(SENSO8/nbiot/data/)(.*)$"
    senso8ControlBase               string  = "SENSO8/nbiot/ctrl/"
)

type SENSO8AddingBLEType1 struct {
	SN      string   `json:"sn"`
	WlParam []string `json:"wl_param"`  // Pair BLE, Trh * 4 list
}

type SENSO8AddingBLEType2 struct {
	SN      string   `json:"sn"`
	WlAdd []string `json:"wl_add"`  // BLE list
}

type SENSO8AddingBLEType3 struct {
	SN      string   `json:"sn"`
	WlParam []string `json:"wl_param"`  // Pair BLE, Trh list
}

type SENSO8DeletingBLEType3 struct {
	SN      string   `json:"sn"`
	WlRM    []string `json:"wl_rm"`
}

type SENSO8SetHeartbeat struct {
	SN          string      `json:"sn"`
	Heartbeat   int         `json:"heartbeat"`  
}
//
//
func (this *SENSO8BLEGWDriver) AddBLEMACType1Controller(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, err := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    bridgeObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyBridgeObjectIdName)
    if err != nil {
        return err
    }
    
    baseTopic, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyTopicBaseName)
    if err != nil {
        return err
    }

    re := regexp.MustCompile(senso8Pattern)
    res := re.FindStringSubmatch(baseTopic)
    if len(res) < 3 {
        return errors.New("wrong senso8 base topic")
    }
    imeiCode := res[2]

    params := pgcore.NewTopicArguments()
    params.TopicName = senso8ControlBase + imeiCode

    var gwMessage SENSO8AddingBLEType1
    gwMessage.SN = imeiCode
    gwMessage.WlParam = make([]string, 0)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.BLEMAC)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.LoTempTh)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.HiTempTh)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.LoHumiTh)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.HiHumiTh)

    jBytes, _ := json.Marshal(gwMessage)
    params.Payload = jBytes

    err = this.pg.CreateControlExecutionStealth(bridgeObjectId, mqttBridgeControlPublishName, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *SENSO8BLEGWDriver) AddBLEMACType2Controller(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, err := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    bridgeObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyBridgeObjectIdName)
    if err != nil {
        return err
    }
    baseTopic, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyTopicBaseName)
    if err != nil {
        return err
    }

    re := regexp.MustCompile(senso8Pattern)
    res := re.FindStringSubmatch(baseTopic)
    if len(res) < 3 {
        return errors.New("wrong senso8 base topic")
    }
    imeiCode := res[2]

    params := pgcore.NewTopicArguments()
    params.TopicName = senso8ControlBase + imeiCode

    var gwMessage SENSO8AddingBLEType2
    gwMessage.SN = imeiCode
    gwMessage.WlAdd = make([]string, 0)
    gwMessage.WlAdd = append(gwMessage.WlAdd, arguments.BLEMAC)

    jBytes, _ := json.Marshal(gwMessage)
    params.Payload = jBytes

    err = this.pg.CreateControlExecutionStealth(bridgeObjectId, mqttBridgeControlPublishName, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *SENSO8BLEGWDriver) AddBLEMACType3Controller(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, err := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    bridgeObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyBridgeObjectIdName)
    if err != nil {
        return err
    }
    baseTopic, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyTopicBaseName)
    if err != nil {
        return err
    }

    re := regexp.MustCompile(senso8Pattern)
    res := re.FindStringSubmatch(baseTopic)
    if len(res) < 3 {
        return errors.New("wrong senso8 base topic")
    }
    imeiCode := res[2]

    params := pgcore.NewTopicArguments()
    params.TopicName = senso8ControlBase + imeiCode

    var gwMessage SENSO8AddingBLEType3
    gwMessage.SN = imeiCode
    gwMessage.WlParam = make([]string, 0)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.BLEMAC)
    gwMessage.WlParam = append(gwMessage.WlParam, arguments.DistTh)

    jBytes, _ := json.Marshal(gwMessage)
    params.Payload = jBytes

    err = this.pg.CreateControlExecutionStealth(bridgeObjectId, mqttBridgeControlPublishName, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *SENSO8BLEGWDriver) DeleteBLEMACController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    arguments, err := pgcore.UnpackBLEMACArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    bridgeObjectId, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyBridgeObjectIdName)
    if err != nil {
        return err
    }
    baseTopic, err := this.pg.GetObjectPropertyValue(controlMessage.ObjectId, MQTTPropertyTopicBaseName)
    if err != nil {
        return err
    }

    re := regexp.MustCompile(senso8Pattern)
    res := re.FindStringSubmatch(baseTopic)
    if len(res) < 3 {
        return errors.New("wrong senso8 base topic")
    }
    imeiCode := res[2]

    params := pgcore.NewTopicArguments()
    params.TopicName = senso8ControlBase + imeiCode

    var gwMessage SENSO8DeletingBLEType3
    gwMessage.SN = imeiCode
    gwMessage.WlRM = make([]string, 0)
    gwMessage.WlRM = append(gwMessage.WlRM, arguments.BLEMAC)

    jBytes, _ := json.Marshal(gwMessage)
    params.Payload = jBytes

    err = this.pg.CreateControlExecutionStealth(bridgeObjectId, mqttBridgeControlPublishName, params.Pack())
    if err != nil {
        return err
    }
    return err
}
//
//
type SENSO8HeartbeatReport struct {
	Ver          int    `json:"ver,omitempty"`
	//SN           string `json:"sn,omitempty"`
	ICCid        string `json:"iccid,omitempty"`
	RSSIFilter   int    `json:"rssi_filter,omitempty"`
	Heartbeat    []int  `json:"heartbeat,omitempty"`
	LowBatt      int    `json:"low_batt,omitempty"`
	BattPwrAlert int    `json:"batt_pwr_alert,omitempty"`
	RegTimeout   int    `json:"reg_timeout,omitempty"`
	Band         []int  `json:"band,omitempty"`
	Report       int    `json:"report,omitempty"`
	LocalPort    int    `json:"local_port,omitempty"`
	Psm          []int  `json:"psm,omitempty"`
	EnShutdown   int    `json:"en_shutdown,omitempty"`
	Secure       int    `json:"secure,omitempty"`
	Firmware     string `json:"firmware,omitempty"`
	Model        string `json:"model,omitempty"`
	ClientPrefix string `json:"client_prefix,omitempty"`
	TopicPrefix  string `json:"topic_prefix,omitempty"`
	Stat         []int  `json:"stat,omitempty"`
}

type SENSO8BLEReport struct {
	Ver     int      `json:"ver,omitempty"`
	//SN      string   `json:"sn,omitempty"`
	WlParam []string `json:"wl_param,omitempty"`
}


type SENSO8SensortMessage struct {
	Ver         int     `json:"ver,omitempty"`
	Msg         int     `json:"msg,omitempty"`
	Time        int64   `json:"time,omitempty"`
	SN          string  `json:"sn,omitempty"`
	Type        int     `json:"type,omitempty"`
	BLEType     int     `json:"ble_type,omitempty"`
	BLEAddr     string  `json:"ble_addr,omitempty"`
	BLERSSI     int     `json:"ble_rssi,omitempty"`
	RAT         string  `json:"RAT,omitempty"`
	RSSI        int     `json:"rssi,omitempty"`
	RSRP        int     `json:"rsrp,omitempty"`
	RSRQ        int     `json:"rsrq,omitempty"`
	Ci          string  `json:"ci,omitempty"`
	Tac         string  `json:"tac,omitempty"`
	Bat         int     `json:"bat,omitempty"`
	BattPow     int     `json:"batt_pow,omitempty"`
	Event       int     `json:"event,omitempty"`
	Debug       []int   `json:"debug,omitempty"`

    //BLE Type1,2
    BLEBatt     int     `json:"ble_batt,omitempty"`

    //BLE Type1
    Temp        int     `json:"temp,omitempty"`
	RH          int     `json:"RH,omitempty"`
    BLEAlert    []int   `json:"ble_alert,omitempty"`   // [lo_temp_alert, hi_temp_alert, lo_humi_alert, hi_humi_alert],

    //BLE Type2
    BLEAdvManu  string  `json:"ble_adv_manu,omitempty"`

    //BLE Type3
	Dist        int     `json:"dist,omitempty"`
	State       int     `json:"state,omitempty"`
}


type SENSO8Type05 struct {
    SENSO8HeartbeatReport
    SENSO8BLEReport
    SENSO8SensortMessage
}

type SENSO8BLEAdvManu struct {
    FWVersion       uint8           `json:"fwVersion"` 
    DeviceId        uint32          `json:"deviceId"`
    DeviceType      uint8           `json:"deviceType"`
    Heartbeat       bool            `json:"heartbeat"`
    LowBattery      bool            `json:"lowLowBattery"`
    Alarm           bool            `json:"alarm"`
    AntiTamper      bool            `json:"antiTamper"`
}

const (
    manuHeartbeatMask   uint8 = 0x01 << 3
    manuLowBatteryMask  uint8 = 0x01 << 2
    manuAlarmMask       uint8 = 0x01 << 1
    manuAntiTamperMask  uint8 = 0x01 << 0
)

func DecodeSENSO8BLEAdvManu(payloadHex string) (*SENSO8BLEAdvManu, error) {
    var err error
    var manu SENSO8BLEAdvManu

    if len(payloadHex) != len("100A7F274608EDFB") {
        return &manu, errors.New("wrong len of hex string: " + payloadHex)
    }

    payloadBytes, err := hex.DecodeString(payloadHex)
    if err != nil {
        return &manu, err
    }

    manu.FWVersion  = payloadBytes[0]
    deviceIdBytes := []byte{ 0, payloadBytes[1], payloadBytes[2], payloadBytes[3] }
    manu.DeviceId   = binary.BigEndian.Uint32(deviceIdBytes)
    manu.DeviceType = payloadBytes[4]

    var eventData uint8 = payloadBytes[5]

    if eventData & manuHeartbeatMask != 0 {
        manu.Heartbeat = true
    }
    if eventData & manuLowBatteryMask != 0 {
        manu.LowBattery = true
    }
    if eventData & manuAlarmMask != 0 {
        manu.Alarm = true
    }
    if eventData & manuAntiTamperMask != 0 {
        manu.AntiTamper = true
    }

    return &manu, err
}

func (this *SENSO8BLEAdvManu) GetJSON() string {
    jBytes, _ := json.Marshal(this)
    return string(jBytes)
} 
//EOF
