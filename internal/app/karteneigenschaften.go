package app

// **Die Karteneigenschaften an der Naht** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wie eine Partie aussieht**; hier steht, **was eine Karte
// ist**. Und `partie.go` steht bei 371 Zeilen (§2).

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// EigenschaftenDTO ist, was eine Karte am Tisch ist — CR 109.3.
//
// **Ein Zeiger am Objekt und kein Verbund**: `null` sagt „hier gibt es
// keine", und das ist etwas anderes als „alle Werte sind leer". Ein
// Spielstein hat keine Karte; ein geheimes Objekt gibt sie nicht her.
type EigenschaftenDTO struct {
	// Name ist der Name der Karte — CR 201.1.
	//
	// **Nicht `ObjektDTO.Name`**: der trägt den Namen eines Spielsteins und
	// ist bei einer Karte leer. Der Grund steht ausführlich bei
	// [domain.Sichteigenschaften].
	Name string `json:"name"`

	// Farben sind die Buchstaben der Manasymbole — `W`, `U`, `B`, `R`, `G`
	// (CR 202.2a), **in der Reihenfolge des Regelwerks**.
	//
	// **Eine leere Liste heißt farblos** (CR 105.2c) — und es ist eine
	// leere Liste und kein `null`, wie überall an dieser Grenze.
	Farben []string `json:"farben"`

	Manawert int `json:"manawert"`

	Supertypen []string `json:"supertypen"`
	Typen      []string `json:"typen"`
	Untertypen []string `json:"untertypen"`

	// Staerke und Widerstandskraft **mit Marken** — CR 122.1a.
	Staerke          KartenwertDTO `json:"staerke"`
	Widerstandskraft KartenwertDTO `json:"widerstandskraft"`

	Loyalitaet   KartenwertDTO `json:"loyalitaet"`
	Verteidigung KartenwertDTO `json:"verteidigung"`

	// UnbekannteTypen sind die Wörter der Typzeile, die das Regelwerk nicht
	// kennt — §8: ein Befund ist kein Fehler.
	//
	// **Eine leere Liste und kein `null`**, wie überall an dieser Grenze.
	UnbekannteTypen []string `json:"unbekannteTypen"`
}

// KartenwertDTO ist eine Zahl auf einer Karte, die auch fehlen oder ein Stern
// sein darf — CR 208.2, CR 208.3.
//
// **`Kartenwert…` und nicht `Wert…`**: `WertDTO` ist in diesem Paket schon
// der Feldwert einer Karte im Bestand (`bestand.go`). Zwei Bedeutungen für
// einen Namen sind eine Falle.
//
// **Drei Felder und keine Zeichenkette**: `steht` unterscheidet „keine" von
// „null", `stern` sagt, dass eine Fähigkeit die Zahl bestimmt (CR 604.3).
// Wer daraus eine Zeichenkette machte, ließe jede Fläche wieder zerteilen.
type KartenwertDTO struct {
	Steht bool `json:"steht"`
	Stern bool `json:"stern"`
	Zahl  int  `json:"zahl"`
}

// eigenschaftenAus bildet die Sicht auf die Eigenschaften ab — **nil bleibt
// nil**.
func eigenschaftenAus(eigen *domain.Sichteigenschaften) *EigenschaftenDTO {
	if eigen == nil {
		return nil
	}

	farben := make([]string, 0, len(eigen.Farben))
	for _, farbe := range eigen.Farben {
		farben = append(farben, string(farbe))
	}

	supertypen := make([]string, 0, len(eigen.Supertypen))
	for _, super := range eigen.Supertypen {
		supertypen = append(supertypen, string(super))
	}

	typen := make([]string, 0, len(eigen.Typen))
	for _, typ := range eigen.Typen {
		typen = append(typen, string(typ))
	}

	untertypen := eigen.Untertypen
	if untertypen == nil {
		untertypen = []string{}
	}

	unbekannt := eigen.UnbekannteTypen
	if unbekannt == nil {
		unbekannt = []string{}
	}

	return &EigenschaftenDTO{
		Name:   eigen.Name,
		Farben: farben, Manawert: eigen.Manawert,
		Supertypen: supertypen, Typen: typen, Untertypen: untertypen,
		Staerke:          kartenwertAus(eigen.Staerke),
		Widerstandskraft: kartenwertAus(eigen.Widerstandskraft),
		Loyalitaet:       kartenwertAus(eigen.Loyalitaet),
		Verteidigung:     kartenwertAus(eigen.Verteidigung),
		UnbekannteTypen:  unbekannt,
	}
}

// kartenwertAus bildet einen Kartenwert ab.
//
// **Nicht `wertAus`**: den Namen trägt in diesem Paket schon eine
// Testvorrichtung, die eine Notiz in eine Zeichenkette gibt. Zwei Bedeutungen
// für einen Namen sind eine Falle, und die Umbenennung kostet hier nichts.
func kartenwertAus(wert domain.Kartenwert) KartenwertDTO {
	return KartenwertDTO{Steht: wert.Steht, Stern: wert.Stern, Zahl: wert.Zahl}
}

// Farbtext gibt die Farben als Buchstabenfolge — `WU`, leer heißt farblos.
//
// # Warum es diese drei Kurzformen gibt
//
// **Weil MCP, REST und die Kommandozeile Text sind.** Dort ist `2/3` die
// Auskunft, die jemand liest; ein Verbund aus `steht`, `stern` und `zahl`
// wäre dieselbe Aussage in einer Form, die ein Mensch erst zusammensetzen
// muss. Dass sie **hier** stehen und nicht an jeder Fläche, ist §2: eine
// Herleitung, nicht drei.
//
// **Der Zeiger darf `nil` sein** — dann ist jede der drei leer, und die
// Flächen lassen das Feld weg.
func (e *EigenschaftenDTO) Farbtext() string {
	if e == nil {
		return ""
	}

	return strings.Join(e.Farben, "")
}

// Typtext gibt die Typzeile, wie sie auf der Karte steht — CR 205.1.
//
// **Mit dem Gedankenstrich der CR** (U+2014), und ohne ihn, wenn es keine
// Untertypen gibt.
func (e *EigenschaftenDTO) Typtext() string {
	if e == nil {
		return ""
	}

	oben := append(append([]string{}, e.Supertypen...), e.Typen...)

	zeile := strings.Join(oben, " ")
	if len(e.Untertypen) == 0 {
		return zeile
	}

	return strings.TrimSpace(zeile + " — " + strings.Join(e.Untertypen, " "))
}

// Wertetext gibt Stärke und Widerstandskraft als `2/3` — leer, wo keine
// stehen (CR 208.3).
//
// **Der Stern bleibt ein Stern** (CR 208.2): was er bedeutet, sagt eine
// Fähigkeit, und eine Zahl an seiner Stelle wäre erfunden.
func (e *EigenschaftenDTO) Wertetext() string {
	if e == nil || !e.Staerke.Steht && !e.Widerstandskraft.Steht {
		return ""
	}

	return kartenwerttext(e.Staerke) + "/" + kartenwerttext(e.Widerstandskraft)
}

// kartenwerttext gibt einen Wert als Text.
//
// **Nicht `zahltext`**: den Namen trägt in diesem Paket schon die Umrechnung
// einer Kennung (`deck.go`).
func kartenwerttext(wert KartenwertDTO) string {
	if wert.Stern {
		return "*"
	}

	return strconv.Itoa(wert.Zahl)
}
