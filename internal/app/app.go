package app

import (
	"os"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// NeueApp baut die Anwendung über einer Datenbankdatei.
func NeueApp(datenbankpfad string) *App {
	anwendung := &App{
		datenbankpfad: datenbankpfad,
		oeffnen:       storage.Oeffnen,

		// **Die Uhr wird hier gesetzt und nicht erst beim ersten Blick.** Ein
		// Feld, das nach dem Start `nil` ist, ist genau die Naht, vor der
		// `naehte_intern_test.go` warnt (#1165).
		uhr:            time.Now,
		paketort:       Paketverzeichnis,
		bilderort:      Bilderverzeichnis,
		downloadort:    downloadortEcht,
		saatort:        Saatverzeichnis,
		vorlagenort:    Vorlagenverzeichnis,
		beispielsetort: Beispielsetverzeichnis,
		mse2orte:       MSE2Orte,
		schnittstellen: echteSchnittstellenorte(),

		// **Die Bildnähte als Verbund** — eingebettet, damit `anwendung.go`
		// unter der Zeilengrenze aus §2 bleibt (`bildnaehte.go`).
		bildnaehte: bildnaehte{packen: schnellPacken},

		// **Die Nähte zum Fenster setzt die Hülle** (`internal/fenster`,
		// ENG-02) — bis dahin gelten die aus `ohnefenster.go`, die nichts tun
		// oder einen Fehler mit Namen geben.
		menuenaht:       menuenaht{menueSetzen: menueOhneFenster, melden: meldenOhneFenster},
		dateiwahl:       dateiwahlOhneFenster,
		ablageLesen:     ablageLesenOhneFenster,
		ablageSchreiben: ablageSchreibenOhneFenster,
		dateiEntfernen:  os.Remove,

		// **Der Rückweg wird hier verdrahtet und nicht in `Starten`.**
		//
		// Er ist der einzige, der gebraucht wird, **wenn `Starten`
		// scheitert** — und dann läuft es nie zu Ende: dort verdrahtet, hätte
		// der Knopf der Wiederherstellungsseite nichts, das er rufen könnte
		// (#962).
		//
		// Keine der beiden Nähte braucht eine Verbindung: sie öffnen ihre Datei
		// selbst und nur lesend.
		bestandsnaehte: bestandsnaehte{
			sicherungspruefer: sicherungsnaht,
			alleinpruefung:    alleinnaht,
		},
	}

	// **Die Sitzungsverwaltung steht ab hier und nicht erst beim ersten
	// Ruf** (N1, #1454). Eine Naht, die später entstünde, wäre bis dahin
	// `nil` — genau das, wovor `naehte_intern_test.go` warnt.
	//
	// **Und sie fragt die Anwendung nach ihrem Wähler**, statt ihn
	// festzuhalten: [App.MitWaehler] darf danach kommen.
	anwendung.peerlage.sitzungen = service.NeuePeersitzungen(
		kontaktrechte{app: anwendung}, waehlernaht{app: anwendung},
		ausweisnaht{app: anwendung}, anwendung.uhr, Sitzungstiefe)
	anwendung.peerlage.einloesegrenze = service.NeueEinloesegrenze(anwendung.uhr)

	return anwendung
}
