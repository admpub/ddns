package dnspod

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/admpub/spider/common/simplejson"
)

func (self *Dnspod) generateHeader(content url.Values) url.Values {
	header := url.Values{}
	if self.Config.LoginToken != "" {
		header.Add("login_token", self.Config.LoginToken)
	} else {
		header.Add("login_email", self.Config.Email)
		header.Add("login_password", self.Config.Password)
	}
	header.Add("format", "json")
	header.Add("lang", "en")
	header.Add("error_on_empty", "no")

	if content != nil {
		for k := range content {
			header.Add(k, content.Get(k))
		}
	}

	return header
}

func (self *Dnspod) apiVersion() {
	self.postData("/Info.Version", nil)
}

func (self *Dnspod) getDomain(name string) int64 {

	var ret int64
	values := url.Values{}
	values.Add("type", "all")
	values.Add("offset", "0")
	values.Add("length", "20")

	response, err := self.postData("/Domain.List", values)

	if err != nil {
		log.Println("Failed to get domain list...")
		return -1
	}

	sjson, parseErr := simplejson.NewJson([]byte(response))

	if parseErr != nil {
		log.Println(parseErr)
		return -1
	}

	if sjson.Get("status").Get("code").MustString() == "1" {
		domains, _ := sjson.Get("domains").Array()

		for _, d := range domains {
			m := d.(map[string]interface{})
			if m["name"] == name {
				id := m["id"]

				switch t := id.(type) {
				case json.Number:
					ret, _ = t.Int64()
				}

				break
			}
		}
		if len(domains) == 0 {
			log.Println("domains slice is empty.")
		}
	} else {
		log.Println("get_domain:status code:", sjson.Get("status").Get("code").MustString())
	}

	return ret
}

func (self *Dnspod) getSubDomain(domainID int64, name string) (string, string) {
	log.Println("debug:", domainID, name)
	var ret, ip string
	value := url.Values{}
	value.Add("domain_id", strconv.FormatInt(domainID, 10))
	value.Add("offset", "0")
	value.Add("length", "1")
	value.Add("sub_domain", name)

	response, err := self.postData("/Record.List", value)

	if err != nil {
		log.Println("Failed to get domain list")
		return "", ""
	}

	sjson, parseErr := simplejson.NewJson([]byte(response))

	if parseErr != nil {
		log.Println(parseErr)
		return "", ""
	}

	if sjson.Get("status").Get("code").MustString() == "1" {
		records, _ := sjson.Get("records").Array()

		for _, d := range records {
			m := d.(map[string]interface{})
			if m["name"] == name {
				ret = m["id"].(string)
				ip = m["value"].(string)
				break
			}
		}
		if len(records) == 0 {
			log.Println("records slice is empty.")
		}
	} else {
		log.Println("get_subdomain:status code:", sjson.Get("status").Get("code").MustString())
	}

	return ret, ip
}

func (self *Dnspod) updateIP(domainID int64, subDomainID string, subDomainName string, ip string) {
	value := url.Values{}
	value.Add("domain_id", strconv.FormatInt(domainID, 10))
	value.Add("record_id", subDomainID)
	value.Add("sub_domain", subDomainName)
	value.Add("record_type", "A")
	value.Add("record_line", "默认")
	value.Add("value", ip)

	response, err := self.postData("/Record.Modify", value)

	if err != nil {
		log.Println("Failed to update record to new IP!")
		log.Println(err)
		return
	}

	sjson, parseErr := simplejson.NewJson([]byte(response))

	if parseErr != nil {
		log.Println(parseErr)
		return
	}

	if sjson.Get("status").Get("code").MustString() == "1" {
		log.Println("New IP updated!")
	}

}

func (self *Dnspod) postData(url string, content url.Values) (string, error) {
	client := &http.Client{}
	values := self.generateHeader(content)
	req, _ := http.NewRequest("POST", "https://dnsapi.cn"+url, strings.NewReader(values.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fmt.Sprintf("GoDNS/0.1 (%s)", self.Config.Email))

	response, err := client.Do(req)

	if err != nil {
		log.Println("Post failed...")
		log.Println(err)
		return "", err
	}

	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)

	return string(resp), nil
}
