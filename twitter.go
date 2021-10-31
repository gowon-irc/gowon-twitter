package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

var (
	tweetUrlRe = regexp.MustCompile(tweetURLRegex)
)

func twit(args string, client *twitter.Client) (msg string, err error) {
	fields := strings.Fields(args)

	if len(fields) == 0 {
		return "Error: pass a user", nil
	}

	arg := strings.Fields(args)[0]

	user, _, err := client.Users.Show(&twitter.UserShowParams{
		ScreenName: arg,
	})

	if err != nil {
		return "", err
	}

	if user.StatusesCount == 0 {
		return fmt.Sprintf("User %s has no tweets", arg), nil
	}

	t := strings.Split(user.Status.Text, "\n")[0]

	return t, nil
}

func tweetFromUrl(args string, client *twitter.Client) (msg string, err error) {
	urls := tweetUrlRe.FindAllString(args, -1)

	if len(urls) == 0 {
		return "", errors.New("no tweet urls found in string")
	}

	outList := []string{}

	for _, url := range urls {
		urlsplit := strings.Split(url, "/")
		id := urlsplit[len(urlsplit)-1]
		iid, _ := strconv.Atoi(id)
		tweet, _, err := client.Statuses.Show(int64(iid), nil)

		if err != nil {
			outList = append(outList, fmt.Sprintf("Could not find tweet with url %s", url))
			continue
		}

		outList = append(outList, fmt.Sprintf("{cyan}Tweet from %s:{clear} %s", tweet.User.Name, tweet.Text))
	}

	return strings.Join(outList, "\n"), nil
}
