package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.mills.io/bitcask/v2"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strings"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "greyris [link to your playlist in spotify]",
	Short: "Sorts you Spotify playlists",
	Long: `This console utility will help you sort your Spotify playlists

Sorting rules: by author name -> by album release date -> by track number in the album

Requires: The Redirect URI of your Spotify App should be "http://localhost:8080/callback"
`,

	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if !strings.HasPrefix(args[0], "https://open.spotify.com/playlist/") {
			return errors.New("invalid link")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		db, err := bitcask.Open("db")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer db.Close()

		client, err := getClient(db)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println(client)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

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

		err := db.Put(secretKey, secret)
		if err != nil {
			return "", err
		}
	}

	return string(secret), nil
}

func getClient(db bitcask.DB) (client *spotify.Client, err error) {
	tokenKey := bitcask.Key("tokenJson")
	var token *oauth2.Token

	id, err := getClientID(db)
	if err != nil {
		log.Fatal(err)
		return
	}

	secret, err := getClientSecret(db)
	if err != nil {
		log.Fatal(err)
		return
	}

	auth := spotifyauth.New(
		spotifyauth.WithClientID(id),
		spotifyauth.WithClientSecret(secret),
		spotifyauth.WithRedirectURL("http://localhost:8080/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))

	tokenJson, err := db.Get(tokenKey)
	if err != nil {
		client = Auth(auth)
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

	client = spotify.New(auth.Client(context.Background(), token))

	return
}

func Auth(auth *spotifyauth.Authenticator) (client *spotify.Client) {
	ch := make(chan *spotify.Client)
	state := "abc123"

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}

		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}

		client = spotify.New(auth.Client(r.Context(), tok))

		_, err = fmt.Fprintf(w, "Login Completed!")
		if err != nil {
			log.Fatal(err)
		}

		ch <- client
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	client = <-ch

	return
}
