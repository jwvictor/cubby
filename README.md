# Cubby
Cubby is a _secure_ blob storage tool optimized for command-line users that can function as a personal organizer, configuration mechanism, or publishing system, among many other things.

In Cubby, users can read or write blobs like so:

```
cubby put "janes phone number" 2125551212 
cubby get "janes phone number" # opens in $EDITOR by default (vim, emacs, etc.)
```

All blobs are encrypted on the client side and only ciphertext is stored on the server.

## Installation
The repository defines two binaries in `cmd`: a client and a server. By default, the client will use the publicly available server at `public.cubbycli.com`, but users can choose to run a private server instead.

## Creating a config file
Create a `cubby-client.yaml` file in your `$HOME` directory with contents as follows:

```
host: https://public.cubbycli.com
port: 3737
options:
  viewer: editor
  body-only: true
user:
  email: jason
  password: passwordhere 
crypto:
  symmetric-key: sometexthere 
  mode: symmetric
```

Next, run:

```
cubby signup
```

You should see a success message and be ready to go.

## Cubby subcommands

### `put`

Usage: `cubby put <key> <value>`
Example: `cubby put myemail "jason@cubbycli.com"`

Puts a new blob into Cubby.

Optional parameters include:
- `-a`: attach file to blob, e.g. `-a file.txt`
- `-2`: add blob as a child to another blob (specified by colon-delimited path), e.g. `-2 path:to:parent`
- `-g`: add a tag, e.g. `-g work`
- `-X`: add a TTL, e.g. `-X 365d`

