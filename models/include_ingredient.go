package Models

type Include_ingredient struct {
	Id_recipe int64 `json:"id_recipe"`
	Id_ingredient int64 `json:"id_ingredient"`
	Count int `json:"count"`
}
