package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Karte über ihre Druckkennung** — Werkstück 2 von Phase 3.6
// (#1167).
//
// # Warum es diesen Weg neben `KarteNachName` gibt
//
// `KarteNachName` fragt `/cards/named` und bekommt **irgendeinen** Druck:
// Scryfall wählt selbst aus. Wer in der Trefferliste einen angeklickt hat,
// meint aber genau den — mit seinem Set, seiner Nummer und seinem Bild. Ein
// Verweis auf „Lightning Bolt" wäre kein Verweis, sondern ein Wort.

// druckadresse glättet eine Druckkennung und gibt Speicherschlüssel und
// Adresse dazu.
//
// # Warum an einer Stelle
//
// **Weil zwei Schreibweisen desselben Drucks zwei Einträge im
// Zwischenspeicher ergäben** — und damit zwei Anfragen an eine Gegenstelle,
// die eine Grenze nennt (§16, Regel 4). Die Regel gilt für jeden Weg, der
// `/cards/:id` fragt, und steht deshalb hier.
//
// **Nachgemessen, nicht erinnert**: Scryfall **nimmt** eine UUID mit großen
// Hexadezimalbuchstaben an und antwortet kanonisch klein —
// `/cards/516E137A-D9D9-4BA7-89FB-E680DDB2C2D1` antwortet mit
// `516e137a-d9d9-4ba7-89fb-e680ddb2c2d1`, live erhoben vom Nachlauf über
// Phase 3.4.
//
// **Was hier nicht behauptet wird**: dass Scryfall jede andere Schreibweise
// annimmt. Gemessen ist die Groß-/Kleinschreibung; mehr steht hier nicht.
func (k *Klient) druckadresse(roh string) (kennung, schluessel, adresse string, fehler error) {
	kennung = strings.ToLower(strings.TrimSpace(roh))

	if kennung == "" {
		return "", "", "", fmt.Errorf(
			"%w: ohne Kennung ist kein Druck gemeint", domain.ErrValidation)
	}

	return kennung,
		"cards/id:" + kennung,
		k.grundadresse + "/cards/" + url.PathEscape(kennung),
		nil
}

// KarteNachKennung holt genau einen Druck über `/cards/:id`.
//
// **Derselbe Speicherschlüssel wie beim Preis.** Beide Wege lesen dieselbe
// Antwort — Karte und Preise stehen in demselben JSON. Zwei Schlüssel wären
// zwei Anfragen für eine Auskunft.
//
// **Ohne Netz kommt der letzte gute Stand**, mit seinem Alter (§16, Regel 5);
// ein 404 bleibt ein fachliches Nein und wird nicht durch einen alten Stand
// ersetzt — das entscheidet [Klient.hole] vor dem Speicher.
func (k *Klient) KarteNachKennung(ctx context.Context, roh string) (
	karte *Karte, alter time.Duration, befunde []domain.Befund, fehler error,
) {
	kennung, schluessel, adresse, fehler := k.druckadresse(roh)
	if fehler != nil {
		return nil, 0, nil, fehler
	}

	gelesen, alter, befunde, fehler := k.karte(ctx, schluessel, adresse)
	if fehler != nil {
		return nil, 0, nil, fehler
	}

	// **Die Antwort muss den Druck nennen, nach dem gefragt wurde.**
	//
	// Dieselbe Wache wie bei `PreiseFuerDruck`, und aus einem schwereren
	// Grund: aus dieser Karte werden **Kartenfeldwerte**, und der Verweis
	// daneben behauptet, aus welchem Druck sie stammen. Eine vertauschte
	// Antwort schriebe eine falsche Herkunft — und §11 sagt, eine falsche
	// Nennung ist schlimmer als keine.
	if !strings.EqualFold(gelesen.Kennung, kennung) {
		return nil, 0, befunde, fmt.Errorf(
			"%w: gefragt war der Druck %q, geantwortet hat %q",
			domain.ErrConflict, kennung, gelesen.Kennung)
	}

	return gelesen, alter, befunde, nil
}

// KarteImVorrat gibt einen Druck aus dem Zwischenspeicher — **ohne zu
// fragen** (P10, #2378): eine Handelszeile zeigt, was dieser Rechner schon
// weiß, und das Öffnen eines Tauschs geht nicht ins Netz. Gelesen wird die
// Antwort von `/cards/:id` und, wo es die nicht gibt, was eine Suche über
// diesen Druck gezeigt hat ([Klient.vorratAblegen]).
//
// **Dieselbe Wache wie [Klient.KarteNachKennung]**: eine abgelegte Antwort,
// die einen anderen Druck nennt, ist kein Vorrat für diesen — sonst stünde an
// der Zeile das Bild einer anderen Karte.
func (k *Klient) KarteImVorrat(roh string) (*Karte, bool) {
	kennung, schluessel, _, fehler := k.druckadresse(roh)
	if fehler != nil {
		return nil, false
	}

	for _, unter := range []string{schluessel, vorratsschluessel(kennung)} {
		inhalt, _, gabEs := k.speicher.Vorrat(unter)
		if !gabEs {
			continue
		}

		gelesen, _, lesefehler := LiesKarte(inhalt)
		if lesefehler == nil && strings.EqualFold(gelesen.Kennung, kennung) {
			return gelesen, true
		}
	}

	return nil, false
}

// vorratsschluessel ist der Ort, an dem eine Suche einen Druck ablegt.
//
// **Nicht derselbe wie bei `/cards/:id`**: dort liest auch der Preisweg, und
// eine Suche gäbe ihm sonst ihren Preis von damals, statt dass er fragt.
func vorratsschluessel(kennung string) string {
	return "druck/vorrat:" + kennung
}

// vorratAblegen legt die Drucke einer Suche je Kennung ab — die Kennung ist
// ein Pflichtfeld ([LiesKarte]), jede gelesene Karte hat eine.
//
// **Ein Fehler beim Ablegen macht die Suche nicht kaputt** — er kostet nur,
// dass eine Handelszeile ihren Namen erst nach dem nächsten Abruf kennt.
func (k *Klient) vorratAblegen(rohe []json.RawMessage, karten []*Karte) {
	for stelle, karte := range karten {
		kennung := strings.ToLower(strings.TrimSpace(karte.Kennung))

		_ = k.speicher.Ablegen(vorratsschluessel(kennung), rohe[stelle]) //nolint:errcheck // der Grund steht darüber
	}
}
