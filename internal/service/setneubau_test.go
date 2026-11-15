// **Ein Set ohne Quelldatei** — Befund **F02** der nativen Prüfung
// (#1746).
//
// # Warum eine eigene Datei
//
// **Weil `setschreiber_test.go` mit diesen Läufen über 400 Zeilen ginge** (§2),
// und die Grenze kennt keine Näherung. Der Schnitt folgt dabei der Frage und
// nicht der Zeilenzahl: dort steht, **was aus einer vorhandenen Datei wird** —
// die Byte-Zusage, eine geänderte Zeile, ein Kommentar, der bleibt. Hier
// steht, **was entsteht, wo es keine gibt**.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Set ohne Quelle wird zusammengesetzt** — Befund **F02** der nativen
// Prüfung (#1746).
//
// # Hier stand das Gegenteil, und es widersprach dem eigenen Vermerk
//
// Der Lauf hieß `TestEinSetOhneQuelleWirdBenannt` und hielt fest, dass der
// Schreiber ein selbst angelegtes Set mit `ErrNotFound` abweist — begründet
// mit der Byte-Zusage. **Die steht hier aber gar nicht auf dem Spiel**: es
// gibt keine Ursprungsdatei, mit der sich vergleichen ließe.
//
// [0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)
// sagt es seit dem 2026-08-15 ausdrücklich: *„Ein Set, das in der Anwendung
// entsteht, hat keine. Dann wird zusammengesetzt, wie `msepaket.Schreibe` es
// für Knoten ohne Rohzeile tut. Die Byte-Zusage gilt dort nicht — es gibt
// nichts, womit zu vergleichen wäre."*
//
// Umgesetzt war das nie, und dieser Lauf ließ die Lücke wie eine Entscheidung
// aussehen. Was sie gekostet hat, steht in der nativen Prüfung von v0.5.0:
// **⌘S und „Sichern unter…" meldeten beide „gesuchten Eintrag nicht
// gefunden", und es entstand keine Datei.** Ein Set, das in dieser Anwendung
// entsteht, ließ sich mit ihr nicht sichern.
func TestEinSetOhneQuelleWirdFrischZusammengesetzt(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "sternengarten.mse-set")

	gesichert, err := service.NeuerSetschreiber(setOhneQuelle(), &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if gesichert.Pfad != ort {
		t.Errorf("Pfad %q, erwartet %q", gesichert.Pfad, ort)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	// **Die Gestalt steht hier und nicht nur ihre Teile.** Ein Lauf, der
	// `strings.Contains` über vier Wörter fährt, wäre auch grün, wenn die
	// Einrückung fehlte oder die Reihenfolge stimmte — und beides entscheidet,
	// ob MSE2 die Datei öffnet (`docs/FORMATE.md`).
	const erwartet = "mse version: 2.0.0\n" +
		"game: probe\n" +
		"stylesheet: probe-kunst\n" +
		"set info:\n" +
		"\ttitle: Sternengarten\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n"

	if string(gelesen) != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// **Und was herauskommt, lässt sich wieder lesen** — FIX-F02.NAT-02.
//
// **Der Lauf darüber prüft das nicht.** Er hält die Bytes fest; ob sie ein
// Paket ergeben, das dieselbe Anwendung wieder öffnet, ist eine andere Frage
// — und genau die stellt die native Prüfung.
func TestDasFrischGeschriebeneSetLaesstSichWiederLesen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "sternengarten.mse-set")

	if _, err := service.NeuerSetschreiber(setOhneQuelle(), &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	baum, zerteilfehler := msepaket.Zerteile("set", gelesen)
	if zerteilfehler != nil {
		t.Fatalf("zerteilen: %v", zerteilfehler)
	}

	for schluessel, erwartet := range map[string]string{
		"mse version": "2.0.0",
		"game":        "probe",
		"stylesheet":  "probe-kunst",
	} {
		knoten := baum.Ersten(schluessel)
		if knoten == nil {
			t.Errorf("%q fehlt — ohne sie öffnet MSE2 das Paket nicht", schluessel)

			continue
		}

		if knoten.Wert != erwartet {
			t.Errorf("%q ist %q, erwartet %q", schluessel, knoten.Wert, erwartet)
		}
	}

	if karten := baum.Alle("card"); len(karten) != 1 {
		t.Fatalf("%d Kartenblöcke, erwartet 1", len(karten))
	}
}

// **Ohne Spiel entsteht nichts** — ein Paket ohne `game:` öffnet in MSE2 nicht,
// und eines zu erfinden wäre eine Falschangabe.
func TestEinSetOhneSpielLaesstSichNichtZusammensetzen(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.koepfe = []domain.SetHead{{ID: 7, Name: "Sternengarten", Style: "probe-kunst"}}

	_, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, filepath.Join(t.TempDir(), "wieder.mse-set"))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler: %v — erwartet war eine Beanstandung", err)
	}
}

