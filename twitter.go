package main

import (
	"fmt"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

func twit(args string, client *twitter.Client) (msg string, err error) {
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
