package main

import (
	"embed"
	"encoding/csv"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed Satellite.png iphone_pro.png android_icon.png
var resourceFS embed.FS

var version = "v2.1"
var isPro bool

const maxHistoryPoints = 30

func getPingColor(ms int) color.Color {
	if ms < 40 {
		return color.NRGBA{R: 0, G: 255, B: 255, A: 255}
	}
	if ms < 85 {
		return color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	}
	return color.NRGBA{R: 255, G: 69, B: 0, A: 255}
}

func setEmbeddedImage(img *canvas.Image, fileName string) {
	data, err := resourceFS.ReadFile(fileName)
	if err != nil {
		return
	}
	img.Resource = fyne.NewStaticResource(fileName, data)
}

func creaModulo(title string, imagePath string, colTitle color.Color) (*canvas.Text, *canvas.Rectangle, *canvas.Image, *canvas.Text, *canvas.Text, fyne.CanvasObject) {
	t := canvas.NewText(title, colTitle)
	t.TextStyle.Bold = true

	data, err := resourceFS.ReadFile(imagePath)
	var img *canvas.Image
	if err != nil {
		img = canvas.NewImageFromResource(nil)
	} else {
		img = canvas.NewImageFromResource(fyne.NewStaticResource(imagePath, data))
	}
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(220, 180))

	v := canvas.NewText("---", color.White)
	v.TextSize = 54
	v.TextStyle.Bold = true

	p := canvas.NewText("STATO: ATTESA...", colTitle)
	p.TextSize = 12

	bg := canvas.NewRectangle(color.NRGBA{R: 10, G: 10, B: 15, A: 255})
	bg.StrokeColor = colTitle
	bg.StrokeWidth = 2
	bg.CornerRadius = 12

	content := container.NewVBox(
		container.NewCenter(t),
		layout.NewSpacer(),
		container.NewCenter(img),
		layout.NewSpacer(),
		container.NewCenter(v),
		layout.NewSpacer(),
		container.NewCenter(p),
	)

	return v, bg, img, t, p, container.NewStack(bg, container.NewPadded(content))
}

func formatStarlinkPingText(ms int) string {
	currentCfg := GetConfig()
	if ms >= 999 {
		return "LOST"
	}
	if ms > currentCfg.LagThresholdMs {
		return fmt.Sprintf("%d MS ⚠", ms)
	}
	return fmt.Sprintf("%d MS", ms)
}

func formatDevicePingText(ms int) string {
	currentCfg := GetConfig()
	if ms >= 999 {
		return "OFFLINE"
	}
	if ms == 0 {
		return "<1 MS"
	}
	if ms > currentCfg.LagThresholdMs {
		return fmt.Sprintf("%d MS ⚠", ms)
	}
	return fmt.Sprintf("%d MS", ms)
}

func pingDisplayColor(ms int) color.Color {
	currentCfg := GetConfig()
	if ms >= 999 {
		return color.NRGBA{R: 255, A: 255}
	}
	if ms > currentCfg.LagThresholdMs {
		return color.NRGBA{R: 255, G: 165, B: 0, A: 255}
	}
	return getPingColor(ms)
}

func appendHistory(history []int, value int) []int {
	history = append(history, value)
	if len(history) > maxHistoryPoints {
		history = history[len(history)-maxHistoryPoints:]
	}
	return history
}

func historyStats(history []int) (avg int, max int) {
	if len(history) == 0 {
		return 0, 0
	}

	sum := 0
	max = 0

	for _, v := range history {
		normalized := v
		if normalized >= 999 {
			normalized = 150
		}
		sum += normalized
		if normalized > max {
			max = normalized
		}
	}

	avg = sum / len(history)
	return avg, max
}

