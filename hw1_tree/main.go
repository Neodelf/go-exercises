package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	err = listOfFiles(out, path, 0, make(map[int]bool), printFiles)
	if err != nil {
		return err
	}
	return nil
}

func listOfFiles(out io.Writer, path string, level int, lastLevels map[int]bool, printFiles bool) error {
	var pathBuffer bytes.Buffer
	pathBuffer.WriteString(path)

	files, err := ioutil.ReadDir(pathBuffer.String())

	if err != nil {
		return fmt.Errorf("Directory is not found")
	}

	var list []os.FileInfo
	for _, file := range files {
		if file.Name() == ".DS_Store" {
			continue
		}
		if printFiles == false {
			if file.IsDir() == true {
				list = append(list, file)
			}
		} else {
			list = append(list, file)
		}
	}

	for _, file := range list {
		fileName := file.Name()

		fmt.Fprint(out, tabulation(level, lastLevels))
		if list[len(list)-1] == file {
			lastLevels[level] = true
			fmt.Fprint(out, "└───")
		} else {
			fmt.Fprint(out, "├───")
		}

		if file.IsDir() == true {
			fmt.Fprintln(out, fileName)
		} else if file.Size() == 0 {
			fmt.Fprintln(out, fileName, "(empty)")
		} else {
			filename := fmt.Sprintf("%s %s%d%s", fileName, "(", file.Size(), "b)")
			fmt.Fprintln(out, filename)
		}

		pathBuffer.WriteString("/")
		pathBuffer.WriteString(fileName)
		listOfFiles(out, pathBuffer.String(), level+1, lastLevels, printFiles)

		delete(lastLevels, level)
		pathBuffer.Reset()
		pathBuffer.WriteString(path)
	}
	return nil
}

func tabulation(level int, lastLevels map[int]bool) string {
	var bytes bytes.Buffer

	for i := 0; i < level; i++ {
		_, ok := lastLevels[i]
		if ok == true {
			bytes.WriteString("	")
		} else {
			bytes.WriteString("│	")
		}
	}
	return bytes.String()
}
