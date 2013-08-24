# Docker 1.0 architecture

## Introduction

This document gives an overview of the planned architecture of Docker 1.0. It is meant to help
contributors, power users, and developers of companion projects to prepare for the 1.0 release.

This architecture will be put in place gradually, over the course of several releases. It requires
major changes to the internals of the engine, and inevitably some of these changes will require
adjustments in the user experience. But the goal is to minimize the number and magnitude of breaking
changes. In particular the remote HTTP api, registry API and overall command-line usage flow should
be preserved with zero or almost zero changes.

This double goal - aggressive internal change and minimal breaking external changes - is
difficult to get right, and carries risks: the risk of stalling development, of alienating users
with too many unnecessary changes, of discouraging developers with reckless API breakage. But that
risk is necessary if we want Docker to live up to the expectations of its community.

So let''s keep the difficulty in mind, focus on the user, and do the best engineering job we can :)



## Engine API

The engine API is the canonical way to interact with the Docker engine programatically. All interactions with the engine
happen through the engine API, directly or indirectly. This includes:

	* Command-line operation of the engine
	* Remote access to the engine via the http remote api
	* Introspection from inside a container

The engine API is designed to natively support a few important things:

	* Isolation and multi-tenancy
	* Easy crud operations on structured data
	* Unix-style process execution, including sending and receiving any number of binary streams in parallel
	* Easy watching of value changes


### Connecting to the engine

The first step to using the engine API is to find an endpoint to connect to. This will depend on the context
in which your program is being executed.

If your program is running *inside* a docker container with introspection privileges, it can get an endpoint
by opening the unix socket at `/.engine.sock` in the containers filesystem.
The endpoint will be scoped to that particular container, and your program will not have access to
anything outside that scope.

If your program is running in the host, *outside* a container, it can get an endpoint scoped to any container by
opening the `.engine.sock` unix socket in the filesystem of that container.

To get full access to all containers in the engine, simply get an endpoint on the `root container` of the engine.
See [filesystem layout] for details.



### The beam protocol

Once your program is connected to the engine, it can start communicating using the Beam protocol,
documented at http://github.com/dotcloud/beam. A short description of Beam is that it implements
querying and watching structured data, running jobs with a unix-style interface,
and sending and receiving multiple binary streams.

The interesting part is that Beam does all this *on top of the Redis protocol*. That makes it extremely
easy to implement client libraries in any language, since most of the work has already been done by
redis client libraries.

The fact that Beam is based on the Redis protocol also facilitates debugging and operations:
you can always whip out the redis cli to inspect the state of the engine, or back it up to a
slave database for later replay or retrieval.


### Inspecting a container

The first function of the engine API is to expose interesting data relative to a container
for inspection by your program. The following data is available for inspection:

	* Engine information: version, uptime, 
	* Container metadata: author, version, package name, creation date, signature.
	* Container configuration: startup processes, ports to expose, directories to persist, default environment, version of docker required.
	* Container history: a complete record of all the operations which led from an empty directory to the current state of the container.
	* Services: the network addresses of remote services accessible from the container.
	* User data: a reserved space for application-specific data.
	* Children: a list of containers nested inside the current container.

Your program can query this data using standard redis commands. It can also watch for changes using the
synchronization features of Beam (see [The Beam protocol]).

Access to inspection data is read-only. All redis commands susceptible to change the data will fail.


### Running jobs

The second function of the engine API is to get the engine to do things on your program''s behalf.

The fundamental unit of work in the docker engine is a job. Everytime the engine interacts with a container - by
executing a process, copying directories, downloading archives, changing default configuration etc. - it
does so with a job.

Using the Beam protocol (which is simply a set of redis commands), your program can instruct the engine
to execute jobs. It can then communication with the jobs through binary streams, much like unix processes.

Below is the typical sequence of running a job:

	* Create new job entry with name (eg. "exec"), arguments (eg. "echo", "hello", "world") and environment (eg. DEBUG=1 HOME=/home/foo)
	* Start new job
	* Read data from output streams (stdout, stderr...) and write data to input streams (stdin...)
	* Wait for job to complete or run more jobs in parallel


### Available jobs

Below is a list of jobs available by default in the engine API. Job names are not case-sensitive.

#### EXEC: execute a process

Syntax: `exec CMD [ARG ...]`

Environment:

	* `restart`: if it exists and is different than "0", the process should be restarted when it exits or is killed.
	* `workdir`: if it exists, sets the working directory in which the process is executed.
	* `timeout`: if this is set, the engine will wait the specified number of seconds, then terminate the process.
	* `user`: if it exists, sets the uid under which the process is executed. As a convenience,
