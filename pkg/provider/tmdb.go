package provider

import (
	"fmt"
	"github.com/KnutZuidema/go-tmdb"
	"github.com/KnutZuidema/godarr/pkg/model"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type TMDBProvider struct {
	client *tmdb.TMDb
	logger logrus.FieldLogger
	kind   model.ItemKind
}

var (
	tmdbTimeFormat = "2006-01-02"
)

func NewTMDBProvider(apiKey string, logger logrus.FieldLogger, kind model.ItemKind) *TMDBProvider {
	client := tmdb.Init(tmdb.Config{
		APIKey: apiKey,
	})
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &TMDBProvider{
		client: client,
		logger: logger.WithField("component", "TMDBProvider"),
		kind:   kind,
	}
}

func (p *TMDBProvider) ListBySearch(search string) ([]*model.Item, error) {
	switch p.kind {
	case model.ItemKindMovie:
		return p.listMovie(search)
	case model.ItemKindTVSeries:
		return p.listTV(search)
	}
	return nil, fmt.Errorf("invalid kind: %v", p.kind)
}

func (p *TMDBProvider) GetByID(id string) (*model.Item, error) {
	tmdbID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	switch p.kind {
	case model.ItemKindMovie:
		return p.getMovieByID(tmdbID)
	case model.ItemKindTVSeries:
		return p.getTVByID(tmdbID)
	}
	return nil, fmt.Errorf("invalid kind: %v", p.kind)
}

func (p *TMDBProvider) getMovieByID(id int) (*model.Item, error) {
	res, err := p.client.GetMovieInfo(id, nil)
	if err != nil {
		return nil, err
	}
	externalIDs, err := p.client.GetMovieExternalIds(id, nil)
	if err != nil {
		return nil, err
	}
	res.ExternalIDs = externalIDs
	item, err := movieItemFromTMDBMovie(res)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (p *TMDBProvider) getTVByID(id int) (*model.Item, error) {
	res, err := p.client.GetTvInfo(id, nil)
	if err != nil {
		return nil, err
	}
	externalIDs, err := p.client.GetTvExternalIds(id, nil)
	if err != nil {
		return nil, err
	}
	res.ExternalIDs = externalIDs
	item, err := tvItemFromTMDBTV(res)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func tvItemFromTMDBTVResult(tv *tmdb.TvSearchResults) ([]*model.Item, error) {
	items := make([]*model.Item, 0, len(tv.Results))
	for _, res := range tv.Results {
		release, err := time.Parse(tmdbTimeFormat, res.FirstAirDate)
		if err != nil {
			release = time.Time{}
		}
		items = append(items, &model.Item{
			ExternalID:  strconv.Itoa(res.ID),
			Kind:        model.ItemKindTVSeries,
			Title:       res.Name,
			ImagePath:   res.PosterPath,
			ReleaseYear: release.Year(),
			Rating:      float64(res.VoteAverage),
			Status:      model.ItemStatusAdded,
		})
	}
	return items, nil
}

func tvItemfromTMDBMovieResult(movie *tmdb.MovieSearchResults) ([]*model.Item, error) {
	items := make([]*model.Item, 0, len(movie.Results))
	for _, res := range movie.Results {
		release, err := time.Parse(tmdbTimeFormat, res.ReleaseDate)
		if err != nil {
			release = time.Time{}
		}
		items = append(items, &model.Item{
			ExternalID:  strconv.Itoa(res.ID),
			Kind:        model.ItemKindMovie,
			Title:       res.Title,
			ImagePath:   res.PosterPath,
			ReleaseYear: release.Year(),
			Rating:      float64(res.VoteAverage),
			Status:      model.ItemStatusAdded,
		})
	}
	return items, nil
}

func movieItemFromTMDBMovie(movie *tmdb.Movie) (*model.Item, error) {
	release, err := time.Parse(tmdbTimeFormat, movie.ReleaseDate)
	if err != nil {
		release = time.Time{}
	}
	return &model.Item{
		ExternalID:  movie.ExternalIDs.ImdbID,
		Kind:        model.ItemKindMovie,
		Title:       movie.Title,
		Description: movie.Overview,
		ImagePath:   movie.PosterPath,
		ReleaseYear: release.Year(),
		Rating:      float64(movie.VoteAverage),
		Status:      model.ItemStatusAdded,
	}, nil
}

func tvItemFromTMDBTV(tv *tmdb.TV) (*model.Item, error) {
	release, err := time.Parse(tmdbTimeFormat, tv.FirstAirDate)
	if err != nil {
		release = time.Time{}
	}
	return &model.Item{
		ExternalID:  strconv.Itoa(tv.ExternalIDs.TvdbID),
		Kind:        model.ItemKindTVSeries,
		Title:       tv.Name,
		Description: tv.Overview,
		ImagePath:   tv.PosterPath,
		ReleaseYear: release.Year(),
		Rating:      float64(tv.VoteAverage),
		Status:      model.ItemStatusAdded,
	}, nil
}

func (p *TMDBProvider) listTV(search string) ([]*model.Item, error) {
	res, err := p.client.SearchTv(search, nil)
	if err != nil {
		return nil, err
	}
	return tvItemFromTMDBTVResult(res)
}

func (p *TMDBProvider) listMovie(search string) ([]*model.Item, error) {
	res, err := p.client.SearchMovie(search, nil)
	if err != nil {
		return nil, err
	}
	return tvItemfromTMDBMovieResult(res)
}
