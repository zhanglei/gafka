package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/go-helix"
	"github.com/funkygao/gocli"
)

type Node struct {
	Ui  cli.Ui
	Cmd string

	admin helix.HelixAdmin
}

func (this *Node) Run(args []string) (exitCode int) {
	var zone string
	cmdFlags := flag.NewFlagSet("node", flag.ContinueOnError)
	cmdFlags.StringVar(&zone, "z", ctx.DefaultZone(), "")
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	this.admin = getConnectedAdmin(zone)
	defer this.admin.Disconnect()

	return
}

func (*Node) Synopsis() string {
	return "Node management"
}

func (this *Node) Help() string {
	help := fmt.Sprintf(`
Usage: %s node [options]

    %s

Options:

    -z zone

   

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
