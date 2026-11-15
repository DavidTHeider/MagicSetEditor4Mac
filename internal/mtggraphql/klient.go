// Package mtggraphql liest Karten über die GraphQL-Schnittstelle von MTGJSON.
//
// **Ein Blatt** (§2, §16): es kennt `internal/domain` und `internal/netz` und
// sonst nichts aus diesem Projekt. Werkstück 4 von Phase 4.8
// (#1321).
//
// # Der Grad von Gewissheit ist hier ein anderer
//
// **Nachgemessen am 2026-09-02** gegen `https://graphql.mtgjson.com/`, ohne
// Zugangsmarke:
//
//   - **Das Schema ist ohne Marke lesbar.** Introspektion antwortet; die
//     Feldnamen in [kartenDTO] sind daran geprüft und nicht geraten.
//   - **Daten sind es nicht.** Eine echte Abfrage gibt **HTTP 200** mit
//     `data: null` und `errors[0].message` „Access denied! You need to be
//     authorized to perform this action!".
//   - **Und der Code dazu lautet `INTERNAL_SERVER_ERROR`.** Das ist
//     irreführend: es ist kein Serverfehler, sondern eine fehlende Marke.
//
// **Das ist mehr, als Vermerk 0056 angenommen hat**, und weniger als bei
// magicthegathering.io: die **Gestalt** der Anfrage lässt sich gegen das
// echte Schema prüfen, die **Antwort** nicht. Was daraus folgt, sagt die
// Oberfläche: „nie gegen den echten Dienst gelaufen".
package mtggraphql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Adresse ist der Endpunkt.
//
// **Nachgelesen am 2026-09-02** unter `https://mtgjson.com/mtggraphql/`.
const Adresse = "https://graphql.mtgjson.com/"

// Abstand ist, wie oft gefragt werden darf.
//
// # Zwei Grenzen, und die kleinere gilt
//
// **Nachgelesen am 2026-09-02**: „The current rate limits are capped at
// `1,000` requests per IP Address per hour and `500` requests per access
// token per hour."
//
// **Die Marke ist die engere**: 500 je Stunde ist eine je 7,2 Sekunden.
// Gedrosselt wird **unter** der genannten Zahl (§16, Regel 4); zehn Sekunden
// lassen 360 zu, mit 28 % Abstand — dasselbe Verhältnis wie bei den anderen
// Diensten.
//
// **Eine Drossel und nicht zwei.** Zwei Grenzen mit demselben Zähler zu
// bedienen ginge nur, wenn dieses Programm die IP mit anderen teilte; es
// fragt aus einem Prozess mit einer Marke, und die engere deckt beide.
const Abstand = 10 * time.Second

// hoechstens ist, wie groß eine Antwort werden darf.
const hoechstens = 8 << 20

// Karte ist eine Karte von MTGJSON — **nach** der Zuordnung.
type Karte struct {
	// Kennung ist die `uuid` von MTGJSON.
	//
	// **Nicht die Scryfall-Kennung.** MTGJSON führt eigene UUIDs; wer sie
	// verwechselte, suchte bei Scryfall nach etwas, das es dort nicht gibt.
	Kennung string

	Name       string
	Manakosten string
	Typzeile   string
	Seltenheit string
	Setcode    string
	Nummer     string

	Regeltext  string
	Flavortext string
	Kuenstler  string

	Kraft      string
	Widerstand string
}

// kartenDTO ist die Gestalt, die der Dienst schickt.
//
// **Die Feldnamen sind am echten Schema geprüft** (2026-09-02, über
// Introspektion): `Card` hat 67 Felder, und diese hier heißen so.
type kartenDTO struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	ManaCost   string `json:"manaCost"`
	Type       string `json:"type"`
	Rarity     string `json:"rarity"`
	SetCode    string `json:"setCode"`
	Number     string `json:"number"`
	Text       string `json:"text"`
	FlavorText string `json:"flavorText"`
	Artist     string `json:"artist"`
	Power      string `json:"power"`
	Toughness  string `json:"toughness"`
}

// antwortDTO ist der Umschlag von GraphQL.
//
// **`errors` steht neben `data` und nicht statt seiner.** Eine Antwort kann
// beides tragen: einen Teil der Daten **und** einen Fehler. Wer nur `data`
// läse, bekäme eine halbe Antwort, die wie eine ganze aussieht.
type antwortDTO struct {
	Daten struct {
		Karten []kartenDTO `json:"cards"`
	} `json:"data"`

	Fehler []fehlerDTO `json:"errors"`
}

