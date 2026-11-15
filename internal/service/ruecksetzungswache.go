package service

// **Wer darf was zurücksetzen** — Werkstück O4 von Phase 5.2
// (#1428).
//
// # Drei Fragen vor jeder Rücksetzung
//
//  1. **Gehört das Ziel dem, der ruft?** Ein Profil aus seiner Liste, sein
//     eigener Nutzer — sonst nichts. Das ist die Zusage der Abnahme: „Kein
//     Profilverwalter erhält durch umfassenden Reset Rechte über andere
//     Eigentümer."
//  2. **Läuft gerade noch etwas?** Dann wird nicht gelöscht, sondern gewartet.
//  3. **Und für die ganze Anwendung: liegt ein [domain.Nachweis] vor?** Eine
//     Rolle reicht dafür nicht — sie nähme anderen Eigentümern ihre Daten weg.
//
// **Die ersten beiden sind Hindernisse, die dritte ist es nicht.** Ein
// Hindernis sagt „so, wie es steht, geht es nicht"; ein fehlender Nachweis
// sagt „hol ihn dir". Die Vorschau der ganzen Anwendung soll zeigen, was
// geschähe — nicht wie eine Sperre aussehen.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ruecksetzungswache sagt, wer ruft, und löst Nachweise ein.
//
// **Erfüllt von [Autorisierung]** — die Schnittstelle steht hier, weil dieser
// Anwendungsfall sie braucht (§2). Zwei Methoden und nicht eine: der Akteur
// beantwortet „gehört dir das?", die Prüfung „darfst du das, und hast du es
// gewollt?".
type Ruecksetzungswache interface {
	Akteur(ctx context.Context) (domain.Akteur, error)
	PruefeMitNachweis(ctx context.Context, wirkung domain.Wirkung,
		kennung string, bindung Nachweisbindung) error
}

// Auftragsschau nennt, was dieser Prozess gerade tut.
//
// **Je Prozess und nicht je Bestand** ([Auftragsbuch]): das Buch steht im
// Arbeitsspeicher. Was ein zweites Fenster tut, sieht diese Schau nicht — die
// Zusage reicht so weit wie die Messung (§10).
type Auftragsschau func() []domain.Auftrag

// MitAuftraegen gibt der Rücksetzung die Schau auf das, was läuft.
//
// **Wahlweise**: die Flächen ohne Fenster führen kein Buch, und ein Prozess
// ohne Aufträge hat keine, die er stoppen müsste.
func (r *Ruecksetzung) MitAuftraegen(schau Auftragsschau) *Ruecksetzung {
	r.auftraege = schau

	return r
}

// gebundenesZiel füllt das Ziel aus dem Akteur und prüft, ob es ihm gehört.
//
// **Eine leere Kennung heißt „meines"** — und nicht „irgendeines". Wer nichts
// nennt, meint das gezeigte Profil und sich selbst; das ist der häufige Fall
// am Fenster, und er soll nicht daran scheitern, dass jemand eine Zahl
// abtippen muss, die er nicht sieht.
//
// **Sie gibt die Profile des Menschen gleich mit** — geladen werden sie
// ohnehin, und ein zweiter Ruf in `betroffeneProfile` hätte einen Fehlerweg
// gehabt, den kein Aufruf betritt: dort **steht** das Profil dann in der
// Liste, weil es hier schon geprüft wurde (§3).
func (r *Ruecksetzung) gebundenesZiel(
	ctx context.Context, art domain.Ruecksetzungsart, ziel Ruecksetzungsziel,
) (Ruecksetzungsziel, []domain.Profil, []domain.Hindernis, error) {
	akteur, fehler := r.wache.Akteur(ctx)
	if fehler != nil {
		return ziel, nil, nil, fehler
	}

	if ziel.ProfilID == 0 {
		ziel.ProfilID = akteur.ProfilID
	}

	if ziel.NutzerID == 0 {
		ziel.NutzerID = akteur.NutzerID
	}

	if ziel.NutzerID != akteur.NutzerID {
		return ziel, nil, []domain.Hindernis{{Art: domain.RuecksetzungFremdesZiel}}, nil
	}

	if !art.LoeschtDaten() {
		return ziel, nil, nil, nil
	}

	// **Der Neuanfang sieht jedes Profil** — er legt die ganze Datei
	// beiseite und nicht nur, was dem gehört, der ruft (Befund 1 des
	// Zweitblicks über Phase G, #2282). Wer ihn darf, entscheidet der
	// Nachweis, der nur am Fenster ausgestellt wird (0078).
	lesen := func(ctx context.Context) ([]domain.Profil, error) {
		return r.ablage.Profile(ctx, akteur.NutzerID)
	}
	if art == domain.RuecksetzungAnwendung {
		lesen = r.ablage.AlleProfile
	}

	profile, profilfehler := lesen(ctx)
	if profilfehler != nil {
		return ziel, nil, nil, profilfehler
	}

	if art != domain.RuecksetzungProfil {
		return ziel, profile, nil, nil
	}

	// **Das eine, und mit seinem Namen** — genau ihn soll jemand gleich
	// darunter abtippen. In der laufenden Anwendung nachgesehen (2026-09-10).
	for _, profil := range profile {
		if profil.ID == ziel.ProfilID {
			return ziel, []domain.Profil{profil}, nil, nil
		}
	}

	return ziel, nil, []domain.Hindernis{{Art: domain.RuecksetzungFremdesZiel}}, nil
}

