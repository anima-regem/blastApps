package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

type progressBarWriter struct {
	bar *pb.ProgressBar
	dst io.Writer
}

func (pw *progressBarWriter) Write(p []byte) (n int, err error) {
	n, err = pw.dst.Write(p)
	pw.bar.Add(n)
	return
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go source_file destination_file")
		return
	}

	sourceFile := os.Args[1]
	destinationFile := os.Args[2]

	err := copyFileWithProgressBar(sourceFile, destinationFile)
	if err != nil {
		fmt.Printf("Error copying file: %s\n", err)
		return
	}

	fmt.Printf("Successfully copied '%s' to '%s'\n", sourceFile, destinationFile)
}

func copyFileWithProgressBar(sourceFile, destinationFile string) error {
	// Open the source file
	src, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Get the file size
	srcStat, err := src.Stat()
	if err != nil {
		return err
	}
	fileSize := srcStat.Size()

	// Create a progress bar
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)

	// Create the custom writer
	writer := &progressBarWriter{
		bar: bar,
		dst: dst,
	}

	// Copy the contents from source to destination using the custom writer
	_, err = io.Copy(writer, src)
	if err != nil {
		return err
	}

	bar.Finish()

	// Flush any buffered data to ensure all data is written to the file
	err = dst.Sync()
	if err != nil {
		return err
	}

	return nil
}
