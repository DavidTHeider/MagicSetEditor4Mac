// Package main ist der Einstiegspunkt der Anwendung.
//
// Hier steht nur die Verdrahtung: Protokoll, Bestand, eingebettetes Frontend.
// Jede Entscheidung, die sich prüfen lässt, gehört nach `internal/app`; wie
// sie das Fenster erreicht, steht in `internal/fenster` — was hier steht,
// sieht kein Test.
//
// In `.covignore` ausgenommen, und zwar mit dem Grund, den diese Datei selbst
// belegt: `fenster.Lauf` öffnet ein echtes Fenster und kehrt erst zurück, wenn
// es geschlossen wird.
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fenster"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcpnaehte"
)

// Das gebaute Frontend wird in die Binärdatei eingebettet.
//
// `all:` nimmt auch Dateien mit führendem Punkt mit — ohne das Präfix ließe
// `go:embed` sie aus, und Vite legt welche an.
//
//go:embed all:frontend/dist
var eingebettet embed.FS

func main() {
	// **Ein Abnahmebau startet nur aus der Hüllenprobe** — sonst lauschte
	// Wails' MCP-Server, und der Bestand wäre der echte. Im gewöhnlichen Bau
	// tut der Ruf nichts (`internal/fenster/abnahme.go`).
	if fehler := fenster.AbnahmeVorbereiten(); fehler != nil {
		log.Fatalf("%v", fehler)
	}

	if fehler := lauf(); fehler != nil {
		// **Der Fehler bekommt ein Fenster** (#797).
		//
		// `log.Fatalf` schreibt nach stderr, und eine doppelgeklickte `.app`
		// hat keins, das jemand sieht: das Symbol hüpfte einmal und war weg.
		// `app.Startfehler` zeigt ihn, **bevor** hier abgebrochen wird — dann
		// steht er in beiden Kanälen, und der zweite erreicht auch jemanden.
		app.Startfehler(fehler)

		log.Fatalf("%v", fehler)
	}
}

// protokollweiche gibt die Weiche der Anwendung — mit der Vorgabe, bis der
// Bestand offen ist und seine Einstellungen gelten (OPS-04, #1721).
//
// **Ein Protokoll, das sich nicht anlegen lässt, hält die Anwendung nicht
// auf.** Wer den Bestand öffnen kann, soll Karten bauen können, auch wenn das
// Verzeichnis daneben nicht beschreibbar ist. **Und es nimmt die anderen
// Kanäle nicht mit** (AK-02): bis OPS-04 stand hier ein Schreiber, der dann
// ganz schwieg — auch ins Unified Log. Die Weiche baut jeden Kanal für sich,
// und die Diagnose sagt, welcher fehlt.
func protokollweiche() *app.Protokollweiche {
	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		log.Printf("Der Ort des Protokolls ließ sich nicht bestimmen: %v", fehler)

		return app.NeueProtokollweiche("")
	}

	return app.NeueProtokollweiche(filepath.Dir(pfad))
}

