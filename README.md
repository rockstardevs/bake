# bake

build and release tool for golang applications


bake is a simple, minimal build, release and versioning tool that combines the go tool, git and google drive api to make it easier to release and maintain golang applications

### What does it do?
Once installed and configured, bake will

  - maintain the current version of your app in a version file (default ./VERSION)
  - either will create a version file, if it doesn't exist, with the intial version 0.1.0
  - or increment version to the next build/minor/major version
  - build your binary (wraps aroung go build) and pass in the new version
  - built binaries are placed by default at .dist/$binary-$version-$platform-$arch
  - commit the new version file to git
  - add a git tag for the new version e.g (v0.1.1)
  - if there are no additional uncommitted changes, git push to remote
  - if configured, upload the build binary to google drive.

Apart from build, all other steps can be enabled/disabled by flags.

## Installation

Either download the latest pre compiled binary for you platform, from the links above.

**OR**

Build it yourself

```
go get github.com/singhsaysdotcom/bake
```

(you need golang installed and GOPATH configured for this to work)

## Configuration

There is no *required* configuration to get started. The is *optional* configuration to use some features.

### Uploading to Google Drive
bake uses the gdrive app from github.com/prasmussen/gdrive for Google Drive integration. You need to generate
an oauth token for the google drive api as a one time step. Currently, you need to install the gdrive app to do
that

```shell
go install github.com/prasmussen/gdrive
$GOPATH/bin/gdrive
```

Follow the interactive prompt to generate and save the oauth token.

## Usage


You need an existing existing git repo containing your source code, to start using bake

### Start managing an unversioned app with bake


```shell
bake next
```

### Build with bake

```shell
bake next
```

to build at the next build number

```shell
bake minor
```

to build the next minor version

```shell
bake major
```

to build the next major version

```shell
bake rebuild
```

to build at the current version

```shell
bake reupload
```

to build at the current version and upload the built binary to google drive.

## TODO/Things that could be better

 - add unit tests
 - add a dry run mode and make that the default
 - consider switching to a native git library instead of forking shell commands
   - not libgit2, because avoid shared lib dependencies
 - add a go get step before builds
 - add a subcommand to configure google drive api oauth token

 - make the upload task modular
 - add other upload providers (e.g. dropbox, http upload, sftp upload etc)


Contributions/Pull Requests are welcome.
