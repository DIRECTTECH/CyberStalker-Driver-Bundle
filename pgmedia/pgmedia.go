/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgmedia


import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    //"strings"

    "app/pgschema"
    "app/pgcore"
    "app/pmlog"
    "net/textproto"
    "strings"

    "github.com/stoewer/go-strcase"
)

const (
    MediaSchemaID string = "72df8b1b-a20c-49b1-85fb-9433a82a8088"
)

func UploadMediaFile(pg *pgcore.Pixcore, mediaURL string, id pgschema.UUID, filename string) (string, error) {
    var err error
    var result pgschema.UUID

    presentName := filepath.Base(filename)
    presentName = strings.TrimSuffix(presentName, filepath.Ext(presentName))
    presentName = strcase.UpperCamelCase(presentName)
    
    exists, err := pg.CheckSchemaExists(MediaSchemaID)
    if err != nil {
        return result, err
    }
    if !exists {
        return result, errors.New("unable register media data because media schema not exists")
    }
    
    _, err = pg.RegisterMediaFile(id, presentName, MediaSchemaID) 
    if err != nil {
        pmlog.LogWarning(err)
        //return result, err
    }

    fileId, err := PutFile(mediaURL, id, pg.GetJWTToken(), filename, presentName)
    if err != nil {
        return result, err
    }

    result = fileId
    return result, err 
}

func PutFile(baseURL string, id pgschema.UUID, jwt string, filename string, presentName string) (pgschema.UUID, error) {

    var err error
    var result pgschema.UUID
    
    url := fmt.Sprintf("%s/upload/%s/%s", baseURL, id, jwt)

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    contenType, err := DetectContentType(filename) // default: "application/octet-stream"
	if err != nil {
		return result, err
	}
    
    header := make(textproto.MIMEHeader)
    contentDisposition := fmt.Sprintf(`form-data; name="uploaded_file"; filename="%s"`, filepath.Base(filename))
	header.Set("Content-Disposition", contentDisposition)
	header.Set("Content-Type", contenType)

	partWriter, err := writer.CreatePart(header)
	if err != nil {
		return result, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return result, err
	}
	defer file.Close()
    _, err = io.Copy(partWriter, file)

    err = writer.Close()
	if err != nil {
		return result, err
	}

    client := &http.Client{}
    resp, err := client.Post(url, writer.FormDataContentType(), body)
    if err != nil {
        return result, err
    }

    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return result, err
    }
    result = pgschema.UUID(respBody) 
    return result, err
}

func DetectContentType(filename string) (string, error) {
    var err error
    var result string

	file, err := os.Open(filename)
	if err != nil {
		
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return result, err
	}
	result = http.DetectContentType(buffer)
	return result, nil
}
//EOF
