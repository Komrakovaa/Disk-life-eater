package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"
	"time"
)

type flags struct {
	flagHelpPtr        *bool
	flagPatternSizePtr *uint
	flagPrefixPtr      *string
	flagFileSizePtr    *uint
	flagPathPtr        *string
	flagRemoveOnExit   *bool
	flagMaxUsedSpace   *uint64
}

var (
	RND_PATTERN_SIZE uint   = 16384
	PREFIX           string = "KILLSSD"
	FILE_SIZE        uint   = 1073741824 //1 Gb
	PATH             string = "/tmp/"
	REMOVE_ON_EXIT   bool   = true
	MAX_USED_SPACE   uint64 = 0
)

var pattern []byte
var bytes_count int64 = 0

func sprintf(s uint) string {
	i := fmt.Sprintf("%06d", s)
	return i
}

func path_free_space(path string) uint64 {
	usage := NewDiskUsage(path)
	return usage.Available()
}

func init_flags() flags {
	var f flags
	f.flagHelpPtr = flag.Bool("h", false, "Help")
	f.flagPatternSizePtr = flag.Uint("b", RND_PATTERN_SIZE, "Random pattern size, bytes")
	f.flagFileSizePtr = flag.Uint("s", FILE_SIZE, "Junk file size")
	f.flagPrefixPtr = flag.String("p", PREFIX, "Junk file prefix")
	f.flagPathPtr = flag.String("path", PATH, "Path for writing junk files")
	f.flagRemoveOnExit = flag.Bool("r", REMOVE_ON_EXIT, "Remove junk on exit")
	f.flagMaxUsedSpace = flag.Uint64("m", MAX_USED_SPACE, "Maximum disk space to use(not implemented)")
	flag.Parse()
	return f
}

func init_params(f flags) {
	RND_PATTERN_SIZE = *f.flagPatternSizePtr
	if RND_PATTERN_SIZE > 1024*1024 {
		RND_PATTERN_SIZE = 1024 * 1024
	}

	PREFIX = *f.flagPrefixPtr
	FILE_SIZE = *f.flagFileSizePtr
	PATH = *f.flagPathPtr
}

func init_pattern() {
	pattern = make([]byte, RND_PATTERN_SIZE)
	_, err := crand.Read(pattern)
	if err != nil {
		fmt.Println("Random pattern init error:", err)
		os.Exit(1)
	}
}

func write_pattern(fp *os.File, size uint) {
	if size > RND_PATTERN_SIZE {
		size = RND_PATTERN_SIZE
	}
	_, err := fp.Write(pattern[0:size])
	if err != nil {
		fmt.Println("Unable write to file:", fp.Name(), "-", err)
		os.Exit(1)
	} else {
		bytes_count += int64(size)
	}

}

func create_junk_file(filename string, size uint) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create file:", filename, "-", err)
		os.Exit(1)
	}
	defer file.Close()

	var i uint
	for i = 0; i < size/RND_PATTERN_SIZE; i++ {
		write_pattern(file, RND_PATTERN_SIZE)
	}
	write_pattern(file, size%RND_PATTERN_SIZE)
}

func remove_all_junk() {
	if !REMOVE_ON_EXIT {
		return
	}
	files, err := os.ReadDir(PATH)
	if err != nil {
		fmt.Println("Unable list path:", err)
		os.Exit(1)
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), PREFIX) {
				err := os.Remove(PATH + file.Name())
				if err != nil {
					fmt.Println("Unable remove path:", file.Name(), "-", err)
				}

			}

		}
	}
}

func delete_rnd_file_with_prefix() {
	var junk_files []string
	files, err := os.ReadDir(PATH)

	if err != nil {
		fmt.Println("Unable list path:", PATH, "-", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), PREFIX) {
				junk_files = append(junk_files, file.Name())
			}

		}
	}

	file_to_remove := junk_files[mrand.Intn(len(junk_files))]

	fmt.Println("Gonna delete...", file_to_remove)
	err = os.Remove(PATH + file_to_remove)
	if err != nil {
		fmt.Println("Can`t delete:", file_to_remove, "-", err)
	}
}

func run(quit chan bool) {
	var file_count uint = 0
	fmt.Println("Starting to write...")
	for {
		select {
		case <-quit:
			return
		default:
			if path_free_space(PATH) < uint64(FILE_SIZE) {
				delete_rnd_file_with_prefix()
			}

			start := time.Now()
			create_junk_file(PATH+PREFIX+sprintf(file_count), FILE_SIZE)
			elapsed := time.Since(start)

			speed := float64(FILE_SIZE) / elapsed.Seconds() / 1048576.0
			fmt.Println("Total writed:", ByteCountDecimal(bytes_count), "Speed:", FloatToFixedPrec(speed, 2), "Mb/s")

			file_count++
		}
	}
}

func main() {
	defer remove_all_junk()

	f := init_flags()
	init_params(f)

	if *f.flagHelpPtr {
		fmt.Println("Disk life Eater v0.1")
		fmt.Println("github.com/Komrakovaa/Disk-life-eater")
		fmt.Println("Usage: diskeater [flags]")
		flag.PrintDefaults()
		os.Exit(0)
	}
	init_pattern()

	quit := make(chan bool)
	go run(quit)
	fmt.Println("Press enter key to exit")
	fmt.Scanln()

	// Quit goroutine
	quit <- true
}
