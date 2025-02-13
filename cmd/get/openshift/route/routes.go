/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package route

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type RoutesItems struct {
	ApiVersion string           `json:"apiVersion"`
	Items      []*routev1.Route `json:"items"`
}

func GetRoutes(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "host/port", "path", "services", "port", "termination", "wildcard"}
	var namespaces []string
	if allNamespacesFlag == true {
		namespace = "all"
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, namespace)
	}

	var data [][]string
	var _RoutesList = RoutesItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items RoutesItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/route.openshift.io/routes.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshal file " + CurrentNamespacePath + "/route.openshift.io/routes.yaml")
			os.Exit(1)
		}

		for _, Route := range _Items.Items {
			if resourceName != "" && resourceName != Route.Name {
				continue
			}

			if outputFlag == "yaml" {
				_RoutesList.Items = append(_RoutesList.Items, Route)
				continue
			}

			if outputFlag == "json" {
				_RoutesList.Items = append(_RoutesList.Items, Route)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_RoutesList.Items = append(_RoutesList.Items, Route)
				continue
			}

			//name
			RouteName := Route.Name
			if allResources {
				RouteName = "route.route.openshift.io/" + RouteName
			}

			//host/port
			hostPort := Route.Spec.Host

			//path
			path := Route.Spec.Path

			//services
			services := Route.Spec.To.Name

			//ports
			port := ""
			if Route.Spec.Port == nil {
				port = "<all>"
			} else {
				port = Route.Spec.Port.TargetPort.String()
			}
			termination := ""
			if Route.Spec.TLS != nil {
				termination = string(Route.Spec.TLS.Termination)
				if Route.Spec.TLS.InsecureEdgeTerminationPolicy != "" {
					termination += "/" + string(Route.Spec.TLS.InsecureEdgeTerminationPolicy)
				}
			}

			//wildcard
			wildcard := string(Route.Spec.WildcardPolicy)
			//labels
			labels := helpers.ExtractLabels(Route.GetLabels())
			_list := []string{Route.Namespace, RouteName, hostPort, path, services, port, termination, wildcard}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 8, _list)

			if resourceName != "" && resourceName == RouteName {
				break
			}
		}
		if namespace != "" && _namespace == namespace {
			break
		}
	}

	if (outputFlag == "" || outputFlag == "wide") && len(data) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var headers []string
	if outputFlag == "" {
		if allNamespacesFlag == true {
			headers = _headers[0:8]
		} else {
			headers = _headers[1:8]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			headers = _headers
		} else {
			headers = _headers[1:]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}

	if len(_RoutesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _RoutesList.Items[0]
	} else {
		resource = _RoutesList
	}

	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(resource)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(resource, jsonPathTemplate)
	}
	return false
}

var Route = &cobra.Command{
	Use:     "route",
	Aliases: []string{"routes", "route.route.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetRoutes(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}
