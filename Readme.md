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

Config files are yaml.  You can look in exampleare in the simpli