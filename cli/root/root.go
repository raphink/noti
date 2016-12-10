package root

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/variadico/noti/cli"
	"github.com/variadico/noti/config"
	"github.com/variadico/noti/runstat"
	"github.com/variadico/noti/triggers"
	"github.com/variadico/vbs"
)

type Command struct {
	flag *cli.Flags
	v    vbs.Printer

	Cmds map[string]cli.Cmd
}

func (c *Command) Args() []string {
	return c.flag.Args()
}

func (c *Command) Parse(args []string) error {
	if err := c.flag.Parse(args); err != nil {
		return err
	}

	c.v.Verbose = c.flag.Verbose
	return nil
}

func (c *Command) Run() error {
	c.v.Println("Running noti command")

	if runtime.GOOS == "darwin" {
		// Prevents noti from running again when user clicks notification.
		if len(os.Args) == 1 {
			p, err := exec.LookPath("noti")
			if err == nil && os.Args[0] == p {
				return nil
			}
		}

		c.v.Println("Locking OS thread")
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	if c.flag.Help {
		fmt.Println(helpText)
		return nil
	}

	return triggers.Run([]string(c.flag.Triggers), c.flag.Args(), c.Notify)
}

func (c *Command) Notify(stats runstat.Result) error {
	c.v.Println("Notifying")

	conf, err := config.File()
	if err != nil {
		c.v.Println(err)
	} else {
		c.v.Println("Found config file")
	}

	// Read default set of notification types.
	if len(conf.DefaultNotifications) == 0 {
		conf.DefaultNotifications = append(conf.DefaultNotifications, "desktop")
	}

	for _, sub := range conf.DefaultNotifications {
		subCmd, found := c.Cmds[sub]
		if !found {
			log.Println("Unknown subcommand:", sub)
			continue
		}

		ncmd, is := subCmd.(cli.NotifyCmd)
		if !is {
			continue
		}

		if err := ncmd.Notify(stats); err != nil {
			log.Printf("Failed to run %s: %s", sub, err)
		}
	}

	return nil
}

func NewCommand() cli.NotifyCmd {
	cmd := &Command{
		flag: cli.NewFlags("noti"),
		v:    vbs.New(),
	}

	return cmd
}
