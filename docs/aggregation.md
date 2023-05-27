# Aggregation

Edge nodes accepts different sizes of files but it doesn't make deals for every files, what it does is it aggregates the files into a bigger file and make a deal for the aggregated file.

Currently, the aggregate size is 1GB per USER (API_KEY). This means that each user can upload files up to 1GB and the edge node will aggregate the files into a single file and make a deal.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
  - There is a running live edge node at https://edge.estuary.tech
- get a API key using this guide [getting an API key](getting-api-key.md)

## Upload a file

![image](https://github.com/application-research/edge-ur/assets/4479171/17d0b7ad-f0b0-48bf-bd7c-16d16231b355)


Once you have a node and API key, you can upload a file to the node using the following command:
```bash
curl --location 'http://localhost:1313/api/v1/content/add' \
--header 'Authorization: Bearer [API_KEY]' \
--form 'data=@"/path/to/file"'
--form 'miners="t017840"' // optional - add a list of miners to pin the file to

{
    "status": "success",
    "message": "File uploaded and pinned successfully. Please take note of the ids.",
    "contents": [
        {
            "ID": 140,
            "name": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y.zip",
            "size": 1157548,
            "cid": "bafybeicgdjdvwes3e5aaicqljrlv6hpdfsducknrjvsq66d4gsvepolk6y",
            "delta_content_id": 0,
            "delta_node_url": "https://node.delta.store",
            "bucket_uuid": "b1e3e0fe-f9ea-11ed-b6d3-d21437f11a21",
            "status": "pinned",
            "last_message": "",
            "miner": "f01963614",
            "replication": 0,
            "created_at": "2023-04-22T13:10:34.177515+02:00",
            "updated_at": "2023-04-22T13:10:34.177515+02:00"
        },
    ]
}
```
*Note that the content has been assigned to a bucket (using bucket_uuid). The bucket is a system object that collects the files which in turn
is used to aggregate the files into a single file.*

## Checking the status of the uploaded content
Once the bucket is filled the edge node will aggregate the files into a single file and make a deal with the specified miner via Delta.

![image](https://github.com/application-research/edge-ur/assets/4479171/b4c3f80d-8b7b-4b16-8c76-61020923a7d2)

## Checking the status by content ID
When the file is uploaded, edge-ur returns a propery called "ID". This is the content ID. You can use this ID to check the status of the content.

Note: The deal_id on the `DealInfo` field will return 0 initially. This is because the deal is not yet made. Once the deal is made, the user needs to hit the status endpoint again
to get the deal_id.

```bash
curl --location 'http://localhost:1313/open/status/content/140'
{
    "data": {
        "content_info": {
            "cid": "bafybeifixryrosiyqtsjl3bnrnvkmfku2eqoumj5lay5jejf3ixvedmwju",
            "name": "random_1685214395N.dat",
            "size": 500000,
            "miner": "t017840"
        },
        "sub_piece_info": {
            "piece_cid": "baga6ea4seaqdgoccoy2x2tqgly4rhhbnewt554xje5jcokdtqhvarpvb5afsehi",
            "size": 524288,
            "comm_pa": "baga6ea4seaqlqz2o7vakjlioebzplj5hzwaukvwtsc6gzso4ifijlmdndvp72lq",
            "size_pa": 8388608,
            "comm_pc": "baga6ea4seaqdgoccoy2x2tqgly4rhhbnewt554xje5jcokdtqhvarpvb5afsehi",
            "size_pc": 524288,
            "status": "StorageDealVerifyData",
            "inclusion_proof": "goIEhFggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCDJTuaLtV+MWLaXHItK/Fug8VPxpMWoX+F4PvMvZ8TrMlgg8sEir+Lzmg9NOgb2LvHdbvOolhyA6T9GCicmWXlp3QSCGgAB/8SRWCD1pf1C0WogMCeY727TCZebQwA9IyDZ8OjqmDGpJ1n7C1ggNzG7maxon2bu9Zc+SpTaGI9N3K5YByT8bz/WDf1IgzNYIFE1nSxP0H7VLSnzTOZPh3RlvdsJYXJbU38xhJcppQQ1WCBXojgaKGUr9H9r73rKZ5vkrt5Ycatc8+ssCBFEiMuFJlggH3rJWVUQ4J6kHEYLF2QwuzIs1vtBLsV8sX2YmkMQNy9YIPx+koKW5Rb6remGso+S1EpPJLk1SFIjN2p5kCe8GPgzWCAIxHs47hO8Q/QbkVwO7ZkRomCGs+1iQBv51YuNGd/2JFggsuR7+xH6zZQfYq9cdQ8+pcxN9RfVxPFtsrTXe67Boy9YIPkiYWDI+Se/3MQYzfIDSTFGAI6u+30CGU1eVIGJAFEIWCAsGpZLuQtZ6/4PbaKa1lrj5BdySo98EXRaQMrB5edAEVgg/uN4zvFkBLGZ7eCxPhG2JP+deE+77YeNgyl+eV4CTwJYII6eJAP6iEz2I39g3yX4PuQNyp7YeetvY1LRUIT1rQ0/WCB1LZaT+hZ1JDlUduMXqYWA8AlHr7ejBUDWJakpHMEqB1ggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCDQtTDbsLTyXF0vKijf7oCLU0EqApMfGMSZ9aJUCGsTJlggC9byZKQHRfNraPkZqWRdUxZ4w2wnvw3iD8oXWPJIdgE=",
            "verifier_data": "gtgqWCgAAYHiA5IgIDM4QnY1fU4GXjkTnC0lp97y6SdSJyhzgeoIvqHoCyIdGgAIAAA="
        },
        "deal_info": {
            "deal_uuid": "",
            "deal_id": 0,
            "status": "StorageDealVerifyData",
            "delta_node": "http://localhost:1414"
        }
    },
    "message": "success"
}
```

### Checking the status of the content by CID
You can check the status of the file using the following command:

Note that a CID can be dealt to different deals so this endpoint will return an array of results.
```bash
curl --location --request GET 'http://localhost:1313/open/status/content/cid/bafybeihl2yxou73d7mro4k3g25xnspjkp3afe7ffydvysypiq2yv5zh6y4' \
--header 'Authorization: Bearer [API_KEY]'
{
    "data": [
        {
            "content_info": {
                "cid": "bafybeihl2yxou73d7mro4k3g25xnspjkp3afe7ffydvysypiq2yv5zh6y4",
                "name": "random_1685214394N.dat",
                "size": 500000
            },
            "sub_piece_info": {
                "piece_cid": "baga6ea4seaqk2rif6gz4mqj4q6pke2bunhi7qh7uwmekn27lhbvhcsbq7wr3aji",
                "size": 524288,
                "comm_pa": "baga6ea4seaqdjgko77z2jmiykg44bzlsyvpsv53b36apt6wwgmdqm4rvw3v6mlq",
                "size_pa": 8388608,
                "comm_pc": "baga6ea4seaqk2rif6gz4mqj4q6pke2bunhi7qh7uwmekn27lhbvhcsbq7wr3aji",
                "size_pc": 524288,
                "status": "StorageDealVerifyData",
                "inclusion_proof": "goIEhFggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCB2qV1fAFCCA+/Zf68f13lBn9mCicwEumC+YrG/Kqx6GlggNUKAEyN5sdKwSVeygJKA3MpZHf+y87iCkETV7s1pLBeCGgAB/8SRWCD1pf1C0WogMCeY727TCZebQwA9IyDZ8OjqmDGpJ1n7C1ggNzG7maxon2bu9Zc+SpTaGI9N3K5YByT8bz/WDf1IgzNYICM1HodgKdT6HNEKwtoNyq4BQ8IeOSJjY4tJC/1eTokLWCBXojgaKGUr9H9r73rKZ5vkrt5Ycatc8+ssCBFEiMuFJlggH3rJWVUQ4J6kHEYLF2QwuzIs1vtBLsV8sX2YmkMQNy9YIPx+koKW5Rb6remGso+S1EpPJLk1SFIjN2p5kCe8GPgzWCAIxHs47hO8Q/QbkVwO7ZkRomCGs+1iQBv51YuNGd/2JFggsuR7+xH6zZQfYq9cdQ8+pcxN9RfVxPFtsrTXe67Boy9YIPkiYWDI+Se/3MQYzfIDSTFGAI6u+30CGU1eVIGJAFEIWCAsGpZLuQtZ6/4PbaKa1lrj5BdySo98EXRaQMrB5edAEVgg/uN4zvFkBLGZ7eCxPhG2JP+deE+77YeNgyl+eV4CTwJYII6eJAP6iEz2I39g3yX4PuQNyp7YeetvY1LRUIT1rQ0/WCB1LZaT+hZ1JDlUduMXqYWA8AlHr7ejBUDWJakpHMEqB1ggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCDQtTDbsLTyXF0vKijf7oCLU0EqApMfGMSZ9aJUCGsTJlggFHI0eeMxGZvmcTnvAh/4L/BYOohDGLYIv9y7O6wfzRE=",
                "verifier_data": "gtgqWCgAAYHiA5IgIK1FBfGzxkE8h56iaDRp0fgf9LMIpuvrOGpxSDD9o7AlGgAIAAA="
            },
            "deal_info": {
                "deal_id": 0,
                "status": "StorageDealVerifyData",
                "delta_node": "http://localhost:1414"
            }
        }
    ]
}
```

### Checking the status of the content by PieceCID
You can check the status of the file using the following command:

```bash
curl --location --request GET 'http://localhost:1313/open/status/content/piece/baga6ea4seaqk2rif6gz4mqj4q6pke2bunhi7qh7uwmekn27lhbvhcsbq7wr3aji' \
--header 'Authorization: Bearer [API_KEY]'
{
    "data": {
        "content_info": {
            "cid": "bafybeihl2yxou73d7mro4k3g25xnspjkp3afe7ffydvysypiq2yv5zh6y4",
            "name": "random_1685214394N.dat",
            "size": 500000,
            "miner": "t017840"
        },
        "sub_piece_info": {
            "piece_cid": "baga6ea4seaqk2rif6gz4mqj4q6pke2bunhi7qh7uwmekn27lhbvhcsbq7wr3aji",
            "size": 524288,
            "comm_pa": "baga6ea4seaqdjgko77z2jmiykg44bzlsyvpsv53b36apt6wwgmdqm4rvw3v6mlq",
            "size_pa": 8388608,
            "comm_pc": "baga6ea4seaqk2rif6gz4mqj4q6pke2bunhi7qh7uwmekn27lhbvhcsbq7wr3aji",
            "size_pc": 524288,
            "status": "StorageDealVerifyData",
            "inclusion_proof": "goIEhFggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCB2qV1fAFCCA+/Zf68f13lBn9mCicwEumC+YrG/Kqx6GlggNUKAEyN5sdKwSVeygJKA3MpZHf+y87iCkETV7s1pLBeCGgAB/8SRWCD1pf1C0WogMCeY727TCZebQwA9IyDZ8OjqmDGpJ1n7C1ggNzG7maxon2bu9Zc+SpTaGI9N3K5YByT8bz/WDf1IgzNYICM1HodgKdT6HNEKwtoNyq4BQ8IeOSJjY4tJC/1eTokLWCBXojgaKGUr9H9r73rKZ5vkrt5Ycatc8+ssCBFEiMuFJlggH3rJWVUQ4J6kHEYLF2QwuzIs1vtBLsV8sX2YmkMQNy9YIPx+koKW5Rb6remGso+S1EpPJLk1SFIjN2p5kCe8GPgzWCAIxHs47hO8Q/QbkVwO7ZkRomCGs+1iQBv51YuNGd/2JFggsuR7+xH6zZQfYq9cdQ8+pcxN9RfVxPFtsrTXe67Boy9YIPkiYWDI+Se/3MQYzfIDSTFGAI6u+30CGU1eVIGJAFEIWCAsGpZLuQtZ6/4PbaKa1lrj5BdySo98EXRaQMrB5edAEVgg/uN4zvFkBLGZ7eCxPhG2JP+deE+77YeNgyl+eV4CTwJYII6eJAP6iEz2I39g3yX4PuQNyp7YeetvY1LRUIT1rQ0/WCB1LZaT+hZ1JDlUduMXqYWA8AlHr7ejBUDWJakpHMEqB1ggcCL2D372rfoXEXpSYZ4wzqgsaAda3xxmd4bsUG7vLRlYINmYh7lzVzqW4ROTZFI2wXsfTHA01yPHqZ9wm7TaYRYrWCDQtTDbsLTyXF0vKijf7oCLU0EqApMfGMSZ9aJUCGsTJlggFHI0eeMxGZvmcTnvAh/4L/BYOohDGLYIv9y7O6wfzRE=",
            "verifier_data": "gtgqWCgAAYHiA5IgIK1FBfGzxkE8h56iaDRp0fgf9LMIpuvrOGpxSDD9o7AlGgAIAAA="
        },
        "deal_info": {
            "deal_uuid": "",
            "deal_id": 0,
            "status": "StorageDealVerifyData",
            "delta_node": "http://localhost:1414"
        }
    },
    "message": "success"
}
```

## View the file using the gateway url
```bash
http://localhost:1313/gw/<cid>
http://localhost:1313/gw/content/<content_id>
```
