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
	Summary map[string]map[string]int `json:"summary"`
	// YYYY-MM, UserName
	Reviews map[string]map[string]int `json:"reviews"`
	// YYYY-MM, labelName
	Pulls map[string]map[string]int `json:"pulls"`
	// YYYY-MM, user
	Committer map[string]map[string]int `json:"committer"`
}

func CreateReport(args CreateReportArgs) error {
	result, e := getResult(args.PullsRequest)
	if e != nil {
		return e
	}

	fileName := "./report2"
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
		if e := excel.WriteFromMap(result.Summary, name, ""); e != nil {
			return e
		}
		if e := excel.WriteFromMap(result.Reviews, name, "review trend (by user)"); e != nil {
			return e
		}
		if e := excel.WriteFromMap(result.Pulls, name, "pull request trend (by label)"); e != nil {
			return e
		}
		if e := excel.WriteFromMap(result.Committer, name, "commit trend (by user)"); e != nil {
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
		Summary:   make(map[string]map[string]int),
		Reviews:   make(map[string]map[string]int),
		Pulls:     make(map[string]map[string]int),
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

		if _, ok := result.Committer[prefixMonth]; !ok {
			result.Summary[prefixMonth] = make(map[string]int)
		}
		if _, ok := result.Pulls[prefixMonth]; !ok {
			result.Pulls[prefixMonth] = make(map[string]int)
		}
		if _, ok := result.Reviews[prefixMonth]; !ok {
			result.Reviews[prefixMonth] = make(map[string]int)
		}
		if _, ok := result.Committer[prefixMonth]; !ok {
			result.Committer[prefixMonth] = make(map[string]int)
		}

		// レビューとPRの総合件数の集計
		result.Summary[prefixMonth]["pull"] += 1
		result.Summary[prefixMonth]["review"] += review.TotalComment
		if review.TotalComment == 0 {
			result.Summary[prefixMonth]["no review pull"] += 1
		} else {
			result.Summary[prefixMonth]["no review pull"] += 0
		}

		// Committerを集計
		result.Committer[prefixMonth][pr.Committer] += 1

		// labelごとにprをの集計 (prとlabelは 1:n のため allと同値にはならない)
		for _, label := range pr.Labels {
			result.Pulls[prefixMonth][label] = result.Pulls[prefixMonth][label] + 1
		}
		// 存在しないケースは "no label" で集計する
		if len(pr.Labels) == 0 {
			result.Pulls[prefixMonth]["no label"] = result.Pulls[prefixMonth]["no label"] + 1
		}

		// user単位のレビュー数集計 (commentの数が多いほど質の良いreviewerと判断)
		for _, comment := range review.Reviewers {
			result.Reviews[prefixMonth][comment.User] = result.Reviews[prefixMonth][comment.User] + comment.Cnt
		}
	}
	return result, nil
}
