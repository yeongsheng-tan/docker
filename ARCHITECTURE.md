# Docker 1.0 architecture

## Introduction


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

Syntax: `start`

`start` looks up the list of startup processes defined in the containers configuration, executes
each of them in a concurrent `exec` job, and waits for them to exit before exiting.

If another `start` is already running in the same container, the second call will fail.


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
port allocation with iptables, etc. - are not implemented in the engine itslef, but in external plugins.

This makes the engine less useful on its own - but also much more portable and customizable.

The engine can download and install plugins dynamically without requiring a restart. It''s also possible 
to bundle a selection of plugins along with the engine and ship them together as a single "distribution".

Plugins are distributed and installed exactly like regular containers. In fact they *are* regular containers.




### Filesystem layout

When the Docker engine starts, it uses a directory on the host filesystem as its `root container`.
All operations of the engine are confined that that directory: it will not affect any other part of the
host filesystem. Multiple engines can run in parallel, as long as they each have a separate root container.


## Execution environment

### Environment variables



## Packaging and distribution




## Container format

### Configuration


### History


