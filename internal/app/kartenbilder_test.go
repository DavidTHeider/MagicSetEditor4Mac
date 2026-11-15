// **Die Vorschaubilder der Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
package app_test

import (
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// listenbildquelle gibt zu jedem Verweis ein Bild und merkt sich die Verweise.
type listenbildquelle struct{ gefragt []string }

func (l *listenbildquelle) Bild(pfad string) (image.Image, error) {
	l.gefragt = append(l.gefragt, pfad)

	bild := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	bild.Set(0, 0, color.NRGBA{G: 255, A: 255})

	return bild, nil
}

// listenbestand gibt drei Karten, von denen zwei ein Bild tragen.
func listenbestand() leseprobe {
	return leseprobe{
		koepfe: []domain.SetHead{{ID: 7, Origin: "/wo/das/set.mse-set"}},
		karten: []*domain.CardRow{
			{ID: 1, Number: 1, Values: []domain.Value{{Field: "kunst", Text: "eins.png"}}},
			{ID: 2, Number: 2, Values: []domain.Value{{Field: "kunst", Text: ""}}},
			{ID: 3, Number: 3, Values: []domain.Value{{Field: "kunst", Text: "drei.png"}}},
		},
	}
}

// listenfassade baut eine Fassade mit Bildfeld, Bestand und Bildquelle.
func listenfassade(quelle render.Bildquelle) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitBestand(listenbestand())
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		// **Ein Wahlfeld davor**, damit „das erste Bildfeld" wirklich geprüft
		// wird und nicht „das erste Feld".
		{Name: "rarity", Kind: domain.KindChoice},
		{Name: "kunst", Kind: domain.KindImage},
		{Name: "zweitbild", Kind: domain.KindImage},
	}})
	anwendung.MitVorschaubildern(service.NeuesVorschaubildwerk(quelle))

	return anwendung
}

// **Gefragte Karten mit Bild bekommen eines, die anderen nicht.**
func TestDieKartenlisteBekommtIhreBilder(t *testing.T) {
	t.Parallel()

	quelle := &listenbildquelle{}

	antwort := listenfassade(quelle).Kartenbilder(probekontext(t), "7", []string{"1", "2"})
	if antwort.Fehler != nil {
		t.Fatalf("Kartenbilder: %+v", antwort.Fehler)
	}

	// **Zwei gefragt, eines geliefert.** Karte 2 trägt keinen Dateinamen, und
	// Karte 3 war nicht gefragt — beides muss fehlen, und aus verschiedenen
	// Gründen.
	if len(antwort.Bilder) != 1 {
		t.Fatalf("es sind %d Bilder: %v", len(antwort.Bilder), antwort.Bilder)
	}

	if !strings.HasPrefix(antwort.Bilder["1"], "data:image/") {
		t.Errorf("das Bild ist kein Daten-URI: %.40s", antwort.Bilder["1"])
	}

	// **Und gefragt wurde genau einmal, unter dem ersten Bildfeld.**
	if len(quelle.gefragt) != 1 ||
		quelle.gefragt[0] != "setpaket:/wo/das/set.mse-set!eins.png" {
		t.Errorf("gefragt wurde %v", quelle.gefragt)
	}
}

// **Ohne Werk gibt es keine Bilder und keinen Fehler.**
//
// Die Kommandozeile hat kein Zeichenwerk; eine Liste ohne Bilder ist dort die
// richtige Antwort und keine Störung.
func TestOhneVorschauwerkGibtEsKeineBilder(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitBestand(listenbestand())

	antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"})

	if antwort.Fehler != nil || len(antwort.Bilder) != 0 {
		t.Errorf("es kam %+v", antwort)
	}
}

// **Eine leere Frage fragt den Bestand gar nicht.**
//
// Die Liste ruft bei jedem Rollen; solange sie noch nichts zeichnet, ist die
// richtige Antwort keine Abfrage.
func TestEineLeereFrageBleibtOhneBestandszugriff(t *testing.T) {
	t.Parallel()

	quelle := &listenbildquelle{}

	antwort := listenfassade(quelle).Kartenbilder(probekontext(t), "7", nil)

	if antwort.Fehler != nil || len(antwort.Bilder) != 0 {
		t.Errorf("es kam %+v", antwort)
	}

	if len(quelle.gefragt) != 0 {
		t.Errorf("die Bildquelle wurde gefragt: %v", quelle.gefragt)
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen.**
func TestKartenbilderWeisenEineKaputteKennungAb(t *testing.T) {
	t.Parallel()

	antwort := listenfassade(&listenbildquelle{}).
		Kartenbilder(probekontext(t), "kein Set", []string{"1"})

	if antwort.Fehler == nil {
		t.Error("die kaputte Kennung kam durch")
	}
}

// **Ohne Bildfeld in der Vorlage bleibt die Liste Text.**
//
// Nicht jedes Spiel hat eines — und MSE2 gibt dann ebenfalls kein Bild aus
// (`ImageCardList::findImageField` liefert einen leeren Zeiger,
// `OnGetItemImage` antwortet `-1`).
func TestOhneBildfeldGibtEsKeineListenbilder(t *testing.T) {
	t.Parallel()

	quelle := &listenbildquelle{}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(listenbestand())
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "name", Kind: domain.KindText},
	}})
	anwendung.MitVorschaubildern(service.NeuesVorschaubildwerk(quelle))

	antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"})

	if antwort.Fehler != nil || len(antwort.Bilder) != 0 {
		t.Errorf("es kam %+v", antwort)
	}

	if len(quelle.gefragt) != 0 {
		t.Errorf("die Bildquelle wurde gefragt: %v", quelle.gefragt)
	}
}

