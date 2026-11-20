package Models

type Include_recipe struct {
	Id_recipe int64 `json:"id_recipe"`
	Id_menu int64 `json:"id_menu"`
	Count int `json:"count"`
}
