// Copyright 2016 The cassandra-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/benbromhead/cassandra-operator/pkg/client"
	"github.com/benbromhead/cassandra-operator/pkg/generated/clientset/versioned"
	"github.com/benbromhead/cassandra-operator/pkg/util/constants"
	"github.com/benbromhead/cassandra-operator/pkg/util/k8sutil"
	"github.com/benbromhead/cassandra-operator/pkg/util/probe"
	"github.com/benbromhead/cassandra-operator/pkg/util/retryutil"
	"github.com/benbromhead/cassandra-operator/test/e2e/e2eutil"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var Global *Framework

type Framework struct {
	opImage    string
	KubeClient kubernetes.Interface
	CRClient   versioned.Interface
	Namespace  string
	S3Cli      *s3.S3
	S3Bucket   string
}

// Setup setups a test framework and points "Global" to it.
func Setup() error {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	opImage := flag.String("operator-image", "", "operator image, e.g. gcr.io/coreos-k8s-scale-testing/cassandra-operator")
	ns := flag.String("namespace", "default", "e2e test namespace")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	Global = &Framework{
		KubeClient: cli,
		CRClient:   client.MustNew(config),
		Namespace:  *ns,
		opImage:    *opImage,
		S3Bucket:   os.Getenv("TEST_S3_BUCKET"),
	}
	return Global.setup()
}

func Teardown() error {
	if err := Global.deleteEtcdOperator(); err != nil {
		return err
	}
	// TODO: check all deleted and wait
	Global = nil
	logrus.Info("e2e teardown successfully")
	return nil
}

func (f *Framework) setup() error {
	if err := f.SetupEtcdOperator(); err != nil {
		return fmt.Errorf("failed to setup cassandra operator: %v", err)
	}
	logrus.Info("cassandra operator created successfully")
	if os.Getenv("AWS_TEST_ENABLED") == "true" {
		if err := f.setupAWS(); err != nil {
			return fmt.Errorf("fail to setup aws: %v", err)
		}
	}
	logrus.Info("e2e setup successfully")
	return nil
}

func (f *Framework) SetupEtcdOperator() error {
	// TODO: unify this and the yaml file in example/
	cmd := []string{"/usr/local/bin/cassandra-operator"}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cassandra-operator",
			Labels: map[string]string{"name": "cassandra-operator"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "cassandra-operator",
					Image:           f.opImage,
					ImagePullPolicy: v1.PullAlways,
					Command:         cmd,
					Env: []v1.EnvVar{
						{
							Name:      constants.EnvOperatorPodNamespace,
							ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
						},
						{
							Name:      constants.EnvOperatorPodName,
							ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"}},
						},
					},
					ReadinessProbe: &v1.Probe{
						Handler: v1.Handler{
							HTTPGet: &v1.HTTPGetAction{
								Path: probe.HTTPReadyzEndpoint,
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
							},
						},
						InitialDelaySeconds: 3,
						PeriodSeconds:       3,
						FailureThreshold:    3,
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	p, err := k8sutil.CreateAndWaitPod(f.KubeClient, f.Namespace, pod, 60*time.Second)
	if err != nil {
		// assuming `kubectl` installed on $PATH
		cmd := exec.Command("kubectl", "-n", f.Namespace, "describe", "pod", "cassandra-operator")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Run() // Just ignore the error...
		logrus.Info("describing cassandra-operator pod:", out.String())
		return err
	}
	logrus.Infof("cassandra operator pod is running on node (%s)", p.Spec.NodeName)

	return e2eutil.WaitUntilOperatorReady(f.KubeClient, f.Namespace, "cassandra-operator")
}

func (f *Framework) DeleteEtcdOperatorCompletely() error {
	err := f.deleteEtcdOperator()
	if err != nil {
		return err
	}
	// On k8s 1.6.1, grace period isn't accurate. It took ~10s for operator pod to completely disappear.
	// We work around by increasing the wait time. Revisit this later.
	err = retryutil.Retry(5*time.Second, 6, func() (bool, error) {
		_, err := f.KubeClient.CoreV1().Pods(f.Namespace).Get("cassandra-operator", metav1.GetOptions{})
		if err == nil {
			return false, nil
		}
		if k8sutil.IsKubernetesResourceNotFoundError(err) {
			return true, nil
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("fail to wait cassandra operator pod gone from API: %v", err)
	}
	return nil
}

func (f *Framework) deleteEtcdOperator() error {
	return f.KubeClient.CoreV1().Pods(f.Namespace).Delete("cassandra-operator", metav1.NewDeleteOptions(1))
}

func (f *Framework) setupAWS() error {
	if err := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", os.Getenv("AWS_CREDENTIAL")); err != nil {
		return err
	}
	if err := os.Setenv("AWS_CONFIG_FILE", os.Getenv("AWS_CONFIG")); err != nil {
		return err
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return err
	}
	f.S3Cli = s3.New(sess)
	return nil
}
