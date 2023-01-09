# Uploader Job to Estuary

- Accepts concurrent uploads (small to large)
- Stores the CID and content on the local blockstore using whypfs
- Save the data on local sqlite DB
- Process each files and call estuary add-ipfs endpoint to make deals for the CID
