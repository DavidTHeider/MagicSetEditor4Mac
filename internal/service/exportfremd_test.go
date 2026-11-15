//go:build fremdpakete

// Jede **echte** Exportvorlage einmal durch — der Nachweis aus
// #51:
// „eine vorhandene Vorlage erzeugt eine Ausgabe".
//
//	MSE_FREMDPAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data" \
//	  go test -tags fremdpakete ./internal/service/ -run Exportvorlagen -v
package service_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilleVorlage gibt zulässigerweise nichts aus.
//
// **`Name Check` meldet Kartennamen, die es anderswo schon gibt.** Findet es
// keinen, ist die leere Ausgabe die richtige Antwort — und keine, die einen
// Lauf rot machen darf (#609).
//
// **Eine benannte Ausnahme und keine gelockerte Regel**: jede andere Vorlage,
// die schweigt, ist weiterhin ein Befund.
const stilleVorlage = "magic-namecheck.mse-export-template"

// ausserReichweite sind die Vorlagen, die an dieser MSE2-Fassung nicht laufen
// **können** ([0036](../../docs/entscheidungen/0036-card-regions-ist-ausser-reichweite.md)).
//
// Sie lesen `sts.card_regions`, und dieses Glied gibt es in `f9d9356` vom
// 2021-07-06 nicht — nachgesehen über den ganzen `src`-Baum, kein Treffer. Das
// Pack zielt auf eine neuere Fassung.
//
// **Sie stehen hier, damit die Zahl unten nicht wie ein unerklärter Rest
// aussieht.** Rot machen sie den Lauf ohnehin nicht; er zählt und benennt,
// statt zu scheitern.
//
// **Nur `Lackey`, und das ist gemessen, nicht erschlossen.** `card_regions`
// steht auch in `magic-cockatrice-v2` und `magic-lackeybot` — die laufen
// trotzdem durch, weil der Zweig mit den Vorgabewerten der Optionen nicht
// betreten wird. Wer sie anders einstellt, kann dort dasselbe treffen; hier
// steht, was dieser Lauf **antrifft**.
var ausserReichweite = map[string]string{
	"magic-lackeyccg.mse-export-template": "card_regions",
}

