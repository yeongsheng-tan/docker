package main

import (
	"fmt"
	"os"
	"flag"
	"log"
	"net"
	"github.com/dotcloud/docker/engine"
	"github.com/dotcloud/beam"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("Usage: mk CMD [ARGS...]\n")
		os.Exit(1)
	}
	jobName := flag.Arg(0)
	var jobArgs []string
	if flag.NArg() > 1 {
		jobArgs = flag.Args()[1:]
	}
	eng, err := engine.New(".")
	if err != nil {
		log.Fatal(err)
	}
	defer eng.Cleanup()
	if err := runJob(eng, jobName, jobArgs...); err != nil {
		log.Fatal(err)
	}
}

func runJob(eng *engine.Engine, name string, args ...string) error {
	cPipe, sPipe := net.Pipe()
	go eng.ServeConn(sPipe)
	client := &beam.Client{Transport: cPipe}
	job := client.Job(name, args)
	job.Streams.WriteTo(os.Stdout, "stdout")
	job.Streams.WriteTo(os.Stderr, "stderr")
	// FIXME: handle stdin
	if err := job.Start(); err != nil {
		return err
	}
	// Wait for job to complete
	if err := job.Wait(); err != nil {
		return err
	}
	// Wait for all inbound streams to drain
	// FIXME: watch out for deadlocks. Who's waiting for who?
	if err := job.Streams.Shutdown(); err != nil {
		return err
	}
	return nil
}