// **Ein Bestandsfehler kommt an.**
//
// Er ist die eine Lage, in der die Liste **nichts** zeigen kann und es sagen
// muss: ohne Karten gibt es auch keine Namen daneben.
func TestEinBestandsfehlerErreichtDieListenbilder(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitBestand(leseprobe{fehler: errStolpert})
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "kunst", Kind: domain.KindImage},
	}})
	anwendung.MitVorschaubildern(service.NeuesVorschaubildwerk(&listenbildquelle{}))

	if antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"}); antwort.Fehler == nil {
		t.Error("der Bestandsfehler kam nicht an")
	}
}

// **Und ein Bild, das nicht zu holen ist, lässt die übrigen stehen** (§8).
//
// Eine Liste, die wegen **einer** kaputten Datei gar nichts zeigte, wäre der
// schlechtere Tausch — und der Zeichenweg meldet denselben Fehler dort, wo er
// etwas bedeutet.
func TestEinKaputtesBildLaesstDieUebrigenStehen(t *testing.T) {
	t.Parallel()

	antwort := listenfassade(&stolperndeBildquelle{}).
		Kartenbilder(probekontext(t), "7", []string{"1", "3"})

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 0 {
		t.Errorf("es kamen Bilder: %v", antwort.Bilder)
	}
}

// stolperndeBildquelle scheitert an jedem Verweis.
type stolperndeBildquelle struct{}

func (stolperndeBildquelle) Bild(string) (image.Image, error) { return nil, errStolpert }

// errStolpert ist der Fehler dieser Datei.
var errStolpert = errors.New("das Bild ist nicht zu holen")

// **Eine Karte ohne den gesuchten Feldnamen bekommt kein Bild.**
//
// Der Fall entsteht bei einem **Stilwechsel**: die Vorlage nennt jetzt ein
// anderes Bildfeld, und die vorhandenen Karten tragen den alten Namen weiter.
// Sie müssen dann bildlos in der Liste stehen und nicht das Bild einer anderen
// Karte zeigen.
func TestEineKarteOhneDasBildfeldBekommtNichts(t *testing.T) {
	t.Parallel()

	quelle := &listenbildquelle{}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(leseprobe{
		koepfe: []domain.SetHead{{ID: 7, Origin: "/wo/das/set.mse-set"}},
		karten: []*domain.CardRow{{
			ID: 1, Number: 1,
			Values: []domain.Value{{Field: "altesfeld", Text: "eins.png"}},
		}},
	})
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "kunst", Kind: domain.KindImage},
	}})
	anwendung.MitVorschaubildern(service.NeuesVorschaubildwerk(quelle))

	antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"})

	if antwort.Fehler != nil || len(antwort.Bilder) != 0 {
		t.Errorf("es kam %+v", antwort)
	}

	if len(quelle.gefragt) != 0 {
		t.Errorf("die Bildquelle wurde gefragt: %v", quelle.gefragt)
	}
}

// **Ohne Bilderort gibt es keine Ablage — und trotzdem Bilder.**
//
// Die Ablage ist der Ort der Bilder, die jemand **in dieser Anwendung**
// eingesetzt hat. Gibt es keinen, liegt jedes Bild im Herkunftspaket, und dort
// wird gefragt. Ein Fehler wäre hier falsch: die Kommandozeile und jede frisch
// eingerichtete Anwendung sind genau in dieser Lage.
func TestOhneBilderortWirdImPaketGefragt(t *testing.T) {
	t.Parallel()

	quelle := &listenbildquelle{}

	anwendung := listenfassade(quelle)
	anwendung.MitBilderort(func() (string, error) { return "", errStolpert })

	if antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(quelle.gefragt) != 1 ||
		quelle.gefragt[0] != "setpaket:/wo/das/set.mse-set!eins.png" {
		t.Errorf("gefragt wurde %v", quelle.gefragt)
	}
}

// **Und mit einem Bilderort wird die Ablage zuerst gefragt.**
//
// Die Gegenprobe zum Lauf darüber: ohne sie wäre nicht zu sehen, ob
// `ablagenort` überhaupt etwas tut — beide Wege endeten im Paket.
func TestMitBilderortEntstehtEinAblagenort(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	if err := os.MkdirAll(filepath.Join(wurzel, "7"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(wurzel, "7", "eins.png"), []byte("x"), 0o600); err != nil {
		t.Fatalf("hinlegen: %v", err)
	}

	quelle := &listenbildquelle{}

	anwendung := listenfassade(quelle)
	anwendung.MitBilderort(func() (string, error) { return wurzel, nil })

	if antwort := anwendung.Kartenbilder(t.Context(), "7", []string{"1"}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(quelle.gefragt) != 1 ||
		quelle.gefragt[0] != "setpaket:"+filepath.Join(wurzel, "7")+"!eins.png" {
		t.Errorf("gefragt wurde %v", quelle.gefragt)
	}
}
