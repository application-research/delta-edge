# Retrieve the file using the retrieval gateway

In this guide, we will show you how to retrieve the file from the edge node using the retrieval gateway.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
- the CID of the file you want to retrieve.

## Retrieving the file
Once you have a node and CID, you can retrieve the file from the node using the following command:
```bash
curl http://localhost:1313/gw/ipfs/bafybeigt7ba7nrauzln4gjffo2msoigcvsqje4jralw45gf7vvyq6xkrtq

OR 

curl http://localhost:1313/gw/bafybeigt7ba7nrauzln4gjffo2msoigcvsqje4jralw45gf7vvyq6xkrtq > file.zip
```

