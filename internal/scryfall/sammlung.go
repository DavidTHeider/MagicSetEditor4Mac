package scryfall

// **Viele Karten in einer Anfrage** — `POST /cards/collection`
// (#1348).
//
// # Was am 2026-08-31 nachgelesen wurde (§16, Regel 1)
//
// Unter <https://scryfall.com/docs/api/cards/collection>:
//
//   - **Höchstens 75 Kennungen je Anfrage.**
//   - Der Weg zählt zu den Kartenwegen: **2/s**, also dieselbe Drossel wie
//     `/cards/search`.
//   - Die Anfrage muss `Content-Type: application/json` tragen.
//   - Der Rumpf ist `{"identifiers": [ … ]}`.
//   - Gültige Kennungsformen sind unter anderem `name`, `name`+`set` und
//     `collector_number`+`set`. **Jede Kennung gibt höchstens eine Karte.**
//   - Was nicht gefunden wird, steht in `not_found`.
//   - Wörtlich: „While cards will be returned in the order that they were
//     requested, cards that aren't found will throw off the mapping of request
//     identifiers to results, so you should not rely on positional index alone
//     while parsing the data."
//
// **Der letzte Satz ist der Grund für [Sammelkarte].** Wer die Antwort über
// den Index zuordnete, hätte bei der ersten fehlenden Karte alles Folgende
// verschoben — und zwar still.
//
// # Warum das nicht [Klient.Druckesuche] ersetzt
//
// Weil es die andere Frage beantwortet. Eine Kennung gibt **eine** Karte;
// „Counterspell" ohne Set gibt die neueste Ausgabe und nicht die 72, unter
// denen jemand wählen soll. Dieser Weg ist für Zeilen, die ihren Druck schon
// benennen — und das sind in einer ausgeführten Deckliste fast alle.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Sammelgrenze ist die Zahl der Kennungen, die in eine Anfrage passen.
//
// **75, nachgelesen** — „A maximum of 75 card references may be submitted per
// request." Wer mehr schickt, bekommt kein Ergebnis, sondern einen Fehler.
const Sammelgrenze = 75

// Kennung benennt eine Karte für die Sammelanfrage.
//
// **Nur die vier Formen, die dieses Programm stellt.** Scryfall kennt acht;
// `mtgo_id`, `multiverse_id`, `oracle_id` und `illustration_id` fragt hier
// niemand, und eine Form ohne Aufrufer wäre eine Zusage ohne Einlösung (§2).
type Kennung struct {
	// ID ist die Druckkennung, wie Scryfall sie vergibt — eine UUID.
	//
	// **Sie ist die genaueste Form**: sie meint einen Druck und nicht eine
	// Karte. Wer sie hat, braucht Set und Nummer nicht.
	//
	// **Nachgetragen am 2026-09-01** für die Formatprüfung, die je Platz
	// einzeln fragte (Befund 6 des Zweitblicks über Phase 4.6,
	// #1373).
	ID string `json:"id,omitempty"`

	// Name ist der Kartenname — allein gibt er die **neueste** Ausgabe.
	Name string `json:"name,omitempty"`

	// Setcode ist der Setcode, wie Scryfall ihn führt.
	Setcode string `json:"set,omitempty"`

	// Nummer ist die Sammlernummer — **eine Zeichenkette**, denn sie ist
	// keine Zahl: es gibt „149a" und „★12".
	Nummer string `json:"collector_number,omitempty"`
}

// Sammelantwort ist, was eine Sammelanfrage zurückgibt.
type Sammelantwort struct {
	// Karten sind die gefundenen — in der Reihenfolge der Anfrage, aber ohne
	// die fehlenden.
	Karten []*Karte

	// NichtGefunden sind die Kennungen, zu denen es nichts gibt.
	//
	// **Sie sind kein Fehler.** Eine Liste mit einem Tippfehler soll die
	// übrigen Zeilen trotzdem bekommen (§8).
	NichtGefunden []Kennung
}

