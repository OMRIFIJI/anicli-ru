package alloha

type quality map[string]string

type hlsSource struct {
	Label   string  `json:"label"`
	Quality quality `json:"quality"`
}

type videoData struct {
	HLSSources []hlsSource `json:"hlsSource"`
}
