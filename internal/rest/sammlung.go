// **Die Sammlung über die Rückschleife** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// # Warum die Sammlung keine Unterressource eines Sets ist
//
// Weil sie an keinem hängt. `/sammlung` steht neben `/sets` und nicht darin:
// was jemand besitzt, ist ein **Druck**, kein Entwurf
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
//
// # Und warum `POST` beim Preisholen
//
// Es **erzeugt** eine Momentaufnahme und fragt dafür einen fremden Dienst.
// Ein `GET` müsste zwischenspeicherbar sein; dieser Weg ist es nicht.

package rest

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungswerkstatt ist, was der Dienst für die Sammlung braucht.
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
type Exemplarsicht struct {
	ID int64 `json:"id"`

	// Fassung ist die Fassung, die gelesen wurde — sie geht beim Ändern
	// zurück (§7, Regel 2).
	//
	// **Ohne sie kann diese Fläche gar nicht ändern.** Am Speicher stünde
	// immer 0, und `WHERE fassung = 0` trifft keine Zeile: der Dienst
	// antwortete auf jedes `PUT` mit 409. Gemessen vom Nachlauf über
	// Phase 3.4 an einem wirklich gestarteten `mse-rest` — und der
	// Abnahmelauf sah es nicht, weil er nur anlegte und las.
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

// sammlung zählt die Sammlung auf.
func (d *Dienst) sammlung(schreiben http.ResponseWriter, anfrage *http.Request) {
	// **Dieselbe Adresse, wahlweise enger** — Werkstück
	// #1500.
	// Wer nichts angibt, bekommt alles, wie bisher; wer eine Sicht oder einen
	// Ausschnitt nennt, bekommt ihn samt beider Zahlen. Ein zweiter Weg für
	// dieselbe Frage wären zwei Verträge.
	abfrage := anfrage.URL.Query()

	if istGanzeSammlung(abfrage) {
		exemplare, fehler := d.naehte.Sammlung.Sammlung(anfrage.Context())
		if fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		ja(schreiben, http.StatusOK, map[string]any{"exemplare": ohneNil(exemplare)})

		return
	}

	fenster, zahlen, fehler := d.naehte.Sammlung.Sammlungsfenster(
		anfrage.Context(),
		abfrage.Get("suche"), abfrage.Get("ausfuehrung"), abfrage.Get("sortierung"),
		zahlAus(abfrage.Get("ab")), zahlAus(abfrage.Get("anzahl")))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"exemplare": ohneNil(fenster), "gesamt": zahlen.Gesamt,
		"insgesamt": zahlen.Insgesamt, "stueck": zahlen.Stueck,
	})
}

// istGanzeSammlung sagt, ob dieser Ruf ohne Sicht und ohne Ausschnitt kam.
func istGanzeSammlung(abfrage url.Values) bool {
	for _, teil := range []string{"suche", "ausfuehrung", "sortierung", "ab", "anzahl"} {
		if abfrage.Has(teil) {
			return false
		}
	}

	return true
}

// zahlAus liest eine Zahl aus der Abfrage — was keine ist, ist null.
func zahlAus(wert string) int {
	zahl, fehler := strconv.Atoi(wert)
	if fehler != nil || zahl < 0 {
		return 0
	}

	return zahl
}

// sammlungFuer zählt auf, was von einem Druck da ist.
func (d *Dienst) sammlungFuer(schreiben http.ResponseWriter, anfrage *http.Request) {
	exemplare, fehler := d.naehte.Sammlung.SammlungFuer(anfrage.Context(),
		anfrage.PathValue("printing"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"exemplare": ohneNil(exemplare)})
}

// exemplarAnlegen nimmt ein Exemplar auf.
func (d *Dienst) exemplarAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var exemplar Exemplarsicht
	if !gelesen(schreiben, anfrage, &exemplar) {
		return
	}

	id, fehler := d.naehte.Sammlung.ExemplarAnlegen(anfrage.Context(), exemplar)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"id": id})
}

// exemplarAendern schreibt ein Exemplar zurück.
//
// **Die Kennung steht im Weg, nicht im Rumpf.** Eine im Rumpf, die von der im
// Weg abweicht, wäre zwei Aussagen über dasselbe — und der Dienst müsste
// entscheiden, welche gilt.
func (d *Dienst) exemplarAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	id, gut := kennung(schreiben, anfrage, "id")
	if !gut {
		return
	}

	var exemplar Exemplarsicht
	if !gelesen(schreiben, anfrage, &exemplar) {
		return
	}

	exemplar.ID = id

	if fehler := d.naehte.Sammlung.ExemplarAendern(anfrage.Context(), exemplar); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"id": id})
}

// exemplarLoeschen nimmt ein Exemplar aus der Sammlung.
func (d *Dienst) exemplarLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	id, gut := kennung(schreiben, anfrage, "id")
	if !gut {
		return
	}

	// **Die Fassung steht in der Abfrage.** Ein `DELETE` hat keinen Rumpf, und
	// ohne sie nähme das Löschen eine fremde Änderung mit, die niemand gesehen
	// hat (Befund 10 des Nachlaufs über Phase 3.4).
	//
	// **Kein `If-Match`**: das setzte ETags an jeder Leseantwort voraus, und
	// die gibt dieser Dienst nirgends heraus. Eine halbe Bedingung wäre
	// schlimmer als eine offene.
	fassung, gutefassung := abfragezahl(schreiben, anfrage, "fassung")
	if !gutefassung {
		return
	}

	if fehler := d.naehte.Sammlung.ExemplarLoeschen(
		anfrage.Context(), id, fassung); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"id": id})
}
