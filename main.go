package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/anzersy/meta2con/mdwatcher"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/urfave/cli"
	"os"
)

var VERSION = "v0.0.0-dev"

func init() {
	logrus.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "meta2con"
	app.Version = VERSION
	app.Usage = "synchronize the metadata to consul"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "metadata-address",
			Usage: "The metadata service address",
			Value: "rancher-metadata",
		},
		cli.StringFlag{
			Name:  "consul-address",
			Usage: "The kv service address",
			Value: "consul",
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	if os.Getenv("RANCHER_DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	exit := make(chan error)

	mdClient := metadata.NewClient(fmt.Sprintf("http://%s/2016-07-29", c.String("metadata-address")))

	conClient, err := api.NewClient(&api.Config{Address: c.String("consul-address")})

	mdSync := &mdwatcher.MetadataToConsul{
		Mdclient:  mdClient,
		Conclient: conClient,
	}

	go func(exit chan<- error) {
		err := mdSync.Synchronize()
		exit <- errors.Wrapf(err, "Synchronized failed.")
	}(exit)

	err = <-exit
	logrus.Errorf("Meta2con exited with error: %v", err)
	return err

}
