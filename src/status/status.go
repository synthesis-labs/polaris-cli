package status

import (
	"fmt"

	polarisv1alpha1 "github.com/synthesis-labs/polaris-client/pkg/client/clientset/versioned"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PrintPolarisStatus prints the status of the project
//
func PrintPolarisStatus(projectName string, client *kubernetes.Clientset, apiextensionClient *apiextension.Clientset, polarisClient *polarisv1alpha1.Clientset, namespace string) error {

	// Define a common polaris project selector
	//
	thisProjectSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("polaris-project=%s", projectName),
	}

	// Start with deployments
	//
	deployments, err := client.AppsV1().Deployments(namespace).List(thisProjectSelector)
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {

		// Find the polaris-component label for this deployment
		//
		component := deployment.Labels["polaris-component"]
		if component == "" {
			return fmt.Errorf("Unable to determine polaris-component from deployment %s", deployment.Name)
		}

		// Define a common component selector
		//
		thisComponentSelector := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("polaris-project=%s,polaris-component=%s", projectName, component),
		}

		// Find the associated container registry
		//
		containerRegistries, err := polarisClient.PolarisV1alpha1().PolarisContainerRegistries(namespace).List(thisComponentSelector)
		if err != nil {
			return err
		}

		// Find the stack associated - using the same labels
		//
		stacks, err := polarisClient.PolarisV1alpha1().PolarisStacks(namespace).List(thisComponentSelector)
		if err != nil {
			return err
		}

		// Find the buildsteps associated - using the same labels
		//
		steps, err := polarisClient.PolarisV1alpha1().PolarisBuildSteps(namespace).List(thisComponentSelector)
		if err != nil {
			return err
		}

		fmt.Println("Deployment:", deployment.Name,
			"(",
			"replicas", deployment.Status.Replicas,
			"ready", deployment.Status.ReadyReplicas,
			")",
		)

		for _, containerRegistry := range containerRegistries.Items {
			fmt.Println("  ContainerRegistry:", containerRegistry.Name, "(", "current", "abcdefgh", "available", "bcdefghi", ")")
		}
		for _, stack := range stacks.Items {
			fmt.Println("              Stack:", stack.Name, "(", stack.Status.Status, ")")
		}
		for _, step := range steps.Items {
			fmt.Println("               Step:", step.Name, "(GET THE STATUS)")
		}
	}

	return nil
}