// **Und ein Set, das der Bestand gar nicht führt, auch nicht.**
func TestEinUnbekanntesSetOhneQuelleWirdBenannt(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.koepfe = []domain.SetHead{}

	_, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, filepath.Join(t.TempDir(), "wieder.mse-set"))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v — erwartet war »nicht gefunden«", err)
	}
}

// setOhneQuelle gibt einen Bestand mit einem selbst angelegten Set.
//
// **So, wie `Importeur.Neu` es anlegt**: der Name steht zugleich als
// `title` unter den Setangaben — erfunden wird hier nichts.
func setOhneQuelle() quellenDouble {
	return quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{
				ID: 7, Name: "Sternengarten", Game: "probe", Style: "probe-kunst",
			}},
			werte: []domain.Value{{Field: "title", Text: "Sternengarten"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Nebelhüterin"},
			}}},
		},
	}
}

// **Die Kopfangaben aus dem Dateikopf kommen mit, wo sie etwas sagen.**
//
// **Und leere nicht** — `apprentice code` steht in jedem kanonischen Set leer,
// `version control` auf `none`. Eine nichtssagende Zeile in eine frische Datei
// zu setzen hieße, dem Set eine Angabe anzudichten, die niemand gemacht hat.
func TestDieKopfangabenDesFormatsKommenMit(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.format = domain.Setkopf{
		MSEFassung:        "2.1.0",
		Spielfassung:      "2020-01-01",
		Stilfassung:       "2021-07-06",
		Versionskontrolle: "none",
	}

	ort := filepath.Join(t.TempDir(), "sternengarten.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	const erwartet = "mse version: 2.1.0\n" +
		"game: probe\n" +
		"stylesheet: probe-kunst\n" +
		"game version: 2020-01-01\n" +
		"stylesheet version: 2021-07-06\n" +
		"version control: none\n" +
		"set info:\n" +
		"\ttitle: Sternengarten\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n"

	if string(gelesen) != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// **Ein Set ohne Stil bekommt keine leere `stylesheet:`-Zeile.**
//
// Sie wäre schlechter als keine: MSE2 suchte dann ein Paket mit leerem Namen.
func TestEinSetOhneStilBekommtKeineLeereZeile(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.koepfe = []domain.SetHead{{ID: 7, Name: "Sternengarten", Game: "probe"}}

	ort := filepath.Join(t.TempDir(), "sternengarten.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if strings.Contains(string(gelesen), "stylesheet") {
		t.Errorf("die Datei nennt ein Stilpaket, obwohl keines gesetzt ist:\n%s", gelesen)
	}
}

// **Ein Formatkopf, der sich nicht lesen lässt, wird gemeldet.**
//
// **Ein eigener Fehlerweg und nicht derselbe wie die Bestandszeile**: sie
// werden nacheinander gelesen, und wer nur den ersten prüft, hat den zweiten
// nie betreten.
func TestEinUnlesbarerFormatkopfWirdGemeldet(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.formatfehler = errors.New("der Kopf ist fort")

	_, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, filepath.Join(t.TempDir(), "wieder.mse-set"))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "der Kopf ist fort") {
		t.Errorf("Fehler: %v — der Grund kommt nicht durch", err)
	}
}
