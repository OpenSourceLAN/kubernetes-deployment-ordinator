package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	sourceFileName := "./tron-template.yaml"

	sourceDeployment := parseYamlFile(sourceFileName)
	expandedDeployments := expandDeployments(*sourceDeployment)

	// for _, value := range expandedDeployments {
	// 	fmt.Printf("%s\n%++v\n%++v\n%++v\n\n",
	// 		value.ObjectMeta.Name,
	// 		value.Spec.Template.ObjectMeta.Labels,
	// 		value.Spec.Selector.MatchLabels,
	// 		value.ObjectMeta.Labels)
	// }

	kubeClient := getKubeClient()
	applyDeployments(expandedDeployments, kubeClient)
}

// Read a yaml file and return the Deployment object it contains
func parseYamlFile(path string) *appsv1.Deployment {
	yamlContents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err.Error())
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlContents), nil, nil)
	if err != nil {
		panic(err.Error())
	}

	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		panic("Could not read yaml file, probably not a deployment?")
	}
	return deployment
}

// Given a Deployment with Replicas, turn it in to n Deployments with 1 replica each
func expandDeployments(deployment appsv1.Deployment) []appsv1.Deployment {
	numberOfReplicas := int32(*deployment.Spec.Replicas)
	*deployment.Spec.Replicas = int32(1)

	generatedDeployments := make([]appsv1.Deployment, int32(numberOfReplicas))

	for i := 1; i <= int(numberOfReplicas); i++ {
		generatedDeployments[i-1] = generateDeploymentForIndex(fmt.Sprintf("%d", i), deployment)
	}

	return generatedDeployments
}

// Update a template deployment to be a numbered deployment
func generateDeploymentForIndex(index string, source appsv1.Deployment) appsv1.Deployment {

	// 3a: deep clone original deployment
	// Some of these have pointers so need to be cloned - careful, lots of minefields here

	//     3b: add additional labels for the ID (eg, label -> replica -> 3)
	source.ObjectMeta.Labels = cloneMap(source.ObjectMeta.Labels)
	source.ObjectMeta.Labels["replica"] = index

	newSelector := cloneLabelSelector(*source.Spec.Selector)
	source.Spec.Selector = &newSelector
	source.Spec.Selector.MatchLabels["replica"] = index

	source.Spec.Template.ObjectMeta.Labels = cloneMap(source.Spec.Template.ObjectMeta.Labels)
	source.Spec.Template.ObjectMeta.Labels["replica"] = index

	//     3c: Replace `###` in each env var value with the ID of the server
	updateEnvironmentValuesForIndex(index, &source)

	//     3d: append n to the Deployment metadata name
	source.ObjectMeta.Name = source.ObjectMeta.Name + "-" + index

	//     3e: create Service definition for this Deployment
	//		TODO

	return source
}

// Given a deployment, replace the placeholders in the EnvVars in each container
func updateEnvironmentValuesForIndex(index string, source *appsv1.Deployment) {
	newContainers := make([]apiv1.Container, len(source.Spec.Template.Spec.Containers))

	for cindx := range source.Spec.Template.Spec.Containers {
		container := source.Spec.Template.Spec.Containers[cindx]

		newEnv := make([]apiv1.EnvVar, len(container.Env))
		for envidx, env := range container.Env {
			env.Value = strings.Replace(env.Value, "###", index, -1)
			newEnv[envidx] = env
		}
		container.Env = newEnv
		newContainers[cindx] = container
	}
	source.Spec.Template.Spec.Containers = newContainers
}

// Naive clone of the Labels of a LabelSelector. Only clones the LabelSelector, not the MatchExpressions
func cloneLabelSelector(input metav1.LabelSelector) metav1.LabelSelector {
	input.MatchLabels = cloneMap(input.MatchLabels)
	return input
}

func cloneMap(input map[string]string) map[string]string {
	newMap := make(map[string]string)
	for key, value := range input {
		newMap[key] = value
	}
	return newMap
}

func applyDeployments(deployments []appsv1.Deployment, client *kubernetes.Clientset) {
	deploymentClient := client.AppsV1().Deployments(apiv1.NamespaceDefault)

	for _, value := range deployments {
		_, err := deploymentClient.Create(&value)
		if err != nil {
			if serr, ok := err.(*errors.StatusError); ok {
				if serr.ErrStatus.Reason != metav1.StatusReasonAlreadyExists {
					panic(serr)
				}
				fmt.Printf("Deployment (inserthere) already exists; updating instead\n")
				_, err := deploymentClient.Update(&value)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	}
}

func getKubeClient() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigLocation())
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func getKubeConfigLocation() string {
	if value, ok := os.LookupEnv("KUBECONFIG"); ok {
		return value
	}
	if value, ok := os.LookupEnv("HOME"); ok {
		return value + "/.kube/config"
	}
	panic("Dunno where kube config is. Set $KUBECONFIG or put it in $HOME/.kube/config")
}
