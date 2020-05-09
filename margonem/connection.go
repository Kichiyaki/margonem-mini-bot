package margonem

import (
	"bot/colly/debug"
	_extensions "bot/colly/extensions"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gocolly/colly/v2/proxy"

	"github.com/gocolly/colly/v2"
)

const (
	appVersion       = "1.3.6"
	baseURL          = "http://www.margonem.pl"
	loginURL         = "https://new.margonem.pl/ajax/login"
	getPlayerDataURL = "http://www.margonem.pl/ajax/getplayerdata.php?app_version=%s"
)

var (
	_logrus = logrus.WithField("package", "margonem")
)

type Connection interface {
	Charlist() []*Character
	UseWholeStamina(charid string, mapid string) error
	UserID() string
	Maplist() (map[string]*Map, error)
}

type connection struct {
	username  string
	password  string
	charlist  []*Character
	proxy     string
	collector *colly.Collector
	headers   http.Header
	mutex     sync.Mutex
}

type Config struct {
	Username string
	Password string
	Proxy    string
}

func Connect(cfg *Config) (Connection, error) {
	entry := _logrus.WithField("cfg", cfg)
	entry.Debug("Connect called")
	c := &connection{
		username: cfg.Username,
		password: cfg.Password,
		proxy:    cfg.Proxy,
	}
	if err := c.init(); err != nil {
		entry.Debugf("Connect err: %s", err.Error())
		return nil, err
	}
	if err := c.login(); err != nil {
		entry.Debugf("Connect err: %s", err.Error())
		return nil, err
	}
	if err := c.getPlayerData(); err != nil {
		entry.Debugf("Connect err: %s", err.Error())
		return nil, err
	}
	entry.Debug("Connect finished")
	return c, nil
}

func (c *connection) Charlist() []*Character {
	return c.charlist
}

func (c *connection) UserID() string {
	id := ""
	for _, cookie := range c.collector.Cookies(baseURL) {
		if cookie.Name == "user_id" {
			id = cookie.Value
		}
	}
	return id
}

func (c *connection) UseWholeStamina(charid string, mapid string) error {
	charid = strings.ToLower(charid)
	entry := _logrus.WithField("charid", charid).WithField("mapid", mapid)
	entry.Debug("UseWholeStamina called")
	character, err := c.findCharacterByID(charid)
	if err != nil {
		entry.Debugf("UseWholeStamina err: %s", err.Error())
		return err
	}
	serverConn := &serverConnection{
		collector: c.collector.Clone(),
		headers:   c.headers.Clone(),
		charID:    character.ID,
		userID:    c.UserID(),
		server:    character.Server(),
	}
	if err := serverConn.init(); err != nil {
		entry.Debugf("UseWholeStamina err: %s", err.Error())
		return err
	}
	for serverConn.canSendAttack(mapid) {
		if err := serverConn.attack(mapid); err != nil {
			entry.Debugf("UseWholeStamina err: %s", err.Error())
			return err
		}
		if err := serverConn.heal(); err != nil {
			entry.Debugf("UseWholeStamina err: %s", err.Error())
			return err
		}
	}

	entry.Debug("UseWholeStamina finished")

	return nil
}

func (c *connection) Maplist() (map[string]*Map, error) {
	character := c.charlist[0]
	_logrus.Debug("Maplist called")
	serverConn := &serverConnection{
		collector: c.collector.Clone(),
		headers:   c.headers.Clone(),
		charID:    character.ID,
		userID:    c.UserID(),
		server:    character.Server(),
	}
	if err := serverConn.init(); err != nil {
		_logrus.Debugf("Maplist err: %s", err.Error())
		return nil, err
	}
	maps := serverConn.maps
	_logrus.Debug("Maplist finished")
	return maps, nil
}

func (c *connection) findCharacterByID(charid string) (*Character, error) {
	var character *Character
	err := fmt.Errorf("findCharacterByID: Cannot find a character with id %s", charid)
	splitted := strings.Split(charid, "#")
	if len(splitted) != 2 {
		return nil, err
	}
	for _, ch := range c.charlist {
		if ch.ID == splitted[1] && ch.Server() == splitted[0] {
			character = ch
		}
	}
	if character == nil {
		return nil, err
	}
	return character, nil
}

func (c *connection) init() error {
	c.collector = colly.NewCollector(
		colly.Debugger(&debug.LogrusDebugger{}),
		colly.Async(true),
		colly.AllowURLRevisit(),
	)
	c.collector.UserAgent = _extensions.RandomMobileUserAgent()
	if c.proxy != "" {
		rp, err := proxy.RoundRobinProxySwitcher(c.proxy)
		if err != nil {
			return fmt.Errorf("init: %s", err.Error())
		}
		c.collector.WithTransport(&http.Transport{
			Proxy: rp,
			DialContext: (&net.Dialer{
				DualStack: true,
			}).DialContext,
			MaxIdleConns: 100,
			TLSNextProto: nil,
		})
	}
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       500 * time.Millisecond,
	})

	c.headers = http.Header{}
	c.headers.Set("User-Agent", c.collector.UserAgent)
	c.headers.Set("X-Unity-Version", "5.6.7f1")
	c.headers.Set("Accept-Encoding", "gzip")
	c.headers.Set("Content-Type", "application/x-www-form-urlencoded")

	return nil
}

type loginResponse struct {
	OK  int    `json:"ok"`
	Msg string `json:"msg"`
}

func (c *connection) login() error {
	var firstErr error
	collector := c.collector.Clone()
	collector.OnResponse(func(res *colly.Response) {
		if firstErr != nil {
			return
		}
		resp := &loginResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("login: %s", err.Error())
			return
		}
		if resp.OK == 0 {
			firstErr = fmt.Errorf("login: %s", resp.Msg)
			return
		}
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("login: %s", err.Error())
	})

	hdr := c.headers.Clone()
	body := url.Values{}
	body.Add("l", c.username)
	body.Add("p", c.password)
	body.Add("security", "true")
	body.Add("h2", "")
	if err := collector.Request("POST",
		loginURL,
		bytes.NewBuffer([]byte(body.Encode())),
		nil,
		hdr); err != nil {
		return fmt.Errorf("login: %s", err.Error())
	}
	collector.Wait()

	return firstErr
}

type playerDataResponse struct {
	OK       int                   `json:"ok"`
	Msg      string                `json:"msg"`
	Charlist map[string]*Character `json:"charlist"`
}

func (c *connection) getPlayerData() error {
	var firstErr error
	collector := c.collector.Clone()
	collector.OnResponse(func(res *colly.Response) {
		if firstErr != nil {
			return
		}
		resp := &playerDataResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("getPlayerData: %s", err.Error())
			return
		}
		if resp.OK == 0 {
			firstErr = fmt.Errorf("getPlayerData: %s", resp.Msg)
			return
		}

		c.mutex.Lock()
		for _, char := range resp.Charlist {
			c.charlist = append(c.charlist, char)
		}
		c.mutex.Unlock()
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("getPlayerData: %s", err.Error())
	})

	hdr := c.headers.Clone()
	if err := collector.Request("POST",
		fmt.Sprintf(getPlayerDataURL, appVersion),
		nil,
		nil,
		hdr); err != nil {
		return fmt.Errorf("getPlayerData: %s", err.Error())
	}
	collector.Wait()

	return firstErr
}
