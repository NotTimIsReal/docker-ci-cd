package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"

	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
)

func main() {
	var alpine string
	getRightImage(&alpine)
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}
	var c conf
	c.getConf()
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("Starting Web Server")
	mux := http.NewServeMux()
	log.Println("Initalised MiddleWare")
	mux.Handle("/", serverHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var parsedBody acceptedBody
			log.Println("POST Request")
			body, raw, err := getBody(r)
			if err != nil {
				log.Println(err)
			}
			if body == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Body Not Found")
				return
			}
			json.Unmarshal(raw, &parsedBody)
			if parsedBody.Repository.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Invalid Body or JSON Format")
				return
			}
			go createNewContainer(client, alpine, "/tmp:/etc", parsedBody.Repository.Name)
			namedContainer, err := getContainerByName(client, parsedBody.Repository.Name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(w, "Docker Container Created With Name: %s", parsedBody.Repository.Name)
			err = client.ContainerStart(context.Background(), namedContainer.ID, types.ContainerStartOptions{})
			if err != nil {
				log.Fatal(err)
			}
		default:
			log.Println("Unsupported Request")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method %s not allowed", r.Method)
		}
	})))
	log.Printf("Listening On Port %s", c.Port)
	log.Fatal(http.ListenAndServe(c.Port, mux))

}
func serverHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Docker-CI-CD")
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Accept", "application/json")
		next.ServeHTTP(w, r)
	})
}

type conf struct {
	Port   string `yaml:"port"`
	Folder struct {
		Binds []struct {
			Name string `yaml:"name"`
			bind string `yaml:"bind"`
		}
	}
}
type acceptedBody struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"respository"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	if c.Port == "" {
		c.Port = ":3000"
	}
	if c.Folder.Binds == nil {
		log.Fatal("Binds Not Found")
		panic("Config File is not valid please check the readme for valid configs")
	}
	return c
}
func createNewContainer(client *client.Client, image string, location string, name string) (containerID string, err error) {
	ctx := context.Background()
	resp, err := client.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"/bin/sh", "-c", "apk add --no-cache git && git pull"},
	}, &container.HostConfig{
		Binds: []string{location},
	}, nil, nil, name)
	if err != nil {
		if err.Error() == "Error response from daemon: No such image: "+image {
			client.ImagePull(ctx, image, types.ImagePullOptions{})
			resp, err := createNewContainer(client, image, location, name)
			if err != nil {
				log.Fatal(err)
			}
			return resp, nil

		}
		return "", err
	}

	return resp.ID, nil

}
func getRightImage(alpine *string) {
	system := runtime.GOOS
	switch system {
	case "linux":
		*alpine = "alpine:latest"
	case "darwin":
		*alpine = "alpine:latest"
	case "windows":
		*alpine = "alpine:latest"
	default:
		*alpine = "alpine:latest"
	}

}
func getBody(r *http.Request) (string, []byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", []byte(""), err
	}
	return string(body), body, nil
}
func getContainerByName(client *client.Client, name string) (container types.Container, err error) {
	ctx := context.Background()
	containers, err := client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return container, err
	}
	for _, c := range containers {
		if c.Names[0] == "/"+name {
			return c, nil
		}
	}
	return container, nil
}
