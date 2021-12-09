/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package main

import (
    "errors"
    "context"
    "flag"
    "fmt"
    "os"

    "path/filepath"
    "time"
    "sync"

    "app/pmlog"
    "app/pmconfig"
    "app/pgschema"
    "app/pgcore"
    "app/pmdrivers"
    "app/pmtools"
)
//
//
func main() {
    app := NewApplication()
    err := app.StartApplication()
    if err != nil {
        pmlog.LogError("app error:", err)
        os.Exit(1)
    }
}
//
type Application struct {
    schema      *pgschema.Schema
    config      *pmconfig.Config
    pg          *pgcore.Pixcore

    objectId    pgschema.UUID
    userId      pgschema.UUID

    appCtx      context.Context
    appCancel   context.CancelFunc

    controlWatcherCtx       context.Context
    controlWatcherCancel    context.CancelFunc
    controlWatcherWG        *sync.WaitGroup

    subscrControlCancel context.CancelFunc
    subscrControlWG     *sync.WaitGroup

    drivers     map[pgschema.UUID]pmdrivers.Driver
}
//
//
func NewApplication() *Application {
    var app Application

    app.appCtx, app.appCancel = context.WithCancel(context.Background())
    app.controlWatcherCtx, app.controlWatcherCancel = context.WithCancel(app.appCtx)

    var controlWatcherWG sync.WaitGroup
    app.controlWatcherWG = &controlWatcherWG 

    var subscrControlWG sync.WaitGroup
    app.subscrControlWG = &subscrControlWG

    app.config  = pmconfig.New()
    app.pg      = pgcore.New(app.appCtx)
    app.schema  = pgschema.NewSchema()

    app.drivers = make(map[pgschema.UUID]pmdrivers.Driver, 0)
    return &app
}
//
//
func (this *Application) GetConfig() error {
    var err error
    exeName := filepath.Base(os.Args[0])
    this.config.Read(exeName + ".yml")

    flag.Usage = func() {
        fmt.Println(exeName + " version " + this.config.Version)
        fmt.Println("")
        fmt.Printf("usage: %s command [option]\n", exeName)
        fmt.Println("")
        flag.PrintDefaults()
        fmt.Println("")
    }
    flag.Parse()
    if len(os.Getenv("CONFIG_API_USERNAME")) > 0 {
        this.config.Core.Username = os.Getenv("CONFIG_API_USERNAME")
    }
    if len(os.Getenv("CONFIG_API_PASSWORD")) > 0 {
        this.config.Core.Password = os.Getenv("CONFIG_API_PASSWORD")
    }
    if len(os.Getenv("CONFIG_API_URL")) > 0 {
        this.config.Core.URL = os.Getenv("CONFIG_API_URL")
    }
    if len(os.Getenv("CONFIG_API_MEDIA")) > 0 {
        this.config.Media.URL = os.Getenv("CONFIG_API_MEDIA")
    }
    if len(os.Getenv("CONFIG_APP_DATADIR")) > 0 {
        this.config.DataDir = os.Getenv("CONFIG_APP_DATADIR")
    }
    return err
}
//
//
func (this *Application) ReStartApplication() error {
    var err error

    this.StopWoWControlSubsrWatcher()
    this.StopWWControlSubsription()
    this.WaitSControlSubsrWatcher()

    err = this.Bind()
    if err != nil {
        return err
    }
    err = this.SetupAppSchema()
    if err != nil {
        return err
    }
    err = this.Bind()
    if err != nil {
        return err
    }
    err = this.GetUserId()
    if err != nil {
        return err
    }
    err = this.GetAppObjectId()
    if err != nil {
        return err
    }
    err = this.CollectDrivers()
    if err != nil {
        return err
    }
    err = this.StartControlSubsription()
    if err != nil {
        return err
    }
    go this.StartControlSubsrWatcher()
    
    return err
}
//
//
func (this *Application) StartApplication() error {
    var err error
    pmlog.LogInfo("trying to start application")
    err = this.GetConfig()
    if err != nil {
        return err
    }

    err = this.DefineAppSchema()
    if err != nil {
        return err
    }
    err = this.ConnectionSetup()
    if err != nil {
        return err
    }

    err = this.Bind()
    if err != nil {
        return err
    }
    err = this.SetupAppSchema()
    if err != nil {
        return err
    }

    err = this.Bind()
    if err != nil {
        return err
    }

    err = this.GetUserId()
    if err != nil {
        return err
    }

    err = this.GetAppObjectId()
    if err != nil {
        return err
    }

    err = this.CollectDrivers()
    if err != nil {
        return err
    }

    //return err

    err = this.StartControlSubsription()
    if err != nil {
        return err
    }
    go this.StartControlSubsrWatcher()

    err = this.StartLoop()
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *Application) ConnectionSetup() error { 
    err := this.pg.Setup(this.config.Core.URL, this.config.Core.Username,
                    this.config.Core.Password, this.config.Core.JwtTTL,
                    this.schema.Metadata.MTags)
    return err
}
//
//
func (this *Application) Bind() error { 
    var err error
    timer := time.NewTicker(bindReconnectInterval * time.Second)
    for _ = range timer.C {
        err := this.pg.Bind()
        if err != nil {
            pmlog.LogInfo("application wainting connection to core, error:", err)
            continue
        }
        pmlog.LogInfo("application connected to core")
        break
    }
    return err
}
//  
//
func (this *Application) GetUserId() error {
    var err error
    userId, err := this.pg.GetUserIdByLogin(this.config.Core.Username)
    if err != nil {
        return err
    }
    pmlog.LogInfo("application received user id:", userId)
    this.userId = userId
    return err
}
//          
//
func (this *Application) SetupAppSchema() error {
    var err error
    //pmlog.LogInfo("trying to import application schema")
    schemaJson := this.schema.GetJSON()
    _, err = this.pg.ImportSchema(schemaJson)
    if err != nil {
        return err
    }
    //pmlog.LogInfo("done import application schema", schemaId)
    return err
}
//
//
func (this *Application) GetAppObjectId() error {
    var err error
    pmlog.LogInfo("trying to get application object id")
    this.objectId, err = this.pg.GetUserProfileId()
    if err != nil {
        return err
    }
    pmlog.LogInfo("application object id is ", this.objectId)
    return err
}
//
//
func (this *Application) CollectDrivers() error {
    var err error
    var driver pmdrivers.Driver

    this.drivers = make(map[pgschema.UUID]pmdrivers.Driver, 0)
    
    driver = pmdrivers.NewSENSO8BLEGWDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8OccuDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8TempDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8DoorDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8PIRDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8WaterLeakDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewSENSO8WaterLeakDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    driver = pmdrivers.NewGenericDriver(this.pg, this.userId)
    this.drivers[driver.GetSchemaId()] = driver

    for i := range this.drivers {
        err = this.drivers[i].WriteSchema()
        if err != nil {
            pmlog.LogError("write driver schema error:", err)
        }
        err = this.drivers[i].SetupSchema()
        if err != nil {
            pmlog.LogError("setup driver schema error:", err)
            return err
        }
    }

    driver = pmdrivers.NewUnidriverDriver(this.pg, this.userId)
    err = driver.SetupSchema()
    if err != nil {
        pmlog.LogError("setup driver schema error:", err)
        return err
    }
    err = driver.WriteSchema()
    if err != nil {
        pmlog.LogError("write driver schema error:", err)
        return err
    }


    schemaList, err := this.pg.ListSchemas()
    if err != nil {
        return err
    }
    for _, schema := range schemaList {
        if pmtools.ArrayIncludes(schema.MTags, pmdrivers.UnidriverDriverTag) {
            pmlog.LogInfo("found unidriver schema:", schema.Id)
            driver = pmdrivers.NewUnidriverDriver(this.pg, this.userId)

            newSchema, err := this.pg.ExportSchema(schema.Id)
            if err != nil {
                return err
            }
            driver.SetSchema(newSchema)
            this.drivers[driver.GetSchemaId()] = driver
        }
    }
    for i := range this.drivers {
        err = this.drivers[i].Init()
        if err != nil {
            pmlog.LogError("init driver error:", err)
            return err
        }
    }
    return err
}
//
//
const (
    aliveMessage            string          = "Alive"
    jwtExpireTh             int64           = 15

    bindReconnectInterval   time.Duration   = 1  // sec
    loopInterval            time.Duration   = 1  // sec
    aliveInterval           time.Duration   = 5  // sec
)
//
//
func (this *Application) StartLoop() error {
    var err error
    pmlog.LogInfo("start application loop")

    for {
        needRestart := false
        time.Sleep(loopInterval * time.Second)

        if (time.Now().Unix() % int64(aliveInterval)) == 0 {
            pmlog.LogInfo("application is still alive")
            _, err = this.pg.UpdateObjectPropertyByName(this.objectId, propertyMessageName, aliveMessage)
            if err != nil {
                pmlog.LogError("error update message property:", err)
                needRestart = true
            }
        }
        expirePeriod := this.pg.GetJWTExpire() - time.Now().Unix()
        if expirePeriod < jwtExpireTh {

            err = this.pg.UpdateJWToken()
            if err != nil {
                pmlog.LogError("error update jwt:", err)
                needRestart = true
            }
        }
        if needRestart {
                for {
                    pmlog.LogInfo("restart application loop")
                    err = this.ReStartApplication()
                    if err == nil {
                        break
                    }
                    time.Sleep(loopInterval * time.Second)
                }
        }
    }
    return err
}
//
//
const subcrRestartWT  time.Duration = 5 // sec
//
//
func (this *Application) StartControlSubsription() error {
    var err error
    pmlog.LogInfo("application trying to start control subscription")

    handler := func(controlMessage pgcore.ControlExecutionMessage) error {
        var err error
        err = this.RouteControlMessage(controlMessage)
        if err != nil {
            pmlog.LogError("*** control error:", err)
        }
        return err
    }

    loopFunc, cancel, err := this.pg.SubscrOnControl(this.appCtx,
                                        this.subscrControlWG, handler)
    if err != nil {
        pmlog.LogError("subsription starting error:", err)
        return err
    }
    this.subscrControlCancel = cancel

    go loopFunc()
    pmlog.LogInfo("application started control subscription")
    return err
}
//
//
func (this *Application) StopWWControlSubsription() error {
    var err error
    pmlog.LogInfo("application trying to stop control subscription")
    this.subscrControlCancel()
    this.subscrControlWG.Wait()
    pmlog.LogInfo("application stoped control subscription")
    return err
}
//
//
//
func (this *Application) StartControlSubsrWatcher() error {
    var err error
    
    pmlog.LogInfo("start control subscription watcher")

    this.controlWatcherCtx, this.controlWatcherCancel = context.WithCancel(this.appCtx)
    this.controlWatcherWG.Add(1)

    for {
        this.subscrControlWG.Wait()
        select {
            case <- this.controlWatcherCtx.Done():
                pmlog.LogInfo("control subscription watcher canceled")
                this.controlWatcherWG.Done()
                return err
            default:
        }

        time.Sleep(subcrRestartWT * time.Second)

        pmlog.LogWarning("application trying to re-start control subscription")
        err := this.StartControlSubsription()
        if err != nil {
            pmlog.LogError("application control subscription error:", err)
            continue
        }
        pmlog.LogWarning("application re-started control subscription")
    }
    return err
}
//
//
//
func (this *Application) StopWoWControlSubsrWatcher() error {
    var err error
    pmlog.LogInfo("application trying to stop control subscription watcher")
    this.controlWatcherCancel()
    return err
}

