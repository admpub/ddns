package store

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

type Host struct {
	Hostname string `redis:"-"`
	IP       string `redis:"ip"`
	Token    string `redis:"token"`
}

func (self *Host) GenerateAndSetToken() {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	hash.Write([]byte(self.Hostname))

	self.Token = fmt.Sprintf("%x", hash.Sum(nil))
}

//IsIPv4 Returns true when this host has a IPv4 Address and false if IPv6
func (self *Host) IsIPv4() bool {
	if strings.Contains(self.IP, ".") {
		return true
	}

	return false
}
