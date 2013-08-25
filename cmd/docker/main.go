package main

import (
	"fmt"
	"os"
	"flag"
	"log"
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
		log.Fatal(fmt.Errorf("Error initializing engine: %s", err))
	}
	defer eng.Cleanup()
	if err := runJob(eng, jobName, jobArgs...); err != nil {
		log.Fatal(fmt.Errorf("Error running job '%s': %s", jobName, err))
	}
}

func runJob(eng *engine.Engine, name string, args ...string) error {
	pipes := engine.NewPipeHub()
	go func() {
		if err := eng.Serve(pipes); err != nil {
			log.Fatal("Error running engine: %s", err)
		}
	}()
	client, err := beam.NewClient(pipes)
	if err != nil {
		return fmt.Errorf("Couldn't initialize beam client: %s", err)
	}
	job, err := client.NewJob(name, args...)
	if err != nil {
		return fmt.Errorf("Couldn't create beam job: %s", err)
	}
	job.Streams.WriteTo(os.Stdout, "stdout")
	job.Streams.WriteTo(os.Stderr, "stderr")
	// FIXME: handle stdin
	if err := job.Start(); err != nil {
		return fmt.Errorf("Couldn't start beam job: %s", err)
	}
	// Wait for job to complete
	if err := job.Wait(); err != nil {
		return fmt.Errorf("Waiting for job failed: %s", err)
	}
	// Wait for all inbound streams to drain
	// FIXME: watch out for deadlocks. Who's waiting for who?
	if err := job.Streams.Shutdown(); err != nil {
		return fmt.Errorf("Failed to shutdown job streams: %s", err)
	}
	return nil
}
