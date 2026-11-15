package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sammelschreiber merkt sich **jede** Schreibung, nicht nur die letzte.
//
// **Das ist der Punkt dieses Tests**: es sind zwei, und die zweite ist die,
// die vergessen werden könnte.
type sammelschreiber struct {
	felder map[string]string
	folge  []string
	fehler error

	// zuege hält fest, **wie viele Aufrufe** es waren und was jeder trug.
	//
	// `folge` sagt das nicht: sie zählt Felder, nicht Züge. Zwischen zwei
	// Zügen liegt ein Stand, in dem das Bild an der Karte hängt und der Name
	// nicht — und genau den prüft `scryfallzug_test.go` (#1163, Z7-02).
	zuege [][]domain.Value

	// herkuenfte hält je Feld fest, woher der Wert kam (#888).
	herkuenfte map[string]string
}

func (s *sammelschreiber) FeldSetzenMitVorwert(
	_ context.Context, _, _ int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	if s.felder == nil {
		s.felder = map[string]string{}
	}

	s.felder[feld] = wert
	s.folge = append(s.folge, feld)

	return "", false, s.fehler
}

// FeldSetzenVonAussen schreibt wie oben und merkt sich die Herkunft (#888).
// FelderVonAussenMitHerkunft ist derselbe Weg samt Feldherkunft (A11, #1450).
func (s *sammelschreiber) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	_ string, _ []domain.KIFeldherkunft,
	_ map[string][]domain.KIBildparameter, _ string,
) (vorher string, hatte bool, fehler error) {
	return s.FelderVonAussen(ctx, setID, karteID, werte, "ki")
}

func (s *sammelschreiber) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	s.zuege = append(s.zuege, werte)

	for _, wert := range werte {
		if _, _, fehler := s.FeldSetzenVonAussen(
			ctx, setID, karteID, wert.Field, wert.Text, herkunft); fehler != nil {
			return "", false, fehler
		}
	}

	return "", false, nil
}

func (s *sammelschreiber) FeldSetzenVonAussen(
	ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	if s.herkuenfte == nil {
		s.herkuenfte = map[string]string{}
	}

	s.herkuenfte[feld] = herkunft

	return s.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
}

// bildholer ist ein Doppel für die Bildquelle der Fassade.
type bildholer struct {
	holung *service.Bilduebernahme
	fehler error
	name   string
	art    string
}

func (b *bildholer) Holen(_ context.Context, name, art string) (
	*service.Bilduebernahme, error,
) {
	b.name, b.art = name, art

	if art != "klein" && art != "normal" && art != "ausschnitt" {
		return nil, fmt.Errorf("%w: %q ist keine bekannte Bildart", domain.ErrValidation, art)
	}

	return b.holung, b.fehler
}

// scryfallbildaufbau stellt eine Anwendung samt Bildquelle bereit.
func scryfallbildaufbau(t *testing.T, holer *bildholer) (
	anwendung *app.App, schreiber *sammelschreiber, wurzel string,
) {
	t.Helper()

	ort := t.TempDir()
	gesammelt := &sammelschreiber{}

	gebaut := app.NeueApp("")
	gebaut.MitEditor(gesammelt,
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&bildplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	gebaut.MitBildablage(service.NeueBildablage(ort))
	gebaut.MitScryfallbildern(holer)

	return gebaut, gesammelt, ort
}

// einBildfund legt ein echtes JPEG ab und beschreibt es als Fund.
func einBildfund(t *testing.T, kuenstler string) *service.Bilduebernahme {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "kunst.jpg")
	schreibeJPEG(t, pfad)

	return &service.Bilduebernahme{
		Pfad: pfad, Art: scryfall.BildNormal, Kuenstler: kuenstler,
		Nennung: "Illustration: " + kuenstler + ". " + scryfall.Copyright,
		Adresse: "https://scryfall.com/card/lea/161",
	}
}

// **Der Satz, für den es diesen Test gibt**: ein Bild von Scryfall kommt nie
// ohne seinen Künstler in den Bestand. Die Bedingung verlangt, dass Künstler
// und Copyright in derselben Oberfläche stehen — und die Ausfuhr ist auch
// eine.
func TestScryfallBildSetztDenKuenstlerMit(t *testing.T) {
	holer := &bildholer{holung: einBildfund(t, "Christopher Rush")}
	anwendung, schreiber, wurzel := scryfallbildaufbau(t, holer)

	vorbereitet := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Lightning Bolt", "normal")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	if holer.name != "Lightning Bolt" || holer.art != "normal" {
		t.Errorf("geholt wurde %q in %q", holer.name, holer.art)
	}

	if vorbereitet.Nennung == "" || vorbereitet.Kuenstler != "Christopher Rush" {
		t.Errorf("die Nennung fehlt in der Antwort: %+v", vorbereitet)
	}

	antwort := anwendung.ScryfallBildEinsetzen(t.Context(), "1", "1", "kunst",
		vorbereitet.Pfad, "Christopher Rush", vorbereitet.Adresse, vorbereitet.Auswahl)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.felder["kunst"] != "kunst1" {
		t.Errorf("das Bildfeld steht auf %q", schreiber.felder["kunst"])
	}

	// **Die eigentliche Zusage.**
	if schreiber.felder["illustrator"] != "Christopher Rush" {
		t.Errorf("illustrator = %q, erwartet den Künstler — ohne ihn verlässt das "+
			"Bild die Anwendung ohne seine Nennung", schreiber.felder["illustrator"])
	}

	if _, err := os.Stat(filepath.Join(wurzel, "7", "kunst1")); err != nil {
		t.Errorf("das Bild fehlt: %v", err)
	}

	// **Und beide Zeilen tragen ihre Herkunft** (#888). Wer den Verlauf liest,
	// soll erkennen, dass hier nicht getippt, sondern übernommen wurde — bei
	// **beiden** Feldern, denn ein Bildname und ein Künstlername sehen danach
	// aus wie jede andere Eingabe.
	for _, feld := range []string{"kunst", "illustrator"} {
		if schreiber.herkuenfte[feld] != "https://scryfall.com/card/lea/161" {
			t.Errorf("Herkunft von %q = %q, erwartet die Kartenseite",
				feld, schreiber.herkuenfte[feld])
		}
	}
}

