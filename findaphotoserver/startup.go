package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ian-kent/go-log/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/api"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/files"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
)

func run(devolopmentMode bool, indexOverride string, aliasOverride string) {
	listenPort := 2000
	easyExit := false
	skipMediaClassifier := false
	api.FindAPhotoVersionNumber = versionString()
	log.Info("FindAPhoto %s", api.FindAPhotoVersionNumber)

	if !common.IsExecWorking(common.ExifToolPath, "-ver") {
		log.Fatalf("exiftool isn't usable (path is '%s')", common.ExifToolPath)
	}
	if !common.IsExecWorking(common.FfmpegPath, "-version") {
		log.Fatalf("ffmpeg isn't usable (path is '%s')", common.FfmpegPath)
	}

	if devolopmentMode {
		fmt.Println("*** Using development mode ***")
		// common.MediaIndexName = "dev-" + common.MediaIndexName
		listenPort = 5000
		easyExit = true
		skipMediaClassifier = true
		if len(aliasOverride) > 0 {
			common.AliasPathOverride = aliasOverride
		}
		common.IndexMakeNoChanges = true
	} else {
		if !common.IsExecWorking(common.IndexerPath, "-v") {
			log.Fatalf("The FindAPhoto Indexer isn't usable (path is '%s')", common.IndexerPath)
		}
		if !common.IsExecWorking(common.MediaClassifierPath, "-v") {
			log.Fatalf("The FindAPhoto Media Classifier isn't usable (path is '%s')", common.MediaClassifierPath)
		}
	}

	if len(indexOverride) > 0 {
		common.MediaIndexName = indexOverride
		fmt.Printf("*** Using index %s ***\n", common.MediaIndexName)
	}

	log.Info("Listening at http://localhost:%d/", listenPort)
	log.Info(" ElasticSearch:: %s/%s", configuration.Current.ElasticSearchURL, common.MediaIndexName)
	log.Info(" Reverse name lookups: %s", configuration.Current.LocationLookupURL)

	common.ElasticSearchServer = configuration.Current.ElasticSearchURL

	checkElasticServerAndIndex()
	checkLocationLookupServer()

	e := configureEcho()

	api.ReindexMedia = func(force bool) {
		go runIndexer(force, false)
	}

	api.ConfigureRouting(e)
	files.ConfigureRouting(e)

	mediaClassifierFunc := func() {
		if !skipMediaClassifier {
			runMediaClassifier(devolopmentMode)
		} else {
			log.Info("Skipping media classifier")
		}
	}

	delayThenIndexFunc := func() {
		if !devolopmentMode {
			time.Sleep(1 * time.Second)
			runIndexer(false, devolopmentMode)
		}
	}

	startServerFunc := func() {
		go mediaClassifierFunc()
		go delayThenIndexFunc()

		err := e.Start(fmt.Sprintf(":%d", listenPort))
		if err != nil {
			log.Fatalf("Failed starting the service: %s", err.Error())
		}
	}

	if !configuration.Current.VipsExists {
		log.Warn("Unable to use the 'vipsthumbnails' command, defaulting to slower slide generation (path is '%s')", common.VipsThumbnailPath)
	}

	if easyExit {
		go startServerFunc()

		fmt.Println("Hit enter to exit")
		var input string
		fmt.Scanln(&input)
	} else {
		startServerFunc()
	}
}

func configureEcho() *echo.Echo {
	e := echo.New()
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return h(util.NewFpContext(c))
		}
	})

	e.Use(fpLogger())

	e.Use(util.Recover())
	e.Use(middleware.CORS())
	e.HidePort = true
	e.HideBanner = true
	return e
}

func fpLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fc := c.(*util.FpContext)
			if err := next(c); err != nil {
				c.Error(err)
			}
			fc.RequestComplete()
			return nil
		}
	}
}

func checkElasticServerAndIndex() {
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(common.ElasticSearchServer),
		elastic.SetSniff(false))

	if err != nil {
		log.Fatalf("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}

	exists, err := client.IndexExists(common.MediaIndexName).Do(context.TODO())
	if err != nil {
		log.Fatalf("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Warn("The index '%s' doesn't exist", common.MediaIndexName)
		err = common.CreateMediaIndex(client)
		if err != nil {
			log.Fatalf("Failed creating index '%s': %+v", common.MediaIndexName, err.Error())
		}
	}

	exists, err = client.IndexExists(common.AliasIndexName).Do(context.TODO())
	if err != nil {
		log.Fatalf("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Warn("The index '%s' doesn't exist", common.AliasIndexName)
		err = common.CreateAliasIndex(client)
		if err != nil {
			log.Fatalf("Failed creating index '%s': %+v", common.AliasIndexName, err.Error())
		}
	}

	err = common.InitializeAliases(client)
	if err != nil {
		log.Fatalf("Failed initializing aliases: %s", err.Error())
	}

	exists, err = client.IndexExists(common.ClarifaiCacheIndexName).Do(context.TODO())
	if err != nil {
		log.Fatalf("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Warn("The index '%s' doesn't exist", common.ClarifaiCacheIndexName)
		err = common.CreateClarifaiClassifyIndex(client)
		if err != nil {
			log.Fatalf("Failed creating index '%s': %+v", common.ClarifaiCacheIndexName, err.Error())
		}
	}
}

func checkLocationLookupServer() {
	url := fmt.Sprintf("%s/api/v1/name?lat=%f&lon=%f", configuration.Current.LocationLookupURL, 47.6216, -122.348133)

	_, err := http.Get(url)
	if err != nil {
		log.Fatalf("The location lookup server values seem to be wrong, a location lookup failed: %s", err.Error())
	}
}
