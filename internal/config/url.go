package config

import (
	"fmt"
	"net/url"
	"strings"
)

type DatabaseConnection struct {
	URL      string
	Password string
	Secrets  []string
}

// ParseDatabaseURL preserves PostgreSQL host syntax while separating passwords.
func ParseDatabaseURL(raw string) (DatabaseConnection, error) {
	invalid := fmt.Errorf("db-url must be a valid PostgreSQL URL using postgres:// or postgresql://")
	scheme, rest, ok := strings.Cut(raw, "://")
	if !ok || (scheme != "postgres" && scheme != "postgresql") {
		return DatabaseConnection{}, invalid
	}
	end := strings.IndexAny(rest, "/?#")
	if end < 0 {
		end = len(rest)
	}
	authority, suffix := rest[:end], rest[end:]
	userinfo, host := "", authority
	if at := strings.LastIndex(authority, "@"); at >= 0 {
		userinfo, host = authority[:at+1], authority[at+1:]
	}
	if _, err := url.PathUnescape(host); err != nil || strings.ContainsAny(host, " \t\r\n") {
		return DatabaseConnection{}, invalid
	}
	// net/url rejects libpq's encoded socket hosts and multi-host authorities.
	// Parse the userinfo, path and query against a placeholder; retain the original host.
	u, err := url.Parse(scheme + "://" + userinfo + "localhost" + suffix)
	if err != nil || u.Fragment != "" {
		return DatabaseConnection{}, invalid
	}
	connection := DatabaseConnection{Secrets: []string{raw}}
	if u.User != nil {
		connection.Password, _ = u.User.Password()
		connection.Secrets = append(connection.Secrets, connection.Password)
		userinfo = url.User(u.User.Username()).String() + "@"
	}
	// libpq uses percent decoding rather than form decoding. Preserve query bytes
	// and ordering, removing only password parameters from the command arguments.
	var query []string
	for _, part := range strings.Split(u.RawQuery, "&") {
		key, value, _ := strings.Cut(part, "=")
		key, err = url.PathUnescape(key)
		if err != nil {
			return DatabaseConnection{}, invalid
		}
		value, err = url.PathUnescape(value)
		if err != nil {
			return DatabaseConnection{}, invalid
		}
		if key == "password" {
			connection.Password = value
			connection.Secrets = append(connection.Secrets, value)
			continue
		}
		query = append(query, part)
	}
	connection.URL = scheme + "://" + userinfo + host + u.EscapedPath()
	if encoded := strings.Join(query, "&"); encoded != "" {
		connection.URL += "?" + encoded
	}
	return connection, nil
}