func drawPingGraph(starlinkHistory, deviceHistory []int, width, height float32) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 12, G: 12, B: 18, A: 255})
	bg.StrokeColor = color.NRGBA{R: 0, G: 190, B: 255, A: 255}
	bg.StrokeWidth = 1.2
	bg.CornerRadius = 14
	bg.Resize(fyne.NewSize(width, height))

	objects := []fyne.CanvasObject{bg}

	for i := 1; i < 4; i++ {
		y := (height / 4) * float32(i)
		line := canvas.NewLine(color.NRGBA{R: 40, G: 40, B: 55, A: 255})
		line.Position1 = fyne.NewPos(0, y)
		line.Position2 = fyne.NewPos(width, y)
		objects = append(objects, line)
	}

	maxPing := 150.0
	mapY := func(v int) float32 {
		value := float64(v)
		if v >= 999 {
			value = maxPing
		}
		if value > maxPing {
			value = maxPing
		}
		return height - float32((value/maxPing)*float64(height-62)) - 18
	}

	buildLineSet := func(history []int, col color.Color) []fyne.CanvasObject {
		var objs []fyne.CanvasObject
		if len(history) < 2 {
			return objs
		}

		stepX := width / float32(maxHistoryPoints-1)
		for i := 1; i < len(history); i++ {
			x1 := stepX * float32(i-1)
			x2 := stepX * float32(i)
			y1 := mapY(history[i-1])
			y2 := mapY(history[i])

			line := canvas.NewLine(col)
			line.StrokeWidth = 2.5
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			objs = append(objs, line)
		}
		return objs
	}

	objects = append(objects, buildLineSet(starlinkHistory, color.NRGBA{R: 0, G: 255, B: 255, A: 255})...)
	objects = append(objects, buildLineSet(deviceHistory, color.NRGBA{R: 255, G: 180, B: 0, A: 255})...)

	title := canvas.NewText("STORICO PING LIVE", color.White)
	title.TextStyle.Bold = true
	title.TextSize = 18
	title.Move(fyne.NewPos(14, 10))

	legend1 := canvas.NewText("■ STARLINK", color.NRGBA{R: 0, G: 255, B: 255, A: 255})
	legend1.TextSize = 12
	legend1.Move(fyne.NewPos(14, 34))

	legend2 := canvas.NewText("■ DEVICE", color.NRGBA{R: 255, G: 180, B: 0, A: 255})
	legend2.TextSize = 12
	legend2.Move(fyne.NewPos(130, 34))

	scaleTop := canvas.NewText("150", color.NRGBA{R: 120, G: 120, B: 140, A: 255})
	scaleTop.TextSize = 10
	scaleTop.Move(fyne.NewPos(width-34, 8))

	scaleMid := canvas.NewText("75", color.NRGBA{R: 120, G: 120, B: 140, A: 255})
	scaleMid.TextSize = 10
	scaleMid.Move(fyne.NewPos(width-28, height/2-6))

	scaleLow := canvas.NewText("0", color.NRGBA{R: 120, G: 120, B: 140, A: 255})
	scaleLow.TextSize = 10
	scaleLow.Move(fyne.NewPos(width-20, height-22))

	objects = append(objects, title, legend1, legend2, scaleTop, scaleMid, scaleLow)

	return container.NewWithoutLayout(objects...)
}

func saveReportTXT(reportPath string, mode string, uptime time.Duration, alarms int, starlinkHistory, deviceHistory []int, current State) error {
	f, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer f.Close()

	starAvg, starMax := historyStats(starlinkHistory)
	devAvg, devMax := historyStats(deviceHistory)

	_, err = fmt.Fprintf(f,
		"NEURALPATH TACTICAL REPORT\n\n"+
			"Timestamp: %s\n"+
			"Mode: %s\n"+
			"Version: %s\n"+
			"PRO: %t\n"+
			"Uptime: %s\n"+
			"Alarms: %d\n"+
			"Current Device: %s\n"+
			"Current Online: %t\n"+
			"Current Starlink Ping: %d ms\n"+
			"Current Device Ping: %d ms\n"+
			"Starlink AVG: %d ms\n"+
			"Starlink MAX: %d ms\n"+
			"Device AVG: %d ms\n"+
			"Device MAX: %d ms\n",
		time.Now().Format("2006-01-02 15:04:05"),
		mode,
		version,
		isPro,
		uptime.Truncate(time.Second).String(),
		alarms,
		current.Device,
		current.IsOnline,
		current.StarlinkPing,
		current.DevicePing,
		starAvg,
		starMax,
		devAvg,
		devMax,
	)
	return err
}

