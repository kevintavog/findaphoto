package main

import (
	"os/exec"

	"github.com/ian-kent/go-log/log"
	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
)

func runMediaClassifier(devMode bool) {
	log.Info("Starting media-classifier")

	var args = []string{
		"-e", configuration.Current.ElasticSearchUrl,
		"-r", configuration.Current.RedisUrl,
		"-a", configuration.Current.ClarifaiApiKey}

	if devMode {
		args = append(args, "-i")
		args = append(args, "dev-")
	}

	err := exec.Command(common.MediaClassifierPath, args...).Run()
	if err != nil {
		log.Fatalf("Failed executing media-classifier for: %s", err.Error())
	}
}
