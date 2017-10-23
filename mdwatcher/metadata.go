package mdwatcher

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/rancher/go-rancher-metadata/metadata"
	"net/http"
)

type MetadataToConsul struct {
	Mdclient  metadata.Client
	Conclient *api.Client
}

func (mc *MetadataToConsul) ListenAndServe(listen string) error {
	http.HandleFunc("/ping", mc.ping)
	logrus.Infof("Listening on %s", listen)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		logrus.Errorf("got error while ListenAndServe: %v", err)
	}
	return err
}

func (mc *MetadataToConsul) ping (rw http.ResponseWriter, req *http.Request) {
	logrus.Debug("Received ping request")
	rw.Write([]byte("OK"))
}

func (mc *MetadataToConsul) Synchronize() error {
	return mc.Mdclient.OnChangeWithError(10, mc.doSynchronization)
}

func (mc *MetadataToConsul) doSynchronization(version string){
	if err := mc.DoSynchronization(version);err != nil{
		logrus.Errorf("DoSynchronization exited with error: %v", err)
	}
}

func (mc *MetadataToConsul) DoSynchronization(str string) error {

	ec := make(chan error)

	cons, _ := mc.Mdclient.GetContainers()

	for _, con := range cons {

		tags := make([]string, 2, 4)

		for k, v := range con.Labels {
			tag := fmt.Sprintf("%s:%s", k, v)
			tags = append(tags, tag)
		}

		var svc *api.AgentServiceRegistration = &api.AgentServiceRegistration{
			ID:      con.Name,
			Name:    fmt.Sprintf("%s-%s", con.StackName, con.ServiceName),
			Tags:    tags,
			Address: con.PrimaryIp,
			Check: &api.AgentServiceCheck{
				Status: con.State,
			},
		}

		err := mc.Conclient.Agent().ServiceDeregister(svc.Name)

		if err != nil {
			ec<- err
		}

		err = mc.Conclient.Agent().ServiceRegister(svc)

		if err != nil {
			ec<- err
		}
	}

	err := <-ec
	return err
}