func saveReportCSV(reportPath string, starlinkHistory, deviceHistory []int) error {
	f, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"index", "starlink_ping_ms", "device_ping_ms"}); err != nil {
		return err
	}

	maxLen := len(starlinkHistory)
	if len(deviceHistory) > maxLen {
		maxLen = len(deviceHistory)
	}

	for i := 0; i < maxLen; i++ {
		s := ""
		d := ""

		if i < len(starlinkHistory) {
			s = strconv.Itoa(starlinkHistory[i])
		}
		if i < len(deviceHistory) {
			d = strconv.Itoa(deviceHistory[i])
		}

		if err := w.Write([]string{strconv.Itoa(i), s, d}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	loadedCfg := LoadConfig()

	initLogger()
	defer closeLogger()

	myApp := app.NewWithID("com.neuralpath.tacticalguard")

	licenseStatus := ResolveLicenseStatus(&loadedCfg)
	accessExpired := licenseStatus.Mode == LicenseModeExpired || licenseStatus.Mode == LicenseModeInvalid

	isPro = licenseStatus.IsPro

	canUseRealMode := licenseStatus.IsPro || licenseStatus.IsTrial
	canUseReports := licenseStatus.IsPro
	canUseNotifications := licenseStatus.IsPro || licenseStatus.IsTrial

	var licenseColor color.Color
	licenseLabel := "PIANO: FREE"

	switch licenseStatus.Mode {
	case LicenseModePro:
		licenseColor = color.NRGBA{R: 0, G: 255, B: 120, A: 255}
		licenseLabel = "PIANO: PRO"
	case LicenseModeTrial:
		licenseColor = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
		licenseLabel = "PIANO: TRIAL"
	default:
		licenseColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
		licenseLabel = "PIANO: FREE"
	}

	if accessExpired {
		licenseLabel = "PIANO: FREE (TRIAL SCADUTO)"
	}

	trialInfoLabel := ""
	if licenseStatus.IsTrial && !accessExpired {
		trialInfoLabel = fmt.Sprintf("PROVA ATTIVA • RESTANO %d GIORNI", licenseStatus.DaysLeft)
	}
	if accessExpired {
		trialInfoLabel = "PROVA SCADUTA • ATTIVA PRO PER SBLOCCARE RETE REALE"
	}
	if licenseStatus.IsPro {
		trialInfoLabel = "LICENZA PRO ATTIVA"
	}

	trialInfoColor := color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	if accessExpired {
		trialInfoColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
	}
	if licenseStatus.IsPro {
		trialInfoColor = color.NRGBA{R: 0, G: 255, B: 120, A: 255}
	}

	trialInfoColor = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	if accessExpired {
		trialInfoColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
	}
	trialInfoColor = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	if accessExpired {
		trialInfoColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
	}
	if licenseStatus.IsPro {
		trialInfoColor = color.NRGBA{R: 0, G: 255, B: 120, A: 255}
	}

	myWindow := myApp.NewWindow(loadedCfg.AppTitle + " " + version)

	if iconData, err := resourceFS.ReadFile("Satellite.png"); err == nil {
		appIcon := fyne.NewStaticResource("Satellite.png", iconData)
		myApp.SetIcon(appIcon)
		myWindow.SetIcon(appIcon)
	}

	startTime := time.Now()
	countAllarmi := 0

	if accessExpired {
		dialog.ShowInformation(
			"Trial scaduta",
			"La prova gratuita è terminata.\nPassa alla versione PRO per riattivare le funzioni premium.",
			myWindow,
		)
	}

	mock := &MockNetwork{}
	mock.SetState(true, "IPHONE", 29, 12)
	realNet := RealNetwork{}

	testMode := true
	var netImpl Network = mock

	colSl := color.NRGBA{R: 0, G: 255, B: 255, A: 255}
	colEme := color.NRGBA{R: 255, G: 165, B: 0, A: 255}

	vSl, _, _, _, pSl, modSl := creaModulo("STARLINK_PRIMARY", "Satellite.png", colSl)
	vEme, bgEme, imgEme, titEme, pEme, modEme := creaModulo("RICERCA BACKUP...", "iphone_pro.png", colEme)

	healthText := canvas.NewText("SISTEMA: MONITORAGGIO ATTIVO", color.White)
	healthText.TextStyle.Bold = true
	healthText.TextSize = 22

	modeText := canvas.NewText("MODALITÀ TEST ATTIVA", color.NRGBA{R: 255, G: 100, B: 100, A: 255})
	modeText.TextStyle.Bold = true

	alarmLogText := canvas.NewText("ALLARMI SESSIONE: 0", color.NRGBA{R: 255, G: 80, B: 80, A: 255})
	alarmLogText.TextStyle.Bold = true

	statsText := canvas.NewText("STARLINK AVG: 0 | MAX: 0    DEVICE AVG: 0 | MAX: 0", color.White)
	statsText.TextStyle.Bold = true
	statsText.TextStyle.Monospace = true
	statsText.TextSize = 13

	alertText := canvas.NewText("NESSUN ALERT", color.NRGBA{R: 120, G: 220, B: 120, A: 255})
	alertText.TextStyle.Bold = true
	alertText.TextSize = 14

	licenseText := canvas.NewText(licenseLabel, licenseColor)
	licenseText.TextStyle.Bold = true
	licenseText.TextSize = 13

	trialInfoText := canvas.NewText(trialInfoLabel, trialInfoColor)
	trialInfoText.TextStyle.Bold = true
	trialInfoText.TextSize = 12

	reportStatusText := canvas.NewText("REPORT: NON ANCORA SALVATO", color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	reportStatusText.TextStyle.Monospace = true
	reportStatusText.TextSize = 12

	creditsText := canvas.NewText("© 2026 NeuralPath Tactical — Lead Dev: Josh Fratocchi", color.NRGBA{R: 150, G: 150, B: 150, A: 255})
	creditsText.TextSize = 10

	uptimeText := canvas.NewText("UPTIME: 00:00:00", color.White)
	uptimeText.TextStyle.Monospace = true

	var starlinkHistory []int
	var deviceHistory []int
	var lastState State
	var sessionMu sync.RWMutex

	copyHistory := func(src []int) []int {
		if len(src) == 0 {
			return nil
		}
		dst := make([]int, len(src))
		copy(dst, src)
		return dst
	}

	getSessionSnapshot := func() (time.Time, int, []int, []int, State) {
		sessionMu.RLock()
		defer sessionMu.RUnlock()

		return startTime, countAllarmi, copyHistory(starlinkHistory), copyHistory(deviceHistory), lastState
	}

	resetSessionData := func() {
		sessionMu.Lock()
		defer sessionMu.Unlock()

		startTime = time.Now()
		countAllarmi = 0
		lastState = State{}
		starlinkHistory = nil
		deviceHistory = nil
	}

	var lastOfflineNotification time.Time
	var lastLagNotification time.Time

	graphWidth := float32(800)
	graphHeight := float32(220)

	graphBackground := canvas.NewRectangle(color.Transparent)
	graphBackground.SetMinSize(fyne.NewSize(graphWidth, graphHeight))

	graphHolder := container.NewMax(
		graphBackground,
		drawPingGraph(starlinkHistory, deviceHistory, graphWidth, graphHeight),
	)

	refreshGraph := func() {
		_, _, starlinkSnapshot, deviceSnapshot, _ := getSessionSnapshot()
		graphHolder.Objects = []fyne.CanvasObject{
			graphBackground,
			drawPingGraph(starlinkSnapshot, deviceSnapshot, graphWidth, graphHeight),
		}
		graphHolder.Refresh()
	}

	refreshModeUI := func() {
		if testMode {
			modeText.Text = "MODALITÀ TEST ATTIVA"
			modeText.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
		} else {
			modeText.Text = "MODALITÀ RETE REALE"
			modeText.Color = color.NRGBA{R: 100, G: 200, B: 255, A: 255}
		}
		modeText.Refresh()
	}
	var switchModeBtn *widget.Button
	var saveReportBtn *widget.Button
	var activateProBtn *widget.Button
	var deactivateLicenseBtn *widget.Button

	resetBtn := widget.NewButton("RESET SESSION DATA", func() {
		resetSessionData()
		resetLogicState()
		refreshGraph()
		statsText.Text = "STARLINK AVG: 0 | MAX: 0    DEVICE AVG: 0 | MAX: 0"
		alertText.Text = "NESSUN ALERT"
		alertText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
		reportStatusText.Text = "REPORT: NON ANCORA SALVATO"
		reportStatusText.Color = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
		alarmLogText.Text = "ALLARMI SESSIONE: 0"
		alarmLogText.Refresh()
		reportStatusText.Refresh()
		logEvent("RESET SESSIONE")
	})

	btnOnlineIphone := widget.NewButton("ONLINE IPHONE", func() {
		if testMode {
			mock.SetState(true, "IPHONE", 29, 12)
			logEvent("TEST: ONLINE IPHONE | Starlink: 29 ms | Device: 12 ms")
		}
	})

	btnOnlineAndroid := widget.NewButton("ONLINE ANDROID", func() {
		if testMode {
			mock.SetState(true, "ANDROID", 28, 18)
			logEvent("TEST: ONLINE ANDROID | Starlink: 28 ms | Device: 18 ms")
		}
	})

	btnLag := widget.NewButton("LAG", func() {
		if testMode {
			mock.SetState(true, "IPHONE", 120, 85)
			logEvent("TEST: LAG | Starlink: 120 ms | Device: 85 ms")
		}
	})

	btnOffline := widget.NewButton("OFFLINE", func() {
		if testMode {
			mock.SetState(false, "NONE", 999, 999)
			logEvent("TEST: OFFLINE")
		}
	})

	testButtons := container.NewGridWithColumns(
		4,
		btnOnlineIphone,
		btnOnlineAndroid,
		btnLag,
		btnOffline,
	)

	updateButtonsState := func() {
		if testMode {
			btnOnlineIphone.Enable()
			btnOnlineAndroid.Enable()
			btnLag.Enable()
			btnOffline.Enable()
		} else {
			btnOnlineIphone.Disable()
			btnOnlineAndroid.Disable()
			btnLag.Disable()
			btnOffline.Disable()
		}
	}

	applyLicenseUI := func() {
		currentCfg := GetConfig()
		licenseStatus = ResolveLicenseStatus(&currentCfg)
		accessExpired = licenseStatus.Mode == LicenseModeExpired || licenseStatus.Mode == LicenseModeInvalid

		isPro = licenseStatus.IsPro
		canUseRealMode = licenseStatus.IsPro || licenseStatus.IsTrial
		canUseReports = licenseStatus.IsPro
		canUseNotifications = licenseStatus.IsPro || licenseStatus.IsTrial

		switch licenseStatus.Mode {
		case LicenseModePro:
			licenseText.Text = "PIANO: PRO"
			licenseText.Color = color.NRGBA{R: 0, G: 255, B: 120, A: 255}
			trialInfoText.Text = "LICENZA PRO ATTIVA"
			trialInfoText.Color = color.NRGBA{R: 0, G: 255, B: 120, A: 255}

		case LicenseModeTrial:
			licenseText.Text = "PIANO: TRIAL"
			licenseText.Color = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
			trialInfoText.Text = fmt.Sprintf("PROVA ATTIVA • RESTANO %d GIORNI", licenseStatus.DaysLeft)
			trialInfoText.Color = color.NRGBA{R: 255, G: 215, B: 0, A: 255}

		default:
			licenseText.Text = "PIANO: FREE (TRIAL SCADUTO)"
			licenseText.Color = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
			trialInfoText.Text = "PROVA SCADUTA • ATTIVA PRO PER SBLOCCARE RETE REALE"
			trialInfoText.Color = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
		}

		if saveReportBtn != nil {
			if canUseReports {
				saveReportBtn.Enable()
				saveReportBtn.SetText("SALVA REPORT")
			} else {
				saveReportBtn.Disable()
				saveReportBtn.SetText("🔒 REPORT AVANZATI SOLO PRO")
			}
			saveReportBtn.Refresh()
		}

		if activateProBtn != nil {
			if isPro {
				activateProBtn.Hide()
			} else {
				activateProBtn.Show()
			}
			activateProBtn.Refresh()
		}

		if deactivateLicenseBtn != nil {
			if isPro {
				deactivateLicenseBtn.Show()
			} else {
				deactivateLicenseBtn.Hide()
			}
			deactivateLicenseBtn.Refresh()
		}

		if !canUseRealMode && !testMode {
			testMode = true
			mock.SetState(true, "IPHONE", 29, 12)
			netImpl = mock
			resetLogicState()
			resetSessionData()
			refreshGraph()
			statsText.Text = "STARLINK AVG: 0 | MAX: 0    DEVICE AVG: 0 | MAX: 0"
			alertText.Text = "NESSUN ALERT"
			alertText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
			reportStatusText.Text = "REPORT: NON ANCORA SALVATO"
			reportStatusText.Color = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
			alarmLogText.Text = "ALLARMI SESSIONE: 0"

			if switchModeBtn != nil {
				switchModeBtn.SetText("PASSA A RETE REALE")
			}
		}

		refreshModeUI()
		updateButtonsState()

		licenseText.Refresh()
		trialInfoText.Refresh()
		reportStatusText.Refresh()
		alarmLogText.Refresh()
		statsText.Refresh()
		alertText.Refresh()
	}

	activateProBtn = widget.NewButton("ATTIVA PRO", func() {
		keyEntry := widget.NewEntry()
		keyEntry.SetPlaceHolder("NP-PRO-XXXX-XXXX-XXXX-XXXX-XXXX")

		content := container.NewVBox(
			widget.NewLabel("Inserisci la chiave licenza PRO locale:"),
			keyEntry,
		)

		dialog.ShowCustomConfirm(
			"Attiva licenza PRO",
			"Attiva",
			"Annulla",
			content,
			func(ok bool) {
				if !ok {
					return
				}

				err := WithConfigWrite(func(c *AppConfig) error {
					_, activateErr := ActivateLicense(c, keyEntry.Text)
					return activateErr
				})
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				applyLicenseUI()

				dialog.ShowInformation(
					"Licenza attivata",
					"Licenza PRO attivata con successo.",
					myWindow,
				)
			},
			myWindow,
		)
	})

	deactivateLicenseBtn = widget.NewButton("DISATTIVA LICENZA", func() {
		dialog.ShowConfirm(
			"Disattiva licenza",
			"Vuoi rimuovere la licenza PRO da questa installazione?",
			func(ok bool) {
				if !ok {
					return
				}

				wasRealMode := !testMode

				err := WithConfigWrite(func(c *AppConfig) error {
					return ClearLicense(c)
				})
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				applyLicenseUI()

				if !canUseRealMode && wasRealMode {
					dialog.ShowInformation(
						"Licenza disattivata",
						"Licenza rimossa. Il trial è scaduto, quindi l'app è tornata automaticamente alla modalità TEST.",
						myWindow,
					)
					return
				}

				dialog.ShowInformation(
					"Licenza disattivata",
					"Licenza rimossa con successo.",
					myWindow,
				)
			},
			myWindow,
		)
	})

	saveReportBtn = widget.NewButton("SALVA REPORT", func() {
		currentCfg := GetConfig()
		startSnapshot, alarmsSnapshot, starlinkSnapshot, deviceSnapshot, stateSnapshot := getSessionSnapshot()
		mode := "TEST"
		if !testMode {
			mode = "RETE REALE"
		}

		timestamp := time.Now().Format("2006-01-02_15-04-05")
		txtPath := filepath.Join(currentCfg.ReportsDir, "report_"+timestamp+".txt")
		csvPath := filepath.Join(currentCfg.ReportsDir, "report_"+timestamp+".csv")

		errTxt := saveReportTXT(
			txtPath,
			mode,
			time.Since(startSnapshot),
			alarmsSnapshot,
			starlinkSnapshot,
			deviceSnapshot,
			stateSnapshot,
		)
		errCSV := saveReportCSV(csvPath, starlinkSnapshot, deviceSnapshot)

		if errTxt == nil && errCSV == nil {
			reportStatusText.Text = "REPORT SALVATO"
			reportStatusText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
			logEvent("REPORT SALVATO: " + txtPath + " | " + csvPath)

			if canUseNotifications {
				myApp.SendNotification(&fyne.Notification{
					Title:   "NeuralPath Tactical Guard",
					Content: "Report salvato con successo.",
				})
			}
		} else {
			reportStatusText.Text = "ERRORE SALVATAGGIO REPORT"
			reportStatusText.Color = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
			logEvent(fmt.Sprintf("ERRORE REPORT: txt=%v csv=%v", errTxt, errCSV))
		}

		reportStatusText.Refresh()
	})

	switchModeBtn = widget.NewButton("PASSA A RETE REALE", func() {
		if !canUseRealMode {
			dialog.ShowInformation(
				"Sblocca la modalità reale",
				"La modalità rete reale è disponibile solo durante la prova gratuita o con licenza PRO.\n\nAttiva PRO per continuare a monitorare la rete reale senza limiti.",
				myWindow,
			)
			return
		}

		if testMode {
			testMode = false
			netImpl = realNet
			resetLogicState()
			resetSessionData()
			refreshGraph()
			statsText.Text = "STARLINK AVG: 0 | MAX: 0    DEVICE AVG: 0 | MAX: 0"
			alertText.Text = "NESSUN ALERT"
			alertText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
			reportStatusText.Text = "REPORT: NON ANCORA SALVATO"
			reportStatusText.Color = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
			alarmLogText.Text = "ALLARMI SESSIONE: 0"
			switchModeBtn.SetText("PASSA A MODALITÀ TEST")
			logEvent("Modalità cambiata: RETE REALE")
		} else {
			testMode = true
			mock.SetState(true, "IPHONE", 29, 12)
			netImpl = mock
			resetLogicState()
			resetSessionData()
			refreshGraph()
			statsText.Text = "STARLINK AVG: 0 | MAX: 0    DEVICE AVG: 0 | MAX: 0"
			alertText.Text = "NESSUN ALERT"
			alertText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
			reportStatusText.Text = "REPORT: NON ANCORA SALVATO"
			reportStatusText.Color = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
			alarmLogText.Text = "ALLARMI SESSIONE: 0"
			switchModeBtn.SetText("PASSA A RETE REALE")
			logEvent("Modalità cambiata: TEST")
		}

		refreshModeUI()
		updateButtonsState()
		alarmLogText.Refresh()
		reportStatusText.Refresh()
	})

	refreshModeUI()
	updateButtonsState()
	applyLicenseUI()

	statsBox := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(statsText),
		container.NewCenter(alertText),
		container.NewCenter(licenseText),
		container.NewCenter(trialInfoText),
		container.NewCenter(reportStatusText),
		layout.NewSpacer(),
	)

	statsCardBg := canvas.NewRectangle(color.NRGBA{R: 15, G: 15, B: 20, A: 255})
	statsCardBg.StrokeColor = color.NRGBA{R: 0, G: 190, B: 255, A: 255}
	statsCardBg.StrokeWidth = 2
	statsCardBg.CornerRadius = 12

	statsCard := container.NewStack(
		statsCardBg,
		container.NewPadded(statsBox),
	)
	statsCard.Resize(fyne.NewSize(graphWidth+40, 130))

	bottomContent := container.NewVBox(
		container.NewCenter(healthText),
		container.NewCenter(modeText),
		container.NewCenter(alarmLogText),
		container.NewCenter(uptimeText),

		layout.NewSpacer(),

		container.NewCenter(container.NewPadded(graphHolder)),

		layout.NewSpacer(),

		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				statsCard,
				layout.NewSpacer(),
			),
		),

		layout.NewSpacer(),
		container.NewCenter(testButtons),
		container.NewCenter(switchModeBtn),
		container.NewCenter(saveReportBtn),
		container.NewCenter(activateProBtn),
		container.NewCenter(deactivateLicenseBtn),
		container.NewCenter(resetBtn),

		layout.NewSpacer(),
		container.NewCenter(creditsText),
	)

	myWindow.SetContent(container.NewBorder(
		nil,
		container.NewPadded(bottomContent),
		nil,
		nil,
		container.NewPadded(container.NewGridWithColumns(2, modSl, modEme)),
	))

	var prevStateKnown bool
	var prevOnline bool
	var prevDevice string
	var prevLag bool

	go func() {
		for {
			currentCfg := GetConfig()
			_, prevAlarms, _, _, _ := getSessionSnapshot()
			state := updateLogic(netImpl, prevAlarms)

			currentLag := state.StarlinkPing > currentCfg.LagThresholdMs || (state.IsOnline && state.DevicePing > currentCfg.LagThresholdMs)

			if !prevStateKnown {
				if state.IsOnline {
					logEvent(fmt.Sprintf("ONLINE - Device: %s | Starlink: %d ms | Device: %d ms",
						state.Device, state.StarlinkPing, state.DevicePing))
				} else {
					logEvent("OFFLINE - Nessun dispositivo connesso")
				}
				if currentLag {
					logEvent("⚠️ LAG RILEVATO")
				}
				prevStateKnown = true
				prevOnline = state.IsOnline
				prevDevice = state.Device
				prevLag = currentLag
			} else {
				if state.IsOnline != prevOnline || state.Device != prevDevice {
					if state.IsOnline {
						logEvent(fmt.Sprintf("ONLINE - Device: %s | Starlink: %d ms | Device: %d ms",
							state.Device, state.StarlinkPing, state.DevicePing))
					} else {
						logEvent("OFFLINE - Nessun dispositivo connesso")
					}
					prevOnline = state.IsOnline
					prevDevice = state.Device
				}

				if currentLag != prevLag {
					if currentLag {
						logEvent("⚠️ LAG RILEVATO")
					} else {
						logEvent("✅ LAG RISOLTO")
					}
					prevLag = currentLag
				}
			}

			if canUseNotifications && !state.IsOnline && time.Since(lastOfflineNotification) > 20*time.Second {
				myApp.SendNotification(&fyne.Notification{
					Title:   "NeuralPath Tactical Guard",
					Content: "Device offline rilevato.",
				})
				lastOfflineNotification = time.Now()
			}

			if canUseNotifications && currentLag && time.Since(lastLagNotification) > 20*time.Second {
				myApp.SendNotification(&fyne.Notification{
					Title:   "NeuralPath Tactical Guard",
					Content: "Lag rilevato sulla rete.",
				})
				lastLagNotification = time.Now()
			}

			var elapsed time.Duration
			var alarmsSnapshot int
			var starAvg, starMax, devAvg, devMax int
			sessionMu.Lock()
			lastState = state
			countAllarmi = state.AlarmCount
			elapsed = time.Since(startTime)
			starlinkHistory = appendHistory(starlinkHistory, state.StarlinkPing)
			if state.IsOnline {
				deviceHistory = appendHistory(deviceHistory, state.DevicePing)
			} else {
				deviceHistory = appendHistory(deviceHistory, 999)
			}
			alarmsSnapshot = countAllarmi
			starAvg, starMax = historyStats(starlinkHistory)
			devAvg, devMax = historyStats(deviceHistory)
			sessionMu.Unlock()

			fyne.Do(func() {
				updateButtonsState()

				alarmLogText.Text = fmt.Sprintf("ALLARMI SESSIONE: %d", alarmsSnapshot)

				uptimeText.Text = fmt.Sprintf(
					"UPTIME: %02d:%02d:%02d",
					int(elapsed.Hours()),
					int(elapsed.Minutes())%60,
					int(elapsed.Seconds())%60,
				)

				vSl.Text = formatStarlinkPingText(state.StarlinkPing)
				vSl.Color = pingDisplayColor(state.StarlinkPing)
				pSl.Text = "STATO: OPERATIVO"

				if state.IsOnline {
					vEme.Text = formatDevicePingText(state.DevicePing)
					vEme.Color = pingDisplayColor(state.DevicePing)
					pEme.Text = "STATO: TELEFONO CONNESSO"

					if state.Device == "IPHONE" {
						titEme.Text = "IPHONE_HOTSPOT"
						setEmbeddedImage(imgEme, "iphone_pro.png")
					} else {
						titEme.Text = "ANDROID_HOTSPOT"
						setEmbeddedImage(imgEme, "android_icon.png")
					}

					if currentLag {
						healthText.Text = "SISTEMA: LAG RILEVATO"
						healthText.Color = color.NRGBA{R: 255, G: 165, B: 0, A: 255}
						alertText.Text = "⚠️ LAG ATTIVO"
						alertText.Color = color.NRGBA{R: 255, G: 165, B: 0, A: 255}
					} else {
						healthText.Text = "SISTEMA: BATTLE READY"
						healthText.Color = color.NRGBA{G: 255, A: 255}
						alertText.Text = "NESSUN ALERT"
						alertText.Color = color.NRGBA{R: 120, G: 220, B: 120, A: 255}
					}

					bgEme.StrokeColor = colEme
				} else {
					vEme.Text = "OFFLINE"
					vEme.Color = color.White
					titEme.Text = "BACKUP SCOLLEGATO"
					pEme.Text = "STATO: NESSUN TELEFONO"
					healthText.Text = "SISTEMA: RISCHIO LAG"
					healthText.Color = color.NRGBA{R: 255, A: 255}
					alertText.Text = "🚨 DEVICE OFFLINE"
					alertText.Color = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
					bgEme.StrokeColor = color.NRGBA{R: 255, A: 255}
				}

				statsText.Text = fmt.Sprintf(
					"STARLINK AVG: %d | MAX: %d    DEVICE AVG: %d | MAX: %d",
					starAvg, starMax, devAvg, devMax,
				)

				refreshGraph()

				vSl.Refresh()
				vEme.Refresh()
				pSl.Refresh()
				pEme.Refresh()
				imgEme.Refresh()
				titEme.Refresh()
				bgEme.Refresh()
				uptimeText.Refresh()
				healthText.Refresh()
				alarmLogText.Refresh()
				modeText.Refresh()
				statsText.Refresh()
				alertText.Refresh()
				reportStatusText.Refresh()
				licenseText.Refresh()
				trialInfoText.Refresh()
			})

			time.Sleep(time.Duration(currentCfg.RefreshIntervalMs) * time.Millisecond)
		}
	}()

	myWindow.Resize(fyne.NewSize(1000, 1380))
	myWindow.ShowAndRun()
}
