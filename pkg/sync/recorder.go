package sync

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SynchronizedBlobRecorder describes a recorder of synchronized blobs
type SynchronizedBlobRecorder struct {
	Blobs  map[string](map[string]int64)
	onDisk *bufio.Writer

	syncC chan int
}

// NewSynchronizedBlobRecorder initialize a SynchronizedBlobRecorder.
func NewSynchronizedBlobRecorder(onDisk string) error {
	if SynchronizedBlobs != nil {
		return nil
	}
	file, err := os.OpenFile(onDisk, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	synchronizedBlobs := &SynchronizedBlobRecorder{
		Blobs:  map[string](map[string]int64){},
		syncC:  make(chan int, 1),
		onDisk: nil,
	}

	// load record file
	// 这个操作主要是初始化，读取文件到内存，但不重复写到文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// each line of the recorder file on disk looks like: "<registry>,<digest_of_layer>,<layer_size>\n"
		content := strings.Split(line, ",")
		if len(content) != 3 || err != nil {
			// ignore the illegal line of recorder file
			// it will take longer for pulling image because of such illegal recorder lines
			//todo why didn't delete this line.
			continue
		}
		var size int64
		size, err = strconv.ParseInt(content[2], 10, 64)
		//todo why didn't handle this err
		err := synchronizedBlobs.Record( content[0] , content[1] , size )
		if err != nil {
			//todo at least print the err
			fmt.Println(err)
		}
	}
	synchronizedBlobs.onDisk = bufio.NewWriter(file)
	SynchronizedBlobs = synchronizedBlobs
	return nil
}

// Record information of a layer that has been synchronized
func (slr *SynchronizedBlobRecorder) Record(registry, digest string, size int64) error {
	slr.LockRecorder()
	defer slr.UnlockRecorder()
	if slr.Blobs[registry] == nil {
		slr.Blobs[registry] = map[string]int64{}
	}
	// registry -> digest -> size

	slr.Blobs[registry][digest] = size
	if slr.onDisk != nil {
		_, err := slr.onDisk.WriteString(registry + "," + digest + "," + strconv.FormatInt(size, 10) + "\n")
		if err != nil {
			//slr.UnlockRecorder()
			//todo  a dead lock maybe happen i changed
			return err
		}
	}
	return nil
}

// Query the recorder if a layer has been synchronized
func (slr *SynchronizedBlobRecorder) Query(registry, digest string) (int64, bool) {
	slr.LockRecorder()
	size, exist := slr.Blobs[registry][digest]
	slr.UnlockRecorder()
	return size, exist
}

// GetRegistryRecords gets records according related to the registry
func (slr *SynchronizedBlobRecorder) GetRegistryRecords(registry string) map[string]int64 {
	slr.LockRecorder()
	recordList := slr.Blobs[registry]
	slr.UnlockRecorder()
	return recordList
}

// UpdateRegistryRecords updates records related to the registry
func (slr *SynchronizedBlobRecorder) UpdateRegistryRecords(registry string, recordList map[string]int64) error {
	slr.LockRecorder()
	for key, value := range recordList {
		if slr.Blobs[registry] == nil {
			slr.Blobs[registry] = map[string]int64{}
		}

		slr.Blobs[registry][key] = value
		if slr.onDisk != nil {
			_, err := slr.onDisk.WriteString(registry + "," + key + "," + strconv.FormatInt(value, 10) + "\n")
			if err != nil {
				slr.UnlockRecorder()
				return err
			}
		}
	}
	slr.UnlockRecorder()
	return nil
}

// Flush records to disk
func (slr *SynchronizedBlobRecorder) Flush() {
	slr.LockRecorder()
	slr.onDisk.Flush()
	slr.UnlockRecorder()
}

// LockRecorder locks the syncC mutex
func (slr *SynchronizedBlobRecorder) LockRecorder() {
	slr.syncC <- 1
}

// UnlockRecorder unlocks the syncC mutex
func (slr *SynchronizedBlobRecorder) UnlockRecorder() {
	<-slr.syncC
}
