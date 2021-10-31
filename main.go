package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/jessevdk/go-flags"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Options struct {
	Prefix         string `short:"P" long:"prefix" env:"GOWON_PREFIX" default:"." description:"prefix for commands"`
	Broker         string `short:"b" long:"broker" env:"GOWON_BROKER" default:"localhost:1883" description:"mqtt broker"`
	ConsumerKey    string `short:"c" long:"consumer-key" env:"GOWON_TWITTER_CONSUMER_KEY" required:"true" description:"twitter consumer key"`
	ConsumerSecret string `short:"C" long:"consumer-secret" env:"GOWON_TWITTER_CONSUMER_SECRET" required:"true" description:"twitter consumer secret"`
	AccessToken    string `short:"a" long:"access-token" env:"GOWON_TWITTER_ACCESS_TOKEN" required:"true" description:"twitter access token"`
	AccessSecret   string `short:"A" long:"access-secret" env:"GOWON_TWITTER_ACCESS_SECRET" required:"true" description:"twitter access secret"`
}

const (
	moduleName               = "twitter"
	mqttConnectRetryInternal = 5
	mqttDisconnectTimeout    = 1000
	tweetURLRegex            = `(http(s)?:\/\/.)?(www\.)?twitter.com/\w+/status/\d+`
)

func genTwitterHandler(client *twitter.Client) func(m gowon.Message) (string, error) {
	return func(m gowon.Message) (string, error) {
		return twit(m.Args, client)
	}
}

func genTweetFromUrlHandler(client *twitter.Client) func(m gowon.Message) (string, error) {
	return func(m gowon.Message) (string, error) {
		return tweetFromUrl(m.Args, client)
	}
}

func defaultPublishHandler(c mqtt.Client, msg mqtt.Message) {
	log.Printf("unexpected message:  %s\n", msg)
}

func onConnectionLostHandler(c mqtt.Client, err error) {
	log.Println("connection to broker lost")
}

func onRecconnectingHandler(c mqtt.Client, opts *mqtt.ClientOptions) {
	log.Println("attempting to reconnect to broker")
}

func onConnectHandler(c mqtt.Client) {
	log.Println("connected to broker")
}

func main() {
	log.Printf("%s starting\n", moduleName)

	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal(err)
	}

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(fmt.Sprintf("gowon_%s", moduleName))
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	mqttOpts.DefaultPublishHandler = defaultPublishHandler
	mqttOpts.OnConnectionLost = onConnectionLostHandler
	mqttOpts.OnReconnecting = onRecconnectingHandler
	mqttOpts.OnConnect = onConnectHandler

	config := oauth1.NewConfig(opts.ConsumerKey, opts.ConsumerSecret)
	token := oauth1.NewToken(opts.AccessToken, opts.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	mr := gowon.NewMessageRouter()
	mr.AddCommand("twitter", genTwitterHandler(client))
	mr.AddRegex(tweetURLRegex, genTweetFromUrlHandler(client))
	mr.Subscribe(mqttOpts, moduleName)

	log.Print("connecting to broker")

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Print("connected to broker")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Println("signal caught, exiting")
	c.Disconnect(mqttDisconnectTimeout)
	log.Println("shutdown complete")
}
