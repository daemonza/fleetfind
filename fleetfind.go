// fleetfind quickly finds a docker process on a fleet cluster.
// This is useful for when the unit is removed from fleet but for some
// reason the docker process still runs.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
)

func fleetSSH(host string, command string) []string {
	sshCmd := "fleetctl ssh " + host + " \"docker ps | grep " + command + "\""
	sshOut, _ := exec.Command("sh", "-c", sshCmd).Output()
	if string(sshOut) == "" {
		return nil
	}

	// remove spaces
	var results []string
	results = append(results, "host:"+host)
	dirty := strings.Split(string(sshOut), " ")
	for _, a := range dirty {
		if a != "" {
			results = append(results, a)
		}
	}

	return results
}

func find(containerName string, action string) {
	cmd := "fleetctl list-machines | awk '/-/ {print $1}' | cut -d. -f1"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		fmt.Println("ERROR : " + err.Error())
		os.Exit(1)
	}

	// put output into a []string
	hostList := strings.Fields(strings.Replace(string(out), "\n", " ", -1))

	// setup wait group on nr hosts found
	var wg sync.WaitGroup
	wg.Add(len(hostList))

	resultsChannel := make(chan []string)

	for host := range hostList {
		fleetMachine := hostList[host]
		go func(fleetMachine string) {
			defer wg.Done()
			unitDetails := fleetSSH(fleetMachine, containerName)
			if action == "stop" && unitDetails != nil {
				fmt.Println("stopping")
			}
			resultsChannel <- unitDetails
		}(fleetMachine)
	}

	go func() {
		for unitDetails := range resultsChannel {
			if unitDetails != nil {
				fmt.Println(unitDetails)
			}
		}
	}()

	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "fleetfind"
	app.Version = "0.0.1"
	// app.Author = "Werner Gillmer"
	// app.Email = "werner.gillmer@gmail.com"
	// app.Authors = []Author{Author{Name: "Werner Gillmer", Email: "werner.gillmer@gmail.com"}}
	app.Commands = []cli.Command{
		{
			Name:  "stop",
			Usage: "stop the docker process and remove the image",
			Action: func(c *cli.Context) {
				find(c.Args().First(), "stop")
			},
		},

		{
			Name:  "list",
			Usage: "find the docker process and display all possible information",
			Action: func(c *cli.Context) {
				find(c.Args().First(), "list")
			},
		},

		{
			Name:  "uptime",
			Usage: "shows only the uptime of the running docker process",
			Action: func(c *cli.Context) {
				println("added task: ", c.Args().First())
			},
		},

		{
			Name:  "host",
			Usage: "shows only the host information where the docker process runs",
			Action: func(c *cli.Context) {
				println("added task: ", c.Args().First())
			},
		},
	}

	app.Run(os.Args)

}
