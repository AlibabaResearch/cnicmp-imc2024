#!/bin/bash

client_num=10
con=0

for i in $(seq 10); do
	if [ $con -lt 5 ] 
	then
		con = $con+1
		echo $1
	fi	
done

