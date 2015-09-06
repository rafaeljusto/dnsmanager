package dnsmanager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

const (
	nsupdateCmd = "nsupdate"
)

func nsupdate(domain Domain, dnsPort int) error {
	cmdFile, err := ioutil.TempFile("/tmp/", "dnsmanager-")
	if err != nil {
		return err
	}
	defer os.Remove(cmdFile.Name())

	// remove all nameservers before adding the new ones
	fmt.Fprintf(cmdFile, "update delete %s NS\n", domain.FQDN)

	for _, ns := range domain.Nameservers {
		fmt.Fprintf(cmdFile, "update add %s 172800 NS %s\n", domain.FQDN, ns.Name)
		fmt.Fprintf(cmdFile, "update delete %s A\n", ns.Name)

		for _, glue := range ns.Glues {
			if glue.To4() != nil {
				fmt.Fprintf(cmdFile, "update add %s 172800 A %s\n", ns.Name, glue.String())
			} else {
				fmt.Fprintf(cmdFile, "update add %s 172800 AAAA %s\n", ns.Name, glue.String())
			}
		}
	}

	// remove all ds records before adding the new ones
	fmt.Fprintf(cmdFile, "update delete %s DS\n", domain.FQDN)

	for _, ds := range domain.DSSet {
		fmt.Fprintf(cmdFile, "update add %s 172800 DS %d %d %d %s\n",
			domain.FQDN, ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest)
	}

	fmt.Fprintln(cmdFile, "send")
	fmt.Fprint(cmdFile, "quit")
	cmdFile.Close()

	cmd := exec.Command(nsupdateCmd, "-l", "-p", strconv.Itoa(dnsPort), cmdFile.Name())

	var cmdErr bytes.Buffer
	cmd.Stderr = &cmdErr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error updating DNS: %s. %s", err.Error(), cmdErr.String())
	}

	return nil
}
