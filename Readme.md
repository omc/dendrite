# Dendrite

## Overview

Dendrite is a log forwarding daemon.  Dendrite can continuously tail both your application and system logs, and emit a variety of formats, including json and statsd.  

## Why dendrite?

### Unified logs are awesome

If you have a unified log stream, it's easy to have tools that consume, forward, and analyze your logs.

### Logging is easier than instrumentation.

All applications log.  Files are easy to read.  Not all applications are instrumented, and there are many disparate instrumentation libraries (jmx, statsd, metrics, ostrich, ...)  Pulling statistics out of log files is easier.

### Dendrite is tiny.

Running the agent on your servers typically consumes less than 5MB of ram, and very very little CPU.

### Dendrite is structured.

Dendrite understands dates, numbers, counters, timings, searchable strings, fields, etc.  Logs are more than lines of text.

### Configure dendrite, not every application.

In the open-source world we live in, it's common for e.g. a rails app to be served by nginx, a rack server, a varnish cache, and rails itself.  Also you'll want slow query logs from your database and redis instance.  It's easy to paste the configs from the dendrite-cookbooks into /etc/dendrite/conf.d, reload dendrite, and be off and running.

## Why not syslog?

Syslog has been the most common unified logging format.  Many applications, however, don't speak to syslog.  If you're using an open-source application such as wordpress, you'll have to learn, download, and install extra plugins to send logs to syslog.  If your wordpress install is behind nginx, then you'll have to patch nginx to speak to syslog.  To use syslog as a unified application stream, you have to configure/patch all of your applications in a unique way.

Additionally, the syslog spec is ill-defined.  There are two RFCs.  The older, more popular one (RFC 3164), doesn't define the encoding (e.g. UTF-8, ascii, etc) of messages.  Syslog facilities won't complain about literal newlines inside your messages, despite syslog being a line-delimited format. Timestamps don't include essential information, such as the year or time zone.  Messages are limited to 1kb.  There is no standard for structured data.  The newer RFC fixes some of these problems.  It adds a very limited implementation of structured data, but also regresses on several fronts (having a well-defined grammar, single timestamp format, and more).

Dendrite scrapes your existing logs (possibly including /var/log/syslog), and re-emits the unified log stream in more modern, sensible, and structured formats.

To log a new application (e.g. apache2) into dendrite, you grab the apache2.yaml config file from github.com/dendrite/cookbooks, put it in /etc/dendrite/conf.d, and reload dendrite.


## Configuration

