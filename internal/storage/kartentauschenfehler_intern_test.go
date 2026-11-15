package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege des Umordnens**
// (#628).
//
// Paketintern und über den Stolpertreiber. **Und sie gehören geprüft, weil ein
// halber Tausch eine Ordnung hinterließe, die niemand angeordnet hat** — im
// schlimmsten Fall zwei Karten auf negativen Stellen, die keine Liste mehr
// zeigt.
func TestDasUmordnenMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht": "",
		"die Abfrage scheitert":         "stolper-fragen",
		"eine Zeile liest sich nicht":   "stolper-stellen-scan",
		"das Zeilenende stolpert":       "stolper-stellen-ende",
		"der Verlauf scheitert":         "stolper-stellen-verlauf",
		"das Parken scheitert":          "stolper-stellen-parken",
		"das Setzen scheitert":          "stolper-stellen-setzen",
		"der Commit scheitert":          "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if err := NeuerKartenschreiber(db).KartenTauschen(probekontext(t), 1, 2); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Ein unlesbarer Stellenweg ist ein Fehler und keine leere Ordnung.**
//
// `hinweg` und `rueckweg` sind JSON in einer Textspalte. Eine Zeile aus einer
// Fassung, die etwas anderes schrieb, kommt hier an — und einen leeren Stand
// zurückzugeben hieße, beide Karten auf Stelle 0 zu setzen.
func TestEinUnlesbarerStellenwegIstEinFehler(t *testing.T) {
	if _, err := stellenstandAus("{kein JSON"); err == nil {
		t.Error("kein Fehler auf einen krummen Stellenweg")
	}
}

// **Ein Eintrag ohne Weg ist ein leerer Stand, kein Fehler.** So kommt eine
// Zeile aus einer Sitzung an, die abgebrochen ist.
func TestEinLeererStellenwegGibtDenLeerenStand(t *testing.T) {
	stand, err := stellenstandAus("")
	if err != nil || stand == nil || *stand != (domain.Stellenstand{}) {
		t.Errorf("%+v, %v — erwartet den leeren Stand", stand, err)
	}
}

// **Und ein Schritt ohne Stand geht ins Leere, statt zu scheitern.**
//
// `stellenGehen` bekommt `nil`, wenn ein Eintrag der Art `reihenfolge` aus
// einer Fassung stammt, die ihn nicht geschrieben hat. Nichts zu tun ist die
// richtige Antwort: eine Ordnung zu erfinden wäre die falsche.
func TestEinSchrittOhneStellenstandTutNichts(t *testing.T) {
	if err := stellenGehen(probekontext(t), nil, nil); err != nil {
		t.Errorf("%v, erwartet keinen Fehler", err)
	}
}

// **Die beiden Wege des Tauschs melden ihre Fehler.**
//
// Eine zurückgerollte Transaktion gibt `ErrTxDone` auf jede Anweisung — damit
// sind beide Zweige erreichbar, das Parken und das Setzen.
func TestDieWegeDesTauschsMeldenIhreFehler(t *testing.T) {
	tx := abgebrocheneTransaktion(t)

	if err := stellenGehen(probekontext(t), tx, &domain.Stellenstand{
		Eine: 1, Andere: 2, EineStelle: 1, AndereStelle: 0,
	}); err == nil {
		t.Error("kein Fehler")
	}
}

// **Und ein Tausch, der beim Bewegen des Verlaufs scheitert, wird gemeldet.**
//
// Der Weg dorthin geht über `schrittGehen`: ein Eintrag der Art
// `reihenfolge`, dessen Ausführung scheitert, darf die Rücknahme nicht still
// für geglückt erklären.
func TestEinGescheiterterTauschImVerlaufWirdGemeldet(t *testing.T) {
	tx := abgebrocheneTransaktion(t)

	schritt := []domain.Verlaufseintrag{{
		KarteID:        1,
		Handlung:       domain.HandlungReihenfolge,
		VorherStellen:  &domain.Stellenstand{Eine: 1, Andere: 2},
		NachherStellen: &domain.Stellenstand{Eine: 1, Andere: 2},
	}}

	for name, zurueck := range map[string]bool{"zurück": true, "vorwärts": false} {
		t.Run(name, func(t *testing.T) {
			if err := schrittGehen(probekontext(t), tx, schritt, 7, zurueck); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein unlesbarer Weg macht den ganzen Eintrag zum Fehler** — beide Seiten.
func TestEinKrummerStellenschrittWirdGemeldet(t *testing.T) {
	for name, wege := range map[string][2]string{
		"der Hinweg":  {"{krumm", "{}"},
		"der Rückweg": {"{}", "{krumm"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := stellenschrittAus(domain.Verlaufseintrag{}, wege[0], wege[1], 0, "fenster")
			if err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
