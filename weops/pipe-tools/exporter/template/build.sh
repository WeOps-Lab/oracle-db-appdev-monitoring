#!/bin/bash

# 单点
for version in 11g 12c 18c 19c 21c; do

  standalone_output_file="standalone_${version}.yaml"
  sed "s/{{VERSION}}/${version}/g;" standalone.tpl > ../standalone/${standalone_output_file}
done

# RAC
for version in 19c; do
  for rac in rac1 rac2; do
    rac_output_file="rac_${version}_${rac}.yaml"
    sed "s/{{VERSION}}/${version}/g; s/{{RAC}}/${rac}/g" rac.tpl > ../rac/${rac_output_file}
  done
done

# dataGuard
for version in 19c; do
    dg_output_file="dg_${version}.yaml"
    sed "s/{{VERSION}}/${version}/g;" dataGuard.tpl > ../dataGuard/${dg_output_file}
done
