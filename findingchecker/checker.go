package findingchecker

import (
	"fmt"
	"github.com/hazcod/go-intigriti"
	"github.com/hazcod/intigriti-slack-announce/config"
	"github.com/hazcod/intigriti-slack-announce/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func schedule(what func(), delay time.Duration, stopChan chan bool) {
	logrus.WithField("check_interval_minutes", delay.Minutes()).Debug("checks scheduled")
	ticker := time.NewTicker(delay)
	go func() {
		for {
			select {
			case <-stopChan:
				logrus.Debug("stopping checker")
				return
			case <-ticker.C:
				what()
			}
		}
	}()
}

func findingExists(config config.Config, finding intigriti.Submission) bool {
	_, found := config.FindingIDs[finding.ID]
	return found
}

func checkForNew(config config.Config, slckEndpoint map[string]slack.Endpoint, intiEndpoint intigriti.Endpoint) (func(), error) {
	return func() {

		logrus.Debug("checking for new findings")
		findings, err := intiEndpoint.GetSubmissions()
		if err != nil {
			logrus.WithError(err).Error("could not fetch from intigriti")
			return
		}

		if len(findings) == 0 {
			logrus.Debug("no findings found")
			return
		}

		for _, finding := range findings {
			fLogger := logrus.WithField("finding_id", finding.ID).WithField("finding_state", finding.State)
			if finding.State != "Triage" {
				logrus.WithField("msg", "Finding is not in Triage state, skipping")
				config.FindingIDs[finding.ID] = finding.ID
				continue
			}
			fLogger.Debug("looking if finding exists")
			if findingExists(config, finding) {
				fLogger.Debugf("finding %s already sent to slack, skipping", finding.ID)
				continue
			}
			if finding.DateCreated.Unix() < config.AppStartTime {
				fLogger.Debugf("Finding was created before application started, skipping")
				config.FindingIDs[finding.ID] = finding.ID
				continue
			}

			routes := []slack.Endpoint{slckEndpoint["sch-bug-bounty-testing"]}
			if channel, ok := config.ProgramChannelMap[finding.Program.Handle]; ok {
				fmt.Printf("%s should be routed to #%s\n", finding.Program.Handle, channel)
				routes = append(routes, slckEndpoint[channel])
			} else {
				fmt.Printf("%s not found in program-channel-map, only routing to default\n", channel)
			}
			for _, route := range routes {
				fmt.Printf("sending to %s \n", route.Channel)
				errs := route.Send(finding)
				if len(errs) > 0 {
					logrus.WithField("errors", fmt.Sprintf("%+v", errs)).Error("could not send to slack")
				} else {
					config.FindingIDs[finding.ID] = finding.ID
				}
			}
		}
	}, nil
}

func RunChecker(config config.Config, clientVersion string) error {

	slackEndpoints := make(map[string]slack.Endpoint)

	for channelName, webhookurl := range config.SlackWebhookURL {
		slackUrl, err := url.Parse(webhookurl)
		if err != nil {
			return errors.Wrap(err, "invalid slack url")
		}
		slackEndpoint := slack.NewEndpoint(*slackUrl, channelName, clientVersion)
		slackEndpoints[channelName] = slackEndpoint
	}

	intigritiEndpoint := intigriti.New(config.IntigritiClientID, config.IntigritiClientSecret)

	checkFunc, err := checkForNew(config, slackEndpoints, intigritiEndpoint)
	if err != nil {
		return errors.Wrap(err, "could not initialize checker")
	}

	// should we ever want to stop it
	stopChan := make(chan bool, 1)

	// recurring runs
	schedule(checkFunc, time.Minute*time.Duration(config.CheckInterval), stopChan)

	logrus.Info("checker is is now running")

	// trigger first run immediately
	checkFunc()

	return nil
}
