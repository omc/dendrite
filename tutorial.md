# Dendrite Tutorial

The goal of this tutorial is to see dendrite in action on your own system.

## Download and unpack dendrite.

Browse to [the downloads page](https://github.com/onemorecloud/dendrite/blob/master/downloads.md) and grab the latest binary for your system.  Then unpack it and change into the resulting directory.  For OS X, this looks something like:

```
:$ wget https://s3.amazonaws.com/dendrite-binaries/darwin-amd64/0.1.0/dendrite-darwin-amd64-0.1.0.tar.gz
:$ tar -zxvf dendrite-darwin-amd64-0.1.0.tar.gz
:$ cd dendrite-darwin-amd64-0.1.0
:$ ls -lh
total 7136
-rw-r--r--  1 kyle  wheel   1.1K Apr 16 14:38 LICENSE
-rw-r--r--  1 kyle  wheel    41B Apr 16 14:38 REVISION
-rw-r--r--@ 1 kyle  wheel   5.4K Apr 16 14:38 Readme.md
-rw-r--r--  1 kyle  wheel     5B Apr 16 14:38 VERSION
drwxr-xr-x  5 kyle  wheel   170B Apr 16 15:41 cookbook
-rwxr-xr-x  1 kyle  wheel   3.5M Apr 16 14:38 dendrite
```

> Alternately, if you have a Go environment and want to run from trunk, you 
> can simply run `go get github.com/onemorecloud/dendrite/cmd/dendrite`.


## Set up a fake configuration.

Dendrite will normally look for /etc/dendrite/config.yaml, but we're going to tell it otherwise.  For now, lets create the following simple config at ./config.yaml:

```yml
global:
  offset_dir: ./offsets
destinations:
  console: "file+json:///dev/stderr"
```

Now, let's make two directories:
```
:$ mkdir -p offsets
:$ mkdir -p conf.d
```

Copy cookbook/syslog.osx.yaml into conf.d.  Dendrite will automatically pick up yaml files in the conf.d relative to the main config file.

```
:$ cp cookbook/syslog.osx.yaml conf.d
```

## Now, let's run dendrite:

```
:$ ./dendrite -f ./config.yaml -l ./example.log
```

We're logging errors and miscellaneous info to ./example.log, but you should be seeing a json representation of your syslog on the screen.  Hit Ctrl-C to leave dendrite.

If you run it again, you should initially see nothing, then as new lines are logged, you'll see them in your console.  While you ran the command before, dendrite was keeping track of what you viewed.  When you restarted it, dendrite read the old offset of where you left off, and restarted at that point.

```
:$ cat offsets/system.log.ptr
143103 
:$ rm offsets/system.log.ptr 
:$ ./dendrite -f ./config.yaml -l ./example.log
```

Remove the pointer file, and dendrite will again play through your system log.

## Next Steps

### Installation and Production Usage

Dendrite is a simple statically linked binary, so you can drop it onto your server.  Typical installed locations are:

* /usr/local/bin/dendrite -- installed binary
* /var/lib/dendrite -- offsets dir
* /var/log/dendrite/dendrite.log -- log
* /etc/dendrite/config.yaml -- main config file
* /etc/dendrite/conf.d/*.yaml -- other config files

Other than perhaps a pid file, that's all you should need.

For now, it's up to you to use chef, init/upstart scripts, monit, etc to manage the installation and running of the binary.  We'll get debian/yum packages up sooner or later (Please contribute!). 
