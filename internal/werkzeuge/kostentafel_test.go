package werkzeuge

// **Die Kostentafel nennt nur Wege, die es gibt** — Folgebefund F1 des
// Lastlaufs QA-08 (#2003).
//
// # Was er hält
//
//   - **Jede Zeile nennt eine Fähigkeit aus `docs/FAEHIGKEITEN.md`**, und
//     jeder ihrer Wege steht dort bei ihr, an derselben Fläche — ein REST-Weg
//     ohne die Namen seiner Platzhalter. Ein umbenannter Weg fiele sonst still
//     auf den Preis seiner Art zurück.
//   - **Eine Zeile hat an jeder Fläche einen Weg**, an der ihre
//     Fähigkeit einen hat und die bucht — nicht die Kommandozeile, wenn die
//     Zeile Lesen ist (#2020). Das ist §16: dieselbe Fähigkeit, dieselben Kosten
//     — eine Zeile, die eine Fläche ausließe, wäre genau der Befund.
//   - **Kein Weg steht in zwei Zeilen.**
//
// # Was er nicht hält
//
// **Ob eine Fähigkeit teuer sein sollte**: das ist eine Entscheidung, und sie
// steht mit Grund in `internal/domain/kostentafel.go`. **Und ob eine neue
// teure Fähigkeit ihre Zeile bekommt** — ohne sie kostet sie überall, was
// ihre Art kostet; das ist gleich, aber vielleicht zu billig. Dass die
// Flächen die Tafel wirklich fragen, halten die Läufe je Fläche und
// `internal/flaechen/kosten_test.go`.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieKostentafelNenntNurWegeDesKatalogs(t *testing.T) {
	_, wege := faehigkeitswege(t)
	gesehen := map[domain.Flaeche]map[string]string{}

	for _, zeile := range domain.Kostentafel() {
		katalog, steht := wege[zeile.Faehigkeit]
		if !steht {
			t.Errorf("%q steht nicht in FAEHIGKEITEN.md", zeile.Faehigkeit)

			continue
		}

		for _, flaeche := range []domain.Flaeche{domain.FlaecheMCP, domain.FlaecheCLI, domain.FlaecheREST} {
			bekannt := katalog[flaeche]
			if flaeche == domain.FlaecheREST {
				bekannt = restwegeOhneNamen(bekannt)
			}

			for _, weg := range zeile.Wege[flaeche] {
				if !slices.Contains(bekannt, weg) {
					t.Errorf("%q: %s %q steht nicht bei ihr im Katalog (dort: %v)",
						zeile.Faehigkeit, flaeche, weg, bekannt)
				}

				if gesehen[flaeche] == nil {
					gesehen[flaeche] = map[string]string{}
				}

				if vorher, doppelt := gesehen[flaeche][weg]; doppelt {
					t.Errorf("%s %q steht in %q und in %q", flaeche, weg, vorher, zeile.Faehigkeit)
				}

				gesehen[flaeche][weg] = zeile.Faehigkeit
			}

			// **Die Kommandozeile bucht nichts** — eine Zeile Lesen nennt ihre Wege
			// dort nicht (`kostenlesen.go` sagt, warum), und ein Abbruch gibt es
			// nur am Dienst.
			if zeile.Klasse == domain.KostenAbbruch || (zeile.Klasse == domain.KostenLesen && flaeche == domain.FlaecheCLI) {
				continue
			}

			if len(bekannt) > 0 && len(zeile.Wege[flaeche]) == 0 {
				t.Errorf("%q kostet %s, aber an %s nach ihrer Art — dort hat sie %v",
					zeile.Faehigkeit, zeile.Klasse, flaeche, bekannt)
			}
		}
	}
}

// restwegeOhneNamen bringt REST-Wege in die Gestalt der Tafel.
func restwegeOhneNamen(wege []string) []string {
	ohne := make([]string, 0, len(wege))
	for _, weg := range wege {
		ohne = append(ohne, domain.RESTWeg(weg))
	}

	return ohne
}
