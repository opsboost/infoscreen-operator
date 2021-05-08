/*
Copyright 2021 OpsBoost Crew <info@opsboost.dev>.

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

package session

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kmapi "kmodules.xyz/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sessionv1 "github.com/OpsBoost/infoscreen-operator/apis/session/v1"
)

// FirefoxReconciler reconciles a Firefox object
type FirefoxReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=session.opsboost.dev,resources=firefoxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=session.opsboost.dev,resources=firefoxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=session.opsboost.dev,resources=firefoxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Firefox object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *FirefoxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("firefox", req.NamespacedName)

	var firefox sessionv1.Firefox
	if err := r.Get(ctx, req.NamespacedName, &firefox); err != nil {
		log.Error(err, "unable to fetch Firefox")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var (
		immediateTermination         int64 = 0
		doNotAutomountServiceAccount       = false
		defaultTempDirSize                 = resource.MustParse("128Mi")
		blockOwnerDeletion                 = true
		ownerIsController                  = true
		err                          error
	)

	firefox.Status.ObservedGeneration = firefox.Generation
	firefox.Status.Phase = sessionv1.ScreenStatusProvisioning
	firefox.Status.Conditions = append(firefox.Status.Conditions, kmapi.Condition{
		Type:               string(sessionv1.ScreenStatusProvisioning),
		Status:             corev1.ConditionTrue,
		ObservedGeneration: firefox.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Changed",
		Message:            "Creating child resources",
	})

	//err = r.Status().Update(context.Background(), &firefox)
	//if err != nil {
	//	return reconcile.Result{}, err
	//}

	envVars := []corev1.EnvVar{
		{
			Name:  "XDG_RUNTIME_DIR",
			Value: "/tmp",
		},
		{
			Name:  "WLR_BACKENDS",
			Value: "headless",
		},
		{
			Name:  "WLR_LIBINPUT_NO_DEVICES",
			Value: "1",
		},
		{
			Name:  "SWAYSOCK",
			Value: "/tmp/sway-ipc.sock",
		},
		{
			Name:  "MOZ_ENABLE_WAYLAND",
			Value: "1",
		},
		{
			Name:  "URL",
			Value: firefox.Spec.Url,
		},
		{
			Name:  "DEBUG",
			Value: "true",
		},
	}

	if firefox.Spec.Target != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TARGET",
			Value: firefox.Spec.Target,
		})
	}

	if firefox.Spec.Credentials != nil {
		if firefox.Spec.Credentials.SecretRef != nil {
			envVars = append(envVars, corev1.EnvVar{
				Name: "LOGIN_USER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: firefox.Spec.Credentials.SecretRef.Name,
						},
						Key: "GF_SECURITY_ADMIN_USER",
					},
				},
			})
			envVars = append(envVars, corev1.EnvVar{
				Name:  "LOGIN_PW_BASE64",
				Value: "false",
			})
			envVars = append(envVars, corev1.EnvVar{
				Name: "LOGIN_PW",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: firefox.Spec.Credentials.SecretRef.Name,
						},
						Key: "GF_SECURITY_ADMIN_PASSWORD",
					},
				},
			})
		}
	}

	if firefox.Spec.Destination != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "URL_PAYLOAD",
			Value: firefox.Spec.Destination,
		})
	}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      firefox.Name,
			Namespace: firefox.Namespace,
			Labels:    firefox.ObjectMeta.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         firefox.TypeMeta.APIVersion,
					Kind:               firefox.TypeMeta.Kind,
					Name:               firefox.Name,
					UID:                firefox.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Controller:         &ownerIsController,
				},
			},
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "temp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    "Memory",
							SizeLimit: &defaultTempDirSize,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:            "firefox",
					Image:           "swayvnc-firefox:latest",
					ImagePullPolicy: corev1.PullNever,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "temp",
							MountPath: "/tmp",
						},
					},
					Env: envVars,
					Ports: []corev1.ContainerPort{
						{
							Name:          "vnc",
							ContainerPort: 5900,
						},
						{
							Name:          "healthz",
							ContainerPort: 5000,
						},
						{
							Name:          "sway-ipc",
							ContainerPort: 7023,
						},
					},
				},
			},
			RestartPolicy:                 corev1.RestartPolicyOnFailure,
			TerminationGracePeriodSeconds: &immediateTermination,
			AutomountServiceAccountToken:  &doNotAutomountServiceAccount,
		},
	}

	if err = r.Create(ctx, &pod); err != nil {
		return ctrl.Result{}, err
	}

	vncService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      firefox.Name,
			Namespace: firefox.Namespace,
			Labels:    firefox.ObjectMeta.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         firefox.TypeMeta.APIVersion,
					Kind:               firefox.TypeMeta.Kind,
					Name:               firefox.Name,
					UID:                firefox.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Controller:         &ownerIsController,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Port:       5900,
					TargetPort: intstr.FromString("vnc"),
				},
			},
			Selector: firefox.Labels,
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	if err = r.Create(ctx, &vncService); err != nil {
		return ctrl.Result{}, err
	}

	internalService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      (firefox.Name + "-pods"),
			Namespace: firefox.Namespace,
			Labels:    firefox.ObjectMeta.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         firefox.TypeMeta.APIVersion,
					Kind:               firefox.TypeMeta.Kind,
					Name:               firefox.Name,
					UID:                firefox.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Controller:         &ownerIsController,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Port:       5900,
					TargetPort: intstr.FromString("vnc"),
				},
				{
					Name:       "sway-ipc",
					Port:       7023,
					TargetPort: intstr.FromString("swap-ipc"),
				},
			},
			Selector:  firefox.Labels,
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	if err = r.Create(ctx, &internalService); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FirefoxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sessionv1.Firefox{}).
		Complete(r)
}
