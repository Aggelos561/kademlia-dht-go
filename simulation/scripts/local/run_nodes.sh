#!/bin/bash

a=$1

python3 run_nodes.py --exec ../../simulation --ip 127.0.0.1 --boot-dir ./bootstraps/nodes_bootstraps/ --out-dir ./logs --a $a