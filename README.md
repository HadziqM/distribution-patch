## Quick pacth for using rain-rust-bot for latest erupe build

1. fix the distribution flow by scanning new query from bot at interval
2. transfom the query to match newer structure

### Building

#### Using Normal Methode

has go and clone this repo

```sh
go build main.go
```
#### Using Nix

use flake to install project into nix ecosystem

```sh
nix profile install github:HadziqM/distribution-patch
```
