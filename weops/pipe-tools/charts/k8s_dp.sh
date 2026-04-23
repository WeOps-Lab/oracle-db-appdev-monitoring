#!/bin/bash

# 部署监控对象
object=oracle
kubectl apply -f ./oracle-monitor-user-init-configmap.yaml -n $object
kubectl apply -f ./oracleDB_11g/svc.yaml -n $object
kubectl apply -f ./oracleDB_11g/statefulset.yaml -n $object
kubectl apply -f ./oracleDB_12c/svc.yaml -n $object
kubectl apply -f ./oracleDB_12c/statefulset.yaml -n $object
kubectl apply -f ./oracleDB_18c/svc.yaml -n $object
kubectl apply -f ./oracleDB_18c/statefulset.yaml -n $object
kubectl apply -f ./oracleDB_19c/pvc.yaml -n $object
kubectl apply -f ./oracleDB_19c/statefulset.yaml -n $object
kubectl apply -f ./oracleDB_21c/pvc.yaml -n $object
kubectl apply -f ./oracleDB_21c/statefulset.yaml -n $object