// fehlerDTO ist ein Fehler, wie GraphQL ihn schickt.
type fehlerDTO struct {
	Meldung     string `json:"message"`
	Erweiterung struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

// Klient fragt die GraphQL-Schnittstelle.
type Klient struct {
	adresse string
	klient  *netz.Klient
	drossel *netz.Drossel
	marke   string
}

// NeuerKlient baut einen Klienten.
//
// **Die Marke kommt herein und wird nicht hier geholt.** Wo sie liegt, weiß
// `internal/schluesselbund`; dieses Paket ist ein Blatt und kennt es nicht.
func NeuerKlient(adresse string, klient *netz.Klient, marke string) *Klient {
	return &Klient{
		adresse: adresse,
		klient:  klient,
		drossel: netz.NeueDrossel(Abstand, time.Now, time.Sleep),
		marke:   marke,
	}
}

// MitDrossel setzt die Drossel — **für Tests**.
func (k *Klient) MitDrossel(drossel *netz.Drossel) *Klient {
	k.drossel = drossel

	return k
}

// kartenabfrage ist die Abfrage nach Karten eines Namens.
//
// **Die Gestalt ist am echten Schema geprüft** (2026-09-02): `cards` nimmt
// `page: PaginationInput!` (`take`, `skip`), `order: ListOrderInput!`
// (`order: ASC|DESC`) und `filter: CardEntityFilterInput!` — und der
// Namensfilter heißt `name_eq`, nicht `name`.
//
// **Der erste Anlauf schrieb `cards(name: …)`** und bekam „Unknown argument
// \"name\" on field \"Query.cards\". Did you mean \"page\"?". Das Schema
// stand zur Verfügung; geraten war es trotzdem.
const kartenabfrage = `query Karten($name: String!) {
  cards(page: {take: 100, skip: 0}, order: {order: ASC},
        filter: {name_eq: $name}) {
    uuid name manaCost type rarity setCode number
    text flavorText artist power toughness
  }
}`

// Karten fragt nach Karten eines Namens.
//
// **Ohne Marke wird gar nicht erst gefragt** (§16, Regel 4): der Dienst
// antwortet ohne sie mit „Access denied", und eine Anfrage, deren Antwort
// feststeht, ist eine zu viel.
func (k *Klient) Karten(ctx context.Context, name string) ([]Karte, error) {
	return k.Abfragen(ctx, kartenabfrage, map[string]any{"name": name})
}

// Abfragen schickt eine beliebige GraphQL-Abfrage.
//
// # Warum es sie neben [Klient.Karten] gibt
//
// **Weil ein GraphQL-Klient genau das anbietet.** `Karten` ist die eine
// Abfrage, die dieses Programm heute stellt; das Schema kennt sieben Wurzeln
// (`status`, `cards`, `cardImage`, `decks`, `metadata`, `prices`, `sets`,
// gemessen am 2026-09-02). Wer die zweite braucht, baut sie hier und nicht
// noch einmal daneben.
//
// **Und der Fehlerweg beim Schreiben der Anfrage lebt hier**: `Karten` reicht
// Zeichenketten durch und kann ihn nicht betreten — ein Zweig, den kein
// Aufruf betritt, ist eine Zusage ohne Rumpf (§3, §10).
func (k *Klient) Abfragen(
	ctx context.Context, abfrage string, variablen map[string]any,
) ([]Karte, error) {
	if err := k.pruefeMarke(); err != nil {
		return nil, err
	}

	rumpf, baufehler := rumpfBauen(abfrage, variablen)
	if baufehler != nil {
		return nil, baufehler
	}

	kenner := netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Marke:   k.marke,
	}

	antwort, holfehler := netz.SendeJSON(
		ctx, k.klient, k.drossel, kenner, k.adresse, rumpf, hoechstens)
	if holfehler != nil {
		return nil, fmt.Errorf("MTGJSON fragen: %w", holfehler)
	}

	return lies(antwort.Rumpf)
}

// pruefeMarke gibt einen Fehler, wenn keine Zugangsmarke da ist.
func (k *Klient) pruefeMarke() error {
	if strings.TrimSpace(k.marke) != "" {
		return nil
	}

	return fmt.Errorf(
		"%w: für MTGJSON fehlt die Zugangsmarke — sie gehört in den "+
			"Schlüsselbund, und es gibt sie nur für Patreon-Unterstützende",
		domain.ErrValidation)
}

// rumpfBauen macht aus Abfrage und Variablen den JSON-Rumpf.
//
// # Warum es sie als eigene Funktion gibt
//
// **Damit ihr Fehlerweg prüfbar ist.** Stand der `json.Marshal` im Aufrufer,
// war sein Fehlerzweig unerreichbar: eine Abbildung aus Zeichenketten
// scheitert nie. Ein Zweig, den kein Aufruf betreten kann, ist eine Zusage
// ohne Rumpf (§3, §10) — und `.covignore` ist dafür ausdrücklich nicht da:
// „Eine unbequeme Verzweigung ist kein Grund."
//
// **Hier ist er erreichbar**, weil die Variablen von außen kommen: ein Wert,
// den JSON nicht kennt, gibt einen Fehler statt einer halben Anfrage.
func rumpfBauen(abfrage string, variablen map[string]any) ([]byte, error) {
	rumpf, err := json.Marshal(map[string]any{
		"query":     abfrage,
		"variables": variablen,
	})
	if err != nil {
		return nil, domain.MitGrund(domain.ErrValidation, err,
			"die Abfrage an MTGJSON lässt sich nicht als JSON schreiben")
	}

	return rumpf, nil
}

