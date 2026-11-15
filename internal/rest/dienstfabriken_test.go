// **Die beiden Dienstfabriken der Tests.**
//
// Aus `zaehlerdoppel_test.go` herausgelöst am 2026-09-02, als die Datei über
// 400 Zeilen wuchs (§2) — der Anlass war die Paketprüfung
// (#1324).
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **was ein Doppel antwortet**; hier, **wie ein Dienst daraus
// entsteht**. Die beiden Fabriken unten sind die einzige Stelle, an der alle
// Nähte eines Dienstes aufgezählt werden — eine neue Naht fällt genau hier
// auf, und das soll sie.

package rest_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func dienstMit(t *testing.T, gezaehlt *zaehlender) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	uhr := func() time.Time { return jetzt }

	return mitHaushalt(t, rest.NeuerDienst(probemarke, uhr,
		rest.Naehte{
			Leser:            gezaehlt,
			Schreiber:        gezaehlt,
			Ausfuhr:          gezaehlt,
			Abgleich:         gezaehlt,
			Bilder:           gezaehlt,
			Auskunft:         gezaehlt,
			Suche:            gezaehlt,
			Katalog:          gezaehlt,
			Peer:             gezaehlt,
			Kontakte:         gezaehlt,
			Sichtfreigaben:   gezaehlt,
			Tausch:           gezaehlt,
			Post:             gezaehlt,
			Verbuchung:       gezaehlt,
			Leitung:          gezaehlt,
			Links:            gezaehlt,
			Konto:            gezaehlt,
			Nachbarn:         gezaehlt,
			Grenzen:          gezaehlt,
			Schnittstellen:   gezaehlt,
			Schalter:         gezaehlt,
			Gesundheit:       gezaehlt,
			Metriken:         gezaehlt,
			Meta:             gezaehlt,
			Vertraeglichkeit: gezaehlt,
			Ruecksprung:      gezaehlt,
			Pflege:           gezaehlt,
			Ruecksetzung:     gezaehlt,
			Index:            gezaehlt,
			Recherche:        gezaehlt,
			Kosten:           gezaehlt,
			Verweise:         gezaehlt,
			Marken:           gezaehlt,
			Ablage:           gezaehlt,
			Ersetzung:        gezaehlt,
			Liste:            gezaehlt,
			Sammeldaten:      gezaehlt,
			Anlage:           gezaehlt,
			Saetze:           gezaehlt,
			Konsole:          gezaehlt,
			Werkstatt:        gezaehlt,
			Sammlung:         gezaehlt,
			Listeneinfuhr:    gezaehlt,
			Deckeinfuhr:      gezaehlt,
			Sammlungsausfuhr: gezaehlt,
			Sammlungssichten: gezaehlt,
			Kartenskripte:    gezaehlt,
			Auszaehlung:      gezaehlt,
			Parameterhilfen:  gezaehlt,
			Autoersetzungen:  gezaehlt,
			Decks:            gezaehlt,
			Turniere:         gezaehlt,
			Symbolbibliothek: gezaehlt,
			Sprachpakete:     gezaehlt,
			Symbol:           gezaehlt,
			Masse:            gezaehlt,
			Loescher:         gezaehlt,
			Sicherung:        gezaehlt,
			Umschlag:         gezaehlt,
			Partien:          gezaehlt,
			Partienverlauf:   gezaehlt,
			Netzweg:          gezaehlt,
			Regelfreigabe:    gezaehlt,
			Preise:           gezaehlt,
			Wortwahl:         gezaehlt,
			Woerter:          gezaehlt,
			Pack:             gezaehlt,
			Angaben:          gezaehlt,
			Probleme:         gezaehlt,
			Stile:            gezaehlt,
			Zusatz:           gezaehlt,
			Pakete:           gezaehlt,
			Nutzer:           gezaehlt,
			Token:            gezaehlt,
			Vorgaben:         gezaehlt,
			Freigaben:        gezaehlt,
			Platz:            gezaehlt,
			Deckuebernahme:   gezaehlt,
			Kartenbilder:     gezaehlt,
			Vorlagen:         gezaehlt,
			Probelauf:        gezaehlt,
			Zahlen:           gezaehlt,
			Verlauf:          gezaehlt,
			Einstellungen:    gezaehlt,
			Austausch:        austauschzaehler{gezaehlt},
			Fassungen:        gezaehlt,
			Anleger:          gezaehlt,
			Auftraege:        gezaehlt,
			Setvorschau:      gezaehlt,
			Dateiablage:      gezaehlt,
			Dateien:          gezaehlt,
			Objektmarken:     gezaehlt,
		}), domain.Vorgabegrenzen(), uhr)
}

