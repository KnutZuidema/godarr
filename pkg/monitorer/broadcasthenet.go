package monitorer

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/KnutZuidema/go-btn"
	"github.com/KnutZuidema/go-btn/pkg/model"
	log "github.com/sirupsen/logrus"
)

type BroadcasTheNetMonitorer struct {
	client         *btn.Client
	logger         log.FieldLogger
	output         chan<- []byte
	searchInterval time.Duration
}

func NewBroadcasTheNetMonitorer(apiKey string, logger log.FieldLogger, output chan<- []byte, interval time.Duration) *BroadcasTheNetMonitorer {
	if logger == nil {
		logger = log.StandardLogger()
	}
	return &BroadcasTheNetMonitorer{
		client:         btn.NewClient(http.DefaultClient, apiKey),
		logger:         logger.WithField("component", "BroadcasTheNetMonitorer"),
		output:         output,
		searchInterval: interval,
	}
}

func (m *BroadcasTheNetMonitorer) Monitor(tvdbID string) (err error) {
	var torrent model.Torrent
	ticker := time.NewTicker(m.searchInterval)
	defer ticker.Stop()
	for range ticker.C {
		torrents, err := m.client.SearchTorrents(model.SearchTorrentOptions{
			TVDbID: tvdbID,
		}, 1, 0)
		if err != nil {
			return err
		}
		if len(torrents) > 0 {
			torrent = torrents[0]
			break
		}
	}
	resp, err := http.Get(torrent.DownloadURL)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			err = e
		}
	}()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	m.output <- buf
	return nil
}
