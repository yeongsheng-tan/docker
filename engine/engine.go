package engine

import (
	"github.com/dotcloud/beam"
	"net"
	"path"
	"os"
	"os/signal"
	"fmt"
)

type Engine struct {
	*beam.Server
	Root string
}

func New(root string) (*Engine, error) {
	srv := beam.NewServer()
	srv.RegisterJob("exec",		JobExec)
	srv.RegisterJob("clone",	JobNotImplemented)
	srv.RegisterJob("ls",		JobNotImplemented)
	srv.RegisterJob("ps",		JobNotImplemented)
	srv.RegisterJob("name",		JobNotImplemented)
	srv.RegisterJob("import",	JobNotImplemented)
	srv.RegisterJob("start",	JobNotImplemented)
	srv.RegisterJob("info",		JobNotImplemented)
	srv.RegisterJob("serve",	JobNotImplemented)
	srv.RegisterJob("echo",		JobNotImplemented)
	srv.RegisterJob("build",	JobNotImplemented)
	srv.RegisterJob("expose",	JobNotImplemented)
	srv.RegisterJob("connect",	JobNotImplemented)
	srv.RegisterJob("prompt",	JobNotImplemented)
	return &Engine{
		Server: srv,
		Root:	root,
	}, nil
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
	return eng.Server.ServeJob(name, args, env, streams, db)
}

func (eng *Engine) ListenAndServe() error {
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
	return eng.Server.Serve(l)
}

