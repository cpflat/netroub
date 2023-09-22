package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/3atlab/netroub/pkg/events"
	"github.com/3atlab/netroub/pkg/model"
	"github.com/3atlab/netroub/pkg/network"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var initalSizes map[string]int64

type ConsoleHook struct{}

func (h *ConsoleHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= logrus.InfoLevel {
		t := entry.Time
		fmt.Print(color.BlueString("[INFO]"))
		fmt.Print(t.Format("2006-01-02 15:04:05 "))
		fmt.Print(color.GreenString(entry.Message), " ")
		if entry.Data["command"] != nil {
			fmt.Println(entry.Data["command"], "duration :", entry.Data["duration"])
		} else {
			fmt.Println("")
		}
	}
	return nil
}
func (h *ConsoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func NewConsoleHook() *ConsoleHook {
	return &ConsoleHook{}
}

func main() {
	app := cli.NewApp()
	app.Name = "Netroub"
	app.Usage = "Netroub is a synthetic data generator from network trouble scenarios"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "Colin Regal-Mezin",
			Email: "colin.regalmezin@gmail.com",
		},
	}
	app.EnableBashCompletion = true
	if len(os.Args) > 1 {
		model.SudoCheck()
		app.Action = runScenario
		app.After = after

	}
	app.Before = before
	app.CustomAppHelpTemplate = model.ConfigTemplate()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "yaml",
			Usage: "Use a .yaml scenario file",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func runScenario(c *cli.Context) error {
	var err error

	controlLogFile, err := os.Create("control.log")
	if err != nil {
		fmt.Println("Error while creating control log file")
		return err
	}
	logrus.SetOutput(controlLogFile)

	//Read the scenario file and sort it by time in an array
	if c.Bool("yaml") {
		err = model.ReadYaml()
		if err != nil {
			return err
		}
	} else {
		err = model.ReadJsonScenar()
		if err != nil {
			return err
		}
	}

	//Read the dot2net data json file containing device information
	err = model.ReadJsonData()
	if err != nil {
		return err
	}

	//Stock the size of all the log file present in the directory of the topo file
	path := model.FindTopoPath()
	initalSizes = make(map[string]int64)
	initalSizes, err = model.StockInitialSize(initalSizes, path)
	if err != nil {
		return err
	}
	//Create the DockerClient which is mandatory for pumba command
	err = network.CreateDockerClient(c)
	if err != nil {
		return err
	}
	//Emulate the network with Containerlab
	err = network.EmulateNetwork()
	if err != nil {
		return err
	}

	nbFile, err := countSubDir()
	if err != nil {
		return err
	}

	for i := 0; i < nbFile; i++ {
		err = network.TcpdumpLog(i)
		if err != nil {
			return err
		}
	}

	//Create a channel to verify routine states
	done := make(chan bool)
	//Run for all the events in the scenario file
	for i := 0; i < len(model.Scenar.Event); i++ {
		go func(index int) {
			event := model.Scenar.Event[index]
			timeToWait := time.Duration(event.BeginTime) * time.Second

			time.Sleep(timeToWait)

			events.ExecuteEvent(index)

			done <- true
		}(i)
	}

	//Wait here until all routines are finished
	for i := 0; i < len(model.Scenar.Event); i++ {
		<-done
	}
	return nil
}

func before(c *cli.Context) error {

	/*Useless*/
	c.Args() //Permit to remove an unsed paramater warning
	/*Useless*/
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})

	logrus.AddHook(NewConsoleHook())
	return nil
}

func after(c *cli.Context) error {

	/*Useless*/
	c.Args() //Permit to remove a unsed paramater warning
	/*Useless*/

	//Find the directory to search log file
	path := model.FindTopoPath()
	//Fill an array with all log file path
	logFiles, err := network.SearchFiles(initalSizes, path)
	if err != nil {
		return err
	}
	//Move tcpdump log files
	network.GetTcpdumpLogs(len(logFiles))
	//Destroy the emulated network
	network.DestroyNetwork()

	err = network.MoveLogFiles(logFiles)
	if err != nil {
		return err
	}
	//Flush log files for the next scenario
	err = network.FlushLogFiles(logFiles)
	if err != nil {
		return err
	}
	return nil
}

func countSubDir() (int, error) {
	count := 0

	file, err := os.Open(model.FindTopoPath())
	if err != nil {
		return count, err
	}
	defer file.Close()

	dir, err := file.ReadDir(-1)
	if err != nil {
		fmt.Println("Error while reading topo dir")
		return count, err
	}
	for _, subDir := range dir {
		if subDir.IsDir() {
			count++
		}
	}
	return count, nil
}
