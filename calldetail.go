package sipengine

import "io"

//CallDetailReporter as an interface is used to generate a STRING format CDR from any user provided
//struct. It is important that it is a string (Whether line of a CSV, JSON, or TSV file) as that is
//How most clients expect to receive CDRs for billing
type CallDetailReporter interface {
	GetCDR() (io.Reader, error)
}

//CallDetailSender is an interface that defines the behavior on how CDRs should be submitted to
//Any system (even internal via channel) for billing / processing
type CallDetailSender interface {
	SendCDR() error
	GetCDR() (io.Reader, error)
}