//go:build abnahme && darwin

package fenster

// **Die Griffe in AppKit** — nur im Abnahmebau (`scripts/bauen.sh abnahme`).
// Was die Leitung ist und warum es sie gibt, steht in `abnahme.go`; hier
// steht, was sie in AppKit tut: die Leitung selbst und die Lage des
// Fensters in dieser Datei, die Leiste in `abnahme_leiste_darwin.go`, Maus,
// Tasten und Ablegen in `abnahme_hand_darwin.go`, die Dateidialoge in
// `abnahme_blatt_darwin.go`. **Jeder Griff läuft auf dem Hauptfaden**
// (`application.InvokeSync…`): AppKit ist nicht fadensicher, und ein Menü,
// das aus einer Goroutine gelesen wird, liest irgendetwas.
//
// **Gelesen in den Kopfdateien des macOS-SDK** (`AppKit.framework/Headers`):
// `NSMenu.h` (`performActionForItemAtIndex:`, `update`), `NSMenuItem.h`,
// `NSApplication.h` (`postEvent:atStart:`, `servicesMenu`), `NSEvent.h`
// (`keyEventWithType:…`), `NSSavePanel.h` und `NSOpenPanel.h`
// (`setDirectoryURL:`, `setNameFieldStringValue:`, `ok:`, `cancel:`),
// `NSWindow.h` (`attachedSheet`, `windowNumber`, `isZoomed`).

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/apple/objc"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Die Vollbildmaske aus `NSWindow.h` — die Arten aus `NSEvent.h` stehen in
// `abnahme.go`, weil die Abfolge eines Zugs ohne AppKit geprüft wird.
const (
	maskeVollbild   uint64 = 1 << 14
	blattBedenkzeit        = 700 * time.Millisecond
)

// punkt und rechteck sind `NSPoint` und `NSRect` — zwei und vier `CGFloat`.
type (
	punkt    struct{ X, Y float64 }
	rechteck struct{ X, Y, B, H float64 }
)

// Fensterlage ist, was AppKit über das Hauptfenster weiß.
type Fensterlage struct {
	Rahmen         rechteck `json:"rahmen"`
	Arbeitsflaeche rechteck `json:"arbeitsflaeche"`
	Gezoomt        bool     `json:"gezoomt"`
	Vollbild       bool     `json:"vollbild"`
	Nummer         int      `json:"nummer"`
	Schluessel     bool     `json:"schluessel"`
	Sichtbar       bool     `json:"sichtbar"`
	Aktiv          bool     `json:"aktiv"`
	Verborgen      bool     `json:"verborgen"`
	Titel          string   `json:"titel"`

	// Vorn ist die Klasse des Schlüsselfensters der Anwendung — leer, wenn
	// sie keines hat. Ist sie vorn, muss es nicht dieses Fenster sein.
	Vorn string `json:"vorn"`

	// Prozess ist die Kennung des Prozesses — damit die Probe ihn beenden
	// kann, wenn ein Drehbuch mittendrin abbricht.
	Prozess int `json:"prozess"`
}

// schluessel ist der Schlüssel des MCP-Servers dieses Laufs — die Leitung
// gibt ihn heraus, und nur sie.
var schluessel string

// leitungsbuch ist das Rufbuch dieses Prozesses — B01 (#2220): die
// Hüllenprobe holt es je Drehbuch über `rufbuch`.
var leitungsbuch = neuesRufbuch()

// rufbuchHuelle legt das Rufbuch um die Hülle des Asset-Servers — als
// äußerste, damit es jeden Bindungsruf sieht (`rufbuch.go`).
func rufbuchHuelle(innen application.Middleware) application.Middleware {
	return func(weiter http.Handler) http.Handler { return leitungsbuch.Huelle(innen(weiter)) }
}

// AbnahmeVorbereiten läuft vor allem anderen (`main.go`): **ohne Leitung
// endet der Abnahmebau**, bevor der Bestand aufgeht und Wails' MCP-Server
// lauscht; mit ihr lauscht der Server nur mit Schlüssel. Wails liest ihn in
// `App.Run` aus der Umgebung (`WAILS_MCP_TOKEN`).
func AbnahmeVorbereiten() error {
	if fehler := abnahmeGrund(os.Getenv("MSE_ABNAHME")); fehler != nil {
		return fehler
	}

	var fehler error
	if schluessel, fehler = mcpSchluessel(rand.Reader); fehler != nil {
		return fehler
	}

	return os.Setenv("WAILS_MCP_TOKEN", schluessel)
}

