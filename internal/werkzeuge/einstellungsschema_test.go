package werkzeuge_test

// **Das Schema der Einstellungen steht in `docs/EINSTELLUNGEN.md`** —
// OPS-01.AK-01 (#1718).
//
// # Was er hält
//
// **Beide Tafeln Zeile für Zeile**: die der Profil- und Geräteeinstellungen
// und die des Menschen, je in Anzeigereihenfolge, mit Bereich, Ebene, Art,
// Werten, Vorgabe und Wirkung — erzeugt aus `internal/domain`. Eine neue
// Einstellung ohne Zeile, eine Zeile ohne Einstellung und eine geänderte
// Vorgabe machen ihn rot, und seine Meldung nennt die Zeile, wie sie lauten
// soll.
//
// # Was er nicht hält
//
// **Die Prosa über den Tafeln** — die Vorrangfolge, die Ebenen, die
// Umgebung. Sie steht im Code an den Stellen, die das Dokument nennt, und
// ob sie dort noch so steht, liest ein Mensch.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// artnamen sind die Arten ohne Liste, wie das Dokument sie schreibt.
var artnamen = map[string]string{domain.ArtGeprueft: "*geprüft*", domain.ArtFrei: "*frei*"}

// schemaWerte schreibt die Werteliste — oder die Art, wenn es keine gibt.
func schemaWerte(art string, werte []string) string {
	if art != domain.ArtListe {
		return artnamen[art]
	}

	teile := make([]string, 0, len(werte))
	for _, wert := range werte {
		teile = append(teile, "`"+wert+"`")
	}

	return strings.Join(teile, ", ")
}

// schemaFeld schreibt einen Wert in Codeschrift — oder einen Strich für leer.
func schemaFeld(wert string) string {
	if wert == "" {
		return "—"
	}

	return "`" + wert + "`"
}

// profilschema ist die Tafel der Profil- und Geräteeinstellungen.
func profilschema() []string {
	reihe := domain.Einstellungsreihe()
	zeilen := make([]string, 0, len(reihe))

	for _, schluessel := range reihe {
		ebene, _ := domain.EbeneVon(schluessel)
		art := domain.EinstellungsArt(schluessel)

		zeilen = append(zeilen, fmt.Sprintf("| %#q | %s | %s | %s | %s | %s |",
			schluessel,
			schemaFeld(domain.EinstellungsBereich(schluessel)),
			schemaFeld(string(ebene)),
			schemaWerte(art, domain.ErlaubteEinstellungswerte(schluessel)),
			schemaFeld(domain.EinstellungsVorgabe(schluessel)),
			domain.EinstellungsWirkung(schluessel)))
	}

	return zeilen
}

// nutzerschema ist die Tafel des Menschen.
//
// **Ohne Bereich und ohne Wirkung**: die Einstellungen des Menschen stehen
// auf keiner Einstellungsseite, sondern an ihrem Ort — Barrierefreiheit,
// Anmeldung —, und wann sie wirken, sagt die Profilzeile gleichen Namens,
// wo es eine gibt.
func nutzerschema() []string {
	reihe := domain.Nutzereinstellungsreihe()
	zeilen := make([]string, 0, len(reihe))

	for _, schluessel := range reihe {
		art := domain.NutzerArt(schluessel)

		zeilen = append(zeilen, fmt.Sprintf("| %#q | %s | %s |",
			schluessel,
			schemaWerte(art, domain.ErlaubteNutzerwerte(schluessel)),
			schemaFeld(domain.NutzerVorgabe(schluessel))))
	}

	return zeilen
}

// tafelNach gibt die Tabellenzeilen unter einer Überschrift, ohne Kopf.
func tafelNach(dokument, ueberschrift string) []string {
	_, rest, gefunden := strings.Cut(dokument, "\n"+ueberschrift+"\n")
	if !gefunden {
		return nil
	}

	zeilen := []string{}

	for zeile := range strings.SplitSeq(rest, "\n") {
		if strings.HasPrefix(zeile, "## ") {
			break
		}

		if strings.HasPrefix(zeile, "| `") {
			zeilen = append(zeilen, zeile)
		}
	}

	return zeilen
}

func TestDasEinstellungsschemaStehtVollstaendigDa(t *testing.T) {
	t.Parallel()

	dokument := lies(t, "docs/EINSTELLUNGEN.md")

	for ueberschrift, erwartet := range map[string][]string{
		"## Die Tafel des Profils und des Geräts": profilschema(),
		"## Die Tafel des Menschen":               nutzerschema(),
	} {
		gelesen := tafelNach(dokument, ueberschrift)
		if strings.Join(gelesen, "\n") == strings.Join(erwartet, "\n") {
			continue
		}

		t.Errorf("%s weicht vom Code ab (%d Zeilen gelesen, %d erwartet). So soll sie lauten:\n%s",
			ueberschrift, len(gelesen), len(erwartet), strings.Join(erwartet, "\n"))
	}
}