// lauf hält den ganzen Ablauf, damit `defer` etwas zu tun bekommt.
//
// `log.Fatalf` ruft `os.Exit`, und `os.Exit` überspringt jedes zurückgestellte
// Kommando. Stünde der Fehlerabbruch weiter unten neben dem `defer`, das den
// Bestand schließt, liefe genau der Fall ins Leere, für den das `defer` da ist.
// Aufgefallen ist das dem Linter (`exitAfterDefer`), nicht dem Lesen.
func lauf() error {
	// Der Bestand wird **vor** dem Fenster geöffnet und migriert.
	//
	// Eine Anwendung, die ihre Datenbank erst beim ersten Zugriff anlegt,
	// meldet einen Fehler darin irgendwo mitten in der Bedienung — hier meldet
	// sie ihn beim Start.
	// **Und `MSE_BESTAND` gilt auch hier** — der Befund vom 2026-09-14.
	//
	// Hier stand `app.Datenbankpfad()`, und damit las **nur** das Fenster den
	// Umgebungsschalter nicht: `mse`, `mse-mcp` und `mse-rest` gehen seit
	// jeher über [app.BestandsortFuerFlaechen]. Das Fenster ist auch eine
	// Fläche, und der Name der Funktion sagte es die ganze Zeit.
	pfad, fehler := app.BestandsortFuerFlaechen()
	if fehler != nil {
		return fmt.Errorf("der Ort der Datenbank ließ sich nicht bestimmen: %w", fehler)
	}

	// **Das Protokoll steht vor dem Bestand** (#848). Was beim Öffnen
	// schiefgeht, ist genau das, wovon man später nur im Protokoll erfährt.
	//
	// **Scheitert es selbst, läuft die Anwendung trotzdem.** Ein Programm,
	// das nicht startet, weil es nicht protokollieren kann, tauscht einen
	// unbequemen Fehler gegen einen unbenutzbaren Zustand.
	weiche := protokollweiche()
	defer func() {
		if schliessfehler := weiche.Schliessen(); schliessfehler != nil {
			log.Printf("Das Protokoll ließ sich nicht schließen: %v", schliessfehler)
		}
	}()

	schreiber := weiche.Logger()

	// **Und er wird der Standard** (#857). `app.Uebersetze` schreibt den
	// unbekannten Fehler auf, bevor es ihn durch den Rückfallsatz ersetzt —
	// und es hat keinen Empfänger, durch den man ein Protokoll reichen
	// könnte. Ohne diese Zeile ginge diese eine Zeile nach stderr, das eine
	// doppelgeklickte `.app` nicht hat.
	slog.SetDefault(schreiber)

	// **Was unter der alten Kennung liegt, zieht um** (R02, #2334, Vermerk
	// 0191) — vor dem Fenster, denn WebKit liest seinen Speicher, sobald es
	// steht. Scheitert ein Teil, startet die Anwendung trotzdem.
	if umzugsfehler := app.KennungUmziehen(context.Background()); umzugsfehler != nil {
		schreiber.Warn("der Umzug von der alten Kennung blieb unvollständig", "grund", umzugsfehler.Error())
	}

	// **Die Kommandozentrale bekommt die Werkzeuge des MCP-Servers** —
	// Werkstück U12 von Phase 5.1 (#1423): dieselben Nähte wie `mse-mcp`.
	anwendung := app.NeueApp(pfad).MitBefehlsquelle(mcpnaehte.Befehle).MitProtokollweiche(weiche)

	// **Die Startmarke vor allem anderen** (F04, #2343): liegt die einer
	// früheren Sitzung noch, endete sie nicht sauber, und das Fenster sagt
	// es. Scheitert das Legen, fehlt nur der Hinweis beim nächsten Mal.
	if markenfehler := anwendung.StartmarkeSetzen(filepath.Join(filepath.Dir(pfad), "Sitzungen")); markenfehler != nil {
		schreiber.Warn("die Startmarke ließ sich nicht legen", "grund", markenfehler.Error())
	}

	// **Und der Pfad von der Kommandozeile** (#891).
	//
	// Er ist **nicht** der Weg des Finders — der geht über
	// `ApplicationOpenedWithFile` in `internal/fenster` — aber
	// der eines Skripts und der einer Shell. `DateiOeffnen` hebt ihn auf, bis
	// das Fenster steht.
	if uebergeben := uebergebeneDatei(os.Args); uebergeben != "" {
		anwendung.DateiOeffnen(uebergeben)
	}

	// **Ein gescheiterter Start beendet die Anwendung nicht mehr** (#962).
	//
	// Hier stand bis zum 2026-08-25 ein `return`. Es gab dann kein Fenster,
	// keine Bindung und keinen Knopf — der Rückweg über eine Sicherung war
	// im gescheiterten Fall **nicht erreichbar**, und genau das ist der Fall,
	// für den es ihn gibt. Eine Anwendung, die sich bei einem Datenbankfehler
	// wortlos beendet, lässt die Menschen allein, die eine Sicherung
	// bräuchten.
	//
	// Das Fenster öffnet trotzdem und fragt `Startbefund`; ist der Bestand
	// nicht offen, zeigt es den Grund und den Weg zur Sicherung.
	if fehler = anwendung.Starten(context.Background()); fehler != nil {
		schreiber.Error("der Bestand ließ sich nicht öffnen — das Fenster öffnet zur Wiederherstellung",
			"grund", fehler.Error())
	} else {
		schreiber.Info("der Bestand ist offen", "ort", pfad)
	}
	defer func() {
		if schliessfehler := anwendung.Beenden(); schliessfehler != nil {
			schreiber.Error("der Bestand ließ sich nicht schließen",
				"grund", schliessfehler.Error())
			log.Printf("Der Bestand ließ sich nicht schließen: %v", schliessfehler)
		}
	}()

	// **Das eingebettete Frontend ohne seinen Ordner** — Wails 2 nahm den
	// Unterordner selbst aus `frontend/dist`; Wails 3 bekommt das Verzeichnis,
	// aus dem es liefert. Scheitern kann `fs.Sub` hier nicht: der Name ist
	// fest und gültig.
	oberflaeche, _ := fs.Sub(eingebettet, "frontend/dist") //nolint:errcheck // siehe darüber

	// **Wie das Fenster aufgeht** — UX-14.AK-01
	// (#1686).
	// Gelesen wird hier, entschieden in `internal/app`, übersetzt in
	// `internal/fenster`. Ohne offenen Bestand kommt die Vorgabe zurück —
	// maximiert, und das ist der Fall, in dem das Fenster zur Wiederherstellung
	// öffnet (#962).
	return fenster.Lauf(anwendung, fenster.Optionen{
		Oberflaeche: oberflaeche,
		Bildweg:     app.Bildweg(anwendung),
		Protokoll:   schreiber,
		Stand:       anwendung.Fensterstand(context.Background()),
	})
}

// uebergebeneDatei gibt den ersten Pfad aus der Kommandozeile, wenn einer da
// ist.
//
// **Ohne die Schalter, die macOS selbst anhängt.** Wird eine Anwendung über
// LaunchServices gestartet, steht in `os.Args` ein `-psn_0_12345` — die
// Prozessseriennummer. Wer ihn als Dateinamen nähme, meldete beim Start einen
// Fehler über eine Datei, die niemand genannt hat.
func uebergebeneDatei(args []string) string {
	for _, arg := range args[1:] {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}

	return ""
}
