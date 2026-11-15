package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildmitschrift merkt sich, wohin geschrieben werden sollte.
type bildmitschrift struct {
	verzeichnis, datei string
	format             service.Bildformat
	karteID            int64
}

func (b *bildmitschrift) Schreibe(_ context.Context, karteID int64,
	verzeichnis, datei string, format service.Bildformat, _ int,
) (*service.Bildschreibung, error) {
	b.karteID, b.verzeichnis, b.datei, b.format = karteID, verzeichnis, datei, format

	return &service.Bildschreibung{Pfad: filepath.Join(verzeichnis, datei)}, nil
}

// **Mit `create directory: true`**, so wie zehn der neunzehn Vorlagen des
// M15-Packs. Ohne die Angabe darf eine Vorlage gar nicht schreiben (#732), und
// diese Tests prüfen die Funktionen und nicht die Erlaubnis — die hat ihre
// eigene Datei, `exportverzeichnis_test.go`.
func vorlageMit(skript string) msepaket.Exportvorlage {
	return msepaket.Exportvorlage{
		KurzerName: "Bilder", Skript: skript, Verzeichnis: true,
	}
}

func eineKarte() []*domain.CardRow {
	return []*domain.CardRow{{ID: 42, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
	}}}
}

// **`write_image_file` schreibt die Karte, die es bekommt** (#216).
func TestWriteImageFileSchreibtDieGemeinteKarte(t *testing.T) {
	mitschrift := &bildmitschrift{}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: mitschrift, Bildziel: "/ziel", Bildbreite: 750,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_image_file(cards.0, file: "bild.jpg")`),
			eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	// **Die Kennung kommt aus der Marke**, nicht aus einem Feld der Karte.
	if mitschrift.karteID != 42 {
		t.Errorf("Karte %d, erwartet 42", mitschrift.karteID)
	}

	if mitschrift.verzeichnis != "/ziel" || mitschrift.datei != "bild.jpg" {
		t.Errorf("%s / %s", mitschrift.verzeichnis, mitschrift.datei)
	}

	// **Das Format steht in der Endung** — die einzige Angabe, die die
	// Vorlagen dazu machen.
	if mitschrift.format != service.FormatJPG {
		t.Errorf("Format %s", mitschrift.format)
	}

	// **Zurück kommt der Dateiname**, nicht der Pfad dieses Rechners: die
	// Vorlage hängt ihn in ein `<img src="…">`.
	if ausgabe != "bild.jpg" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **`directory:` liegt *unter* dem Ziel — die Vorlage benennt, sie wählt nicht.**
func TestWriteImageFileLegtDasUnterverzeichnisUnterDasZiel(t *testing.T) {
	mitschrift := &bildmitschrift{}

	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: mitschrift, Bildziel: "/ziel", Bildbreite: 0,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(),
			vorlageMit(`write_image_file(cards.0, directory: "bilder", file: "a.png")`),
			eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if mitschrift.verzeichnis != filepath.Join(string(filepath.Separator)+"ziel", "bilder") {
		t.Errorf("%q", mitschrift.verzeichnis)
	}
}

// **Die Gegenprobe zur Ausbruchsgrenze** (#51).
//
// Eine Vorlage darf keinen Ort wählen. Was nach einem Pfad aussieht, wird zu
// einem Namen gemacht oder fällt weg — geschrieben wird in jedem Fall unter
// dem Verzeichnis, das der Aufrufer bestimmt hat.
func TestEineVorlageKannNichtAusDemZielverzeichnisAusbrechen(t *testing.T) {
	faelle := []struct{ name, genannt string }{
		{"zwei Ebenen hinauf", `"../.."`},
		{"ein absoluter Pfad", `"/etc"`},
		{"ein Pfad mit Trennern", `"a/../../b"`},
		{"nur ein Punkt", `"."`},
		{"leer", `""`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			mitschrift := &bildmitschrift{}

			_, _, err := probelauf(service.Ausfuhrmittel{
				Bilder: mitschrift, Bildziel: "/ziel", Bildbreite: 0,
				Verzeichnis: "ausfuhr-files",
			}).
				Fuehre(t.Context(), vorlageMit(
					`write_image_file(cards.0, directory: `+fall.genannt+`, file: "a.png")`),
					eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			if !strings.HasPrefix(mitschrift.verzeichnis, "/ziel") {
				t.Errorf("geschrieben nach %q — außerhalb des Ziels", mitschrift.verzeichnis)
			}

			if strings.Contains(mitschrift.verzeichnis, "..") {
				t.Errorf("%q trägt einen Ausbruch", mitschrift.verzeichnis)
			}
		})
	}
}

// **Ohne Bildschreiber gibt es die Funktion nicht.**
//
// Die ehrlichere Antwort als eine Datei an einem Ort, den niemand gewählt hat.
func TestOhneBildschreiberIstWriteImageFileUnbekannt(t *testing.T) {
	_, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlageMit(`write_image_file(cards.0, file: "a.png")`),
			eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "write_image_file") {
		t.Errorf("die Meldung nennt die Funktion nicht: %v", err)
	}
}

// **Was keine Karte ist, wird abgewiesen.**
//
// Ohne diese Prüfung schriebe `write_image_file("text", …)` die Karte mit der
// Kennung, die eine leere Marke ergibt — also gar keine, still.
func TestWriteImageFileWeistEtwasAbDasKeineKarteIst(t *testing.T) {
	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: &bildmitschrift{}, Bildziel: "/ziel", Bildbreite: 0,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_image_file("kein Kartenwert", file: "a.png")`),
			eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ohne `file:` ist es ein Tippfehler.**
