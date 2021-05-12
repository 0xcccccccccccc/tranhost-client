package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mozillazg/go-pinyin"
	"github.com/skip2/go-qrcode"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var uri fyne.URI
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
	Openfile, err := os.Open(uri.Path())
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
	w.Header().Set("Content-Disposition", "attachment; filename="+uri.Name())
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	application.SendNotification(&fyne.Notification{Title: "传输开始", Content: "请求方:" + r.RemoteAddr})
	io.Copy(w, Openfile) //'Copy' the file to the client
	application.SendNotification(&fyne.Notification{Title: "传输完成", Content: "请求方:" + r.RemoteAddr})
	return
}

func fileOpenCallback(closer fyne.URIReadCloser, err error) {
	if err == nil {
		uri = closer.URI()

		http.HandleFunc("/"+uri.Name(), fileTransferCallback)

		if p2pStatus.protocol == IPV6 {
			go http.ListenAndServe("[::]:60000", nil)
		} else {
			go http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(p2pStatus.inner_port)), nil)
		}
		password := ""
		pe := widget.NewPasswordEntry()
		form := []*widget.FormItem{{Text: "密码", Widget: pe}}
		dialog.NewForm("设置密码", "设定密码", "不需要", form, func(b bool) {
			if b {
				md5ctx := md5.New()
				md5ctx.Write([]byte(pe.Text))
				cipherStr := md5ctx.Sum(nil)
				password = hex.EncodeToString(cipherStr)
			} else {
				password = ""
			}
			resp, err := http.Post("https://"+SERVER_ADDR+"/postfile", "application/x-www-form-urlencoded",
				strings.NewReader( /*"protocol="+string(p2pStatus.protocol)+"&"+*/
					"ip="+p2pStatus.ip+"&"+
						"port="+strconv.Itoa(int(p2pStatus.external_port))+"&"+
						"filename="+uri.Name()+"&"+
						"password="+password+"&"+
						"captcha="+strconv.Itoa(captchaCalc(p2pStatus.ip+strconv.Itoa(int(p2pStatus.external_port))+uri.Name()))))
			if err == nil {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil && resp.Status == "200 OK" {
					mainView.fileLabel.SetText("https://" + SERVER_ADDR + "" + string(body))
					mainView.fileLabel.SetURLFromString("https://" + SERVER_ADDR + string(body))
					mainView.pinyinLabel.SetText(strings.Join(pinyin.LazyPinyin(string(body), pinyin.Args{Style: pinyin.Tone}), "/"))
					var png []byte
					png, _ = qrcode.Encode("https://"+SERVER_ADDR+string(body), qrcode.Medium, 256)
					img := canvas.NewImageFromReader(bytes.NewReader(png), "qrcode.png")
					img.SetMinSize(fyne.Size{
						Width:  256,
						Height: 256,
					})
					dialog.ShowCustom("分享成功", "返回", img, window)
					mainView.successLabel.SetText("分享成功，关闭程序以取消分享")
					mainView.browseButton.SetText("显示二维码")
					mainView.browseButton.OnTapped = func() {
						dialog.ShowCustom("分享成功", "返回", img, window)
					}

				} else {
					dialog.ShowError(errors.New("与服务器的连接中断，请重试"), window)
				}

			}
		}, window).Show()

	}
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
	window.Resize(fyne.Size{Height: 600, Width: 800})
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
