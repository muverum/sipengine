package sipengine

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"gortc.io/sdp"
)

//Message is the raw definition of the SIP structure (including SIP and SDP fields)
type Message struct {
	layers.SIP
	ctx context.Context
	//Raw is retained for potential audit purposes
	raw                []byte
	SessionDescription *sdp.Message
	//Detail is the CDR reporter used for collecting and Reporting the data
	//In most cases it will also be the sender.
	Detail CallDetailSender
}

func (m Message) GetOriginalBytes() []byte {
	return m.raw
}

func NewMessage(data io.Reader, ctx context.Context) (*Message, error) {
	s := new(Message)
	s.Detail = &MessageDetail{}
	s.ctx = ctx
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, data)

	if err != nil {
		return s, err
	}

	s.raw = buf.Bytes()
	s.Headers = make(map[string][]string)

	//Process SIP Data
	sipContent, sdpContent, err := splitMessage(bytes.NewReader(s.raw))

	if err != nil {
		return nil, err
	}

	err = s.DecodeFromBytes(sipContent, gopacket.NilDecodeFeedback)

	if err != nil {
		return s, err
	}

	//Process SDP Data
	s.SessionDescription, err = sdp.Decode(sdpContent)

	if err != nil {
		return s, err
	}

	return s, err
}

func splitMessage(input io.Reader) (sip []byte, rtp []byte, err error) {

	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	var sipLines []string
	var rtpLines []string

	dividerReached := false

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		if !dividerReached {
			sipLines = append(sipLines, scanner.Text())
		} else {
			rtpLines = append(rtpLines, scanner.Text())
		}

		if strings.Contains(strings.ToLower(scanner.Text()), "content-length") {
			dividerReached = true
		}
	}

	return []byte(strings.Join(sipLines, "\n")), []byte(strings.Join(rtpLines, "\n")), nil

}
