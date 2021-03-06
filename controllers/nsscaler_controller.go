/*
Copyright 2021.

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

package controllers

import (
	"context"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	operatorsv1alpha1 "github.com/lwabish/namespace-scaler-operator/api/v1alpha1"
)

// NSScalerReconciler reconciles a NSScaler object
type NSScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operators.wubw.fun,resources=nsscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.wubw.fun,resources=nsscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operators.wubw.fun,resources=nsscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the NSScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
// todo: 借鉴lru，理论上来讲，一个命名空间所有pod里最后更新过的pod的时间可以作为命名空间的最后活跃时间
// 增加参数：保留活跃命名空间的个数，按最后活跃时间排序，超过保留个数的命名空间里的deploy全部缩成0
// todo: 增加一个无用的spec参数，用来手动触发协调循环
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *NSScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// your logic here
	instance := &operatorsv1alpha1.NSScaler{}
	err := r.Get(ctx, req.NamespacedName, instance)
	logger.Info("starting")
	if err != nil {
		// cr被删了
		if errors.IsNotFound(err) {
			logger.Info("deleted", "error", err)
			// 如果需要清理
			return ctrl.Result{}, nil
		}
		logger.Info("error encountered", "error", err)
		// 读取错误，重新协调
		return ctrl.Result{}, err
	}

	// 获取ns列表
	nsList := &coreV1.NamespaceList{}
	err = r.List(ctx, nsList, &client.ListOptions{})
	if err != nil {
		logger.Error(err, "error list namespace")
		return ctrl.Result{}, err
	}

	var protectedNamespaces = make(map[string]struct{}, len(instance.Spec.ActiveNamespaceSuffixes))
	var p = instance.Spec.ScopePrefix
	var noPodObserved = true

	for _, s := range instance.Spec.ActiveNamespaceSuffixes {
		protectedNamespaces[p+s] = struct{}{}
	}

	for _, v := range nsList.Items {
		thisNamespace := v.Name
		if strings.HasPrefix(thisNamespace, p) {
			if _, ok := protectedNamespaces[thisNamespace]; !ok {
				// 符合前缀的命名空间，且不处于活跃状态的
				logger.Info("found target ns", "namespace", thisNamespace)

				// 拿到所有deploy
				deploys := &appsV1.DeploymentList{}
				err = r.List(ctx, deploys, &client.ListOptions{Namespace: thisNamespace})
				if err != nil {
					logger.Error(err, "err list deploy")
					return ctrl.Result{}, err
				}

				// 把replica不等于0的scale到0
				for _, d := range deploys.Items {
					if d.Spec.Replicas != nil && *d.Spec.Replicas > 0 {
						*d.Spec.Replicas = 0
						err = r.Update(ctx, &d, &client.UpdateOptions{})
						if err != nil {
							logger.Error(err, "err update deploy")
							return ctrl.Result{}, err
						}
						logger.Info("NsScaler scaled[namespace+name]:", d.Namespace, d.Name)
					}
				}

				// 尝试维护状态
				if !noPodObserved {
					pods := &coreV1.PodList{}
					err = r.List(ctx, pods, &client.ListOptions{Namespace: thisNamespace})
					if len(pods.Items) > 0 {
						noPodObserved = false
					}
				}
			}
		}
	}

	// 更新cr的status
	logger.Info("reconcile result status: ", "noPodObserved", noPodObserved)
	instance.Status.Done = noPodObserved
	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NSScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.NSScaler{}).
		Complete(r)
}
