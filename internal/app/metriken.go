package app

// **Begrenzte Metriken** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Was gezählt wird
//
// Jede **geprüfte Wirkung**: der Riegel aus O1 fragt vor jedem Datei- und
// Netzweg, ob dieser Aufruf darf, und der Ausgang wird gezählt — nach Fläche,
// Wirkung und Ausgang. Alle drei sind Aufzählungen; die Zahl der Reihen ist
// damit **vierzig**, gleich wie viele Karten oder Profile es gibt. Warum das
// eine Eigenschaft des Typs und keine Absicht des Aufrufers ist, steht bei
// `service.Messpunkt`.
//
// # Was **nicht** gezählt wird, und warum das dazugehört
//
// Ein Lesen des Bestands geht an keinem Riegel vorbei (`service.Pruefer` sagt
// warum) und steht deshalb in keiner Reihe. Diese Metrik beantwortet „was hat
// dieser Prozess **gewirkt**", nicht „wie viele Rufe kamen an".
//
// **Und ein Ruf, den der Zwischenspeicher beantwortet, zählt nicht** — das
// ist an der Verdrahtung abgelesen und nicht erschlossen: der Riegel hängt
// am `netz.Klient` (`scryfallverdrahtung.go`, `MitErlaubnis`), der
// Zwischenspeicher liegt darüber im `scryfall.Klient`. Wer zweimal dieselbe
// Karte sucht, sendet einmal — und die Zahl sagt das. Gemessen am
// 2026-09-10: dieselbe Suche zweimal ergab **eine** Zeile mit `anzahl: 1`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// MessreiheDTO ist eine Reihe für eine Fläche.
type MessreiheDTO struct {
	Flaeche string `json:"flaeche"`
	Wirkung string `json:"wirkung"`
	Ausgang string `json:"ausgang"`
	Anzahl  int64  `json:"anzahl"`
}

// MetrikenAntwort trägt den Stand dieses Prozesses.
type MetrikenAntwort struct {
	Reihen []MessreiheDTO `json:"reihen"`

	// Grenze ist die Zahl der Reihen, die es überhaupt geben kann.
	//
	// **Sie steht in der Antwort und nicht nur in der Doku**: wer die Reihen
	// liest, soll ohne Nachschlagen sehen, dass sie nicht wachsen — das ist
	// die Zusage des Werkstücks, und eine Zusage gehört dorthin, wo sie
	// geprüft werden kann.
	Grenze int `json:"grenze"`

	// Laufzeiten ist die Verteilung je Fläche und Art (OPS-05), in den
	// Eimern aus EimerMs und einem offenen darüber. **Die Gestalt aus
	// `service`**: sie trägt ihre JSON-Marken schon und nichts, was die
	// Oberfläche nicht sehen darf.
	Laufzeiten []service.Laufzeitreihe `json:"laufzeiten"`
	EimerMs    []int64                 `json:"eimerMs"`

	// Spur ist die Ablaufspur — leer, solange sie aus ist.
	Spur service.Spurauskunft `json:"spur"`

	// Speicher sind die Zahlen der Zwischenspeicher fremder Dienste.
	Speicher SpeicherDTO `json:"speicher"`

	// Datenbank fehlt, wo dieser Prozess keinen Bestand hält.
	Datenbank *DatenbankDTO `json:"datenbank"`
	Prozess   ProzessDTO    `json:"prozess"`

	// Seit und Bis sind der Erhebungszeitraum, AufwandMikros das, was das
	// Erheben selbst gekostet hat, Ungemessen das, was fehlt (AK-04).
	Seit          string   `json:"seit"`
	Bis           string   `json:"bis"`
	AufwandMikros int64    `json:"aufwandMikros"`
	Ungemessen    []string `json:"ungemessen"`

	Fehler *Fehler `json:"fehler"`
}

// Metriken gibt den Zählstand dieses Prozesses.
//
// **Ohne Naht und ohne Bestand**: gezählt wird im Arbeitsspeicher dieses
// Prozesses (`service.Messstand`). Ein Zähler im Bestand kostete einen
// Schreibvorgang je Prüfung und wäre eine Zahl über **alle** Prozesse — eine
// andere Frage als die, die hier gestellt wird.
//
// **`imProfil` steht in der ersten Zeile, obwohl dieser Weg kein Profil
// braucht** — der Wächter über die Bindungen verlangt sie von jeder, und sein
// Kopf sagt warum: eine Aufzählung der Ausnahmen wäre eine zweite Liste, die
// veralten kann. Was sie hier zusätzlich tut: sie würfelt die
// Vorgangskennung, falls Wails den Ruf ohne eine hereinreicht.
func (a *App) Metriken(ctx context.Context) MetrikenAntwort {
	ctx = a.imProfil(ctx)

	return metrikenAus(ctx, a.db)
}

// metrikenAus stellt die Antwort zusammen — mit den Zahlen der Verbindung,
// wenn es eine gibt.
func metrikenAus(ctx context.Context, db *sql.DB) MetrikenAntwort {
	// **Ein abgebrochener Ruf bekommt keine Zahlen.** Das Lesen selbst dauert
	// nichts; was hier zählt, ist die Antwort an ein Fenster, das schon
	// weitergeblättert hat — es soll den Abbruch sehen und nicht einen Stand,
	// den niemand mehr wollte (§7, Regel 2).
	if fehler := ctx.Err(); fehler != nil {
		return MetrikenAntwort{
			Reihen: []MessreiheDTO{}, Grenze: service.Messreihengrenze,
			Laufzeiten: []service.Laufzeitreihe{}, EimerMs: service.Laufzeiteimer,
			Spur: service.Spurauskunft{Punkte: []service.Spurpunkt{}}, Ungemessen: []string{},
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	stand := service.Messstand()

	antwort := MetrikenAntwort{
		Reihen: make([]MessreiheDTO, 0, len(stand)),
		Grenze: service.Messreihengrenze,
	}

	for _, reihe := range stand {
		antwort.Reihen = append(antwort.Reihen, MessreiheDTO{
			Flaeche: string(reihe.Flaeche),
			Wirkung: string(reihe.Wirkung),
			Ausgang: string(reihe.Ausgang),
			Anzahl:  reihe.Anzahl,
		})
	}

	betriebAnhaengen(&antwort, db)

	return antwort
}

// Metriken ist die Fensterbindung.
func (b *Bindungen) Metriken(ctx context.Context) MetrikenAntwort {
	return b.app(ctx).Metriken(b.ruf(ctx))
}
