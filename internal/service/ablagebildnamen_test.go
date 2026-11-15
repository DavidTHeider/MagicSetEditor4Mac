package service_test

// **Was ein Bildname aus der Zwischenablage sein darf** —
// #1282,
// Befund 3.
//
// **Eine eigene Datei**, weil `ablageruecknahme_test.go` mit ihr über 400
// Zeilen ginge (§2). Der Schnitt folgt der Frage: dort steht, **was nicht
// liegen bleibt**, hier, **was gar nicht erst hineinkommt**.

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein fremder Bildname bricht nicht aus dem Zielverzeichnis aus** —
// #1282,
// Befund 3.
//
// # Warum das ein Sicherheitsbefund ist
//
// Der Name eines `image_data`-Blocks kommt aus der **Zwischenablage** — also
// von außen. `namenBelegen` verbindet ihn unmittelbar mit dem Zielpfad, und
// `filepath.Join` räumt `..` dabei auf, statt es abzuweisen: aus
// `image_data: ../1/bild <base64>` wird eine Datei im Verzeichnis eines
// **anderen** Sets.
//
// Der `#nosec`-Vermerk an jener Stelle sagt, der Name entstehe intern. Für
// diesen Aufrufer stimmte das nicht.
//
// **Verdeckt wurde es davon, dass die Naht gar nicht verdrahtet war** (Befund
// 1) — mit ihr ist der Weg unmittelbar erreichbar.
//
// # Zwei Dinge, die dieser Lauf braucht
//
// **Das Nachbarverzeichnis muss es geben.** Ohne es scheitert das Schreiben an
// „no such file or directory", und der Lauf wäre grün, ohne etwas geprüft zu
// haben — der erste Anlauf war genau das.
//
// **Und der Suchbegriff darf nicht im Pfad stehen.** `t.TempDir()` nimmt den
// Namen des Unterlaufs in den Pfad auf; wer nach ihm sucht, findet ihn immer.
func TestEinFremderBildnameBrichtNichtAus(t *testing.T) {
	t.Parallel()

	// **Ohne Sonderzeichen im Namen des Unterlaufs**, siehe oben.
	for nummer, name := range []string{
		"../1/schmuggel",
		"..\\1\\schmuggel",
		"/tmp/schmuggel",
		"unter/schmuggel",
	} {
		t.Run(strconv.Itoa(nummer), func(t *testing.T) {
			t.Parallel()

			vor := neueBildvorrichtung(t)
			vor.karteMitWert("image", "image1")

			// **Das Nachbarverzeichnis anlegen** — im Betrieb ist es da,
			// sobald das erste Set ein Bild führt.
			nachbar := vor.bilder.Verzeichnis(1)
			if err := os.MkdirAll(nachbar, 0o700); err != nil {
				t.Fatalf("Nachbarverzeichnis: %v", err)
			}

			inhalt := "game_name: probe\ncard:\n\timage: image1\n" +
				"image_data: " + name + " " +
				base64.StdEncoding.EncodeToString([]byte("keinBild")) + "\n"

			einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt)

			// **Abgewiesen wird das Bild, nicht die Karte** (§8): ein Inhalt
			// mit einem unbrauchbaren Bild wird mit einem Befund eingefügt,
			// nicht verworfen.
			if fehler != nil {
				t.Fatalf("Einfuegen: %v", fehler)
			}

			if len(einfuegung.Befunde) == 0 {
				t.Errorf("kein Befund zu %q — der Name ist stillschweigend "+
					"angenommen worden", name)
			}

			ziel := vor.bilder.Verzeichnis(2)

			for _, ort := range []string{nachbar, "/tmp"} {
				eintraege, err := os.ReadDir(ort)
				if err != nil {
					continue
				}

				for _, eintrag := range eintraege {
					if strings.HasPrefix(eintrag.Name(), "schmuggel") {
						t.Errorf("%q ist in %q entstanden — der Name %q ist "+
							"aus %q ausgebrochen",
							eintrag.Name(), ort, name, ziel)
					}
				}
			}

			// **Und im Zielverzeichnis auch kein Unterverzeichnis.**
			eintraege, err := os.ReadDir(ziel)
			if err != nil {
				return
			}

			for _, eintrag := range eintraege {
				if eintrag.IsDir() {
					t.Errorf("%q ist im Zielverzeichnis entstanden — ein "+
						"Bildname legt kein Verzeichnis an", eintrag.Name())
				}
			}
		})
	}
}

// **Ein Punkt ist kein Dateiname** — die Grenzfälle von `istEinfacherName`.
func TestEinPunktIstKeinBildname(t *testing.T) {
	t.Parallel()

	for nummer, name := range []string{"", ".", ".."} {
		t.Run(strconv.Itoa(nummer), func(t *testing.T) {
			t.Parallel()

			vor := neueBildvorrichtung(t)
			vor.karteMitWert("image", "image1")

			inhalt := "game_name: probe\ncard:\n\timage: image1\n" +
				"image_data: " + name + " " +
				base64.StdEncoding.EncodeToString([]byte("keinBild")) + "\n"

			einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt)
			if fehler != nil {
				t.Fatalf("Einfuegen: %v", fehler)
			}

			// **Der leere Name kommt gar nicht bis zur Prüfung** — `Cut`
			// findet dann kein Leerzeichen, und der Block wird übersprungen.
			// Bei `.` und `..` steht ein Befund da.
			if name != "" && len(einfuegung.Befunde) == 0 {
				t.Errorf("kein Befund zu %q", name)
			}

			if len(vor.bilderImZielset(t)) != 0 {
				t.Errorf("es liegt eine Datei da: %v", vor.bilderImZielset(t))
			}
		})
	}
}

