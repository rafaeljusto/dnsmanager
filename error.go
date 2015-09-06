package dnsmanager

import "fmt"

const (
	// GenericErrorCodeInvalidFQDN is returned when a domain name has an
	// invalid format
	GenericErrorCodeInvalidFQDN GenericErrorCode = iota

	// GenericErrorCodeBlockedTLD alerts that a domain with a forbidden TLD is
	// trying to be registered in the system
	GenericErrorCodeBlockedTLD
)

type GenericErrorCode int

type GenericError struct {
	Code GenericErrorCode
}

func NewGenericError(code GenericErrorCode) GenericError {
	return GenericError{
		Code: code,
	}
}

func (g GenericError) Error() string {
	return fmt.Sprintf("dnsmanager: generic error (%d)", g.Code)
}

const (
	// DNSErrorMissingGlue nameserver is bellow the domain name, so we need a
	// glue record to query
	DNSErrorMissingGlue DNSErrorCode = iota

	// DNSErrorCodeQueryFailed occurs while checking a nameserver for DNS or
	// DNSSEC checks and something happens in the connection
	DNSErrorCodeQueryFailed

	// DNSErrorCodeNotAuthoritative occurs when checking a domain authority in
	// a nameserver, and there's no AA flag or response
	DNSErrorCodeNotAuthoritative

	// DNSErrorCodeInvalidFQDN is returned when a nameserver name has an
	// invalid format
	DNSErrorCodeInvalidFQDN
)

type DNSErrorCode int

type DNSError struct {
	Code    DNSErrorCode
	Index   int
	Details error
}

func NewDNSError(code DNSErrorCode, index int, err error) DNSError {
	return DNSError{
		Code:    code,
		Index:   index,
		Details: err,
	}
}

func (d DNSError) Error() string {
	msg := fmt.Sprintf("dnsmanager: dns check error (%d) in nameserver index %d",
		d.Code, d.Index)

	if d.Details != nil {
		msg += fmt.Sprintf(". details: %s", d.Details)
	}

	return msg
}

const (
	// DNSSECErrorCodeAlgorithmDontMatch the algorithm used in the DS doesn't
	// match with what we found in the DNSKEY response
	DNSSECErrorCodeAlgorithmDontMatch DNSSECErrorCode = iota

	// DNSSECErrorCodeDigestDontMatch the digest doesn't match from what we
	// calculate from the DNSKEY response
	DNSSECErrorCodeDigestDontMatch

	// DNSSECErrorCodeDNSKEYNotSEP the public key used as DS is not a secure
	// entry point of the zone
	DNSSECErrorCodeDNSKEYNotSEP

	// DNSSECErrorCodeDSNotFound the DS did not match with any DNSKEY keytag
	// of the keyset
	DNSSECErrorCodeDSNotFound
)

type DNSSECErrorCode int

type DNSSECError struct {
	Code    DNSSECErrorCode
	NSIndex int
	DSIndex int
}

func NewDNSSECError(code DNSSECErrorCode, nsIndex, dsIndex int) DNSSECError {
	return DNSSECError{
		Code:    code,
		NSIndex: nsIndex,
		DSIndex: dsIndex,
	}
}

func (d DNSSECError) Error() string {
	return fmt.Sprintf("dnsmanager: dnssec check error (%d) in nameserver index %d and ds index %d",
		d.Code, d.NSIndex, d.DSIndex)
}

type ErrorBox struct {
	Errors []error
}

func (e *ErrorBox) Append(err error) *ErrorBox {
	e.Errors = append(e.Errors, err)
	return e
}

func (e ErrorBox) Unpack() error {
	if len(e.Errors) > 0 {
		return e
	}

	return nil
}

func (e ErrorBox) Error() string {
	var msg string
	for i := range e.Errors {
		msg += fmt.Sprintf("error %d: %v\n", e.Errors[i])
	}
	return msg
}
