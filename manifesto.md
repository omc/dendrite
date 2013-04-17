# Dendrite Manifesto
 
We're running two search platfoms now--websolr.com and bonsai.io.  We host a bunch of people's search engines, and those search engines generate a whole bunch of logs.  And we do stuff with them.  We'll get paged if specific things show up in the garbage collection logs.  We have cron jobs that curl individual customer's search engines, and record the latencies.  We have a bunch of traditional monit-like alerts as well, on disk usage, memory, responsiveness on a port, etc.  Things like that.  But there's so much more we could do.
 
We could deliver our customers realtime tails of their logs in the browser.  We could send an email to customers if they're using the service in an inefficient manner, or connecting from an IP address in a different EC2 region.  But in order to get there, we need a good way to extract the information we need out of all of the logs and services we have running, and we need to pipe that information around our infrastructure.
 
When I was at twitter, we had a pretty amazing setup.  You could insert a one-liner anywhere in the webapp, and  and you could log a custom json entry.  You could log when a page got loaded, a button got clicked, whatever.  
 
    {"action":"click", "page":"search-results", "result_id": 900000000, "time": ...} 
 
The entry had a pretty well-defined (and large) schema, but all of the entries would get funneled into one giant store.  You can see an example of this pretty easily.  Open chrome's network inspector on twitter, and browse around the site until you see a POST to /i/jot.  Look up the request body, and it'll be a bunch of really detailed json log entries about what you were just doing.  And you could go to the giant store and write a hadoop query to analyze some subset, or trigger some event in near realtime based upon a Storm workflow.
 
Great, so I want something like that, maybe.  But since I run an infrastructure provider, I don't especially care about what's happening in browsers--I want to know what's happening on servers.  And so do a bunch of you guys.

If you're running a mid-sized Rails app on your own, you might have a technology stack that looks something like the following:

* Ruby
* Rails
* Linux
* Thin
* Nginx
* Mysql
* Redis
* Resque
* ElasticSearch
* Cron

That's a lot of things to keep running.  Wouldn't you like to know if you've gotten a bunch of entries in your Mysql slow.log?  Or if you're getting a traffic spike?  You can write little scripts to grep all your logs, but it gets tiring, and you spend a lot of time managing them.  Or you can configure/patch all of your applications to log metrics in some specific way, but that's annoying too.  Gotta make sure that the latest version of nginx didn't break things.

So what we'd like to have is a daemon sitting on your system, aggregating all of your logs, parsing them into a common, useful format.  Then you could ship that data to Statsd/Graphite, Papertrail, logstash, greylog2, Elasticsearch, Splunk, a Mongo table, that in-house hacky thing, etc.  For us, we're going to use Graphite, StatsD, and we're going to pipe it into a websockets api we wrote.  People come up with great stuff for this all the time.

> Side note: I love the idea of using @cantino's Huginn for easy-to-write
> custom agents that do peak-detection and send email, etc.  It could be
> like a roll-my-own PagerDuty with weak AI.  New stuff every week.

Ideally, this daemon should only consume a few megabytes of memory, it should have minimal dependencies, it should be efficient.  The common open source agents for this stuff are written in Java, and depend on not only the Java dependencies, but random C deps like libtokyocabinet.  Sorry, not simple enough, and I'd like to use that -Xmx256M of RAM for my applications (and not worry about GC thrashing killing my cpu utilization).

So golang seems great for this sort of thing.  It's fast, has a small memory footprint, deploying a program is just copying an executable with no dependencies, and cross-compiling golang for a bunch of architectures took a few minutes to setup.  To some degree, this also lets you avoid the language wars.  Setting up a Ruby environment as a Python guy or vice versa is a pain, and you never fully trust it.  I mean, how does easy_install or pip work?  What's the difference between Python 2.6 and 2.7?  Is RVM for servers?  Nah, just give me an executable.

So I built an agent.  Its small, written in golang, and takes as input 
a bunch of little yaml files describing how to parse log formats.  Each file is mostly a regular expression and a short, simple description of how to take each submatch group and translate it into a column in a database table.

Then, you have another file that lists the urls you'd like to send the data to.  Stuff like `stats: "udp+statsd://foo.bar.com:4000"`.  Right now, we've implemented a couple simple things that we're going to use (json, statsd, tcp, udp, files), but we'd like to see the community step up and add syslog, http, etc.  Or I'm sure we'll get to some of those sooner or later.

It's pretty easy to extend dendrite (that's what we called it) with a new protocol (like http) or a new encoding (like xml, yay!).  It's writing one function, and adding that to a lookup table.  And the parsers are pretty easy and well-documented too.  

So use dendrite and contribute!  Please!  I'd like to be setting up a server in 201X, and copy the 14 relevant yaml files from the cookbooks folder into /etc/dendrite/conf.d, and have a complete picture of all the metrics and logs across all the relevant services running on my servers.  I bet you'd like this too.