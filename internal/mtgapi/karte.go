// Package mtgapi liest Karten von magicthegathering.io.
//
// **Ein Blatt** (§2, §16): es kennt `internal/domain` und `internal/netz` und
// sonst nichts aus diesem Projekt. Werkstück 3 von Phase 4.8
// (#1321),
// entschieden in [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md).
//
// # Warum kein SDK
//
// `github.com/MagicTheGathering/mtg-sdk-go` ist MIT und täte, was es soll. Es
// kommt trotzdem nicht ins `go.mod` — die drei Gründe stehen in 0056, und der
// schwerste ist: es bringt seinen **eigenen** Transport mit, und damit lägen
// Drossel, Frist und der Umgang mit dem 403 in einer fremden Bibliothek.
//
// # Drei Schichten, und die dritte ist die, die man weglässt
//
// Transport (`internal/netz`), fremdes DTO (`kartenDTO`), Zuordnung
// (`alsKarte`). Wer den fremden Verbund direkt weiterreichte, hätte die
// Gestalt eines fremden Anbieters in seiner Fachlogik.
package mtgapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Karte ist eine Karte von magicthegathering.io — **nach** der Zuordnung.
//
// **Deutsche Feldnamen, englische Werte.** Der Typ ist ein eigener Begriff
// dieses Projekts ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md));
// was in ihm steht, kommt aus einer fremden Datei und bleibt, wie es dort
// steht.
type Karte struct {
	// Kennung ist die `id` des Dienstes.
	//
	// **Nachgelesen am 2026-09-02**: „made up by doing an SHA1 hash of
	// setCode + cardName + cardImageName". Sie ist damit **nicht** die
	// Scryfall-Kennung und auch keine Multiverse-ID.
	Kennung string

	Name       string
	Manakosten string

	// Typzeile ist die ganze Zeile — `type`, nicht `types`.
	Typzeile string

	Seltenheit string
	Setcode    string
	Setname    string
	Nummer     string

	Regeltext  string
	Flavortext string
	Kuenstler  string

	// Kraft und Widerstand sind **Zeichenketten**, weil `*` eine gültige
	// Angabe ist. Eine Zahl daraus zu machen hieße, „*/*" zu verlieren.
	Kraft      string
	Widerstand string

	Bildadresse string

	// Legalitaeten sind Format → Zustand, beides **kleingeschrieben**.
	//
	// **Warum umgeschrieben wird**: MTG API sagt „Legal", Scryfall „legal".
	// Zwei Schreibweisen für dieselbe Aussage wären zwei Wahrheiten, und die
	// Formatprüfung müsste beide kennen.
	Legalitaeten map[string]string
}

// antwortDTO ist die Gestalt, die der Dienst schickt.
type antwortDTO struct {
	Karten []kartenDTO `json:"cards"`
}

// kartenDTO ist eine Karte mit **seinen** Feldnamen.
//
// **Nachgelesen am 2026-09-02** unter `https://docs.magicthegathering.io/`.
// Was hier fehlt, fehlt mit Absicht: `foreignNames`, `rulings`, `variations`,
// `printings` und die Vanguard-Felder beantworten keine Frage, die dieses
// Programm heute stellt — und ein Feld, das niemand liest, ist eine Zusage,
// die niemand hält.
type kartenDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ManaCost string `json:"manaCost"`
	Type     string `json:"type"`
	Rarity   string `json:"rarity"`

	Set     string `json:"set"`
	SetName string `json:"setName"`
	Number  string `json:"number"`

	Text   string `json:"text"`
	Flavor string `json:"flavor"`
	Artist string `json:"artist"`

	Power     string `json:"power"`
	Toughness string `json:"toughness"`

	ImageURL string `json:"imageUrl"`

	Legalities []legalitaetDTO `json:"legalities"`
}

// legalitaetDTO ist ein Format samt seinem Zustand.
type legalitaetDTO struct {
	Format   string `json:"format"`
	Legality string `json:"legality"`
}

// zustaende sind die Werte, die `legality` haben darf.
//
// **Eine Aufzählung und kein Kleinschreiben.** Wer bloß kleinschriebe, nähme
// jeden Wert an — auch einen, den der Anbieter neu erfunden hat. Dann stünde
// er in einer Karte, und die Formatprüfung deutete ihn als „unbekannt heißt
// unzulässig" oder als „unbekannt heißt egal"; welches von beidem, entschiede
// der Zufall (§16: Drift wird erkannt, nicht erlitten).
var zustaende = map[string]string{
	"Legal":      "legal",
	"Not Legal":  "not_legal",
	"Restricted": "restricted",
	"Banned":     "banned",
}

