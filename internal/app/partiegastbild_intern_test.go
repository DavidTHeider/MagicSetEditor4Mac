package app

// **Am Gast keine Zeile des Wirts** — P9
// (#2377),
// Befund C-12, Vermerk 0198: zwei Anwendungen, der Gast sitzt am Tisch des
// Wirts mit einem Deck aus dessen Bestand. Seine Hand sieht er — aber die
// Kartenkennungen darin sind Zeilen **des Wirts**; in seinem eigenen Bestand
// nachgeschlagen, fänden sie eine andere Karte. Er bekommt deshalb keine, und
// an ihrer Stelle die Bildkennung.

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

func TestDerGastSiehtKeineZeileDesWirts(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID := einTischMitEigenenKarten(t, wirt)
	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	bilder := bilderAmGast(gast)
	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	if antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), "", PartiemitbringDTO{}); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	select {
	case bild := <-bilder:
		gesehen := 0

		for _, objekt := range bild.Partie.Objekte {
			if objekt.Geheim || objekt.Karte == "" {
				continue
			}

			gesehen++

			if objekt.KarteID != "" {
				t.Errorf("%s trägt die Zeile %s des Wirts", objekt.Kennung, objekt.KarteID)
			}

			if objekt.Bild == "" {
				t.Errorf("%s trägt keine Bildkennung", objekt.Kennung)
			}
		}

		if gesehen == 0 {
			t.Fatal("der Gast sieht keine einzige Karte — der Lauf prüfte nichts")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("kein Tischbild am Gast")
	}
}

// bilderAmGast fängt die Tischbilder, die das Fenster des Gastes bekommt.
func bilderAmGast(wer *seite) chan PartieAntwort {
	bilder := make(chan PartieAntwort, 64)
	vorher := wer.anwendung.melden

	wer.anwendung.melden = func(ctx context.Context, ereignis string, daten ...any) {
		vorher(ctx, ereignis, daten...)

		if ereignis != EreignisPartietischbild || len(daten) < 2 {
			return
		}

		if bild, passt := daten[1].(PartieAntwort); passt {
			select {
			case bilder <- bild:
			default:
			}
		}
	}

	return bilder
}

// einTischMitEigenenKarten legt beim Wirt eine begonnene Partie an, deren
// Decks aus **eigenen** Karten bestehen — echte Drucke trügen keine Zeile, und
// der Lauf prüfte nichts.
func einTischMitEigenenKarten(t *testing.T, wirt *seite) string {
	t.Helper()

	ctx := riegelkontext(t)
	karte := eineKarteImBestand(t, wirt.anwendung.db)
	ablage := storage.NeueDeckablage(wirt.anwendung.db)

	deck, fehler := ablage.Anlegen(ctx, domain.Deck{Name: "Eigen", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("Deck anlegen: %v", fehler)
	}

	if _, fehler := ablage.PlatzSetzen(ctx, deck.ID, domain.Deckplatz{
		KarteID: karte, Menge: 20, Bereich: domain.BereichHauptdeck,
	}); fehler != nil {
		t.Fatalf("Platz setzen: %v", fehler)
	}

	deckID := strconv.FormatInt(deck.ID, 10)
	flaeche := PartieflaecheUeber(wirt.anwendung.db)

	antwort := flaeche.PartieAusWuenschen(wirt.ctx, "magic", "anna", []Sitzwunsch{
		{Kennung: "anna", Name: "Anna", DeckID: deckID},
		{Kennung: "bert", Name: "Bert", DeckID: deckID},
	})
	if antwort.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
	}

	beginnAbschliessen(t, ueberDieFlaeche(wirt.ctx, flaeche), antwort.Partie)

	// **Am Wirt trägt seine eigene Karte ihre Zeile** — sonst prüfte das Nein
	// am Gast nur eine Vorrichtung ohne Zeilen.
	amWirt := flaeche.Partie(wirt.ctx, antwort.Partie.ID, "anna")
	mitZeile := false

	for _, objekt := range amWirt.Partie.Objekte {
		mitZeile = mitZeile || objekt.KarteID != ""
	}

	if !mitZeile {
		t.Fatal("am Wirt trägt keine Karte eine Zeile")
	}

	return antwort.Partie.ID
}

// bildtexturen zeichnet jede Karte als dasselbe kleine Bild und legt es ins
// Fach — wie das Texturwerk der Anwendung, ohne Stilpaket.
type bildtexturen struct {
	texturprobe
	fach *vorschauweg.Texturfach
	png  []byte
}

func (b *bildtexturen) Textur(
	ctx context.Context, karte *domain.CardRow,
	seite service.Kartenseite, guete service.Texturguete,
) (service.Textur, error) {
	textur, fehler := b.texturprobe.Textur(ctx, karte, seite, guete)
	if fehler != nil {
		return textur, fehler
	}

	textur.Adresse = b.fach.Hinterlege(textur.Schluessel, b.png, 2, 3)
	textur.Breite, textur.Hoehe = 2, 3

	return textur, nil
}

// **Der Gast bekommt das Bild über die Leitung** — gezeichnet vom Wirt mit
// seiner Zeile, abgelegt im eigenen Fach, unter der Bildkennung.
func TestDerGastBekommtDasKartenbildVomWirt(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	bild := einKleinesPNG(t)
	wirt.anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	wirt.anwendung.kartenleser = &texturkartenleser{}
	wirt.anwendung.MitTexturwerk(&bildtexturen{fach: wirt.anwendung.spielbilder, png: bild})
	gast.anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)

	partieID := einTischMitEigenenKarten(t, wirt)
	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	bilder := bilderAmGast(gast)
	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}
	}()

	if antwort := gast.anwendung.PartieBeitreten(gast.ctx, string(wirt.kennung), "", PartiemitbringDTO{}); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	var (
		objekt ObjektDTO
		partie string
	)

	select {
	case bild := <-bilder:
		partie = bild.Partie.ID
		for _, eines := range bild.Partie.Objekte {
			if !eines.Geheim && eines.Bild != "" {
				objekt = eines
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatal("kein Tischbild am Gast")
	}

	if objekt.Bild == "" {
		t.Fatal("der Gast sieht keine Karte mit Bildkennung")
	}

	antwort := gast.anwendung.Fremdkartenbild(gast.ctx, partie, objekt.Kennung, objekt.Bild, "vorne", "tisch")
	if antwort.Fehler != nil || antwort.Textur.Adresse == "" {
		t.Fatalf("Fremdkartenbild: %+v", antwort)
	}

	abgelegt, da := gast.anwendung.spielbilder.Hole(antwort.Textur.Schluessel)
	if !da || !bytes.Equal(abgelegt, bild) {
		t.Errorf("im Fach des Gastes liegt %d Bytes, da=%v", len(abgelegt), da)
	}

	// **Ein zweites Mal aus dem Fach** — ohne den Wirt zu fragen.
	if zweite := gast.anwendung.Fremdkartenbild(gast.ctx, partie, objekt.Kennung, objekt.Bild, "vorne", "tisch"); zweite.Textur.Adresse != antwort.Textur.Adresse {
		t.Errorf("das zweite Mal: %+v", zweite)
	}
}

// einKleinesPNG gibt ein gültiges Bild von 2 × 3 Punkten.
func einKleinesPNG(t *testing.T) []byte {
	t.Helper()

	var puffer bytes.Buffer
	if fehler := png.Encode(&puffer, image.NewNRGBA(image.Rect(0, 0, 2, 3))); fehler != nil {
		t.Fatalf("png: %v", fehler)
	}

	return puffer.Bytes()
}
