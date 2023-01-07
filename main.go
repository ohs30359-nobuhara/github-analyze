package main

import (
	"github.com/spf13/cobra"
	"ohs30359/github-analyze/cmd"
	"time"
)

func main() {
	command := &cobra.Command{
		Use: "analyze",
		Run: func(c *cobra.Command, params []string) {
			var (
				args = cmd.CreateReportArgs{}
				e    error
			)

			args.Org, e = c.PersistentFlags().GetString("org")
			if e != nil {
				panic(e.Error())
			}

			args.Repo, e = c.PersistentFlags().GetString("repo")
			if e != nil {
				panic(e.Error())
			}

			args.Token, e = c.PersistentFlags().GetString("token")
			if e != nil {
				panic(e.Error())
			}

			args.Base, e = c.PersistentFlags().GetString("base")
			if e != nil {
				panic(e.Error())
			}

			args.Host, e = c.PersistentFlags().GetString("host")
			if e != nil {
				panic(e.Error())
			}

			args.Output, e = c.PersistentFlags().GetString("out")
			if e != nil || !(args.Output == "json" || args.Output == "excel") {
				panic(e.Error())
			}

			from, _ := c.PersistentFlags().GetString("from")
			args.From, e = time.Parse(time.RFC3339, from+"T00:00:00Z")
			if e != nil {
				panic(e.Error())
			}

			to, _ := c.PersistentFlags().GetString("to")
			args.To, e = time.Parse(time.RFC3339, to+"T00:00:00Z")
			if e != nil {
				panic(e.Error())
			}

			if e := cmd.CreateReport(args); e != nil {
				panic(e.Error())
			}
		},
	}
	command.PersistentFlags().String("org", "", "organization ex) ohs30359")
	command.PersistentFlags().String("repo", "", "repository ex) github-analyze")
	command.PersistentFlags().String("token", "", "github access token")
	command.PersistentFlags().String("base", "main", "target branch (default main)")
	command.PersistentFlags().String("from", "", "date from")
	command.PersistentFlags().String("to", "", "date to")
	command.PersistentFlags().String("host", "api.github.com", "github api host")
	command.PersistentFlags().String("out", "excel", "excel or json")

	if e := command.Execute(); e != nil {
		panic(e.Error())
	}
}
