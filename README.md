# Uploader Job to Estuary

## Goal/Purpose
To allow the customer to have a better UX using estuary, we need to give a dedicated light node for them to upload and retrieve their CIDs. To do this, we decoupled the upload and retrieval aspect from the Estuary API node so we can create a node that can live on the "edge" closer to the customer.

By decoupling this to a light node, we achieve the following:
- dedicated node assignment for each customer. The customer or user can now launch an edge node and use it for both uploading to Estuary and retrieval using the same API keys issued from Estuary.
- offload the Estuary API node and get it to focus on deal-making process rather than consuming massive, concurrent HTTP uploads
- switches the upload protocol. The user still needs to upload via HTTP but the edge node will use bitswap to transfer the files over to Estuary.
![image](https://user-images.githubusercontent.com/4479171/211378054-ab24e2b6-6273-45fd-ad24-a98dbeb14fbe.png)


## Features
- Accepts concurrent uploads (small to large)
- Stores the CID and content on the local blockstore using whypfs
- Save the data on local sqlite DB
- Process each files and call estuary add-ipfs endpoint to make deals for the CID
- uses estuary api (`pinning/pins`) endpoint to pin files on estuary

## HL Architecture/Process flow
![image](https://user-images.githubusercontent.com/4479171/211354164-2df9b2be-ff77-4749-871b-3a5932e0b857.png)

# Build
## `go build`
```
go build -tags netgo -ldflags '-s -w' -o edge-ur
```

# Running 
## Create the `.env` file
```
DB_NAME=edge-ur
MODE=remote-pin
REMOTE_PIN_ENDPOINT=https://api.estuary.tech/pinning/pins

## For local CLI only
API_KEY=[REDACTED]

# JOB Frequency
BUCKET_ASSIGN=10
UPLOAD_PROCESS=30
DEAL_CHECK=86400

# Deal config
DELETE_AFTER_DEAL_MADE=false
```

## Running the daemon
```
./edge-ur daemon
```


## Running the CLI
While running the daemon, the user can run the following commands to add file or dir to local instance
```
./edge-ur pin-file <path>
./edge-ur pin-dir <path>
```

This will create an entry on the contents table, assigned to a bucket which will then be pushed to Estuary and the delegates.

# Gateway
This node comes with it's own gateway to serve directories and files.

View the gateway using:
- https://localhost:1313
- https://localhost:1313/dashboard
- https://localhost:1313/gw/ipfs/:cid

# Pin and make a storage deal for your file(s) on Estuary
```
curl --location --request POST 'http://localhost:1313/api/v1/content/add' \
--header 'Authorization: Bearer [ESTUARY_API_KEY]' \
--form 'data=@"/path/to/file"'
```

# Pin and make a storage deal for your cid(s) on Estuary
```
curl --location --request POST 'http://localhost:1313/api/v1/content/cid/bafybeihxodfkobqiovfgui6ipealoabr2u3bhor765z47wxdthrgn7rvyq' \
--header 'Authorization: Bearer [ESTUARY_API_KEY]'
```

## Status check
This will return the status of the file(s) or cid(s) on edge-ur. It'll also return the estuary content_id.
```
curl --location --request GET 'http://localhost:1313/api/v1/status/5' \
--header 'Authorization: Bearer [ESTUARY_API_KEY]'
```
