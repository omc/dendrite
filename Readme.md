# Dendrite

## Overview

Dendrite scrapes your existing logs (possibly including /var/log/syslog), and re-emits the unified log stream in modern, sensible, structured formats, like json and statsd, over common protocols such as tcp(+tls), udp, http(s), write(2), or RFC5424 syslog.

## Why dendrite?

### Unified, structured logs are awesome

If you have a unified log stream, it's easy to have tools that consume, forward, and analyze your logs.

### Logging is easier than instrumentation.

All applications log.  Files are easy to read.  Not all applications are instrumented, and there are many disparate instrumentation libraries (jmx, statsd, metrics, ostrich, ...)  Pulling statistics out of log files is easier.

### Dendrite is tiny.

Running the agent on your servers typically consumes less than 5MB of ram, and very very little CPU.

### Dendrite is structured.

Dendrite understands dates, numbers, counters, timings, searchable strings, fields, etc.  Logs are more than lines of text.

### Configure dendrite, not every application.

In the open-source world we live in, it's common for e.g. a rails app to be served by nginx, a rack server, a varnish cache, and rails itself.  Also you'll want slow query logs from your database and redis instance.  It's easy to paste the configs for each of these services from the dendrite-cookbooks into /etc/dendrite/conf.d, reload dendrite, and be off and running.

## Configuration

Dendrite will load a config file at /etc/dendrite/conf.yaml.  This is overridable with the -f flag.  Dendrite will then load any yaml files it can find in a conf.d directory below the config file.  These files will be merged into the main config file.  By convention, conf.d config files should usually only contain one source or destination group each.

The primary yaml file follows the format looks like this:

    global:
      data_directory: [/var/lib/dendrite]
      # more keys may be added in later dendrite versions
    sources:
      # ... (usually empty, delegated to the conf.d)
    destinations:
      # ... (usually empty, delegated to the conf.d)
      
A typical conf.d yaml file looks like:

    sources:
      # a key/name for the service
      syslog:
      
        # Astericks, etc are useful.  Syntax is documented at
        # http://golang.org/pkg/path/filepath/#Match
        glob: /var/log/system.log

        # The log lines are parsed with a RE2 regex
        # (https://code.google.com/p/re2/wiki/Syntax). Named matching groups
        # become columns in the structured output.
        #
        # This pattern parses my OS X syslog.  Syslog isn't consistent, 
        # so this may not work on your system.
        pattern: (?P<date>.*?:[0-9]+) (?P<user>\S+) (?P<prog>\w+)\[(?P<pid>\d+)\]: (?P<text>.*)
        
        # The output of the regexp can be post-processed.  This allows you
        # to specify type information, etc.
        #
        # Current field types are string, date, tokenized, int, as well as
        # gauge, timing, and metric.  The last few types are specialized 
        # integers, and will be treated differently by statsd.
        fields:
          timestamp:
            name: date
            # This will output a UNIX timestamp (seconds since epoch)
            type: date
            # see http://golang.org/pkg/time/#Parse
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
            # unneccessary.  All named match groups are implicitly turned into 
            # string fields.  However, since I used the "text" match group  
            # above, the implicit string match no longer exists.
            type: string
          pid:
            type: int
                
Or for a destination conf.d yaml file:
    
    destinations:
      # a key/name for the destination
      stats:
        # current protocols: udp, tcp, file
        protocol: udp
        address: foo.bar.com:1234
        # current encodings: json, statsd
        encoding: statsd


Look in the cookbook directory for more examples.