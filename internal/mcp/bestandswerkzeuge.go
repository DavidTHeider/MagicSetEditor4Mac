// **Pakete, Kennzahlen, Verlauf und Einstellungen als Werkzeuge**
// (#1107).
//
// # Warum ein Modell die Rücknahme braucht
//
// Es schreibt, während niemand hinsieht. Bis hierher konnte es jede Änderung
// nur einzeln zurückdrehen — über den Rückweg, den jedes schreibende Werkzeug
// nennt —, und für eine Massenänderung gab es keinen. `schreiben.zuruecknehmen`
// nimmt **einen Schritt** des Verlaufs zurück, gleich wie viele Karten er
// berührt hat.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Bestandsquelle ist, was die acht Werkzeuge dieser Datei brauchen.
type Bestandsquelle interface {
	Pakete(ctx context.Context) (pakete []Paketzeile, befunde []string, fehler error)
	PaketBeziehen(ctx context.Context, adresse, stand, unterverzeichnis, name string) (
		Paketbezug, error)
	Statistik(ctx context.Context, setID int64) (Setzahlen, error)
	Verteilung(ctx context.Context, setID int64, feld string) ([]Wertzahl, error)
	// Zuruecknehmen nimmt einen Schritt zurück — den jüngsten des Sets, oder
	// mit `vorgang` **genau diesen** (#1393).
	Zuruecknehmen(ctx context.Context, setID int64, vorgang string) (Bewegung, error)
	Wiederherstellen(ctx context.Context, setID int64, vorgang string) (Bewegung, error)

	// Verlauf gibt die Schritte des Sets, der jüngste zuerst (U6, #1417).
	Verlauf(ctx context.Context, setID int64) ([]Verlaufsschritt, error)
	Einstellungen(ctx context.Context) ([]Einstellungszeile, error)
	EinstellungSetzen(ctx context.Context, schluessel, wert string) error

	// **Und der Austausch** — Werkstück O3 von Phase 5.2
	// (#1427).
	// Die Werkzeuge dazu stehen in `einstellungsaustausch.go`; hier steht,
	// dass dieselbe Quelle sie bedient — eine zweite Naht für dieselbe Tafel
	// wären zwei Wege zu einer Sache.
	Austauschquelle
}

// Paketzeile ist ein installiertes Paket.
type Paketzeile struct {
	Art      string `json:"art"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Pfad     string `json:"pfad"`
	Herkunft string `json:"herkunft"`

	// Die Übersicht — DOM-01.AK-01 (#1696), **erweiternd** nach §16:
	// Prüfsumme des Bezugs, Herkunftsart (mitgeliefert, bezogen, uebernommen,
	// eigen), ob es vom Stand beim Einbinden abweicht, was es braucht, wer es
	// braucht und wie viele Sets es benutzen.
	Pruefsumme   string   `json:"pruefsumme"`
	Herkunftsart string   `json:"herkunftsart"`
	Geaendert    bool     `json:"geaendert"`
	Braucht      []string `json:"braucht"`
	GebrauchtVon []string `json:"gebrauchtVon"`
	Benutzt      int      `json:"benutzt"`
}

// Paketbezug ist der Ausgang eines Paketbezugs.
type Paketbezug struct {
	Ort         string   `json:"ort"`
	Stand       string   `json:"stand"`
	Pruefsumme  string   `json:"pruefsumme"`
	Uebergangen []string `json:"uebergangen"`
}

// Setzahlen sind die Kennzahlen eines Sets.
type Setzahlen struct {
	Karten   int    `json:"karten"`
	Felder   int    `json:"felder"`
	Leere    int    `json:"leere"`
	Belegung []Zahl `json:"belegung"`
}

// Zahl ist eine Zeile der Belegung.
type Zahl struct {
	Name   string `json:"name"`
	Karten int    `json:"karten"`
}

// Wertzahl ist eine Zeile der Verteilung — samt den Karten des Werts,
// aufsteigend (UXR-F6, #2013).
// Dieselbe Auskunft wie am Fenster: die Fläche, die einen Balken zählt, nennt
// auch seine Karten. Die Belegung nicht — eine Karte steht in vielen Feldern.
type Wertzahl struct {
	Name      string   `json:"name"`
	Karten    int      `json:"karten"`
	KartenIDs []string `json:"kartenIDs"`
}

// Bewegung sagt, was eine Rücknahme bewegt hat.
type Bewegung struct {
	Bewegt  bool   `json:"bewegt"`
	KarteID string `json:"karteID"`
	Feld    string `json:"feld"`
	Umfang  int    `json:"umfang"`
	Urheber string `json:"urheber"`
}

// Einstellungszeile ist eine Einstellung samt ihren erlaubten Werten.
type Einstellungszeile struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`
	Bereich    string   `json:"bereich"`

	// **Was gilt, wo es hingehört und von wem es kommt** — Werkstück O3 von
	// Phase 5.2 (#1427).
	//
	// **`Wirksam` weicht von `Wert` genau dann ab, wenn dort `erben` steht**
	// oder eine Schranke greift; `Ebene` sagt, wie weit die Einstellung
	// reist; `Herkunft` sagt, wessen Entscheidung der wirksame Wert war. Ein
	// Modell, das nur `Wert` sähe, setzte auf einer Ebene, auf der es nichts
	// bewirkt.
	Wirksam  string `json:"wirksam"`
	Ebene    string `json:"ebene"`
	Herkunft string `json:"herkunft"`
	// Wirkung sagt, **wann** ein geänderter Wert gilt — `sofort` oder
	// `neustart` (UX-04.AK-02,
	// #1676).
	// Dieselbe Antwort wie am Fenster: sie hängt daran, wie der Wert
	// **gelesen** wird, nicht daran, wo jemand ihn umlegt (§16).
	Wirkung string `json:"wirkung"`

	// Vorgabe ist der Wert, mit dem die Anwendung anfängt (#1422).
	Vorgabe string `json:"vorgabe"`
}

