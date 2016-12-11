package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

func main() {
	githubToken := flag.String("github-token", "", "Github token")
	webhookUrl := flag.String("webhook-url", "", "Webhook URL")

	if *githubToken == "" || *webhookUrl == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Starting")
	err := registerWebHook("gjacquet", "consul-backup-tool", *githubToken, *webhookUrl)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8080", nil)
}

func getHook(repositories *github.RepositoriesService, owner string, repo string, url string) (*github.Hook, error) {
	hooks, _, err := repositories.ListHooks(owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, b := range hooks {
		fmt.Printf("Current hook: %v, looking for %v", b.Config["url"], url)
		if b.Config["url"] == url {
			fmt.Printf("found hook")
			return b, nil
		}
	}

	return nil, nil
}

func registerWebHook(owner string, repo string, token string, url string) error {
	fmt.Println("registering hook")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	repositories := client.Repositories

	hook, err := getHook(repositories, owner, repo, url)
	if err != nil {
		return err
	}
	if hook == nil {
		_, _, err := repositories.CreateHook(owner, repo, &github.Hook{
			Name:   github.String("web"),
			Events: []string{"pull_request_review_comment", "issue_comment"},
			Active: github.Bool(true),
			Config: map[string]interface{}{
				"url":          url,
				"secret":       "toto",
				"content_type": "json",
			},
		})

		fmt.Println("Created hook")

		if err != nil {
			return err
		}
	} else {
		_, _, err := repositories.EditHook(owner, repo, *hook.ID, &github.Hook{
			// FIXME should be add_events
			Events: []string{"pull_request_review_comment", "issue_comment"},
			Active: github.Bool(true),
			Config: map[string]interface{}{
				"url":          "http://87fdb834.ngrok.io",
				"secret":       "toto",
				"content_type": "json",
			},
		})

		fmt.Println("Updated hook")

		if err != nil {
			return err
		}
	}

	return nil
}

func processIssueCommentEvent(event github.IssueCommentEvent) error {
	fmt.Printf("Message: %v", event.Comment)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte("toto"))
	if err != nil {
		fmt.Println("Error in payload validation")

	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		fmt.Println("Error parsing hook")
	}
	fmt.Printf("Received event: %v", event)

	switch event := event.(type) {
	case github.IssueCommentEvent:
		processCommitCommentEvent(event)
	default:
		fmt.Printf("Received event of type %v\n", event.(type))
	}
}
