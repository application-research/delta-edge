# Uploader Job to Estuary

- Accepts concurrent uploads (small to large)
- Stores the CID and content on the local blockstore using whypfs
- Save the data on local sqlite DB
- Process each files and call estuary add-ipfs endpoint to make deals for the CID

![image](https://user-images.githubusercontent.com/4479171/211347885-8f8f2218-e7b6-4b5a-bdc1-0a6e0faa4b73.png)


# Build
## `go build`
```
go build -tags netgo -ldflags '-s -w' -o edge-ur
```


# Running the daemon
Running the daemon will initialize the node configuration and the gateway at port 1313
```
./edge-ur daemon
```

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

# Pin make a storage deal for your cid(s) on Estuary
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