// neuerDienst baut einen Dienst über willigen Doubles.
//
// **Willig, nicht leer.** Die vier Klassen prüfen den Weg **bis** zum
// Anwendungsfall; was dahinter liegt, hat seine eigenen Tests. Ein Double, das
// scheiterte, machte aus „gültig" einen Fehlerfall und aus der Klasse eine
// Behauptung.
func neuerDienst(t *testing.T) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	return neuerDienstUm(t, func() time.Time { return jetzt })
}

// neuerDienstUm ist derselbe Dienst mit einer Uhr, die der Test stellt (#1429).
func neuerDienstUm(t *testing.T, jetzt func() time.Time) *rest.Dienst {
	t.Helper()

	return mitHaushalt(t, rest.NeuerDienst(probemarke, jetzt,
		rest.Naehte{
			Leser:            williger{},
			Schreiber:        williger{},
			Ausfuhr:          williger{},
			Abgleich:         williger{},
			Bilder:           williger{},
			Auskunft:         williger{},
			Suche:            williger{},
			Katalog:          williger{},
			Peer:             williger{},
			Kontakte:         williger{},
			Sichtfreigaben:   williger{},
			Tausch:           williger{},
			Post:             williger{},
			Verbuchung:       williger{},
			Leitung:          williger{},
			Links:            williger{},
			Konto:            williger{},
			Nachbarn:         williger{},
			Grenzen:          williger{},
			Schnittstellen:   williger{},
			Schalter:         williger{},
			Gesundheit:       williger{},
			Metriken:         williger{},
			Meta:             williger{},
			Vertraeglichkeit: williger{},
			Ruecksprung:      williger{},
			Pflege:           williger{},
			Ruecksetzung:     williger{},
			Index:            williger{},
			Recherche:        williger{},
			Kosten:           williger{},
			Verweise:         williger{},
			Marken:           williger{},
			Ablage:           williger{},
			Ersetzung:        williger{},
			Liste:            williger{},
			Sammeldaten:      williger{},
			Anlage:           williger{},
			Saetze:           williger{},
			Konsole:          williger{},
			Werkstatt:        williger{},
			Sammlung:         williger{},
			Listeneinfuhr:    williger{},
			Deckeinfuhr:      williger{},
			Sammlungsausfuhr: williger{},
			Sammlungssichten: williger{},
			Kartenskripte:    williger{},
			Auszaehlung:      williger{},
			Parameterhilfen:  williger{},
			Autoersetzungen:  williger{},
			Decks:            williger{},
			Turniere:         williger{},
			Symbolbibliothek: williger{},
			Sprachpakete:     williger{},
			Symbol:           williger{},
			Masse:            williger{},
			Loescher:         williger{},
			Sicherung:        williger{},
			Umschlag:         williger{},
			Partien:          williger{},
			Partienverlauf:   williger{},
			Netzweg:          williger{},
			Regelfreigabe:    williger{},
			Preise:           williger{},
			Wortwahl:         williger{},
			Woerter:          williger{},
			Pack:             williger{},
			Angaben:          williger{},
			Probleme:         williger{},
			Stile:            williger{},
			Zusatz:           williger{},
			Pakete:           williger{},
			Nutzer:           williger{},
			Token:            williger{},
			Vorgaben:         williger{},
			Freigaben:        williger{},
			Platz:            williger{},
			Deckuebernahme:   williger{},
			Kartenbilder:     williger{},
			Vorlagen:         williger{},
			Probelauf:        williger{},
			Zahlen:           williger{},
			Verlauf:          williger{},
			Einstellungen:    williger{},
			Austausch:        austauschprobe{},
			Fassungen:        williger{},
			Anleger:          williger{},
			Auftraege:        williger{},
			Setvorschau:      williger{},
			Dateiablage:      williger{},
			Dateien:          williger{},
			Objektmarken:     williger{},
		}), domain.Vorgabegrenzen(), jetzt)
}

// **Und die Sicherung des ganzen Bestands** (#1169).
