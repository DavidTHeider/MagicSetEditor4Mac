package app

// **Ein Gast setzt sich mit seinem eigenen Deck** — P9b
// (#2377):
// zwei Anwendungen, je mit eigenem Bestand. Der Wirt lädt Bert an einen Sitz
// ohne Deck ein; Bert bringt eines aus **seinem** Bestand mit. Danach zeigen
// Berts Karten an seinem Fenster auf **seine** Zeilen — und beim Wirt auf
// keine, denn dort fänden sie eine andere Karte.

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// einEigenesDeck legt im Bestand dieser Seite ein Deck aus einer eigenen
// Karte an und gibt Deck und Zeile.
func einEigenesDeck(t *testing.T, wer *seite, menge int) (deckID string, zeile int64) {
	t.Helper()

	ctx := riegelkontext(t)
	zeile = eineKarteImBestand(t, wer.anwendung.db)
	ablage := storage.NeueDeckablage(wer.anwendung.db)

	deck, fehler := ablage.Anlegen(ctx, domain.Deck{Name: "Eigen", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("Deck anlegen: %v", fehler)
	}

	if _, fehler := ablage.PlatzSetzen(ctx, deck.ID, domain.Deckplatz{
		KarteID: zeile, Menge: menge, Bereich: domain.BereichHauptdeck,
	}); fehler != nil {
		t.Fatalf("Platz setzen: %v", fehler)
	}

	return strconv.FormatInt(deck.ID, 10), zeile
}

// einTischMitEinladung legt beim Wirt eine Partie an, deren zweiter Sitz
// auf diesen Gast wartet.
func einTischMitEinladung(t *testing.T, wirt, gast *seite) string {
	t.Helper()

	deckID, _ := einEigenesDeck(t, wirt, 20)

	antwort := PartieflaecheUeber(wirt.anwendung.db).PartieAusWuenschen(wirt.ctx, "magic", "anna", []Sitzwunsch{
		{Kennung: "anna", Name: "Anna", DeckID: deckID},
		{Kennung: "bert", Name: "Bert", Eingeladen: string(gast.kennung)},
	})
	if antwort.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
	}

	if antwort := wirt.anwendung.PartieFreigeben(wirt.ctx, antwort.Partie.ID, "anna", []string{"bert"}); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	return antwort.Partie.ID
}

// amTischAnnehmen nimmt beim Wirt die nächste Sitzung an.
func amTischAnnehmen(wirt *seite) chan error {
	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	return angenommen
}

// bisBertSitzt wartet auf das erste Tischbild, in dem Berts Sitz nicht mehr
// wartet.
func bisBertSitzt(t *testing.T, bilder chan PartieAntwort) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		select {
		case bild := <-bilder:
			for _, sitz := range bild.Partie.Sitze {
				if sitz.Kennung == "bert" && !sitz.Wartet && sitz.Karten > 0 {
					return
				}
			}
		case <-frist:
			t.Fatal("Bert sitzt nicht")

			return
		}
	}
}

func TestEinGastSetztSichMitSeinemEigenenDeck(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)
	MitPartienleitung(gast.anwendung, gast.anwendung.db)

	partieID := einTischMitEinladung(t, wirt, gast)
	deckID, zeile := einEigenesDeck(t, gast, 9)

	bilder := bilderAmGast(gast)
	angenommen := amTischAnnehmen(wirt)

	if antwort := gast.anwendung.PartieBeitreten(gast.ctx, string(wirt.kennung), "",
		PartiemitbringDTO{DeckID: deckID}); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	bisBertSitzt(t, bilder)

	// **Berts Hand zeigt auf seine Zeilen** — gezeichnet wird bei ihm.
	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "starthand/bert", Art: "spieler.starthand",
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	hand := bisZurHand(t, bilder)
	for _, objekt := range hand {
		if objekt.KarteID != strconv.FormatInt(zeile, 10) || objekt.Bild == "" {
			t.Errorf("Berts Karte %s: Zeile %q, Bild %q", objekt.Kennung, objekt.KarteID, objekt.Bild)
		}
	}

	// **Beim Wirt zeigt sie auf keine** — sein Bestand hat Berts Karten nicht.
	beimWirt := PartieflaecheUeber(wirt.anwendung.db).Partie(wirt.ctx, partieID, "bert")
	for _, objekt := range beimWirt.Partie.Objekte {
		if objekt.Eigner == "bert" && objekt.KarteID != "" {
			t.Errorf("beim Wirt trägt %s die Zeile %s", objekt.Kennung, objekt.KarteID)
		}
	}
}

