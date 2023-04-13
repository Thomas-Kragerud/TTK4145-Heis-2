Heis implementert med UDP, Peer2Peer og dynamic reasigner




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