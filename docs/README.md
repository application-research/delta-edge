# Edge UR (Upload and Retrieve) Service

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
- global replication factor can be set in the config file. This will make the node replicate the data to the specified number of nodes.
- For 32GB and above, the node will split the file into 32GB chunks and make deals for each chunk and car them. **[WIP]** 

# Author
Protocol Labs Outercore Engineering.
