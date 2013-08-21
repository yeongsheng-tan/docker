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



### The beam protocol


### Inspecting a container


### Running jobs


### Available jobs


### Navigating containers





## Docker on the host


### Cross-platform support



### Runtime dependencies


### Plugins



### Filesystem layout


## Execution environment

### Environment variables



## Packaging and distribution




## Container format

### Configuration


### History