func TestExportvorlagenAusFremdpaketen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("Wurzel lesen: %v", err)
	}

	// **Fünf gefüllte Felder, und der Rest kommt aus dem Spiel** (#609).
	//
	// Hier standen nur diese fünf, und zwölf der achtzehn Fehlschläge waren
	// „die Abbildung führt kein Glied `notes`" und seinesgleichen. Das war
	// **kein** Befund über die Ausfuhr, sondern einer über diese Datei — §3 in
	// der Gegenrichtung: eine Vorrichtung, die **knausriger** ist als der
	// Betrieb, sagt über den Stand genauso wenig wie eine großzügigere.
	//
	// **Behoben ist es nicht hier, sondern im Ausfuhrpfad**: `kartenliste`
	// füllt jedes Feld auf, das das Spiel erklärt. Diese fünf bleiben, weil
	// eine Karte mit Inhalt mehr prüft als eine leere.
	karten := []*domain.CardRow{{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
		{Field: "casting cost", Text: "2R"},
		{Field: "type", Text: "Spontanzauber"},
		{Field: "rule text", Text: "Füge drei Schaden zu."},
		{Field: "rarity", Text: "common"},
	}}}

	// **Ein Set ohne eingetragene Angaben** (#609). Die Felder kommen aus dem
	// Spiel und stehen leer da — genau der Zustand eines frisch angelegten
	// Sets, und der ist der strengere Fall.
	var setwerte []domain.Value

	gelesen, gelaufen, ausser := 0, 0, 0
	gruende := map[string]int{}

	// **Kandidaten sind die Pakete, die dastehen** — nicht eine Zahl im
	// Quelltext. Was übersprungen wurde und warum, steht danach einzeln da;
	// eine Vorlage, die still verschwindet, ist der Fall, um den es geht.
	kandidaten := 0
	uebersprungen := map[string]string{}

	for _, eintrag := range eintraege {
		if msepaket.ArtAusName(eintrag.Name()) != msepaket.ArtExportvorlage {
			continue
		}

		kandidaten++

		paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, eintrag.Name()))
		if fehler != nil {
			uebersprungen[eintrag.Name()] = "ließ sich nicht öffnen: " + fehler.Error()

			continue
		}

		// **Erst die Module, dann das Modell** — wie `Vorlagenliste.vorlageAus`
		// es tut (#216).
		//
		// Hier stand der Griff nach `Exportmodell()` unmittelbar hinter dem
		// Öffnen, und damit maß dieser Test einen Weg, den die Anwendung nicht
		// geht: zwei Vorlagen beginnen ihr Skript mit `include file:`, und
		// unaufgelöst scheitert das am Zerteiler. Eine Messung, die an der Naht
		// vorbeigeht, misst nicht den Stand, sondern die Vorrichtung — genau
		// die Klasse, vor der `AGENTS.md` §3 warnt.
		_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // ein fehlendes Modul ist ein Befund

		vorlage := paket.Exportmodell()
		_ = paket.Schliessen() //nolint:errcheck // gelesen ist gelesen

		// **Das Spielskript gehört dazu** — `Vorlagenliste` legt es bei (#216).
		if spielpaket, fehler := msepaket.Oeffne(
			filepath.Join(wurzel, vorlage.Spiel+"."+string(msepaket.ArtSpiel))); fehler == nil {
			_ = spielpaket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // ein fehlendes Modul ist ein Befund
			vorlage.Spielskript = spielpaket.Initskript()

			// **Und die Kartenfelder** (#609) — wie `Vorlagenliste.vorlageAus`
			// es tut. Ohne sie misst dieser Lauf einen Weg, den die Anwendung
			// nicht geht.
			spielmodell, _ := spielpaket.Spielmodell()
			vorlage.Kartenfelder = spielmodell.Fields
			vorlage.Setfelder = spielmodell.SetFields

			_ = spielpaket.Schliessen() //nolint:errcheck // gelesen ist gelesen
		}

		if vorlage.Skript == "" {
			uebersprungen[eintrag.Name()] = "trägt kein Skript"

			continue
		}

		gelesen++

		// **Mit einem Bildschreiber, der nichts schreibt.** Gemessen wird, ob
		// das *Skript* durchläuft; ein echter Rasterer verlangte hier eine
		// Datenbank, einen Stil und ein Fenster (#216).
		// **Und mit Beigaben, die nichts ablegen** (#657) — aus demselben
		// Grund: gemessen wird, ob das Skript durchläuft, nicht ob die Platte
		// schreibbar ist.
		ausgabe, _, lauffehler := ausfuhrprobe{
			ausfuehrer: service.NeuerVorlagenausfuehrer().
				MitSymbolen(echteSymbolquelle(wurzel),
					render.NeuerRasterer(msepaket.NeueBildquelle(wurzel))),
			mittel: service.Ausfuhrmittel{
				Bilder:            leerschreiber{},
				Bildziel:          t.TempDir(),
				Beigaben:          leerbeigaben{},
				Symbolzeichnungen: msepaket.NeueBildquelle(wurzel),
				Setpaket:          "",
				Verzeichnis:       "ausfuhr-files",
			},
		}.Fuehre(t.Context(), vorlage, karten, setwerte, nil)
		if lauffehler != nil {
			meldung := lauffehler.Error()
			if stelle := strings.LastIndex(meldung, ": "); stelle > 0 {
				meldung = meldung[stelle+2:]
			}

			gruende[meldung]++

			// **Mit dem Namen der Vorlage**, nicht nur mit der Zahl. Die
			// Aufstellung darunter sagt, *wie oft* etwas fehlt; sie sagt nicht,
			// *wo* — und #216
			// verlangt, jeden Zerteilfehler einzeln anzusehen. Dafür muss man
			// die Datei aufmachen können.
			if grund, jenseits := ausserReichweite[eintrag.Name()]; jenseits {
				ausser++

				t.Logf("  ⊘ %s: außer Reichweite (%s, Vermerk 0036): %s",
					eintrag.Name(), grund, lauffehler.Error())
			} else {
				t.Logf("  ✗ %s: %s", eintrag.Name(), lauffehler.Error())
			}

			continue
		}

		// **Eine leere Ausgabe ist kein Durchlauf** (#609). Hier stand
		// `gelaufen++` **hinter** der Meldung und ohne `continue`: eine
		// Vorlage mit null Zeichen zählte als durchgelaufen und machte den
		// Lauf zugleich rot. Beides gleichzeitig ist keine Aussage — und
		// **die eine Vorlage, die #609 als „durchgelaufen" führte, war genau
		// diese**: `magic-namecheck` gibt null Zeichen aus. Die wahre Zahl war
		// null.
		if ausgabe == "" {
			// **Und bei dieser einen ist es richtig so.** `Name Check` meldet
			// Kartennamen, die anderswo schon vergeben sind; findet es nichts,
			// hat es nichts zu sagen. Eine Vorlage, die schweigt, weil es
			// nichts zu melden gibt, ist kein Fehlschlag — sie ist nur kein
			// Beleg dafür, dass eine Ausgabe entsteht.
			if eintrag.Name() != stilleVorlage {
				t.Errorf("%s: leere Ausgabe ohne Fehler", eintrag.Name())
			}

			t.Logf("  · %s: leere Ausgabe (nichts zu melden)", eintrag.Name())

			continue
		}

		gelaufen++

		t.Logf("  ✓ %s: %d Zeichen Ausgabe", eintrag.Name(), len(ausgabe))
	}

	// **Die Zahl steht mit dem, was sie nicht enthält** (§10): manche Vorlagen
	// können an dieser MSE2-Fassung nicht laufen, und das ist entschieden, nicht
	// offen ([0036](../../docs/entscheidungen/0036-card-regions-ist-ausser-reichweite.md)).
	//
	// **Gezählt wird, was dieser Lauf angetroffen hat**, nicht die Länge der
	// Liste: `M15-Magic-Pack` führt eine der beiden Vorlagen, der volle Pack
	// beide.
	t.Logf("%d Exportvorlagen gelesen, %d liefen durch (%d außer Reichweite)",
		gelesen, gelaufen, ausser)

	for grund, anzahl := range gruende {
		t.Logf("  %3d × %s", anzahl, grund)
	}

	// **Gelesen werden muss jede** — und „jede" heißt: jede, die in der
	// gesetzten Wurzel liegt.
	//
	// **Hier stand `!= 19`**, und das band die Zusage an **eine** Paketwurzel:
	// gegen `full-magic-pack` meldete der Lauf „22 Vorlagen gelesen, erwartet
	// 19" und wurde rot, ohne dass etwas kaputt war
	// (#779).
	// Rot heißt „etwas ist kaputt", nicht „du hast das andere Pack gesetzt".
	//
	// **Die Zahl aus dem Verzeichnis ist zugleich die schärfere Zusage**: sie
	// nennt die Übersprungenen einzeln, statt eine Differenz zu melden.
	for name, grund := range uebersprungen {
		t.Errorf("  %s: %s", name, grund)
	}

	if gelesen != kandidaten {
		t.Errorf("%d von %d Vorlagen gelesen", gelesen, kandidaten)
	}

	// **Wie viele *laufen*, ist die verfolgte Zahl und noch keine Zusage.**
	// Sie steht im Protokoll, nicht in einer Bedingung: was fehlt, ist die
	// Standardbibliothek (#25) und die Verkettung von Funktionen mit `+`, und
	// beides ist eigener Umfang
	// (#202).
	//
	// **Eine Bedingung darüber wäre eine Zusage, die wir nicht halten** — und
	// ein Test, der sie behauptete, wäre rot, ohne dass jemand etwas falsch
	// gemacht hätte. Was er stattdessen hält: keine Vorlage gibt **still** eine
	// leere Ausgabe zurück.
}

