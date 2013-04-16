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

Now, let's run dendrite:

```
:$ ./dendrite -c ./config.yaml -l ./example.log
```

