# PGet

## Building

* Install go tooling
    * https://golang.org
    * Or use *brew install go*
* Execute *go build*

## Configuration

Requires a configuration file named "pget.json" with customer_id and pin in one of the following locations:

* /etc/pget/
* $HOME/.pget/
* ./

**Example:**

```json
{
  "customer_id": "",
  "pin": ""
}
```

## Usage 

Using *--help* on commands gives you further options

```
usage: pget [<flags>] <command> [<args> ...]

Premiumize Get

Flags:
  --help   Show context-sensitive help (also try --help-long and --help-man).
  --debug  Dump parsed premiumize.me responses

Commands:
  help [<command>...]
    Show help.

  list
    List torrents

  tree [<name>]
    Print tree of the torrent files

  download [<flags>] [<name>]
    Downloads the content of a given torrent

  upload [<link>]
    Upload a torrent file or magnet link

  watch [<flags>]
    Watch for local or remote files to upload/download
```

## Example

The following will

* Upload torrent or magnet files from the ./upload directory (--upload)
* Download finished downloads to the ./download directory (--download)
* Ignore directory hierarchies (--flatten)
* Only download video files (--video-only)
* Only load stuff that has previously been uploaded (--strict)
 
```bash
./pget watch --upload upload --download download --video-only --flatten --strict 
```