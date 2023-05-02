# SteamQuery by devusSs

## Disclaimer

DO NOT (!) USE THIS PROGRAM WITHOUT KNOWING WHAT IT DOES OR GETTING LECTURED BY THE CREATOR (devusSs).

This program is not affiliated with any services mentioned inside or outside of it.

This program may also not be safe for production use. It should be seen as a fun / site project by the developer.

Please contact the owner (devusSs) at devuscs@gmail.com to resolve any issues (especially trademarks and copyright stuff).

## Setup

Create a new raw [Google Sheet](https://docs.google.com/) and set it up like so (must not exactly match):
![sheets.jpg](./docs/googlesheet.png)
`Make sure there is a cell for last updated, individual price per item, total value of items, value difference to last run and errors each.`

The price /item column will be updated for each item specified in your config file (check below).

To calculate total values you would want to use Google's own sheet functions like:

```sh
= PRODUCT(CELL1 * CELL2)
```

or

```sh
= SUM(CELL1 + CELL2)
```

<br/>

Go to [Google Cloud](https://cloud.google.com/), set up a new project and enable the Google Drive API and the Google Sheets API. Then add a service account and generate keys in json format. Place that json file in a directory of your choice. Using the [files](files/) directory in the projects directory is recommended however while naming the json file `gcloud.json`. Make sure to add the auto generated E-Mail in your Google Cloud service account as an editor on your Google sheet.

<br/>

Go to your [Steam dev settings](https://steamcommunity.com/dev/apikey) and generate an API key. It will be needed to query the status of the Steam API / Sessions Managers / Community Status of CSGO. It will also be used to query your CSGO inventory on Steam later on.
You will also need to set your Steam ID 64 (fetch it via sites like [SteamIDUK](https://steamid.uk)) so we can query your Steam (CSGO) inventory programmatically.

<br/>

Then create a `config.json` file in a directory of your choice. Using the [files](files/) directory in the projects directory is recommended however while naming the config `config.json`.
<br/>
Use the [example config file](files/config.example.json) to create your own `config.json` file:

<br/>

## Building and running the app

Either download an already compiled program from the [releases](https://github.com/devusSs/steamquery/releases) section or clone the repository and compile the program yourself. You will need the [Go(lang)](https://go.dev) binaries for that.

You can then run the app with either default flags or use the defined flags `-c` and `-g` to set your config file path and your gcloud.json file path respectively.

Errors will usually be self-explanatory. Any weird errors may require the use of [Google](https://google.com) or [creating an issue](https://github.com/devusSs/steamquery/issues) on Github.

## Further features (soonTM)

- automatically query items from Steam inventories
- automatically query items from Google sheets to make use of config easier
