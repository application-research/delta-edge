# Check the status of your content

In this guide, we will show you how to check the status of your content on the edge node.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
- get a API key using this guide [getting an API key](getting-api-key.md)
- content id from the upload file guide [uploading a file](upload_file.md)


## Checking the status of the file
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

