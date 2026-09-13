#!/bin/bash

bootstrap=$1
a=$2

./simulation -mode client -file $bootstrap -a $a 2> client.log