// leerschreiber tut so, als hätte er geschrieben.
//
// **Er prüft nichts und schreibt nichts.** Was `write_image_file` mit seinen
// Argumenten macht, steht in `exportbild_test.go`; hier geht es allein darum,
// ob die neunzehn Vorlagen bis zu ihrer Ausgabe kommen.
// echteSymbolquelle gibt die Symbolschriften der Paketwurzel — **keine
// Attrappe** (§3).
//
// Der Messlauf soll sagen, ob `symbols_to_html` mit den Schriften des Packs
// läuft; eine Attrappe, die jede Schrift liefert, sagte etwas über die
// Vorrichtung (#658).
func echteSymbolquelle(wurzel string) *Stilspeicherzugriff {
	return &Stilspeicherzugriff{speicher: service.NeuerStilspeicher(nil, wurzel)}
}

// Stilspeicherzugriff reicht die Symbolschriften durch.
type Stilspeicherzugriff struct{ speicher *service.Stilspeicher }

func (s *Stilspeicherzugriff) Symbolschrift(name string) (msepaket.Symbolschrift, bool) {
	return s.speicher.Symbolschrift(name)
}

type leerschreiber struct{}

func (leerschreiber) Schreibe(_ context.Context, _ int64, verzeichnis, datei string,
	_ service.Bildformat, _ int,
) (*service.Bildschreibung, error) {
	return &service.Bildschreibung{Pfad: filepath.Join(verzeichnis, datei)}, nil
}

// leerbeigaben nimmt jede Datei an und schreibt keine.
type leerbeigaben struct{}

func (leerbeigaben) AusVorlage(string) ([]byte, error) { return nil, nil }

func (leerbeigaben) Lege(datei string, _ []byte) (string, error) { return datei, nil }

func (leerbeigaben) LegeSet(context.Context, string) (string, error) { return "set", nil }
