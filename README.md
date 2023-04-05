# Edge Node

**Edge is currently under heavy development and a massive optimization release is coming soon**

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
- periodically checks the status of the deals and update the database.
- For 32GB and above, the node will split the file into 32GB chunks and make deals for each chunk and car them.

# Build
## `go build`
```
go build -tags netgo -ldflags '-s -w' -o edge-cli
```

# Running
## Create the `.env` file
```
DB_NAME=edge-urdb
DELTA_NODE_API=https://cake.delta.store
EDGE_NODE_API_KEY=EST4c96295b-27bc-4358-94c6-f547b61246c1ARY
DEAL_CHECK=600
```

## Running the daemon
```
./edge-cli daemon --repo=/tmp/blockstore
```


# Gateway
This node comes with it's own gateway to serve directories and files.

View the gateway using:
- https://localhost:1313/gw/:cid
- https://localhost:1313/gw/ipfs/:cid

# Pin and make a storage deal for your file(s) on Estuary
```
curl --location --request POST 'http://localhost:1313/api/v1/content/add' \
--header 'Authorization: Bearer [ESTUARY_API_KEY]' \
--form 'data=@"/path/to/file"'
{
    "status": "success",
    "message": "File uploaded and pinned successfully. Please take note of the ID.",
    "id": 5,
    "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y"
}
```

# Status check
This will return the status of the file(s) or cid(s) on edge-ur. It'll also return the estuary content_id.
```
curl --location --request GET 'http://localhost:1313/api/v1/status/1' \
--header 'Authorization: Bearer [ESTUARY_API_KEY]'
{
    "content": {
        "ID": 1,
        "name": "aqua-plugin-231.8109.147.zip",
        "size": 1157548,
        "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y",
        "delta_content_id": 2705,
        "status": "transfer-finished",
        "last_message": "transfer-finished",
        "miner": "",
        "created_at": "2023-04-05T09:08:11.839358-04:00",
        "updated_at": "2023-04-05T09:09:00.105453-04:00"
    }
}
```

# Author
Protocol Labs Outercore Engineering
