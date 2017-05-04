// Package makross is a high productive and modular web framework in Golang.

// Package content provides content negotiation handlers for the makross.
package content

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/insionng/makross"
)

// MIME types
const (
	JSON = makross.MIME_JSON
	XML  = makross.MIME_XML
	XML2 = makross.MIME_XML2
	HTML = makross.MIME_HTML
)

// DataWriters lists all supported content types and the corresponding data writers.
// By default, JSON, XML, and HTML are supported. You may modify this variable before calling TypeNegotiator
// to customize supported data writers.
var DataWriters = map[string]makross.DataWriter{
	JSON: &JSONDataWriter{},
	XML:  &XMLDataWriter{},
	XML2: &XMLDataWriter{},
	HTML: &HTMLDataWriter{},
}

// TypeNegotiator returns a content type negotiation handler.
//
// The method takes a list of response MIME types that are supported by the application.
// The negotiator will determine the best response MIME type to use by checking the "Accept" HTTP header.
// If no match is found, the first MIME type will be used.
//
// The negotiator will set the "Content-Type" response header as the chosen MIME type. It will call makross.Context.SetDataWriter()
// to set the appropriate data writer that can write data in the negotiated format.
//
// If you do not specify any supported MIME types, the negotiator will use "text/html" as the response MIME type.
func TypeNegotiator(formats ...string) makross.Handler {
	if len(formats) == 0 {
		formats = []string{HTML}
	}
	for _, format := range formats {
		if _, ok := DataWriters[format]; !ok {
			panic(format + " is not supported")
		}
	}

	return func(c *makross.Context) error {
		format := NegotiateContentType(c.Request, formats, formats[0])
		DataWriters[format].SetHeader(c.Response)
		c.SetDataWriter(DataWriters[format])
		return nil
	}
}

// JSONDataWriter sets the "Content-Type" response header as "application/json" and writes the given data in JSON format to the response.
type JSONDataWriter struct{}

func (w *JSONDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
}

func (w *JSONDataWriter) Write(res http.ResponseWriter, data interface{}) (err error) {
	enc := json.NewEncoder(res)
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

// XMLDataWriter sets the "Content-Type" response header as "application/xml; charset=UTF-8" and writes the given data in XML format to the response.
type XMLDataWriter struct{}

func (w *XMLDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/xml; charset=UTF-8")
}

func (w *XMLDataWriter) Write(res http.ResponseWriter, data interface{}) (err error) {
	var bytes []byte
	if bytes, err = xml.Marshal(data); err != nil {
		return
	}
	_, err = res.Write(bytes)
	return
}

// HTMLDataWriter sets the "Content-Type" response header as "text/html; charset=UTF-8" and calls makross.DefaultDataWriter to write the given data to the response.
type HTMLDataWriter struct{}

func (w *HTMLDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "text/html; charset=UTF-8")
}

func (w *HTMLDataWriter) Write(res http.ResponseWriter, data interface{}) error {
	return makross.DefaultDataWriter.Write(res, data)
}
