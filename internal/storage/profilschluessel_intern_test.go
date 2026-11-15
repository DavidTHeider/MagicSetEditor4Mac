package storage

import (
	"slices"
	"strings"
	"testing"
)

// **Ein Schlüssel an einer Profiltafel nennt das Profil** — Befund 1 des
// Zweitblicks über Phase 6.1.
//
// # Der Befund, für den es diesen Wächter gibt
//
// **`deckleihen` trug `profil_id` und hatte `kennung` als Schlüssel.** Eine
// Leihe liegt aber bei **beiden** Seiten mit derselben Kennung — und führt
// jemand zwei Profile in einem Bestand, leiht er von einem ins andere. Das
// `ON CONFLICT (kennung) DO UPDATE` der Annahme traf dann die Zeile des
// Verleihers: der Entleiher bekam keine eigene, und Verbrauch wie Widerruf
// schrieb er in ein fremdes Profil. Die Besitzläufe aus #1815 sahen es nicht,
// denn die Ablage fragte nach dem Profil — nur der Schlüssel tat es nicht.
//
// # Was er hält
//
// **Jede eindeutige Bedingung einer Tafel mit `profil_id` nennt das Profil**
// — oder die Tafel steht mit Grund in [schluesselOhneProfil]. Der
// Rowid-Schlüssel (`INTEGER PRIMARY KEY`) zählt nicht: ihn vergibt der
// Bestand und nie ein Aufrufer.
//
// # Was er nicht hält
//
// **Ob ein Grund trägt.** Er steht da; nachgelesen wird er im Review — und
// jeder hier sagt, **wer** die Kennung vergibt.

// schluesselOhneProfil sind eindeutige Bedingungen ohne `profil_id` — mit
// Grund, und jeder nennt, wer die Kennung vergibt.
//
// **Hier stand auch `kostenposten(kennung)`**, mit dem Grund, der Dienst würfle
// die Kennung. Der Nachlauf über Phase 6.1 (Befund 4) las nach: sie kommt vom
// Aufrufer, und die Probe eines Modells heißt in jedem Profil gleich. Seit
// Wanderung 0082 nennt der Schlüssel das Profil.
var schluesselOhneProfil = map[string]string{
	"freigabenachweise(kennung)": "die Kennung würfelt der Dienst beim Ausstellen " +
		"(`Nachweise.Ausstellen`), und das Einlösen prüft Nutzer und Profil (`Nachweis.Gilt`)",
	"partien(id)": "die Kennung würfelt die Anwendung beim Anlegen (`NeuePartiekennung`, " +
		"`crypto/rand`); eine Einfuhr bekommt eine neue (`Partieleitung.Einfuhr`, `domain.Kopiere`)",
	"retrieval_dokumente(generation, kennung)": "heute gibt es nur öffentliche Dokumente " +
		"(`profil_id` 0, `retrievalkorpus.go`); ein privates bräuchte das Profil im Schlüssel",
}

func TestJederSchluesselEinerProfiltafelNenntDasProfil(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	gesehen := map[string]bool{}

	for _, tafel := range eigeneTafeln(t, ctx, db) {
		if einWert(t, ctx, db,
			`SELECT count(*) FROM pragma_table_info(?) WHERE name = 'profil_id'`, tafel) == "0" {
			continue
		}

		for _, index := range zeilenAus(t, ctx, db,
			`SELECT name FROM pragma_index_list(?) WHERE "unique" = 1`, tafel) {
			var spalten []string
			for _, spalte := range zeilenAus(t, ctx, db,
				`SELECT name FROM pragma_index_info(?) ORDER BY seqno`, index[0]) {
				spalten = append(spalten, spalte[0])
			}

			name := tafel + "(" + strings.Join(spalten, ", ") + ")"
			if slices.Contains(spalten, "profil_id") {
				continue
			}

			gesehen[name] = true

			if schluesselOhneProfil[name] == "" {
				t.Errorf("%s ist eindeutig, ohne das Profil zu nennen — eine gleiche Kennung "+
					"in zwei Profilen träfe dieselbe Zeile", name)
			}
		}
	}

	for name := range schluesselOhneProfil {
		if !gesehen[name] {
			t.Errorf("%s steht in schluesselOhneProfil, nennt das Profil aber oder ist fort", name)
		}
	}
}
