# Retrieve the file using the retrieval gateway

In this guide, we will show you how to retrieve the file from the edge node using the retrieval gateway.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
- the CID of the file you want to retrieve.


# Gateway home
If you know the CID, you can check the gateway home page by opening the base url on your browser
http://localhost:1313/

![image](https://user-images.githubusercontent.com/4479171/230234478-80f27572-6615-4dde-8507-39701acdd9ee.png)

## Retrieving the file
Once you have a node and CID, you can retrieve the file from the node using the following command:
```bash
curl http://localhost:1313/gw/ipfs/bafybeigt7ba7nrauzln4gjffo2msoigcvsqje4jralw45gf7vvyq6xkrtq

OR 

curl http://localhost:1313/gw/bafybeigt7ba7nrauzln4gjffo2msoigcvsqje4jralw45gf7vvyq6xkrtq > file.zip
```
