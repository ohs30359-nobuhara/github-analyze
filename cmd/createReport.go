package cmd

import (
	"encoding/json"
	"fmt"
	"ohs30359/github-analyze/pkg/github"
	"time"
)

type CreateReportArgs struct {
	github.PullsRequest
}

type Result struct {
	// YYYY-MM, UserName
	Reviews map[string]map[string]int `json:"reviews"`
	// YYYY-MM, labelName
	Prs map[string]map[string]int `json:"prs"`
	// YYYY-MM, user
	Committer map[string]map[string]int `json:"committer"`
}

func CreateReport(args CreateReportArgs) error {
	result := Result{
		Reviews:   make(map[string]map[string]int),
		Prs:       make(map[string]map[string]int),
		Committer: make(map[string]map[string]int),
	}

	prs, e := github.GetPulls(args.PullsRequest)
	if e != nil {
		return e
	}

	for _, pr := range prs.Prs {
		time.Sleep(time.Second * 1)
		// 並列で叩いてもいいが下手をするとDDosになるため1件ごとにdelayをかけて取得する
		req := github.ReviewRequest{PrId: pr.PrNumber}
		req.Org = args.Org
		req.Repo = args.Repo
		req.Token = args.Token

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

	st, e := json.Marshal(result)
	if e != nil {
		return e
	}

	fmt.Println(string(st))
	return nil
}
