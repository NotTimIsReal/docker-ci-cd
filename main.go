package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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
		removeContainersDead(client)
		switch r.Method {
		case "POST":
			var parsedBody acceptedBody
			log.Println("POST Request")
			val, err := parsedBody.getBody(r)
			if err != nil {
				log.Println(err)
			}
			if val.Repository.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Invalid Body or JSON Format")
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			bind, err := c.getBinds(parsedBody.Repository.Name)
			if err != nil {
				fmt.Fprintf(w, "Repository Flagged")
				log.Print("Repository Flagged")
			}
			fmt.Fprintf(w, "Docker Container Created With Name: %s", parsedBody.Repository.Name)
			log.Println(c)
			createAndStartContainer(client, alpine, bind, parsedBody.Repository.Name)
			r.Body.Close()

		default:
			log.Println("Unsupported Request")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method %s not allowed", r.Method)
		}
	})))
	log.Printf("Listening On Port %s", c.Port)
	go repeatEvery(5*time.Second, func() {
		removeContainersDead(client)
	})
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
	Port  string `yaml:"port"`
	Binds []struct {
		Name string `yaml:"name"`
		Bind string `yaml:"bind"`
	} `yaml:"binds"`
}

type acceptedBody struct {
	Repository Repo `json:"repository"`
}
type Repo struct {
	Name string `json:"name"`
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
	if c.Binds == nil {
		log.Fatal("Binds Not Found")
		panic("Config File is not valid please check the readme for valid configs")
	}
	return c
}
func getRightImage(alpine *string) {
	system := runtime.GOOS
	switch system {
	case "linux":
		*alpine = "ghcr.io/nottimisreal/alpinewithgit"
	case "darwin":
		*alpine = "ghcr.io/nottimisreal/alpinewithgit"
	case "windows":
		*alpine = "ghcr.io/nottimisreal/alpinewithgit"
	default:
		*alpine = "ghcr.io/nottimisreal/alpinewithgit"
	}

}
func (b *acceptedBody) getBody(r *http.Request) (*acceptedBody, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = json.Unmarshal(body, b)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return b, nil
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
func (c *conf) getBinds(containerName string) (string, error) {
	for _, bind := range c.Binds {
		if bind.Name == containerName {
			return bind.Bind, nil
		}
	}
	return "", errors.New("no such container")
}
func createAndStartContainer(client *client.Client, image string, location string, name string) (containerID string) {
	ctx := context.Background()
	resp, err := client.ContainerCreate(ctx, &container.Config{
		Image: image,
		//run git pull and exit
		Cmd: []string{"/bin/sh", "-c", "git pull", " && exit"},
	}, &container.HostConfig{
		Binds: []string{location},
	}, nil, nil, name)
	if err != nil {
		if err.Error() == "Error response from daemon: No such image: "+image {
			client.ImagePull(ctx, image, types.ImagePullOptions{})
			resp := createAndStartContainer(client, image, location, name)

			return resp

		} else {
			log.Fatal(err)
		}
	}
	go func() {
		err := client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}()
	return resp.ID
}
func removeContainersDead(c *client.Client) {
	ctx := context.Background()
	containers, err := c.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("status", "exited"),
		),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, container := range containers {
		c.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
	}

}
func repeatEvery(d time.Duration, f func()) {
	for {
		f()
		time.Sleep(d)
	}
}