// abnahmeAnschliessen öffnet die Leitung am Ort aus `MSE_ABNAHME` — ohne ihn
// ist der Abnahmebau gar nicht gestartet ([AbnahmeVorbereiten]). **Und erst,
// wenn die Anwendung steht**: vorher gibt es keinen Hauptfaden, auf den ein
// Griff gehen könnte (`InvokeSync` vor `Run` endet in einem Absturz, gesehen
// am 2026-09-26).
func abnahmeAnschliessen(wails *application.App, fenster *application.WebviewWindow) {
	ort := os.Getenv("MSE_ABNAHME")

	// **Der Ort kommt aus der Umgebung**, und `gosec` nennt das zu Recht einen
	// Weg für einen fremden Pfad (G703). Hier legt ihn die Probe selbst an,
	// in ihrem eigenen Zeitverzeichnis, und es gibt ihn nur im Abnahmebau.
	ort = filepath.Clean(ort)

	wails.Event.OnApplicationEvent(events.Common.ApplicationStarted, gesichert("die Abnahmeleitung", func(*application.ApplicationEvent) {
		// **Nur eine alte Leitung wird geräumt**, nie eine Datei, die dort
		// zufällig liegt — dann scheitert das Öffnen und sagt es.
		//nolint:gosec // G703 steht oben
		if stand, fehler := os.Lstat(ort); fehler == nil && stand.Mode()&os.ModeSocket != 0 {
			_ = os.Remove(ort) //nolint:errcheck,gosec // G703 steht oben
		}

		hoerer, fehler := (&net.ListenConfig{}).Listen(wails.Context(), "unix", ort)
		if fehler != nil {
			slog.Error("die Abnahmeleitung ließ sich nicht öffnen", "grund", fehler.Error())

			return
		}

		if fehler := os.Chmod(ort, 0o600); fehler != nil { //nolint:gosec // G703 steht oben
			slog.Error("die Abnahmeleitung ließ sich nicht sperren", "grund", fehler.Error())
			_ = hoerer.Close() //nolint:errcheck // der Grund steht schon im Protokoll

			return
		}

		leitungOffen.Store(true)

		go annehmen(hoerer, fenster)
	}))
}

func annehmen(hoerer net.Listener, fenster *application.WebviewWindow) {
	for {
		verbindung, fehler := hoerer.Accept()
		if fehler != nil {
			return
		}

		go bedienen(verbindung, fenster)
	}
}

func bedienen(verbindung net.Conn, fenster *application.WebviewWindow) {
	defer func() { _ = verbindung.Close() }() //nolint:errcheck // gelesen ist gelesen

	leser := bufio.NewScanner(verbindung)
	leser.Buffer(make([]byte, 64*1024), 1<<20)
	schreiber := json.NewEncoder(verbindung)

	for leser.Scan() {
		befehl, fehler := befehlLesen(leser.Bytes())

		antwort := antwortAus(nil, fehler)
		if fehler == nil {
			antwort = ausfuehren(befehl, fenster)
		}

		if schreiber.Encode(antwort) != nil {
			return
		}
	}
}

