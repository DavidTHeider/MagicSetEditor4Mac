// **Die Übernahme aus einer vorhandenen MSE2-Installation** —
// #56.
//
// **Aus `paketverwalter.go` herausgelöst**, als sie über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort steht, wie **ein** Paket verwaltet wird —
// installieren, entfernen, auffrischen —, hier, wie eine **fremde Sammlung**
// als Ganzes hereinkommt. Das ist ein anderer Gegenstand: sie führt Hunderte,
// sie liegt woanders, und was dabei schiefgeht, ist ein Befund je Paket und
// kein Abbruch.

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Uebernahme ist der Ausgang einer Übernahme aus einer MSE2-Installation.
type Uebernahme struct {
	// Uebernommen sind die Pakete, die dabei entstanden sind.
	Uebernommen []domain.Paket

	// Uebergangen sind die, die schon dalagen.
	//
	// **Übergangen und nicht überschrieben** — derselbe Grund wie beim
	// Installieren: wer eine Vorlage bearbeitet hat, verlöre sie sonst, und
	// nichts sagte es ihm.
	Uebergangen []string

	Befunde []domain.Befund
}

// Uebernehmen holt alle Vorlagenpakete aus einem Verzeichnis.
//
// # Der Hauptweg, und warum
//
// **Es ist der einzige Bezug, bei dem die Rechtslage klar ist**
// (#55): die
// Dateien liegen bereits auf dem Rechner der Nutzenden. Was über das Netz
// kommt, gehört jemand anderem (§11).
//
// # Was er nicht tut
//
// **Er sucht nicht von sich aus.** Welches Verzeichnis gemeint ist, sagt der
// Aufrufer; die üblichen Orte schlägt die Oberfläche vor und liest sie nicht
// ungefragt.
//
// **Er geht eine Ebene tief.** MSE2 legt seine Pakete unmittelbar unter `data/`
// ab; wer tiefer suchte, fände die Module in den Paketen und hielte sie für
// Pakete.
func (v *Paketverwalter) Uebernehmen(ctx context.Context, ort string) (*Uebernahme, error) {
	if err := v.ortPruefen(); err != nil {
		return nil, err
	}

	angaben, err := os.Stat(ort)
	if err != nil || !angaben.IsDir() {
		return nil, domain.NewFieldError("ort", domain.SchluesselValidierung,
			"%q ist kein Verzeichnis mit Vorlagen.", ort)
	}

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		return nil, fmt.Errorf("%q lesen: %w", ort, err)
	}

	ergebnis := &Uebernahme{
		Uebernommen: []domain.Paket{},
		Uebergangen: []string{},
		Befunde:     []domain.Befund{},
	}

	var vermerke []Vermerk

	for _, eintrag := range eintraege {
		quelle := filepath.Join(ort, eintrag.Name())

		gelesen, befund := lies(quelle, eintrag.Name(), nil)

		switch {
		case befund != nil:
			ergebnis.Befunde = append(ergebnis.Befunde, *befund)

		case gelesen == nil || !eintrag.IsDir():
			// Kein Paket, ein Set oder ein Archiv — nichts davon gehört in ein
			// Vorlagenverzeichnis (der Grund steht bei `lies` und `Installieren`).

		default:
			vermerk, uebergangen := v.uebernimm(quelle, eintrag.Name())
			if uebergangen {
				ergebnis.Uebergangen = append(ergebnis.Uebergangen, eintrag.Name())

				continue
			}

			if vermerk == nil {
				ergebnis.Befunde = append(ergebnis.Befunde, domain.Befund{
					Datei:   eintrag.Name(),
					Meldung: "das Paket ließ sich nicht übernehmen",
				})

				continue
			}

			vermerke = append(vermerke, *vermerk)
		}
	}

	// **Ein Abgleich für alle, nicht einer je Paket.** Eine MSE2-Installation
	// führt Hunderte; ein Abgleich je Paket wäre dieselbe Arbeit hundertfach.
	//
	// **Und er hält kein Abbruch mehr auf** — siehe `Installieren`. Die Pakete
	// **liegen** im Verzeichnis, teils Hunderte; ein `ctx`, der hier abbricht,
	// meldete einen Fehler über eine Übernahme, die stattgefunden hat, und
	// ließe sie sämtlich außerhalb des Katalogs zurück.
	ohneAbbruch := context.WithoutCancel(ctx)

	befunde, err := v.Vermerken(ohneAbbruch, vermerke...)
	ergebnis.Befunde = append(ergebnis.Befunde, befunde...)

	if err != nil {
		return ergebnis, err
	}

	ergebnis.Uebernommen = v.uebernommene(ohneAbbruch, vermerke)

	return ergebnis, nil
}

// uebernimm kopiert ein Paket; der zweite Wert sagt, dass es schon dalag.
func (v *Paketverwalter) uebernimm(quelle, name string) (*Vermerk, bool) {
	ziel := filepath.Join(v.verwaltet, name)
	if _, err := os.Stat(ziel); err == nil {
		return nil, true
	}

	if err := kopierePaket(quelle, ziel); err != nil {
		_ = os.RemoveAll(ziel) //nolint:errcheck // der Befund steht beim Aufrufer

		return nil, false
	}

	return &Vermerk{Ort: ziel, Herkunft: quelle}, false
}

// uebernommene sucht die frisch übernommenen Pakete im Katalog.
func (v *Paketverwalter) uebernommene(ctx context.Context, vermerke []Vermerk) []domain.Paket {
	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		// **Kein zweiter Fehlerweg für dieselbe Auskunft.** Der Abgleich ist
		// geglückt; was danach beim Lesen schiefgeht, macht die Übernahme nicht
		// ungeschehen (§7, Regel 3). Die Liste bleibt dann leer, die Zahl der
		// Übergangenen und die Befunde stehen trotzdem.
		return []domain.Paket{}
	}

	orte := make(map[string]bool, len(vermerke))
	for _, vermerk := range vermerke {
		orte[vermerk.Ort] = true
	}

	uebernommen := make([]domain.Paket, 0, len(vermerke))

	for _, paket := range pakete {
		if orte[paket.Pfad] {
			uebernommen = append(uebernommen, paket)
		}
	}

	return uebernommen
}
