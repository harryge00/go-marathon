/*
Copyright 2015 The go-marathon Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/gambol99/go-marathon"
	"log"
)

var marathonURL string
var timeout int

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080", "the url for the Marathon endpoint")
	flag.IntVar(&timeout, "timeout", 600, "listen to events for x seconds")
}

func assert(err error) {
	if err != nil {
		log.Fatalf("Failed, error: %s", err)
	}
}

func main() {
	flag.Parse()
	stop := make(chan struct{})
	run(stop)
}
// Workflow:
// 1. deployment_info -> 2.
//
//
func run(stop <-chan struct{}) {
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	config.EventsTransport = marathon.EventsTransportSSE
	log.Printf("Creating a client, Marathon: %s", marathonURL)

	client, err := marathon.NewClient(config)
	assert(err)

	// Register for events
	apievents, err := client.AddEventsListener(marathon.EventIDAPIRequest)
	assert(err)
	events, err := client.AddEventsListener(marathon.EventIDApplications)
	assert(err)
	deployments, err := client.AddEventsListener(marathon.EventIDDeploymentStepSuccess)
	assert(err)

	done := false
	for {
		if done {
			break
		}
		select {
		case <-stop:
			log.Printf("Exiting the loop")
			done = true
		case event:= <- apievents:
			log.Printf("Received api-post-event: %v", event)
			var app *marathon.EventAPIRequest
			app = event.Event.(*marathon.EventAPIRequest)
			log.Printf("labels: %v", *app.AppDefinition.Labels)
			log.Printf("port: %v", app.AppDefinition.PortDefinitions)
		case event := <-events:
			log.Printf("Received application event: %s", event)
		case event := <-deployments:
			log.Printf("Received deployment event: %v", event)
			var deployment *marathon.EventDeploymentStepSuccess
			deployment = event.Event.(*marathon.EventDeploymentStepSuccess)
			log.Printf("deployment step: %v", deployment.CurrentStep)
		}
	}

	log.Printf("Removing our subscription")
	client.RemoveEventsListener(events)
	client.RemoveEventsListener(apievents)
	client.RemoveEventsListener(deployments)
}