package generatethumbnail

import (
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"github.com/nfnt/resize"
	"github.com/twinj/uuid"
)

var GeneratedImage int64
var FailedImage int64
var GeneratedVideo int64
var FailedVideo int64
var ThumbnailsCreated int64
var VipsExists bool

type ThumbnailInfo struct {
	FullPath    string
	AliasedPath string
	MimeType    string
}

const thumbnailMaxHeightDimension = 170

var queue chan *ThumbnailInfo
var waitGroup sync.WaitGroup

func Start() {
	ratio := 1.0
	if !VipsExists {
		ratio = 0.5
	}
	numConsumers := common.RatioNumCpus(float32(ratio))
	queue = make(chan *ThumbnailInfo, 10000)
	waitGroup.Add(numConsumers)

	log.Info("Thumbnail generation using %d consumers", numConsumers)
	for idx := 0; idx < numConsumers; idx++ {
		go func() {
			dequeue()
			waitGroup.Done()
		}()
	}
}

func Done() {
	close(queue)
}

func Wait() {
	waitGroup.Wait()
}

func Enqueue(fullPath, aliasedPath, mimeType string) {
	thumbnailInfo := &ThumbnailInfo{
		FullPath:    fullPath,
		AliasedPath: aliasedPath,
		MimeType:    mimeType,
	}
	queue <- thumbnailInfo
}

func dequeue() {
	var mediaType []string
	var thumbPath string
	var err error

	for thumbnailInfo := range queue {
		mediaType = strings.Split(thumbnailInfo.MimeType, "/")
		thumbPath = common.ToThumbPath(thumbnailInfo.AliasedPath)
		if len(mediaType) < 1 {
			log.Error("Invalid media type: '%s' for %s", thumbnailInfo.MimeType, thumbnailInfo.FullPath)
			continue
		}

		err = common.CreateDirectory(path.Dir(thumbPath))
		if err != nil {
			log.Error("Unable to create directory for '%s'", thumbPath)
			continue
		}

		switch strings.ToLower(mediaType[0]) {
		case "video":
			generateVideo(thumbnailInfo.FullPath, thumbPath)
		case "image":
			generateImage(thumbnailInfo.FullPath, thumbPath)
		default:
			log.Error("Unhandled mediaType: %s (%s) for %s", thumbnailInfo.MimeType, mediaType, thumbnailInfo.FullPath)
		}

		atomic.AddInt64(&ThumbnailsCreated, 1)
		if ThumbnailsCreated%500 == 0 {
			log.Info("Generated thumbnail %d [%s]", ThumbnailsCreated, thumbnailInfo.FullPath)
		}
	}
}

func generateImage(fullPath, thumbPath string) {

	var err error
	if VipsExists {
		err = createVipsThumbnails(fullPath, thumbPath)
	} else {
		err = createNfntThumbnail(fullPath, thumbPath)
	}

	if err != nil {
		log.Error("Failed thumbnail generation on %s: %s", fullPath, err.Error())
		atomic.AddInt64(&FailedImage, 1)
	} else {
		atomic.AddInt64(&GeneratedImage, 1)
	}
}

func generateVideo(fullPath, thumbPath string) {
	tmpFilename := path.Join(os.TempDir(), "findAPhoto", "thumbnails", uuid.NewV4().String()+".JPG")
	defer os.Remove(tmpFilename)

	err := common.CreateDirectory(path.Dir(tmpFilename))
	if err != nil {
		log.Error("Unable to create temporary directory for thumbnail generation (%s): %s", tmpFilename, err.Error())
	}

	out, err := exec.Command(common.FfmpegPath, "-i", fullPath, "-ss", "00:00:01.0", "-vframes", "1", tmpFilename).Output()
	if err != nil {
		atomic.AddInt64(&FailedVideo, 1)
		log.Error("Failed executing ffmpeg for '%s': %s (%s)", fullPath, err.Error(), out)
	}

	if exists, _ := common.PathExists(tmpFilename); !exists {
		// The video may not be long enough to grab a frame at the 1 second, so try the first frame
		out, err = exec.Command(common.FfmpegPath, "-i", fullPath, "-ss", "00:00:00.0", "-vframes", "1", tmpFilename).Output()
		if err != nil {
			atomic.AddInt64(&FailedVideo, 1)
			log.Error("Failed executing ffmpeg for '%s': %s (%s)", fullPath, err.Error(), out)
		}
	}

	if err := createNfntThumbnail(tmpFilename, thumbPath); err != nil {
		log.Error("Failed thumbnail generation on %s: %s", tmpFilename, err.Error())
		atomic.AddInt64(&FailedVideo, 1)
	} else {
		atomic.AddInt64(&GeneratedVideo, 1)
	}
}

func createNfntThumbnail(imageFilename, thumbFilename string) error {
	file, err := os.Open(imageFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	thumb := resize.Resize(0, thumbnailMaxHeightDimension, image, resize.NearestNeighbor)
	savedThumbnailFile, err := os.Create(thumbFilename)
	if err != nil {
		return err
	}
	defer savedThumbnailFile.Close()

	jpeg.Encode(savedThumbnailFile, thumb, &jpeg.Options{Quality: 85})
	return nil
}

func createVipsThumbnails(imageFilename, thumbFilename string) error {
	_, err := exec.Command(common.VipsThumbnailPath, "-d", "-s", "10000x"+strconv.Itoa(thumbnailMaxHeightDimension), "-f", thumbFilename+"[optimize_coding,strip]", imageFilename).Output()
	return err
}
