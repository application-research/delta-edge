# Uploader Job to Estuary

- Accepts concurrent uploads (small to large)
- Stores the CID and content on the local blockstore using whypfs
- Save the data on local sqlite DB
- Process each files and call estuary add-ipfs endpoint to make deals for the CID


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

# Pin your files to Estuary
```
curl --location --request POST 'http://localhost:1313/api/v1/content/add' \
--header 'Authorization: Bearer [YOUR API KEY]' \
--form 'data=@"/path/to/file"'
```