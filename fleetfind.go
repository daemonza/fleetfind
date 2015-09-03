// fleetfind quickly finds a docker process on a fleet cluster
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
)

// FleetUnit contains details about the docker process
type FleetUnit struct {
	Image   string // docker image name
	Uptime  string // uptime of the process
	Host    string // host where the process run
	Command string // command that started the process
}

func fleetSSH(host string, command string) (unit *FleetUnit) {
	sshCmd := "fleetctl ssh " + host + " \"docker ps | grep " + command + "\""
	sshOut, _ := exec.Command("sh", "-c", sshCmd).Output()
	if string(sshOut) == "" {
		return nil
	}

	fleetUnit := new(FleetUnit)
	fleetUnit.Image = string(sshOut)
	fleetUnit.Host = host
	return fleetUnit
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

	resultsChannel := make(chan *FleetUnit)

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
				fmt.Println(unitDetails.Host)
				fmt.Println(unitDetails.Image)
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
