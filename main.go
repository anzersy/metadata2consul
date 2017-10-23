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

func main() {
	app := cli.NewApp()
	app.Name = "meta2con"
	app.Version = VERSION
	app.Usage = "synchronize the metadata to consul"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug, d",
			Usage: "Used for debugging",
			EnvVar: "RANCHER_DEBUG",
		},
		cli.StringFlag{
			Name:  "metadata-address",
			Usage: "The metadata service address",
			Value: "169.254.169.250",
			EnvVar: "METADATA_ADDRESS",
		},
		cli.StringFlag{
			Name:  "consul-address",
			Usage: "The kv service address",
			Value: "consul",
			EnvVar: "CONSUL_ADDRESS",
		},
		cli.StringFlag{
			Name:  "listen",
			Usage: "Expose health check API",
			Value: "localhost:9527",
			EnvVar: "LISTEN",
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) error {

	if c.Bool("debug"){
		logrus.SetLevel(logrus.DebugLevel)
	}

	listen := c.String("listen")

	exit := make(chan error)

	mdClient := metadata.NewClient(fmt.Sprintf("http://%s/2016-07-29", c.String("metadata-address")))

	conClient, err := api.NewClient(&api.Config{Address: c.String("consul-address")})

	if err != nil {
		err = errors.Wrapf(err, "Inited consul client failed.")
		return
	}

	mdSync := &mdwatcher.MetadataToConsul{
		Mdclient:  mdClient,
		Conclient: conClient,
	}

	go func(exit chan<- error) {
		err := mdSync.Synchronize()
		exit <- errors.Wrapf(err, "Synchronized failed.")
	}(exit)

	go func(exit chan<- error) {
		err := mdSync.ListenAndServe(listen)
		exit <- errors.Wrapf(err, "Listen failed.")
	}(exit)

	err = <-exit
	logrus.Errorf("Meta2con exited with error: %v", err)
	return err

}
