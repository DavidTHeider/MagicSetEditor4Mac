package app

// **Was aus einem angenommenen Umschlag wird** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// # Drei Wege, und sie sind ausdrücklich nicht einer
//
//  1. **Ein Set** wird als neues Set angelegt — der Weg aus `N4`.
//  2. **Ein Deck** wird als neues Deck angelegt, samt seinen Plätzen.
//  3. **Eine Sammlungssicht** wird **nicht** angenommen. Sie wird angesehen.
//
// # Warum die dritte keine Annahme ist
//
// **Weil sie kein Besitz ist.** Sie sagt, was ein Gegenüber **hat**; sie in
// die eigene Sammlung zu schreiben hieße, fremde Karten als eigene zu führen.
// Das Issue sagt es in einem Satz: „Empfang/Import ist keine physische
// Abbuchung und garantiert keine exklusive Eigentumsübertragung." Wohin sie
// gehört, ist der **Abgleich** (`T4`) — und dorthin kommt sie als Stand und
// nicht als Bestand.
//
// # Und Karte und Vorlage stehen nebenan
//
// **Weil beide ein Ziel brauchen, das der Umschlag nicht nennt** — eine Karte
// gehört in ein **Set**, eine Vorlage in ein **Paket**. Seit dem 2026-09-19
// steht die Wahl **im Ruf**, und die zwei Wege stehen in `umschlagkarte.go`
// (#1639).

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UmschlagSicht liest eine empfangene Sammlungssicht — **ohne zu schreiben**.
//
// **Sie ist der Stand eines Gegenübers**, und was daraus wird, entscheidet
// der Abgleich (`T4`). In den eigenen Bestand geht sie nie.
func (a *App) UmschlagSicht(
	ctx context.Context, pfad string,
) FreigabesichtAntwort {
	ctx = a.imProfil(ctx)

	leer := FreigabesichtAntwort{Zeilen: []SichtkarteDTO{}}

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	roh, fehler := lesenAusSendung(sendung, austausch.ArtSammlung, Sichtdatei)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	zeilen := []SichtkarteDTO{}
	if fehler := json.Unmarshal(roh, &zeilen); fehler != nil {
		leer.Fehler = Uebersetze(ctx, domain.MitGrund(domain.ErrValidation,
			fehler, "die Sammlungssicht lesen"))

		return leer
	}

	return FreigabesichtAntwort{Zeilen: zeilen}
}

// UmschlagDeck legt ein empfangenes Deck als neues Deck an.
//
// **Als neues und nicht über ein vorhandenes.** Ein Umschlag von außen darf
// nichts überschreiben; wer zweimal annimmt, hat zwei Decks — und sieht es.
func (a *App) UmschlagDeck(ctx context.Context, pfad string) DeckAntwort {
	ctx = a.imProfil(ctx)

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	roh, fehler := lesenAusSendung(sendung, austausch.ArtDeck, Deckdatei)
	if fehler != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	var gelesen DeckDTO

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, fehler, "das Deck lesen"))}
	}

	return a.deckAnnehmen(ctx, gelesen)
}

