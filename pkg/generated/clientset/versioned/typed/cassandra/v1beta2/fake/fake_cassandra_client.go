/*
Copyright 2017 The etcd-operator Authors

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
package fake

import (
	v1beta2 "github.com/benbromhead/cassandra-operator/pkg/generated/clientset/versioned/typed/cassandra/v1beta2"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCassandraV1beta2 struct {
	*testing.Fake
}

func (c *FakeCassandraV1beta2) CassandraClusters(namespace string) v1beta2.CassandraClusterInterface {
	return &FakeCassandraClusters{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCassandraV1beta2) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
