package kodik

type videoData struct {
	Links map[string][]struct {
		Src string `json:"src"`
	} `json:"links"`
}
