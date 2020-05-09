package margonem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	mobileTokenSalt = "humantorch-"
	initURL         = "http://%s.margonem.pl/engine?t=init&initlvl=%d&mucka=%g&aid=%s&mobile=1"
	eventURL        = "http://%s.margonem.pl/engine?t=%s&aid=%s&mobile=1&ev=&mobile_token=%s"
)

type serverConnection struct {
	mutex          sync.Mutex
	collector      *colly.Collector
	headers        http.Header
	mobileToken    string
	mobileTokenMD5 string
	charID         string
	server         string
	userID         string
	hp             int
	maxhp          int
	stamina        int
	maps           map[string]*Map
	items          map[string]*Item
}

type serverResponse struct {
	Alert       string `json:"alert"`
	T           string `json:"t"`
	E           string `json:"e"`
	MobileToken string `json:"mobile_token"`
	Character   struct {
		Stamina *int `json:"stamina"`
		Stats   struct {
			HP    *int `json:"hp"`
			MaxHP *int `json:"maxhp"`
		} `json:"warrior_stats"`
	} `json:"h"`
	MobileMaps map[string]*Map  `json:"mobile_maps"`
	Items      map[string]*Item `json:"item"`
	Loot       struct {
		Source string         `json:"source"`
		States map[string]int `json:"states"`
	} `json:"loot"`
}

func (c *serverConnection) canSendAttack(mapid string) bool {
	if m, ok := c.maps[mapid]; ok {
		return c.stamina >= m.StaminaCostFight
	}
	return false
}

func (c *serverConnection) handleNewRequest(req *colly.Request) {
	cookie := &http.Cookie{
		Value:  c.charID,
		Name:   "mchar_id",
		Domain: ".margonem.pl",
	}
	req.Headers.Set("cookie", cookie.String())
	urlValues := req.URL.Query()
	if _, ok := urlValues["ev"]; ok {
		urlValues["ev"] = []string{
			fmt.Sprintf("%f", float64(time.Now().UnixNano())/1000000000),
		}
		req.URL.RawQuery = urlValues.Encode()
	}
}

func (c *serverConnection) init() error {
	var firstErr error
	mucka := generateMucka()
	collector := c.collector.Clone()
	collector.OnRequest(c.handleNewRequest)
	collector.OnResponse(func(res *colly.Response) {
		if firstErr != nil {
			return
		}
		resp := &serverResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("initServerConnection: %s", err.Error())
			return
		}
		if firstErr = c.handleResponse(resp); firstErr != nil {
			return
		}
		if resp.T == "stop" || resp.Alert != "" || resp.E != "ok" {
			msg := resp.Alert
			if msg == "" {
				msg = resp.E
			}
			firstErr = fmt.Errorf("initServerConnection: %s", msg)
			return
		}
		initLvl := res.Request.URL.Query()["initlvl"][0]
		c.mutex.Lock()
		if initLvl == "1" {
			c.mobileToken = resp.MobileToken
			c.mobileTokenMD5 = hash(mobileTokenSalt + c.mobileToken)
			c.stamina = *resp.Character.Stamina
			c.hp = *resp.Character.Stats.HP
			c.maxhp = *resp.Character.Stats.MaxHP
		} else if initLvl == "2" {
			c.maps = resp.MobileMaps
		} else if initLvl == "3" {
			c.items = resp.Items
		}
		c.mutex.Unlock()
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("initServerConnection: %s", err.Error())
	})
	hdr := c.headers.Clone()
	for i := 1; i <= 4; i++ {
		uri := fmt.Sprintf(initURL, c.server, i, mucka, c.userID)
		if i > 1 {
			uri += "&mobile_token=" + c.mobileTokenMD5
		}
		if err := collector.Request("POST",
			uri,
			nil,
			nil,
			hdr); err != nil {
			return fmt.Errorf("initServerConnection: %s", err.Error())
		}
		collector.Wait()
		if firstErr != nil {
			return firstErr
		}
	}

	return c.sendSingleEvent()
}

func (c *serverConnection) sendSingleEvent() error {
	var firstErr error
	collector := c.collector.Clone()
	hdr := c.headers.Clone()

	collector.OnRequest(c.handleNewRequest)
	collector.OnResponse(func(res *colly.Response) {
		if firstErr != nil {
			return
		}
		resp := &serverResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("sendSingleEvent: %s", err.Error())
			return
		}
		if firstErr = c.handleResponse(resp); firstErr != nil {
			return
		}
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("sendSingleEvent: %s", err.Error())
	})

	if err := collector.Request("POST",
		fmt.Sprintf(eventURL, c.server, "_", c.userID, c.mobileTokenMD5),
		nil,
		nil,
		hdr); err != nil {
		return fmt.Errorf("sendSingleEvent: %s", err.Error())
	}
	collector.Wait()
	return firstErr
}

