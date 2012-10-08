package ossapi

import (
    "encoding/xml"
    "io"
)



type Error struct {
    Code string
    Message string
    Request string
    HostId string
}


func ParseXmlError(content *io.Reader) (*Error, error) {
    xmlError := &Error{}
    decoder := xml.NewDecoder(content)
    err := decoder.Decode(xmlError)
    if err != nil {
        return (nil, err)
    }
    return (xmlError, err)
}