if the value is not an integer, `/etc/passwd` is looked up in the container to determine the uid. Default=root.

Additionally, all job environment variables are passed to the process as an overlay on its default
environment (see [environment variables] under [execution environment]).


#### IMPORT: set the root filesystem to the contents of an archive

Syntax: `import -|URL`

`import` unpacks the contents of a tar archive into the root filesystem, erasing all previous
content.

If its first argument is `-`, it reads the contents of the archive from stdin. Otherwise it
downloads it from the specified URL.

The archive may be compressed with the following formats: identity (uncompressed), gzip, bzip2 or xz.


#### RUN: execute all startup processes

Syntax: `run [ARG...]`

`run` executes the container''s entry point. Think of it as "double-clicking" on a container to make
it do something. See [Run entrypoint].

If another `run` is already running in the same container, the job is aborted.


#### BUILD: execute the container''s build script

Syntax `build`

`build` looks for a build script and executes it. The goal of a build script is to transform a
non-runnable container (typically a source code repository) into its runnable form (typically
a compiled binary and its run dependencies).

The build sequence is the following:

* Look for `Dockerfile` at the root of the container. If it doesn''t exist, abort the build.

* If the executable bit is set for `Dockerfile`, execute it with `exec /Dockerfile`. Otherwise
execute it as a shell script with `exec sh /Dockerfile`. This will work because of the binaries
guaranteed to be made available by the engine in every container (see [available binaries] in
[execution environment]).


The build system in 1.0 supports all versions of the Dockerfile syntax, all the way back to
version 0.4 when the `build` command was first introduced.


#### SET: change container metadata


#### WIRE: make the ports of the current container accessible to another

Syntax: `wire consumer [consumer...]`



#### EXPOSE: advertise that a container is listening on a network port

Syntax: `expose portspec`


### Navigating containers





## Docker on the host


### Cross-platform support

The Docker engine can run on all major operating systems, including Linux, Windows, OSX and Solaris, and all
major hardware platforms, including x86, x86_64 and arm.

The core functionalities of the engine are the same across all operating systems and hardware platforms.
Because of its modular architecture, many high-level features of Docker are available as plugins.
Some of these plugins, like registry support or the http api, are portable and available on all platforms.
Others, like lxc and aufs support, are system-specific by nature and are only available on certain platforms.


### Runtime dependencies

The Docker engine is shipped as a static binary and doesn''t execute any external programs. In other words,
it has no runtime dependencies other than the kernel and harware it was built for.



## Plugins

Many of the features which made Docker popular - process isolation with lxc, copy-on-write with aufs,
port allocation with iptables, etc. - are not implemented in the engine itself, but in external plugins.

This makes the engine less useful on its own - but also much more portable and customizable.

The engine can download and install plugins dynamically without requiring a restart. It''s also possible 
to bundle a selection of plugins along with the engine and ship them together as a single "distribution".

Plugins are distributed and installed exactly like regular containers. In fact they *are* regular containers.




### Filesystem layout

When the Docker engine starts, it uses a directory on the host filesystem as its `root container`.
All operations of the engine are confined that that directory: it will not affect any other part of the
host filesystem. Multiple engines can run in parallel, as long as they each have a separate root container.


## Execution environment

This section describes the environment available to processes executed inside a container
by an `exec` job (see [EXEC] in [Available jobs]).

When writing a program which will run inside a docker container, you can rely on this environment
always being available.


### Environment variables

The following environment variables are always set by default:

```
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOME=/
```

These defaults may be overriden by the container configuration (See [container configuration]
in [container format]), or in the job environment (see [running jobs] in [the engine api]).


### Filesystem alterations

An `exec` job never modifies the container''s filesystem, with 2 exceptions:

* The unix socket `/.engine.sock` is crested for introspection (see [runtime introspection] below)
* The directory `/.docker` is created and reserved for the use of the docker engine. Processes
may traverse and read it, but must not modify it or its contents.


### Available binaries

### Built-in shell


### Runtime introspection

A program may dynamically introspect various properties of its runtime environment by accessing the
engine API. See [Engine API].

Typical examples of introspection include looking up the address of remote services (see [Service discovery]),
determining the version of the docker runtime (see [engine information]), creating and manipulating
other containers (see [children]).


## Container format on disk