func TestWriteImageFileBrauchtEinenDateinamen(t *testing.T) {
	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: &bildmitschrift{}, Bildziel: "/ziel", Bildbreite: 0,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_image_file(cards.0)`), eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein Fehler des Schreibers bleibt ein Fehler.**
func TestEinFehlerBeimSchreibenBrichtDieVorlageAb(t *testing.T) {
	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: sperriger{}, Bildziel: "/ziel", Bildbreite: 0,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_image_file(cards.0, file: "a.png")`),
			eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

type sperriger struct{}

func (sperriger) Schreibe(_ context.Context, _ int64, _, _ string,
	_ service.Bildformat, _ int,
) (*service.Bildschreibung, error) {
	return nil, errors.New("die Platte ist voll")
}

// **Ohne Stellungsargument ist es kein Aufruf.**
func TestWriteImageFileBrauchtEineKarte(t *testing.T) {
	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: &bildmitschrift{}, Bildziel: "/ziel", Bildbreite: 0,
		Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_image_file(file: "a.png")`), eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein gescheiterter Satz im Initskript wird als Befund genannt** (#216).
//
// Er bricht nicht ab — aber verschwiegen wird er auch nicht. Der stille Weg
// wäre genau das ohne diese Meldung.
//
// **Der Auslöser war bis #240 ein Vorausblick im Muster**, den Gos `regexp`
// nicht kannte. Den kann dieses Programm inzwischen
// ([`msescript.Muster`](../msescript/muster.go)), also musste hier ein anderer
// her — und der taugt besser: ein unbekannter Name ist der Fall, der in einem
// bezogenen Paket tatsächlich vorkommt, und er wird nicht eines Tages behoben.
func TestEinVorbehaltImInitskriptWirdGenannt(t *testing.T) {
	_, befunde, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), msepaket.Exportvorlage{
			KurzerName:  "V",
			Skript:      `"fertig"`,
			Spielskript: "x := gibt_es_nicht(\"a\")\ny := 1\n",
		}, nil, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "gibt_es_nicht") ||
		!strings.Contains(befunde[0].Ort, "Zeile 1") {
		t.Errorf("%+v", befunde[0])
	}
}
