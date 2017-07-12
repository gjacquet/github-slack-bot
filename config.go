package main

type GithubWebhookConfiguration struct {
	Url    *string
	Secret *string
	Events []string
}

type GithubConfiguration struct {
	Owner   *string
	Repo    *string
	Token   *string
	Webhook *GithubWebhookConfiguration
}

type SlackConfiguration struct {
}

type ApplicationConfiguration struct {
	Github *GithubConfiguration
}
