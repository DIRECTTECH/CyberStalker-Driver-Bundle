/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmdrivers

import (
    "errors"
    "encoding/json"
    "strings"
    "io"
    "path"
    "strconv"
    "reflect"
    "regexp"
    "fmt"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
    //"app/pmtools"
)

const (
    UnidriverDriverSchemaId         pgschema.UUID   = "813820b8-94dd-48db-b526-161dbdf66caf"
    UnidriverDriverSchemaVersion    string          = "1.22"

    UnidriverPropertyConfigSchemaIdName    string      = "SchemaId"
    UnidriverPropertyConfigObjectIdName    string      = "ObjectId"

    UnidriverPropertyTempName     string           = "Temp"
    UnidriverPropertyHumiName     string           = "Humi"

    UnidriverDriverTag            string           = "unidriver"
)

type UnidriverDriver struct {
    MQTTDriver
}
//
//
func (this *UnidriverDriver) SetSchemaId(schemaId pgschema.UUID) {
    this.schema.Metadata.Id = schemaId
}
//
//
func (this *UnidriverDriver) SetSchema(schema pgschema.Schema) {
    this.schema = &schema
}
//
//
func NewUnidriverDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *UnidriverDriver {
    var driver UnidriverDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = UnidriverDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "Unidriver Template"
    metadata.Enabled              = true
    metadata.MExternalId          = UnidriverDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MTags                = append(metadata.MTags, UnidriverDriverTag)
    metadata.MVersion             = UnidriverDriverSchemaVersion
    metadata.Name                 = "Unidriver Template"
    metadata.Type                 = pgschema.MetadataTypeDevice
    metadata.MManufacturer        = "Pixel"
    metadata.MPicture             = driver.GetMediaId()

    schema := pgschema.NewSchema()
    schema.Metadata     = metadata

    schema.Controls     = append(schema.Controls, driver.newControlDecodePayload())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgTopicName())
    schema.Controls     = append(schema.Controls, driver.newControlDecodePayloadArgPayload())

    schema.Controls     = append(schema.Controls, driver.newSetTopicBaseControl())
    schema.Controls     = append(schema.Controls, driver.newSetTopicBaseControlArgTopicBase())

    //schema.Controls     = append(schema.Controls, driver.newSetAutoProvisionControl())
    //schema.Controls     = append(schema.Controls, driver.newSetAutoProvisionControlArgEnable())

    //schema.Properties   = append(schema.Properties, driver.newMQTTBridgeIdProperty())
    schema.Properties   = append(schema.Properties, driver.newTopicBaseProperty())

    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())

    //schema.Properties   = append(schema.Properties, driver.newConfigSchemaIdProperty())
    //schema.Properties   = append(schema.Properties, driver.newConfigObjectIdProperty())
    //schema.Properties   = append(schema.Properties, driver.newAutoProvisionProperty())

    schema.Properties   = append(schema.Properties, driver.newTempProperty())
    schema.Properties   = append(schema.Properties, driver.newHumiProperty())
    schema.Properties   = append(schema.Properties, driver.newResultProperty())

    driver.schema           = schema
    driver.schemaOwnerId    = schemaOwnerId

    return &driver
}
//
//
func (this *UnidriverDriver) newTempProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = UnidriverPropertyTempName
    property.Type           = pgschema.IntType
    property.Description    = "Temperature (json:temp)"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *UnidriverDriver) newHumiProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = UnidriverPropertyHumiName
    property.Type           = pgschema.IntType
    property.Description    = "Humidity (json:RH)"
    property.GroupName      = BasicPropertyGroupMeasurementName
    return property
}
//
//
func (this *UnidriverDriver) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
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

        case MQTTControlSetTopicBaseName:
            err = this.SetTopicBaseController(controlMessage)

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
const MaxMQTTPayloadSize int = 64 * 1024
const MaxMQTTPayloadFragment int = 1 * 1024
//
func (this *UnidriverDriver) DecodePayloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    targetClassId := this.GetSchemaId()
    if targetClassId != controlMessage.HelperData.SchemaId {
        return errors.New("class id is worng")
    }



    if len(targetClassId) == 0 {
        return errors.New("Unidriver target class id is not defined")
    }
    pmlog.LogDebug("#### targetClassId", targetClassId)

    targetObjectId := controlMessage.ObjectId
    if len(targetObjectId) == 0 {
        return errors.New("Unidriver target object id is not defined")
    }

    
    arguments, _ := pgcore.UnpackTopicArguments(controlMessage.Params)
    //pmlog.LogDebug("### payloag", string(arguments.Payload))

    if len(arguments.Payload) > MaxMQTTPayloadSize {
        pmlog.LogWarning("driver", this.GetSchemaId(), "error: mqtt payload is more 64k")
        return errors.New("mqtt payload is more 64k")
    }

    _, err = this.pg.UpdateObjectPropertyByName(targetObjectId, MQTTPropertyResultName, string(arguments.Payload))
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }

    sourceKeyMap, err := UnidriverJSON2Path(string(arguments.Payload))
    if err != nil {
        return err
    }
    for key, value := range sourceKeyMap {
        pmlog.LogDebug("### ==== keyMap:", key, "=", value)
    }

    propKeyMap := make(map[string]string)

    for _, item := range this.schema.Properties {
        propKey, match, err := UnidriverGetPath(item.Description)
        if err != nil {
            continue
        }
        if !match {
            continue
        }
        if len(propKey) == 0 {
            continue
        }
        pmlog.LogDebug("#### property:", item.Property, item.Description, "->", propKey )
        propKeyMap[propKey] = item.Property
    }

    for propKey, propDest := range propKeyMap {
        value, exists := sourceKeyMap[propKey]
        if exists {
            pmlog.LogDebug("#### wrote", value, "--->" , propDest)
            switch reflect.TypeOf(value).Kind() {
                case reflect.Float64:
                    valueFloat64, ok := value.(float64)
                    if ok {
                        valueString := fmt.Sprintf("%.0f", valueFloat64)
                        _, err = this.pg.UpdateObjectPropertyByName(targetObjectId, propDest, valueString)
                        if err != nil {
                            pmlog.LogInfo("error update message property:", err)
                        }
                    }
                case reflect.String:
                    valueString, ok := value.(string)
                    if ok {
                        _, err = this.pg.UpdateObjectPropertyByName(targetObjectId, propDest, valueString)
                        if err != nil {
                            pmlog.LogInfo("error update message property:", err)
                        }
                    }
                case reflect.Bool:
                    valueBool, ok := value.(bool)
                    if ok {
                        valueString := strconv.FormatBool(valueBool)
                        _, err = this.pg.UpdateObjectPropertyByName(targetObjectId, propDest, valueString)
                        if err != nil {
                            pmlog.LogInfo("error update message property:", err)
                        }
                    }
                default:
                    pmlog.LogDebug("#### error, unk type", reflect.TypeOf(value).Kind())
            }
        }
    }  
    return err
}


