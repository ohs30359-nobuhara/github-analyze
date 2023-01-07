package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PullsRequest struct {
	commonRequestIF
	Base string
	From time.Time
	To   time.Time
}

type PullsResponse struct {
	Prs []prSchema
}

type prSchema struct {
	Id         int
	Committer  string
	Labels     []string
	MergedDate string
	ClosedDate string
	PrNumber   int
}

// GetPulls PRの一覧を取得
func GetPulls(args PullsRequest) (PullsResponse, error) {
	var result PullsResponse

	// pulls apiを使うと30件までしか取れないためissue経由で取得
	url := fmt.Sprintf("https://%s/search/issues?q=is:pr+state:closed+base:%s+repo:%s/%s", args.Host, args.Base, args.Org, args.Repo)
	// page index 0 で取得 (以降のpageはloopで回す)
	items, nextUrl, e := getPullsPageNation(url, args.Token)
	pr, next := pageNation(items, args.From, args.To)
	result.Prs = append(result.Prs, pr...)

	// pageNationはnextUrlが消えるまで繰り返し取得 (負荷を考慮して1秒delay)
	for len(nextUrl) != 0 || next {
		time.Sleep(time.Second * 1)
		items, nextUrl, e = getPullsPageNation(nextUrl, args.Token)
		pr, next := pageNation(items, args.From, args.To)
		result.Prs = append(result.Prs, pr...)

		// 指定期間外なら処理中断
		if !next {
			break
		}
	}

	if e != nil {
		return PullsResponse{}, e
	}

	return result, nil
}

// pageNation pageに含まれるprを取得してresponse形式で返す
func pageNation(items []pullResponseSchema, from time.Time, to time.Time) ([]prSchema, bool) {
	var (
		prs  []prSchema
		next = true
	)
	for _, item := range items {
		// ラベル情報の取得
		var labels []string
		for _, label := range item.Labels {
			labels = append(labels, label.Name)
		}

		// 全件取得されるので擬似的に 帰還を絞る (sortされているので対象期間外のレコードが一度出れば以降は該当レオードは存在しない)
		closedDate, _ := time.Parse(time.RFC3339, item.Closed)
		if closedDate.After(from) && closedDate.Before(to) {
			// before < closed < to → 指摘期間内
			pr := prSchema{
				Id:         item.Id,
				Committer:  item.User.Login,
				Labels:     labels,
				MergedDate: item.Merged,
				ClosedDate: item.Closed,
				PrNumber:   item.Number,
			}
			prs = append(prs, pr)
			fmt.Printf("pr: %s, date: %s は対象期間内 \n", strconv.Itoa(item.Number), closedDate.String())
		} else if !closedDate.After(from) {
			// closed > before → これ以上過去のデータは不要
			fmt.Printf("pr: %s, date: %s は対象期間外 ※以降のprはすべてskip\n", strconv.Itoa(item.Number), closedDate.String())
			next = false
			break
		}
	}
	return prs, next
}

// getPullsPageNation
func getPullsPageNation(url string, token string) ([]pullResponseSchema, string, error) {
	fmt.Printf("request to %s \n", url)
	req, e := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if e != nil {
		return nil, "", e
	}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := new(http.Client)
	res, e := client.Do(req)
	if e != nil {
		return nil, "", e
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", errors.New(res.Status)
	}

	body, e := io.ReadAll(res.Body)
	if e != nil {
		return nil, "", e
	}

	var dataset struct {
		Items []pullResponseSchema `json:"items"`
	}
	if e := json.Unmarshal(body, &dataset); e != nil {
		return nil, "", e
	}

	next := res.Header.Get("Link")
	if len(next) == 0 {
		return dataset.Items, "", e
	}

	var nextUrl string
	urls := strings.Split(next, `,`)
	for _, test := range urls {
		if strings.Contains(test, `rel="next"`) {
			nextUrl = test
		}
	}
	nextUrl = strings.Replace(nextUrl, "<", "", -1)
	nextUrl = strings.Replace(nextUrl, `>; rel="next"`, "", -1)
	nextUrl = strings.TrimSpace(nextUrl)

	return dataset.Items, nextUrl, e
}
