package service_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probevorlage liest das Prüfstück.
func probevorlage(t *testing.T) msepaket.Exportvorlage {
	t.Helper()

	paket, err := msepaket.Oeffne(
		filepath.Join("..", "..", "testdaten", "probe-export.mse-export-template"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return paket.Exportmodell()
}

// zweiKarten gibt zwei Karten mit Namen und Farbe.
func zweiKarten() []*domain.CardRow {
	return []*domain.CardRow{
		{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "Feuerball"}, {Field: "farbe", Text: "rot"},
		}},
		{ID: 2, Number: 2, Values: []domain.Value{
			{Field: "name", Text: "Eisball"}, {Field: "farbe", Text: "blau"},
		}},
	}
}

// **Eine Exportvorlage erzeugt eine Ausgabe** (#51).
//
// # Was das Skript sieht
//
// `cards` als Liste, `card` je Durchgang, `options` als Abbildung ihrer
// Wahlmöglichkeiten. Der Wert des **letzten** Ausdrucks ist die Ausgabe — so
// schreiben es alle 19 Vorlagen von `M15-Magic-Pack`.
func TestEineExportvorlageErzeugtEineAusgabe(t *testing.T) {
	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), probevorlage(t), zweiKarten(), nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	erwartet := "Meine Karten\nFeuerball (rot)\nEisball (blau)\n"
	if ausgabe != erwartet {
		t.Errorf("Ausgabe %q, erwartet %q", ausgabe, erwartet)
	}
}

// **Die Vorgaben gelten, wenn niemand wählt** — und eine Wahl schlägt sie.
func TestEineWahlSchlaegtDieVorgabe(t *testing.T) {
	ausgabe, _, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), probevorlage(t), zweiKarten(),
		nil, map[string]string{"nur namen": "ja", "kopfzeile": "Nur die Namen"})
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "Nur die Namen\nFeuerball\nEisball\n" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Ein Skript, das sich nicht ausrechnen lässt, ist ein Fehler und kein
// halbes Ergebnis.**
//
// Anders als bei einer Karte (§8): dort ist eine Ebene weniger brauchbar, hier
// wäre eine halbe Datei eine, die jemand für vollständig hält.
func TestEinKaputtesVorlagenskriptWirdGemeldet(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = "gibtsnicht()"

	if _, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, zweiKarten(), nil, nil); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Hinweis, keine Grenze — und seit #227 steht das auch so da.**
//
// # Was er ist
//
// Ein Namensfilter über die Bibliothek. Er fällt auf, wenn jemand eine Funktion
// `write_something` nennt, und **sonst nichts**: eine Funktion namens `sichere`
// oder `ablegen` liefe grün durch, und eine bestehende Funktion, die eines
// Tages intern eine Datei öffnet, ebenfalls.
//
// # Wo die Grenze steht
//
// In `schichten.TestDieSpracheKannNichtsAnfassen`: `internal/msescript`
// importiert nichts, womit sich eine Datei, ein Prozess oder eine Verbindung
// anfassen ließe. Wer das nicht **kann**, tut es nicht — gleich, wie seine
// Funktionen heißen. Bis
// #227 hing die
// Zusage aus #51 an diesem Namensfilter hier, und das war die halbe Zusage: ein
// Wächter, der sie hält, macht die andere Hälfte unsichtbar (§10).
//
// # Warum er trotzdem bleibt
//
// **Weil er etwas anderes sagt.** Die Importgrenze hält, dass die Sprache
// nichts anfassen **kann**; dieser Filter hält, dass ihre Namen nicht so tun,
// als könnten sie es. Eine Funktion `write_image_file` **in der Bibliothek**
// wäre auch dann eine Verwechslung, wenn sie nichts schriebe — sie gehört in
// die Umgebung, die der Aufrufer stellt (#216).
func TestAuffaelligeDateifunktionsnamenWerdenGemeldet(t *testing.T) {
	// **Die Wörter sind eng gewählt.** Der erste Anlauf führte `remove`, und
	// `remove_tags` schlug an — ein Wächter, der bei jedem zweiten Namen
	// klingelt, wird abgeschaltet, statt gelesen zu werden.
	verdaechtig := []string{
		"write_", "read_", "open", "file", "path", "exec", "system", "dir",
	}

	for _, name := range service.NamenDerBibliothek() {
		klein := strings.ToLower(name)

		for _, wort := range verdaechtig {
			if strings.Contains(klein, wort) {
				t.Errorf("die Bibliothek führt %q — greift sie ins Dateisystem?", name)
			}
		}
	}
}

// **Eine Vorlage ohne Skript ist ein Fehler, kein leerer Text.**
//
// Zwei der Pakete in `M15-Magic-Pack` tragen die Endung und kein `script:` —
// wer daraus eine leere Datei machte, schriebe sie und meldete Erfolg.
func TestEineVorlageOhneSkriptWirdAbgewiesen(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = "   \n  "

	if _, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, zweiKarten(), nil, nil); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Wahl, die die Vorlage nicht führt, wird benannt.**