func TestScryfallBildOhneKuenstlerSchreibtKeinenIllustrator(t *testing.T) {
	// **Kein Künstler heißt kein leeres Feld.** Ein `illustrator: ""` würde
	// einen vorhandenen Namen löschen — das wäre eine Verschlechterung durch
	// eine Nennungspflicht.
	holer := &bildholer{holung: einBildfund(t, "")}
	anwendung, schreiber, _ := scryfallbildaufbau(t, holer)

	vorbereitet := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Namenlos", "normal")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	antwort := anwendung.ScryfallBildEinsetzen(
		probekontext(t), "1", "1", "kunst", vorbereitet.Pfad, "",
		vorbereitet.Adresse, vorbereitet.Auswahl)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if _, gesetzt := schreiber.felder["illustrator"]; gesetzt {
		t.Errorf("illustrator wurde auf %q gesetzt, obwohl kein Künstler bekannt ist",
			schreiber.felder["illustrator"])
	}
}

func TestScryfallBildMeldetFehlerUebersetzt(t *testing.T) {
	t.Run("ohne Zugang", func(t *testing.T) {
		anwendung := app.NeueApp("")

		antwort := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Egal", "normal")
		if antwort.Fehler == nil {
			t.Fatal("der fehlende Zugang blieb ohne Fehler")
		}
	})

	t.Run("Fehler des Abrufs", func(t *testing.T) {
		holer := &bildholer{fehler: domain.ErrNotFound}
		anwendung, _, _ := scryfallbildaufbau(t, holer)

		antwort := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Gibtesnicht", "normal")
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
			t.Fatalf("Fehler = %+v", antwort.Fehler)
		}
	})

	t.Run("unlesbare Kennung", func(t *testing.T) {
		holer := &bildholer{holung: einBildfund(t, "Wer auch immer")}
		anwendung, _, _ := scryfallbildaufbau(t, holer)

		antwort := anwendung.ScryfallBild(t.Context(), "keine Zahl", "kunst", "Egal", "normal")
		if antwort.Fehler == nil {
			t.Fatal("die unlesbare Kennung blieb ohne Fehler")
		}
	})

	t.Run("unbekannte Bildart", func(t *testing.T) {
		holer := &bildholer{holung: einBildfund(t, "Wer auch immer")}
		anwendung, _, _ := scryfallbildaufbau(t, holer)

		antwort := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Egal", "riesig")
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung {
			t.Fatalf("Fehler = %+v", antwort.Fehler)
		}
	})

	t.Run("Fehler beim Einsetzen", func(t *testing.T) {
		holer := &bildholer{holung: einBildfund(t, "Wer auch immer")}
		anwendung, schreiber, _ := scryfallbildaufbau(t, holer)
		schreiber.fehler = domain.ErrNotFound

		vorbereitet := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Egal", "normal")

		antwort := anwendung.ScryfallBildEinsetzen(t.Context(), "1", "1", "kunst",
			vorbereitet.Pfad, "Wer auch immer", vorbereitet.Adresse, vorbereitet.Auswahl)
		if antwort.Fehler == nil {
			t.Fatal("der Schreibfehler blieb ohne Meldung")
		}
	})
}

// **Ein unlesbares Bild kommt bis an die Fläche durch** —
// #1005.
//
// # Warum das ein eigener Fall ist
//
// Der Bildabruf kann glücken und die Datei trotzdem unbrauchbar sein — eine
// Weiterleitung auf eine Fehlerseite etwa liefert 200 und HTML. **Dann darf
// kein leerer Zuschnitt aufgehen**, sondern die Meldung muss ankommen.
//
// Der Fall ist nicht erfunden: beim Bauen dieses Wegs stand hier ein
// base64-Block, den ich für ein PNG hielt. `image.Decode` sagte dazu „not
// enough pixel data" — und die Meldung, die dabei herauskam, benannte nichts.
func TestScryfallBildWeistEinUnlesbaresBildAb(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "keinbild.png")
	if fehler := os.WriteFile(pfad, []byte("<html>Fehlerseite</html>"), 0o600); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	holer := &bildholer{holung: &service.Bilduebernahme{
		Pfad: pfad, Art: scryfall.BildNormal, Kuenstler: "Wer auch immer",
		Nennung: "Illustration: Wer auch immer. " + scryfall.Copyright,
	}}

	anwendung, _, _ := scryfallbildaufbau(t, holer)

	antwort := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Kaputt", "normal")
	if antwort.Fehler == nil {
		t.Fatal("die unlesbare Datei kam durch")
	}

	if antwort.Pfad != "" {
		t.Errorf("es wurde ein Zuschnitt geöffnet: %q", antwort.Pfad)
	}
}

// GruppeVonAussenMitHerkunft gibt es an diesem Doppel **nur der
// Schnittstelle wegen** (A13.1b, #1452).
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine geschrieben hat — §3: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
func (*sammelschreiber) GruppeVonAussenMitHerkunft(
	context.Context, []domain.Kartenschritt,
) error {
	return errors.New("dieses Doppel schreibt keine Gruppen")
}