func (this *Application) WaitSControlSubsrWatcher() error {
    var err error
    pmlog.LogInfo("application waiting stop control subscription watcher")
    this.controlWatcherWG.Wait()
    pmlog.LogInfo("application stoped control subscription watcher")
    return err
}
//
//
const (
    mqttPropertyTopicBaseName   string = "TopicBase"
    mqttPropertyBridgeName      string = "BRIDGE"
    driverPropertyResponseStatusName    string = "RESPONSE_STATUS"
)

const (
    appSchemaVersion            string  = "1.31"

    mqttAppSchemaTag            string  = "mqtt app"
    mqttDriverAppSchemaTag      string  = "mqtt driver"    
    mqttBridgeAppSchemaTag      string  = "mqtt bridge"    
    mqttDriverSchemaTag         string  = "mqtt device"

    appProfileTag               string  = "app profile"
    applicationTag              string  = "application"

    controlPublishName          string  = "SendDownlink"
    controlReloadName           string  = "Reload"
    controlTestModuleName       string  = "TestModule"
    controlDecodePayloadName    string  = "DecodePayload"

    propertyGroupMeasurement    string  = "Measurements"
    propertyGroupCredential     string  = "Credentials"
    propertyGroupTopics         string  = "Topics"
    propertyGroupHealthCheck    string  = "HealthCheck"

    propertyStatusName          string  = "Status"
    propertyMessageName         string  = "Message"
    propertyTimeoutName         string  = "Timeout"

    propertyStatusDefaultValue  string  = "true"
    propertyTimeoutDefaultValue string  = "120"
)