// bestandswerkzeuge gibt die sechs lesenden und die fünf schreibenden.
//
// **Die drei des Austauschs kommen aus `einstellungsaustausch.go`** —
// Werkstück O3 von Phase 5.2 (#1427). Sie stehen dort, weil die Datei sonst
// über die Zeilengrenze ginge (§2); registriert werden sie hier, damit es
// eine Stelle bleibt, an der man sieht, was der Bestand kann.
func bestandswerkzeuge(quelle Bestandsquelle) (lesend, schreibend []Werkzeug) {
	austauschlesend, austauschschreibend := austauschwerkzeuge(quelle)

	return append([]Werkzeug{
			paketeLesen(quelle), statistikLesen(quelle),
			verteilungLesen(quelle), einstellungenLesen(quelle),
		}, austauschlesend...),
		append([]Werkzeug{
			paketBeziehen(quelle), bewegungswerkzeug(quelle, true),
			bewegungswerkzeug(quelle, false), einstellungSetzen(quelle),
		}, austauschschreibend...)
}

// paketeLesen frischt den Katalog auf und nennt ihn.
func paketeLesen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.pakete", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die installierten Vorlagenpakete. Frischt den Katalog vorher von der Platte auf.",
		Schema:     objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			pakete, befunde, fehler := quelle.Pakete(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"pakete": liste(pakete), "befunde": liste(befunde),
			})
		},
	}
}

// paketBeziehen holt ein Vorlagenpaket.
//
// **Es steht bei den schreibenden, obwohl es kein Set anfasst**: es lädt aus
// dem Netz und legt Dateien an. Ein Gegenüber, das die Liste auf die lesenden
// kürzt, soll es nicht behalten.
func paketBeziehen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.paket.beziehen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Rechner: lädt ein Vorlagenpaket aus dem Netz und legt es ab. Der Stand ist ein Commit, kein Zweig.",
		Schema: objekt(map[string]any{
			"adresse":          text("Die Adresse des Repositoriums."),
			"stand":            text("Der Commit-Hash — ein Zweig zeigt morgen woanders hin."),
			"unterverzeichnis": text("Das Verzeichnis im Archiv, das die Pakete enthält."),
			"name":             text("Der Name, unter dem abgelegt wird."),
		}, "adresse", "stand", "unterverzeichnis", "name"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				Adresse          string `json:"adresse"`
				Stand            string `json:"stand"`
				Unterverzeichnis string `json:"unterverzeichnis"`
				Name             string `json:"name"`
			}

			if err := json.Unmarshal(argumente, &wahl); err != nil {
				return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
			}

			ergebnis, fehler := quelle.PaketBeziehen(
				ctx, wahl.Adresse, wahl.Stand, wahl.Unterverzeichnis, wahl.Name)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// statistikLesen zählt ein Set aus.
func statistikLesen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.statistik", Art: Lesend, Idempotent: true,
		Erklaerung: "Zählt ein Set aus: Karten, Felder, leere Karten und die Belegung je Feld.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			zahlen, fehler := quelle.Statistik(ctx, wahl)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(zahlen)
		},
	}
}

// verteilungLesen sagt, wie viele Karten welchen Wert eines Felds tragen.
func verteilungLesen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.verteilung", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, wie viele Karten welchen Wert eines Feldes tragen, " +
			"und nennt zu jedem Wert die Karten (»kartenIDs«, aufsteigend).",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"feld":  text("Der Feldname aus »lesen.felder«."),
		}, "setID", "feld"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := optionswahl(argumente)
			if err != nil {
				return "", err
			}

			werte, fehler := quelle.Verteilung(ctx, wahl.SetID, wahl.Feld)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"verteilung": liste(werte)})
		},
	}
}

// einstellungenLesen nennt jede Einstellung samt ihren erlaubten Werten.
func einstellungenLesen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.einstellungen", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Einstellungen der Anwendung samt den erlaubten Werten.",
		Schema:     objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			alle, fehler := quelle.Einstellungen(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"einstellungen": liste(alle)})
		},
	}
}

// einstellungSetzen schreibt eine Einstellung fort.
func einstellungSetzen(quelle Bestandsquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.einstellung", Art: Schreibend,
		Erklaerung: "ÄNDERT eine Einstellung der Anwendung. Der Schlüssel und die erlaubten Werte kommen aus »lesen.einstellungen«.",
		Schema: objekt(map[string]any{
			"schluessel": text("Der Schlüssel aus »lesen.einstellungen«."),
			"wert":       text("Einer der dort genannten Werte."),
		}, "schluessel", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				Schluessel string `json:"schluessel"`
				Wert       string `json:"wert"`
			}

			if err := json.Unmarshal(argumente, &wahl); err != nil {
				return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
			}

			if fehler := quelle.EinstellungSetzen(
				ctx, wahl.Schluessel, wahl.Wert); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"gesetzt": wahl.Schluessel})
		},
	}
}
