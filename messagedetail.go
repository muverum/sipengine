package sipengine

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
)

//MessageDetail is the default implementation of the CDR type data structure that will be placed
//Into the message, but can be replaced with anything that meets the requirements. It would make
//Most sense to embed this struct into another that overrides the interface methods to achieve
//both a uniform data structure (plus additions) as well as desired behaviors for CDR generation
//and transmission
type MessageDetail struct {
	CallID string `json:"call_id"`
	From string `json:"from"`
	To string `json:"to"`
	Start time.Time `json:"start"`
	End time.Time `json:"end"`
	billingSystem struct {
		Address string
	}
}

//GetCDR is the MessageDetail variant that simply prints the struct out in pretty JSON form
func (m MessageDetail) GetCDR() (io.Reader, error) {
	jsonbytes, err := json.MarshalIndent(m, "", "    ")
	return bytes.NewReader(jsonbytes), err
}

//SendCDR for the MessageDetail type is a basic implementation of the interface that posts the json content
//Of the message to a remote source
func (m MessageDetail) SendCDR() error {
	if cdr, err := m.GetCDR(); err == nil {
		resp, err := http.Post(m.billingSystem.Address, "application/json", cdr)

		if err != nil {
			return errors.Wrap(err, "error occurred during billing request")
		}

		if resp.StatusCode != 200 {
			return errors.Wrap(err, "response code not ok from billing system")
		}
	}

	return nil
}

