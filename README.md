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

### Book Binding Commands

The first step once you have your pdf is to separate out the pdf into a folder
of all the images (one image for each page in the pdf book). To accomplish this
I use a command line program named "magick" which can be downloaded here:
(https://imagemagick.org/script/download.php) 

```
magick convert -density 300 input.pdf -quality 100 -depth 8 -sharpen 0x1.0
~/Desktop/temp/output-%04d.png
```

For instance if you were navigated to your Desktop within terminal (cd
~/Desktop) The above command would convert a file on your desktop that was
named "input.pdf" into a folder of images within a new folder "temp" on your
desktop. NOTE this step can take a LONG time to run, like 45min+ depending on
the size of your book, I think that if you replace "input.pdf" with
"input.pdf[1-3]" in the above command it will do this task for only 3 pages
(taking much less time) which will give you a sense if the command is working
or not.

The next step is to reorder all those images for proper printing and further
add them all to a single pdf. This is accomplished with my little be piece of
software. The software will need to get compile from source, here is the link
to the code and the installation instructions

https://github.com/rigelrozanski/multitool#installation-from-source

Once installed the command you could use from your Desktop within terminal is: 

```
mt pdf alt-book "temp/"
```

(optional step) if you want to get fancy and change the margins around a bit to
maximize the text size on the page you can play with the margins with the
following flags (change those 0.3 numbers around, they're the margin in inches)

``` 
mt pdf alt-book "temp/" --xmar 0.3 --ymar 0.3 
```

This would compile all the images into your desired file... once you have that
file, it'll be dead easy to print/cut/bind.

