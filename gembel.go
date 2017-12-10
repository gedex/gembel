package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	ghClient *github.Client
	ghCtx    context.Context

	reHex = regexp.MustCompile("^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$")

	version = "master"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	Labels       []Label
	Repositories []string
}

type Label struct {
	Name    string
	Color   string
	Replace string
}

type Action int

const (
	Create Action = iota
	Update
)

var actions = [...]string{
	"create",
	"update",
}

func (a Action) String() string {
	return actions[a]
}

type Result struct {
	Action Action
	From   Label
	To     Label
	Error  error
}

func (r Result) String() string {
	prefix := "[OK]"
	if r.Error != nil {
		prefix = "[FAIL]"
	}

	var ret string
	switch r.Action {
	case Update:
		ret = fmt.Sprintf("%s Updated label named '%s' with color '%s' to '%s' with color '%s'", prefix, r.From.Name, r.From.Color, r.To.Name, r.To.Color)
	case Create:
		ret = fmt.Sprintf("%s Created label named '%s' with color '%s'", prefix, r.To.Name, r.To.Color)
	}

	return ret
}

func (c *Config) check() error {
	if len(c.Labels) == 0 {
		return errors.New("Empty labels in config file")
	}
	if len(c.Repositories) == 0 {
		return errors.New("Empty target repositories in config file")
	}

	m := make(map[string]bool, 0)
	for i, label := range c.Labels {
		if label.Name == "" {
			return errors.New("label name can not be empty")
		}
		if label.Color == "" {
			return errors.New("label color can not be empty")
		}
		if strings.HasPrefix(label.Color, "#") {
			label.Color = strings.TrimPrefix(label.Color, "#")
			c.Labels[i].Color = label.Color
		}

		if !reHex.MatchString(label.Color) {
			return errors.New("label color must be in 6 character hex code")
		}

		if _, ok := m[label.Name]; ok {
			return fmt.Errorf("%s in `replaces` is used more than once", label.Name)
		}
	}

	for _, repo := range c.Repositories {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format %s, shoud be user/repo", repo)
		}
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid repo format %s, shoud be user/repo", repo)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage(errors.New("missing config file"))
	}
	if os.Getenv("GITHUB_TOKEN") == "" {
		usage(errors.New("empty GITHUB_TOKEN in env"))
	}
	c, err := ReadConfig(os.Args[1])
	if err != nil {
		usage(err)
	}
	run(c)
}

func ReadConfig(path string) (c *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()

	fc, err := ioutil.ReadAll(f)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(fc, &c); err != nil {
		return c, fmt.Errorf("json unmarshal error: %s", err)
	}
	err = c.check()

	return c, err
}

func run(c *Config) {
	ghCtx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ghCtx, ts)
	ghClient = github.NewClient(tc)

	for _, repoPath := range c.Repositories {
		fmt.Printf("Update labels in repo %s...\n", repoPath)
		results, err := UpdateRepo(repoPath, c.Labels)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		printResults(results)
	}
}

func printResults(results []Result) {
	for _, result := range results {
		fmt.Printf("* %s\n", result)
	}
	fmt.Println("")
}

func UpdateRepo(repoPath string, labels []Label) (results []Result, err error) {
	// First, get all labels, mapped to their colors, from the repoOwner.
	repoLabels, err := GetRepoLabels(repoPath)
	if err != nil {
		return results, err
	}

	// Foreach labels from config:
	// - If label name exists in current labels, perform update. Probably
	//   the color changes.
	// - If label replace found in repoLabels, perform update.
	// - If no match create new label.
	var result Result
	for _, label := range labels {
		if color, ok := repoLabels[label.Name]; ok {
			result = Result{
				Action: Update,
				From: Label{
					Name:  label.Name,
					Color: color,
				},
				To:    label,
				Error: UpdateLabel(repoPath, label.Name, label),
			}
		} else if color, ok := repoLabels[label.Replace]; ok {
			result = Result{
				Action: Update,
				From: Label{
					Name:  label.Replace,
					Color: color,
				},
				To:    label,
				Error: UpdateLabel(repoPath, label.Replace, label),
			}
		} else {
			result = Result{
				Action: Create,
				From:   Label{},
				To:     label,
				Error:  CreateLabel(repoPath, label),
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func UpdateLabel(repoPath, labelName string, label Label) error {
	parts := strings.Split(repoPath, "/")
	owner, repo := parts[0], parts[1]

	ghLabel := &github.Label{
		Name:  &label.Name,
		Color: &label.Color,
	}

	if _, _, err := ghClient.Issues.EditLabel(ghCtx, owner, repo, labelName, ghLabel); err != nil {
		return err
	}

	return nil
}

func CreateLabel(repoPath string, label Label) error {
	parts := strings.Split(repoPath, "/")
	owner, repo := parts[0], parts[1]

	ghLabel := &github.Label{
		Name:  &label.Name,
		Color: &label.Color,
	}

	if _, _, err := ghClient.Issues.CreateLabel(ghCtx, owner, repo, ghLabel); err != nil {
		return err
	}

	return nil
}

func GetRepoLabels(repoPath string) (m map[string]string, err error) {
	parts := strings.Split(repoPath, "/")
	owner, repo := parts[0], parts[1]
	opt := &github.ListOptions{
		PerPage: 100,
	}

	m = make(map[string]string)
	for {
		repoLabels, resp, err := ghClient.Issues.ListLabels(ghCtx, owner, repo, opt)
		if err != nil {
			return m, err
		}
		for _, label := range repoLabels {
			m[label.GetName()] = label.GetColor()
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return m, err
}

func getVersion() string {
	return fmt.Sprintf("%v, commit %v, built at %v", version, commit, date)
}

func usage(err error) {
	fmt.Printf("Error: %v\n", err)
	fmt.Printf(`
Name:
  gembel - bulk update issue labels of GitHub repositories.

Version:
  %s

Usage:
  gembel <config-file>

  To specifiy GITHUB_TOKEN when running it:

  GITHUB_TOKEN=token gembel <config-file>
`, getVersion())

	os.Exit(1)
}
