# Upgrade Guide

This document shows how to safely upgrade the operator to a desired version while preserving the cluster's state and data whenever possible. It is assumed that the preexisting cluster is configured to create and store backups to persistent storage. See the [backup config guide](../backup_config.md) for details.

### Backup safety precaution:
First create a backup of your current cluster before starting the upgrade process. See the [backup service guide](../backup_service.md) on how to create a backup.

In the case of an upgrade failure you can restore your cluster to the previous state from the previous backup. See the [spec examples](https://github.com/benbromhead/cassandra-operator/blob/master/doc/user/spec_examples.md#three-members-cluster-that-restores-from-previous-pv-backup) on how to do that.


### Upgrade operator deployment
An [in-place update](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/#in-place-updates-of-resources) can be performed when the upgrade is compatible, i.e we can upgrade the operator without affecting the cluster.

To upgrade an operator deployment the image field `spec.template.spec.containers.image` needs to be changed via an in-place update.

Change the image field to `quay.io/benbromhead/cassandra-operator:vX.Y.Z`  where `vX.Y.Z` is the desired version.
```bash
$ kubectl edit deployment/etcd-operator
# make the image change in your editor then save and close the file
```

### Incompatible upgrade
In the case of an incompatible upgrade, the process requires restoring a new cluster from backup. See the [incompatible upgrade guide](incompatible_upgrade.md) for more information.


## v0.5.x -> v0.6.0

### Breaking Change

The v0.6.0 release removes [operator S3 flag](https://github.com/benbromhead/cassandra-operator/blob/v0.5.1/doc/user/backup_config.md#operator-level-configuration).

**Note:** if your cluster is not using operator S3 flag, then you just need recreate etcd-operator deployment with the `v0.6.0` image.

If your cluster is using operator S3 flag for backup and want to use S3 backup in v0.6.0, then you need to migrate your cluster to use [cluster level backup](../backup_config.md#S3-on-aws).

Steps for migration:

- Create a backup of you current data using the [backup service guide](../backup_service.md#http-api-v1).

- Create a new AWS secret `<aws-secret>` for [cluster level backup](../backup_config.md#s3-on-aws):

  `$ kubectl -n <namespace> create secret generic <aws-secret> --from-file=$AWS_DIR/credentials --from-file=$AWS_DIR/config`

- Change the backup field in your existing cluster backup spec to enable [cluster level backup](../backup_config.md#s3-on-aws):

  From

  ```yaml
  backup:
    backupIntervalInSecond: 30
    maxBackups: 5
    storageType: "S3"
  ```

  to

  ```yaml
  backup:
    backupIntervalInSecond: 30
    maxBackups: 5
    storageType: "S3"
    s3:
      s3Bucket: <your-s3-bucket>
      awsSecret: <aws-secret> 
  ```

- Apply the cluster backup spec:
  `$ kubectl -n <namespace> apply -f <your-cluster-deployment>.yaml`

- Update deployment spec to use `etcd-operator:v0.6.0` and remove the dependency on operator level S3 backup flags:

  From

  ```yaml
  spec:
    containers:
    - name: etcd-operator
      image: quay.io/benbromhead/cassandra-operator:v0.5.2
      command: 
        - /usr/local/bin/etcd-operator
        - --backup-aws-secret=aws
        - --backup-aws-config=aws
        - --backup-s3-bucket=<your-s3-bucket>
  ```

  to 

  ```yaml
  spec:
    containers:
    - name: etcd-operator
      image: quay.io/benbromhead/cassandra-operator:v0.6.0
      command: 
        - /usr/local/bin/etcd-operator
  ```

- Apply the updated deployment spec:
  `$ kubectl -n <namespace> apply -f <your-etcd-operator-deployment>.yaml`

## v0.4.x -> v0.5.x
For any `0.4.x` versions, please update to `0.5.0` first.

The `0.5.0` release introduces a breaking change in moving from [TPR](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/) to [CRD](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/). To preserve the cluster state across the upgrade, the cluster must be recreated from backup after the upgrade.
### Prerequisite:
Kubernetes cluster version must be 1.7+.

### Steps to upgrade:
- Create a backup of your current cluster data. See the following guides on how to enable and create backups:
    - [backup config guide](https://github.com/benbromhead/cassandra-operator/blob/master/doc/user/backup_config.md) to enable backups fro your cluster
    - [backup service guide](https://github.com/benbromhead/cassandra-operator/blob/master/doc/user/backup_service.md) to create a backup
- Delete the cluster TPR object. The etcd-operator will delete all resources(pods, services, deployments) associated with the cluster:
    - `kubectl -n <namespace> delete cluster <cluster-name>`
- Delete the etcd-operator deployment
- Delete the TPR
    - `kubectl delete thirdpartyresource cluster.etcd.coreos.com`
- Replace the existing RBAC rules for the etcd-operator by editing or recreating the `ClusterRole` with the [new rules for CRD](https://github.com/benbromhead/cassandra-operator/blob/master/doc/user/rbac.md#create-clusterrole).
- Recreate the etcd-operator deployment with the `0.5.0` image.
- Create a new cluster that restores from the backup of the previous cluster. The new cluster CR spec should look as follows:
```
apiVersion: "etcd.database.coreos.com/v1beta2"
kind: "EtcdCluster"
metadata:
  name: <cluster-name>
spec:
  size: <cluster-size>
  version: "3.1.8"
  backup:
    # The same backup spec used to save the backup of the previous cluster
    . . .
    . . .
  restore:
    backupClusterName: <previous-cluster-name>
    storageType: <storage-type-of-backup-spec>
```
The two points of interest in the above CR spec are:
  1. The `apiVersion` and `kind` fields have been changed as mentioned in the `0.5.0` release notes
  2. The `spec.restore` field needs to be specified according your backup configuration. See [spec examples guide](https://github.com/benbromhead/cassandra-operator/blob/master/doc/user/spec_examples.md#three-members-cluster-that-restores-from-previous-pv-backup) on how to specify the `spec.restore` field for your particular backup configuration.

## v0.3.x -> v0.4.x
Upgrade to `v0.4.0` first.
See the release notes of `v0.4.0` for noticeable changes: https://github.com/benbromhead/cassandra-operator/releases/tag/v0.4.0

## v0.2.x -> v0.3.x
### Prerequisite:
To upgrade to `v0.3.x` the current operator verison **must** first be upgraded to `v0.2.6+`, since versions < `v0.2.6` are not compatible with `v0.3.x` .

### Noticeable changes:
- `Spec.Backup.MaxBackups` update:
  - If you have `Spec.Backup.MaxBackups < 0`, previously it had no effect.
    Now it will get rejected. Please remove it.
  - If you have `Spec.Backup.MaxBackups == 0`, previously it had no effect.
    Now it will create backup sidecar that allows unlimited backups.
    If you don't want backup sidecar, just don't set any backup policy.
