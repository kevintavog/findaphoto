package main

import (
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/helpers"
	"github.com/kevintavog/findaphoto/indexer/steps"
	"github.com/kevintavog/findaphoto/indexer/steps/checkindex"
	"github.com/kevintavog/findaphoto/indexer/steps/checkthumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/generatethumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/getexif"
	"github.com/kevintavog/findaphoto/indexer/steps/indexmedia"
	"github.com/kevintavog/findaphoto/indexer/steps/resolveplacename"
	"github.com/kevintavog/findaphoto/indexer/steps/scanner"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

func main() {
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphotoindexer")

	if !common.IsExecWorking(common.ExifToolPath, "-ver") {
		log.Fatal("exiftool isn't usable (path is '%s')", common.ExifToolPath)
	}
	if !common.IsExecWorking(common.FfmpegPath, "-version") {
		log.Fatal("ffmpeg isn't usable (path is '%s')", common.FfmpegPath)
	}
	generatethumbnail.VipsExists = common.IsExecWorking(common.VipsThumbnailPath, "--vips-version")

	app := cli.App("indexer", "The FindAPhoto indexer")
	app.Spec = "-p -s -r -l [-a] [-i] [--reindex] [-v]"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index (for development) (optional)")
	scanPath := app.StringOpt("p path", "", "The path to recursively index")
	server := app.StringOpt("s server", "", "The URL for the ElasticSearch server")
	redisServer := app.StringOpt("r", "", "The URL for the Redis server")
	locationLookupUrl := app.StringOpt("l", "", "The URL for the location lookup (ReverseNameLookup)")
	forceIndex := app.BoolOpt("reindex", false, "Force everything to be re-indexed; current index not deleted. (optional)")
	aliasPathOverride := app.StringOpt("a", "", "The alias path override, for development")
	app.Version("v", "Show the version and exit")
	app.Action = func() {

		// common.IndexMakeNoChanges = true
		common.MediaIndexName = *indexPrefix + common.MediaIndexName
		common.RedisServer = *redisServer
		common.AliasPathOverride = *aliasPathOverride

		log.Info("%s: FindAPhoto scanning %s, and indexing to %s/%s; using %d/%d CPU's",
			time.Now().Format("2006-01-02"),
			*scanPath,
			*server,
			common.MediaIndexName,
			runtime.NumCPU(),
			runtime.GOMAXPROCS(0))
		log.Info("Using %s to resolve locations to placename", *locationLookupUrl)

		if common.IndexMakeNoChanges {
			log.Info("NOT making any changes")
		}

		checkindex.ForceIndex = *forceIndex
		if checkindex.ForceIndex {
			log.Warn("Re-indexing all documents")
		}

		if !common.IndexMakeNoChanges {
			c, err := redis.DialURL(*redisServer)
			if err != nil {
				log.Fatalf("Unable to connect to Redis server: %s [%s]", err, *redisServer)
			}
			defer c.Close()
		}

		common.ElasticSearchServer = *server
		resolveplacename.LocationLookupUrl = *locationLookupUrl

		checkServerAndIndex()
		alias, err := common.AliasForPath(*scanPath)
		if err != nil {
			log.Fatalf("Unable to get alias for '%s': %s", *scanPath, err.Error())
		}

		if !generatethumbnail.VipsExists {
			log.Warn("Unable to use the 'vipsthumbnails' command, defaulting to slower slide generation (path is '%s')", common.VipsThumbnailPath)
		}

		scanStartTime := time.Now()
		helpers.InitializeDuplicates()
		classifymedia.Start()
		scanner.Scan(*scanPath, alias)
		scanDuration := time.Now().Sub(scanStartTime).Seconds()
		emitStats(scanDuration)

		if !common.IndexMakeNoChanges {
			err = common.UpdateLastIndexed(alias)
			if err != nil {
				log.Warn("Failed updating indexed date: '%s'", err.Error())
			}
		}
	}

	app.Run(os.Args)
}

func emitStats(seconds float64) {
	filesPerSecond := float64(scanner.SupportedFilesFound) / seconds

	log.Info("[%01.3f seconds, %01.2f files/second], Scanned %d folders and %d files, found %d supported files.",
		seconds, filesPerSecond,
		scanner.DirectoriesScanned, scanner.FilesScanned, scanner.SupportedFilesFound)

	log.Info("%d failed repository checks, %d badly formatted json responses, %d failed signatures",
		checkindex.BadJson, checkindex.CheckFailed, checkindex.SignatureGenerationFailed)

	log.Info("%d exiftool invocations, %d failed",
		getexif.ExifToolInvocations, getexif.ExifToolFailed)

	log.Info("%d locations lookup attempts, %d location lookup failures, %d server errors, %d other failures",
		resolveplacename.PlacenameLookups, resolveplacename.FailedLookups, resolveplacename.ServerErrors, resolveplacename.Failures)

	log.Info("%d image thumbnails created, %d failed; %d video thumbnails created, %d failed; %d failed thumbnail checks",
		generatethumbnail.GeneratedImage, generatethumbnail.FailedImage, generatethumbnail.GeneratedVideo, generatethumbnail.FailedVideo, checkthumbnail.FailedChecks)

	log.Info("%d files indexed, %d duplicates ignored, %d failed and %d added due to detected changes",
		indexmedia.IndexedFiles, helpers.DuplicatesIgnored, indexmedia.FailedIndexAttempts, indexmedia.ChangedFiles)

	log.Info("%d media scanned, %d removed from the index",
		scanner.MediaScanned, scanner.MediaRemoved)
}

func checkServerAndIndex() {
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(common.ElasticSearchServer),
		elastic.SetSniff(false))

	if err != nil {
		log.Fatal("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}

	exists, err := client.IndexExists(common.MediaIndexName).Do(context.TODO())
	if err != nil {
		log.Fatal("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Fatal("The index '%s' doesn't exist", common.MediaIndexName)
	}
	exists, err = client.IndexExists(common.AliasIndexName).Do(context.TODO())
	if err != nil {
		log.Fatal("Failed querying index: %s", err.Error())
	}
	if !exists {
		log.Fatal("The index '%s' doesn't exist", common.AliasIndexName)
	}

	err = common.InitializeAliases(client)
	if err != nil {
		log.Fatal("Failed initializing aliases: %s", err.Error())
	}
}
