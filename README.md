# Cubby: personal data management for command-line users

[www.cubbycli.com](https://www.cubbycli.com)

```bash
cubby put ideas 
cubby publish put ideas
cubby get ideas
```

## Main features
* Command line-based, free, open source, cloud note-taking and storage tool
* A [Neovim plugin](https://github.com/jwvictor/nvim-cubby) to access your Cubby blobs from the best editor ever
* End-to-end AES-encrypted by default 
* Full version history, with the ability to revert to any prior version
* Share personal notes, config files, and data between your computers
* Publish potentially sensitive data using end-to-end encrypted web sharing
* Collaborate with other Cubby users securely and efficiently
* Maintain privacy in the presence of spyware 
* 100MB of storage totally free forever 

## Preamble

Your average (technically proficient) user of the command-line -- for all the tools and skills they have
access to -- no doubt has numerous text files scattered around their filesystem containing important
bits of information. Things like:

1. To-do lists
2. Passwords and other user credentials
3. Notes
4. Cryptographic keys
5. Settings, like terminal settings in `.bashrc` or SSH key settings in `.ssh`

Many of these are scattered around in random text files and documents.

Worse yet, others -- like your favorite configs and keys -- need to be arduously copied among computers in a 
vain attempt to tame chaos and keep your config files in sync.

Notes and to-do lists become siloed between "work computer" and "personal computer," and one must take care
to always be at the right computer when having thoughts or questions about a particular domain.

Alas, Cubby is here to fix all these problems. It was designed specifically for the highly technical, 
CLI-loving programmer type. In Cubby, viewing your to-do list takes some 4 keystrokes with the
standard set of aliases, whereas viewing a to-do list on Notion would involve clicking around clumsily 
like a mere technical peasant.

Add up all that time and you're losing years of your life to clunky tools designed for people who don't know how to 
use `vim` and refuse to learn.

All that to say -- try Cubby now. There's a super quick and easy install script that sets up your configs,
downloads the binary, and makes an account if necessary. Five minutes to learn Cubby now may save you
years of productivity yet.

## What is Cubby?

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
use case examples below for ideas on where to start.

## Use case examples

Some example use cases for Cubby:

- [Unifying shell configs across all your computers](https://public.cubbycli.com/v1/post/jason/cubbyrc/view)
- [Privacy while using work computers](https://public.cubbycli.com/v1/post/jason/notescompanylaptop/view)
- [Writing a developer blog with Cubby](https://public.cubbycli.com/v1/post/jason/blogging-with-cubby/view)
- [A general note on the motivation for Cubby](https://public.cubbycli.com/v1/post/jason/note-from-the-author/view)

## Installation

### Quick install

For most flavors of UNIX/Linux and Mac OS X, simply run:

```bash
curl -s -S -L https://www.cubbycli.com/static/install.sh | bash 
```

This will set up a config file in `$HOME/.cubby` and download the Cubby binary to `$HOME/.cubby/bin`.

### Long install

The repository defines two binaries in `cmd`: a client and a server. By default, the client will use the publicly
available server at `public.cubbycli.com`, but users can choose to run a private server instead.

#### Creating a config file

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

Before doing the walkthrough, be sure to install Cubby via the instructions in the "Installation" section
above. (For most users, this simply involves running the install script.)

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
blog post we'll share with the world. This first one, like most blog posts, will not be encrypted;
we want anyone who wants to read it to be able to read it. (Of course, all data is _delivered_
over a secure connection, regardless of user encryption settings.)

To accomplish this, we'll add some new things to our `cubby put`:

* We'll introduce the encryption mode (`-C`) flag. This flag allows you to specify what type of user-level
  of encryption to use for this blob. Yours is defaulted to `symmetric` encryption in `cubby-client.yaml` file,
  so for a blob we want to share publicly, we need to override that setting and make it unencrypted with `-C none`.
* We'll make this blob as a child blob under the parent `posts` using paths (as discussed under "Concepts").

All together, it looks like this:

```bash
cubby put -C none posts:helloworld
```

In order to see how this created a child blob under the parent `posts`, run:

```bash
cubby list
```

This will list out your blobs, suitably indented to illustrate parent-child relationships. You'll see your
`helloworld` blob is under `posts`, for example.

Now let's open up our post and edit it to contain some content for our blog post:

```bash
cubby get posts:helloworld
```

You've probably noticed that paths are represented in Cubby with a colon (`:`). Anywhere that you're 
expected to pass a blob ID of some kind, you can pass a colon-separated path like this instead, and
it will be automatically resolved to the proper blob ID.

Let's use `cubby get ` again to fill out the post with some Markdown like the following so we have 
interesting content to look at for the rest of the tutorial:

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
Web: https://public.cubbycli.com:443/v1/post/jason8081/helloworld/view
API: https://public.cubbycli.com:443/v1/post/jason8081/helloworld
```

If you go the URL listed under `Web`, you'll see your post fully rendered as HTML. To make changes to your post, 
simply edit your blob with:

```bash
cubby get posts:helloworld
```

Any changes you make will be automatically reflected at the post URL.

But what if you want to share something secret, and you only want a specific audience to be
able to decrypt your message? Easy enough. In this last example, we'll still make our post publicly
accessible, but we'll encrypt it so the user has to enter an encryption passphrase when they
go to read it.

However, we want this to be a passphrase we're comfortable sharing -- in other words, we don't want to reuse our standard
Cubby encryption passphrase, which secures our _personal_ data.

To accomplish this, we'll introduce the encryption key (`-K`) flag. This flag allows you to use a different encryption key
from the one configured in your `cubby-client.yaml` file. Here, we're setting the passphrase to `share_password`.

```bash
cubby put -K share_password posts:secret_post
cubby get -K share_password posts:secret_post # edit the post contents
cubby publish put posts:secret_post
```

When you go to the web link returned by `cubby publish`, you'll see a page prompting you to enter the encryption passphrase.
Once you enter `share_password` and click "Decrypt", the content of your post will appear as fully
rendered HTML.

Now, when you go to the URL for that post, no encryption password will be required in order
to view the content.

## Core Cubby concepts

### Blobs

The central concept in Cubby is the blob. A Cubby "blobspace" is simply a collection of blobs, where each
blob can itself have "child" blobs, like files in a directory. A blob has a number of attributes:

1. A title, similar to a filename, which is used to uniquely identify the blob
2. Some body text, which may be encrypted or plaintext
3. A type, which specifies the type of content and informs how it is rendered on the web app
4. Tags, which can be any string, and which allow you to quickly `cubby search` for particular types of blobs
5. Attached files, which can be unencrypted or encrypted
6. Expire time, which optionally specifies a "time-to-live", i.e. a time at which the blob is
   automatically deleted
7. A complete version history
8. Some number of "child" blobs

### Paths

Paths allow you to address blobs in your blobspace. They are simply colon-separated lists
of blob titles, from root to leaf.

For example, say we have a blobspace set up so the output of `cubby list` is as follows:

```
 posts
  helloworld
  plaintext_post
```

To address the `posts` "parent" blob, we would simply use `posts` (e.g. `cubby get posts`). 
To address one of its children,  however, such as `helloworld`, we would use 
`posts:helloworld` (e.g. `cubby get posts:helloworld`).

If we add a child below `helloworld` called `hw_child`, we could address it with the
path `posts:helloworld:hw_child`, i.e. `cubby get posts:helloworld:hw_child`.

#### Using prefix paths for maximum finger speed

The colon-separated notation gives us an easy way to name our blobs and locate them
within their hierarchical blobspace. But we're still doing a lot of unnecessary typing!
That's where prefix paths come in. Instead of providing the full blob titles as the
segments in a path, you can instead use _prefixes_ wherever the use of such prefix
would not create ambiguity. (If it does create ambiguity, an arbitrary response wil
be returned.)

Say, for example, you have a blobspace that has these blobs:

```text
 places
  alabama
  austria
  burbank
```

If I'm looking up `burbank`, I could run:

```bash
cubby get p:b
```

Since there is only one root blob starting with `p` and one child of that root 
starting with `b`, the path `p:b` is unambiguous. If, however, I wanted to look
up `austria`, I'd need to instead do:

```bash
cubby get p:au
```

Because there are two blobs under `places` starting with `a`, I need to provide an
extra letter to make the path unambiguous.

### Shared blobs and publications

Any blob -- encrypted with any key or in plaintext -- can be "published" via Cubby. The details of how
the shared version of the blob work depend on the configuration of the blob itself. The main two details
to consider are the _type_ of the blob and its encryption settings.

#### Types for shared blobs

Currently, the only "special" type for shared blobs is Markdown, which is also the default type for a new blob, 
as demonstrated in the walkthrough above. Blobs with type set to `markdown` will be treated like
"blog posts" -- they will be visible both via API and on the web at a generated URL. On the web, the Markdown 
will be properly converted into HTML and syntax highlighting will be applied to code snippets.

#### Encryption settings for shared blobs

If a blob is encrypted, the users attempting to access it -- whether by web, CLI, or API -- will need to have
the proper symmetric key passphrase in order to decrypt its contents. Typically, users don't want to give out
their personal encryption key, so there are two things we can do instead:

1. Pass `-C none` on the `cubby put`, which disables encryption completely. When this blob is published,
   it will be a fully public post that anyone with the URL can view.
2. Pass `-K <key>` with a special key only for this share. When this blob is published, users will only
   be able to view it with that key. You can communicate this offline to the intended audience, and
   this way the shared data is never exposed to the server or to the public.

### Configs

Virtually all configs -- every one except the title/data fields for `cubby put` -- are controlled by both Cobra
and Viper. This means you can set your defaults in `cubby-client.yaml` and override them as needed with
command line flags. Running `cubby help` will show you a full list of both the config file key and
command line switch versions for each variable. (Note that, for config file keys, the period `.` denotes
keys that are nested in the YAML file, e.g. `crypto.mode`, where `mode` is under the block `crypto`.)

## Cubby subcommands

### `attachments`

View and download file attachments to blobs.

Users can pass attachments with the `-a` flag to `cubby put`, e.g.:

```bash
cubby put -a README.md files:readme
```

The `attachments` subcommand allows you to view the attachments with `cubby attachments <blob path>`,
which produces a list of files attached the blob. To download one or more files, which may or may
not be encrypted, use the `-F` switch, e.g.:

```bash
cubby attachments <blob path> -F <filename> 
```

Usage: `cubby attachments <blob path> [-F <filename>]`

Usage: `cubby attachments backup:file -F file.dat`

### `cat`

Output the body of a blob to STDOUT.

Effectively, `cubby cat` is equivalent to `cubby get -b=true -V=stdout`.

Usage: `cubby cat <key>`

Example: `cubby cat todo`

### `get`

View and update a blob.

`get` opens the blob's body contents in the viewer configured in `options.viewer`. By default, this is set to `editor`,
which opens the blob in your system's default editor (as defined by `$EDITOR`). It can also be set to `stdout`
or `viewer` for read-only use cases.

Nested blobs can be retrieved by using colon-delimited paths of form `root-blob:child-blob` --
e.g. `notes:client meeting`.

Usage: `cubby get <key>`

Example: `cubby get todo`

Optional parameters include:

- `-V`: override configured viewer (one of `editor`, `stdout`, or `viewer`) -- `stdout` is often used when calling Cubby from a script 
- `-b`: when using stdout, shows body only (default: true)

### `grep`

Pull down each blob in your blobspace, decrypt it, and search it for a substring. Regular
expression support is forthcoming.

Usage: `cubby grep [-i] <substring>`

Example: `cubby grep bonjour`

Optional parameters include:

- `-i`: case insensitive search 

### `help`

The most important command of all: `help`. You can use `help` (or `-h`) on both the root `cubby` command
or on any individual subcommands. For example, to get a list of all subcommands and global options,
run `cubby help`, but to get options for `cubby put` specifically, run `cubby put -h` (for some commands, `help`
can be mistakenly interpreted as a blob path).

Usage: `cubby help`

Example: `cubby help`

### `list`

List blobs.

`list` shows all active blobs, intended to illustrate parent-child relationships.

Usage: `cubby list`

### `push`

Appends a single line of text to a blob.

Usage: `cubby push <path> <new line>`

Usage: `cubby push notes "TODO: buy thanksgiving turkey"`

### `put`

Puts a new blob into Cubby.

Usage: `cubby put <key> <value>`

Example: `cubby put myemail "jason@cubbycli.com"`

Optional parameters include:

- `-a`: attach file to blob, e.g. `-a file.txt`
- `-2`: add blob as a child to another blob (specified by colon-delimited path), e.g. `-2 path:to:parent`
- `-g`: add a tag, e.g. `-g work`
- `-X`: add a TTL, e.g. `-X 365d`

### `revert`

Revert a blob to a prior version.

Blob should be specified by its ID or path. The user will be prompted to view the version
history and provide a revision number to which Cubby will revert the blob's contents.

Usage: `cubby revert <blob path>`

Usage: `cubby revert blog:hello-world`

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


## A note from the author

Cubby was written, first and foremost, to solve my personal needs. The default Cubby server now
publicly available at `public.cubbycli.com` was previously for personal use only. But as I've
used Cubby more and more, I've come to think that it's worth sharing with the world in hopes that you
all find it useful as well.

I wanted a way to interact with my personal data that was fast, efficient, and centered around
a command-line interface. To be clear, I don't mean a CLI _utility_, a companion to some larger app
or service, but rather an experience designed from the ground up for the command-line. With Cubby, all
my personal data is just a few keystrokes away; my default editor is the highly-powered `vim`, which
lets me edit my stuff -- again -- _efficiently_. With the help of macros and buffers and yanks.

Likewise, important files, such as cryptographic keys and identity files, are immediately available
at every computer I have. Simply running `cubby get <whatever>` pulls down these critical pieces
of data that I need to copy to every computer, cloud server, and container I run.

Aside from being an extremely fast an efficient way to work with personal data, Cubby is useful for
managing _personal_ notes in which you may sometimes mention matters from work. For example, you remember
something to do for a client at work, and you want to write "get back to client X" on your to-do
list. Exposing that information to the eyes of whatever service provider you're using _could_ violate
your company's security policies.

So, Cubby encrypts your data _at the source_ and only sends ciphertext up to the server. When you
want to view the data, Cubby pulls down the ciphertext and uses a decryption key you provide to
decrypt the data locally. As such, the name of "client X" was never exposed to Cubby's servers,
keeping you safe if you need to mention confidential information.

The other problem of mine that Cubby solves is sharing. Sharing covers everything from sharing a
password or cryptographic key with a single coworker to publishing a blog post to the entire world.

I've often felt that the mere inertia of blogging and the pressure to compile a suitable "blog"
to host my _oeuvres_ often discouraged me from publishing my thoughts to the world. With Cubby,
any Markdown blob you store can be swiftly published with `cubby publish put <blob name>`. A
unique URL is generated that you can share with the world, where readers can view your
(now-rendered) Markdown post in all its glory.

And the same applies for sharing secrets with other people. You can set a custom encryption
key for a blob containing a password with `-K` and share it with your coworker using
`cubby publish put -r <user email>`.  As long as your coworker has a Cubby account and
you tell them the encryption passphrase you used (if any), they will be able to pull down
the ciphertext and decrypt it.

This base set of features lets me build all sorts of interesting little tools and workflows
that make my job easier. It allows me to keep myself organized and provides me with a
"cloud filesystem" of sorts that I can tightly integrate into existing scripts and
workflows -- and one that, importantly, allows me to use the tools with which I'm most
efficient.

