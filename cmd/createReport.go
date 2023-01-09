package cmd

import (
	"encoding/json"
	"fmt"
	"ohs30359/github-analyze/pkg/excel"
	"ohs30359/github-analyze/pkg/github"
	"os"
	"time"
)

type CreateReportArgs struct {
	github.PullsRequest
	Output string
}

type result struct {
	// YYYY-MM, UserName
	Reviews map[string]map[string]int `json:"reviews"`
	// YYYY-MM, labelName
	Prs map[string]map[string]int `json:"prs"`
	// YYYY-MM, user
	Committer map[string]map[string]int `json:"committer"`
}

func CreateReport(args CreateReportArgs) error {
	result, e := getResult(args.PullsRequest)
	if e != nil {
		return e
	}

	fileName := "./report"
	switch args.Output {
	case "json":
		jsonStr, e := json.Marshal(result)
		if e != nil {
			return e
		}

		f, e := os.Create(fileName + ".json")
		if e != nil {
			return e
		}

		defer f.Close()
		if _, e := f.Write(jsonStr); e != nil {
			return e
		}
	case "excel":
		name := fileName + ".xlsx"
		if e := excel.WriteFromMap(result.Reviews, name, "user別review実施数"); e != nil {
			return e
		}
		if e := excel.WriteFromMap(result.Prs, name, "label別pr数"); e != nil {
			return e
		}
		if e := excel.WriteFromMap(result.Committer, name, "user別実装数"); e != nil {
			return e
		}
	}
	return nil
}

// getResult 集計結果を取得
func getResult(args github.PullsRequest) (result, error) {
	prs, e := github.GetPulls(args)
	if e != nil {
		return result{}, e
	}

	result := result{
		Reviews:   make(map[string]map[string]int),
		Prs:       make(map[string]map[string]int),
		Committer: make(map[string]map[string]int),
	}

	for _, pr := range prs.Prs {
		time.Sleep(time.Second * 1)
		// 並列で叩いてもいいが下手をするとDDosになるため1件ごとにdelayをかけて取得する
		req := github.ReviewRequest{PrId: pr.PrNumber}
		req.Org = args.Org
		req.Repo = args.Repo
		req.Token = args.Token
		req.Host = args.Host

		// review取得に失敗した場合はskip (TODO: retryのほうが良いので変更する)
		review, e := github.GetReview(req)
		if e != nil {
			fmt.Println(e.Error())
			continue
		}

		// 月ごとに集計を行うのでcloseした日付から付きを取得
		// ※ resultSetの第二回層のmapは初期化されないのでnilを考慮する必要あり
		closeDate, _ := time.Parse(time.RFC3339, pr.ClosedDate)
		prefixMonth := closeDate.Format("2006-01")

		if _, ok := result.Prs[prefixMonth]; !ok {
			result.Prs[prefixMonth] = make(map[string]int)
		}
		if _, ok := result.Reviews[prefixMonth]; !ok {
			result.Reviews[prefixMonth] = make(map[string]int)
		}
		if _, ok := result.Committer[prefixMonth]; !ok {
			result.Committer[prefixMonth] = make(map[string]int)
		}

		// レビューとPRの総合件数の集計
		result.Prs[prefixMonth]["all"] += 1
		result.Reviews[prefixMonth]["all"] += review.TotalComment

		// Committerを集計
		result.Committer[prefixMonth][pr.Committer] += 1

		// labelごとにprをの集計 (prとlabelは 1:n のため allと同値にはならない)
		for _, label := range pr.Labels {
			result.Prs[prefixMonth][label] = result.Prs[prefixMonth][label] + 1
		}

		// user単位のレビュー数集計 (commentの数が多いほど質の良いreviewerと判断)
		for _, comment := range review.Reviewers {
			result.Reviews[prefixMonth][comment.User] = result.Reviews[prefixMonth][comment.User] + comment.Cnt
		}
	}
	return result, nil
}
