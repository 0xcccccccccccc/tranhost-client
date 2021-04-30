package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mozillazg/go-pinyin"
	"github.com/skip2/go-qrcode"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var uri fyne.URI
var ip string
var window fyne.Window
var application fyne.App
var fileLabel *widget.Hyperlink
var browseButton *widget.Button
var qrImg *fyne.Container
var successLabel *widget.Label
var pinyinLabel *widget.Label



func fileTransferCallback(w http.ResponseWriter,r *http.Request)  {
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
	application.SendNotification(&fyne.Notification{Title: "传输开始", Content: "请求方:"+r.RemoteAddr})
	io.Copy(w, Openfile) //'Copy' the file to the client
	application.SendNotification(&fyne.Notification{Title: "传输完成", Content: "请求方:"+r.RemoteAddr})
	return
}

func fileOpenCallback(closer fyne.URIReadCloser,err error){
	if(err==nil){
		uri =closer.URI()

		http.HandleFunc("/"+uri.Name(), fileTransferCallback)
		go http.ListenAndServe("[::]:60000",nil)

		password:=""
		pe:=widget.NewPasswordEntry()
		form:=[]*widget.FormItem{{Text: "密码", Widget: pe}}
		dialog.NewForm("设置密码","设定密码","不需要",form, func(b bool) {
			if(b){
				md5ctx:=md5.New()
				md5ctx.Write([]byte(pe.Text))
				cipherStr:=md5ctx.Sum(nil)
				password=hex.EncodeToString(cipherStr)
			}else {
				password=""
			}
			resp,err:=http.Post("https://tran.host/postfile","application/x-www-form-urlencoded",
				strings.NewReader("ipv6="+ip+"&"+
									"port="+strconv.Itoa(60000)+"&"+
									"filename="+uri.Name()+"&"+
									"password="+password))
			if(err==nil){
				defer resp.Body.Close()
				body,err:=ioutil.ReadAll(resp.Body)
				if(err==nil && resp.Status=="200 OK"){
					fileLabel.SetText("https://tran.host"+string(body))
					fileLabel.SetURLFromString("https://tran.host"+string(body))
					pinyinLabel.SetText(strings.Join(pinyin.LazyPinyin(string(body),pinyin.Args{Style:pinyin.Tone }),"/"))
					var png []byte
					png, _ = qrcode.Encode("https://tran.host"+string(body), qrcode.Medium, 256)
					img:=canvas.NewImageFromReader(bytes.NewReader(png),"qrcode.png")
					img.SetMinSize(fyne.Size{
						Width:  256,
						Height:256,
					})
					dialog.ShowCustom("分享成功","返回",img, window)
					//browseButton.SetText("取消分享"+ uri.Name()+"并退出")
					//browseButton.OnTapped= func() {
					//	os.Exit(0)
					//}
					successLabel.SetText("分享成功，关闭程序以取消分享")
					browseButton.SetText("显示二维码")
					browseButton.OnTapped= func() {
						dialog.ShowCustom("分享成功","返回",img, window)
					}

				}else{
					dialog.ShowError(errors.New("与服务器的连接中断，请重试"), window)
				}

			}
		}, window).Show()



	}
}


func main() {

	// Disable HTTP Proxies
	os.Setenv("https_proxy","")
	os.Setenv("http_proxy","")

	resp,err:=http.Get("http://v6.ip.zxinc.org/getip")
	if(err==nil){
		defer resp.Body.Close()
		body,err:=ioutil.ReadAll(resp.Body)
		if(err==nil){
			ip =string(body)
		}
	}
	//cwd,err:=os.Getwd()
		//os.Setenv("FYNE_FONT",cwd+string(os.PathSeparator)+"FZHTJW.TTF")
		application = app.New()
		application.Settings().SetTheme(&myTheme{})
		window = application.NewWindow("Tran")
		fileLabel =widget.NewHyperlink("",&url.URL{})
		pinyinLabel=widget.NewLabel("")

		browseButton =widget.NewButton("选择你欲传输的文件", func() {
			dialog.ShowFileOpen(fileOpenCallback, window)
		})

		//qrImg=container.NewVBox()
		window.Resize(fyne.Size{Height: 600,Width: 800})

		logo :=canvas.NewImageFromResource(resourceIconPng)
		logo.SetMinSize(fyne.NewSize(256,256))

		logo.FillMode=canvas.ImageFillOriginal
		logo.ScaleMode=canvas.ImageScaleSmooth


		successLabel=widget.NewLabel("你已经接入到分享网络")
		successLabel.Alignment=fyne.TextAlignCenter
		fileLabel.Alignment=fyne.TextAlignCenter
		pinyinLabel.Alignment=fyne.TextAlignCenter
		pinyinLabel.TextStyle.Monospace=true


		window.SetContent(container.NewVBox(
			widget.NewLabel(""),
			logo,
			widget.NewLabel(""),
			successLabel,
			fileLabel,
			pinyinLabel,
			browseButton,
		))

		if(len(ip)<2){
			dialog.ShowConfirm("网络错误","没有检测到IPV6地址,请尝试关闭代理并重新连接网络", func(b bool) {
				os.Exit(-1)
			}, window)
		}
		window.ShowAndRun()

}
