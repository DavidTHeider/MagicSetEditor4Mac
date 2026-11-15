// **Die Sammlung über MCP** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Wofür ein Sprachmodell das braucht
//
// „Trag mir die drei Karten ein, die ich gestern gekauft habe" ist genau die
// Art Aufgabe, für die es sich lohnt, nicht zu klicken. Und §16 verlangt es
// ohnehin: was das Fenster kann, können die anderen Flächen.
//
// # Was hier ausdrücklich nicht steht
//
// **Ein Weg, der Preise holt.** Der geht ins Netz, und das gehört in
// `preise.go` daneben — nicht, weil es nicht ginge, sondern damit ein Modell,
// das Exemplare einträgt, nicht nebenbei dreihundert Abrufe auslöst.

package mcp

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungswerkstatt ist, was die Werkzeuge der Sammlung brauchen.
type Sammlungswerkstatt interface {
	Sammlung(ctx context.Context) ([]Exemplarsicht, error)

	// Sammlungsfenster gibt den Ausschnitt einer Sicht samt beider Zahlen —
	// Werkstück #1500.
	Sammlungsfenster(ctx context.Context, suche, ausfuehrung, sortierung string,
		ab, anzahl int) (fenster []Exemplarsicht, zahlen domain.Fensterzahlen, fehler error)
	SammlungFuer(ctx context.Context, printing string) ([]Exemplarsicht, error)
	ExemplarAnlegen(ctx context.Context, exemplar Exemplarsicht) (int64, error)
	ExemplarAendern(ctx context.Context, exemplar Exemplarsicht) error
	ExemplarLoeschen(ctx context.Context, id, fassung int64) error
}

