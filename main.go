// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

//go:generate go run generator/packages.go

// Example service program that beeps.
//
// The program demonstrates how to create Windows service and
// install / remove it on a computer. It also shows how to
// stop / start / pause / continue any service, and how to
// write to event log. It also shows how to use debug
// facilities available in debug package.
//
package main

import (
	_ "github.com/Openera/winserv/p/a"

	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/Openera/winserv/p/b"

	// lots of imports

	_ "bufio"
	_ "bytes"
	_ "compress/gzip"
	_ "context"
	_ "crypto/tls"
	_ "crypto/x509"
	_ "crypto/x509/pkix"
	_ "encoding/binary"
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	_ "flag"
	_ "io"
	_ "io/ioutil"
	_ "log"
	_ "log/syslog"
	_ "math"
	_ "math/rand"
	_ "net"
	_ "net/http"
	_ "net/textproto"
	_ "net/url"
	_ "os/exec"
	_ "os/signal"
	_ "path/filepath"
	_ "reflect"
	_ "regexp"
	_ "runtime"
	_ "sort"
	_ "strconv"
	_ "strings"
	_ "sync"
	_ "syscall"
	_ "testing"
	_ "unsafe"

	_ "github.com/Openera/winserv/p/c"

	"github.com/Openera/winserv/out"

	_ "github.com/Openera/winserv/p/d"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	_ "github.com/Openera/winserv/p/e"
)

func init() {
	out.LogString("main/init")
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	out.LogString("main/main")

	const svcName = "myservice"

	if (len(os.Args) >= 2) && (strings.ToLower(os.Args[1]) == "nonservice") {
		out.LogString("main/nonservice")
		for {
			time.Sleep(1 * time.Minute)
		}
	}

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		runService(svcName, true)
		return
	case "install":
		out.LogString("main/install")
		err = installService(svcName, "my service")
	case "remove":
		out.LogString("main/remove")
		err = removeService(svcName)
	case "start":
		out.LogString("main/start")
		err = startService(svcName)
	case "stop":
		out.LogString("main/stop")
		err = controlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		out.LogString("main/pause")
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		out.LogString("main/continue")
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		out.LogString("main/invalid")
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func installService(name, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	s, err = m.CreateService(name, exepath,
		mgr.Config{DisplayName: desc, Description: desc, StartType: mgr.StartAutomatic})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(10 * time.Second)
	slowtick := time.Tick(1 * time.Minute)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case <-tick:
			// beep()
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				out.LogString(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error

	out.LogString(fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		out.LogString(fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	out.LogString(fmt.Sprintf("%s service stopped", name))
}
