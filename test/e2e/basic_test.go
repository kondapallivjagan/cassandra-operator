// Copyright 2017 The cassandra-operator Authors
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

package e2e

import (
	"os"
	"testing"
	"time"

	api "github.com/benbromhead/cassandra-operator/pkg/apis/cassandra/v1beta2"
	"github.com/benbromhead/cassandra-operator/test/e2e/e2eutil"
	"github.com/benbromhead/cassandra-operator/test/e2e/framework"
)

var retryCount = 20

func TestCreateCluster(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}
	f := framework.Global
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, e2eutil.NewCluster("test-cassandra-", 3))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, retryCount, testEtcd); err != nil {
		t.Fatalf("failed to create 3 members cassandra cluster: %v", err)
	}
}

// TestPauseControl tests the user can pause the operator from controlling
// an cassandra cluster.
func TestPauseControl(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}

	f := framework.Global
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, e2eutil.NewCluster("test-cassandra-", 3))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	names, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, retryCount, testEtcd)
	if err != nil {
		t.Fatalf("failed to create 3 members cassandra cluster: %v", err)
	}

	updateFunc := func(cl *api.CassandraCluster) {
		cl.Spec.Paused = true
	}
	if testEtcd, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, retryCount, updateFunc); err != nil {
		t.Fatalf("failed to pause control: %v", err)
	}

	// TODO: this is used to wait for the CR to be updated.
	// TODO: make this wait for reliable
	time.Sleep(5 * time.Second)

	if err := e2eutil.KillMembers(f.KubeClient, f.Namespace, names[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := e2eutil.WaitUntilPodSizeReached(t, f.KubeClient, 2, retryCount, testEtcd); err != nil {
		t.Fatalf("failed to wait for killed member to die: %v", err)
	}
	if _, err := e2eutil.WaitUntilPodSizeReached(t, f.KubeClient, 3, retryCount, testEtcd); err == nil {
		t.Fatalf("cluster should not be recovered: control is paused")
	}

	updateFunc = func(cl *api.CassandraCluster) {
		cl.Spec.Paused = false
	}
	if testEtcd, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, retryCount, updateFunc); err != nil {
		t.Fatalf("failed to resume control: %v", err)
	}

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, retryCount, testEtcd); err != nil {
		t.Fatalf("failed to resize to 3 members cassandra cluster: %v", err)
	}
}

//func TestCassandraUpgrade(t *testing.T) {
//	if os.Getenv(envParallelTest) == envParallelTestTrue {
//		t.Parallel()
//	}
//	f := framework.Global
//	origEtcd := e2eutil.NewCluster("test-cassandra-", 3)
//	origEtcd = e2eutil.ClusterWithVersion(origEtcd, "3.0.16")
//	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, origEtcd)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defer func() {
//		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	err = e2eutil.WaitSizeAndVersionReached(t, f.KubeClient, "3.0.16", 3, retryCount, testEtcd)
//	if err != nil {
//		t.Fatalf("failed to create 3 members cassandra cluster: %v", err)
//	}
//
//	updateFunc := func(cl *api.CassandraCluster) {
//		cl = e2eutil.ClusterWithVersion(cl, "3.1.8")
//	}
//	_, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, retryCount, updateFunc)
//	if err != nil {
//		t.Fatalf("fail to update cluster version: %v", err)
//	}
//
//	// We have seen in k8s 1.7.1 env it took 35s for the pod to restart with the new image.
//	err = e2eutil.WaitSizeAndVersionReached(t, f.KubeClient, "3.1.8", 3, retryCount, testEtcd)
//	if err != nil {
//		t.Fatalf("failed to wait new version cassandra cluster: %v", err)
//	}
//}