func (this *Application) DefineAppSchema() error {
    var err error
    schema := pgschema.NewSchema()
    metadata := pgschema.NewMetadata()
    metadata.Id                   = this.config.AppSchemaId
    metadata.MExternalId          = this.config.AppSchemaId

    metadata.MTags                = append(metadata.MTags, applicationTag)
    metadata.MTags                = append(metadata.MTags, appProfileTag)
    metadata.MTags                = append(metadata.MTags, mqttDriverAppSchemaTag)

    metadata.MVersion             = appSchemaVersion
    metadata.Name                 = "MQTT Drivers"
    metadata.Description          = "MQTT Drivers"
    metadata.Type                 = pgschema.MetadataTypeApp
    schema.Metadata = metadata

    schema.Controls     = append(schema.Controls, this.newReloadControl())
    schema.Controls     = append(schema.Controls, this.newTestModuleControl())
    
    schema.Properties   = append(schema.Properties, this.newStatusProperty())
    schema.Properties   = append(schema.Properties, this.newMessageProperty())
    schema.Properties   = append(schema.Properties, this.newTimeoutProperty())

    this.schema = schema
    return err
}
//
//
func (this *Application) newReloadControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Reload application"
    control.Hidden          = false
    control.RPC             = controlReloadName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
//
//
func (this *Application) newTestModuleControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Test module"
    control.Hidden          = false
    control.RPC             = controlTestModuleName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