// **Ein Bildname, den `Bildablage.Bytes` selbst abweist** — die zweite
// Grenze, eine Ebene tiefer (#1282,
// Befund 3).
//
// `ablagebilder.go` weist einen Weg schon vorher ab. Diese Zeile prüft die
// Grenze **selbst**: sie gilt auch für den nächsten Aufrufer, der einen Namen
// von außen bekommt und nicht daran denkt.
func TestDieBildablageWeistEinenWegAb(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)

	for nummer, name := range []string{"unter/tief", "..", ""} {
		t.Run(strconv.Itoa(nummer), func(t *testing.T) {
			t.Parallel()

			if _, fehler := vor.bilder.Bytes(2, name, []byte("x")); fehler == nil {
				t.Errorf("%q wurde angenommen", name)
			}
		})
	}
}

// **Ein Bildverweis ohne Datei wird benannt** — im laufenden Fenster gesehen
// am 2026-08-28.
//
// Die Bilder eines geöffneten Sets liegen in seinem Herkunftspaket und nicht
// in dieser Anwendung ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
// Der Verweis geht allein mit, und `image0` heißt im Zielset etwas anderes —
// dort stünde dann ein **fremdes** Bild. Das darf nicht still geschehen.
func TestEinVerweisOhneBildWirdBenannt(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.ablage = vor.ablage.MitFeldern(felderdoppel{})

	// **Eine Karte mit Bildwert, aber ohne abgelegte Datei** — genau der Fall
	// eines Sets, dessen Bilder im Paket liegen.
	vor.karteMitWert("image", "image0")

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if strings.Contains(inhalt, "image_data") {
		t.Fatalf("es ist eine Datei mitgekommen, obwohl keine dalag:\n%s", inhalt)
	}

	einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt)
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(einfuegung.Befunde) == 0 {
		t.Fatal("kein Befund — der Verweis ist still durchgegangen")
	}

	if !strings.Contains(einfuegung.Befunde[0].Meldung, "image0") {
		t.Errorf("der Befund nennt das Bild nicht: %v", einfuegung.Befunde)
	}
}

// **Und wenn die Datei mitkommt, steht kein Befund da.**
func TestEinMitgekommenesBildGibtKeinenBefund(t *testing.T) {
	t.Parallel()

	vor, _ := bildKarte(t)
	vor.ablage = vor.ablage.MitFeldern(felderdoppel{})

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt)
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(einfuegung.Befunde) != 0 {
		t.Errorf("Befunde, obwohl das Bild mitkam: %v", einfuegung.Befunde)
	}
}

// **Was das Zielset nicht speichert, kommt nicht mit — und wird benannt.**
//
// Ein `.mse-set` speichert `time created`; die Vorlage rechnet es, und
// `FelderVonAussen` weist jeden Schreibversuch ab. Ohne diese Wache scheiterte
// **jedes** Einfügen einer echten Karte und ließ eine leere zurück — im
// laufenden Fenster nachgestellt am 2026-08-28
// (#1282).
func TestEinBerechnetesFeldKommtNichtMit(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	doppel.karten[1][0].Values = append(doppel.karten[1][0].Values,
		domain.Value{Field: "time created", Text: "2020-01-01"})

	ablagen := kartenablage(doppel).MitFeldern(felderdoppel{})

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	einfuegung, fehler := ablagen.Einfuegen(t.Context(), 2, inhalt)
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	for _, wert := range doppel.werte[einfuegung.Karten[0]] {
		if wert.Field == "time created" {
			t.Errorf("`time created` wurde geschrieben — die Vorlage rechnet es")
		}
	}

	if len(einfuegung.Befunde) == 0 {
		t.Error("kein Befund — das übergangene Feld ist still verschwunden")
	}
}

// **Ohne Feldquelle geht alles durch** — die Wahlfreiheit, wie bei den Bildern.
func TestOhneFeldquelleGehtAllesDurch(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	doppel.karten[1][0].Values = append(doppel.karten[1][0].Values,
		domain.Value{Field: "time created", Text: "2020-01-01"})

	ablagen := kartenablage(doppel)

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	einfuegung, fehler := ablagen.Einfuegen(t.Context(), 2, inhalt)
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(einfuegung.Befunde) != 0 {
		t.Errorf("Befunde ohne Feldquelle: %v", einfuegung.Befunde)
	}
}

// **Und ein Fehler beim Lesen der Felder hält nichts auf.**
//
// Er hieße, dass die Vorlage nicht lesbar ist; dann weiß niemand, was
// gespeichert wird, und der Schreiber entscheidet es wie bisher.
func TestEinFehlerBeimFeldlesenHaeltDasEinfuegenNichtAuf(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	ablagen := kartenablage(doppel).MitFeldern(
		felderdoppel{fehler: errors.New("keine Vorlage")})

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if _, fehler := ablagen.Einfuegen(t.Context(), 2, inhalt); fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}
}

// felderdoppel sagt, welche Felder ein Set speichert.
type felderdoppel struct{ fehler error }

func (f felderdoppel) Kartenfelder(
	context.Context, int64,
) ([]domain.Field, error) {
	return []domain.Field{
		{Name: "name", Gespeichert: true},
		{Name: "rule text", Gespeichert: true},
		{Name: "image", Kind: domain.KindImage, Gespeichert: true},
		{Name: "time created", Gespeichert: false},
	}, f.fehler
}