//
// Sie ist fast immer ein Tippfehler — und still übergangen sähe die Ausgabe
// aus wie eine, in der die Wahl gewirkt hat.
func TestEineUnbekannteWahlWirdBenannt(t *testing.T) {
	_, befunde, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), probevorlage(t), zweiKarten(),
		nil, map[string]string{"gibtsnicht": "egal"})
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Ort != "gibtsnicht" {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Eine Karte trägt jedes Feld ihres Spiels, auch das leere**
// (#609).
//
// **ported.** In MSE legt der Kartenkonstruktor für **jedes** erklärte Feld
// einen Wert an:
//
//	Card::Card() … { data.init(game_for_reading()->card_fields); }
//
// `src/data/card.cpp`, Zeilen 20–30, und `IndexMap::init`
// (`src/util/index_map.hpp`, Zeilen 48–62), Fassung f9d9356 vom 2021-07-06.
// `card.notes` gibt es dort **immer** — leer, wenn die Datei nichts
// geschrieben hat.
//
// **Bei uns fehlte es**, und seit ein fehlendes Glied ein Fehler ist statt
// `nil`, scheiterte jede Vorlage daran, die ein ausgelassenes Feld liest.
// Zwölf der achtzehn Fehlschläge über `M15-Magic-Pack` hatten diese Gestalt.
//
// **Der Zeichenpfad hatte den Schritt längst** — `mitVorgaben` in
// `planumgebung.go`. Der Ausfuhrpfad bekam die Kartenfelder gar nicht erst.
func TestEineKarteTraegtJedesFeldIhresSpiels(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(cards[0].notizen) + "|" + to_string(cards[0].name)`
	vorlage.Kartenfelder = []domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "notizen", Kind: domain.KindText},
	}

	karten := []*domain.CardRow{{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
	}}}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "|Feuerball" {
		t.Errorf("Ausgabe %q, erwartet `|Feuerball`", ausgabe)
	}
}

// **Ein Feld, das die Karte führt, schlägt das leere** — die Vorgabe füllt
// nur auf, sie überschreibt nicht.
func TestDasFeldDerKarteSchlaegtDasLeere(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(cards[0].name)`
	vorlage.Kartenfelder = []domain.Field{{Name: "name", Kind: domain.KindText}}

	karten := []*domain.CardRow{{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
	}}}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "Feuerball" {
		t.Errorf("Ausgabe %q, erwartet `Feuerball`", ausgabe)
	}
}

// **Ein Feldname mit Leerzeichen wird kanonisch** — `rule text` heißt im
// Skript `rule_text`, und das gilt für das aufgefüllte Feld wie für das
// vorhandene.
func TestEinAufgefuelltesFeldWirdKanonischBenannt(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(cards[0].rule_text) + "!"`
	vorlage.Kartenfelder = []domain.Field{{Name: "rule text", Kind: domain.KindText}}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, []*domain.CardRow{{ID: 1}}, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "!" {
		t.Errorf("Ausgabe %q, erwartet `!`", ausgabe)
	}
}

// **`notes` steht an jeder Karte, gleich was das Spiel erklärt**
// (#609).
//
// **ported.** Es ist kein Feld, sondern ein Glied der Karte —
// `String notes;` (`src/data/card.hpp`, Zeile 39) und `REFLECT(notes)`
// (`src/data/card.cpp`, Zeile 91), Fassung f9d9356 vom 2021-07-06.
//
// **Und das ist kein ausgedachter Sonderfall**: `magic.mse-game` erklärt gar
// kein Feld `notes` — nur `sorted notes` —, und trotzdem lesen fünf der
// neunzehn Exportvorlagen des Packs `card.notes`.
func TestNotesStehtAnJederKarte(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(cards[0].notes) + "/" + to_string(cards[1].notes)`

	karten := []*domain.CardRow{
		{ID: 1, Values: []domain.Value{{Field: "notes", Text: "!noexport"}}},
		{ID: 2},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "!noexport/" {
		t.Errorf("Ausgabe %q, erwartet `!noexport/`", ausgabe)
	}
}

// **Das Set trägt jedes seiner Felder, auch das leere**
// (#609).
//
// Dieselbe Regel wie an der Karte, eine Ebene höher: MSEs `Set` legt für jedes
// erklärte `set field:` einen Wert an
// (`Set::Set`, `src/data/set.cpp`, Zeilen 30–36, Fassung f9d9356 vom
// 2021-07-06). `set.title` gibt es dort immer.
//
// **Neun der achtzehn Fehlschläge über `M15-Magic-Pack` waren `set.*`** —
// sechsmal `title`, dreimal `set_code`.
func TestDasSetTraegtJedesSeinerFelder(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(set.title) + "|" + to_string(set.set_code)`
	vorlage.Setfelder = []domain.Field{
		{Name: "title", Kind: domain.KindText},
		{Name: "set code", Kind: domain.KindText},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, []*domain.CardRow{{ID: 1}},
			[]domain.Value{{Field: "title", Text: "Mein Probeset"}}, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "Mein Probeset|" {
		t.Errorf("Ausgabe %q, erwartet `Mein Probeset|`", ausgabe)
	}
}

// **Ohne Setangaben ist `set` eine leere Abbildung und kein fehlender Name.**
//
// Der Unterschied ist die Meldung, die jemand liest: „die Abbildung führt kein
// Glied `title`" sagt, welches Feld fehlt; „der Name `set` ist hier nicht
// bekannt" sagt, dass die Ausfuhr gar keine Setangaben kennt.
func TestOhneSetangabenMeldetDasGliedUndNichtDenNamen(t *testing.T) {
	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(set.title)`

	_, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, []*domain.CardRow{{ID: 1}}, nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "führt kein Glied") {
		t.Errorf("%v — erwartet eine Meldung über das Glied", err)
	}
}
