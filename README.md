
## Elevator System
The elevator system is a distributed system that uses a peer-to-peer (P2P) network architecture to communicate between elevator nodes. Specifically, it utilizes the User Datagram Protocol (UDP) for broadcasting messages across the network. UDP is a connectionless protocol that allows for fast transmission of data packets, making it a suitable choice for applications that require low latency and real-time communication.

In this system, each elevator node is responsible for managing its own state and making decisions based on that state. To ensure consistency across the system, each elevator node broadcasts its state information to all other nodes using UDP broadcasts. This approach allows all nodes to have the same information about the state of the entire system, which enables them to make informed decisions.

The use of a P2P network architecture allows for a decentralized system that is fault-tolerant and can scale to accommodate additional nodes. It also reduces the risk of a single point of failure, which can be a significant advantage in large distributed systems.

To mitigate version control issues, the system only sends information about itself, rather than the entire system. This means that each node only needs to be aware of the state of the other nodes, rather than the entire system. This approach reduces the amount of information that needs to be transmitted and helps to ensure consistency across the system.



## Command Line Arguments

This application accepts the following command-line arguments:

1. `--port`: Specifies the port number for this elevator. (mandatory)
2. `--id`: Sets the ID of this elevator. (mandatory)
3. `--floors`: Determines the number of floors in the elevator system. (default: 4)
4. `--gui`: Enables the graphical user interface (GUI) when set to true. (default: false)
5. `--sound`: Activates the sound when set to true. (default: false)

### Usage

You can provide command-line arguments when starting the application like this:
```
go run main.go --port="1234" --id="1" --floors=6 --gui=ture --sound=true
```


## Installing External Dependencies

This project uses several external packages that need to be installed before running the application. To download and install the required packages, use the following commands:

```
go get -u github.com/faiface/beep
go get -u github.com/faiface/beep/mp3
go get -u github.com/faiface/beep/speaker
go get -u github.com/faiface/pixel
go get -u github.com/faiface/pixel/imdraw
go get -u github.com/faiface/pixel/pixelgl
golang.org/x/image/colornames
go get github.com/oleiade/lane

```
The above commands will download and install the necessary packages, allowing you to run the application without issues.


## Running on macOS using Docker

Since the hall request assigner executable is not compatible with macOS, you can use Docker to run the application on a macOS system. To set up Docker for this purpose, follow the steps below:

1. Install Docker Desktop for Mac from the official website: https://www.docker.com/products/docker-desktop

2. Open a terminal and navigate to the project directory.

3. Make sure the `Dockerfile` is in the same folder as the `hall_request_executables`. 

4. Build the Docker image by running the following command in the terminal:
```
docker build -t dock_hra
```
This command will create a Docker image named `dock_hra` using the provided `Dockerfile`.

5. Now you can run the application. The project is configured to use the Docker container for executing the hall request assigner when running on macOS.

Please note that you may need to give Docker permission to access your project directory. You can do this by going to Docker Desktop Preferences > Resources > File Sharing and adding your project directory to the list.

Be aware that running the application on macOS might result in slightly slower performance due to the application uses a Docker container running an Ubuntu image as a compatibility layer.

This additional layer of virtualization can introduce some overhead, which may affect the overall performance of the application on macOS. However, the actual impact on performance would depend on various factors, such as your hardware and the resources allocated to the Docker container.

While the performance difference might be negligible for many use cases, it is important to keep this consideration in mind when running the application on macOS.