package main

import (
	"fmt"
	"os"
	"os/exec"
	cobra "github.com/spf13/cobra"
	"path/filepath"
)

var rootCmd = &cobra.Command{
    Use:   "docker-scripts",
    Short: "A CLI tool for docker scripts",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Specify a sub-command please")
    },
}

var createCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a script to run/stop docker containers with compose",
    Run:   createHandler,
}

var scriptName, appName, appFolder string

func createHandler(cmd *cobra.Command, args []string) {
	// TODO: Implement script creation logic based on user inputs
	scriptName, appName, appFolder = getUserInputForScriptCreation()

	if err := validateScriptInputs(scriptName, appName, appFolder); err != nil {
		fmt.Println("Error:", err)
		return
	}

	
	fmt.Printf("Creating script: %s\n", scriptName)
	// here we need to do the actual script creation and write it to a file and run the chmod command to make it executable
	createScript := generateScriptContent(appFolder, appName)

		// Decide the script folder based on sudo privileges
	var scriptFolder string
	if os.Geteuid() == 0 {
		// If running with sudo, use system-wide /usr/local/bin
		scriptFolder = "/usr/local/bin"
	} else {
		// If not running with sudo, use a default folder
		scriptFolder = filepath.Join(getUserHomeDir(), "docker-scripts")
	}
		// Create the folder if it doesn't exist
	if err := os.MkdirAll(scriptFolder, 0755); err != nil {
		fmt.Println("Error creating script folder:", err)
		return
	}


	// Write script to file
	scriptFileName := filepath.Join(scriptFolder, fmt.Sprintf("%s.sh", scriptName))

	err := writeScriptToFile(scriptFileName, createScript)
	if err != nil {
		fmt.Println("Error creating script file", err)
		return
	}

	err = makeScriptExecutable(scriptFileName)
	if err != nil {
		fmt.Println("Error making script executable:", err)
		return
	}

	if os.Geteuid() != 0 {
		fmt.Println("To make the script globally accessible, move it to /usr/local/bin:")
		fmt.Printf("sudo mv %s /usr/local/bin/\n", scriptFileName)
	}

	fmt.Printf("Script %s created successfully.\n", scriptName)
}

// TODO: Implement runHandler
// check for available scripts in the created dir and make a select for which to run
// run the selected script

func getUserInputForScriptCreation() (string, string, string) {
	fmt.Print("Enter the name of the script: ")
	fmt.Scanln(&scriptName)

	fmt.Print("Enter the name of the Docker app: ")
	fmt.Scanln(&appName)

	fmt.Print("Enter the location of the Docker project folder: ")
	fmt.Scanln(&appFolder)

	return scriptName, appName, appFolder
}

func validateScriptInputs(scriptName, appName, appFolder string) error {
	if scriptName == "" || appName == "" || appFolder == "" {
		return fmt.Errorf("all fields must be provided")
	}

	return nil
}

func generateScriptContent(appFolder, appName string) string {
	scriptContent := fmt.Sprintf(`#!/bin/bash
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

	func makeScriptExecutable(fileName string) error{

		cmd := exec.Command("chmod", "+x", fileName)
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
func main() {
	rootCmd.AddCommand(createCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


