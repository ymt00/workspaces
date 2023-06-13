package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Node struct
type Node struct {
	ID            int    `json:"id"`
	AppID         string `json:"app_id"`
	Nodes         []Node `json:"nodes"`
	FloatingNodes []Node `json:"floating_nodes"`
}

// Workspace struct
type Workspace struct {
	ID             int    `json:"id"`
	Num            int    `json:"num"`
	Focused        bool   `json:"focused"`
	Representation string `json:"representation"`
	Nodes          []Node `json:"nodes"`
	FloatingNodes  []Node `json:"floating_nodes"`
}

// type Container struct {
// 	ID int `json:"id"`
// }

type WindowEvent struct {
	Change string `json:"change"`
	// Con    Container `json:"container"`
}

var (
	icons = map[string]string{}
)

func readIconFile() {
	// read apps icon file from argument
	if len(os.Args) < 2 {
		panic("Path for icon file was not passsed as argument")
	}

	appsIconFile, err := os.ReadFile(os.Args[1])
	if err != nil {
		// error reading file icon
		panic("Error reading icon file")
	}

	json.Unmarshal(appsIconFile, &icons)
}

// get all Container from workspace
func getRecursiveApps(nodes []Node, apps *[]string) {
	for _, n := range nodes {
		if n.AppID != "" {
			*apps = append(*apps, n.AppID)
		}
		getRecursiveApps(append(n.Nodes, n.FloatingNodes...), apps)

	}
}

func getApps(workspace Workspace, apps *[]string) {
	re := regexp.MustCompile(`H|V|T|S|\[|\]`)
	representation := re.ReplaceAllString(workspace.Representation, "")

	if representation != "" {
		*apps = append(*apps, strings.Split(representation, " ")...)
	}

	getRecursiveApps(append(workspace.Nodes, workspace.FloatingNodes...), apps)
}

// getActiveWorkspace get the active workspace based on the Container ID
// func getFocusedWorkspace() Workspace {
// 	cmd := exec.Command("swaymsg", "-r", "-t", "get_workspaces")
// 	out, _ := cmd.CombinedOutput()
// 	var workspaces []Workspace
// 	json.Unmarshal(out, &workspaces)
// 	var workspace Workspace
// 	for _, w := range workspaces {
// 		if w.Focused {
// 			workspace = w
// 			break
// 		}
// 	}
// 	return workspace
// }

func getWorkspaces() []Workspace {
	cmd := exec.Command("swaymsg", "-rt", "get_workspaces")

	out, _ := cmd.CombinedOutput()

	var workspaces []Workspace

	json.Unmarshal(out, &workspaces)

	return workspaces
}

func getIcon(name string) string {
	icon := icons[strings.ToLower(name)]
	if icon == "" {
		return icons["generic"]
	}

	return icon
}

func setWorkspaceName(num int, apps []string) {
	var icons string

	for _, app := range apps {
		icons = icons + " " + getIcon(app)
	}

	n := strconv.Itoa(num)

	if icons == "" {
		exec.Command(
			"swaymsg", "rename", "workspace",
			"number", n,
			"to", n,
		).Run()
	} else {
		exec.Command(
			"swaymsg", "rename", "workspace",
			"number", n,
			"to", n+":"+icons,
		).Run()
	}
}

func subscribeWindowEvent() {
	cmd := exec.Command("swaymsg", "-rmt", "subscribe", "[\"window\"]")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic("ERROR: Swaymsg was unable to subscribe to window event.")
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer([]byte(""), 1024)
	cmd.Start()

	for scanner.Scan() {
		// scanOut := scanner.Bytes()
		// var event WindowEvent
		// json.Unmarshal(scanOut, &event)
		// con := event.Con
		// fmt.Printf("\nEvent: %s\nID: %d\n", event.Change, con.ID)
		// if event.Change == "move" {
		// 	for _, w := range getWorkspaces() {
		// 		var apps []string
		// 		getApps(w, &apps)
		// 		setWorkspaceName(w.Num, apps)
		// 	}
		// } else {
		// 	var apps []string
		// 	fw := getFocusedWorkspace()
		// 	getApps(fw, &apps)
		// 	setWorkspaceName(fw.Num, apps)
		// }

		for _, w := range getWorkspaces() {
			var apps []string

			getApps(w, &apps)

			setWorkspaceName(w.Num, apps)
		}

	}
}

func main() {
	readIconFile()

	subscribeWindowEvent()
}
