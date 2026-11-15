// Package mtgjsonpreise liest die Preise, die MTGJSON führt.
//
// **Werkstück A3 von Phase 5.3**
// (#1442).
//
// # Warum ein eigenes Paket neben `internal/mtgjson`
//
// `internal/mtgjson` liest und schreibt das **Dateiformat** MTG JSON — eine
// Datei, die jemand mitbringt, und was `magic-set-generator` liest. Hier geht
// es um den **Dienst** unter <https://mtgjson.com>: ein anderer Vertrag, ein
// anderer Transport, andere Grenzen. Das Issue verlangt die Trennung
// ausdrücklich („der Netzadapter/Bulkcache bekommt einen klaren Ort mit
// Zuständigkeitsnachweis, statt eine vorhandene Produktionsanbindung zu
// behaupten").
//
// # Die Gestalt, die gelesen wird
//
// **Nachgelesen am 2026-09-11** unter <https://mtgjson.com/data-models/price/>
// samt `price-formats`, `price-list` und `price-points`:
//
//	"<uuid>": { "paper": { "<anbieter>": { "currency": "USD",
//	                                       "retail":  { "normal": {"2026-09-10": 4.17} },
//	                                       "buylist": { … } } },
//	            "mtgo":  { "cardhoarder": … } }
//
// Fünf Ebenen unter der Kennung, und **jede** davon gehört zur Reihe
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).
//
// # Was hier nicht steht
//
// **Der Crosswalk.** Welche MTGJSON-Kennung zu welchem Druck dieses Bestands
// gehört, weiß der Bestand — dieses Paket ist ein Blatt (§2) und bekommt die
// Zuordnung als Karte übergeben. Was nicht darin steht, wird übersprungen:
// die Datei führt jede Karte, die es gibt, und der Bestand nur die, die
// jemand besitzt.
package mtgjsonpreise

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Grenzen sind die Schranken, unter denen gelesen wird.
//
// **Null heißt: die Vorgabe** — nicht „unbegrenzt". Eine fremde Datei ohne
// Deckel ist der Weg, auf dem ein Gigabyte in den Arbeitsspeicher kommt (§16).
type Grenzen struct {
	// HoechstensBytes ist der Deckel auf dem **ausgepackten** Strom.
	HoechstensBytes int64

	// HoechstensStaende ist der Deckel auf dem Ergebnis.
	HoechstensStaende int
}

// Die Vorgaben, wenn eine Grenze nicht genannt ist.
//
// **Zwei Gigabyte ausgepackt**: `AllPrices.json` führt neunzig Tage und wiegt
// nach Angabe des Anbieters über ein Gigabyte. Die Grenze liegt darüber, damit
// eine gewachsene Datei nicht am Deckel scheitert — und weit genug darunter,
// dass ein Packbomben-Strom nicht durchläuft.
//
// **Zwei Millionen Stände**: bei rund hunderttausend Drucken im Bestand und
// zwanzig Reihen je Druck ist das die Größenordnung, und mehr wäre keine
// Sammlung mehr, sondern ein Katalog.
const (
	VorgabeBytes   int64 = 2 << 30
	VorgabeStaende       = 2_000_000
)

// mitVorgaben füllt, was nicht genannt ist.
func (g Grenzen) mitVorgaben() Grenzen {
	if g.HoechstensBytes <= 0 {
		g.HoechstensBytes = VorgabeBytes
	}

	if g.HoechstensStaende <= 0 {
		g.HoechstensStaende = VorgabeStaende
	}

	return g
}

// LiesGepackt liest einen gzip-Strom.
//
// **Der Deckel liegt auf dem ausgepackten Strom**, nicht auf dem gepackten:
// zehn Megabyte gzip können ein Terabyte werden, und eine Grenze auf der
// gepackten Seite wäre keine.
func LiesGepackt(quelle io.Reader, crosswalk map[string]string, grenzen Grenzen) (
	[]domain.Preisstand, []domain.Befund, error,
) {
	entpackt, fehler := gzip.NewReader(quelle)
	if fehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Preisdatei ist kein gzip")
	}

	defer func() { _ = entpackt.Close() }() //nolint:errcheck // gelesen ist gelesen

	return Lies(entpackt, crosswalk, grenzen)
}

