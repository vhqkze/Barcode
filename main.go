package main

import (
	"fmt"
	"image"
	"strings"
	"vhqkze/Barcode/config"

	"math/rand"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/widget"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/flopp/go-findfont"
)

var conf *config.Config

func main() {
	reload()
	if _, ok := os.LookupEnv("FYNE_FONT"); ok == false {
        fmt.Println("FYNE_FONT not set")
        os.Setenv("FYNE_FONT", getFont())
	}
	if conf.Theme == "light" {
		os.Setenv("FYNE_THEME", "light")
	} else if conf.Theme == "dark" {
		os.Setenv("FYNE_THEME", "dark")
	}

	show_help_table := false
	ticker := time.NewTicker(time.Second * time.Duration(conf.AutoRefresh.Interval))

	myApp := app.New()
	window := myApp.NewWindow(" ")
	myCanvas := window.Canvas()

	tracking_number := getTrackingNumber()
	content := updateCanvas(tracking_number)
	myCanvas.SetContent(content)
	fmt.Println(tracking_number)
	lastShipId := ""

	refresh := desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierShortcutDefault}
	myCanvas.AddShortcut(&refresh, func(shortcut fyne.Shortcut) {
		ticker.Stop()
		reload()
		tracking_number = getTrackingNumber()
		content = updateCanvas(tracking_number)
		myCanvas.SetContent(content)
	})
	myCanvas.AddShortcut(&fyne.ShortcutCopy{}, func(shortcut fyne.Shortcut) {
		window.Clipboard().SetContent(tracking_number)
	})
	myCanvas.AddShortcut(&fyne.ShortcutPaste{}, func(shortcut fyne.Shortcut) {
		clip_content := window.Clipboard().Content()
		tracking_number = strings.TrimSpace(clip_content)
		fmt.Println(tracking_number)
		content = updateCanvas(tracking_number)
		myCanvas.SetContent(content)
	})
	myCanvas.SetOnTypedKey(func(ke *fyne.KeyEvent) {
		keys := []fyne.KeyName{fyne.KeyB, fyne.KeyN, fyne.KeyM, fyne.KeyR, fyne.KeySpace, fyne.KeyH, fyne.KeyO, fyne.KeyP}
		if ke.Name == fyne.KeyB { // 显示/隐藏运单号
			conf.TrackingNumber.Show = !conf.TrackingNumber.Show
		} else if ke.Name == fyne.KeyN { // 显示/隐藏姓名
			conf.Username.Show = !conf.Username.Show
		} else if ke.Name == fyne.KeyM { // 显示/隐藏手机号
			conf.Mobile.Show = !conf.Mobile.Show
		} else if ke.Name == fyne.KeyR || ke.Name == fyne.KeySpace { // 刷新运单条码
			tracking_number = getTrackingNumber()
		} else if ke.Name == fyne.KeyH { // 显示货架码
			tracking_number = getShelfNum()
			conf.Username.Show = false
			conf.Mobile.Show = false
		} else if ke.Name == fyne.KeyO {
			tracking_number = tracking_number_add_one(tracking_number, -1)
		} else if ke.Name == fyne.KeyP {
			tracking_number = tracking_number_add_one(tracking_number, 1)
		}

		if contain(keys, ke.Name) {
			content = updateCanvas(tracking_number)
			myCanvas.SetContent(content)
			fmt.Println(tracking_number)
		}

		if ke.Name == fyne.KeySlash { // 显示帮助
			show_help_table = !show_help_table
			if show_help_table {
				myCanvas.SetContent(showHelp())
			} else {
				myCanvas.SetContent(content)
			}
		}

		if ke.Name == fyne.KeyS {
			// 循环随机
			conf.AutoRefresh.Enable = !conf.AutoRefresh.Enable
			if conf.AutoRefresh.Enable {
				tracking_number = getTrackingNumber()
				content = updateCanvas(tracking_number)
				myCanvas.SetContent(content)
				fmt.Println(tracking_number)
				ticker = time.NewTicker(time.Second * time.Duration(conf.AutoRefresh.Interval))
				go func() {
					for range ticker.C {
						tracking_number = getTrackingNumber()
						content = updateCanvas(tracking_number)
						myCanvas.SetContent(content)
						fmt.Println(tracking_number)
					}
				}()
			} else {
				ticker.Stop()
			}
		}
		numberKeys := map[fyne.KeyName]string{
			fyne.Key0: "0",
			fyne.Key1: "1",
			fyne.Key2: "2",
			fyne.Key3: "3",
			fyne.Key4: "4",
			fyne.Key5: "5",
			fyne.Key6: "6",
			fyne.Key7: "7",
			fyne.Key8: "8",
			fyne.Key9: "9",
		}
		if value, ok := numberKeys[ke.Name]; ok {
			// 按下数字键，记录保存为ShipId
			lastShipId += value
		}
		if (ke.Name == fyne.KeyReturn || ke.Name == fyne.KeyEnter) && lastShipId != "" {
			fmt.Println("change shipId to", lastShipId)
			shipId, err := strconv.Atoi(lastShipId)
			lastShipId = ""
			if err != nil {
				fmt.Println(err)
			}
			conf.TrackingNumber.ShipId = shipId
			tracking_number = getTrackingNumber()
			content = updateCanvas(tracking_number)
			myCanvas.SetContent(content)
			fmt.Println(tracking_number)
		}
	})

	window.ShowAndRun()
}

func getShelfNum() string {
	rand.Seed(time.Now().UnixNano())
	leftLength := rand.Intn(2) + 1
	rightLength := rand.Intn(2) + 1
	return getRandNum(leftLength) + "-" + getRandNum(rightLength)
}

