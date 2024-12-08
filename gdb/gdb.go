// Package gdb interacts with thegamedb.net's API.
package gdb

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	gamesdb "github.com/J-Swift/thegamesdb-swagger-client-go"
)

var apiClient = gamesdb.NewAPIClient(gamesdb.NewConfiguration())

// Publishers

type publishers struct {
	mux        sync.Mutex
	publishers *map[string]gamesdb.Publisher
}

var publishersCache = publishers{}

func getPublishers(ctx context.Context, apikey string) map[string]gamesdb.Publisher {
	pubs, resp, err := apiClient.PublishersApi.PublishersExecute(
		apiClient.PublishersApi.Publishers(ctx).
			Apikey(apikey),
	)

	if err != nil || resp.StatusCode != 200 {
		return make(map[string]gamesdb.Publisher)
	}

	return pubs.Data.Publishers
}

func getCachedPublishers(ctx context.Context, apikey string) map[string]gamesdb.Publisher {
	publishers := publishersCache.publishers
	if publishers != nil {
		return *publishers
	}

	publishersCache.mux.Lock()
	defer publishersCache.mux.Unlock()

	publishers = publishersCache.publishers
	if publishers == nil {
		apiPublishers := getPublishers(ctx, apikey)
		publishers = &apiPublishers
		publishersCache.publishers = publishers
	}

	return *publishers
}

// Developers

type developers struct {
	mux        sync.Mutex
	developers *map[string]gamesdb.Developer
}

var developersCache = developers{}

func getDevelopers(ctx context.Context, apikey string) map[string]gamesdb.Developer {
	pubs, resp, err := apiClient.DevelopersApi.DevelopersExecute(
		apiClient.DevelopersApi.Developers(ctx).
			Apikey(apikey),
	)
	if err != nil || resp.StatusCode != 200 {
		return make(map[string]gamesdb.Developer)
	}

	return pubs.Data.Developers
}

func getCachedDevelopers(ctx context.Context, apikey string) map[string]gamesdb.Developer {
	developers := developersCache.developers
	if developers != nil {
		return *developers
	}

	developersCache.mux.Lock()
	defer developersCache.mux.Unlock()

	developers = developersCache.developers
	if developers == nil {
		apiDevelopers := getDevelopers(ctx, apikey)
		developers = &apiDevelopers
		developersCache.developers = developers
	}

	return *developers
}

// Genres

type genres struct {
	mux    sync.Mutex
	genres *map[string]gamesdb.Genre
}

var genresCache = genres{}

func getGenres(ctx context.Context, apikey string) map[string]gamesdb.Genre {
	pubs, resp, err := apiClient.GenresApi.GenresExecute(
		apiClient.GenresApi.Genres(ctx).
			Apikey(apikey),
	)

	if err != nil || resp.StatusCode != 200 {
		return make(map[string]gamesdb.Genre)
	}

	return pubs.Data.Genres
}

func getCachedGenres(ctx context.Context, apikey string) map[string]gamesdb.Genre {
	genres := genresCache.genres
	if genres != nil {
		return *genres
	}

	genresCache.mux.Lock()
	defer genresCache.mux.Unlock()

	genres = genresCache.genres
	if genres == nil {
		apiGenres := getGenres(ctx, apikey)
		genres = &apiGenres
		genresCache.genres = genres
	}

	return *genres
}

// ParsedDeveloper is a normalized GamesDB Developer
type ParsedDeveloper struct {
	ID   int
	Name string
}

func toParsedDeveloper(apiDeveloper gamesdb.Developer) ParsedDeveloper {
	return ParsedDeveloper{
		ID:   int(apiDeveloper.Id),
		Name: apiDeveloper.Name,
	}
}

// ParsedGenre is a normalized GamesDB Genre
type ParsedGenre struct {
	ID   int
	Name string
}

func toParsedGenre(apiGenre gamesdb.Genre) ParsedGenre {
	return ParsedGenre{
		ID:   int(apiGenre.Id),
		Name: apiGenre.Name,
	}
}

// ParsedPublisher is a normalized GamesDB Publisher
type ParsedPublisher struct {
	ID   int
	Name string
}

func toParsedPublisher(apiPublisher gamesdb.Publisher) ParsedPublisher {
	return ParsedPublisher{
		ID:   int(apiPublisher.Id),
		Name: apiPublisher.Name,
	}
}

// ParsedGameImage is a normalized GamesDB GameImage
type ParsedGameImage struct {
	ID       int
	Type     string
	Side     string
	Filename string
}

