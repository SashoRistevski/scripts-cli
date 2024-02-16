package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "docker-scripts",
	Short: "A CLI tool for docker scripts",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Docker Scripts")
		fmt.Println("Please select an option to continue")
		fmt.Println("1. Create a script")
		fmt.Println("2. Run a script")
		fmt.Println("3. Exit")
		fmt.Print("Enter your choice: ")
		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			createCmd.Run(createCmd, args)
		case 2:
			runCmd.Run(runCmd, args)
		case 3:
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid choice. Exiting...")
			os.Exit(1)
		}
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a script to run/stop docker containers with compose",
	Run:   createHandler,
}
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a script to start/stop docker containers",
	Run:   runHandler,
}
var scriptFolder string

var scriptName, appName, appFolder string

func createHandler(cmd *cobra.Command, args []string) {
	scriptName, appName, appFolder = getUserInputForStartScriptCreation()

	if err := validateScriptInputs(scriptName, appName, appFolder); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Creating start/stop scripts for: %s\n", scriptName)

	createStartScript := generateStartScriptContent(appFolder, appName)
	createStopScript := generateStopScriptContent(appFolder)

	scriptFolder = filepath.Join(getUserHomeDir(), "docker-scripts")
	if err := os.MkdirAll(scriptFolder, 0755); err != nil {
		fmt.Println("Error creating script folder:", err)
		return
	}

	// Write script to file
	startScriptFileName := filepath.Join(scriptFolder, fmt.Sprintf("start_%s.sh", scriptName))
	stopScriptFileName := filepath.Join(scriptFolder, fmt.Sprintf("stop_%s.sh", scriptName))

	err, done := writeStartScriptToFile(startScriptFileName, createStartScript)
	if done {
		return
	}
	err, done = writeStopScriptToFile(stopScriptFileName, createStopScript)
	if done {
		return
	}

	err = makeScriptExecutable(startScriptFileName, stopScriptFileName)
	if err != nil {
		fmt.Println("Error making script executable:", err)
		return
	}
	fmt.Printf("Script %s created successfully.\n", scriptName)
}

func writeStartScriptToFile(scriptFileName string, createStartScript string) (error, bool) {
	err := writeScriptToFile(scriptFileName, createStartScript)
	if err != nil {
		fmt.Println("Error creating script file", err)
		return nil, true
	}
	return err, false
}

func writeStopScriptToFile(scriptFileName string, createStartScript string) (error, bool) {
	err := writeScriptToFile(scriptFileName, createStartScript)
	if err != nil {
		fmt.Println("Error creating script file", err)
		return nil, true
	}
	return err, false
}

func runHandler(cmd *cobra.Command, args []string) {
	scriptFolder = filepath.Join(getUserHomeDir(), "docker-scripts")
	files, err := os.Open(scriptFolder)
	if err != nil {
		fmt.Println(err)
		return
	}
	scripts, err := files.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	var options []string
	for _, script := range scripts {
		if filepath.Ext(script.Name()) == ".sh" {
			options = append(options, script.Name())
		}
	}
	fmt.Println("Select a script to run")
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	var choice int
	fmt.Scanln(&choice)
	if choice < 1 || choice > len(options) {
		fmt.Println("Invalid choice")
		return
	} else {
		runScript(choice)
	}
}

func getUserInputForStartScriptCreation() (string, string, string) {
	fmt.Print("Enter the name of the script (example: project1):  ")
	fmt.Scanln(&scriptName)

	fmt.Print("Enter the app folder name (example: project_1): ")
	fmt.Scanln(&appName)

	fmt.Print("Enter the location of the project folder(example: /var/www/code): ")
	fmt.Scanln(&appFolder)

	return scriptName, appName, appFolder
}

func validateScriptInputs(scriptName, appName, appFolder string) error {
	if scriptName == "" || appName == "" || appFolder == "" {
		return fmt.Errorf("all fields must be provided")
	}

	return nil
}

func generateStartScriptContent(appFolder, appName string) string {
	scriptContent := fmt.Sprintf(
		`#!/bin/bash
	# Navigate to project directory
	echo "Navigating to %s..."
	cd %s || exit

	# Check if Docker containers are up
	if docker ps -q -f name=%s_* | grep -q "."; then
		echo "Docker Compose containers are already up."
		echo "Entering shell inside app container..."
		docker compose exec %s sh
	else
		# Start Docker Compose
		echo "Starting Docker Compose..."
		if docker compose start; then
			echo "Entering shell inside app container..."
			docker compose exec %s sh
		else
			echo "Docker Compose startup failed."
		fi
		fi`, appFolder, appFolder, appName, appName, appName)

	return scriptContent
}

func generateStopScriptContent(appFolder string) string {
	scriptContent := fmt.Sprintf(
		`#!/bin/bash

	# Change directory to the specified path
	cd %s || exit
	# Check if Docker Compose services are running
	if docker compose ps | grep -q "Up"; then
	# Stop Docker Compose services
	docker compose stop
	else
	echo "Docker services are already stopped."
	fi`, appFolder)

	return scriptContent
}

func writeScriptToFile(fileName, content string) error {
	scriptFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating script file:", err)
		return err
	}
	defer scriptFile.Close()

	_, err = scriptFile.WriteString(content)
	if err != nil {
		fmt.Println("Error writing script content:", err)
		return err
	}
	return nil
}

func makeScriptExecutable(startScriptName string, stopScriptName string) error {

	cmd := exec.Command("chmod", "+x", startScriptName, stopScriptName)
	return cmd.Run()

}

func getUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	return home
}

func runScript(choice int) {
	scriptFolder = filepath.Join(getUserHomeDir(), "docker-scripts")
	files, err := os.Open(scriptFolder)
	if err != nil {
		fmt.Println(err)
		return
	}
	scripts, err := files.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	var options []string
	for _, script := range scripts {
		if filepath.Ext(script.Name()) == ".sh" {
			options = append(options, script.Name())
		}
	}
	scriptFileName := filepath.Join(scriptFolder, options[choice-1])
	cmd := exec.Command(scriptFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running script:", err)
	}
}
