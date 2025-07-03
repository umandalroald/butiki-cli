package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const dataFileName = ".butiki_commands.json"

func getDataFilePath() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, dataFileName)
}

func loadCommands() (map[string]string, error) {
	filePath := getDataFilePath()
	commands := make(map[string]string)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return commands, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &commands)
	return commands, err
}

func saveCommands(commands map[string]string) error {
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getDataFilePath(), data, 0644)
}

func addCommand(label, command string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	if _, exists := commands[label]; exists {
		fmt.Printf("[!] Label '%s' already exists. Use 'update' instead.\n", label)
		return
	}
	commands[label] = command
	saveCommands(commands)
	fmt.Printf("[+] Added command under label '%s'.\n", label)
}

func updateCommand(label, command string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	if _, exists := commands[label]; !exists {
		fmt.Printf("[!] Label '%s' does not exist. Use 'add' instead.\n", label)
		return
	}
	commands[label] = command
	saveCommands(commands)
	fmt.Printf("[~] Updated command under label '%s'.\n", label)
}

func deleteCommand(label string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	if _, exists := commands[label]; !exists {
		fmt.Printf("[!] Label '%s' not found.\n", label)
		return
	}
	delete(commands, label)
	saveCommands(commands)
	fmt.Printf("[-] Deleted command '%s'.\n", label)
}

func listCommands() {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	if len(commands) == 0 {
		fmt.Println("[i] No commands stored.")
		return
	}
	fmt.Println("Stored Commands:")
	for label, cmd := range commands {
		fmt.Printf(" - %s: %s\n", label, cmd)
	}
}

func runCommand(label string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	cmdStr, exists := commands[label]
	if !exists {
		fmt.Printf("[!] Command '%s' not found.\n", label)
		return
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("[*] Running '%s': %s\n", label, cmdStr)
	err = cmd.Run()
	if err != nil {
		fmt.Println("[!] Error running command:", err)
	}
}

func editCommand(label string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}

	oldCommand, exists := commands[label]
	if !exists {
		fmt.Printf("[!] Label '%s' not found.\n", label)
		return
	}

	tmpFile, err := os.CreateTemp("", "butiki_edit_*.txt")
	if err != nil {
		fmt.Println("[!] Failed to create temp file:", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(oldCommand)
	tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	editedBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		fmt.Println("[!] Could not read file:", err)
		return
	}
	newCommand := strings.TrimSpace(string(editedBytes))

	if newCommand != "" && newCommand != oldCommand {
		commands[label] = newCommand
		saveCommands(commands)
		fmt.Printf("[~] Updated command '%s'.\n", label)
	} else {
		fmt.Println("[i] No changes made.")
	}
}

func searchCommands(keyword string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	found := false
	for label, command := range commands {
		if strings.Contains(label, keyword) || strings.Contains(command, keyword) {
			fmt.Printf(" - %s: %s\n", label, command)
			found = true
		}
	}
	if !found {
		fmt.Println("[i] No matching commands found.")
	}
}

func exportCommands(filePath string) {
	commands, err := loadCommands()
	if err != nil {
		fmt.Println("Error loading commands:", err)
		return
	}
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		fmt.Println("Failed to export:", err)
		return
	}
	os.WriteFile(filePath, data, 0644)
	fmt.Printf("[⇨] Exported to %s\n", filePath)
}

func importCommands(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("[!] Failed to read file:", err)
		return
	}
	var imported map[string]string
	err = json.Unmarshal(data, &imported)
	if err != nil {
		fmt.Println("[!] Invalid JSON format.")
		return
	}

	commands, _ := loadCommands()
	for k, v := range imported {
		commands[k] = v
	}
	saveCommands(commands)
	fmt.Printf("[✓] Imported %d command(s).\n", len(imported))
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  butiki add <label> <command>")
	fmt.Println("  butiki update <label> <command>")
	fmt.Println("  butiki delete <label>")
	fmt.Println("  butiki list")
	fmt.Println("  butiki run <label>")
	fmt.Println("  butiki edit <label>")
	fmt.Println("  butiki search <keyword>")
	fmt.Println("  butiki export <filename>")
	fmt.Println("  butiki import <filename>")
}

func main() {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}

	action := args[1]

	switch action {
	case "add":
		if len(args) < 4 {
			fmt.Println("[!] Usage: butiki add <label> <command>")
			return
		}
		label := args[2]
		command := strings.Join(args[3:], " ")
		addCommand(label, command)

	case "update":
		if len(args) < 4 {
			fmt.Println("[!] Usage: butiki update <label> <command>")
			return
		}
		label := args[2]
		command := strings.Join(args[3:], " ")
		updateCommand(label, command)

	case "delete":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki delete <label>")
			return
		}
		label := args[2]
		deleteCommand(label)

	case "list":
		listCommands()

	case "run":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki run <label>")
			return
		}
		label := args[2]
		runCommand(label)

	case "edit":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki edit <label>")
			return
		}
		label := args[2]
		editCommand(label)

	case "search":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki search <keyword>")
			return
		}
		searchCommands(args[2])

	case "export":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki export <filename>")
			return
		}
		exportCommands(args[2])

	case "import":
		if len(args) < 3 {
			fmt.Println("[!] Usage: butiki import <filename>")
			return
		}
		importCommands(args[2])

	default:
		printUsage()
	}
}
