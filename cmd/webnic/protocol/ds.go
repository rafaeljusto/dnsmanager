package protocol

import (
	"strconv"

	"github.com/rafaeljusto/dnsmanager"
)

type DS struct {
	KeyTag     string `json:"keytag"`
	Algorithm  int    `json:"algorithm"`
	DigestType int    `json:"digestType"`
	Digest     string `json:"digest"`
}

func NewDS(ds dnsmanager.DS) DS {
	return DS{
		KeyTag:     strconv.FormatUint(uint64(ds.KeyTag), 10),
		Algorithm:  int(ds.Algorithm),
		DigestType: int(ds.DigestType),
		Digest:     ds.Digest,
	}
}

func (d DS) Convert() (ds dnsmanager.DS, err error) {
	keytag, err := strconv.ParseUint(d.KeyTag, 10, 16)
	if err != nil {
		return
	}

	ds.KeyTag = uint16(keytag)
	ds.Algorithm = uint8(d.Algorithm)
	ds.DigestType = uint8(d.DigestType)
	ds.Digest = d.Digest
	return
}