// laufende gibt die Hindernisse aus dem, was gerade läuft.
//
// **Nur vor einer Löschung.** Eine Wahl zurückzunehmen nimmt niemandem etwas
// weg; einem laufenden Schreiber den Bestand wegzuziehen schon.
func (r *Ruecksetzung) laufende(art domain.Ruecksetzungsart) []domain.Hindernis {
	if r.auftraege == nil || !art.LoeschtDaten() {
		return nil
	}

	var hindernisse []domain.Hindernis

	for _, auftrag := range r.auftraege() {
		if auftrag.Stand != domain.AuftragLaeuft {
			continue
		}

		hindernisse = append(hindernisse, domain.Hindernis{
			Art: domain.RuecksetzungLaufenderAuftrag, Stelle: auftrag.Was,
		})
	}

	return hindernisse
}

// erlaubt fragt Rolle und Nachweis.
//
// **Der fehlende Nachweis kommt zuerst**, denn er ist der Satz, den ein Mensch
// lesen soll: „hol dir die Freigabe". Die Absage der Rolle danach ist die
// Auskunft für den, der sie hat.
func (r *Ruecksetzung) erlaubt(
	ctx context.Context, art domain.Ruecksetzungsart, kennung string,
) error {
	if art.BrauchtNachweis() && kennung == "" {
		return fmt.Errorf("%w: %s", domain.ErrErlaubnis,
			satzZuHindernis(domain.Hindernis{Art: domain.RuecksetzungNachweisFehlt}))
	}

	return r.wache.PruefeMitNachweis(ctx, domain.WirkungSchreiben, kennung,
		Nachweisbindung{Zweck: domain.RuecksetzungsZweck, Ziel: string(art)})
}

// satzZuHindernis macht aus einer Marke einen Satz für die Oberfläche.
//
// **Hier und nicht in vier Flächen**: ein Hindernis, das eine Rücksetzung
// **anhält**, ist ein Fehler und kein Zustand — und ein Fehler trägt seinen
// Satz (§8). Die Marken der **Vorschau** bleiben Marken; dort baut jede Fläche
// ihren eigenen.
func satzZuHindernis(hindernis domain.Hindernis) string {
	switch hindernis.Art {
	case domain.RuecksetzungOffeneRueckkehr:
		return "Über Mengen in diesem Profil steht nach einer Rückkehr nichts " +
			"fest. Sieh sie nach und quittiere die Rückkehr, bevor du löschst."
	case domain.RuecksetzungLetzterNutzer:
		return "Das ist der letzte Mensch in diesem Bestand — ihn zu löschen " +
			"ließe niemanden übrig."
	case domain.RuecksetzungFremdesZiel:
		return "Das gehört dir nicht. Zurücksetzen lässt sich, was dir gehört."
	case domain.RuecksetzungLaufenderAuftrag:
		return "Hier läuft noch etwas: " + hindernis.Stelle +
			". Warte, bis es fertig ist."
	case domain.RuecksetzungNachweisFehlt:
		return "Die ganze Anwendung zurückzusetzen verlangt eine ausdrückliche " +
			"Freigabe an diesem Rechner."
	case domain.RuecksetzungOhneTausch:
		return "Die ganze Anwendung setzt dieser Dienst nicht zurück: er hält " +
			"den Bestand, solange er läuft. Das geht am Fenster oder an der " +
			"Kommandozeile (mse zuruecksetzen anwendung)."
	}

	return "Diese Rücksetzung geht gerade nicht."
}
