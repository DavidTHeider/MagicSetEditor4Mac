package service

// **Was aus einem Suchergebnis eine Recherche macht** — Werkstück A16 von
// Phase 5.3
// (#1515).
//
// # Der Schnitt folgt der Frage
//
// **Dort steht, wonach gesucht wird; hier, was daraus wird** — bündeln,
// ausschließen, schneiden, und vor allem: **warum nichts kam**.
//
// # „Nichts gefunden" heißt sechs Dinge
//
// **Und das Issue verlangt, sie zu unterscheiden**: „Leere Treffer
// unterscheiden von Index fehlt, Modell fehlt, unzureichender Quelle,
// gesperrtem Zugriff und Budgetgrenze." Fünf davon sagen jemandem, was zu
// tun ist; eine sagt, dass es die Antwort nicht gibt.

import (
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ohneIndex ist die Antwort, wenn gar kein Index dasteht.
func ohneIndex(stand Indexstand) domain.Rechercheergebnis {
	return domain.Rechercheergebnis{
		Treffer: []domain.Recherchetreffer{},
		Modus:   string(ModusOhneIndex),
		Leere:   domain.LeereOhneIndex,
		Grund:   stand.Grund,
	}
}

// pruefeKorpusstand weist einen verlangten Stand ab, den es nicht gibt.
//
// **Das Issue erlaubt einen „optionalen Corpusstand"** in der Anfrage. Wer
// ihn nennt, meint einen bestimmten — und einen anderen zu liefern wäre die
// Mischung, die A15 ausschließt.
func pruefeKorpusstand(anfrage domain.Rechercheanfrage, stand Indexstand) error {
	if anfrage.Korpusstand == "" {
		return nil
	}

	for _, korpus := range stand.Manifest.Korpusstaende {
		if korpus.Revision == anfrage.Korpusstand {
			return nil
		}
	}

	return domain.MitGrund(domain.ErrConflict, nil,
		"der Index steht auf einem anderen Stand als %q — eine Antwort "+
			"daraus gehörte zu einem Korpus, der nicht gemeint war",
		anfrage.Korpusstand)
}

// sprachenAus macht aus einer Sprache die Liste, die der Filter erwartet.
func sprachenAus(sprache string) []string {
	if strings.TrimSpace(sprache) == "" {
		return nil
	}

	return []string{sprache}
}

// antwortAus macht aus dem Suchergebnis die Recherche.
func (d *Recherchedienst) antwortAus(
	ergebnis Suchergebnis, anfrage domain.Rechercheanfrage,
	stand Indexstand, ausgeschlossen []string, stelle int, abdruck string,
) domain.Rechercheergebnis {
	treffer := domain.GebuendeltNach(
		domain.OhneAusgangskarte(alsRecherchetreffer(ergebnis.Treffer),
			ausgeschlossen))

	antwort := domain.Rechercheergebnis{
		Modus:             string(ergebnis.Modus),
		Grund:             ergebnis.Grund,
		Modell:            stand.Manifest.Modell,
		Modellrevision:    stand.Manifest.Modellrevision,
		Suchkonfiguration: stand.Manifest.Suchkonfiguration,

		// **Beide, und keiner ersetzt den anderen**: der eine bindet die
		// Liste, der andere den Raum.
		Indexfingerabdruck: indexabdruck(stand),
		Raumabdruck:        stand.Manifest.Raum().Abdruck(),
		Ausgangsstand:      anfrage.Ausgangsstand,

		Korpusstaende:    stand.Manifest.Korpusstaende,
		Einschraenkungen: einschraenkungen(anfrage, stand, ergebnis),
	}

	antwort.Treffer, antwort.Fortsetzung = d.seiteAus(
		treffer, anfrage.Grenze, stelle, abdruck)
	antwort.Leere, antwort.Grund = leereAus(antwort, ergebnis, stand)

	return antwort
}

// seiteAus schneidet die Seite heraus und baut den Zeiger für die nächste.
func (d *Recherchedienst) seiteAus(
	treffer []domain.Recherchetreffer, grenze, stelle int, abdruck string,
) (seite []domain.Recherchetreffer, fortsetzung string) {
	if stelle >= len(treffer) {
		return []domain.Recherchetreffer{}, ""
	}

	bis := min(stelle+grenze, len(treffer))
	seite = slices.Clone(treffer[stelle:bis])

	// **Die Ränge zählen über die ganze Liste** und nicht je Seite: ein
	// Treffer auf Platz elf heißt elf, auch wenn er der erste der zweiten
	// Seite ist.
	for stellung := range seite {
		seite[stellung].Rang = stelle + stellung + 1
	}

	if bis >= len(treffer) {
		return seite, ""
	}

	// **Der Zeiger entsteht nur, wenn es etwas fortzusetzen gibt.** Einer,
	// der auf das Ende zeigt, führte zu einer leeren Seite und sähe aus wie
	// ein Fehler.
	return seite, domain.Recherchezeiger{
		Stelle: bis, Abdruck: abdruck, Erstellt: d.jetzt(),
	}.AlsZeichenkette()
}

// leereAus sagt, warum nichts kam.
func leereAus(
	antwort domain.Rechercheergebnis, ergebnis Suchergebnis, stand Indexstand,
) (leere domain.Rechercheleere, grund string) {
	if len(antwort.Treffer) > 0 {
		return domain.LeereKeine, antwort.Grund
	}

	if !stand.Suchbar {
		return domain.LeereOhneModell, stand.Grund
	}

	if ergebnis.Modus == ModusLexikalisch {
		return domain.LeereOhneModell, ergebnis.Grund
	}

	return domain.LeereNichtsGefunden,
		"Zu dieser Frage steht nichts im Index."
}

// einschraenkungen nennt, was nicht getan wurde.
//
// **Das Issue verlangt „begründete Einschränkungen"** — eine Sprache, die im
// Index gar nicht vorkommt, oder ein Korpus, der nicht gebaut wurde, sind
// keine Fehler und dürfen trotzdem nicht unsichtbar bleiben.
func einschraenkungen(
	anfrage domain.Rechercheanfrage, stand Indexstand, ergebnis Suchergebnis,
) []string {
	gesammelt := make([]string, 0)

	gefuehrt := map[domain.Retrievalquelle]bool{}
	for _, korpus := range stand.Manifest.Korpusstaende {
		gefuehrt[korpus.Quelle] = true
	}

	// **Nur die Quellen, in denen wirklich gesucht wurde.** Eine private
	// ohne Freigabe fehlt nicht im Index — sie wurde nicht gefragt, und das
	// ist eine andere Auskunft (A13.4, #1452).
	for _, quelle := range anfrage.Suchquellen() {
		if !gefuehrt[quelle] {
			gesammelt = append(gesammelt, fmt.Sprintf(
				"Der Korpus %q steht nicht im Index; in ihm wurde nicht gesucht.",
				quelle))
		}
	}

	// **Und was die Freigabe ausgelassen hat, steht als eigener Satz da.**
	//
	// **Nicht als Fehler**: ohne Freigabe zu suchen ist richtig, und die
	// Antwort ist nur enger, als sie sein könnte. Wer es nicht sagte, ließe
	// „nichts gefunden" wie eine Aussage über den Bestand aussehen (§8).
	for _, korpus := range anfrage.AusgelasseneQuellen() {
		gesammelt = append(gesammelt, fmt.Sprintf(
			"Der Korpus %q ist nicht freigegeben; in ihm wurde nicht gesucht.",
			korpus))
	}

	if ergebnis.Modus == ModusLexikalisch {
		gesammelt = append(gesammelt,
			"Gesucht wurde nur im Volltext. "+ergebnis.Grund)
	}

	return gesammelt
}

// alsRecherchetreffer übersetzt die Treffer des Index.
func alsRecherchetreffer(treffer []domain.Retrievaltreffer) []domain.Recherchetreffer {
	umgesetzt := make([]domain.Recherchetreffer, 0, len(treffer))

	for _, einzelner := range treffer {
		dokument := einzelner.Dokument

		umgesetzt = append(umgesetzt, domain.Recherchetreffer{
			Rang: einzelner.Rang, Kennung: dokument.Kennung,
			Quelle:         dokument.Quelle,
			Kennungssystem: dokument.Kennungssystem,
			Fachkennung:    dokument.Fachkennung,
			Semantik:       dokument.Semantik,
			Titel:          dokument.Titel,
			Ausschnitt:     einzelner.Ausschnitt,
			Text:           dokument.Text,
			Sprache:        dokument.Sprache,
			Regelnummer:    dokument.Regelnummer,
			Urheber:        dokument.Urheber,
			Abschnitt:      dokument.Abschnitt,
			Von:            dokument.Von,
			Bis:            dokument.Bis,
			Bezugsadresse:  dokument.Bezugsadresse,
			Quellrevision:  dokument.Quellrevision,
			Quelldatum:     dokument.Quelldatum,
			Abgerufen:      dokument.Abgerufen.UTC().Format(zeitform),
		})
	}

	return umgesetzt
}

// zeitform ist, wie eine Abrufzeit hinausgeht.
const zeitform = "2006-01-02T15:04:05Z07:00"
