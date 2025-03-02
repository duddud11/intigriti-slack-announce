package slack

import (
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/hazcod/go-intigriti"
	"regexp"
)

var (
	regexRemove = regexp.MustCompile(`[\\W_]+`)
)

// remove all chars except alphanumeric
func cleanStr(str string) string {
	return string(regexRemove.ReplaceAll([]byte(str), []byte("")))
}

func buildMessage(f intigriti.Submission) slack.Payload {
	attach := slack.Attachment{}

	/*
		attach.AddField(slack.Field{
			Title: "Created",
			Value: f.Timestamp.Format("2006-01-02 15:04:05"),
		})
	*/

	attach.AddField(slack.Field{
		Title: "Severity",
		Value: f.Severity,
	})
	attach.AddField(slack.Field{
		Title: "Type",
		Value: f.Type,
	})

	if f.Endpoint != "" {
		attach.AddField(slack.Field{
			Title: "Endpoint",
			Value: f.Endpoint,
		})
	}

	return slack.Payload{
		Username: "intigriti",
		Text: fmt.Sprintf("A new *%s* finding by *%s* for *%s*: <%s|%s>",
			cleanStr(f.Severity), cleanStr(f.Researcher.Username), cleanStr(f.Program.Name), f.URL, cleanStr(f.Title)),
		Attachments: []slack.Attachment{attach},
	}
}
