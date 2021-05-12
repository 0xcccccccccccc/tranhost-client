package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"time"
)

type TranView struct {
	browseButton *widget.Button
	logo         *canvas.Image
	successLable *widget.Label
	mainContent  *fyne.Container
	fileLabel    *widget.Hyperlink
	pinyinLabel  *widget.Label
	successLabel *widget.Label
}

func NewTranView() TranView {
	newView := TranView{}
	newView.fileLabel = widget.NewHyperlink("", &url.URL{})
	newView.pinyinLabel = widget.NewLabel("")

	newView.browseButton = widget.NewButton("选择你欲传输的文件", func() {
		dialog.ShowFileOpen(fileOpenCallback, window)
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
	newView.mainContent = container.NewVBox(
		widget.NewLabel(""),
		newView.logo,
		widget.NewLabel(""),
		newView.successLabel,
		newView.fileLabel,
		newView.pinyinLabel,
		newView.browseButton,
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
