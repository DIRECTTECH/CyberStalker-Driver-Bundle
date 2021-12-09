/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pgtools

import (
    "errors"
    "fmt"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
)
//
// SetupObject()
//
func SetupObject(pg *pgcore.Pixcore, objectId pgschema.UUID, schemaId pgschema.UUID, name string, description string) error {
    var err error

    pmlog.LogInfo("trying to create object", objectId)

    object := pgschema.NewObject()
    object.Id               = objectId  
    object.SchemaId         = schemaId
    object.Name             = name
    object.Description      = description

    objectExists, err := pg.CheckObjectExists(objectId)
    if err != nil {
        return err
    }
    if objectExists {
        pmlog.LogInfo("object", objectId, "already exists")
        return err
    }
    objectId, err = pg.CreateObject(object)
    if err != nil {
        return err
    }
    pmlog.LogInfo("object", objectId, "created")
    
    objectExists, err = pg.CheckObjectExists(objectId)
    if err != nil {
        return err
    }
    if !objectExists {
        message := fmt.Sprintf("cannot create object %s with schema id %s", object.Id, object.SchemaId)
        err = errors.New(message)
        return err
    }
    pmlog.LogInfo("object", objectId, "exists")
    return err 
}

const (
    propertyMACName string = "MAC"
)


func SetupBeacon(pg *pgcore.Pixcore, objectId pgschema.UUID, schemaId pgschema.UUID, name string, description string, mac string) error {
    
    var err error
    err = SetupObject(pg, objectId, schemaId, name, description)
    if err != nil {
        return err
    }
    _, err = pg.UpdateObjectPropertyByName(objectId, propertyMACName, mac)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
        return err
    }
    return err
}
//EOF
