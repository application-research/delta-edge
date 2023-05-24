# EdgeUR (Upload/Retrieve)

[![CodeQL](https://github.com/application-research/edge-ur/actions/workflows/codeql.yml/badge.svg)](https://github.com/application-research/edge-ur/actions/workflows/codeql.yml)

*Edge is currently under heavy development.*

## Goal/Purpose
Dedicated light node to upload and retrieve their CIDs. To do this, we decoupled the upload and retrieval aspect from the Estuary API node so we can create a node that can live on the "edge" closer to the customer.

By decoupling this to a light node, we achieve the following:
- dedicated node assignment for each customer. The customer or user can now launch an edge node and use it for both uploading to Estuary and retrieval using the same API keys issued from Estuary.
- switches the upload protocol. The user still needs to upload via HTTP but the edge node will transfer the file over to a delta node to make deals.

![image](https://user-images.githubusercontent.com/4479171/227985970-58bfead8-0906-4f2e-b7ae-b314508ee3e5.png)

## Features
- Only supports online/e2e verified deals for now.
- Accepts concurrent uploads (small to large)
- Stores the CID and content on the local blockstore using whypfs
- Save the data on local sqlite DB
- retries the storage deals if it fails. Uses delta `auto_retry` feature.
- periodically checks the status of the deals and update the database.
- For 32GB and above, the node will split the file into 32GB chunks and make deals for each chunk and car them. **[WIP]**

## Quick how to

### Upload a file with a given miner
We have a working live server that you can use. To run basic upload, please get API key first. You can get one here https://auth.estuary.tech/register-new-token

```
curl --location 'http://localhost:3000/api/v1/content/add' \
--header 'Authorization: Bearer [APIKEY]' \
--form 'data=@"/file.zip"' \
--form 'miners="f0137168,f0717969"' // list of miners (optional)
{
    "status": "success",
    "message": "File uploaded and pinned successfully. Please take note of the ids.",
    "contents": [
        {
            "ID": 53,
            "name": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y.zip",
            "size": 1157548,
            "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y",
            "delta_content_id": 0,
            "delta_node_url": "https://node.delta.store",
            "status": "pinned",
            "last_message": "",
            "miner": "f0137168",
            "replication": 0,
            "created_at": "2023-04-22T14:40:43.10918+02:00",
            "updated_at": "2023-04-22T14:40:43.10918+02:00"
        },
        {
            "ID": 54,
            "name": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y.zip",
            "size": 1157548,
            "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y",
            "delta_content_id": 0,
            "delta_node_url": "https://node.delta.store",
            "status": "pinned",
            "last_message": "",
            "miner": "f0717969",
            "replication": 0,
            "created_at": "2023-04-22T14:40:43.112161+02:00",
            "updated_at": "2023-04-22T14:40:43.112161+02:00"
        }
    ]
}
```

### Check status
Get the content id and use the following call
```
curl --location 'http://localhost:3000/api/v1/status/53' \
--header 'Authorization: Bearer [APIKEY]'
{
    "content": {
        "ID": 53,
        "name": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y.zip",
        "size": 1157548,
        "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y",
        "delta_content_id": 67547,
        "delta_node_url": "https://node.delta.store",
        "status": "uploaded-to-delta",
        "last_message": "",
        "miner": "f0137168",
        "replication": 0,
        "created_at": "2023-04-22T14:40:43.10918+02:00",
        "updated_at": "2023-04-22T14:40:46.564971+02:00"
    }
}
```
### View the file using the gateway url
```
http://localhost:1313/gw/<cid>
http://localhost:1313/gw/content/<content_id>
```

## Getting Started
To get started, follow the guide [here](docs/README.md).

# Author
Protocol Labs Outercore Engineering
