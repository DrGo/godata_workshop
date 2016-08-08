package main

// This script is based on:
//   https://github.com/dghubble/go-twitter/blob/master/examples/streaming.go
//
// It requires the go-twitter library:
//   https://github.com/dghubble/go-twitter
//
// To use this script you must first configure the application with
// Twitter and obtain API keys.  The steps are as follows:
//
// 1. You must have a Twitter account that is linked to your mobile
// device.
//
// 2. Visit this page: https://apps.twitter.com/app/new. You can enter
// whatever you want in the first three fields, and leave the fourth
// field blank.  Then check the developer agreement box and press the
// button to create your application.
//
// 3. Record the consumer key and secret provided by Twitter.  Then
// click on the "Keys and access tokens" tab and press the button on
// that page to obtain a Twitter access key and secret.
//
// Once you have these keys and tokens, you can run this script using:
//   go run streaming.go -consumer-key=## -consumer-secret=## -access-token=## -access-secret=##
// replacing ## as appropriate.
//
// Configure the FILTER section as desired.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func main() {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter Client
	client := twitter.NewClient(httpClient)

	lang := make(map[string]int)

	// Demultiplex stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		fmt.Printf("%s\n", tweet.Text)
		x := tweet.Lang
		lang[x] = lang[x] + 1
		fmt.Printf("%v\n", lang)
	}

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"zika"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()
}
