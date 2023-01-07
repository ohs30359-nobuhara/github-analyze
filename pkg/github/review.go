package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type ReviewRequest struct {
	commonRequestIF
	PrId int
}

type ReviewResponse struct {
	TotalComment int
	Reviewers    []struct {
		User string
		Cnt  int
	}
}

// GetReview 指定したPRのReviewを取得
func GetReview(args ReviewRequest) (ReviewResponse, error) {
	type Response []struct {
		User struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"user"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/reviews", args.Org, args.Repo, strconv.Itoa(args.PrId))
	fmt.Printf("request to %s \n", url)
	req, e := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if e != nil {
		return ReviewResponse{}, e
	}

	req.Header.Set("Authorization", "Token "+args.Token)
	req.Header.Set("Content-Type", "application/json")

	client := new(http.Client)
	res, e := client.Do(req)
	if e != nil {
		return ReviewResponse{}, e
	}
	defer res.Body.Close()

	body, e := io.ReadAll(res.Body)
	if e != nil {
		return ReviewResponse{}, e
	}

	if res.StatusCode != http.StatusOK {
		return ReviewResponse{}, errors.New(res.Status)
	}

	var data Response
	if e := json.Unmarshal(body, &data); e != nil {
		return ReviewResponse{}, e
	}

	var gheResp ReviewResponse

	// ユーザーごとのcomment数を集計
	m := make(map[string]int)
	for _, user := range data {
		u := user.User.Login
		if _, ok := m[u]; !ok {
			m[u] = 0
		}
		m[u] += 1
		gheResp.TotalComment += 1
	}

	for user, cnt := range m {
		gheResp.Reviewers = append(gheResp.Reviewers, struct {
			User string
			Cnt  int
		}{User: user, Cnt: cnt})
	}

	return gheResp, nil
}
