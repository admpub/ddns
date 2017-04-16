package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net"
	"regexp"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/defaults"
	"github.com/webx-top/echo/engine/standard"
)

func RunWebService(conn *RedisConnection) {
	t := BuildTemplate()

	defaults.Get("/", func(c echo.Context) error {
		buf := new(bytes.Buffer)
		err := t.Execute(buf, echo.H{"domain": DdnsDomain})
		if err != nil {
			return err
		}
		return c.HTML(buf.String())
	})

	defaults.Get("/available/:hostname", func(c echo.Context) error {
		hostname, valid := ValidHostname(c.Param("hostname"))

		return c.JSON(echo.H{
			"available": valid && !conn.HostExist(hostname),
		})
	})

	defaults.Get("/new/:hostname", func(c echo.Context) error {
		hostname, valid := ValidHostname(c.Param("hostname"))

		if !valid {
			return c.JSON(echo.H{"error": "This hostname is not valid"}, 404)
		}

		if conn.HostExist(hostname) {
			return c.JSON(echo.H{
				"error": "This hostname has already been registered.",
			}, 403)
		}

		host := &Host{Hostname: hostname, IP: "127.0.0.1"}
		host.GenerateAndSetToken()

		conn.SaveHost(host)

		return c.JSON(echo.H{
			"hostname":    host.Hostname,
			"token":       host.Token,
			"update_link": fmt.Sprintf("/update/%s/%s", host.Hostname, host.Token),
		})
	})

	defaults.Get("/update/:hostname/:token", func(c echo.Context) error {
		hostname, valid := ValidHostname(c.Param("hostname"))
		token := c.Param("token")

		if !valid {
			return c.JSON(echo.H{"error": "This hostname is not valid"}, 404)
		}

		if !conn.HostExist(hostname) {
			return c.JSON(echo.H{
				"error": "This hostname has not been registered or is expired.",
			}, 404)
		}

		host := conn.GetHost(hostname)

		if host.Token != token {
			return c.JSON(echo.H{
				"error": "You have supplied the wrong token to manipulate this host",
			}, 403)
		}

		ip, err := GetRemoteAddr(c)
		if err != nil {
			return c.JSON(echo.H{
				"error": "Your sender IP address is not in the right format",
			}, 400)
		}

		host.IP = ip
		conn.SaveHost(host)

		return c.JSON(echo.H{
			"current_ip": ip,
			"status":     "Successfuly updated",
		})
	})

	defaults.Run(standard.New(DdnsWebListenSocket))
}

// GetRemoteAddr Get the Remote Address of the client. At First we try to get the
// X-Forwarded-For Header which holds the IP if we are behind a proxy,
// otherwise the RemoteAddr is used
func GetRemoteAddr(c echo.Context) (string, error) {
	headerData := c.Header("X-Forwarded-For")

	if len(headerData) > 0 {
		return headerData, nil
	}
	ip, _, err := net.SplitHostPort(c.Request().RemoteAddress())
	return ip, err
}

//BuildTemplate Get index template from bindata
func BuildTemplate() *template.Template {
	html, err := template.New("index.html").Parse(indexTemplate)
	HandleErr(err)

	return html
}

var alphaNumeric = regexp.MustCompile("^[a-z0-9]{1,32}$")

func ValidHostname(host string) (string, bool) {
	valid := alphaNumeric.Match([]byte(host))
	return host, valid
}
