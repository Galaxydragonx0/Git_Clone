package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {

	// trying out flags here
	catfile := flag.NewFlagSet("cat-file", flag.ExitOnError)
	catPara := catfile.Bool("p", true, "prints the of content type")

	lstree := flag.NewFlagSet("ls-tree", flag.ExitOnError)
	nameOnly := lstree.Bool("name-only", false, "outputs only the names of directory if added")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		catfile.Parse(os.Args[2:])
		blobReader(catfile.Args()[1])

		if *catPara {
			// display the content types
		}

	case "hash-object":
		createBlob(os.Args[3])

	case "ls-tree":
		// setting the arugments
		lstree.Parse(os.Args[2:])
		treeHash := lstree.Args()[0]
		if *nameOnly {
			treeReader(treeHash)
		}

	case "write-tree":
		path, err := os.Getwd()
		if err != nil {
			fmt.Println("Working directory could not be found !!!")
		}
		treeString, _ := treeWriter(path)

		fmt.Println(treeString)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

// function for reading the blob
func blobReader(hash string) {
	// get the path of the stored object

	// slice up our hashed string to give us this
	path := fmt.Sprintf(".git/objects/%s/%s", hash[0:2], hash[2:])

	file, _ := os.Open(path)

	//open new zlib reader to read the object

	r, err := zlib.NewReader(io.Reader(file))
	if err != nil {
		fmt.Println("Could not read the file", err)
	}

	c, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("The file could not be read", err)
	}

	// the format of the decoded object states it is separated by a null character so we deal with this like so:
	// spilt the string by the null character which x00 in bytes
	content := strings.Split(string(c), "\x00")

	//print the second part of the content to the user

	fmt.Print(content[1])

	r.Close()

}

func createBlob(file string) {
	// read the file contents
	f, err := os.ReadFile(file)

	if err != nil {
		fmt.Println("Could not read content of the file /n", err)
	}

	// get the size of the file
	s, err := os.Stat(file)
	if err != nil {
		fmt.Println("Could not read file size /n", err)
	}

	// create the blob object in plain text
	blob := fmt.Sprintf("blob %d\x00%s", s.Size(), f)

	//hash the complete blob object name using SHA-1 hashing
	h := sha1.Sum([]byte(blob))

	// convert [Size]bytes into string
	hash := string(h[:])

	//take the hash name and slice it up into the parts required for object storage
	pathPrefix := hash[0:2]
	pathSuffix := hash[2:]

	// make the path
	storePath := fmt.Sprintf(".git/objects/%s/%s", pathPrefix, pathSuffix)

	// store the contents of the blob in a buffer
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(blob))
	w.Close()

	// create the parent directory of blob storeage
	err = os.MkdirAll(fmt.Sprintf(".git/objects/%s", pathPrefix), os.ModePerm)
	if err != nil {
		fmt.Println("Could not create the required path to store file", err)
		os.Exit(1)
	}

	// store the file
	err = os.WriteFile(storePath, b.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(hash)

}

func treeReader(treeHash string) {

	// take the hash and split it to get the directory of the tree hash
	parentDir := treeHash[0:2]
	mainDir := treeHash[2:]

	//use the path to open the file
	path := fmt.Sprintf(".git/objects/%s/%s", parentDir, mainDir)

	file, _ := os.Open(path)

	//open new zlib reader to read the object

	r, err := zlib.NewReader(io.Reader(file))
	if err != nil {
		fmt.Println("Could not read the file", err)
	}

	t, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("The file could not be read", err)
	}

	// after reading it in format the output accordingly

	// split the header up
	fileParts := strings.Split(string(t), " ")

	// need to skip the header after splitting by a space

	// make an empty array of string that points to nil
	nodes := make([]string, 0)

	// loop through the rest of the file and split by nil
	for i := 2; i < len(fileParts); i++ {
		// for every mode name hash line, we split them by the null byte and add them to the array
		nodes = append(nodes, strings.Split(fileParts[i], string('\000'))[0])
	}

	// print out the array
	for _, node := range nodes {
		fmt.Println(node)
	}

}

func treeWriter(path string) (string, []byte) {

	// get the current directory
	path, err := os.Getwd()

	if err != nil {
		fmt.Println("Could not find directory??!!")
	}

	// create a struct to save hash and name
	type entry struct {
		name       string
		treeString string
	}

	var entries []entry

	// keep track of the size of the current tree
	var treeSize int

	// read the current files in the directory
	files, err := os.ReadDir(path)

	// loop through entries in the path
	for _, file := range files {

		// if it is the .git file we skip it
		if file.Name() == ".git" {
			continue
		}

		// if the file is a directory we want to recursivley get the tree hash

		if file.IsDir() {
			_, treeHash := treeWriter(fmt.Sprintf("%s/%s", path, file.Name()))
			treeBodyLine := fmt.Sprintf("040000 %s\x00%s", file.Name(), string(treeHash))
			entries = append(entries, entry{file.Name(), treeBodyLine})

			treeSize += len([]byte(treeBodyLine))
			continue
		}

		// if it is a normal file we do this
		// get file hash from helper function
		fileHash := createFileBlob(fmt.Sprintf("%s/%s", path, file.Name()))

		// create the string for the tree body to stitch together
		treeBodyLine := fmt.Sprintf("100644 %s\x00%s", file.Name(), string(fileHash))
		entries = append(entries, entry{file.Name(), treeBodyLine})

		// add to treeSize
		treeSize += len([]byte(treeBodyLine))
	}

	// read this first then read if it is directory

	// this sorts the array in ascending order
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	headerString := fmt.Sprintf("tree %d\x00", treeSize)

	// stitch together entries here
	var treeBody string

	for _, entry := range entries {
		treeBody += entry.treeString
	}

	hashString := fmt.Sprintf(headerString + treeBody)

	//hash it and return tree object hash
	sha1 := sha1.New()

	return hashString, sha1.Sum([]byte(hashString))

}

func createFileBlob(path string) []byte {
	//read the file
	f, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Error: File could not be read!!")
	}

	//get size of the file

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error: File could not be read!!")
	}

	blobString := fmt.Sprintf("blob %d\x00%s", fileInfo.Size(), string(f))

	// create hash

	sha1 := sha1.New()

	return sha1.Sum([]byte(blobString))

}