func ausfuehren(befehl Abnahmebefehl, fenster *application.WebviewWindow) Abnahmeantwort {
	switch befehl.Was {
	case "menue":
		return antwortAus(application.InvokeSyncWithResult(echteLeiste), nil)
	case "klick":
		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			return klicken(befehl.Pfad, 0)
		}))
	case "imfenster":
		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			return klicken(befehl.Pfad, fensterID(fenster))
		}))
	case "taste":
		druck, fehler := tasteAus(befehl.Taste, befehl.Mit)
		if fehler == nil {
			application.InvokeSync(func() { druecken(fenster, druck) })
		}

		return antwortAus(nil, fehler)
	case "klicke":
		schritte, fehler := klickschritte(Ort{X: befehl.X, Y: befehl.Y}, befehl.Anzahl, befehl.Mit)
		if fehler != nil {
			return antwortAus(nil, fehler)
		}

		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			return klickeAn(fenster, schritte)
		}))
	case "ziehe":
		schritte, fehler := zugschritte(befehl.Weg, befehl.Halten)
		if fehler != nil {
			return antwortAus(nil, fehler)
		}

		return antwortAus(nil, mausSchicken(fenster, schritte))
	case "loslassen":
		return antwortAus(nil, mausSchicken(fenster,
			[]mausschritt{{art: artMausHoch, ort: Ort{X: befehl.X, Y: befehl.Y}}}))
	case "bewege":
		return antwortAus(nil, mausSchicken(fenster, bewegeschritte(Ort{X: befehl.X, Y: befehl.Y})))
	case "rechtsklick":
		return antwortAus(nil, mausSchicken(fenster, rechtsklickschritte(Ort{X: befehl.X, Y: befehl.Y})))
	case "rufbuch":
		return antwortAus(leitungsbuch.Abzug(befehl.Leeren), nil)
	case "speicher":
		// **Der Go-Heap im Dauerbetrieb** (E06, #2278): Stand und, mit
		// `datei`, ein Heap-Profil — die Probe holt es zu Beginn, in der Mitte
		// und am Ende eines Laufs.
		return antwortAus(speicherabzug(befehl.Datei))
	case "stoere":
		leitungsbuch.Stoeren(befehl.Stoerungen)

		return antwortAus(nil, nil)
	case "aktivieren":
		application.InvokeSync(func() { aktivieren(fenster) })

		return antwortAus(nil, nil)
	case "blatt":
		return antwortAus(application.InvokeSyncWithResult(func() Blattstand {
			return blattLesen(fenster)
		}), nil)
	case "ablegen":
		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			return ablegen(fenster, befehl.Pfad, befehl.X, befehl.Y)
		}))
	case "ueberziehen":
		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			_, _, fehler := ueberziehen(fenster, befehl.Pfad, befehl.X, befehl.Y)

			return fehler
		}))
	case "wegziehen":
		return antwortAus(nil, application.InvokeSyncWithError(wegziehen))
	case "dialogantwort":
		vorbereitet.Lock()
		vorbereitet.bereit, vorbereitet.datei = true, befehl.Datei
		vorbereitet.Unlock()

		return antwortAus(nil, nil)
	case "blattordner":
		return antwortAus(nil, application.InvokeSyncWithError(func() error {
			blatt := blattAn(fenster)
			if blatt == 0 {
				return errKeinBlatt
			}

			objc.Send[struct{}](blatt, objc.Sel("setDirectoryURL:"), dateiadresse(befehl.Datei))

			return nil
		}))
	case "blattantwort":
		return antwortAus(nil, blattBeantworten(fenster, befehl))
	case "fenster":
		return antwortAus(application.InvokeSyncWithResult(func() Fensterlage {
			return lageLesen(fenster)
		}), nil)
	case "zwischenablage":
		return antwortAus(ablage.lesen(), nil)
	case "schluessel":
		return antwortAus(schluessel, nil)
	case "aufklappen":
		return antwortAus(aufklappen(fenster, befehl.Pfad))
	case "beenden":
		application.InvokeAsync(func() {
			objc.Send[struct{}](nsApp(), objc.Sel("terminate:"), objc.ID(0))
		})

		return antwortAus(nil, nil)
	}

	return antwortAus(nil, fmt.Errorf("unbekannter Befehl %q", befehl.Was))
}

func nsApp() objc.ID {
	return objc.Send[objc.ID](objc.ID(objc.GetClass("NSApplication")), objc.Sel("sharedApplication"))
}

func fensterID(fenster *application.WebviewWindow) objc.ID {
	return objc.ID(uintptr(fenster.NativeWindow()))
}

func text(objekt objc.ID, eigenschaft string) string {
	if objekt == 0 {
		return ""
	}

	return objc.IDToString(objc.Send[objc.ID](objekt, objc.Sel(eigenschaft)))
}

func klassenname(objekt objc.ID) string {
	if objekt == 0 {
		return ""
	}

	return text(objekt, "className")
}

func istArt(objekt objc.ID, klasse string) bool {
	return objc.Send[bool](objekt, objc.Sel("isKindOfClass:"), objc.ID(objc.GetClass(klasse)))
}

func lageLesen(fenster *application.WebviewWindow) Fensterlage {
	id := fensterID(fenster)
	app := nsApp()

	return Fensterlage{
		Rahmen: objc.Send[rechteck](id, objc.Sel("frame")),
		Arbeitsflaeche: objc.Send[rechteck](objc.Send[objc.ID](id, objc.Sel("screen")),
			objc.Sel("visibleFrame")),
		Gezoomt:    objc.Send[bool](id, objc.Sel("isZoomed")),
		Vollbild:   objc.Send[uint64](id, objc.Sel("styleMask"))&maskeVollbild != 0,
		Nummer:     objc.Send[int](id, objc.Sel("windowNumber")),
		Schluessel: objc.Send[bool](id, objc.Sel("isKeyWindow")),
		Sichtbar:   objc.Send[bool](id, objc.Sel("isVisible")),
		Aktiv:      objc.Send[bool](app, objc.Sel("isActive")),
		Verborgen:  objc.Send[bool](app, objc.Sel("isHidden")),
		Titel:      text(id, "title"),
		Vorn:       klassenname(objc.Send[objc.ID](app, objc.Sel("keyWindow"))),
		Prozess:    os.Getpid(),
	}
}
