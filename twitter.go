package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	twitterscraper "github.com/n0madic/twitter-scraper"
)

var (
	tweetUrlRe = regexp.MustCompile(tweetURLRegex)
)

func twit(args string, scraper *twitterscraper.Scraper) (msg string, err error) {
	fields := strings.Fields(args)

	if len(fields) == 0 {
		return "Error: pass a user", nil
	}

	arg := strings.Fields(args)[0]

	for tweet := range scraper.GetTweets(context.Background(), arg, 1) {
		if tweet.Error != nil {
			return "", tweet.Error
		}

		return tweet.Text, nil
	}

	return fmt.Sprintf("User %s has no tweets", arg), nil
}

func tweetFromUrl(args string, scraper *twitterscraper.Scraper) (msg string, err error) {
	urls := tweetUrlRe.FindAllString(args, -1)

	if len(urls) == 0 {
		return "", errors.New("no tweet urls found in string")
	}

	outList := []string{}

	for _, u := range urls {
		up, err := url.Parse("https://" + u)

		if err != nil {
			outList = append(outList, fmt.Sprintf("Could not parse tweet with url %s", u))
			continue
		}

		us := strings.Split(up.Path, "/")
		id := us[len(us)-1]

		tweet, err := scraper.GetTweet(id)

		if err != nil {
			outList = append(outList, fmt.Sprintf("Could not find tweet with url %s", u))
			continue
		}

		outList = append(outList, fmt.Sprintf("{cyan}Tweet from %s:{clear} %s", tweet.Name, tweet.Text))
	}

	return strings.Join(outList, "\n"), nil
}
