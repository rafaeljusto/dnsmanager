package dnsmanager

type Nameserver struct {
	Name string
	Glue string
}

type DS struct {
	KeyTag     int
	Algorithm  int
	DigestType uint8
	Digest     string
}

type Domain struct {
	Name        string
	Nameservers []Nameserver
	DSs         []DS
}
