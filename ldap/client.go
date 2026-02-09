package ldap

import (
	"crypto/tls"
	"fmt"
	"log"
	"strconv"

	"github.com/go-ldap/ldap/v3"

	"nsscache-http/config"
	"nsscache-http/models"
)

// Client wraps an LDAP connection
type Client struct {
	cfg  *config.LDAPConfig
	conn *ldap.Conn
}

// NewClient creates a new LDAP client
func NewClient(cfg *config.LDAPConfig) *Client {
	return &Client{cfg: cfg}
}

// Connect establishes a connection to the LDAP server
func (c *Client) Connect() error {
	var conn *ldap.Conn
	var err error

	address := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)

	if c.cfg.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.cfg.SkipVerify,
			ServerName:         c.cfg.Host,
		}
		conn, err = ldap.DialTLS("tcp", address, tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	c.conn = conn
	return nil
}

// Bind authenticates with the LDAP server
func (c *Client) Bind() error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.Bind(c.cfg.BindDN, c.cfg.BindPassword)
}

// Close closes the LDAP connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// FetchUsers queries LDAP for posixAccount entries
func (c *Client) FetchUsers() ([]models.User, error) {
	searchRequest := ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.cfg.UserFilter,
		[]string{"uid", "uidNumber", "gidNumber", "cn", "gecos", "homeDirectory", "loginShell"},
		nil,
	)

	result, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	users := make([]models.User, 0, len(result.Entries))
	for _, entry := range result.Entries {
		name := entry.GetAttributeValue("uid")

		uid, err := strconv.Atoi(entry.GetAttributeValue("uidNumber"))
		if err != nil || uid == 0 {
			log.Printf("skipping user %q: invalid or zero uidNumber", name)
			continue
		}
		gid, err := strconv.Atoi(entry.GetAttributeValue("gidNumber"))
		if err != nil || gid == 0 {
			log.Printf("skipping user %q: invalid or zero gidNumber", name)
			continue
		}

		// Use gecos if available, fall back to cn
		gecos := entry.GetAttributeValue("gecos")
		if gecos == "" {
			gecos = entry.GetAttributeValue("cn")
		}

		users = append(users, models.User{
			Name:   name,
			Passwd: "x",
			UID:    uid,
			GID:    gid,
			GECOS:  gecos,
			Dir:    entry.GetAttributeValue("homeDirectory"),
			Shell:  entry.GetAttributeValue("loginShell"),
		})
	}

	return users, nil
}

// FetchGroups queries LDAP for posixGroup entries
func (c *Client) FetchGroups() ([]models.Group, error) {
	searchRequest := ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.cfg.GroupFilter,
		[]string{"cn", "gidNumber", "memberUid", "member"},
		nil,
	)

	result, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("group search failed: %w", err)
	}

	groups := make([]models.Group, 0, len(result.Entries))
	for _, entry := range result.Entries {
		name := entry.GetAttributeValue("cn")

		gid, err := strconv.Atoi(entry.GetAttributeValue("gidNumber"))
		if err != nil || gid == 0 {
			log.Printf("skipping group %q: invalid or zero gidNumber", name)
			continue
		}

		// Collect members from both memberUid (traditional POSIX) and member (FreeIPA/AD style DNs)
		members := entry.GetAttributeValues("memberUid")
		memberDNs := entry.GetAttributeValues("member")

		// Parse uid from member DNs (e.g., "uid=sasha,cn=users,cn=accounts,dc=example,dc=com" -> "sasha")
		for _, dn := range memberDNs {
			if uid := extractUIDFromDN(dn); uid != "" {
				members = append(members, uid)
			}
		}

		groups = append(groups, models.Group{
			Name:    name,
			Passwd:  "x",
			GID:     gid,
			Members: members,
		})
	}

	return groups, nil
}

// extractUIDFromDN extracts the uid from a DN like "uid=sasha,cn=users,..."
func extractUIDFromDN(dn string) string {
	// Parse the DN to find the uid component
	parsedDN, err := ldap.ParseDN(dn)
	if err != nil {
		return ""
	}
	for _, rdn := range parsedDN.RDNs {
		for _, attr := range rdn.Attributes {
			if attr.Type == "uid" {
				return attr.Value
			}
		}
	}
	return ""
}