// Lies deutet den Umschlag — **öffentlich für den Vertragstest**.
//
// §16 verlangt einen Test gegen eingecheckte Beispiele; er braucht einen Weg
// hinein, der nicht durch das Netz führt. Ein Vertragstest, der einen
// Testserver bräuchte, prüfte den Testserver mit.
func Lies(rumpf []byte) ([]Karte, error) { return lies(rumpf) }

// lies deutet den Umschlag.
//
// # Warum der Statuscode hier nichts entscheidet
//
// **Nachgemessen am 2026-09-02**: „Access denied" kommt mit **200**, ein
// Feldfehler mit **400** — und beide tragen ihre Auskunft in `errors`. Wer
// den Code läse, sähe im ersten Fall einen Erfolg (§16).
func lies(rumpf []byte) ([]Karte, error) {
	var antwort antwortDTO

	if err := json.Unmarshal(rumpf, &antwort); err != nil {
		return nil, domain.MitGrund(domain.ErrValidation, err,
			"die Antwort von MTGJSON ist kein gültiges JSON")
	}

	if len(antwort.Fehler) > 0 {
		return nil, alsFehler(antwort.Fehler)
	}

	karten := make([]Karte, 0, len(antwort.Daten.Karten))

	for _, dto := range antwort.Daten.Karten {
		if err := pflichtfelder(dto); err != nil {
			return nil, err
		}

		karten = append(karten, alsKarte(dto))
	}

	return karten, nil
}

// pflichtfelder prüft, was dasein muss.
//
// # Zwei, und `setCode` ist keines
//
// **`uuid` und `name`**: ohne die Kennung lässt sich die Karte nicht
// wiederfinden, ohne den Namen nicht anzeigen. Beides macht sie unbrauchbar,
// und eine unbrauchbare Karte still durchzulassen wäre die schlechteste
// Antwort (§8).
//
// **`setCode` ist hier kein Pflichtfeld** — anders als bei
// magicthegathering.io. Der Grund steht im Schema (gemessen am 2026-09-02):
// MTGJSON kennt `AtomicCard`, eine Karte **ohne** Ausgabe. Wer `setCode`
// verlangte, wiese eine gültige Antwort ab.
//
// # Die Reihenfolge ist festgeschrieben
//
// Dieselbe Überlegung wie in `internal/scryfall` und `internal/mtgapi`: ein
// `map`-Literal würfelte sie, und dieselbe Antwort nennte zweimal ein anderes
// fehlendes Feld (#1289).
func pflichtfelder(dto kartenDTO) error {
	for _, fall := range []struct{ feld, wert string }{
		{"uuid", dto.UUID},
		{"name", dto.Name},
	} {
		if strings.TrimSpace(fall.wert) == "" {
			return fmt.Errorf(
				"%w: die Karte von MTGJSON hat kein Feld %q — das ist ein Pflichtfeld",
				domain.ErrValidation, fall.feld)
		}
	}

	return nil
}

// alsFehler macht aus den GraphQL-Fehlern einen Go-Fehler.
//
// **„Access denied" wird als Zugangsfehler gedeutet und nicht als
// Serverfehler.** Der Dienst nennt dafür `INTERNAL_SERVER_ERROR`; das ist
// irreführend, und wer den Code weiterreichte, meldete „der Dienst ist
// kaputt", wo „deine Marke gilt nicht" gemeint ist (§8).
func alsFehler(fehler []fehlerDTO) error {
	meldungen := make([]string, 0, len(fehler))
	for _, eine := range fehler {
		meldungen = append(meldungen, eine.Meldung)
	}

	zusammen := strings.Join(meldungen, "; ")

	if strings.Contains(zusammen, "Access denied") {
		return fmt.Errorf(
			"%w: MTGJSON weist den Zugang ab — die Zugangsmarke fehlt oder "+
				"gilt nicht: %s", domain.ErrValidation, zusammen)
	}

	return fmt.Errorf("%w: MTGJSON meldet: %s", domain.ErrValidation, zusammen)
}

// alsKarte ist die Zuordnung — **an einer Stelle** (§16).
func alsKarte(dto kartenDTO) Karte {
	return Karte{
		Kennung:    dto.UUID,
		Name:       dto.Name,
		Manakosten: dto.ManaCost,
		Typzeile:   dto.Type,
		Seltenheit: dto.Rarity,
		Setcode:    dto.SetCode,
		Nummer:     dto.Number,
		Regeltext:  dto.Text,
		Flavortext: dto.FlavorText,
		Kuenstler:  dto.Artist,
		Kraft:      dto.Power,
		Widerstand: dto.Toughness,
	}
}
