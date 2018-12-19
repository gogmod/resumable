package resumable

import (
    "crypto/rand"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func generateSessionID() string {
    b := make([]byte,8)
    rand.Read(b)
    fmt.Sprintf("%X",b)
}

func generateContentRange(index uint64, fileChunk int, partSize int, totalSize int64) string{
    var contentRange string
    if index == 0 {
        contentRange = "bytes 0-" + fmt.Sprintf("%v", partSize) + "/" + fmt.Sprintf("%v", totalSize)
    }else{
        from := uint64(fileChunk) * index
        to := uint64(fileChunk) * (index + 1)
        contentRange = "bytes " + fmt.Sprintf("%v", from) + "-" + fmt.Sprintf("%v",to) + "/" + fmt.Sprintf("%v",totalSize)
    }
    return contentRange
}

func parseContentRange(contentRange string) (totalSize int, partFrom int64, partTo int64) {
    contentRange = strings.Replace(contentRange, "bytes ", "", -1)
    fromTo := strings.Split(contentRange, "/")[0]
    totalSize, err := strconv.ParseInt(strings.Split(contentRange,"/")[1], 10, 64)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    splitted := strings.Split(fromTo, "-")
    partFrom, err = strconv.ParseInt(splitted[0],10,64)
    if err != nil {
        fmt.Println(err)
		os.Exit(1)
    }
    partTo, err = strconv.ParseInt(splitted[1],10,64)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    return totalSize, partFrom, partTo
}

func parseBody(body string) int64 {
    fromTo := strings.Split(body, "/")[0]
    splitted := strings.Split(fromTo, "-")
    partTo, err := strconv.ParseInt(splitted[1],10,64)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    return partTo
}

func fileExists(filePath string) bool {
    if _, err := os.Stat(filePath); !os.IsNotExist(err) {
        return true
    }
    return false
}

func ensureDir(dirPath string) {
    if _, err := os.Stat(dirPath); os.IsNotExist(err){
        os.MkdirAll(dirPath, os.ModePerm)
    }
}




