// bisZurHand wartet auf ein Tischbild, in dem Berts Hand offen liegt.
func bisZurHand(t *testing.T, bilder chan PartieAntwort) []ObjektDTO {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		select {
		case bild := <-bilder:
			var hand []ObjektDTO

			for _, objekt := range bild.Partie.Objekte {
				if objekt.Eigner == "bert" && objekt.Zone == string(domain.ZoneHand) && !objekt.Geheim {
					hand = append(hand, objekt)
				}
			}

			if len(hand) > 0 {
				return hand
			}
		case <-frist:
			t.Fatal("Berts Hand kommt nicht")

			return nil
		}
	}
}

// **Der Wirt sieht Berts Karte mit Berts Bild** — gezeichnet bei Bert, denn
// nur sein Bestand kennt die Zeile.
func TestDerWirtSiehtDieKarteDesGastesMitDessenBild(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)
	MitPartienleitung(gast.anwendung, gast.anwendung.db)

	bild := einKleinesPNG(t)
	wirt.anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	gast.anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	gast.anwendung.kartenleser = &texturkartenleser{}
	gast.anwendung.MitTexturwerk(&bildtexturen{fach: gast.anwendung.spielbilder, png: bild})

	partieID := einTischMitEinladung(t, wirt, gast)
	deckID, _ := einEigenesDeck(t, gast, 9)

	bilder := bilderAmGast(gast)
	angenommen := amTischAnnehmen(wirt)

	if antwort := gast.anwendung.PartieBeitreten(gast.ctx, string(wirt.kennung), "",
		PartiemitbringDTO{DeckID: deckID}); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	bisBertSitzt(t, bilder)

	objekt := bertsKarteAufDemTisch(t, wirt, gast, partieID, bilder)

	antwort := wirt.anwendung.Fremdkartenbild(wirt.ctx, partieID, objekt.Kennung, objekt.Bild, "vorne", "tisch")
	if antwort.Fehler != nil || antwort.Textur.Adresse == "" {
		t.Fatalf("Fremdkartenbild: %+v", antwort)
	}

	// **Und für einen anderen Tisch nicht** (Zweitblick über Phase 7,
	// Befund 4) — auch nicht aus dem Fach.
	if fremd := wirt.anwendung.Fremdkartenbild(wirt.ctx, "eine-andere", objekt.Kennung, objekt.Bild, "vorne", "tisch"); fremd.Fehler == nil || fremd.Fehler.Code != CodeKonflikt {
		t.Errorf("für einen anderen Tisch: %+v", fremd)
	}

	if abgelegt, da := wirt.anwendung.spielbilder.Hole(antwort.Textur.Schluessel); !da || !bytes.Equal(abgelegt, bild) {
		t.Errorf("im Fach des Wirts liegen %d Bytes, da=%v", len(abgelegt), da)
	}
}