// Sammlung holt bis zu [Sammelgrenze] Karten in **einer** Anfrage.
//
// **Sie teilt nicht selbst auf.** Wer mehr hat, ruft mehrmals — und sieht
// dabei, was es kostet. Eine Funktion, die stillschweigend sechzig Anfragen
// stellte, wäre genau die Wartezeit ohne Lebenszeichen, gegen die es die
// Schübe gibt.
func (k *Klient) Sammlung(ctx context.Context, kennungen []Kennung) (
	antwort Sammelantwort, alter time.Duration, befunde []domain.Befund, fehler error,
) {
	if len(kennungen) == 0 {
		return Sammelantwort{Karten: []*Karte{}, NichtGefunden: []Kennung{}}, 0, nil, nil
	}

	if len(kennungen) > Sammelgrenze {
		return Sammelantwort{}, 0, nil, fmt.Errorf(
			"%w: %d Kennungen auf einmal — Scryfall nimmt höchstens %d",
			domain.ErrValidation, len(kennungen), Sammelgrenze)
	}

	// **Ohne Fehlerzweig, und das ist eine Aussage**: [Kennung] trägt drei
	// Zeichenketten, und `json.Marshal` scheitert daran nicht — es kann an
	// Kanälen, Funktionen und Zyklen scheitern, und keines davon steht hier.
	// Ein Zweig, den kein Test betreten kann, wäre nach §3 keiner, den es
	// geben darf.
	//nolint:errcheck // der Grund steht darüber
	rumpf, _ := json.Marshal(struct {
		Kennungen []Kennung `json:"identifiers"`
	}{Kennungen: kennungen})

	roh, alter, fehler := k.speicher.Hole(ctx, sammelschluessel(rumpf), func() ([]byte, error) {
		return k.sende(ctx, "/cards/collection", rumpf)
	})
	if fehler != nil {
		return Sammelantwort{}, 0, nil, fehler
	}

	return liesSammelantwort(roh, alter)
}

// sende stellt die Anfrage und deutet den Status wie [Klient.holeMit].
func (k *Klient) sende(ctx context.Context, weg string, rumpf []byte) ([]byte, error) {
	antwort, fehler := netz.SendeJSON(ctx, k.klient, k.drossel, k.kenner,
		k.grundadresse+weg, rumpf, hoechstens)
	if fehler != nil {
		return nil, fehler
	}

	// **Ein 404 gibt es hier nicht**, und das ist kein Versehen: eine
	// Sammelanfrage, die nichts findet, gibt eine leere Liste und füllt
	// `not_found`. Ein 404 wäre eine Aussage über den **Weg**.
	if antwort.Status != netz.StatusGut {
		return nil, fmt.Errorf("%w: Scryfall antwortet mit %d",
			domain.ErrValidation, antwort.Status)
	}

	return antwort.Rumpf, nil
}

// sammelschluessel gibt den Namen, unter dem eine Sammelantwort liegt.
//
// **Über den Rumpf und nicht über die Adresse**, denn die ist für jede Anfrage
// dieselbe — der Rumpf trägt die Frage. Wer die Adresse nähme, bekäme für jede
// Sammelanfrage die Antwort der ersten.
//
// # Warum der Rumpf und nicht die Felder
//
// Hier standen bis zum 2026-09-01 die drei Felder, mit `|` und `\n`
// dazwischen. Das ging schief, sobald eine Kennung **nur** eine UUID trug: die
// drei Felder waren dreimal leer, jeder Schub bekam denselben Schlüssel, und
// der zweite bekam die Karten des ersten aus dem Zwischenspeicher — still,
// mit einem plausiblen Ergebnis.
//
// **Und es war nicht der einzige Fall.** Ein Kartenname mit einem `|` darin
// glich einem Namen mit Setcode; dasselbe Muster, nur seltener. Der
// marshallierte Rumpf **ist** die gestellte Frage: zwei verschiedene Fragen
// haben verschiedene Rümpfe. Länge kostet nichts, denn [netz.Speicher] legt
// jeden Schlüssel ohnehin als Prüfsumme ab.
func sammelschluessel(rumpf []byte) string {
	return "cards/collection:" + string(rumpf)
}

// liesSammelantwort deutet die Antwort — **streng in den Pflichtfeldern**.
//
// §16: „Pflichtfelder streng, Zusatzfelder tolerant." `object` und `data` sind
// Pflicht; ein fehlendes `not_found` heißt: nichts fehlte.
func liesSammelantwort(roh []byte, alter time.Duration) (
	Sammelantwort, time.Duration, []domain.Befund, error,
) {
	var liste struct {
		Objekt        string            `json:"object"`
		Daten         []json.RawMessage `json:"data"`
		NichtGefunden []Kennung         `json:"not_found"`
	}

	if err := json.Unmarshal(roh, &liste); err != nil {
		return Sammelantwort{}, 0, nil, domain.MitGrund(domain.ErrValidation, err,
			"die Sammelantwort von Scryfall ist kein gültiges JSON")
	}

	if liste.Objekt != "list" {
		return Sammelantwort{}, 0, nil, fmt.Errorf(
			"%w: die Sammelantwort von Scryfall hat kein Feld %q — das ist ein Pflichtfeld",
			domain.ErrValidation, "object")
	}

	var befunde []domain.Befund

	karten := make([]*Karte, 0, len(liste.Daten))

	for _, eine := range liste.Daten {
		gelesen, weitere, lesefehler := LiesKarte(eine)
		if lesefehler != nil {
			return Sammelantwort{}, 0, nil, lesefehler
		}

		karten = append(karten, gelesen)
		befunde = append(befunde, weitere...)
	}

	fehlend := liste.NichtGefunden
	if fehlend == nil {
		fehlend = []Kennung{}
	}

	return Sammelantwort{Karten: karten, NichtGefunden: fehlend}, alter, befunde, nil
}
