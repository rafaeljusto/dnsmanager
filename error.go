package dnsmanager

import (
	"fmt"
	"strconv"
)

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
	switch g.Code {
	case GenericErrorCodeInvalidFQDN:
		return "dnsmanager: invalid FQDN"
	case GenericErrorCodeBlockedTLD:
		return "dnsmanager: blocked TLD"
	}

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

	// DNSErrorCodeInvalidIPv4Glue is used when the informed IPv4 is invalid
	DNSErrorCodeInvalidIPv4Glue
)

type DNSErrorCode int

func (d DNSErrorCode) String() string {
	switch d {
	case DNSErrorMissingGlue:
		return "missing glue record"
	case DNSErrorCodeQueryFailed:
		return "query failed"
	case DNSErrorCodeNotAuthoritative:
		return "not authoritative"
	case DNSErrorCodeInvalidFQDN:
		return "invalid FQDN"
	case DNSErrorCodeInvalidIPv4Glue:
		return "invalid IPV4 glue record"
	}

	return strconv.Itoa(int(d))
}

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
	msg := fmt.Sprintf("dnsmanager: dns check error (%s) in nameserver index %d",
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

func (d DNSSECErrorCode) String() string {
	switch d {
	case DNSSECErrorCodeAlgorithmDontMatch:
		return "algorithm don't match"
	case DNSSECErrorCodeDigestDontMatch:
		return "digest don't match"
	case DNSSECErrorCodeDNSKEYNotSEP:
		return "DNSKEY is not SEP"
	case DNSSECErrorCodeDSNotFound:
		return "DS not found"
	}

	return strconv.Itoa(int(d))
}

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
	return fmt.Sprintf("dnsmanager: dnssec check error (%s) in nameserver index %d and ds index %d",
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
		msg += fmt.Sprintf("error %d: %s\n", i, e.Errors[i])
	}
	return msg
}
