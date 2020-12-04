package kube

import (
	"bytes"
	"coordinator/conf"
	"coordinator/utils"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

var availableFunctions = template.FuncMap{
	//"GenerateTest": GenerateTestFunc,
}

// 检查模版文件
func CheckTemplate() {

}

func Read(filePath string, data interface{}, funcs template.FuncMap) ([]byte, error) {
	tmpl, err := template.New(filepath.Base(filePath)).
		Funcs(funcs).
		ParseFiles(filePath)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deployment prometheus
func PromDeployYaml(number string) (string, []byte, error) {
	var promYamlData = map[string]string{
		"App":           "prometheus" + number,
		"Namespace":     conf.YamlConf.Namespace,
		"Name":          "prometheus" + number,
		"PromImage":     conf.YamlConf.PrometheusImage,
		"PromStorage":   "--storage.tsdb.path=/storage/prometheus" + number + "/tsdb",
		"ThanosImage":   conf.YamlConf.ThanosImage,
		"ThanosStorage": "--tsdb.path=/storage/prometheus" + number + "/tsdb",
		"InitStorage":   "mkdir -pv /storage/prometheus" + number + ";chmod -R 777 /storage/prometheus" + number,
		"BbStorage":     conf.YamlConf.BusyboxImage,
		"Config":        "prometheus" + number + "-config",
	}

	filePath := "/coordinator/conf/prometheus-deploy-template.yaml"
	data, err := Read(filePath, promYamlData, availableFunctions)
	if err != nil {
		return "", nil, err
	}

	deployFilename := "/storage/install/prometheus-deploy-" + number + ".yml"
	err = ioutil.WriteFile(deployFilename, data, 0777)
	return deployFilename, data, err
}

// svc prometheus
func PromSvcYaml(number string) (string, []byte, error) {
	var promSvcData = map[string]interface{}{
		"App":       "prometheus" + number,
		"Namespace": conf.YamlConf.Namespace,
		"Name":      "prometheus" + number,
		"NodePort":  30600 + utils.StrTOInt(number),
	}
	filePath := "/coordinator/conf/prometheus-svc-template.yaml"
	data, err := Read(filePath, promSvcData, availableFunctions)
	if err != nil {
		return "", nil, err
	}

	svcFilename := "/storage/install/prometheus-svc-" + number + ".yml"
	err = ioutil.WriteFile(svcFilename, data, 0777)
	return svcFilename, data, err
}

// config prometheus
func configToString(number string) (string, error) {
	var promConfigData = map[string]interface{}{
		"App":             "prometheus" + number,
		"Area":            conf.YamlConf.Area,
		"BareMetalServer": "prometheus" + number + "_bareMetalCluster",
		"K8sServer":       "prometheus" + number + "_k8sCluster",
	}

	filePath := "/coordinator/conf/prometheus-config-template.yaml"
	data, err := Read(filePath, promConfigData, availableFunctions)
	return string(data), err
}

// configmap prometheus
func PromCmYaml(number string) (string, []byte, error) {
	s, _ := configToString(number)
	var promCmData = map[string]interface{}{
		"ConfigName": "prometheus" + number + "-config",
		"Namespace":  conf.YamlConf.Namespace,
		"Config":     s,
	}
	filePath := "/coordinator/conf/prometheus-cm-template.yaml"
	data, err := Read(filePath, promCmData, availableFunctions)
	if err != nil {
		return "", nil, err
	}

	cmFilename := "/storage/install/prometheus-config-" + number + ".yml"
	err = ioutil.WriteFile(cmFilename, data, 0777)
	return cmFilename, data, err
}

// 创建三个部署文件 deploy svc cm
func PreYamlFile(number string) (string, []byte, string, []byte, string, []byte, error) {
	deployFilename, deploy, err := PromDeployYaml(number)
	svcFilename, svc, err := PromSvcYaml(number)
	cmFilename, cm, err := PromCmYaml(number)
	return deployFilename, deploy, svcFilename, svc, cmFilename, cm, err
}