// Lies liest den Strom und gibt die Stände der genannten Kennungen.
//
// **Strömend und nicht in einem Zug** (§16): die Datei wiegt ein Gigabyte, und
// `json.Unmarshal` legte sie zweimal in den Arbeitsspeicher — einmal als Bytes
// und einmal als Baum.
//
// **Zahlen als Text.** `json.Decoder.UseNumber` gibt die Ziffernfolge, wie sie
// dasteht; sie geht durch [domain.MarktpreisAus] und wird zur kleinsten Einheit. Ein
// `float64` dazwischen wäre Rundungsdrift in einer Geldrechnung — und genau
// das verbietet A3.
func Lies(quelle io.Reader, crosswalk map[string]string, grenzen Grenzen) (
	[]domain.Preisstand, []domain.Befund, error,
) {
	grenzen = grenzen.mitVorgaben()

	leser := &zaehlenderLeser{unten: quelle, grenze: grenzen.HoechstensBytes}

	zerteiler := json.NewDecoder(leser)
	zerteiler.UseNumber()

	staende, befunde, fehler := bisZuDenKarten(zerteiler, crosswalk, grenzen)

	// **Der Deckel schlägt jede andere Meldung.**
	//
	// Der Zerteiler sieht einen abgeschnittenen Strom und sagt „das fängt
	// nicht mit einem Objekt an" — wahr, und die falsche Auskunft: abgeschnitten
	// hat **ihn** der Deckel. Wer die äußere Meldung läse, suchte einen Fehler
	// in der Datei, den es nicht gibt.
	if leser.zuGross != nil {
		return nil, nil, leser.zuGross
	}

	return staende, befunde, fehler
}

// bisZuDenKarten geht zum Feld `data` und liest die Kennungen darunter.
func bisZuDenKarten(
	zerteiler *json.Decoder, crosswalk map[string]string, grenzen Grenzen,
) ([]domain.Preisstand, []domain.Befund, error) {
	if fehler := bisZumFeld(zerteiler, "data"); fehler != nil {
		return nil, nil, fehler
	}

	return karten(zerteiler, crosswalk, grenzen)
}

// zaehlenderLeser hält den Deckel auf dem ausgepackten Strom.
type zaehlenderLeser struct {
	unten  io.Reader
	grenze int64
	gerade int64

	// zuGross ist der Fehler, sobald der Deckel greift.
	zuGross error
}

func (z *zaehlenderLeser) Read(puffer []byte) (int, error) {
	gelesen, fehler := z.unten.Read(puffer)

	z.gerade += int64(gelesen)

	// **Null Bytes zurück, nicht die gelesenen.**
	//
	// Der erste Anlauf gab beides — die Bytes **und** den Fehler —, und
	// `json.Decoder` nahm die Bytes: passte das ganze Dokument in einen
	// einzigen Lesevorgang, war es damit vollständig, und der Fehler wurde nie
	// gefragt. Der Deckel war da und griff nicht (§10: eine Zusage, die
	// niemand verletzen kann, ist keine geprüfte Zusage).
	if z.gerade > z.grenze {
		z.zuGross = fmt.Errorf(
			"%w: die Preisdatei ist über %d Byte groß — so viel führt MTGJSON nicht",
			domain.ErrValidation, z.grenze)

		return 0, z.zuGross
	}

	return gelesen, fehler
}

// bisZumFeld liest bis zum Anfang des genannten Objektfelds.
//
// **Und nicht bis zum ersten Vorkommen des Wortes**: `meta` steht davor, und
// ein Wert darin könnte „data" heißen. Gezählt wird die Tiefe.
func bisZumFeld(zerteiler *json.Decoder, feld string) error {
	if _, fehler := zerteiler.Token(); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Preisdatei fängt nicht mit einem Objekt an")
	}

	for zerteiler.More() {
		marke, fehler := zerteiler.Token()
		if fehler != nil {
			return domain.MitGrund(domain.ErrValidation, fehler,
				"die Preisdatei ist unlesbar")
		}

		if name, istText := marke.(string); istText && name == feld {
			return objektAnfangen(zerteiler, feld)
		}

		// **Der Wert wird übersprungen**, nicht gelesen: `meta` ist klein,
		// aber die Regel gilt für jedes Feld, das vor `data` stehen könnte.
		var weg json.RawMessage
		if fehler := zerteiler.Decode(&weg); fehler != nil {
			return domain.MitGrund(domain.ErrValidation, fehler,
				"die Preisdatei ist unlesbar")
		}
	}

	return fmt.Errorf("%w: die Preisdatei nennt kein %s", domain.ErrValidation, feld)
}

// objektAnfangen verlangt, dass hinter dem Feld ein Objekt steht.
//
// **Und prüft es wirklich.** Der erste Anlauf las die Marke und gab zurück;
// bei `{"data": 42}` war das die Zahl, und danach meldete der Zerteiler
// wahrheitsgemäß „nichts mehr da". Der Lauf war grün und hatte **null**
// Kennungen gelesen — ein leerer Bestand sieht genauso aus.
func objektAnfangen(zerteiler *json.Decoder, feld string) error {
	marke, fehler := zerteiler.Token()
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Preisdatei nennt %s ohne Inhalt", feld)
	}

	if auf, istMarke := marke.(json.Delim); !istMarke || auf != '{' {
		return fmt.Errorf("%w: die Preisdatei führt unter %s kein Objekt",
			domain.ErrValidation, feld)
	}

	return nil
}
