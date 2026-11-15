package austausch

// **Einen Umschlag lesen** — und zwar in dieser Reihenfolge: Behälter,
// Manifest, Unterschrift, Quarantäne, **dann erst** Inhalt.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Sendung ist ein gelesener Umschlag — **noch nicht übernommen**.
//
// **Der Unterschied ist der ganze Punkt von N4**: was hier steht, ist geprüft
// und liegt in einem eigenen Verzeichnis. Was davon in den Bestand kommt,
// entscheidet ein Mensch (#1457: „benötigt fachliche Übernahmefreigabe").
type Sendung struct {
	Manifest Manifest

	// Verzeichnis ist, wo die geprüften Objekte liegen.
	Verzeichnis string

	// Urheber ist die Kennung, deren Unterschrift getragen hat.
	Urheber peer.Kennung
}

// Lies prüft einen Umschlag und packt ihn in ein eigenes Verzeichnis.
//
// # Die Reihenfolge ist die Zusage
//
//  1. **Der Behälter** — `sicherung` prüft Größen, Zahl der Einträge und
//     Pfade, bevor irgendetwas geschrieben wird.
//  2. **Das Manifest** — was steht drin, und in welcher Fassung.
//  3. **Die Unterschrift** — über die **rohen Bytes** des Manifests, nicht
//     über eine wieder zusammengesetzte Form: zwei Kodierungen derselben
//     Struktur ergeben zwei verschiedene Bytefolgen.
//  4. **Die Quarantäne** — Grenzen, Namen, Abhängigkeiten — und jedes Objekt
//     gegen den Behälter: dieselbe Länge, dieselbe Summe.
//  5. **Erst dann** die Objekte, und `sicherung` hält jedes an Länge und Summe.
//
// **Wer die Reihenfolge umdreht, prüft Bytes, die schon dastehen.**
func Lies(pfad, verzeichnis string) (Sendung, error) {
	// **Zuerst nur Manifest und Unterschrift.** `EntpackenNur` holt genau
	// die beiden Einträge; der Rest bleibt im Archiv, bis die Quarantäne
	// durch ist.
	pruefort := filepath.Join(verzeichnis, ".pruefung")

	//
	// **Jeder an seiner eigenen Grenze** (SEC-03, #2138): hier galten die 16
	// GiB je Teil einer Sicherung, und beide wurden ganz eingelesen, bevor
	// die Unterschrift geprüft war.
	behaelter, fehler := sicherung.EntpackenGedeckelt(
		pfad, pruefort, aeusseresManifest, HoechstensUmschlagmanifest, map[string]int64{
			Manifestname: HoechstensManifest, Unterschriftsname: HoechstensUnterschrift,
		})
	if fehler != nil {
		// **Der Satz wird ersetzt und nicht durchgereicht.** `sicherung`
		// spricht von „der Sicherung" — sie ist sein Gegenstand, und dieses
		// Paket leiht sich nur seinen Auspacker. Wer hier einen Umschlag
		// öffnet und „die Sicherung lässt sich nicht öffnen" liest, sucht am
		// falschen Ort; angestrichen vom Flächenlauf über N4, der an allen
		// drei Flächen denselben falschen Satz sah.
		//
		// **Der Grund reist mit** (§8): der Satz ist für Menschen, der Grund
		// fürs Protokoll.
		return Sendung{}, domain.MitGrund(domain.ErrValidation, fehler,
			"diese Datei ist kein Umschlag — wähle eine andere")
	}

	roh, fehler := os.ReadFile(filepath.Join(pruefort, Manifestname)) //#nosec G304 -- der Ort ist unser eigener
	if fehler != nil {
		return Sendung{}, domain.MitGrund(domain.ErrValidation, fehler,
			"das Manifest des Umschlags lesen")
	}

	unterschrift, fehler := os.ReadFile( //#nosec G304 -- dito
		filepath.Join(pruefort, Unterschriftsname))
	if fehler != nil {
		return Sendung{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Unterschrift des Umschlags lesen")
	}

	var manifest Manifest
	if fehler := json.Unmarshal(roh, &manifest); fehler != nil {
		return Sendung{}, domain.MitGrund(domain.ErrValidation, fehler,
			"das Manifest des Umschlags deuten")
	}

	urheber := peer.Kennung(manifest.Urheber)

	// **Über die rohen Bytes**, und der Grund steht oben.
	if fehler := peer.Pruefe(urheber, roh, unterschrift); fehler != nil {
		return Sendung{}, fmt.Errorf("%w: %w", domain.ErrUmschlagUnterschrift,
			fehler)
	}

	if fehler := pruefeManifest(manifest); fehler != nil {
		return Sendung{}, fehler
	}

	if fehler := pruefeGegenBehaelter(manifest, behaelter); fehler != nil {
		return Sendung{}, fehler
	}

	namen := make([]string, 0, len(manifest.Objekte))
	for _, objekt := range manifest.Objekte {
		namen = append(namen, objekt.Name)
	}

	inhaltsort := filepath.Join(verzeichnis, "inhalt")

	// **Ausgepackt wird nach dem Behälter**, und der hält Länge und Summe
	// jedes Teils fest — seit [pruefeGegenBehaelter] dieselben wie die
	// unterschriebenen. Eine zweite Prüfung danach fände nichts mehr.
	if _, fehler := sicherung.EntpackenNurAls(
		pfad, inhaltsort, aeusseresManifest, namen...); fehler != nil {
		return Sendung{}, fehler
	}

	return Sendung{
		Manifest: manifest, Verzeichnis: inhaltsort, Urheber: urheber,
	}, nil
}

// pruefeGegenBehaelter hält das unterschriebene Manifest gegen das des
// Behälters — **bevor** ausgepackt wird.
//
// **Das äußere Manifest sagt, was der Behälter trägt; das innere sagt, was
// der Absender unterschrieben hat.** Stimmen beide nicht überein, ist der
// Umschlag nachträglich geändert worden — und genau das soll auffallen. Bis
// QA-06b fiel es erst nach dem Auspacken auf, und nur an der Summe.
//
// # Warum die Quarantäne das braucht
//
// **Ihre Grenzen gelten der Angabe `groesse`, und gegen die Bytes hielt sie
// niemand.** `sicherung` packt nach dem **äußeren** Manifest aus und hält
// dort Länge und Summe jedes Teils fest; stimmen beide mit dem inneren
// überein, gilt die Angabe, an der die Quarantäne gemessen hat, auch für
// das, was auf die Platte kommt. Bis QA-06b
// (#1967)
// trug ein unterschriebener Umschlag mit `groesse: 1` beliebig viel herein.
func pruefeGegenBehaelter(manifest Manifest, behaelter sicherung.Beschreibung) error {
	teile := make(map[string]sicherung.Teilangabe, len(behaelter.Teile))
	for _, teil := range behaelter.Teile {
		teile[teil.Name] = teil
	}

	for _, objekt := range manifest.Objekte {
		if teil, da := teile[objekt.Name]; !da || teil.Groesse != objekt.Groesse || teil.Summe != objekt.Summe {
			return fmt.Errorf("%w: %q ist im Umschlag nicht das, was das Manifest sagt",
				domain.ErrUmschlagInhalt, objekt.Name)
		}
	}

	return nil
}
