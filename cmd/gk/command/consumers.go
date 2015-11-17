package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/funkygao/golib/color"
)

type Consumers struct {
	Ui         cli.Ui
	Cmd        string
	onlineOnly bool
}

func (this *Consumers) Run(args []string) (exitCode int) {
	var (
		cluster string
		zone    string
	)
	cmdFlags := flag.NewFlagSet("consumers", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", "", "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	cmdFlags.BoolVar(&this.onlineOnly, "l", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if zone == "" {
		forAllZones(func(zone string, zkzone *zk.ZkZone) {
			this.printConsumers(zone, zkzone, cluster)
		})

		return
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZonePath(zone)))
	this.printConsumers(zone, zkzone, cluster)

	return
}

// Print all controllers of all clusters within a zone.
func (this *Consumers) printConsumers(zone string, zkzone *zk.ZkZone, clusterFilter string) {
	this.Ui.Output(zone)
	zkzone.WithinClusters(func(name, path string) {
		if clusterFilter != "" && clusterFilter != name {
			return
		}

		zkcluster := zkzone.NewCluster(name)
		this.Ui.Output(strings.Repeat(" ", 4) + name)
		for group, consumers := range zkcluster.ConsumerGroups() {
			if len(consumers) > 0 {
				this.Ui.Output(fmt.Sprintf("\t%s %s", color.Green("☀︎"), group))
				for _, c := range consumers {
					this.Ui.Output(fmt.Sprintf("\t\t%s", c))
				}
			} else if !this.onlineOnly {
				this.Ui.Output(fmt.Sprintf("\t%s %s", color.Yellow("☔︎"), group))
			}
		}
	})

}

func (*Consumers) Synopsis() string {
	return "Print consumer groups from Zookeeper"
}

func (this *Consumers) Help() string {
	help := fmt.Sprintf(`
Usage: %s consumers [options]

	Print consumer groups from Zookeeper

Options:

  -z zone
  	Only print kafka controllers within this zone.

  -c cluster

  -l 
  	Only show online consumer groups.

`, this.Cmd)
	return strings.TrimSpace(help)
}
