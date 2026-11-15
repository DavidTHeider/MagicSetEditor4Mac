// Package scrydex liest Karten bei Scrydex.
//
// **Ein Blatt** (§2, §16): es kennt `internal/domain` und `internal/netz` und
// sonst nichts aus diesem Projekt. Werkstück 5 von Phase 4.8
// (#1321).
//
// # Nur Magic
//
// Scrydex führt **sechs** Spiele; hier zählt eines. Pokémon, Lorcana, Gundam,
// One Piece und Riftbound sind ausdrücklich außerhalb (Vermerk 0056).
//
// # Der Grad von Gewissheit
//
// **Es gibt hier kein Konto** — Vermerk 0056, Entscheidung 4. Geprüft ist
// unsere Seite der Grenze: dass **beide** Kopfzeilen gesetzt werden, die
// Gestalt der Anfrage, die Deutung der Antwort. Ob der Dienst so antwortet,
// wie dieses Paket annimmt, ist **nicht** gemessen.
//
// **Eines ist doch gemessen**, und es widerspricht der Doku: siehe
// [Zugang.Fehlt].
package scrydex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Adresse ist die Grundadresse.
//
// **Nachgelesen am 2026-09-02** unter
// `https://scrydex.com/docs/getting-started/authentication`.
const Adresse = "https://api.scrydex.com"

// Kartenweg ist der Weg zu den Magic-Karten.
//
// **Nachgelesen am 2026-09-02** unter
// `https://scrydex.com/docs/magicthegathering/cards`.
const Kartenweg = "/magicthegathering/v1/cards"

// Seitengroesse ist, wie viele Karten eine Seite trägt.
//
// **Nachgelesen am 2026-09-02**: „The highest allowable value is `100`".
const Seitengroesse = 100

// Abstand ist, wie oft gefragt werden darf.
//
// # Die genannte Grenze und die, die zählt
//
// **Nachgelesen am 2026-09-02**: „There is a rate limit of 100 requests per
// second applied per-second across all API requests." Daneben stehen
// **Kontingente**, die Geld kosten — und die sind die eigentliche Grenze:
// „Each request consumes a specific number of credits."
//
// **Eine Sekunde Abstand, nicht ein Hundertstel.** Die technische Grenze
// erlaubt hundertmal mehr; was sie nicht erlaubt, ist die Rechnung am
// Monatsende. Ein Programm, das im Namen der Nutzenden Geld ausgibt, drosselt
// nicht an der technischen Grenze, sondern an der, die weh tut (§16, Regel 4
// im Sinn und nicht im Buchstaben).
const Abstand = time.Second

// hoechstens ist, wie groß eine Antwort werden darf.
const hoechstens = 8 << 20

// Zugang sind die beiden Angaben, die Scrydex verlangt.
//
// **Zwei und nicht eine** (nachgelesen am 2026-09-02): `X-Api-Key` **und**
// `X-Team-ID`. „Even solo users need a team."
type Zugang struct {
	Schluessel string
	Team       string
}

// Fehlt sagt, welche Angabe fehlt — leer heißt: keine.
//
// # Die Doku sagt etwas anderes als der Dienst
//
// `https://scrydex.com/docs/getting-started/authentication` sagt am
// 2026-09-02: „Requests made without authentication headers will proceed, but
// with heavily reduced rate limits."
//
// **Gemessen am selben Tag** gegen `api.scrydex.com`:
//
//	HTTP/2 401
//	{"error":{"message":"Invalid credentials","code":"INVALID_CREDENTIALS"}}
//
// Es geht **nicht** weiter. Deshalb wird ohne Zugang gar nicht erst gefragt —
// eine Anfrage, deren Antwort feststeht, ist eine zu viel.
//
// **Und die Meldung nennt die fehlende Angabe.** „Zugang fehlt" hilft nicht,
// wenn jemand den Schlüssel längst eingetragen hat und nur die Team-Kennung
// vergaß (§8).
func (z Zugang) Fehlt() string {
	switch {
	case strings.TrimSpace(z.Schluessel) == "" && strings.TrimSpace(z.Team) == "":
		return "der API-Schlüssel und die Team-Kennung"
	case strings.TrimSpace(z.Schluessel) == "":
		return "der API-Schlüssel"
	case strings.TrimSpace(z.Team) == "":
		return "die Team-Kennung"
	}

	return ""
}

