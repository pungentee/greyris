# Greyris
This console utility will help you sort your Spotify playlists

Sorting rules: by author name -> by album release date -> by track number in the album

## Requires
- Installed Go
- [Spotify App](https://developer.spotify.com/dashboard) with a Redirect URI that has the value `http://localhost:8080/callback` (Required for proper authentication)
- The playlist you want to sort must be yours

## Install
```shell
$ go install github.com/Pungentee/greyris@latest
```

## Usage
```shell
# sort playlist
$ greyris [link]
$ greyris [alias] 
$ greyris all # sorts all aliases
$ greyris [link] all # combine

# add alias
$ greyris add [name] [link]
$ greyris add [name] [link] [name] [link]... # add multiple aliases 

# print list of all aliases
$ greyris list

# remove alias
$ greyris remove [alias]
$ greyris remove [alias] [alias]... # remove multiple aliases 
```