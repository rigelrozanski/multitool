# Multitool

## Installation From Source

1. make sure you [have Go installed][1] and [put $GOPATH/bin in your $PATH][2]
2. clone this repository while navigated to $GOPATH/src/github.com/rigelrozanski/ 
   with `git clone https://github.com/rigelrozanski/multitool.git`
3. go to the repo directory with `cd $GOPATH/src/github.com/rigelrozanski/multitool`
4. run `make install`

[1]: https://golang.org/doc/install
[2]: https://github.com/tendermint/tendermint/wiki/Setting-GOPATH 

### Slack

Often when attempting to copy and paste text from a slack conversation there is 
excess whitespace and timestamps riddled throughout. This tool will remove the whitespace
and pesky timestamps automatically within the clipboard. Usage

```
<copy text to the clipboard from slack>
mt slck [names]
<paste text into desired output>
```

Note that the `[names]` arguments are optional but add extra whitespace between the 
conversation paragraphs from different participants. The name is the slack handle of
the participant. 

### Got

Adapted from https://github.com/ebuchman/got

Got is a tool for doing things with go and git that the go tool doesn't do for you by default. It's particularly useful for working with 
forks of open source projects, where import paths become an issue.

##### Usage
Replace strings in an entire directory tree
```
got replace [-d <depth> -p <dirpath>] <oldString> <newString>
```
Switch import paths to upstream repo, pull, switch back
```
got pull <remote> <branch>
```
This does the same thing as running replace, followed by a commit, followed by a git pull, followed by a replace which undoes the first replace.

Check out the same branch across many repositories
```
got checkout develop
```
This will `cd` into every directory in the current one and run `git checkout develop`. If you want to specify exceptions, add arguments of the form `<repo>:<branch>`, ie
```
got checkout develop myrepo:newfeature repo2:master
```
will `cd` into every repo and run `git checkout develop` except in `myrepo` it will do `git checkout newfeature` and in `repo2` will do `git checkout master`. To only `cd` into some directories, just list them, ie. 
```
got checkout develop myrepo anotherepo repo2:master
```
will checkout `myrepo` and `anotherepo` onto develop and `repo2` onto master while leaving all others alone.

See which branch every repo in a directory is on 
```
got branch
```
Toggle the import path in a go directory between the Godep/ vendored packages and the $GOPATH based location

```
got dep --local github.com/myorg/myrepo
```
and 
```
got dep --vendor github.com/myorg/myrepo
```

