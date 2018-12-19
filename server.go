package resumable

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"time"
)

type uploadFile struct {
	file       *os.File
	name       string
	tmpPath    string
	status     string
	size       int64
	transfered int64
}

var files = make(map[string]uploadFile)

type fileStorage struct {
	TmpPath string
	Path    string
}

//FileStorage ...
var FileStorage = fileStorage{
	Path:    "./files",
	TmpPath: ".tmp",
}

//HTTPHandler ....
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	ensureDir(FileStorage.Path)
	ensureDir(FileStorage.TmpPath)

	sessionID := r.Header.Get("Session-ID")
	contentRange := r.Header.Get("Content-Range")

	if r.Method != "POST" || sessionID == "" || contentRange == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Invalid request."))
		return
	}

	var upload uploadFile

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	totalSize, partFrom, partTo := parseContentRange(contentRange)

	if partFrom == 0 {
		_, ok := files[sessionID]
		if !ok {
			w.WriteHeader(http.StatusCreated)
			_, params, errs := mime.ParseMediaType(r.Header.Get("Content-Disposition"))
			if errs != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errs.Error()))
			}
			fileName := params["filename"]

			newFile := FileStorage.TmpPath + "/" + sessionID
			_, err = os.Create(newFile)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			f, errs := os.OpenFile(newFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if errs != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errs.Error()))
			}
			files[sessionID] = uploadFile{
				file:    f,
				name:    fileName,
				tmpPath: newFile,
				status:  "created",
				size:    totalSize,
			}

		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	upload = files[sessionID]
	upload.status = "uploading"
	_, err = upload.file.Write(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	upload.file.Sync()
	upload.transfered = partTo

	w.Header().Set("Content-Length", string(len(body)))
	w.Header().Set("Connection", "close")
	w.Header().Set("Range", contentRange)
	w.Write([]byte(contentRange))

	if partTo >= totalSize {
		moveToPath(sessionID)
		upload.file.Close()
		delete(files, sessionID)
	}
}

func moveToPath(id string) {
	uploadFile := files[id]
	filePath := FileStorage.Path + "/" + uploadFile.name
	if fileExists(filePath) {
		t := time.Now().Format(time.RFC3339)
		filePath = FileStorage.Path + "/" + t + "-" + uploadFile.name
	}

	err := os.Rename(uploadFile.tmpPath, filePath)
	if err != nil {
		os.Exit(1)
	}
}