//
//
func (this *Application) newStatusProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyStatusName
    property.Type           = pgschema.BoolType
    property.Description    = "Application online"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = propertyStatusDefaultValue
    return property
}
//
//
func (this *Application) newMessageProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyMessageName
    property.Type           = pgschema.StringType
    property.Description    = "Status message"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = pgschema.TimeUnixEpoch
    return property
}
//
//
func (this *Application) newTimeoutProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyTimeoutName
    property.Type           = pgschema.IntType
    property.Description    = "Timeout for offline status"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = propertyTimeoutDefaultValue 
    return property
}
//
//
const (
    testReportMessage string = "module test successful"
)
//
//
func (this *Application) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    switch controlMessage.Name {
        case controlReloadName,
            controlTestModuleName:
            err = this.RouteAppControlMessage(controlMessage)

        case pmdrivers.BasicControlDecodePayloadName,
                pmdrivers.SENSO8ControlAddBLEMACType1Name,
                pmdrivers.SENSO8ControlAddBLEMACType2Name,
                pmdrivers.SENSO8ControlAddBLEMACType3Name,
                pmdrivers.SENSO8ControlAddBLEMACName,
                pmdrivers.SENSO8ControlDeleteBLEMACName,
                pmdrivers.MQTTControlSetAutoProvisionName,
                pmdrivers.MQTTControlSetMQTTBridgeIdName,
                pmdrivers.MQTTControlSetTopicBaseName:
            err = this.RouteDriversControlMessage(controlMessage)

        default:
            pmlog.LogError("unknown control message:", controlMessage.GetJSON())
            err = errors.New(fmt.Sprintf("unknown control message name %s", controlMessage.Name))
    }
    return err
}
//
//
func (this *Application) RouteAppControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    err = this.pg.UpdateControlExecutionAck(controlMessage.Id)
    if err != nil {
        return err
    }
    switch controlMessage.Name {
        case controlReloadName:
            err = this.RestartController(controlMessage)
        case controlTestModuleName:
            this.LogController(controlMessage)
        default:
            pmlog.LogError("unknown control message:", controlMessage.GetJSON())
            err = errors.New(fmt.Sprintf("unknown control message name %s", controlMessage.Name))
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
func (this *Application) RouteDriversControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    pmlog.LogDebug("### control message:", controlMessage.GetJSON())
    pmlog.LogDebug("### control helper:", controlMessage.HelperData.SchemaId)

    for key := range this.drivers {
        if controlMessage.HelperData.SchemaId == this.drivers[key].GetSchemaId() {
            
            pmlog.LogDebug("### control message routed to", this.drivers[key].GetSchemaId())
            return this.drivers[key].RouteControlMessage(controlMessage)
        }
    }
    err = errors.New(fmt.Sprintf("cannot found driver class for route message name %s", controlMessage.Name))
    return err
}
//
//
func (this *Application) LogController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    pmlog.LogInfo("*** log controller message:", controlMessage.GetJSON())
    return err
}
//
//
func (this *Application) RestartController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    pmlog.LogWarning("application received restart command")
    this.ReStartApplication()
    return err
}
//EOF
