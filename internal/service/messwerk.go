// **Was gezählt wird, und warum es endlich bleibt** — Werkstück O9 von Phase
// 5.2 (#1433).
//
// # Die Falle, um die es geht
//
// Das Issue nennt sie beim Namen: „Hohe Zahl verschiedener Objekte erhöht
// Metrikserien nicht unbegrenzt." Wer eine Karten-, Profil- oder Pfadkennung
// als Merkmal einer Messreihe nimmt, bekommt so viele Reihen, wie es Objekte
// gibt — und ein Betriebswerkzeug, das sie liest, fällt daran um.
//
// # Die Antwort ist der Schlüsseltyp
//
// [Messpunkt] hat **drei Felder, und alle drei sind Aufzählungen**: die
// Fläche (vier), die Wirkung (fünf), der Ausgang (zwei). Damit gibt es
// höchstens **vierzig** Reihen, gleich wie viele Karten jemand anlegt. Das ist
// keine Absicht des Aufrufers, sondern eine Eigenschaft des Typs — und ein
// Wächter hält, dass kein viertes Feld dazukommt
// (`internal/werkzeuge/messdimensionen_test.go`).

package service

import (
	"context"
	"sort"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Messpunkt ist der Schlüssel einer Messreihe.
type Messpunkt struct {
	Flaeche domain.Dienstmodus
	Wirkung domain.Wirkung
	Ausgang domain.Ausgang
}

// Messreihe ist ein Messpunkt mit seinem Stand.
type Messreihe struct {
	Flaeche domain.Dienstmodus `json:"flaeche"`
	Wirkung domain.Wirkung     `json:"wirkung"`
	Ausgang domain.Ausgang     `json:"ausgang"`
	Anzahl  int64              `json:"anzahl"`
}

// Messreihengrenze ist die Zahl der Reihen, die es überhaupt geben kann —
// vier Flächen mal **sechs** Wirkungen mal zwei Ausgänge.
//
// **Gerechnet und nicht abgeschrieben** (N7, #1568): ein Einschub in
// [domain.Wirkungen] überholte eine abgeschriebene Zahl still, und die Grenze
// wäre zu niedrig, ohne dass es jemand sieht. Sie zählt die Aufzählungen
// selbst.
var Messreihengrenze = len(domain.Dienstmodi()) * len(domain.Wirkungen()) *
	len(domain.Ausgaenge())

// messwerk zählt die geprüften Wirkungen dieses Prozesses.
//
// # Warum ein Paketwert und kein durchgereichter Zähler
//
// **Dasselbe Argument wie bei `slog.Default()`** (`internal/app/fehler.go`):
// [Autorisierung.Pruefe] wird von achtzehn Wurzeln aus gebaut — jede Fläche,
// jede Naht ihre eigene —, und ein Zähler, den jede von ihnen durchreichen
// müsste, wäre bei der ersten vergessenen Wurzel stumm. Gezählt wird, was
// dieser **Prozess** tut; das ist die Einheit, über die ein Betriebswerkzeug
// ohnehin fragt.
var messwerk = struct {
	sperre sync.Mutex
	stand  map[Messpunkt]int64
}{stand: map[Messpunkt]int64{}}

// Zaehle nimmt einen Messpunkt auf.
//
// **Ein unbekannter Wert wird nicht gezählt.** Er kann nicht entstehen — die
// drei Typen sind geschlossene Aufzählungen, und jede Rufstelle setzt
// Konstanten —, und wenn er es doch täte, wäre die Grenze von vierzig Reihen
// keine Zusage mehr. Was hier passiert, ist deshalb kein Sicherheitsnetz für
// einen erwarteten Fall, sondern die Stelle, an der die Zusage eingelöst wird.
func Zaehle(punkt Messpunkt) {
	if !bekannt(punkt) {
		return
	}

	messwerk.sperre.Lock()
	defer messwerk.sperre.Unlock()

	messwerk.stand[punkt]++
}

// bekannt sagt, ob alle drei Werte aus ihrer Aufzählung stammen.
func bekannt(punkt Messpunkt) bool {
	return ausDerAufzaehlung(domain.Dienstmodi(), punkt.Flaeche) &&
		ausDerAufzaehlung(domain.Wirkungen(), punkt.Wirkung) &&
		ausDerAufzaehlung(domain.Ausgaenge(), punkt.Ausgang)
}

func ausDerAufzaehlung[W comparable](vorrat []W, wert W) bool {
	for _, kandidat := range vorrat {
		if kandidat == wert {
			return true
		}
	}

	return false
}

// Messstand gibt die Reihen mit einem Stand über null — **in fester
// Reihenfolge**, damit zwei Abrufe hintereinander dieselbe Ausgabe geben und
// ein Vergleich etwas bedeutet.
func Messstand() []Messreihe {
	messwerk.sperre.Lock()
	defer messwerk.sperre.Unlock()

	reihen := make([]Messreihe, 0, len(messwerk.stand))
	for punkt, anzahl := range messwerk.stand {
		reihen = append(reihen, Messreihe{
			Flaeche: punkt.Flaeche, Wirkung: punkt.Wirkung,
			Ausgang: punkt.Ausgang, Anzahl: anzahl,
		})
	}

	sort.Slice(reihen, func(links, rechts int) bool {
		if reihen[links].Flaeche != reihen[rechts].Flaeche {
			return reihen[links].Flaeche < reihen[rechts].Flaeche
		}

		if reihen[links].Wirkung != reihen[rechts].Wirkung {
			return reihen[links].Wirkung < reihen[rechts].Wirkung
		}

		return reihen[links].Ausgang < reihen[rechts].Ausgang
	})

	return reihen
}

// MesswerkZuruecksetzen leert den Stand — **für Tests**, damit einer nicht die
// Zahlen des vorigen sieht.
func MesswerkZuruecksetzen() {
	messwerk.sperre.Lock()
	defer messwerk.sperre.Unlock()

	messwerk.stand = map[Messpunkt]int64{}
}

// gezaehlt trägt den Ausgang einer Prüfung ein und gibt ihn unverändert
// zurück.
//
// **An dieser einen Stelle**, weil hier jede Wirkung vorbeikommt, die über
// den Bestand hinausgeht — und nur solche werden geprüft (siehe [Pruefer]).
// Ein Zähler an jeder Bindung zählte Bindungen; gezählt werden soll, was
// **wirkt**.
func gezaehlt(ctx context.Context, wirkung domain.Wirkung, fehler error) error {
	ausgang := domain.AusgangGewaehrt
	if fehler != nil {
		ausgang = domain.AusgangAbgewiesen
	}

	Zaehle(Messpunkt{
		Flaeche: DienstmodusAus(ctx), Wirkung: wirkung, Ausgang: ausgang,
	})

	return fehler
}