func (c *serverConnection) heal() error {
	if c.maxhp <= c.hp {
		return nil
	}
	var firstErr error
	collector := c.collector.Clone()
	hdr := c.headers.Clone()
	collector.OnRequest(c.handleNewRequest)
	collector.OnResponse(func(res *colly.Response) {
		resp := &serverResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("heal: %s", err.Error())
			return
		}
		if firstErr = c.handleResponse(resp); firstErr != nil {
			return
		}

	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("heal: %s", err.Error())
	})
	for itemID, item := range c.items {
		if item.Cl == 16 && strings.Contains(item.Stat, "leczy") {
			values, err := parseStatStr(item.Stat, "leczy", "amount")
			if err != nil {
				return err
			}
			amount := values["amount"]
			add := values["leczy"]
			for amount > 0 && c.hp < c.maxhp {
				if err := collector.Request("POST",
					c.getHealRequestURL(itemID),
					nil,
					nil,
					hdr); err != nil {
					return fmt.Errorf("heal: %s", err.Error())
				}
				collector.Wait()
				if firstErr != nil {
					return firstErr
				}
				c.mutex.Lock()
				c.hp += add
				c.mutex.Unlock()
				amount--
			}
		} else if item.Cl == 16 && strings.Contains(item.Stat, "fullheal") {
			values, err := parseStatStr(item.Stat, "fullheal")
			if err != nil {
				return err
			}
			add := values["fullheal"]
			if add > 0 && c.hp < c.maxhp {
				if err := collector.Request("POST",
					c.getHealRequestURL(itemID),
					nil,
					nil,
					hdr); err != nil {
					return fmt.Errorf("heal: %s", err.Error())
				}
				collector.Wait()
				if firstErr != nil {
					return firstErr
				}
				c.mutex.Lock()
				c.hp += add
				c.mutex.Unlock()
			}
		}
		if c.hp >= c.maxhp {
			break
		}
	}
	return nil
}

func (c *serverConnection) attack(mapID string) error {
	if c.stamina == 0 {
		return nil
	}
	if _, ok := c.maps[mapID]; !ok {
		return fmt.Errorf("attack: Cannot find map with id %s", mapID)
	}

	var firstErr error
	collector := c.collector.Clone()
	hdr := c.headers.Clone()
	collector.OnRequest(c.handleNewRequest)
	collector.OnResponse(func(res *colly.Response) {
		resp := &serverResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("attack: %s", err.Error())
			return
		}
		if firstErr = c.handleResponse(resp); firstErr != nil {
			return
		}
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("attack: %s", err.Error())
	})
	for i, add := range []string{
		fmt.Sprintf("&a=attack&town_id=%s", mapID),
		"&a=f",
		"&a=quit",
	} {
		if err := collector.Request("POST",
			fmt.Sprintf(eventURL,
				c.server,
				"fight"+add,
				c.userID,
				c.mobileTokenMD5),
			nil,
			nil,
			hdr); err != nil {
			return fmt.Errorf("attack: %s", err.Error())
		}
		collector.Wait()
		if firstErr != nil {
			return firstErr
		}
		if i == 1 {
			time.Sleep(1500)
		}
	}

	return c.sendSingleEvent()
}

func (c *serverConnection) getHealRequestURL(itemID string) string {
	return fmt.Sprintf(eventURL,
		c.server,
		fmt.Sprintf("moveitem&id=%s&st=1", itemID),
		c.userID,
		c.mobileTokenMD5)
}

func (c *serverConnection) handleLoot(res *serverResponse) error {
	var firstErr error
	collector := c.collector.Clone()
	hdr := c.headers.Clone()

	collector.OnRequest(c.handleNewRequest)
	collector.OnResponse(func(res *colly.Response) {
		if firstErr != nil {
			return
		}
		resp := &serverResponse{}
		if err := json.Unmarshal(res.Body, resp); err != nil {
			firstErr = fmt.Errorf("handleLoot: %s", err.Error())
			return
		}
		if firstErr = c.handleResponse(resp); firstErr != nil {
			return
		}
	})
	collector.OnError(func(res *colly.Response, err error) {
		if firstErr != nil {
			return
		}
		firstErr = fmt.Errorf("handleLoot: %s", err.Error())
	})
	urls := []string{
		fmt.Sprintf(eventURL,
			c.server,
			"loot&not=&want=&must=&final=1",
			c.userID,
			c.mobileTokenMD5),
	}
	for id := range res.Loot.States {
		urls = append(urls, fmt.Sprintf(eventURL,
			c.server,
			"shop&sell="+id,
			c.userID,
			c.mobileTokenMD5))
	}
	for _, url := range urls {
		if err := collector.Request("POST",
			url,
			nil,
			nil,
			hdr); err != nil {
			return fmt.Errorf("handleLoot: %s", err.Error())
		}
		collector.Wait()
		if firstErr != nil {
			return firstErr
		}
	}

	return nil
}

func (c *serverConnection) handleResponse(res *serverResponse) error {
	if res.T == "stop" || res.Alert != "" || res.E != "ok" {
		msg := res.Alert
		if msg == "" {
			msg = res.E
		}
		return fmt.Errorf("handleResponse: %s", msg)
	}

	c.mutex.Lock()
	if res.Character.Stamina != nil {
		c.stamina = *res.Character.Stamina
	}

	if res.Character.Stats.HP != nil {
		c.hp = *res.Character.Stats.HP
	}

	if res.Character.Stats.MaxHP != nil {
		c.maxhp = *res.Character.Stats.MaxHP
	}
	c.mutex.Unlock()
	if res.Loot.Source == "fight" {
		if err := c.handleLoot(res); err != nil {
			return err
		}
	}

	return nil
}
