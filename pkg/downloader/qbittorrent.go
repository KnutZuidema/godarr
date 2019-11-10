package downloader

import (
	"bytes"
	"time"

	"github.com/KnutZuidema/go-qbittorrent"
	"github.com/KnutZuidema/go-qbittorrent/pkg/model"
	"github.com/anacrolix/torrent/metainfo"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type QBitTorrentDownloader struct {
	logger        log.FieldLogger
	output        chan<- string
	client        *qbittorrent.Client
	checkInterval time.Duration
}

func NewQBitTorrentDownloader(username, password, url string, logger log.FieldLogger, output chan<- string, interval time.Duration) (*QBitTorrentDownloader, error) {
	if logger == nil {
		logger = log.StandardLogger()
	}
	client := qbittorrent.NewClient(url, logger)
	if err := client.Login(username, password); err != nil {
		return nil, err
	}
	return &QBitTorrentDownloader{
		logger:        logger.WithField("component", "QBitTorrentDownloader"),
		client:        client,
		output:        output,
		checkInterval: interval,
	}, nil
}

func (d QBitTorrentDownloader) Download(file []byte) error {
	info, err := metainfo.Load(bytes.NewBuffer(file))
	if err != nil {
		return err
	}
	if err := d.client.Torrent.AddFiles(map[string][]byte{uuid.NewV4().String(): file}, &model.AddTorrentsOptions{
		Category: "godarr",
	}); err != nil {
		return err
	}
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()
	for range ticker.C {
		res, err := d.client.Torrent.GetProperties(info.HashInfoBytes().String())
		if err != nil {
			return err
		}
		if res.PiecesHave == res.PiecesNum {
			d.output <- res.SavePath
			break
		}
	}
	return nil
}