func getTrackingNumber() string {
	shipId := conf.TrackingNumber.ShipId
	ship := map[int]string{
		1:   "773" + getRandNum(12), // 申通
		85:  "YT4" + getRandNum(12), // 圆通
		44:  "SF" + getRandNum(13),  // 顺丰
		115: "753" + getRandNum(11), // 中通
		119: "58" + getRandNum(10),  // 天天
		118: "12" + getRandNum(11),  // 邮政EMS
		131: "DPK" + getRandNum(12), // 德邦
		132: "23" + getRandNum(11),  // 邮政快递包裹
		3:   "552" + getRandNum(12), // 百世
		384: "JT" + getRandNum(13),  // 极兔
		340: "JD" + getRandNum(13),  // 京东
	}
	if value, ok := ship[shipId]; ok {
		return value
	}
	return "ShipId ERROR"
}

func getRandNum(l int) string {
	rand.Seed(time.Now().UnixNano())
	num := ""
	for i := 0; i < l; i++ {
		num += strconv.Itoa(rand.Intn(10))
	}
	return num
}

func toImage(tracking_number string) image.Image {
	cs, _ := code128.Encode(tracking_number)
	// 设置图片像素大小
	qrCode, _ := barcode.Scale(cs, 350, 100)
	return qrCode
}

func updateBarcode(cv *canvas.Image) {
	cv = canvas.NewImageFromImage(toImage(getTrackingNumber()))
	cv.Refresh()
}

func updateCanvas(tr_no string) *fyne.Container {
	receiver := widget.NewLabel("")
	if conf.Username.Show && conf.Mobile.Show {
		receiver = widget.NewLabel(conf.Username.Text + " " + conf.Mobile.Text)
	} else if conf.Username.Show {
		receiver = widget.NewLabel(conf.Username.Text)
	} else if conf.Mobile.Show {
		receiver = widget.NewLabel(conf.Mobile.Text)
	}

	image := canvas.NewImageFromImage(toImage(tr_no))
	image.FillMode = canvas.ImageFillOriginal

	tr := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
	if conf.TrackingNumber.Show {
		tr = widget.NewLabelWithStyle(tr_no, fyne.TextAlignCenter, fyne.TextStyle{})
	}
	// cont := container.New(layout.NewVBoxLayout(), receiver, layout.NewSpacer(), image, tr)
	cont := container.New(layout.NewVBoxLayout(), widget.NewLabel(""), image, tr, receiver)
	return cont
}

func contain(s []fyne.KeyName, str fyne.KeyName) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func showHelp() *fyne.Container {
	// data := [][]string{
	// 	{"快捷键", "说明"},
	// 	{"?", "显示/隐藏帮助信息"},
	// 	{"b", "显示/隐藏条形码下对应的运单号"},
	// 	{"n", "显示/隐藏姓名"},
	// 	{"m", "显示/隐藏手机号"},
	// 	{"s", "开启/关闭自动更新条码"},
	// 	{"r/<space>", "使用随机生成的运单号生成条码"},
	// 	{"ctrl+c/cmd+c", "复制当前条码对应的运单号"},
	// 	{"ctrl+v/cmd+v", "使用剪贴板中的运单号生成条码"},
	// 	{"ctrl+r/cmd+r", "重新读取配置文件"},
	// }
	// table := widget.NewTable(
	// 	func() (int, int) {
	// 		return len(data), len(data[0])
	// 	},
	// 	func() fyne.CanvasObject {
	// 		// t := widget.NewLabelWithStyle("wide content", fyne.TextAlignCenter, fyne.TextStyle{})
	// 		t := widget.NewLabel("wide content")
	// 		// t.SetMinSize(fyne.NewSize(200, 300))
	// 		// t.Resize(fyne.NewSize(3000, 800))
	// 		return t
	// 	},
	// 	func(i widget.TableCellID, o fyne.CanvasObject) {
	// 		o.(*widget.Label).SetText(data[i.Row][i.Col])
	// 	})
	title := widget.NewLabelWithStyle("快递单号条形码生成器", fyne.TextAlignCenter, fyne.TextStyle{})
	author := widget.NewLabelWithStyle("author: vhqkze", fyne.TextAlignCenter, fyne.TextStyle{})
	// content_table := container.New(layout.NewMaxLayout(), table, layout.NewSpacer())
	content := container.New(layout.NewVBoxLayout(), title, layout.NewSpacer(), author)
	// content_table.Resize(fyne.NewSize(900, 5000))
	// content_table.MinSize().AddWidthHeight(3000, 4000)
	return content
}

func reload() {
	configFile := config.GetConfigFile()
	conf = config.InitConfig(configFile)
}

func getFont() string {
	if fontPath, err := findfont.Find("Arial Unicode.ttf"); err == nil {
		fmt.Println("found font")
		return fontPath
	}
	return ""
}

// 将字符串截取后4位并转换为数字，加1，然后转为字符后保留后4位
func tracking_number_add_one(tracking_number string, offset int) string {
	tn := strings.TrimSpace(tracking_number)
	tn_len := len(tn)
	tn_sub := tn[tn_len-4:]
	tn_sub_int, _ := strconv.Atoi(tn_sub)
	tn_sub_int = tn_sub_int + offset
	if tn_sub_int > 9999 {
		tn_sub_int = tn_sub_int - 10000
	} else if tn_sub_int < 0 {
		tn_sub_int = tn_sub_int + 10000
	}
	// 将 tn_sub_int 转为长度为4的字符

	// tn_new := strconv.Itoa(tn_sub_int)
	tn_new := fmt.Sprintf("%04d", tn_sub_int)
	tn_new = tn[:tn_len-4] + tn_new
	return tn_new
}
