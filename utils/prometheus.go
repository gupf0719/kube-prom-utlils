package kube

import (
	"coordinator/log"
	"coordinator/models"
	"coordinator/utils"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
)

// prometheus 扩容
func CreateProm() error {
	number := models.GetNumber() // 获取prom集群内部编号
	num := utils.IntToStr(number + 1)
	log.Info("create prom number ", number)

	// 1. cm
	_, cm, err := PromCmYaml(num)
	if err != nil {
		return err
	}
	err = DynamicK8s("create", cm)
	if err != nil {
		return err
	}
	// 2. svc
	_, svc, err := PromSvcYaml(num)
	if err != nil {
		_ = DynamicK8s("delete", cm)
		return err
	}
	err = DynamicK8s("create", svc)
	if err != nil {
		_ = DynamicK8s("delete", cm)
		return err
	}
	// 3. deploy
	_, deploy, err := PromDeployYaml(num)
	if err != nil {
		_ = DynamicK8s("delete", cm)
		_ = DynamicK8s("delete", svc)
		return err
	}
	err = DynamicK8s("create", deploy)
	if err != nil {
		_ = DynamicK8s("delete", cm)
		_ = DynamicK8s("delete", svc)
		return err
	}

	var prometheus models.Prometheus
	prometheus.Id = bson.NewObjectId()
	prometheus.Url = "http://prometheus" + num + ":9090"
	prometheus.State = "using"
	prometheus.Number = number + 1
	prometheus.ServerName = "prometheus" + num

	_ = prometheus.AddProm()
	return err
}

// 移除 prometheus
func DelProm(number string) error {
	cmFile := "/storage/install/prometheus-config-" + number + ".yml"
	cmData, err := ioutil.ReadFile(cmFile)
	if err != nil {
		return err
	}
	err = DynamicK8s("delete", cmData)
	if err != nil {
		return err
	}
	err = os.Remove(cmFile)

	svcFile := "/storage/install/prometheus-svc-" + number + ".yml"
	svcData, err := ioutil.ReadFile(svcFile)
	if err != nil {
		return err
	}
	err = DynamicK8s("delete", svcData)
	if err != nil {
		return err
	}
	err = os.Remove(svcFile)

	deployFile := "/storage/install/prometheus-deploy-" + number + ".yml"
	deployData, err := ioutil.ReadFile(deployFile)
	if err != nil {
		return err
	}
	err = DynamicK8s("delete", deployData)
	if err != nil {
		return err
	}
	err = os.Remove(deployFile)

	_ = models.DelPromBySM("prometheus" + number)

	return err
}