// bertsKarteAufDemTisch bringt den Beginn zu Ende, legt eine Karte aus
// Berts Hand auf den Tisch — **über Berts eigene Wünsche**, wie im Betrieb:
// nur so kennt seine letzte Sicht die Karte — und gibt sie, wie der Wirt sie
// sieht.
func bertsKarteAufDemTisch(
	t *testing.T, wirt, gast *seite, partieID string, bilder chan PartieAntwort,
) ObjektDTO {
	t.Helper()

	for _, art := range []string{"spieler.starthand", "spieler.behalten"} {
		if antwort := wirt.anwendung.PartieBefehl(wirt.ctx, domain.Spielbefehl{
			PartieID: partieID, Kennung: art + "/anna", Absender: "anna", Art: art,
		}); antwort.Fehler != nil {
			t.Fatalf("%s für Anna: %+v", art, antwort.Fehler)
		}

		if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
			Kennung: art + "/bert", Art: art,
		}); antwort.Fehler != nil {
			t.Fatalf("%s für Bert: %+v", art, antwort.Fehler)
		}
	}

	gelegt := bisZurHand(t, bilder)[0]
	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "schieben/" + gelegt.Kennung, Art: "objekt.schieben",
		Nutzlast: map[string]any{"objekt": gelegt.Kennung, "zone": string(domain.ZoneTisch)},
	}); antwort.Fehler != nil {
		t.Fatalf("schieben: %+v", antwort.Fehler)
	}

	bisAufDemTisch(t, bilder, gelegt.Kennung)

	for _, objekt := range PartieflaecheUeber(wirt.anwendung.db).Partie(wirt.ctx, partieID, "anna").Partie.Objekte {
		if objekt.Kennung == gelegt.Kennung {
			if objekt.KarteID != "" || objekt.Bild == "" {
				t.Fatalf("beim Wirt: Zeile %q, Bild %q", objekt.KarteID, objekt.Bild)
			}

			return objekt
		}
	}

	t.Fatal("beim Wirt liegt Berts Karte nicht")

	return ObjektDTO{}
}

// bisAufDemTisch wartet, bis Berts Fenster die Karte auf dem Tisch sieht.
func bisAufDemTisch(t *testing.T, bilder chan PartieAntwort, kennung string) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		select {
		case bild := <-bilder:
			for _, objekt := range bild.Partie.Objekte {
				if objekt.Kennung == kennung && objekt.Zone == string(domain.ZoneTisch) {
					return
				}
			}
		case <-frist:
			t.Fatal("Berts Karte kommt nicht auf den Tisch")

			return
		}
	}
}

// **Ein Gast ohne Deck liest die Absage des Wirts** — P9c, gefunden am
// nativen Fenster (`netzpartie09`): der Grund kam an, aber das Fenster sprach
// nur von einer abgerissenen Leitung — und erst nach „Nachsehen“.
func TestEinGastOhneDeckLiestDieAbsage(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)
	MitPartienleitung(gast.anwendung, gast.anwendung.db)
	einTischMitEinladung(t, wirt, gast)

	gemeldet := make(chan struct{}, 4)
	vorher := gast.anwendung.melden
	gast.anwendung.melden = func(ctx context.Context, ereignis string, daten ...any) {
		vorher(ctx, ereignis, daten...)

		if ereignis == EreignisPartiebeitritt {
			gemeldet <- struct{}{}
		}
	}

	angenommen := amTischAnnehmen(wirt)
	if antwort := gast.anwendung.PartieBeitreten(gast.ctx, string(wirt.kennung), "", PartiemitbringDTO{}); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	// **Gewartet wird auf die Wirkung, nicht auf das erste Ereignis** (§3):
	// unter Last kam zuerst „pausiert" an und die Absage mit dem nächsten —
	// rot in der Vorprüfung vom 2026-10-19, grün allein und unter `-race`.
	frist := time.After(10 * time.Second)

	for {
		select {
		case <-gemeldet:
		case <-frist:
			t.Fatalf("das Fenster erfuhr die Absage nicht: %+v",
				gast.anwendung.PartieBeitrittStand(gast.ctx))
		}

		stand := gast.anwendung.PartieBeitrittStand(gast.ctx)
		if !stand.Laeuft && strings.HasPrefix(stand.Absage, "dieser Tisch wartet auf dein Deck") {
			return
		}
	}
}
