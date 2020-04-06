package margonem

type Character struct {
	ID   string `json:"id"`
	Nick string `json:"nick"`
	Lvl  string `json:"lvl"`
	Prof string `json:"prof"`
	Icon string `json:"icon"`
	Last string `json:"last"`
	DB   string `json:"db"`
	Sex  string `json:"sex"`
}

func (ch *Character) Server() string {
	return serverNameDecoder.Replace(ch.DB)
}
