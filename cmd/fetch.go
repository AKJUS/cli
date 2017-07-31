package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/exercism/cli/api"
	"github.com/exercism/cli/config"
	"github.com/exercism/cli/user"
	app "github.com/urfave/cli"
)

// Fetch downloads exercism problems and writes them to disk.
func Fetch(ctx *app.Context) error {
	c, err := config.New(ctx.GlobalString("config"))
	if err != nil {
		log.Fatal(err)
	}
	client := api.NewClient(c)

	args := ctx.Args()
	var problems []*api.Problem

	if ctx.Bool("all") {
		if len(args) > 0 {
			trackID := args[0]
			fmt.Printf("\nFetching all problems for the %s track...\n\n", trackID)
			p, err := client.FetchAll(trackID)
			if err != nil {
				log.Fatal(err)
			}
			problems = p
		} else {
			log.Fatalf("You must supply a track to fetch all exercises")
		}
	} else {
		p, err := client.Fetch(args)
		if err != nil {
			log.Fatal(err)
		}
		problems = p
	}

	submissionInfo, err := client.Submissions()
	if err != nil {
		log.Fatal(err)
	}

	if err := setSubmissionState(problems, submissionInfo); err != nil {
		log.Fatal(err)
	}

	dirs, err := filepath.Glob(filepath.Join(c.Dir, "*"))
	if err != nil {
		log.Fatal(err)
	}

	dirMap := make(map[string]bool)
	for _, dir := range dirs {
		dirMap[dir] = true
	}
	hw := user.NewHomework(problems, c)

	if len(ctx.Args()) == 0 {
		if err := hw.RejectMissingTracks(dirMap); err != nil {
			log.Fatal(err)
		}
	}

	if err := hw.Save(); err != nil {
		log.Fatal(err)
	}

	hw.Summarize(user.HWAll)

	return nil
}

func setSubmissionState(problems []*api.Problem, submissionInfo map[string][]api.SubmissionInfo) error {
	for _, problem := range problems {
		langSubmissions := submissionInfo[problem.TrackID]
		for _, submission := range langSubmissions {
			if submission.Slug == problem.Slug {
				problem.Submitted = true
			}
		}
	}

	return nil
}
