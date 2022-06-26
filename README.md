# Cubby: a personal data storage tool for command-line users

Cubby is a _secure_ blob storage tool optimized for command-line users that can function as a personal organizer,
note-taking tool, time tracker, secret sharing platform, configuration mechanism, or blog publishing system, among many
other things.

At the most basic level, Cubby allows users to read and write blobs like so:

```
cubby put "janes phone number" 2125551212 
cubby get "janes phone number" # opens in $EDITOR by default (vim, emacs, etc.)
```

These blobs are _securely_ stored on the server without exposing any sensitive data to the server at all -- all data is
AES-128 encrypted (and decrypted) on the client side. Only ciphertext is stored on the server.

Taking a few minutes to set up and learn Cubby can save the average user a huge amount of time, effort, and mental
overhead by providing an easy, straightforward way to store important bits of information. Take a look at the
Use Case Examples below for ideas on where to start.

## Installation

The repository defines two binaries in `cmd`: a client and a server. By default, the client will use the publicly
available server at `public.cubbycli.com`, but users can choose to run a private server instead.

## Creating a config file

Create a `cubby-client.yaml` file in your `$HOME` directory with contents as follows:

```
host: https://public.cubbycli.com
port: 3737
options:
  viewer: editor
  body-only: true
user:
  email: jason@email.com
  password: passwordhere 
crypto:
  symmetric-key: sometexthere 
  mode: symmetric
```

Next, run:

```
cubby signup
```

You should see a success message and you're ready to go.

## Walkthrough

Before doing the walkthrough, be sure to make your config file and run `cubby signup` as explained in the 
Installation section above.

The first time you start up Cubby, you won't have any blobs. Let's start by making a first blob -- a to-do list.
Start by running:

```bash
cubby put todo
```

This creates a new, empty blob at the root of our blobspace called `todo`. 

Next, we're going to edit this blob. Before we do that, make sure you have set an `$EDITOR` environment variable
to edit the blob body. If none is provided, we default to `vim`. If you wanted to use `nano`, for example, you could do:

```bash
export EDITOR=nano
```

Now we can edit our blob. Simply run:

```bash
cubby get todo
```

This will open up your `todo` blob in vim. Edit it to contain whatever you want, and then save and exit when
you're done. Because our configuration file specifies encryption mode `symmetric`, our updates will be 
encrypted automatically before being sent to the server (using a key derived from the password you set
in place of `passwordhere`).

Let's make a new blob that will ultimately contain a public-facing blob post. To do this, let's first
initialize a blob called `posts` at the root of our blobspace:

```bash
cubby put posts
```

We're actually not going to use `posts` to store content -- it's being used kind of like a "directory" to
organize our posts in one spot. 

Let's make a child blob of `posts` (like a "file" in the "directory") that will actually store a 
blog post we'll share with the world. In this example, we want to encrypt our post so that a reader
needs to enter a passphrase in order to decrypt and view the content. However, we want this to be a 
passphrase we're comfortable sharing -- in other words, we don't want to reuse our standard
Cubby encryption passphrase.

To accomplish this, we'll add some new things to our `cubby put`:

* We'll introduce the type (`-T`) flag. This flag allows you to set a 
  "file type" for the blob. For our blog post, we'll use the type `markdown` so Cubby knows our data
  can be rendered using a Markdown parser. So next we'll do:
* We'll introduce the encryption key (`-K`) flag. This flag allows you to use a different encryption key
  from the one configured in your `cubby-client.yaml` file. Here, we're setting the passphrase to
  `share_password`.
* We'll add in the `-2` flag to specify the _path_ of the parent under which to put this  blob.

All together, it looks like this:

```bash
cubby put -T markdown -2 posts -K=share_password helloworld
```

In order to see how this created a child blob under the parent `posts`, run:

```bash
cubby list
```

This will list out your blobs, suitably indented to illustrate parent-child relationships. You'll see your
`helloworld` blob is under `posts`, for example.

Now let's open up our post and edit it to contain some content with (we'll need to pass the 
required passphrase with `-K` when we interact with this blob):

```bash
cubby get posts:helloworld -K=share_password
```

You've probably noticed that paths are represented in Cubby with a colon (`:`). Anywhere that you're 
expected to pass a blob ID of some kind, you can pass a colon-separated path like this instead, and
it will be automatically resolved to the proper blob ID.

Let's fill out the post with some Markdown like the following so we have interesting content to
look at for the rest of the tutorial:

```markdown
# Hello, world!

Hello, world! Welcome to Cubby. We hope it saves you a ton of time and aggravation.
```

Next, we're going to "publish" this blob for the world to see. That's super easy with Cubby! Just
run:

```bash
cubby publish put posts:helloworld
```

You'll see something like:

```
No permissions provided - defaulting to public...
Published successfully, getting URL...
Web: http://localhost:3737/v1/post/jason8081/helloworld/view
API: http://localhost:3737/v1/post/jason8081/helloworld
```

If you go the URL listed under `Web`, you'll see a page prompting you to enter the encryption passphrase.
Once you enter `share_password` and click "Decrypt", the content of your post will appear as fully
rendered HTML.

To make changes to your post, simply edit your blob with:

```bash
cubby get posts:helloworld
```

Any changes you make will be automatically reflected at the post URL.

## Core Cubby concepts

### Blobs

### Paths

### Shared blobs and publications

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

`get` opens the blob's body contents in the viewer configured in `options.viewer`. By default, this is set to `editor`,
which opens the blob in your system's default editor (as defined by `$EDITOR`). It can also be set to `stdout`
or `viewer` for read-only use cases.

Nested blobs can be retrieved by using colon-delimited paths of form `root-blob:child-blob` --
e.g. `notes:client meeting`.

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

By default, this subcommand takes the username and password from your `cubby-client.yaml` file. Be sure to set up this
config file before running signup.

Usage: `cubby signup`

