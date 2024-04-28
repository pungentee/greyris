package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyAuth "github.com/zmb3/spotify/v2/auth"
	"go.mills.io/bitcask/v2"
	"golang.org/x/oauth2"
)

// Track for store only useful data
type Track struct {
	artist           string
	albumReleaseDate string
	trackNumber      int
}

// getDB returns a bitcask.DB that locate in ~/.greyris/<name>
// don't forgot call `defer db.Close()`
func getDB(name string, new bool) (bitcask.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(homeDir, ".greyris", name)

	if new {
		err = os.RemoveAll(path)
		if err != nil {
			return nil, err
		}
	}

	db, err := bitcask.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

// getIdByLink returns id from link
func getIdByLink(link string) spotify.ID {
	cut, _ := strings.CutPrefix(link, "https://open.spotify.com/playlist/")
	return spotify.ID(cut[:22])
}

// getClientID finds client ID in db and returns client ID
// if not found, then requests it from the user and stores it in the database
func getClientID(db bitcask.DB) (string, error) {
	idKey := bitcask.Key("clientID")
	id, err := db.Get(idKey)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)

		for len(id) != 32 {
			fmt.Print("Enter your Client ID: ")
			id, err = reader.ReadBytes('\n')
			if err != nil {
				return "", err
			}

			id = id[:len(id)-1]
			if len(id) != 32 {
				fmt.Println("error: invalid Client ID (length of it must be 32 characters)")
			}
		}

		err = db.Put(idKey, id)
		if err != nil {
			return "", err
		}
	}

	return string(id), nil
}

// getClientSecret finds client Secret in db and returns client ID
// if not found, then requests it from the user and stores it in the database
func getClientSecret(db bitcask.DB) (string, error) {
	secretKey := bitcask.Key("secretID")
	secret, err := db.Get(secretKey)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)

		for len(secret) != 32 {
			fmt.Print("Enter your Client Secret: ")
			secret, err = reader.ReadBytes('\n')
			if err != nil {
				return "", err
			}

			secret = secret[:len(secret)-1]
			if len(secret) != 32 {
				fmt.Println("error: invalid Client Secret (length of it must be 32 characters)")
			}
		}

		err = db.Put(secretKey, secret)
		if err != nil {
			return "", err
		}
	}

	return string(secret), nil
}

// getAuthenticator returns configured authenticator
func getAuthenticator(clientID, clientSecret string) *spotifyAuth.Authenticator {
	return spotifyAuth.New(
		spotifyAuth.WithClientID(clientID),
		spotifyAuth.WithClientSecret(clientSecret),
		spotifyAuth.WithRedirectURL("http://localhost:8080/callback"),
		spotifyAuth.WithScopes(
			spotifyAuth.ScopePlaylistReadPrivate,
			spotifyAuth.ScopePlaylistModifyPrivate,
			spotifyAuth.ScopePlaylistModifyPublic))
}

// login returns *spotify.Client
// if a token is stored in db, then log in with it
// if token is not found in db, then using authenticate log in and store token in db
func login(db bitcask.DB) (*spotify.Client, error) {
	var token *oauth2.Token
	tokenKey := bitcask.Key("tokenJson")

	clientID, err := getClientID(db)
	if err != nil {
		return nil, err
	}

	clientSecret, err := getClientSecret(db)
	if err != nil {
		return nil, err
	}

	authenticator := getAuthenticator(clientID, clientSecret)

	// getting token from db
	tokenJson, err := db.Get(tokenKey)
	if err != nil {
		// if token not found, then log in user and store token in db
		client := authenticate(authenticator)
		token, err = client.Token()
		if err != nil {
			return nil, err
		}

		tokenJson, err = json.Marshal(token)
		if err != nil {
			return nil, err
		}

		err = db.Put(tokenKey, tokenJson)
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	err = json.Unmarshal(tokenJson, &token)
	if err != nil {
		return nil, err
	}

	// if token already stores in db just use it for get spotify.Client
	return spotify.New(authenticator.Client(context.Background(), token)), nil
}

// authenticate log in user and returns *spotify.Client
func authenticate(authenticator *spotifyAuth.Authenticator) *spotify.Client {
	// copied from https://github.com/zmb3/spotify/blob/master/examples/authenticate/authcode/authenticate.go
	ch := make(chan *spotify.Client)
	state := "abc123"

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := authenticator.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "<h1>Couldn't get token</h1>", http.StatusForbidden)
			log.Fatal(err)
		}

		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}

		client := spotify.New(authenticator.Client(r.Context(), tok))

		_, err = fmt.Fprintf(w, "<h1>Login Completed!</h1><h3>You can close this page</h3>")
		if err != nil {
			log.Fatal(err)
		}

		ch <- client
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := authenticator.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	client := <-ch

	return client
}

