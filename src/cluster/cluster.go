package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Check for KUBECONFIG environment variable, then otherwise find it in local directory ~/.kube/config
//
func getKubeConfig() (string, error) {
	var kubeconfig string

	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}

	kubeconfig = filepath.Join(home, ".kube/config")

	return kubeconfig, nil
}

// ConnectToCluster connects and returns the kubernetes client
//
func ConnectToCluster() (*kubernetes.Clientset, *apiextension.Clientset, string, error) {
	kubeConfigPath, err := getKubeConfig()
	if err != nil {
		return nil, nil, "", err
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: kubeConfigPath,
		},
		&clientcmd.ConfigOverrides{},
	)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, "", err
	}

	// create the clientset
	//
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create the apiextensions client
	//
	apiextensionsClient, err := apiextension.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Lookup the namespace we might be configured to
	//
	ns, _, _ := clientConfig.Namespace()

	return clientset, apiextensionsClient, ns, err
}

// EnsureOperatorInstalled ensures and installs the operator plus whatever depedencies
//
func EnsureOperatorInstalled(client *kubernetes.Clientset, apiextensionClient *apiextension.Clientset, namespace string, environmentVariables map[string]string) error {
	// SERVICE ACCOUNT transcribed from the operator-sdk deploy/folder
	//
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polaris-operator",
		},
	}
	fmt.Print("ServiceAccount Creating... ")
	_, err := client.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Updating... ")
		_, err = client.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// ROLE transcribed from the operator-sdk deploy/folder
	//
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polaris-operator",
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
					"services",
					"endpoints",
					"persistentvolumeclaims",
					"events",
					"configmaps",
					"secrets",
				},
				Verbs: []string{
					"*",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"namespaces",
				},
				Verbs: []string{
					"get",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
					"daemonsets",
					"replicasets",
					"statefulsets",
				},
				Verbs: []string{
					"*",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{
					"monitoring.coreos.com",
				},
				Resources: []string{
					"servicemonitors",
				},
				Verbs: []string{
					"get",
					"create",
				},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{
					"polaris.synthesis.co.za",
				},
				Resources: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
	}
	fmt.Print("Role Creating... ")
	_, err = client.RbacV1().Roles(namespace).Create(role)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Updating... ")
		_, err = client.RbacV1().Roles(namespace).Update(role)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// ROLEBINDING transcribed from the operator-sdk deploy/folder
	//
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polaris-operator",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "polaris-operator",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      "polaris-operator",
				Namespace: namespace,
			},
		},
	}
	fmt.Print("RoleBinding Creating... ")
	_, err = client.RbacV1().RoleBindings(namespace).Create(roleBinding)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Updating... ")
		_, err = client.RbacV1().RoleBindings(namespace).Update(roleBinding)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// CRD - polarisstacks.polaris.synthesis.co.za transcribed from the operator-sdk deploy/folder
	//
	polarisStackCrd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polarisstacks.polaris.synthesis.co.za",
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "polaris.synthesis.co.za",
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Kind:     "PolarisStack",
				ListKind: "PolarisStackList",
				Plural:   "polarisstacks",
				Singular: "polarisstack",
				ShortNames: []string{
					"ps",
					"stack",
					"stacks",
				},
				Categories: []string{
					"all",
				},
			},
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Version: "v1alpha1",
		},
	}
	fmt.Print("CustomResourceDefinition (polarisstacks.polaris.synthesis.co.za) Creating... ")
	_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisStackCrd)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Deleting... ")
		err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("polarisstacks.polaris.synthesis.co.za", &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		fmt.Print("Waiting... ")
		time.Sleep(2 * time.Second)
		fmt.Print("Creating... ")
		_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisStackCrd)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// CRD - polarissourcerepositories.polaris.synthesis.co.za transcribed from the operator-sdk deploy/folder
	//
	polarisSourceRepositoryCrd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polarissourcerepositories.polaris.synthesis.co.za",
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "polaris.synthesis.co.za",
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Kind:     "PolarisSourceRepository",
				ListKind: "PolarisSourceRepositoryList",
				Plural:   "polarissourcerepositories",
				Singular: "polarissourcerepository",
				ShortNames: []string{
					"psr",
					"sourcerepository",
					"repository",
				},
				Categories: []string{
					"all",
				},
			},
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Version: "v1alpha1",
		},
	}
	fmt.Print("CustomResourceDefinition (polarissourcerepositories.polaris.synthesis.co.za) Creating... ")
	_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisSourceRepositoryCrd)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Deleting... ")
		err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("polarissourcerepositories.polaris.synthesis.co.za", &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		fmt.Print("Waiting... ")
		time.Sleep(2 * time.Second)
		fmt.Print("Creating... ")
		_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisSourceRepositoryCrd)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// CRD - polariscontainerregistries.polaris.synthesis.co.za transcribed from the operator-sdk deploy/folder
	//
	polarisContainerRegistryCrd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polariscontainerregistries.polaris.synthesis.co.za",
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "polaris.synthesis.co.za",
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Kind:     "PolarisContainerRegistry",
				ListKind: "PolarisContainerRegistryList",
				Plural:   "polariscontainerregistries",
				Singular: "polariscontainerregistry",
				ShortNames: []string{
					"pcr",
					"containerregistry",
					"registry",
				},
				Categories: []string{
					"all",
				},
			},
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Version: "v1alpha1",
		},
	}
	fmt.Print("CustomResourceDefinition (polariscontainerregistries.polaris.synthesis.co.za) Creating... ")
	_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisContainerRegistryCrd)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Deleting... ")
		err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("polariscontainerregistries.polaris.synthesis.co.za", &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		fmt.Print("Waiting... ")
		time.Sleep(2 * time.Second)
		fmt.Print("Creating... ")
		_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisContainerRegistryCrd)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// CRD - polariscontainerregistries.polaris.synthesis.co.za transcribed from the operator-sdk deploy/folder
	//
	polarisBuildPipelineCrd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polarisbuildpipelines.polaris.synthesis.co.za",
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "polaris.synthesis.co.za",
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Kind:     "PolarisBuildPipeline",
				ListKind: "PolarisBuildPipelineList",
				Plural:   "polarisbuildpipelines",
				Singular: "polarisbuildpipeline",
				ShortNames: []string{
					"pbp",
					"buildpipeline",
					"ci",
				},
				Categories: []string{
					"all",
				},
			},
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Version: "v1alpha1",
		},
	}
	fmt.Print("CustomResourceDefinition (polarisbuildpipelines.polaris.synthesis.co.za) Creating... ")
	_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisBuildPipelineCrd)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Deleting... ")
		err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("polarisbuildpipelines.polaris.synthesis.co.za", &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		fmt.Print("Waiting... ")
		time.Sleep(2 * time.Second)
		fmt.Print("Creating... ")
		_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisBuildPipelineCrd)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// CRD - polarisbuildsteps.polaris.synthesis.co.za transcribed from the operator-sdk deploy/folder
	//
	polarisBuildStepCrd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polarisbuildsteps.polaris.synthesis.co.za",
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: "polaris.synthesis.co.za",
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Kind:     "PolarisBuildStep",
				ListKind: "PolarisBuildStepList",
				Plural:   "polarisbuildsteps",
				Singular: "polarisbuildstep",
				ShortNames: []string{
					"pbs",
					"step",
					"steps",
				},
				Categories: []string{
					"all",
				},
			},
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Version: "v1alpha1",
		},
	}
	fmt.Print("CustomResourceDefinition (polarisbuildsteps.polaris.synthesis.co.za) Creating... ")
	_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisBuildStepCrd)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Deleting... ")
		err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("polarisbuildsteps.polaris.synthesis.co.za", &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		fmt.Print("Waiting... ")
		time.Sleep(2 * time.Second)
		fmt.Print("Creating... ")
		_, err = apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(polarisBuildStepCrd)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	// DEPLOYMENT - transcribed from the operator-sdk deploy/folder
	//
	var replicas int32 = 1
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "polaris-operator",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "polaris-operator",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "polaris",
						"name": "polaris-operator",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "polaris-operator",
					Containers: []v1.Container{
						v1.Container{
							Name:            "polaris-operator",
							Image:           "tomwells/polaris-operator:1.0.0",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: 60000,
									Name:          "metrics",
								},
							},
							Command: []string{
								"polaris-operator",
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"stat",
											"/tmp/operator-sdk-ready",
										},
									},
								},
								InitialDelaySeconds: 4,
								PeriodSeconds:       10,
								FailureThreshold:    1,
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name:  "OPERATOR_NAME",
									Value: "polaris-operator",
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Print("Deployment Creating... ")

	// Add additional environment variables provided
	//
	for k, v := range environmentVariables {
		fmt.Println(" Adding variable", k, v)
		newVar := v1.EnvVar{Name: k, Value: v}
		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, newVar)
	}

	_, err = client.AppsV1().Deployments(namespace).Create(deployment)
	if k8serrors.IsAlreadyExists(err) {
		fmt.Print("Updating... ")
		_, err = client.AppsV1().Deployments(namespace).Update(deployment)
	}
	if err != nil {
		return err
	}
	fmt.Print("Done!\n")

	return nil
}

// GetPolarisOperator finds and returns the pod(s) running the operator
//
func GetPolarisOperator(client *kubernetes.Clientset, apiextensionClient *apiextension.Clientset, namespace string) error {
	podsClient := client.CoreV1().Pods(namespace)

	pods, err := podsClient.List(metav1.ListOptions{
		LabelSelector: "app=polaris,name=polaris-operator",
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Println("Found polaris-operator", pod.Name)
	}

	if len(pods.Items) == 0 {
		fmt.Println("Polaris operator not found. Installing now...")

	}

	return nil
}
