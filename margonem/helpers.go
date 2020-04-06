package margonem

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	serverNameDecoder = strings.NewReplacer(
		`#`, "",
	)
)

func hash(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateMucka() float32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32()
}

func random(min, max int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(max-min) + min
}

func parseStatStr(stat string, fields ...string) (map[string]int, error) {
	m := make(map[string]int)
	for _, s := range strings.Split(stat, ";") {
		for _, f := range fields {
			if strings.Contains(s, f) {
				val, err := strconv.Atoi(strings.Split(s, "=")[1])
				if err != nil {
					fmt.Println(s, val)
					return nil, fmt.Errorf("parseStatStr: %s", err.Error())
				}
				m[f] = val
				break
			}
		}
	}
	return m, nil
}