// deckAnnehmen legt das gelesene Deck an und füllt es.
//
// # Warum ein gleichnamiges Deck einen Zusatz bekommt
//
// **Weil ein Name in diesem Bestand eindeutig ist und im fremden nicht.** Wer
// sein eigenes Deck zurückbekommt oder von zwei Leuten ein „Standard" erhält,
// liefe sonst gegen einen Konflikt, den er nicht verursacht hat. Der Zusatz
// ist **sichtbar**: stillschweigend umzubenennen wäre schlimmer als
// abzulehnen, und abzulehnen wäre unbrauchbar.
//
// **Einmal und nicht in einer Schleife.** Wer auch den Zusatz schon vergeben
// hat, bekommt den Konflikt — mit einem Satz, der sagt, was los ist.
func (a *App) deckAnnehmen(ctx context.Context, gelesen DeckDTO) DeckAntwort {
	angelegt := a.DeckAnlegen(ctx, gelesen.Name, gelesen.Spiel, gelesen.Notiz)

	if angelegt.Fehler != nil && angelegt.Fehler.Code == CodeKonflikt {
		angelegt = a.DeckAnlegen(ctx,
			gelesen.Name+" ("+DialogtexteFuer(a.sprache).Uebernommen+")",
			gelesen.Spiel, gelesen.Notiz)
	}

	if angelegt.Fehler != nil || angelegt.Deck == nil {
		return angelegt
	}

	// **Die Plätze gehen einzeln hinein.** Ein Platz, den dieser Rechner
	// nicht legen kann — weil die Karte fehlt —, ist ein Befund und kein
	// Grund, das Deck zu verwerfen (§8). **Und kein Grund, die übrigen
	// auszulassen**: bis #1995 gab der erste gescheiterte Platz zurück, und
	// alle danach fehlten still.
	//
	// **Eine eigene Karte des Absenders kommt nicht mit** (FMT-04, #1693):
	// ihre Kennung gilt in **seinem** Bestand. Hier zeigte dieselbe Zahl auf
	// eine andere Karte oder auf keine — bis FMT-04 wurde sie trotzdem
	// gelegt, und das Deck trug still eine fremde Karte.
	var fremde, nichtGelegt []string

	for _, platz := range gelesen.Plaetze {
		if platz.KarteID != "" {
			name := platz.Bezeichnung
			if name == "" {
				name = "#" + platz.KarteID
			}

			fremde = append(fremde, name)

			continue
		}

		gelegt := a.DeckplatzLegen(ctx, angelegt.Deck.ID, "",
			platz.Printing, platz.Menge, platz.Bereich)
		if gelegt.Fehler != nil {
			nichtGelegt = append(nichtGelegt, fmt.Sprintf("%s ×%d", platz.Printing, platz.Menge))
		}
	}

	gelesenesDeck := a.Deck(ctx, angelegt.Deck.ID)

	// **Angenommen ist angenommen** (§7, Regel 3; #1970): scheitert das
	// Nachlesen, ist das Deck trotzdem da. Die Antwort nennt es, wie es
	// angelegt wurde — ohne Deck wüsste niemand, dass es eines gibt, und wer
	// es noch einmal annähme, legte ein zweites an.
	if gelesenesDeck.Fehler != nil {
		gelesenesDeck = DeckAntwort{Deck: angelegt.Deck, Nachladefehler: gelesenesDeck.Fehler}
	}

	if befund := annahmebefund(fremde, nichtGelegt); befund != nil {
		gelesenesDeck.Fehler = Uebersetze(ctx, befund)
	}

	return gelesenesDeck
}

// annahmebefund nennt, was vom Deck nicht ankam — fremde eigene Karten,
// nicht legbare Plätze oder beides. **Beides in einem Befund**: die Antwort
// trägt einen, und der eine darf den anderen nicht verdecken (#1995).
func annahmebefund(fremde, nichtGelegt []string) *domain.FieldError {
	karten := strings.Join(fremde, ", ")
	plaetze := strings.Join(nichtGelegt, ", ")

	switch {
	case len(fremde) > 0 && len(nichtGelegt) > 0:
		return domain.NewFieldError("plaetze", domain.SchluesselFremdeKartenUndNichtGelegte,
			"Eigene Karten des Absenders kamen nicht mit — ihre Kennung gilt nur in seinem Bestand: %s. Und diese Plätze ließen sich hier nicht legen: %s. Prüfe beides und lege es von Hand nach.",
			karten, plaetze).MitWerten(map[string]string{"karten": karten, "plaetze": plaetze})
	case len(fremde) > 0:
		return domain.NewFieldError("plaetze", domain.SchluesselFremdeEigeneKarten,
			"Eigene Karten des Absenders kamen nicht mit — ihre Kennung gilt nur in seinem Bestand: %s.",
			karten).MitWerten(map[string]string{"karten": karten})
	case len(nichtGelegt) > 0:
		return domain.NewFieldError("plaetze", domain.SchluesselNichtGelegtePlaetze,
			"Diese Plätze ließen sich hier nicht legen: %s. Das übrige Deck ist angenommen; prüfe die Plätze und lege sie von Hand nach.",
			plaetze).MitWerten(map[string]string{"plaetze": plaetze})
	default:
		return nil
	}
}

// lesenAusSendung holt die eine Datei einer Art aus einer Sendung.
func lesenAusSendung(
	sendung geleseneSendung, art austausch.Objektart, name string,
) ([]byte, error) {
	for _, objekt := range sendung.Manifest.Objekte {
		if objekt.Art != art {
			continue
		}

		// #nosec G304 -- der Ort ist das Wegwerfverzeichnis dieser Sendung
		roh, fehler := os.ReadFile(filepath.Join(sendung.Verzeichnis, name))
		if fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"%s aus dem Umschlag lesen", name)
		}

		return roh, nil
	}

	return nil, fmt.Errorf("%w: dieser Umschlag enthält kein %s",
		domain.ErrValidation, art)
}
