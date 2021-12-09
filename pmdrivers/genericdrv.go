/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmdrivers

import (
    "app/pgschema"
    "app/pgcore"
)

const (
    genericDriverSchemaId       pgschema.UUID   = "6a34e442-cc3c-4586-853e-9058e1fd7739"
    genericDriverSchemaVersion  string          = "1.36"
)

type GenericDriver struct {
    MQTTDriver
}

func NewGenericDriver(pg *pgcore.Pixcore, schemaOwnerId pgschema.UUID) *GenericDriver {
    var driver GenericDriver

    driver.pg = pg

    metadata := pgschema.NewMetadata()
    metadata.Id                   = genericDriverSchemaId
    metadata.ApplicationOwner     = schemaOwnerId
    metadata.Description          = "Generic MQTT Device"
    metadata.Enabled              = true
    metadata.MExternalId          = MQTTDriverSchemaId
    metadata.MTags                = append(metadata.MTags, MQTTDriverSchemaTag)
    metadata.MVersion             = genericDriverSchemaVersion
    metadata.Name                 = "Generic MQTT Device"
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

    schema.Controls     = append(schema.Controls, driver.newSetMQTTBridgeIdControl())
    schema.Controls     = append(schema.Controls, driver.newSetMQTTBridgeIdArgBridgeId())

    schema.Properties   = append(schema.Properties, driver.newMQTTBridgeIdProperty())
    schema.Properties   = append(schema.Properties, driver.newTopicBaseProperty())

    schema.Properties   = append(schema.Properties, driver.newDCPowerProperty())        
    schema.Properties   = append(schema.Properties, driver.newBatteryLevelProperty())
    schema.Properties   = append(schema.Properties, driver.newBatteryLowProperty())

    schema.Properties   = append(schema.Properties, driver.newResponseStatusProperty())
    schema.Properties   = append(schema.Properties, driver.newResponseTimeoutProperty())

    schema.Properties   = append(schema.Properties, driver.newResultProperty())

    driver.schema = schema
    driver.schemaOwnerId = schemaOwnerId

    return &driver
}



//EOF
