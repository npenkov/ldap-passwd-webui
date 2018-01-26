package app

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"gopkg.in/ldap.v2"
)

// SecurityProtocol protocol type
type SecurityProtocol int

// Note: new type must be added at the end of list to maintain compatibility.
const (
	SecurityProtocolUnencrypted SecurityProtocol = iota
	SecurityProtocolLDAPS
	SecurityProtocolStartTLS
)

// LDAPClient Basic LDAP authentication service
type LDAPClient struct {
	Name             string // canonical name (ie. corporate.ad)
	Host             string // LDAP host
	Port             int    // port number
	SecurityProtocol SecurityProtocol
	SkipVerify       bool
	UserBase         string // Base search path for users
	UserDN           string // Template for the DN of the user for simple auth
	Enabled          bool   // if this LDAPClient is disabled
}

func bindUser(l *ldap.Conn, userDN, passwd string) error {
	log.Printf("\nBinding with userDN: %s", userDN)
	err := l.Bind(userDN, passwd)
	if err != nil {
		log.Printf("\nLDAP auth. failed for %s, reason: %v", userDN, err)
		return err
	}
	log.Printf("\nBound successfully with userDN: %s", userDN)
	return err
}

func (ls *LDAPClient) sanitizedUserDN(username string) (string, bool) {
	// See http://tools.ietf.org/search/rfc4514: "special characters"
	badCharacters := "\x00()*\\,='\"#+;<>"
	if strings.ContainsAny(username, badCharacters) {
		log.Printf("\n'%s' contains invalid DN characters. Aborting.", username)
		return "", false
	}

	return fmt.Sprintf(ls.UserDN, username), true
}

func dial(ls *LDAPClient) (*ldap.Conn, error) {
	log.Printf("\nDialing LDAP with security protocol (%v) without verifying: %v", ls.SecurityProtocol, ls.SkipVerify)

	tlsCfg := &tls.Config{
		ServerName:         ls.Host,
		InsecureSkipVerify: ls.SkipVerify,
	}
	if ls.SecurityProtocol == SecurityProtocolLDAPS {
		return ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ls.Host, ls.Port), tlsCfg)
	}

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ls.Host, ls.Port))
	if err != nil {
		return nil, fmt.Errorf("Dial: %v", err)
	}

	if ls.SecurityProtocol == SecurityProtocolStartTLS {
		if err = conn.StartTLS(tlsCfg); err != nil {
			conn.Close()
			return nil, fmt.Errorf("StartTLS: %v", err)
		}
	}

	return conn, nil
}

// ModifyPassword : modify user's password
func (ls *LDAPClient) ModifyPassword(name, passwd, newPassword string) error {
	if len(passwd) == 0 {
		return fmt.Errorf("Auth. failed for %s, password cannot be empty", name)
	}
	l, err := dial(ls)
	if err != nil {
		ls.Enabled = false
		return fmt.Errorf("LDAP Connect error, %s:%v", ls.Host, err)
	}
	defer l.Close()

	var userDN string
	log.Printf("\nLDAP will bind directly via UserDN template: %s", ls.UserDN)

	var ok bool
	userDN, ok = ls.sanitizedUserDN(name)
	if !ok {
		return fmt.Errorf("Error sanitizing name %s", name)
	}
	bindUser(l, userDN, passwd)

	log.Printf("\nLDAP will execute password change on: %s", userDN)
	req := ldap.NewPasswordModifyRequest(userDN, passwd, newPassword)
	_, err = l.PasswordModify(req)

	return err
}

// NewLDAPClient : Creates new LDAPClient capable of binding and changing passwords
func NewLDAPClient() *LDAPClient {

	securityProtocol := SecurityProtocolUnencrypted
	if envBool("LPW_ENCRYPTED", true) {
		securityProtocol = SecurityProtocolLDAPS
		if envBool("LPW_START_TLS", false) {
			securityProtocol = SecurityProtocolStartTLS
		}
	}

	return &LDAPClient{
		Host:             envStr("LPW_HOST", ""),
		Port:             envInt("LPW_PORT", 636), // 389
		SecurityProtocol: securityProtocol,
		SkipVerify:       envBool("LPW_SSL_SKIP_VERIFY", false),
		UserDN:           envStr("LPW_USER_DN", "uid=%s,ou=people,dc=example,dc=org"),
		UserBase:         envStr("LPW_USER_BASE", "ou=people,dc=example,dc=org"),
	}
}
