# Spammer script

In this guide, we will show you how to spam the edge node with 5GB files.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
- get a API key using this guide [getting an API key](getting-api-key.md)

## Set it up, and run it
Create a shell script with the following.s
```
#!/bin/bash

while :
do
  ms=$(date +%s%N)
  dd if=/dev/random of=random_"$ms".dat bs=10000 count=500000
	curl --location --request POST 'http://localhost:1313/api/v1/content/add' \
  --header 'Authorization: Bearer [REDACTED]' \
  --form 'data=@"./random_'${ms}'.dat"'
  rm random_$ms.dat
done
```

Set the proper permissions
```
chmod +xrw spammer.sh
```

Run the script
```
./spammer.sh
```

This will create a 5B file, upload it to the edge node, and delete it. It will repeat this process until you stop it.