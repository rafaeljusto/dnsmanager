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
	fmt.Fprintf(cmdFile, "update delete %s NS\n", domain.Name)

	for _, ns := range domain.Nameservers {
		fmt.Fprintf(cmdFile, "update add %s 172800 NS %s\n", domain.Name, ns.Name)
		fmt.Fprintf(cmdFile, "update delete %s A\n", ns.Name)

		if len(ns.Glue) > 0 {
			fmt.Fprintf(cmdFile, "update add %s 172800 A %s\n", ns.Name, ns.Glue)
		}
	}

	// remove all ds records before adding the new ones
	fmt.Fprintf(cmdFile, "update delete %s DS\n", domain.Name)

	for _, ds := range domain.DSs {
		fmt.Fprintf(cmdFile, "update add %s 172800 DS %d %d %d %s\n",
			domain.Name, ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest)
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
