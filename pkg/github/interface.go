package github

// API共通IF
type commonRequestIF struct {
	Org   string
	Repo  string
	Token string
	Host  string
}

// pulls api response
type pullResponseSchema struct {
	Id     int    `json:"Id"`
	State  string `json:"State"`
	Closed string `json:"closed_at"`
	Merged string `json:"merged_at"`
	Number int    `json:"number"`
	Labels []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"labels"`
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}
