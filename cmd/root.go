/*
Copyright © 2023 Tymofii Kliuiev <pungentee@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go.mills.io/bitcask/v2"
	"log"
	"os"
	"strings"
)

var (
	idKey     = bitcask.Key("clientID")
	secretKey = bitcask.Key("secretID")
	logger    = log.Default()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "greyris [link to your playlist in spotify]",
	Short: "Sorts you Spotify playlists",
	Long: `This console utility will help you sort your Spotify playlists

Sorting rules: by author name -> by album release date -> by track number in the album
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
			logger.Fatal(err)
			return
		}
		defer db.Close()

		clientID, clientSecret, err := getUserInfo(db)
		if err != nil {
			logger.Fatal(err)
			return
		}

		fmt.Println(clientID, clientSecret)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// getUserInfo get Client ID and Secret from the database
// if this data isn't available, then ask it from the user
// and store the answer in the database and return it
func getUserInfo(db bitcask.DB) (clientID, clientSecret string, err error) {
	reader := bufio.NewReader(os.Stdin)

	id, err := db.Get(idKey)
	if err != nil {
		for len(id) != 32 {
			fmt.Print("Enter your Client ID: ")
			id, err = reader.ReadBytes('\n')
			if err != nil {
				return "", "", err
			}

			id = id[:len(id)-1]
			if len(id) != 32 {
				fmt.Println("error: invalid Client ID (length of it must be 32 characters)")

			}
		}

		err = db.Put(idKey, id)
		if err != nil {
			return "", "", err
		}
	}

	secret, err := db.Get(secretKey)
	if err != nil {
		for len(secret) != 32 {
			fmt.Print("Enter your Client Secret: ")
			secret, err = reader.ReadBytes('\n')
			if err != nil {
				return "", "", err
			}

			secret = secret[:len(secret)-1]
			if len(secret) != 32 {
				fmt.Println("error: invalid Client Secret (length of it must be 32 characters)")

			}
		}

		err := db.Put(secretKey, secret)
		if err != nil {
			return "", "", err
		}
	}

	return string(id), string(secret), nil
}
