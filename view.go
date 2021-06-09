package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cavaliercoder/grab"
	"github.com/mozillazg/go-pinyin"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TranView struct {
	browseButton   *widget.Button
	logo           *canvas.Image
	successLable   *widget.Label
	mainContent    *fyne.Container
	fileLabel      *widget.Hyperlink
	pinyinLabel    *widget.Label
	successLabel   *widget.Label
	downloadButton *widget.Button
	entry1         *widget.Entry
	entry2         *widget.Entry
	entry3         *widget.Entry
	entry4         *widget.Entry
}

func uploadFile(uri fyne.URI, relay bool) {
	http.HandleFunc("/"+uri.Name(), fileTransferCallback)

	if p2pStatus.protocol == IPV6 {
		go http.ListenAndServe("[::]:60000", nil)
	} else {
		go http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(p2pStatus.inner_port)), nil)
	}
	password := ""
	pe := widget.NewPasswordEntry()
	form := []*widget.FormItem{{Text: "密码", Widget: pe}}
	if relay == false {
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
					go appendHash(string(body), uri.Path())
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
	} else {
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
				go appendHash(string(body), uri.Path())
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
				//dialog.ShowCustom("分享成功", "返回", img, window)
				mainView.successLabel.SetText("分享成功，关闭程序以取消分享")
				mainView.browseButton.SetText("显示二维码")
				mainView.browseButton.OnTapped = func() {
					dialog.ShowCustom("分享成功", "返回", img, window)
				}

			}
		}
	}
}

func fileOpenCallback(closer fyne.URIReadCloser, err error) {
	if err == nil {
		uploadFile(closer.URI(), false)
		//defer closer.Close()
	}
}

type DownloadContext struct {
	done       bool
	csrf_token string
	password   string
}

func askPasswordBlock() string {
	//passwnd:=application.NewWindow("需要密码")
	//passwnd.Show()
	password := "can't be such password"
	pe := widget.NewPasswordEntry()
	form := []*widget.FormItem{{Text: "密码", Widget: pe}}
	dialog.NewForm("密码", "确定", "取消", form, func(b bool) {
		if b {
			md5ctx := md5.New()
			md5ctx.Write([]byte(pe.Text))
			cipherStr := md5ctx.Sum(nil)
			password = hex.EncodeToString(cipherStr)
		} else {
			password = ""
		}
	}, window).Show()

	return password

}
func tryDownload(req *grab.Request) DownloadContext {
	client := grab.NewClient()
	resp := client.Do(req)
	downwnd := application.NewWindow("下载中")
	downbar := widget.NewProgressBar()
	downwnd.SetContent(container.NewHBox(downbar))
	downwnd.Show()
	for {
		if resp.IsComplete() {
			//checkhead:=make([]byte,1024)

			//resp.HTTPResponse.Body.Read(checkhead)
			//eigen:=[]byte{0x8B,0xF7,0x8F,0x93,0x51,0x65,0x63,0xD0,0x53,0xD6,0x5B,0xC6,0x78,0x01,0x67,0x65,0x63,0xD0,0x53,0xD6}
			if resp.HTTPResponse.Status == "401 Unauthorized" {
				downwnd.Close()
				return DownloadContext{done: false}

			} else {
				downbar.Value = 1.0
				downwnd.Close()
				dialog.ShowInformation("成功", uri.Path()+"保存成功", window)
				go uploadFile(uri, true)
				return DownloadContext{done: true}
			}
		} else {
			downbar.Value = resp.Progress()
		}
	}
}
func downloadProcess(path string) {
	req, _ := grab.NewRequest(path, "http://"+SERVER_ADDR+"/"+mainView.entry1.Text+"/"+mainView.entry2.Text+"/"+mainView.entry3.Text+"/"+mainView.entry4.Text)
	//tryDownload(req)
	ctx := tryDownload(req)
	if !ctx.done {
		pe := widget.NewPasswordEntry()
		form := []*widget.FormItem{{Text: "密码", Widget: pe}}
		dialog.NewForm("密码", "确定", "取消", form, func(b bool) {
			var password string
			if b {
				md5ctx := md5.New()
				md5ctx.Write([]byte(pe.Text))
				cipherStr := md5ctx.Sum(nil)
				password = hex.EncodeToString(cipherStr)
			} else {
				password = ""
			}
			req, _ := grab.NewRequest(uri.Path(), "http://"+SERVER_ADDR+"/"+mainView.entry1.Text+"/"+mainView.entry2.Text+"/"+mainView.entry3.Text+"/"+mainView.entry4.Text+"?password="+password)
			tryDownload(req)
		}, window).Show()

	}
}
func fileSaveCallback(closer fyne.URIWriteCloser, err error) {
	if err == nil {
		uri = closer.URI()
		go downloadProcess(uri.Path())
		closer.Close()
	}
}
func NewTranView() TranView {
	newView := TranView{}
	newView.fileLabel = widget.NewHyperlink("", &url.URL{})
	newView.pinyinLabel = widget.NewLabel("")

	newView.browseButton = widget.NewButton("选择你欲传输的文件", func() {
		dialog.ShowFileOpen(fileOpenCallback, window)
	})
	newView.downloadButton = widget.NewButton("下载", func() {
		dialog.ShowFileSave(fileSaveCallback, window)
	})

	//qrImg=container.NewVBox()

	newView.logo = canvas.NewImageFromResource(resourceIconPng)
	newView.logo.SetMinSize(fyne.NewSize(256, 256))

	newView.logo.FillMode = canvas.ImageFillOriginal
	newView.logo.ScaleMode = canvas.ImageScaleSmooth

	newView.successLabel = widget.NewLabel("你已经接入到分享网络")
	newView.successLabel.Alignment = fyne.TextAlignCenter
	newView.fileLabel.Alignment = fyne.TextAlignCenter
	newView.pinyinLabel.Alignment = fyne.TextAlignCenter
	newView.pinyinLabel.TextStyle.Monospace = true
	newView.entry1 = widget.NewEntry()
	newView.entry2 = widget.NewEntry()
	newView.entry3 = widget.NewEntry()
	newView.entry4 = widget.NewPasswordEntry()
	newView.mainContent = container.NewVBox(
		widget.NewLabel(""),
		newView.logo,
		widget.NewLabel(""),
		newView.successLabel,
		newView.fileLabel,
		newView.pinyinLabel,
		newView.browseButton,
		widget.NewLabel(""),
		widget.NewSeparator(),
		widget.NewLabel(""),
		container.NewCenter(widget.NewLabel("使用客户端进行下载，能加速自己和邻居的下载速度！")),
		container.NewAdaptiveGrid(4, newView.entry1, newView.entry2, newView.entry3, newView.entry4),
		newView.downloadButton,
	)

	return newView
}

