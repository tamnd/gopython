package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const inventoryPath = "testdata/cpython_python_files.txt"

func main() {
	list := flag.Bool("list", false, "print tracked files")
	root := flag.String("root", "", "local CPython Python directory")
	strict := flag.Bool("strict", false, "fail when local files differ from tracked files")
	flag.Parse()

	tracked, err := loadTrackedInventory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *list || *root == "" {
		for _, name := range tracked {
			fmt.Println(name)
		}
	}

	if *root == "" {
		return
	}

	local, err := listLocalFiles(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	missing, extra := compare(tracked, local)
	if len(missing) == 0 && len(extra) == 0 {
		fmt.Printf("inventory matches %s\n", *root)
		return
	}

	if len(missing) > 0 {
		fmt.Println("missing from local checkout:")
		for _, name := range missing {
			fmt.Println(name)
		}
	}
	if len(extra) > 0 {
		fmt.Println("extra in local checkout:")
		for _, name := range extra {
			fmt.Println(name)
		}
	}
	if *strict {
		os.Exit(1)
	}
}

func loadTrackedInventory() ([]string, error) {
	path, err := findUp(inventoryPath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseInventory(string(data)), nil
}

func findUp(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("%s not found", name)
		}
		dir = next
	}
}

func parseInventory(data string) []string {
	lines := strings.Split(data, "\n")
	names := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		names = append(names, line)
	}
	sort.Strings(names)
	return names
}

func listLocalFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func compare(tracked, local []string) (missing []string, extra []string) {
	trackedSet := make(map[string]bool, len(tracked))
	localSet := make(map[string]bool, len(local))
	for _, name := range tracked {
		trackedSet[name] = true
	}
	for _, name := range local {
		localSet[name] = true
	}
	for _, name := range tracked {
		if !localSet[name] {
			missing = append(missing, name)
		}
	}
	for _, name := range local {
		if !trackedSet[name] {
			extra = append(extra, name)
		}
	}
	return missing, extra
}