func toParsedGameImage(apiGameImage gamesdb.GameImage) ParsedGameImage {
	return ParsedGameImage{
		ID:       int(apiGameImage.GetId()),
		Type:     apiGameImage.GetType(),
		Side:     apiGameImage.GetSide(),
		Filename: apiGameImage.GetFilename(),
	}
}

// ParsedImageSizeBaseUrls is a normalized GamesDB ImageBaseUrlMeta
type ParsedImageSizeBaseUrls struct {
	Original string
	Thumb    string
}

func toParsedImageSizeBaseUrls(apiBaseURLMeta gamesdb.ImageBaseUrlMeta) ParsedImageSizeBaseUrls {
	return ParsedImageSizeBaseUrls{
		Original: apiBaseURLMeta.Original,
		Thumb:    apiBaseURLMeta.Thumb,
	}
}

// ParsedGame is  a normalized GamesDB Game
type ParsedGame struct {
	ID          int
	Name        string
	ReleaseDate string
	//Platform    int
	Players    int
	Overview   string
	Developers []ParsedDeveloper
	Genres     []ParsedGenre
	Publishers []ParsedPublisher

	Images        map[string][]ParsedGameImage
	ImageBaseUrls ParsedImageSizeBaseUrls
}

// GetGame gets the game information from the DB.
func GetGame(ctx context.Context, apikey string, gameID string) (*ParsedGame, error) {
	var resp *http.Response
	var err error

	// TODO(jpr): remove unneeded fields
	//fields := "players,publishers,genres,overview,last_updated,rating,platform,coop,youtube,os,processor,ram,hdd,video,sound,alternates"
	fields := "players,publishers,genres,overview,platform"

	if gameID == "" {
		return nil, fmt.Errorf("must provide an ID or Name")
	}
	games, resp, err := apiClient.GamesApi.GamesByGameIDExecute(
		apiClient.GamesApi.GamesByGameID(ctx).
			Apikey(apikey).
			Id(gameID).
			Fields(fields),
	)

	if err != nil {
		return nil, fmt.Errorf("getting game url:%s, error:%s", resp.Request.URL, err)
	}

	if len(games.Data.Games) == 0 {
		return nil, fmt.Errorf("game not found")
	}

	apiGame := games.Data.Games[0]
	res := &ParsedGame{
		ID:          int(apiGame.GetId()),
		Name:        apiGame.GetGameTitle(),
		ReleaseDate: apiGame.GetReleaseDate(),
		Players:     int(apiGame.GetPlayers()),
		Overview:    apiGame.GetOverview(),
	}

	allGenres := getCachedGenres(ctx, apikey)
	genres := []ParsedGenre{}
	for _, genreID := range apiGame.Genres {
		if apiGenre, ok := allGenres[strconv.Itoa(int(genreID))]; ok {
			genres = append(genres, toParsedGenre(apiGenre))
		}
	}
	res.Genres = genres

	allDevelopers := getCachedDevelopers(ctx, apikey)
	developers := []ParsedDeveloper{}
	for _, developerID := range apiGame.Developers {
		if apiDeveloper, ok := allDevelopers[strconv.Itoa(int(developerID))]; ok {
			developers = append(developers, toParsedDeveloper(apiDeveloper))
		}
	}
	res.Developers = developers

	allPublishers := getCachedPublishers(ctx, apikey)
	publishers := []ParsedPublisher{}
	for _, publisherID := range apiGame.Publishers {
		if apiPublisher, ok := allPublishers[strconv.Itoa(int(publisherID))]; ok {
			publishers = append(publishers, toParsedPublisher(apiPublisher))
		}
	}
	res.Publishers = publishers

	images, _, err := apiClient.GamesApi.GamesImagesExecute(
		apiClient.GamesApi.GamesImages(ctx).
			Apikey(apikey).
			GamesId(strconv.Itoa(res.ID)),
	)
	if err == nil {
		res.ImageBaseUrls = toParsedImageSizeBaseUrls(images.Data.BaseUrl)

		parsedImages := make(map[string][]ParsedGameImage)
		for key, val := range images.Data.Images {
			result := parsedImages[key]
			for _, image := range val {
				result = append(result, toParsedGameImage(image))
			}
			parsedImages[key] = result
		}
		res.Images = parsedImages
	}

	return res, nil
}

// IsUp returns if thegamedb.net is up.
func IsUp(ctx context.Context, apikey string) bool {
	_, resp, err := apiClient.GamesApi.GamesByGameIDExecute(
		apiClient.GamesApi.GamesByGameID(ctx).
			Apikey(apikey).
			Id("1"),
	)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}
