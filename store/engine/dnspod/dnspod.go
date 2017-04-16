package dnspod

import "github.com/admpub/ddns/store"

var _ store.Storer = &Dnspod{}

type Dnspod struct {
	Domain string
	Key    string
	Config Config
}

func (self *Dnspod) GetHost(name string) *store.Host {
	host := store.Host{Hostname: name}
	self.HostToken(&host)
	domainID := self.getDomain(self.Domain)
	if domainID == -1 {
		return &host
	}
	subDomainID, ip := self.getSubDomain(domainID, name)
	if len(subDomainID) == 0 || len(ip) == 0 {
		return &host
	}
	host.IP = ip
	host.GenerateAndSetToken(self.Key)
	return &host
}

func (self *Dnspod) SaveHost(host *store.Host) {
	domainID := self.getDomain(self.Domain)
	if domainID == -1 {
		return
	}
	subDomainID, ip := self.getSubDomain(domainID, host.Hostname)
	if len(subDomainID) == 0 || len(ip) == 0 {
		return
	}
	self.updateIP(domainID, subDomainID, host.Hostname, host.IP)
}

func (self *Dnspod) HostExist(name string) bool {
	domainID := self.getDomain(self.Domain)
	if domainID == -1 {
		return false
	}
	subDomainID, ip := self.getSubDomain(domainID, name)
	if len(subDomainID) == 0 || len(ip) == 0 {
		return false
	}
	return true
}

func (self *Dnspod) HostToken(host *store.Host) string {
	host.GenerateAndSetToken(self.Key)
	return host.Token
}

func (self *Dnspod) Close() error {
	return nil
}