A docker container is a directory managed by the Docker engine. The contents of that directory are
referred to as the container''s *filesystem*. The engine uses the filesystem as a sandboxed environment
to execute *jobs* (see [Running jobs]). There are no mandatory requirements for a container''s
filesystem - even an empty directory qualifies as a container.

Optionally, special configuration can be added to a container''s filesystem to communicate requirements
to the engine (See [Run entrypoint).

A container''s filesystem represents the entirety of its content: there is no need for a separate configuration
file, manifest or other companion data. It''s all there. This means that copying a container''s filesystem
is equivalent to duplicating a container.


## Packaging and distribution



### Run entrypoint 

A container may carry on its filesystem a `Dockerfile`, which makes the container "runnable" by
defining an entry point for the `RUN` command (See [RUN]). The Dockerfile is also used
by the `BUILD` command, which is a special case of RUN (see [BUILD]).

A Dockerfile can be written using 2 different syntaxes: simplified and advanced. The two syntaxes
are interchangeable, and can even be mixed in the same Dockerfile.


#### Simplified Dockerfile syntax

The simplified syntax is a minimal DSL introduced in version 0.4 to facilitate the use of
Dockerfiles for *building* new containers from source. For this reason the simplified syntax
is mostly used by Dockerfiles shipped with source code, as a "Makefile on steroids".

Don''t worry, you can also use the simplified syntax as an entrypoint for fully built containers.
If the notion of using a Dockerfile for something other than building confuses you, please refer
to [Source repositories are containers too!] in [Concepts].

The simplified syntax is a stripped down "pseudo-shell" syntax:

* The file is scanned line by line.
* Leading whitespaces are ignored.
* Lines made of whitespaces or starting with '#' are skipped.
* The remaining lines are split in 2 parts, using the first whitespace as a separator. The left
part is used as the command name, the right part as the argument.
* Commands are executed in sequence.
* If a command fails, the execution of the Dockerfile is aborted.

Below is a list of available commands. Command names are not case-sensitive.

* `FROM <pkg>`: create a new container, install the package `pkg` in it, and use it as the target for
all subsequent commands until the next FROM (see [NEW] and [INSTALL] in [Available jobs]).

* `MAINTAINER [name] <email>`: set the maintainer field in the target container''s metadata
(see [SET] in [Available jobs]).

* `RUN <cmd> [arg...]`: execute a process in the target container and wait for it to complete.

* `EXPOSE <por>t`: modify the configuration of the target container to expose the given port at runtime.

* `ENV <key> <value>`: set the default value of the environment variable `key` to `value` in the configuration
of the target container.

* `USER <uid>|<username>`: set the default user of the startup processes in the configuration
of the target container.

* `WORKDIR <path>`: set the default working directory of the startup processes in the configuration
of the target container.

* `VOLUME <path>`: add `path` to the list of directories to persist in the configuration of the target
container.

* `ENTRYPOINT`: 

* `CMD`: 

Commands are executed in sequence, line-by-line. Multi-line commands are not supported.


#### Advanced Dockerfile syntax

### Configuration


### History



## Concepts

This section explains various concepts which guide the design of Docker, with a particular emphasis
on new concepts introduced between 0.6 and 1.0.


### Source repositories are containers too!

If you are used to the 0.5 architecture, you are used using Dockerfile for *building containers*.
What is this nonsense about also using them as a *run* entrypoint for containers?

Well, that is because a *build* is simply what happens when you *run* a directory full of source
code.

A docker container can be defined as "any directory the docker engine can run". Technically, even an
empty directory is a valid docker container! And source directories - freshly checked out of your
git repository - are containers too!

Why is that cool? Well, running a container is analogous to a double-click: it makes the selected
object "do something", whatever that is. If you double-click an app, it runs. If you double-click
an installer, it installs its contents onto the system. And if you double-clicked a source code
repository... wouldn''t it be cool if it just built itself? Well, now it can!



### Containers are nested


Containers can contain other containers, and so on recursively with no limit of depth.

This makes it much easier to reason about several use cases, including *orchestration* (a stack
of containers is represented by a parent container), *build environments* (a container with all
the build dependencies can assemble a child container with only the runtime bits), *multi-tenancy*
(3 different users get 3 different remote api endpoints in 3 different containers to manage
3 different fleets of children containers), and so on.

Nesting also makes the engine code smaller and more generic. Since a container is just a directory
which the docker engine can run, the root context in which the docker engine runs
(usually /var/lib/docker) is itself a container. "All the containers" is another way to say
"all the children containers of the root container" - and the same code can implement both.


## Open questions


* How do we accomodate future extensions of the container configuration? (freeze/restore,
auto

