#!/bin/bash
kubectl delete -f ./exporter/configMap -n oracle
kubectl delete -f ./exporter/standalone -n oracle