// getAllItemsList return list with all tracks from playlist
func getAllItemsList(client *spotify.Client, playlistID spotify.ID) (result []spotify.PlaylistItem, err error) {
	tracks, err := client.GetPlaylistItems(context.Background(), playlistID) // getting first page
	if err != nil {
		return nil, err
	}

	for {
		result = append(result, tracks.Items...)            // adding each track from page in result
		err = client.NextPage(context.Background(), tracks) // changing to next page
		if errors.Is(err, spotify.ErrNoMorePages) {         // break if it is last page
			break
		} else if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// itemsToTracks converts spotify.PlaylistItem list to Track list
func itemsToTracks(items []spotify.PlaylistItem) (tracks []Track) {
	for _, value := range items {
		track := value.Track.Track
		artistName := strings.ToLower(track.Artists[0].Name)  // convert every name to lower case for correct comparing
		artistName, _ = strings.CutPrefix(artistName, "the ") // if the prefix "the" exists, then cut it

		// adding each track to the list only with useful information
		tracks = append(tracks, Track{
			trackNumber:      track.TrackNumber,
			artist:           artistName,
			albumReleaseDate: track.Album.ReleaseDate,
		})
	}
	return
}

// sortTrackList return sorted track list
func sortTrackList(tracks []Track) []Track {
	isSorted := false
	for !isSorted {
		isSorted = true
		for i := 0; i < len(tracks)-1; i++ {
			if tracks[i].artist > tracks[i+1].artist { // if current artist greater next then swap them
				isSorted = false
				tracks[i], tracks[i+1] = tracks[i+1], tracks[i]
			} else if tracks[i].artist == tracks[i+1].artist {
				// checking album release dates
				if tracks[i].albumReleaseDate > tracks[i+1].albumReleaseDate {
					isSorted = false
					tracks[i], tracks[i+1] = tracks[i+1], tracks[i]
				} else if tracks[i].albumReleaseDate == tracks[i+1].albumReleaseDate {
					// checking tack number in album
					if tracks[i].trackNumber > tracks[i+1].trackNumber {
						isSorted = false
						tracks[i], tracks[i+1] = tracks[i+1], tracks[i]
					}
				}
			}
		}
	}

	return tracks
}

// reorderPlaylist reorders playlist by comparing initial and modified
func reorderPlaylist(client *spotify.Client, id spotify.ID, initial, modified []Track) error {
	for newIndex, value := range modified {
		oldIndex := indexOf(value, initial)
		if newIndex != oldIndex {
			_, err := client.ReorderPlaylistTracks(context.Background(), id, spotify.PlaylistReorderOptions{
				RangeStart:   oldIndex,
				InsertBefore: newIndex,
			})
			if err != nil {
				return err
			}
		}
		initial = moveElement(initial, oldIndex, newIndex) // syncing initial list with Spotify playlist
	}

	return nil
}

// indexOf returns the index of an element if it matches the argument element
// if element not found, returns a -1.
func indexOf[T comparable](element T, slice []T) int {
	for index, value := range slice {
		if element == value {
			return index
		}
	}
	return -1
}

// insert inserts element by index to slice
func insert[T any](slice []T, value T, index int) []T {
	return append(slice[:index], append([]T{value}, slice[index:]...)...)
}

// removeIndex removes from slice by index
func removeIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

// moveElement moves element
func moveElement[T any](slice []T, from int, to int) []T {
	value := slice[from]
	return insert(removeIndex(slice, from), value, to)
}

func removeValue[T comparable](l []T, item T) []T {
	out := make([]T, 0)
	for _, element := range l {
		if element != item {
			out = append(out, element)
		}
	}
	return out
}

func getAllIDsFromDB(db bitcask.DB) (IDs []string, err error) {
	err = db.ForEach(func(key bitcask.Key) error {
		link, err := db.Get(key)
		if err != nil {
			return err
		}

		IDs = append(IDs, string(link))

		return nil
	})
	if err != nil {
		return nil, err
	}

	return IDs, nil
}
