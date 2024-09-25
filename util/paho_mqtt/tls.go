package paho_mqtt

import (
	"crypto/tls"
	"errors"
	"time"
)

func checkPeerCertValidityBounds(cs tls.ConnectionState) error {
	currenTime := time.Now().UTC()
	for _, peerCert := range cs.PeerCertificates {
		if currenTime.Before(peerCert.NotBefore) || currenTime.After(peerCert.NotAfter) {
			return errors.New("certificate expired or not valid yet")
		}
	}
	return nil
}
