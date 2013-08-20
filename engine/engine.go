package engine

import (
	"github.com/dotcloud/beam"
	"path"
	"os"
	"os/signal"
	"fmt"
	"net"
)

type Engine struct {
	*beam.Worker
	Root string
	redisTransport beam.Connector
}

func New(root string) (*Engine, error) {
	// FIXME: setup in-memory redis database,
	// and connect the worker to it
	// For now, we connect to an actual redis database,
	// and use that as a helper
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil, fmt.Errorf("Can't connect to helper redis DB. REDIS_ADDR env variable is not set.")
	}
	eng := &Engine{
		Root: root,
		redisTransport: &beam.NetTransport{"tcp", addr},
	}
	// Setup the worker and attach it to the redis db.
	w := beam.NewWorker(eng.redisTransport, "/jobs")
	w.RegisterJob("exec",	JobExec)
	w.RegisterJob("clone",	JobNotImplemented)
	w.RegisterJob("ls",	JobNotImplemented)
	w.RegisterJob("ps",	JobNotImplemented)
	w.RegisterJob("name",	JobNotImplemented)
	w.RegisterJob("import",	JobNotImplemented)
	w.RegisterJob("start",	JobNotImplemented)
	w.RegisterJob("info",	JobNotImplemented)
	w.RegisterJob("serve",	JobNotImplemented)
	w.RegisterJob("echo",	JobNotImplemented)
	w.RegisterJob("build",	JobNotImplemented)
	w.RegisterJob("expose",	JobNotImplemented)
	w.RegisterJob("connect",JobNotImplemented)
	w.RegisterJob("prompt",	JobNotImplemented)
	eng.Worker = w
	return eng, nil
}

func (eng *Engine) Cleanup() {
	Debugf("Cleaning up engine")
	os.Remove(eng.sockPath())
}

// sockPath returns the path of the unix socket used by the engine to listen for new connections
func (eng *Engine) sockPath() string {
	return path.Join(eng.Root, ".engine.sock")
}


// ServeJob overrides the default Beam job handler to add an append-only journal.
func (eng *Engine) ServeJob(name string, args []string, env map[string]string, streams *beam.Streamer, db beam.DB) (err error) {
	// Start action in journal
	// journalPath := path.Join(eng.Root, "/.docker/history")
	// id, err := godj.Start(journalPath, name, args...)
	// if err != nil {
	//	return err
	// }
	// End action in journal
	// defer godj.End(journalPath, err)
	return eng.Worker.ServeJob(name, args, env, streams, db)
}

func (eng *Engine) ListenAndServe() error {
	// FIXME: this should start the redis database, then run the worker main loop
	// against it.
	// For now, we setup a proxy to the helper redis db.
	sockPath := eng.sockPath()
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		if c, dialErr := net.Dial("unix", sockPath); dialErr != nil {
			Debugf("Cleaning up leftover unix socket\n")
			os.Remove(sockPath)
			l, err = net.Listen("unix", sockPath)
			if err != nil {
				return err
			}
		} else {
			c.Close()
			return err
		}
	}
	Debugf("Setting up signals")
	signals := make(chan os.Signal, 128)
	signal.Notify(signals, os.Interrupt, os.Kill)
	go func() {
		for sig := range signals {
			fmt.Printf("Caught %s. Closing socket\n", sig)
			l.Close()
		}
	}()
	// FIXME: do we need to remove the socket?
	return eng.Serve(l)
}

func (eng *Engine) Serve(l net.Listener) error {
	return Proxy(l, eng.redisTransport)
}