// Karte ist eine Karte von Scrydex — **nach** der Zuordnung.
type Karte struct {
	Kennung    string
	Name       string
	Manakosten string
	Typzeile   string
	Seltenheit string

	// Nummer ist die Sammlernummer im Set — sie macht die Zuordnung
	// eindeutig, wo Name und Setcode es nicht tun (A5, #1444).
	Nummer string

	Setcode string
	Setname string

	Regeltext string
	Kuenstler string

	Kraft      string
	Widerstand string
}

// kartenDTO ist die Gestalt, die der Dienst schickt.
//
// **Nachgelesen am 2026-09-02**: `id`, `name`, `mana_cost`, `mana_value`,
// `type`, `types`, `rarity`, `rules`, `artist`, `expansion`, …
//
// **Schlangenschrift, nicht Kamelhöcker.** Scryfall und MTG API schreiben
// `manaCost`; Scrydex schreibt `mana_cost`. Wer das verwechselt, bekommt
// leere Felder ohne Fehlermeldung — das JSON ist gültig, es passt nur nichts.
type kartenDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ManaCost  string    `json:"mana_cost"`
	Type      string    `json:"type"`
	Rarity    string    `json:"rarity"`
	Nummer    string    `json:"number"`
	Rules     regeltext `json:"rules"`
	Artist    string    `json:"artist"`
	Power     string    `json:"power"`
	Toughness string    `json:"toughness"`

	// Expansion trägt Setcode und Setname — **in einem Unterverbund**.
	Expansion struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"expansion"`
}

// antwortDTO ist der Umschlag.
//
// **Nachgelesen am 2026-09-02**: `status`, `data`, `page`, `pageSize`,
// `totalCount`. Der **Fehlerfall** trägt eine andere Gestalt — `error` mit
// `message` und `code` —, und das ist gemessen und nicht gelesen.
type antwortDTO struct {
	Daten      []kartenDTO `json:"data"`
	Gesamtzahl int         `json:"totalCount"`

	Fehler *struct {
		Meldung string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Klient fragt Scrydex.
type Klient struct {
	grundadresse string
	klient       *netz.Klient
	drossel      *netz.Drossel
	zugang       Zugang
}

// NeuerKlient baut einen Klienten.
//
// **Der Zugang kommt herein und wird nicht hier geholt.** Wo er liegt, weiß
// `internal/schluesselbund`; dieses Paket ist ein Blatt und kennt es nicht.
func NeuerKlient(grundadresse string, klient *netz.Klient, zugang Zugang) *Klient {
	return &Klient{
		grundadresse: strings.TrimSuffix(grundadresse, "/"),
		klient:       klient,
		drossel:      netz.NeueDrossel(Abstand, time.Now, time.Sleep),
		zugang:       zugang,
	}
}

// MitDrossel setzt die Drossel — **für Tests**.
func (k *Klient) MitDrossel(drossel *netz.Drossel) *Klient {
	k.drossel = drossel

	return k
}

// Karten sucht Karten nach Namen.
func (k *Klient) Karten(ctx context.Context, name string) ([]Karte, error) {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return nil, fmt.Errorf(
			"%w: für Scrydex fehlt %s — beides gehört in den Schlüsselbund",
			domain.ErrValidation, fehlt)
	}

	// **Die Abfrage ist Lucene-artig** (nachgelesen am 2026-09-02): `q` nimmt
	// `feld:"wert"`. Die Anführungszeichen halten einen Namen mit Leerzeichen
	// zusammen — ohne sie wäre „Black Lotus" zwei Bedingungen.
	adresse := fmt.Sprintf(`%s%s?q=%s&page_size=%d`,
		k.grundadresse, Kartenweg,
		url.QueryEscape(`name:"`+name+`"`), Seitengroesse)

	// **Beide Kopfzeilen gehören zum Kenner** — sie sind die Art, wie sich
	// dieses Programm bei Scrydex vorstellt, und nichts anderes.
	kenner := netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Zusatz: map[string]string{
			"X-Api-Key": k.zugang.Schluessel,
			"X-Team-ID": k.zugang.Team,
		},
	}

	antwort, holfehler := netz.Hole(
		ctx, k.klient, k.drossel, kenner, adresse, hoechstens)
	if holfehler != nil {
		return nil, fmt.Errorf("Scrydex fragen: %w", holfehler)
	}

	return lies(antwort)
}

// LiesKarten deutet einen Rumpf samt Statuscode — **öffentlich für den
// Vertragstest**.
//
// §16 verlangt einen Test gegen eingecheckte Beispiele; er braucht einen Weg
// hinein, der nicht durch das Netz führt.
func LiesKarten(rumpf []byte, status int) ([]Karte, error) {
	return lies(netz.Antwort{Status: status, Rumpf: rumpf})
}

// lies deutet die Antwort.
func lies(antwort netz.Antwort) ([]Karte, error) {
	var gelesen antwortDTO

	if err := json.Unmarshal(antwort.Rumpf, &gelesen); err != nil {
		// **Hier bleibt das zweite `%w`.** `json.Unmarshal` gibt zurück, was
		// ein eigenes `UnmarshalJSON` gemeldet hat — und `regeltext.go` sagt
		// dort, **welches Feld** welchen Unsinn trug. Das ist die Auskunft,
		// nicht die Ursache dahinter.
		return nil, fmt.Errorf(
			"%w: die Antwort von Scrydex ist kein gültiges JSON: %w",
			domain.ErrValidation, err)
	}

	// **Der Fehler im Rumpf steht vor dem Statuscode.** Er sagt mehr: „Invalid
	// credentials" ist eine Auskunft, „401" ist eine Zahl.
	if gelesen.Fehler != nil {
		return nil, fmt.Errorf("%w: Scrydex meldet: %s",
			domain.ErrValidation, gelesen.Fehler.Meldung)
	}

	if antwort.Status != netz.StatusGut {
		return nil, fmt.Errorf("%w: Scrydex antwortet mit %d",
			domain.ErrValidation, antwort.Status)
	}

	karten := make([]Karte, 0, len(gelesen.Daten))

	for _, dto := range gelesen.Daten {
		if err := pflichtfelder(dto); err != nil {
			return nil, err
		}

		karten = append(karten, alsKarte(dto))
	}

	return karten, nil
}

// pflichtfelder prüft, was dasein muss.
//
// **`id` und `name`**: ohne die Kennung lässt sich die Karte nicht
// wiederfinden, ohne den Namen nicht anzeigen.
//
// **`expansion.code` ist keines.** Scrydex führt sechs Spiele und mehrere
// Sprachen; ob jede Antwort eine Ausgabe trägt, ist hier **nicht** zu
// messen — es gibt kein Konto. Ein Pflichtfeld, das auf einer Annahme
// beruht, wiese irgendwann eine gültige Antwort ab.
//
// **Die Reihenfolge ist festgeschrieben**, wie in den Schwesterpaketen.
func pflichtfelder(dto kartenDTO) error {
	for _, fall := range []struct{ feld, wert string }{
		{"id", dto.ID},
		{"name", dto.Name},
	} {
		if strings.TrimSpace(fall.wert) == "" {
			return fmt.Errorf(
				"%w: die Karte von Scrydex hat kein Feld %q — das ist ein Pflichtfeld",
				domain.ErrValidation, fall.feld)
		}
	}

	return nil
}

// alsKarte ist die Zuordnung — **an einer Stelle** (§16).
func alsKarte(dto kartenDTO) Karte {
	return Karte{
		Kennung:    dto.ID,
		Name:       dto.Name,
		Manakosten: dto.ManaCost,
		Typzeile:   dto.Type,
		Seltenheit: dto.Rarity,
		Nummer:     dto.Nummer,
		Setcode:    dto.Expansion.Code,
		Setname:    dto.Expansion.Name,
		Regeltext:  string(dto.Rules),
		Kuenstler:  dto.Artist,
		Kraft:      dto.Power,
		Widerstand: dto.Toughness,
	}
}
