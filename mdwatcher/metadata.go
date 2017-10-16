package mdwatcher

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/rancher/go-rancher-metadata/metadata"
	"sync"
)

type MetadataToConsul struct {
	Mdclient  metadata.Client
	Conclient *api.Client
	Lock      sync.Mutex
}

func (mc *MetadataToConsul) Synchronize() error {
	return mc.Mdclient.OnChangeWithError(10, mc.DoSynchronization)
}

func (mc *MetadataToConsul) DoSynchronization(str string) {
	var err error

	mc.Lock.Lock()
	defer mc.Lock.Unlock()

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

		err = mc.Conclient.Agent().ServiceDeregister(svc.Name)

		if err != nil {
			logrus.Infof("Deregister service %s failed.", svc.Name)
		} else {
			logrus.Infof("Deregister service %s successfully.", svc.Name)
		}

		err = mc.Conclient.Agent().ServiceRegister(svc)

		if err != nil {
			logrus.Infof("Register service %s failed.", svc.Name)
		} else {
			logrus.Infof("Register service %s successfully.", svc.Name)
		}
	}
}
