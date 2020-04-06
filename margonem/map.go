package margonem

type Map struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	StaminaCostFight int    `json:"stamina_cost_fight"`
	StaminaCostBoss  int    `json:"stamina_cost_boss"`
}
