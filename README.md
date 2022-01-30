# Docker-CI-CD

## What Is This?
The Docker-CI-CD is a tool that helps you to make every commit and push to your github repositories happen seamlessly and automatically. It works by making a new container to you existing bind of the main container and pulling the latest github commit with linux alpine and then restarting your main container and removing the container that pulled from the github commit. This ensures that your main container is always up to date with the latest changes while not having to manually restart the container.

## Installation
In order to install this tool you need to download the latest version of the compiled binary from [Github](https://github.com/NotTimIsReal/docker-ci-cd), you also need to install the latest version of docker.

### Darwin:
  For Darwin (MacOS) you can just move the binary to your /usr/local/bin directory.
### Linux:
  For Linux you can just move the binary to your /usr/bin directory.
### Windows:
  For Windows you can just move the binary to your /c/Windows/System32 directory.

## Usage
### Configuation:
You are required to generate your own config.yaml file, the format should look like so:
```yaml
port: :8080
binds: 
  - name: ybabackendv2
    bind: "path-to-git-repo:/home/root"
```
The `port` is the port that the server will listen on. To add more binds you should add another `binds` entry. The `name` in the `Binds` entry is the name of repository that the post request is from. The `bind` is the path that the server will bind to. The yaml file should be located in the project directory.

### Github:
The Github webhook is used to trigger the build. The Github webhook should be configured to post to the url of `https://your-server-ip:THEPORTYOUSPECIFED/`. 

## Running
To run the web-server you can run the following command:
```bash 
ci-cd
```
This should start the webserver and start logging to a file called logs.txt

## Daemonizing
To daemonize the web-server you can make your own service file to run the server.

