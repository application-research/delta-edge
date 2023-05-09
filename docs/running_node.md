# Running a node

Edge is a node that can be used to make storage deals on the Filecoin network. It is a daemon that runs in the background and can be controlled using the `edge` command.

Edge **REQUIRES** a running delta node to make deals. You can run a delta node locally or use a remote node.

## Install

### Clone the repo
``` 
git clone https://github.com/application-research/edge-ur
```

### build with `go build`
```
cd edge-ur
go build -tags netgo -ldflags '-s -w' -o edge
```

### build with `make`
```
make all
```

## Create the `.env` file
Copy the `.env.example` file to `.env` and update the values as needed.
```
# node information
NODE_NAME=edge
NODE_DESCRIPTION=Edge node
DB_DSN=edge-urdb

# Service URLs
AUTH_SVC_API=https://auth.estuary.tech
DELTA_NODE_API=https://node.delta.store
```

## Running
```
./edge daemon

Starting Edge daemon...
Setting up the Edge node... 
Setting up the Edge node... Done
Total memory: 233213760 bytes
Total memory: 132527064 bytes
Total system memory: 233213760 bytes
Total heap memory: 191365120 bytes
Heap in use: 168304640 bytes
Stack in use: 22544384 bytes
Total storage:  994662584320
Total number of CPUs: 10
Number of CPUs that this Delta will use: 10

 _______    ________   ________   _______                    ___  ___   ________     
|\  ___ \  |\   ___ \ |\   ____\ |\  ___ \                  |\  \|\  \ |\   __  \    
\ \   __/| \ \  \_|\ \\ \  \___| \ \   __/|    ____________ \ \  \\\  \\ \  \|\  \   
 \ \  \_|/__\ \  \ \\ \\ \  \  ___\ \  \_|/__ |\____________\\ \  \\\  \\ \   _  _\  
  \ \  \_|\ \\ \  \_\\ \\ \  \|\  \\ \  \_|\ \\|____________| \ \  \\\  \\ \  \\  \| 
   \ \_______\\ \_______\\ \_______\\ \_______\                \ \_______\\ \__\\ _\ 
    \|_______| \|_______| \|_______| \|_______|                 \|_______| \|__|\|__|

Starting API server
API server up and running on port 1313


   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.9.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:1313

```
