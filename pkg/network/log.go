package network

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/3atlab/netroub/pkg/model"
)

func SearchFiles(initalSizes map[string]int64, root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		initalSize, exist := initalSizes[path]

		if exist && info.Size() != initalSize && !strings.Contains(path, "control.log") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error while searching log file")
		return nil, err
	}

	return files, nil
}

func MoveLogFiles(logFiles []string) error {
	//Retrieve the time for the name
	t := time.Now()
	//Generate a name for the directory
	dirName := strconv.Itoa(int(t.Month())) + "_" + strconv.Itoa(t.Day()) + "_" + strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute()) + ":" + strconv.Itoa(t.Second()) + "_" + model.Scenar.ScenarioName

	if _, err := os.Stat(model.Scenar.LogPath + "/" + model.Scenar.ScenarioName); os.IsNotExist(err) {
		err = os.Mkdir(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName, os.ModePerm)
		if err != nil {
			fmt.Println("Error while creating new directory")
			return err
		}
	}

	//Create the directory for the scenario
	err := os.Mkdir(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName+"/"+dirName, os.ModePerm)
	if err != nil {
		fmt.Println("Error while creating log directory")
		return err
	}

	//Fill the directory with the different logs generated
	for path := range logFiles {
		err := os.Mkdir(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName+"/"+dirName+"/r"+strconv.Itoa(path+1), os.ModePerm)
		if err != nil {
			fmt.Println("Error while creating device directory")
			return err
		}
		src, err := os.Open(logFiles[path])
		if err != nil {
			fmt.Println("Error while opening log file")
			return err
		}
		defer src.Close()
		destFile := filepath.Join(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName+"/"+dirName+"/r"+strconv.Itoa(path+1), filepath.Base(logFiles[path]))
		dst, err := os.Create(destFile)
		if err != nil {
			fmt.Println("Error while creating new file")
			return err
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			fmt.Println("Error while copying log into the new file")
			return err
		}
	}

	err = MoveControlLogs(dirName)
	if err != nil {
		return err
	}

	for i := 0; i < len(logFiles); i++ {

		err = MoveTcpdumpLogs(dirName, "r"+strconv.Itoa(i+1), i)
		if err != nil {
			return err
		}

	}

	return nil
}

func MoveControlLogs(dirName string) error {
	//Move the control log file in the created directory
	control, err := os.Open("control.log")
	if err != nil {
		fmt.Println("Error while opening control log file")
		return err
	}
	defer control.Close()
	destFile := filepath.Join(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName+"/"+dirName, filepath.Base("control.log"))
	dst, err := os.Create(destFile)
	if err != nil {
		fmt.Println("Error while creating new control log file")
		return err
	}
	_, err = io.Copy(dst, control)
	if err != nil {
		fmt.Println("Error while copying control log into the new file")
		return err
	}
	err = os.Remove("control.log")
	if err != nil {
		return err
	}
	return nil
}

func MoveTcpdumpLogs(dirName string, device string, index int) error {

	err := os.Mkdir(model.Scenar.LogPath+"/"+model.Scenar.ScenarioName+"/"+dirName+"/"+device+"/tcpdump", 0777)
	if err != nil {
		return err
	}

	for _, inter := range model.Devices.Nodes[index].Interfaces {
		tcpdumpFile, err := os.Open(model.FindTopoPath() + device + "/tcpdump/tcpdump_" + inter.Name + ".log")
		if err != nil {
			fmt.Println("Error while opening tcpdump log file")
			return err
		}
		defer tcpdumpFile.Close()

		dst, err := os.Create(model.Scenar.LogPath + "/" + model.Scenar.ScenarioName + "/" + dirName + "/" + device + "/tcpdump/tcpdump_" + inter.Name + ".log")
		if err != nil {
			fmt.Println("Error while creating new tcpdump log file")
			return err
		}
		_, err = io.Copy(dst, tcpdumpFile)
		if err != nil {
			fmt.Println("Error while copying tcpdump log into the new file")
			return err
		}

	}

	/*err = os.Remove(model.FindTopoPath() + device + "/tcpdump")
	if err != nil {
		return err
	}*/
	return nil

}

func GetTcpdumpLogs(nbFile int) error {

	containerNameArray := strings.Split(model.Scenar.Event[0].Host, "-")
	containerName := strings.Join(containerNameArray[:len(containerNameArray)-1], "-")

	for i := 0; i < nbFile; i++ {
		cmd := exec.Command("sudo", "docker", "cp", containerName+"-r"+strconv.Itoa(i+1)+":/tcpdump", model.FindTopoPath()+"r"+strconv.Itoa(i+1)+"/")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error while moving tcpdump directory")
			log.Println(string(output))
			return err
		}
	}

	return nil
}

func FlushLogFiles(logFiles []string) error {
	for path := range logFiles {
		err := os.Truncate(logFiles[path], 0)
		if err != nil {
			fmt.Println("Error while flushing file")
			return err
		}
	}
	return nil
}

func TcpdumpLog(index int) error {

	containerNameArray := strings.Split(model.Scenar.Event[0].Host, "-")
	containerName := strings.Join(containerNameArray[:len(containerNameArray)-1], "-")

	file, err := os.Create(model.FindTopoPath() + "/r" + strconv.Itoa(index+1) + "/" + "tcpdump.sh")
	if err != nil {
		fmt.Println("Error while creating tcpdump.sh file")
		return err
	}
	defer file.Close()

	err = os.Chmod(model.FindTopoPath()+"/r"+strconv.Itoa(index+1)+"/"+"tcpdump.sh", 0775)
	if err != nil {
		fmt.Println("Error while changing tcpdump.sh permission")
		return err
	}

	cmd := exec.Command("sudo", "docker", "exec", "-d", containerName+"-r"+strconv.Itoa(index+1), "mkdir", "tcpdump")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while creating tcpdump directory")
		log.Println(string(output))
		return err
	}

	_, err = file.WriteString("#!/bin/sh \n")
	if err != nil {
		fmt.Println("Error while writing in tcpdump.sh file")
		return err
	}

	for _, inter := range model.Devices.Nodes[index].Interfaces {
		_, err = file.WriteString("tcpdump -i " + inter.Name + " -n -v > tcpdump/tcpdump" + "_" + inter.Name + ".log & \n")
		if err != nil {
			fmt.Println("Error while writing in tcpdump.sh file")
			return err
		}
	}

	cmd = exec.Command("sudo", "docker", "cp", model.FindTopoPath()+"/r"+strconv.Itoa(index+1)+"/"+"tcpdump.sh", containerName+"-r"+strconv.Itoa(index+1)+":/")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while copying tcpdump script in the host container")
		log.Println(string(output))
		return err
	}

	cmd = exec.Command("sudo", "docker", "exec", "-d", containerName+"-r"+strconv.Itoa(index+1), "./tcpdump.sh")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while starting tcpdump")
		log.Println(string(output))
		return err
	}

	return nil
}
