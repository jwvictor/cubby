# Cubby: a personal data storage tool for command-line power users
Cubby is a _secure_ blob storage tool optimized for command-line users that can function as a personal organizer, note-taking tool, time tracker, secret sharing platform, configuration mechanism, or blog publishing system, among many other things.

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

Puts a new blob into Cubby.

Usage: `cubby put <key> <value>`

Example: `cubby put myemail "jason@cubbycli.com"`

Optional parameters include:
- `-a`: attach file to blob, e.g. `-a file.txt`
- `-2`: add blob as a child to another blob (specified by colon-delimited path), e.g. `-2 path:to:parent`
- `-g`: add a tag, e.g. `-g work`
- `-X`: add a TTL, e.g. `-X 365d`

### `push`

Appends a single line of text to a blob.

Usage: `cubby push <path> <new line>`

Usage: `cubby push notes "TODO: buy thanksgiving turkey"`

### `get`

View and update a blob. 

`get` opens the blob's body contents in the viewer configured in `options.viewer`. By default, this is set to `editor`, which opens the blob in your system's default editor (as defined by `$EDITOR`). It can also be set to `stdout` or `viewer` for read-only use cases.

Nested blobs can be retrieved by using colon-delimited paths of form `root-blob:child-blob` -- e.g. `notes:client meeting`.

Usage: `cubby get <key>`

Example: `cubby get myemail`

### `list`

List blobs. 

`list` shows all active blobs, intended to illustrate parent-child relationships.

Usage: `cubby list`

### `search`

Search blobs. 

Searches titles, tags, and unencrypted body text.

Usage: `cubby search <substring>`

Example: `cubby search email`

### `signup`

Signs up a user identity.

By default, this subcommand takes the username and password from your `cubby-client.yaml` file. Be sure to set up this config file before running signup.

Usage: `cubby signup`

