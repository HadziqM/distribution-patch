## Quick patch for using rain-rust-bot for latest erupe build

1. fix the distribution flow by scanning new query from bot at interval
2. transform the query to match newer structure

### Building

#### Using Normal Method

has go and clone this repo

```sh
go build main.go
```
#### Using Nix

use flake to install project into nix ecosystem

```sh
nix profile install github:HadziqM/distribution-patch
```

#### Usage

```sh
distribution-patch -db_url="postgres://username:password@host/database"
```

can set `-interval=5` to manually set interval (second), default 5 sec
