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

// This file was automatically generated by lister-gen

package v1beta2

import (
	v1beta2 "github.com/benbromhead/cassandra-operator/pkg/apis/cassandra/v1beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CassandraClusterLister helps list CassandraClusters.
type CassandraClusterLister interface {
	// List lists all CassandraClusters in the indexer.
	List(selector labels.Selector) (ret []*v1beta2.CassandraCluster, err error)
	// CassandraClusters returns an object that can list and get CassandraClusters.
	CassandraClusters(namespace string) CassandraClusterNamespaceLister
	CassandraClusterListerExpansion
}

// cassandraClusterLister implements the CassandraClusterLister interface.
type cassandraClusterLister struct {
	indexer cache.Indexer
}

// NewCassandraClusterLister returns a new CassandraClusterLister.
func NewCassandraClusterLister(indexer cache.Indexer) CassandraClusterLister {
	return &cassandraClusterLister{indexer: indexer}
}

// List lists all CassandraClusters in the indexer.
func (s *cassandraClusterLister) List(selector labels.Selector) (ret []*v1beta2.CassandraCluster, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta2.CassandraCluster))
	})
	return ret, err
}

// CassandraClusters returns an object that can list and get CassandraClusters.
func (s *cassandraClusterLister) CassandraClusters(namespace string) CassandraClusterNamespaceLister {
	return cassandraClusterNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CassandraClusterNamespaceLister helps list and get CassandraClusters.
type CassandraClusterNamespaceLister interface {
	// List lists all CassandraClusters in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta2.CassandraCluster, err error)
	// Get retrieves the CassandraCluster from the indexer for a given namespace and name.
	Get(name string) (*v1beta2.CassandraCluster, error)
	CassandraClusterNamespaceListerExpansion
}

// cassandraClusterNamespaceLister implements the CassandraClusterNamespaceLister
// interface.
type cassandraClusterNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CassandraClusters in the indexer for a given namespace.
func (s cassandraClusterNamespaceLister) List(selector labels.Selector) (ret []*v1beta2.CassandraCluster, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta2.CassandraCluster))
	})
	return ret, err
}

// Get retrieves the CassandraCluster from the indexer for a given namespace and name.
func (s cassandraClusterNamespaceLister) Get(name string) (*v1beta2.CassandraCluster, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta2.Resource("cassandracluster"), name)
	}
	return obj.(*v1beta2.CassandraCluster), nil
}
