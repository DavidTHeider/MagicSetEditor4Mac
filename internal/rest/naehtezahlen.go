// **Kennzahlen und Verlauf** — die Nähte der Auszähl- und Bewegungsendpunkte,
// aus `naehte.go` herausgelöst, damit die Datei unter der Zeilengrenze bleibt
// (§2; Werkstück O3 von Phase 5.2 hat sie mit der Austauschnaht darüber
// geschoben).

package rest

import "context"

// Kennzahlen ist, was die zwei Auszählendpunkte brauchen.
type Kennzahlen interface {
	Statistik(ctx context.Context, setID int64) (Setzahlen, error)
	Verteilung(ctx context.Context, setID int64, feld string) ([]Wertzahl, error)
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
// Dieselbe Gestalt wie am Fenster und über MCP.
type Wertzahl struct {
	Name      string   `json:"name"`
	Karten    int      `json:"karten"`
	KartenIDs []string `json:"kartenIDs"`
}

// Verlauf ist, was die zwei Bewegungsendpunkte brauchen.
type Verlauf interface {
	// Zuruecknehmen nimmt den jüngsten geltenden Schritt des Sets zurück —
	// oder mit `vorgang` **genau diesen** (#1393).
	Zuruecknehmen(ctx context.Context, setID int64, vorgang string) (Bewegung, error)
	Wiederherstellen(ctx context.Context, setID int64, vorgang string) (Bewegung, error)

	// Verlauf gibt die Schritte des Sets, der jüngste zuerst (U6, #1417).
	Verlauf(ctx context.Context, setID int64) ([]Verlaufsschritt, error)
}

// Verlaufsschritt ist ein Schritt des Verlaufs — dieselbe Gestalt wie am
// Fenster und über MCP.
type Verlaufsschritt struct {
	ID              string `json:"id"`
	Zeitpunkt       string `json:"zeitpunkt"`
	Handlung        string `json:"handlung"`
	KarteID         string `json:"karteID"`
	Feld            string `json:"feld"`
	Vorher          string `json:"vorher"`
	GabVorher       bool   `json:"gabVorher"`
	Nachher         string `json:"nachher"`
	GabNachher      bool   `json:"gabNachher"`
	Zurueckgenommen bool   `json:"zurueckgenommen"`
	Urheber         string `json:"urheber"`
	Herkunft        string `json:"herkunft"`
	Vorgang         string `json:"vorgang"`
	Umfang          int    `json:"umfang"`
}

// Bewegung sagt, was eine Rücknahme bewegt hat.
//
// **`Bewegt` unterscheidet „nichts mehr da" von „getan"** — ohne es sähe ein
// Aufrufer keinen Unterschied zwischen dem Ende des Verlaufs und einem Schritt.
type Bewegung struct {
	Bewegt  bool   `json:"bewegt"`
	KarteID string `json:"karteID"`
	Feld    string `json:"feld"`
	Umfang  int    `json:"umfang"`
	Urheber string `json:"urheber"`
}
