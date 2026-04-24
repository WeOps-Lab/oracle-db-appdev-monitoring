#!/bin/bash

# 删除监控对象
object=oracle
kubectl delete -f ./oracleDB_11g/statefulset.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_11g/svc.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_12c/statefulset.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_12c/svc.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_18c/statefulset.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_18c/svc.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_19c/statefulset.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_19c/pvc.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_21c/statefulset.yaml -n $object --ignore-not-found
kubectl delete -f ./oracleDB_21c/pvc.yaml -n $object --ignore-not-found
kubectl delete configmap oracle-monitor-user-init -n $object --ignore-not-found
