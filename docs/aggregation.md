# Aggregation

Edge nodes accepts different sizes of files but it doesn't make deals for every files, what it does is it aggregates the files into a bigger file and make a deal for the aggregated file.

Currently, the aggregate size is 1GB per USER (API_KEY). This means that each user can upload files up to 1GB and the edge node will aggregate the files into a single file and make a deal.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
- get a API key using this guide [getting an API key](getting-api-key.md)

## Upload a file
Once you have a node and API key, you can upload a file to the node using the following command:
```bash
curl --location 'http://localhost:1313/api/v1/content/add' \
--header 'Authorization: Bearer [API_KEY]' \
--form 'data=@"/path/to/file"'
--form 'miners="f01963614"' // optional - add a list of miners to pin the file to
--form 'make_deal="false"' // optional - make a deal with the miners. Default is true.

{
    "status": "success",
    "message": "File uploaded and pinned successfully. Please take note of the ids.",
    "contents": [
        {
            "ID": 51,
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

## Checking the status of the bucket
Once the bucket is filled (i.e 1GB) the edge node will aggregate the files into a single file and make a deal with the specified miner via Delta.
You can check the status of the bucket using the following command:
```bash
curl --location 'http://localhost:1313/open/status/bucket/b1e3e0fe-f9ea-11ed-b6d3-d21437f11a21'
{"bucket":{"ID":11,"uuid":"b1e3e0fe-f9ea-11ed-b6d3-d21437f11a21","name":"b1e3e0fe-f9ea-11ed-b6d3-d21437f11a21","size":0,"delta_content_id":0,"delta_node_url":"https://node.delta.store","miner":"f01963614","piece_cid":"","piece_size":0,"inclusion_proof":"","cid":"","status":"open","last_message":"","created_at":"2023-05-24T04:34:01.463498Z","updated_at":"2023-05-24T04:34:01.463498Z"}}
````

**Bucket has the following state**
- Open - this means the bucket is still accepting files
- Processing - this means the bucket is now closed and is being processed to make a deal.
- Uploaded-to-Delta - this means the bucket has been uploaded to Delta and is waiting for the deal to be made.

## Checking the status of the content
You can check the status of the file using the following command:
```bash
curl --location --request GET 'http://localhost:1313/api/v1/status/1' \
--header 'Authorization: Bearer [API_KEY]'
{
    "content": {
        "ID": 1,
        "name": "aqua-plugin-231.8109.147.zip",
        "size": 1157548,
        "cid": "bafybeigt7ba7nrauzln4gjffo2msoigcvsqje4jralw45gf7vvyq6xkrtq",
        "delta_content_id": 2705,
        "status": "transfer-finished",
        "last_message": "transfer-finished",
        "miner": "f01794610",
        "created_at": "2023-04-05T09:08:11.839358-04:00",
        "updated_at": "2023-04-05T09:09:00.105453-04:00"
    }
}
```

## View the file using the gateway url
```
http://localhost:1313/gw/<cid>
http://localhost:1313/gw/content/<content_id>
```
