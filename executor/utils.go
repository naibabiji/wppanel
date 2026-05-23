package executor

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

func generatePassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func isValidDomain(domain string) bool {
	if len(domain) < 3 || len(domain) > 253 {
		return false
	}
	re := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	return re.MatchString(domain)
}

func buildServerNames(domain string, aliases []string) string {
	names := []string{domain}
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias != "" {
			names = append(names, alias)
		}
	}
	return strings.Join(names, " ")
}
