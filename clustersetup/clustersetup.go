package clustersetup

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	clustertemplatev1alpha1 "github.com/rawagner/cluster-templates-operator/api/v1alpha1"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterSetupLabel    = "clustertemplate.openshift.io/setup-name"
	ClusterSetupInstance = "clustertemplate.openshift.io/cluster-instance"
)

func CreateSetupPipelines(
	ctx context.Context,
	log logr.Logger,
	k8sClient client.Client,
	clusterTemplate clustertemplatev1alpha1.ClusterTemplate,
	clusterTemplateInstance *clustertemplatev1alpha1.ClusterTemplateInstance,
	kubeconfigSecret string,
) error {
	for _, clusterSetup := range clusterTemplate.Spec.ClusterSetup {
		log.Info("Create PipelineRun", "name", clusterSetup.Name)

		pipelineRun := &pipeline.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: clusterTemplateInstance.Name + "-",
				Namespace:    clusterTemplateInstance.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "ClusterTemplateInstance",
						APIVersion: clustertemplatev1alpha1.APIVersion,
						Name:       clusterTemplateInstance.Name,
						UID:        clusterTemplateInstance.UID,
					},
				},
				Labels: map[string]string{ClusterSetupInstance: clusterTemplateInstance.Name, ClusterSetupLabel: clusterSetup.Name},
			},
			Spec: pipeline.PipelineRunSpec{
				PipelineRef: &clusterSetup.PipelineRef,
				Workspaces: []pipeline.WorkspaceBinding{
					{
						Name: "kubeconfigSecret",
						Secret: &corev1.SecretVolumeSource{
							SecretName: kubeconfigSecret,
						},
					},
				},
			},
		}
		log.Info("Submit PipelineRun ", "name", clusterSetup.Name)
		err := k8sClient.Create(ctx, pipelineRun)
		if err != nil {
			return err
		}
	}
	return nil
}
