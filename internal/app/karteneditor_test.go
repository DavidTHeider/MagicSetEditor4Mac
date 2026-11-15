package app_test

import (
	"context"
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schreibprobe merkt sich, was geschrieben wurde.
type schreibprobe struct {
	karteID int64
	feld    string
	wert    string
	fehler  error

	// feldherkuenfte hält, was eine KI-Übernahme mitschickte (A11, #1450).
	feldherkuenfte []domain.KIFeldherkunft

	// erwartet hält, welcher Stand als Bedingung ankam (Befund 2 des
	// Zweitblicks über Phase 5.3).
	erwartet []string
}

func (s *schreibprobe) FeldSetzenMitVorwert(
	_ context.Context, _, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	s.karteID, s.feld, s.wert = karteID, feld, wert

	return "", false, s.fehler
}

// FeldSetzenVonAussen schreibt wie oben und merkt sich die Herkunft (#888).
// FelderVonAussenMitHerkunft ist derselbe Weg samt Feldherkunft (A11, #1450).
func (s *schreibprobe) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	_ string, herkuenfte []domain.KIFeldherkunft,
	_ map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	s.feldherkuenfte = append(s.feldherkuenfte, herkuenfte...)
	s.erwartet = append(s.erwartet, erwartet)

	return s.FelderVonAussen(ctx, setID, karteID, werte, "ki")
}

func (s *schreibprobe) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	for _, wert := range werte {
		if _, _, fehler := s.FeldSetzenVonAussen(
			ctx, setID, karteID, wert.Field, wert.Text, herkunft); fehler != nil {
			return "", false, fehler
		}
	}

	return "", false, nil
}

func (s *schreibprobe) FeldSetzenVonAussen(
	ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	_ = herkunft

	return s.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
}

type karteprobe struct {
	karte  *domain.CardRow
	fehler error

	// **Die Herkünfte gehören zur Karte** (#1025): dieselbe Naht liest beides,
	// damit kein Aufrufer die Nennung vergessen kann.
	herkuenfte      map[string]string
	herkunftsfehler error
}

func (k karteprobe) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return k.karte, k.fehler
}

func (k karteprobe) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return k.herkuenfte, k.herkunftsfehler
}

func editor(schreiber *schreibprobe, karte karteprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(schreiber, karte, service.NeuerPlanbauer(),
		render.NeuerRasterer(ohneBilder{}))

	return anwendung
}

type ohneBilder struct{}

func (ohneBilder) Bild(string) (image.Image, error) { return nil, errors.New("keine Bildquelle") }

func TestFeldSetzenReichtDurch(t *testing.T) {
	schreiber := &schreibprobe{}

	if antwort := editor(schreiber, karteprobe{}).
		FeldSetzen(probekontext(t), "1", "7", "name", "Feuerball"); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if schreiber.karteID != 7 || schreiber.feld != "name" || schreiber.wert != "Feuerball" {
		t.Errorf("%+v", schreiber)
	}
}

// **Der Feldname wird geprüft, der Wert nicht.**
//
// Ein leerer Feldname ergäbe eine Zeile, die niemand wiederfindet; ein leerer
// Wert ist ein gültiger Wert — wer ein Feld leert, meint das.
func TestEinLeererFeldnameWirdAbgewiesenUndEinLeererWertNicht(t *testing.T) {
	schreiber := &schreibprobe{}
	anwendung := editor(schreiber, karteprobe{})

	for _, feld := range []string{"", "   "} {
		befund := anwendung.FeldSetzen(t.Context(), "1", "7", feld, "egal").Fehler
		if befund == nil {
			t.Fatalf("der Feldname %q kam durch", feld)
		}

		// **Der Fehler erscheint am verursachenden Feld** (§6, und die zweite
		// Zusage aus #40).
		if befund.Feld != "feld" || befund.Code != app.CodeValidierung {
			t.Errorf("%+v", befund)
		}
	}

	if antwort := anwendung.FeldSetzen(t.Context(), "1", "7", "name", ""); antwort.Fehler != nil {
		t.Errorf("ein leerer Wert wurde abgewiesen: %+v", antwort.Fehler)
	}
}