// LiesKarten liest die Antwort auf `/cards`.
//
// **Pflichtfelder streng, Zusatzfelder tolerant** (§16). Fehlt eines der
// drei, die eine Karte ausmachen, ist es ein Fehler; ein neues Feld beim
// Anbieter stört nicht.
func LiesKarten(roh []byte) ([]Karte, []domain.Befund, error) {
	var antwort antwortDTO

	if err := json.Unmarshal(roh, &antwort); err != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, err,
			"die Antwort von magicthegathering.io ist kein gültiges JSON")
	}

	var (
		karten  = make([]Karte, 0, len(antwort.Karten))
		befunde []domain.Befund
	)

	for _, dto := range antwort.Karten {
		eine, gefunden, err := alsKarte(dto)
		if err != nil {
			return nil, nil, err
		}

		karten = append(karten, eine)
		befunde = append(befunde, gefunden...)
	}

	return karten, befunde, nil
}

// alsKarte ist die Zuordnung — **an einer Stelle** (§16).
func alsKarte(dto kartenDTO) (Karte, []domain.Befund, error) {
	if err := pflichtfelder(dto); err != nil {
		return Karte{}, nil, err
	}

	legal, befunde := alsLegalitaeten(dto)

	return Karte{
		Kennung:      dto.ID,
		Name:         dto.Name,
		Manakosten:   dto.ManaCost,
		Typzeile:     dto.Type,
		Seltenheit:   dto.Rarity,
		Setcode:      dto.Set,
		Setname:      dto.SetName,
		Nummer:       dto.Number,
		Regeltext:    dto.Text,
		Flavortext:   dto.Flavor,
		Kuenstler:    dto.Artist,
		Kraft:        dto.Power,
		Widerstand:   dto.Toughness,
		Bildadresse:  sichereBildadresse(dto.ImageURL),
		Legalitaeten: legal,
	}, befunde, nil
}

// gathererKlartext ist der Anfang, mit dem magicthegathering.io jedes Bild
// nennt.
const gathererKlartext = "http://gatherer.wizards.com/"

// sichereBildadresse nennt ein Gatherer-Bild über `https`.
//
// **Weil die WebView Klartext nicht holt** — ihre Schutzrichtlinie kennt nur
// `https`-Ursprünge (`internal/fenster/schutznaht.go`). Gemessen am
// 2026-10-07: dieselbe Adresse über `http` antwortet mit 301 auf ihr
// `https`-Gegenstück. Ein anderer Ursprung bleibt, wie er ist: ob er `https`
// spricht, ist nicht gemessen.
func sichereBildadresse(adresse string) string {
	if rest, ok := strings.CutPrefix(adresse, gathererKlartext); ok {
		return "https://gatherer.wizards.com/" + rest
	}

	return adresse
}

// pflichtfelder prüft, was dasein muss.
//
// # Die Reihenfolge ist festgeschrieben, und das ist der Punkt
//
// Dieselbe Überlegung wie in `internal/scryfall`: ein `map`-Literal würfelte
// die Reihenfolge, und dieselbe Antwort nennte zweimal ein anderes fehlendes
// Feld. Gefunden hat das ein Fuzzlauf in Phase 4.2 (#1289).
func pflichtfelder(dto kartenDTO) error {
	for _, fall := range []struct {
		feld string
		wert string
	}{
		{"id", dto.ID},
		{"name", dto.Name},
		{"set", dto.Set},
	} {
		if strings.TrimSpace(fall.wert) == "" {
			return fmt.Errorf(
				"%w: die Karte von magicthegathering.io hat kein Feld %q — das ist ein Pflichtfeld",
				domain.ErrValidation, fall.feld)
		}
	}

	return nil
}

// alsLegalitaeten macht aus der Liste eine Abbildung.
//
// **Ein unbekannter Zustand wird ein Befund und kommt nicht in die
// Abbildung.** Ihn aufzunehmen hieße, eine Auskunft weiterzugeben, die
// niemand deuten kann; ihn stillschweigend fallen zu lassen hieße, dass
// niemand von der Änderung erfährt (§8, §16).
func alsLegalitaeten(dto kartenDTO) (map[string]string, []domain.Befund) {
	if len(dto.Legalities) == 0 {
		return nil, nil
	}

	var (
		legal   = make(map[string]string, len(dto.Legalities))
		befunde []domain.Befund
	)

	for _, eine := range dto.Legalities {
		zustand, kennt := zustaende[eine.Legality]
		if !kennt {
			befunde = append(befunde, domain.Befund{
				Datei: dto.ID,
				Ort:   "legalities." + eine.Format,
				Art:   domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"magicthegathering.io nennt für %q den Zustand %q — den kennt "+
						"dieses Programm nicht", eine.Format, eine.Legality),
			})

			continue
		}

		legal[strings.ToLower(eine.Format)] = zustand
	}

	return legal, befunde
}
