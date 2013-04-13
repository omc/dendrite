# Dendrite

## Overview

Dendrite scrapes your existing logs, and re-emits the unified log stream in modern, sensible, structured formats, like JSON and StatsD, over common protocols such as TCP(+TLS), UDP, HTTP(S), write(2), or RFC5424 Syslog.

## Why Dendrite?

### Unified, structured logs and metrics are awesome

If you have a unified log stream, it's easy to build and use tools that consume, forward, and analyze your logs.

### Logging is easier than instrumentation.

All applications generate logs. Not all applications are instrumented for metrics. On top of which, there are many disparate instrumentation libraries, such as JMX, StatsD, Metrics, Ostrich, and others.

Files are easy to read. Extracting metrics and statistics out of log files can be much easier than instrumenting an entire application to emit metrics.

### Configure dendrite, not every application.

In today's open-source environment, it's common for, e.g., a Ruby on Rails app to be served by HAProxy, Nginx, a Varnish server, a Rack server, and Rails itself. And then you'll want slow query logs from your database and your Redis server. The list goes on.

It's easy to create and share useful configuration cookbooks for each of these services, drop them into your `/etc/dendrite/conf.d`, reload Dendrite, and be off and running with real-time metrics.

### Dendrite is structured.

Logs are more than lines of text. Dendrite understands dates, numbers, counters, timings, searchable strings, fields, etc.

### Dendrite is tiny.

Running the agent on your servers typically consumes less than 5MB of RAM, and very little CPU.

## Configuration

Dendrite will load a config file at `/etc/dendrite/conf.yaml`. This is overridable with the `-f` flag. Dendrite will then load any YAML files it can find in a `conf.d` directory below the main `conf.yml` config file. The configurations in these files will be merged into the main config file. By convention, `conf.d` config files should only contain one source or destination group each.

The primary YAML file follows a format looks like this:

```yml
global:
  data_directory: [/var/lib/dendrite]
  # more keys may be added in later dendrite versions
sources:
  # ... (usually empty, delegated to the conf.d)
destinations:
  # ... (usually empty, delegated to the conf.d)
```

A typical conf.d yaml file looks like:

```yml
sources:
  # a key/name for the service
  syslog:
  
    # Astericks, etc are useful. Syntax is documented at
    # http://golang.org/pkg/path/filepath/#Match
    glob: /var/log/system.log

    # The log lines are parsed with a RE2 regex
    # (https://code.google.com/p/re2/wiki/Syntax). Named matching groups
    # become columns in the structured output.
    #
    # This pattern parses my OS X syslog. Syslog isn't consistent, 
    # so this may not work on your system.
    pattern: (?P<date>.*?:[0-9]+) (?P<user>\S+) (?P<prog>\w+)\[(?P<pid>\d+)\]: (?P<text>.*)
    
    # The output of the regexp can be post-processed. This allows you
    # to specify type information, etc.
    #
    # Current field types are string, date, tokenized, int, timestamp,
    # as well as gauge, timing, and metric. The last few types are 
    # specialized integers, and will be treated differently by statsd.
    fields:
      # tstamp is the field name in the output.
      tstamp:
        # date is the name of the regex match group.
        name: date
        type: timestamp
        format: Jan _2 15:04:05
      line: 
        # you can match numbered subgroups, in addition to named ones.
        group: 0
      tokens: 
        name: text
        # this will create an array of the matched tokens.
        type: tokenized
        pattern: \S+\b
      text: 
        # If there wasn't the tokens field above, this would be 
        # unneccessary. All named match groups are implicitly turned into 
        # string fields. However, since I used the "text" match group  
        # above, the implicit string match no longer exists.
        type: string
      pid:
        type: int
```

Or for a destination `conf.d` YAML file:

```yml
destinations:
  # a key/name for the destination
  stats:
    # current protocols: udp, tcp, file
    protocol: udp
    address: foo.bar.com:1234
    # current encodings: json, statsd
    encoding: statsd
```

Look in the cookbook directory for more examples.

## Contributing

Join us in our [Dendrite HipChat](https://www.hipchat.com/gKr8c8S4o) room to give feedback and chat about how you might be interested in using and/or contributing.

## Getting Dendrite

For binaries of versioned releases, please look at our [downloads page](./downloads.md)