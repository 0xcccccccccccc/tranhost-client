package main

import (
	"crypto/md5"
	"encoding/hex"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"github.com/OneOfOne/xxhash"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var lastURI fyne.URI
var p2pStatus P2PType
var application fyne.App
var window fyne.Window
var wait_window fyne.Window
var mainView TranView

func captchaCalc(s string) int {
	md5ctx := md5.New()
	i := 0
	for {
		md5ctx.Write([]byte(s + strconv.Itoa(i)))
		cipherStr := md5ctx.Sum(nil)
		if hex.EncodeToString(cipherStr)[:3] == "000" {
			return i
		} else {
			i += 1
			md5ctx.Reset()
		}

	}
}

func fileTransferCallback(w http.ResponseWriter, r *http.Request) {
	//Check if file exists and open
	Openfile, err := os.Open(lastURI.Path())
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}

	FileHeader := make([]byte, 512)
	Openfile.Read(FileHeader)
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+lastURI.Name())
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	//application.SendNotification(&fyne.Notification{Title: "传输开始", Content: "请求方:" + r.RemoteAddr})
	io.Copy(w, Openfile) //'Copy' the file to the client
	//application.SendNotification(&fyne.Notification{Title: "传输完成", Content: "请求方:" + r.RemoteAddr})
	return
}
func appendHash(uuid string, path string) {
	f, _ := os.Open(path)
	h := xxhash.New32()
	io.Copy(h, f)
	checksum := strconv.FormatInt(int64(h.Sum32()), 16)
	http.Post("https://"+SERVER_ADDR+"/appendhash", "application/x-www-form-urlencoded",
		strings.NewReader( /*"protocol="+string(p2pStatus.protocol)+"&"+*/
			"hash="+checksum+"&"+"uuid="+uuid))
}

func InitProcess() {
	// Disable HTTP Proxies
	os.Setenv("https_proxy", "")
	os.Setenv("http_proxy", "")

	p2pStatus = P2PHelper()
	mainView = NewTranView()
	if p2pStatus.protocol == UNAVALIABLE {
		dialog.ShowConfirm("网络错误", "无法连接到共享网络,请尝试关闭代理并重新连接网络", func(b bool) {
			os.Exit(-1)
		}, window)
	}
	window.Resize(fyne.Size{Width: 400, Height: 800})
	window.SetContent(mainView.mainContent)
}
func main() {
	application = app.New()
	application.Settings().SetTheme(&myTheme{})
	window = application.NewWindow("Tran")

	go InitProcess()
	window.SetContent(NewWaitContainer())
	window.Resize(fyne.Size{Width: 400, Height: 50})
	window.ShowAndRun()
}