// Eine Karte, die es nicht gibt, kommt als „nicht gefunden" an — nicht als
// Konflikt. Das Frontend zeigt dann eine Meldung, statt nachzuladen.
func TestEineUnbekannteKarteKommtAlsNichtGefundenAn(t *testing.T) {
	schreiber := &schreibprobe{fehler: domain.ErrNotFound}

	befund := editor(schreiber, karteprobe{}).FeldSetzen(probekontext(t), "1", "7", "name", "x").Fehler
	if befund == nil || befund.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", befund)
	}
}

func TestFeldSetzenWeistEineUnlesbareKennungAb(t *testing.T) {
	befund := editor(&schreibprobe{}, karteprobe{}).
		FeldSetzen(probekontext(t), "1", "abc", "name", "x").Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// Die Vorschau gibt ein Bild, das ein `src` unmittelbar aufnimmt.
func TestDieVorschauGibtEinBild(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	antwort := editor(&schreibprobe{}, karte).Vorschau(probekontext(t), "1", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	vorschau := antwort.Vorschau

	// **Eine Adresse und keine Data-URI** (#1270) — dass sie zu den Bytes
	// führt, sagt [TestDieVorschauadresseFuehrtZumBild].
	if !strings.HasPrefix(vorschau.Bild, app.Bildwegpfad) {
		t.Errorf("%.40s …", vorschau.Bild)
	}

	if vorschau.Breite != int(service.Kartenbreite) {
		t.Errorf("Breite %d", vorschau.Breite)
	}

	// **Die Textebene kann noch nicht gezeichnet werden** (#105) — und das steht
	// als Befund da, statt still zu fehlen.
	if len(vorschau.Befunde) == 0 {
		t.Error("keine Befunde — die Textebene müsste sich melden")
	}
}

// Eine größere Vergrößerung ergibt ein größeres Bild aus **demselben** Plan.
func TestDieVorschauSkaliert(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{ID: 1}}
	anwendung := editor(&schreibprobe{}, karte)

	klein := anwendung.Vorschau(t.Context(), "1", 1).Vorschau
	gross := anwendung.Vorschau(t.Context(), "1", 2).Vorschau

	if gross.Breite != klein.Breite*2 {
		t.Errorf("%d gegen %d", gross.Breite, klein.Breite)
	}
}

func TestDieVorschauMeldetIhreFehler(t *testing.T) {
	faelle := map[string]struct {
		karte   karteprobe
		kennung string
		groesse float64
	}{
		"eine unlesbare Kennung": {karteprobe{}, "abc", 1},
		"eine Karte, die es nicht gibt": {
			karteprobe{fehler: domain.ErrNotFound}, "1", 1,
		},
		"eine unbrauchbare Vergrößerung": {
			karteprobe{karte: &domain.CardRow{ID: 1}}, "1", 0,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			befund := editor(&schreibprobe{}, fall.karte).
				Vorschau(probekontext(t), fall.kennung, fall.groesse).Fehler
			if befund == nil {
				t.Fatal("der Fehler ging verloren")
			}
		})
	}
}

// Eine laufende Vorschau lässt sich abbrechen (§7).
func TestDieVorschauLaesstSichAbbrechen(t *testing.T) {
	ctx, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	karte := karteprobe{karte: &domain.CardRow{ID: 1}}

	if editor(&schreibprobe{}, karte).Vorschau(ctx, "1", 1).Fehler == nil {
		t.Error("der Abbruch wurde übergangen")
	}
}

// GruppeVonAussenMitHerkunft gibt es an diesem Doppel **nur der
// Schnittstelle wegen** (A13.1b, #1452).
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine geschrieben hat — §3: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
func (*schreibprobe) GruppeVonAussenMitHerkunft(
	context.Context, []domain.Kartenschritt,
) error {
	return errors.New("dieses Doppel schreibt keine Gruppen")
}

// **Ohne einen Schreiber, der es sagt, bleibt das Gespeicherte offen** (#2535)
// — das Fenster bestätigt dann wie bisher, statt einen Wert anzunehmen.
func TestFeldSetzenOhneErgebnisNenntNichts(t *testing.T) {
	antwort := editor(&schreibprobe{}, karteprobe{}).
		FeldSetzen(probekontext(t), "1", "7", "name", "Feuerball")
	if antwort.Fehler != nil || antwort.Gespeichert != nil {
		t.Errorf("%+v", antwort)
	}
}
