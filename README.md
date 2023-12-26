# Greyris
This console utility will help you sort your Spotify playlists

Sorting rules: by author name -> by album release date -> by track number in the album

## Requires
- Installed Go
- [Spotify App](https://developer.spotify.com/dashboard) with a Redirect URI that has the value `http://localhost:8080/callback`
- The playlist you want to sort must be yours

## Install
```shell
$ go install github.com/Pungentee/greyris
```

## Usage
```shell
$ greyris <link to spotify playlist>
```