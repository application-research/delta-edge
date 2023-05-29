# Aggregation

Edge nodes accepts different sizes of files but it doesn't make deals for every files, what it does is it aggregates the files into a bigger file and make a deal for the aggregated file.

Currently, the aggregate size is 1GB per USER (API_KEY). This means that each user can upload files up to 1GB and the edge node will aggregate the files into a single file and make a deal.

## Pre-requisites
- make sure you have a edge node running either locally or remote. Use this guide [running a node](running_node.md) to run a node.
- identify the edge node host.
  - There is a running live edge node at https://edge.estuary.tech
- get a API key using this guide [getting an API key](getting-api-key.md)

## Upload a file
Files that are less than the aggregate size will automatically be part of a bucket. A bucket is a system object that collects all the files, bundle them all together to create a deal.

### How it works
![image](https://github.com/application-research/edge-ur/assets/4479171/17d0b7ad-f0b0-48bf-bd7c-16d16231b355)

### To upload a file:
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
Once the bucket is filled the edge node will aggregate the files into a single file and make a deal with the specified miner via Delta. Anyone can access the status of the CID using the status endpoint.

### How it works
![image](https://github.com/application-research/edge-ur/assets/4479171/b4c3f80d-8b7b-4b16-8c76-61020923a7d2)

### Checking the status by content ID
When the file is uploaded, edge-ur returns a propery called "ID". This is the content ID. You can use this ID to check the status of the content.

Note: The deal_id on the `DealInfo` field will return 0 initially. This is because the deal is not yet made. Once the deal is made, the user needs to hit the status endpoint again
to get the deal_id.

```bash
curl --location 'http://localhost:1313/open/status/content/1500'
{
   "data":{
      "content_info":{
         "cid":"bafybeib4druxhywgq6pw3o4yyfdf4xlcqtyuzzpjqkjztonlz2ui5p5yyq",
         "selective_car_cid":"",
         "name":"random_1685323144198576133.dat",
         "size":500000,
         "miner":"t017840"
      },
      "sub_piece_info":{
         "piece_cid":"baga6ea4seaqpqiuaevlybzqw32oey227npulipxeyz6xecm4owwqkfo2lx5fapy",
         "size":524288,
         "comm_pa":"baga6ea4seaqprqf7z6eb3syawm4nd22xiltiaoupsn3sse5ympva42kwiojd2ea",
         "size_pa":8388608,
         "comm_pc":"baga6ea4seaqpqiuaevlybzqw32oey227npulipxeyz6xecm4owwqkfo2lx5fapy",
         "size_pc":524288,
         "status":"StorageDealAwaitingPreCommit",
         "inclusion_proof":{
            "proofIndex":{
               "index":"0x000000000001ffc3",
               "path":[
                  "0x5f51015a6c5a60cb34ca2e67fdd19c3de1a3d234dc532de0e6e5e74642266027",
                  "0xe7f44f3b94879cf48d0ca14664c21b0a30e568f65f8059937d1643cd30184a28",
                  "0xdaf9ca3c2b71c078c0aca6d3268743c9f0895579243dc738c394a68013c9602d",
                  "0x57a2381a28652bf47f6bef7aca679be4aede5871ab5cf3eb2c08114488cb8526",
                  "0x1f7ac9595510e09ea41c460b176430bb322cd6fb412ec57cb17d989a4310372f",
                  "0xfc7e928296e516faade986b28f92d44a4f24b935485223376a799027bc18f833",
                  "0x08c47b38ee13bc43f41b915c0eed9911a26086b3ed62401bf9d58b8d19dff624",
                  "0xb2e47bfb11facd941f62af5c750f3ea5cc4df517d5c4f16db2b4d77baec1a32f",
                  "0xf9226160c8f927bfdcc418cdf203493146008eaefb7d02194d5e548189005108",
                  "0x2c1a964bb90b59ebfe0f6da29ad65ae3e417724a8f7c11745a40cac1e5e74011",
                  "0xfee378cef16404b199ede0b13e11b624ff9d784fbbed878d83297e795e024f02",
                  "0x8e9e2403fa884cf6237f60df25f83ee40dca9ed879eb6f6352d15084f5ad0d3f",
                  "0x752d9693fa167524395476e317a98580f00947afb7a30540d625a9291cc12a07",
                  "0x7022f60f7ef6adfa17117a52619e30cea82c68075adf1c667786ec506eef2d19",
                  "0xd99887b973573a96e11393645236c17b1f4c7034d723c7a99f709bb4da61162b",
                  "0xd0b530dbb0b4f25c5d2f2a28dfee808b53412a02931f18c499f5a254086b1326",
                  "0xc61a433ad9181ecd0494d8a3ad68342e65cfd91195bb4a35f52498872d58ad17"
               ]
            },
            "proofSubtree":{
               "index":"0x0000000000000003",
               "path":[
                  "0x055bda1951ac6a8dbb39528d4af0af42d95adcdb679e77af39b96d64dde52f1f",
                  "0xf9b01b48888fa61f24df8a954626ae680283f23d4cd8016dd8088a8cc6dab51e",
                  "0xa490ccea6aee523cb1d7bb5f9d2a944ca228cbd46c065e1cad2aacb5fc54c92f",
                  "0xda2369882f204474ac2d7fefc31406706f125bced0907210f270855d3d311216"
               ]
            }
         },
         "verifier_data":{
            "CommPc":{
               "/":"baga6ea4seaqpqiuaevlybzqw32oey227npulipxeyz6xecm4owwqkfo2lx5fapy"
            },
            "SizePc":524288
         }
      },
      "deal_info":{
         "deal_id":79457,
         "status":"StorageDealAwaitingPreCommit",
         "delta_node":"https://hackfs.delta.estuary.tech"
      }
   },
   "message":"success"
}
```

### Checking the status of the content by CID
You can check the status of the file using the following command:

Note that a CID can be dealt to different deals so this endpoint will return an array of results.
```bash
curl --location --request GET 'http://localhost:1313/open/status/content/cid/bafybeihl2yxou73d7mro4k3g25xnspjkp3afe7ffydvysypiq2yv5zh6y4' 
{
   "data":[
      {
         "content_info":{
            "cid":"bafybeihwrdaysfhiutv62gifebmx3hzghqqi7riljgz7vfuvlwabrokhui",
            "selective_car_cid":"",
            "name":"random_1685323335525204940.dat",
            "size":500000,
            "miner":"t017840"
         },
         "sub_piece_info":{
            "piece_cid":"baga6ea4seaqi2mhuqczxjj4j5db6udry5sc752offux7lp6pjnc2ocljesz6kiq",
            "size":524288,
            "comm_pa":"baga6ea4seaqdkiyu664i6pjn76wrc4yfit2jm4lyai377tixd4krcvtenswdekq",
            "size_pa":8388608,
            "comm_pc":"baga6ea4seaqi2mhuqczxjj4j5db6udry5sc752offux7lp6pjnc2ocljesz6kiq",
            "size_pc":524288,
            "status":"StorageDealAwaitingPreCommit",
            "inclusion_proof":{
               "proofIndex":{
                  "index":"0x000000000001ffc1",
                  "path":[
                     "0xbee18debaad248941d4f50e3ac1f17a73fa425c2d7d9948c7609b1ef473c253f",
                     "0x7ccef61c7c7f2f379ea99354ba667e63529890d5a74114477450d8b8d9238132",
                     "0x693b4e3429c834452a0e5bc7255255bd0b57ac5033fc6abb6c25b773e669251e",
                     "0x57a2381a28652bf47f6bef7aca679be4aede5871ab5cf3eb2c08114488cb8526",
                     "0x1f7ac9595510e09ea41c460b176430bb322cd6fb412ec57cb17d989a4310372f",
                     "0xfc7e928296e516faade986b28f92d44a4f24b935485223376a799027bc18f833",
                     "0x08c47b38ee13bc43f41b915c0eed9911a26086b3ed62401bf9d58b8d19dff624",
                     "0xb2e47bfb11facd941f62af5c750f3ea5cc4df517d5c4f16db2b4d77baec1a32f",
                     "0xf9226160c8f927bfdcc418cdf203493146008eaefb7d02194d5e548189005108",
                     "0x2c1a964bb90b59ebfe0f6da29ad65ae3e417724a8f7c11745a40cac1e5e74011",
                     "0xfee378cef16404b199ede0b13e11b624ff9d784fbbed878d83297e795e024f02",
                     "0x8e9e2403fa884cf6237f60df25f83ee40dca9ed879eb6f6352d15084f5ad0d3f",
                     "0x752d9693fa167524395476e317a98580f00947afb7a30540d625a9291cc12a07",
                     "0x7022f60f7ef6adfa17117a52619e30cea82c68075adf1c667786ec506eef2d19",
                     "0xd99887b973573a96e11393645236c17b1f4c7034d723c7a99f709bb4da61162b",
                     "0xd0b530dbb0b4f25c5d2f2a28dfee808b53412a02931f18c499f5a254086b1326",
                     "0x7eacee2ef23658551f68d1d577413a7cbaa85a5af436a4501914392c4d941e23"
                  ]
               },
               "proofSubtree":{
                  "index":"0x0000000000000001",
                  "path":[
                     "0x6f4d1b33f5949c03e0101448312da702124842a0ecd9c9a7430d6d64c5fe6810",
                     "0x33d1bf66e6612eb4c6c5f1adce4fd57d639b5447fa982841d3f9180333357c16",
                     "0xfc61255147302c4f89137e7ca69018ec08852218b49e797dd689088cdb793a00",
                     "0x41d687c129527729e1b27f27824284495f802dc3ed74fdac6d94690d9f05f70a"
                  ]
               }
            },
            "verifier_data":{
               "CommPc":{
                  "/":"baga6ea4seaqi2mhuqczxjj4j5db6udry5sc752offux7lp6pjnc2ocljesz6kiq"
               },
               "SizePc":524288
            }
         },
         "deal_info":{
            "deal_id":79483,
            "status":"StorageDealAwaitingPreCommit",
            "delta_node":"https://hackfs.delta.estuary.tech"
         }
      }
   ]
}
```

## View the file using the gateway url
```bash
http://localhost:1313/gw/<cid>
http://localhost:1313/gw/content/<content_id>
```
