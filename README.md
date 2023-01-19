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

Later on there will be a function to create this file through a dialog, but right now it is not possible to do that

After creating this file, you can upload files to paperless by just calling `paperless-uploader path/to/file`

## Status, Goals

This is a work in progress. I use it daily to automatically upload things from my Downloads folders and from my scanner's folder and it seems to work fine. I have a couple of things I want to add (for example tagging so that I automatically set tags for different sources) but I have very little time to work on this. If you want to expand on it, feel free to send PRs or just outright fork this, I don't particularly care. I have very little interest right now to turn this into something else than a very specifically made tool for my own needs.

You can see my own Todo list of sorts in the [Issues](/issues)

## License

The bit of code that is there is licensed under the MIT License (See [LICENSE](LICENSE))