// Exemplarsicht ist ein Exemplar, wie diese Fläche es sieht.
//
// **Eine eigene Gestalt und kein `domain.Exemplar`** (§2): `mcp` ist ein Blatt
// und kennt nur die Domäne — aber der **Preis als Text** ist eine Entscheidung
// der Naht, nicht der Domäne, und sie steht in `internal/app`.
type Exemplarsicht struct {
	ID int64 `json:"id"`

	// Fassung ist die Fassung, die gelesen wurde — sie geht beim Ändern
	// zurück (§7, Regel 2). Der Grund und was ihr Fehlen gekostet hat,
	// steht bei [rest.Exemplarsicht].
	Fassung int64 `json:"fassung"`

	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Foil ist der **alte** Wahrheitswert — abgeleitet aus [Exemplarsicht.Finish]
	// und seit T2 abgekündigt ([0119](../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md)).
	//
	// **Er bleibt, weil draußen jemand steht, der ihn liest** (0033, Regel 6:
	// erweiternd). Beim Schreiben entscheidet `finish`, wenn es dasteht; sonst
	// entscheidet er.
	//
	// **Für `etched` ist er `false`** — eine geätzte Karte ist kein Foil,
	// sondern eine eigene Ausführung. Ein Gegenüber, das nur ihn kennt und
	// eine geätzte Zeile zurückschreibt, macht sie damit zu `nonfoil`; das ist
	// der Preis dieser Verträglichkeit und steht im Vermerk.
	Foil bool `json:"foil"`

	// Finish ist die physische Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish          string `json:"finish"`
	Erwerbsdatum    string `json:"erwerbsdatum"`
	Erwerbspreis    string `json:"erwerbspreis"`
	Erwerbswaehrung string `json:"erwerbswaehrung"`
	Notiz           string `json:"notiz"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette wie jede Kennung, leer für keine; Vorlaeufig sagt, dass
	// der Druck keinen Alias hat. Beides **erweiternd**: alte Aufrufer lesen
	// weiter `printing`.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	// Name, Setcode und Nummer sind der **lesbare Bezug** des Drucks
	// (#1200).
	//
	// **Der letzte gute Stand und nicht die Identität** — die ist `printing`.
	// Ohne Netz steht hier, was zuletzt bekannt war (§16, Regel 5); ohne
	// Angabe bleibt die Kennung, und `bezeichnung` sagt sie.
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Lagerort ist, wo das Stück Papier liegt — Freitext
	// (Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
	Lagerort string `json:"lagerort"`

	// Wunsch sagt, dass diese Zeile **kein Besitz** ist (0057).
	//
	// **Und sie gehört in die Eindeutigkeit**: denselben Druck kann jemand
	// zugleich besitzen und sich wünschen. Ohne dieses Feld an der Fläche war
	// die zweite Zeile hier ein Konflikt — angestrichen vom Nachlauf über
	// Phase 4.5.
	Wunsch bool `json:"wunsch"`

	// Wunschrang und Wunschart gelten **nur an einem Wunsch** (T4, #1469) —
	// `niedrig`, `normal`, `hoch`, `sehrhoch` und `genau`, `jederdruck`.
	// An einer besessenen Zeile stehen sie auf ihrer Vorgabe.
	Wunschrang string `json:"wunschrang"`
	Wunschart  string `json:"wunschart"`

	// Bezeichnung ist der Bezug, wie eine Fläche ihn zeigt — oder die Kennung.
	Bezeichnung string `json:"bezeichnung"`
}

// exemplarfelder beschreibt die Argumente eines Exemplars.
//
// **Einmal aufgeschrieben.** Anlegen und Ändern nehmen dieselben Felder; zwei
// Beschreibungen daneben liefen beim ersten neuen Feld auseinander.
func exemplarfelder() map[string]any {
	return map[string]any{
		"printing": text("Die Scryfall-Kennung des DRUCKS — eine UUID. " +
			"Nicht der Name der Karte: »Lightning Bolt« gibt es dutzendfach, " +
			"besessen wird ein bestimmter Druck."),
		"menge":   zahl("Wie viele Stück. Mindestens eines."),
		"zustand": text("mint, nearMint, excellent, good, lightPlayed, played oder poor. Ohne Angabe nearMint."),
		"sprache": text("Der Sprachcode des Drucks, wie Scryfall ihn führt. Ohne Angabe »en«."),
		"foil": jaNein("ABGEKÜNDIGT seit 0.5.0 — benutze »finish«. " +
			"Ob es die veredelte Ausführung ist. Er entscheidet nur, " +
			"solange »finish« fehlt, und kann »etched« nicht sagen."),
		"finish": text("Die physische Ausführung: nonfoil, foil oder etched. " +
			"Ohne Angabe nonfoil."),
		"erwerbsdatum": text("Der Tag des Erwerbs als JJJJ-MM-TT. " +
			"Leer, wenn niemand ihn weiß."),
		"erwerbspreis": text("Der bezahlte Preis, wie man ihn schreibt: »4.17«. " +
			"Leer heißt: nicht erfasst. Ein Komma statt eines Punktes wird abgewiesen."),
		"erwerbswaehrung": text("Der Währungscode zum Erwerbspreis, etwa »EUR«. " +
			"Gehört zusammen mit dem Preis: eines allein wird abgewiesen."),
		"notiz": text("Freier Text."),

		// **Der lesbare Bezug** (#1200) — wahlfrei, und ohne ihn steht die
		// Kennung da. Ein Zwang dazu hieße, jede Fläche ans Netz zu binden.
		"name":    text("Der Kartenname zum Druck — wahlfrei, damit die Zeile lesbar ist."),
		"setcode": text("Der Setcode zum Druck, etwa »m11« — wahlfrei."),
		"nummer":  text("Die Sammlernummer zum Druck, etwa »149« — wahlfrei."),

		// **Ordnung und Absicht** (0057) — beides wahlfrei, und ohne Angabe
		// ist es ein Besitz ohne Lagerort.
		"lagerort": text("Wo das Stück liegt — Freitext: »Kiste 4«, " +
			"»Ordner Rot, Seite 12«, »bei Anna«. Leer heißt: nicht erfasst."),
		"wunsch": jaNein("Ob diese Zeile ein WUNSCH ist statt Besitz. " +
			"Ohne Angabe ist es Besitz. Denselben Druck kann jemand zugleich " +
			"besitzen und sich wünschen — das sind zwei Zeilen."),

		// **Nur an einem Wunsch** (T4, #1469): an einem Besitz bedeuten sie
		// nichts, und die Domäne räumt sie dort weg.
		"wunschrang": text("Wie dringend dieser WUNSCH ist: niedrig, normal, " +
			"hoch oder sehrhoch. Ohne Angabe normal. An einem Besitz bedeutet " +
			"er nichts."),
		"wunschart": text("Woran dieser WUNSCH hängt: »genau« meint genau " +
			"diesen Druck, »jederdruck« die Karte in jedem Druck. Ohne Angabe " +
			"»genau« — die engere Lesart, damit niemandem etwas angeboten " +
			"wird, das er nicht genau so gewollt hat."),
	}
}

// sammlungswerkzeuge gibt die fünf Werkzeuge der Sammlung.
func sammlungswerkzeuge(werkstatt Sammlungswerkstatt) []Werkzeug {
	return append(sammlungslesewerkzeuge(werkstatt),
		sammlungsschreibwerkzeuge(werkstatt)...)
}

// sammlungslesewerkzeuge gibt die beiden lesenden.
func sammlungslesewerkzeuge(werkstatt Sammlungswerkstatt) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.sammlung", Art: Lesend,
			Erklaerung: "Zählt die eigene Kartensammlung auf — je Druck, Zustand, " +
				"Sprache und Ausführung eine Zeile. ÄNDERT NICHTS. Das sind " +
				"BESESSENE Karten, nicht die entworfenen eines Sets: eine Sammlung " +
				"hängt an keinem Set. **Bei großen Sammlungen einen Ausschnitt " +
				"holen**: »anzahl« begrenzt die Zeilen, »ab« überspringt welche, " +
				"»suche« filtert über Name, Druck, Setcode und Notiz, " +
				"»ausfuehrung« über »foil«/»nonfoil«, »sortierung« ordnet nach " +
				"»bezeichnung«, »menge« oder »erwerbsdatum«. Ohne Angaben kommt " +
				"alles. »gesamt« nennt die Zeilen der Sicht, »insgesamt« die der " +
				"ganzen Sammlung.",
			Schema: objekt(map[string]any{
				"suche":       text("filtert über Name, Druck, Setcode und Notiz"),
				"ausfuehrung": text("»alle«, »foil« oder »nonfoil«"),
				"sortierung":  text("»bezeichnung«, »menge« oder »erwerbsdatum«"),
				"ab":          zahl("wie viele Zeilen übersprungen werden"),
				"anzahl":      zahl("wie viele Zeilen höchstens kommen; 0 heißt alle"),
			}),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Suche       string `json:"suche"`
					Ausfuehrung string `json:"ausfuehrung"`
					Sortierung  string `json:"sortierung"`
					Ab          int    `json:"ab"`
					Anzahl      int    `json:"anzahl"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				fenster, zahlen, fehler := werkstatt.Sammlungsfenster(ctx,
					argumente.Suche, argumente.Ausfuehrung, argumente.Sortierung,
					argumente.Ab, argumente.Anzahl)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{
					"Exemplare": ohneNil(fenster),
					"gesamt":    zahlen.Gesamt,
					"insgesamt": zahlen.Insgesamt,
					"stueck":    zahlen.Stueck,
				})
			},
		},
		{
			Name: "lesen.sammlungfuer", Art: Lesend,
			Erklaerung: "Zählt auf, was von einem bestimmten Druck da ist. ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{
				"printing": text("Die Scryfall-Kennung des Drucks."),
			}, "printing"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Printing string `json:"printing"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				exemplare, fehler := werkstatt.SammlungFuer(ctx, argumente.Printing)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Exemplare": ohneNil(exemplare)})
			},
		},
	}
}

// sammlungsschreibwerkzeuge gibt die drei schreibenden.
//
// # Keines sagt Idempotenz zu
//
// **Kein schreibendes Werkzeug dieses Servers tut das**, und der Grund steht in
// `idempotenz_test.go`: eine Änderung ist eine Wirkung auf die Umgebung, und
// ein zweiter Ruf ist eine zweite.
//
// Für `schreiben.exemplaraendern` sähe es verlockend aus — zweimal dieselben
// Werte gesetzt gibt zweimal dieselbe Antwort. Für `schreiben.exemplarloeschen`
// nicht: der zweite Ruf sagt „die Zeile gibt es nicht". Lieber zu wenig
// zusagen als eine Zusage, die an einer Vorrichtung hängt, die großzügiger ist
// als der Betrieb (§3).
func sammlungsschreibwerkzeuge(werkstatt Sammlungswerkstatt) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.exemplar", Art: Schreibend,
			Erklaerung: "ÄNDERT DIE SAMMLUNG: nimmt ein Exemplar auf. Dieselbe Karte in " +
				"demselben Zustand, derselben Sprache und derselben Ausführung " +
				"gibt es nur EINMAL — ein zweites Anlegen wird abgewiesen; wer die " +
				"Menge erhöhen will, ändert die Zeile.",
			Schema: objekt(exemplarfelder(), "printing", "menge"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var exemplar Exemplarsicht
				if err := lies(roh, &exemplar); err != nil {
					return "", err
				}

				id, fehler := werkstatt.ExemplarAnlegen(ctx, exemplar)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"ID": id})
			},
		},
		{
			Name: "schreiben.exemplaraendern", Art: Schreibend,
			Erklaerung: "ÄNDERT DIE SAMMLUNG: schreibt ein Exemplar zurück. KENNUNG und FASSUNG " +
				"kommen aus »lesen.sammlung«. Alle Felder werden gesetzt, auch die " +
				"weggelassenen — wer nur die Menge ändern will, schickt den Rest mit.",
			Schema: objekt(mitKennung(exemplarfelder()),
				"id", "fassung", "printing", "menge"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var exemplar Exemplarsicht
				if err := lies(roh, &exemplar); err != nil {
					return "", err
				}

				if fehler := werkstatt.ExemplarAendern(ctx, exemplar); fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Geaendert": exemplar.ID})
			},
		},
		{
			Name: "schreiben.exemplarloeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT DIE SAMMLUNG: nimmt ein Exemplar heraus. KENNUNG und FASSUNG " +
				"kommen aus »lesen.sammlung«. Das ist eine Aussage " +
				"über den BESITZ, nicht über eine Karte — die Karte bleibt, wo sie " +
				"ist. Ein zweiter Ruf sagt, dass es die Zeile nicht gibt.",
			Schema: objekt(map[string]any{
				"id": zahl("Die Kennung des Exemplars aus »lesen.sammlung«."),
				"fassung": zahl("Die Fassung aus »lesen.sammlung«. " +
					"Stimmt sie nicht mehr, hat jemand anderes geschrieben, " +
					"und der Ruf wird abgewiesen — sonst nähme dein Löschen " +
					"eine Änderung mit, die du nie gesehen hast."),
			}, "id", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					ID      int64 `json:"id"`
					Fassung int64 `json:"fassung"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if fehler := werkstatt.ExemplarLoeschen(ctx, argumente.ID, argumente.Fassung); fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Geloescht": argumente.ID})
			},
		},
	}
}

// mitKennung ergänzt die Felder um die Kennung.
func mitKennung(felder map[string]any) map[string]any {
	felder["id"] = zahl("Die Kennung des Exemplars aus »lesen.sammlung«.")

	// **Und die Fassung dazu.** Ohne sie kommt am Speicher 0 an, und
	// `WHERE fassung = 0` trifft keine Zeile: **jeder** Ruf endete im
	// Konflikt. Ein Modell konnte über diese Fläche nichts ändern.
	felder["fassung"] = zahl("Die Fassung aus »lesen.sammlung«. " +
		"Sie sagt, welchen Stand du gelesen hast — stimmt sie nicht mehr, " +
		"hat jemand anderes geschrieben, und der Ruf wird abgewiesen, " +
		"statt die fremde Änderung zu überschreiben. Erst wieder lesen.")

	return felder
}

// ohneNil macht aus einer fehlenden Liste eine leere.
//
// **Sonst stünde `null` in der Antwort**, wo ein Modell eine Liste erwartet —
// dieselbe Falle wie am Fenster (§3).
func ohneNil(exemplare []Exemplarsicht) []Exemplarsicht {
	if exemplare == nil {
		return []Exemplarsicht{}
	}

	return exemplare
}
