# Paperless ngx File uploader

**This is a work in progress**. The goal is to have a rather minimal feature set to allow it to easily be integrated into other tools like FileJuggler or Hazel without having to resort to FTP upload, for example.

## Build

Currently I don't provide builds, so you're on your own to install Go, check out the repo and build the executable with `go build .`

## Usage

Create a paperless-uploader.ini in your users config directory. This varies per operating system. On Windows it is `%AppData%`, on MacOS it is `$HOME/Library/Application Support`, on Linux it is `$HOME/.config`. It should look like this:

```ini
[server]
url = "https://url-to-your-paperless-instance"
username = "your paperless username"
password = "your paperless password"
```

### Login

Alternatively, you can run `paperless-uploader --login` and it will ask you for your instance and login details and write them to the config file.

**please note that this saves the password to your instance in cleartext.**

After creating this file, you can upload files to paperless by just calling `paperless-uploader path/to/file`

You can provide more than one file and all of them are going to be uploaded.

### Tags

You can specify a tag to add to the document (And paperless-uploader will make sure that tag exists) by adding `--tag tagname` to the commandline. This allows you to tag different sources, for example.

## Status, Goals

This is a work in progress. I use it daily to automatically upload things from my Downloads folders and from my scanner's folder and it seems to work fine. I have a couple of things I want to add but I have very little time to work on this. If you want to expand on it, feel free to send PRs or just outright fork this, I don't particularly care. I have very little interest right now to turn this into something else than a very specifically made tool for my own needs.

You can see my own Todo list of sorts in the [Issues](/issues)

## License

The bit of code that is there is licensed under the MIT License (See [LICENSE](LICENSE))
