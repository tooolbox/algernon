package engine

import (
	"github.com/mholt/certmagic"
	log "github.com/sirupsen/logrus"
	"github.com/xyproto/algernon/utils"
	"net/http"
	"path/filepath"
)

// GetDomainDirectories returns a slice of all directory names that qualifies
// for serving contents for specific domains, when using the --domain flag.
// This function has no caching and is intented to be called only when configuring
// Let's Encrypt with certmagic.
func (ac *Config) GetDomainDirectories() []string {
	if !ac.serverAddDomain {
		log.Warn("Looking for domain names without --domain being specified")
	}
	domains := []string{}
	if ac.fs.IsDir(ac.serverDirOrFilename) {
		for _, fileOrDirectory := range utils.GetFilenames(ac.serverDirOrFilename) {
			dirname := filepath.Join(ac.serverDirOrFilename, fileOrDirectory)
			if ac.fs.IsDir(dirname) && certmagic.HostQualifies(fileOrDirectory) {
				// This directory qualifies for holding the contents for a domain
				domains = append(domains, fileOrDirectory)
			}
		}
	}
	return domains
}

// GetDomain returns the domain of a request (up to ":", if any)
func GetDomain(req *http.Request) string {
	for i, r := range req.Host {
		if r == ':' {
			return req.Host[:i]
		}
	}
	return req.Host
}

// Save the served domain to the slice of domains, which can be used with Let's Encrypt
func (ac *Config) CollectDomain(domain string) {
	if domain == "" {
		return
	}
	found := false
	ac.domainMut.RLock()
	for _, existingDomain := range ac.domains {
		if domain == existingDomain {
			found = true
			break
		}
	}
	ac.domainMut.RUnlock()
	if !found {
		ac.domainMut.Lock()
		if ac.domains == nil {
			ac.domains = []string{domain}
		} else {
			ac.domains = append(ac.domains, domain)
		}
		ac.domainMut.Unlock()
	}
}

// Return a slice of the currently accessed domains
func (ac *Config) Domains() []string {
	// Lock for reading
	ac.domainMut.RLock()
	// Create a copy
	domainSliceCopy := make([]string, len(ac.domains))
	copy(domainSliceCopy, ac.domains)
	// Unlock and return
	defer ac.domainMut.RUnlock()
	return domainSliceCopy
}