type WaitView struct {
	container    *fyne.Container
	anim         *fyne.Animation
	spinningLogo *canvas.Image
}

// themedBox is a simple box that change its background color according
// to the selected theme
type themedBox struct {
	widget.BaseWidget
}

func newThemedBox() *themedBox {
	b := &themedBox{}
	b.ExtendBaseWidget(b)
	return b
}

func (b *themedBox) CreateRenderer() fyne.WidgetRenderer {
	b.ExtendBaseWidget(b)
	bg := canvas.NewRectangle(theme.ForegroundColor())
	return &themedBoxRenderer{bg: bg, objects: []fyne.CanvasObject{bg}}
}

type themedBoxRenderer struct {
	bg      *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *themedBoxRenderer) Destroy() {
}

func (r *themedBoxRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
}

func (r *themedBoxRenderer) MinSize() fyne.Size {
	return r.bg.MinSize()
}

func (r *themedBoxRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *themedBoxRenderer) Refresh() {
	r.bg.FillColor = theme.ForegroundColor()
	r.bg.Refresh()
}

func makeAnimationCurveItem(label string, curve fyne.AnimationCurve, yOff float32) (
	text *widget.Label, box fyne.CanvasObject, anim *fyne.Animation) {
	text = widget.NewLabel(label)
	text.Alignment = fyne.TextAlignCenter
	text.Resize(fyne.NewSize(380, 30))
	text.Move(fyne.NewPos(0, yOff))
	box = newThemedBox()
	box.Resize(fyne.NewSize(30, 30))
	box.Move(fyne.NewPos(0, yOff))

	anim = canvas.NewPositionAnimation(
		fyne.NewPos(0, yOff), fyne.NewPos(380, yOff), time.Second, func(p fyne.Position) {
			box.Move(p)
			box.Refresh()
		})
	anim.Curve = curve
	anim.AutoReverse = true
	anim.RepeatCount = 1
	return
}

func NewWaitContainer() *fyne.Container {
	label1, box1, a1 := makeAnimationCurveItem("载入中", fyne.AnimationEaseInOut, 0)
	a1.RepeatCount = fyne.AnimationRepeatForever
	a1.Start()
	return container.NewHBox(
		container.NewWithoutLayout(
			label1,
			box1,
		))

}