type KeyMap = map[string]interface{}

func UnidriverJSON2Path(jsonData string) (KeyMap, error) {
    var err error
    decoder := json.NewDecoder(strings.NewReader(jsonData))
    keyMap := make(KeyMap)

    _, err = decoder.Token()
    if err != nil {
        return keyMap, err
    }

    err = UnidriverRDecode(decoder, "/", keyMap)
    if err == io.EOF {
        err = nil
    }
    return keyMap, err
}

func UnidriverRDecode(decoder *json.Decoder, jPath string, keyMap KeyMap) error {
    var err error

    for {
        key, err := decoder.Token()
        if err != nil {
            return err
        }
        if key == json.Delim('}') {
            return err
        }

        value, err := decoder.Token()
        if err != nil {
            return err
        }

        jPath := path.Join(jPath, key.(string))

        if value == json.Delim('{') {
            err = UnidriverRDecode(decoder, jPath, keyMap)
            continue
        }
        if value == json.Delim('[') {
            num := 0
            for {
                value, err := decoder.Token()
                if err != nil {
                    return err
                }
                if value == json.Delim(']') {
                    break
                }
                sNum := "["+ strconv.Itoa(num) + "]"
                aPath := jPath + sNum  
                if value == json.Delim('{') {
                    err = UnidriverRDecode(decoder, aPath, keyMap)
                    continue
                }
                keyMap[aPath] = value
                num += 1
            }
            continue
        }
        keyMap[jPath] = value
    }
    return err
}

const (
    tagPattern          string = "^.*\\([\\s]*json:(.*)[\\s]*\\).*$"
    pathSeparator       string = "/"
    altPathSeparator    string = "."
    doublePathSeparator string = "//"
)

func UnidriverGetPath(source string) (string, bool, error) {
    var result  string
    var match   bool
    var err     error
    source = strings.ReplaceAll(source, "\n", "")
    source = strings.ReplaceAll(source, "\t", " ")

    match, _ = regexp.MatchString(tagPattern, source)
    if err != nil {
        match = false
        return result, match, err
    }
    if match {
        re := regexp.MustCompile(tagPattern)
        resultArray := re.FindStringSubmatch(source)
        if len(resultArray) < 2 {
            match = false
            return result, match, err
        }
        result = resultArray[1]
        result = pathSeparator + result
        result = strings.ReplaceAll(result, altPathSeparator, pathSeparator)
        pathElems := strings.Split(result, pathSeparator)
        result = ""
        for _, elem := range pathElems {
            if len(elem) > 0 {
                result += pathSeparator + elem
            }
        }
        return result, match, err
    }
    return result, match, err
}
//EOF
