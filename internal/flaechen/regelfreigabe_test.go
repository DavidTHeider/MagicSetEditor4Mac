package flaechen_test

// **Die Regelfreigabe über jede Fläche** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum das einen eigenen Lauf bekommt
//
// **Weil hier eine Sperre über eine Fläche geht.** Das Issue verlangt es
// wörtlich: „Turniersperre lässt sich weder über UI noch REST/CLI/MCP
// umgehen." Eine Sperre, die an einer Fläche anders antwortet als an der
// anderen, ist keine — und der Wächter über §16 sieht nur, **dass** es die
// Fähigkeit überall gibt, nicht, dass sie dasselbe sagt.
//
// # Was hier geprüft wird
//
// **Dass `streng` an jeder Fläche abgewiesen wird**, solange acht von elf
// Bereichen von Hand gemacht werden — und dass jede dabei denselben Grund
// nennt.
//
// **Und dass der Sandkasten an jeder durchkommt.** Ein Lauf, der nur das Nein
// prüfte, ließe eine Fläche durchgehen, die **alles** abweist (§10).

import (
	"net/http"
	"strings"
	"testing"
)

// **Streng wird an jeder Fläche abgewiesen — mit demselben Grund.**
func TestStrengWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	for _, fall := range []struct {
		name   string
		fragen func(strenge string) string
	}{
		{"Kommandozeile", func(strenge string) string {
			// **Ohne `muss`**: `mse regelfreigabe` gibt bei einer Absage den
			// Rückgabewert 1, und das ist die Antwort und kein Fehler.
			ausgabe, _, _ := kommandozeile(t, umgebung,
				"regelfreigabe", "magic", strenge, "2", "--json")

			return ausgabe
		}},
		{"MCP", func(strenge string) string {
			return mussMCP(t, umgebung, "lesen.regelfreigabe",
				map[string]any{
					"spiel": "magic", "strenge": strenge, "sitze": 2,
				})
		}},
		{"REST", func(strenge string) string {
			return mussREST(t, an, http.MethodGet,
				"/regelfreigabe?spiel=magic&strenge="+strenge+"&sitze=2", "")
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			// **Der Sandkasten kommt durch.**
			//
			// **Er steht zuerst**, denn ohne ihn sagte dieser Lauf nur, dass
			// etwas abgewiesen wird — auch eine Fläche, die alles abweist,
			// wäre dann grün (§10).
			offen := fall.fragen("sandkasten")

			if !strings.Contains(offen, `"erlaubt":true`) &&
				!strings.Contains(offen, `"erlaubt": true`) {
				t.Errorf("der Sandkasten kommt nicht durch: %s", offen)
			}

			// **Und streng nicht.**
			abgewiesen := fall.fragen("streng")

			if !strings.Contains(abgewiesen, `"erlaubt":false`) &&
				!strings.Contains(abgewiesen, `"erlaubt": false`) {
				t.Errorf("streng kommt durch: %s", abgewiesen)
			}

			// **Mit demselben Grund** — zwei von dreizehn Bereichen.
			//
			// **Acht von zwölf waren es bis zum 2026-09-17**: G9 hat
			// `zugstruktur` und `prioritaet` auf `automatisch` gestellt
			// (#1592), `M5` die zustandsbasierten Aktionen
			// (#1607).
			// **Und der dreizehnte kam mit `M19`**
			// (#1620):
			// der Commanderbereich, und im **Grundprofil** ist er eine
			// Lücke — dort gibt es die Variante nicht. **Und `M9`
			// (#1611)
			// hat die Wechselwirkungen geschlossen**.
			//
			// **Und seit dem 2026-09-19 sind es zwei**
			// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
			// #1626):
			// der Commander zählt im Grundprofil nicht mehr mit — eine
			// Sandkastenpartie spielt die Variante nicht, und ein Bereich,
			// den niemand schließen soll, ist keine Lücke. Übrig sind
			// Spielsteine und Typen.
			if !strings.Contains(abgewiesen, "2 von 13") {
				t.Errorf("der Grund fehlt: %s", abgewiesen)
			}

			// **Und die Quelle steht dabei, samt Fassung** (§11).
			if !strings.Contains(abgewiesen, "CR 400") ||
				!strings.Contains(abgewiesen, "2026-08-19") {
				t.Errorf("die Regelquelle fehlt: %s", abgewiesen)
			}
		})
	}
}

// **Modus und Abwahl wirken an jeder Fläche gleich** — PLAY-02.AK-01 und
// AK-02 (#1710).
//
// # Was hier geprüft wird
//
// **Dass die Abwahl im Anfängermodus überall abgewiesen wird** — mit
// demselben Ausweg im Satz —, **dass sie im freien Modus überall durchgeht**
// und **dass die Matrix danach überall dasselbe sagt**. Ein Wächter sieht,
// dass es die Fähigkeit gibt; dass sie dasselbe tut, sieht nur ein Lauf, der
// sie dreimal fährt (§16).
func TestDieRegelwahlWirktAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	for _, fall := range []struct {
		name   string
		fragen func(modus, ohne string) string
	}{
		{"Kommandozeile", func(modus, ohne string) string {
			ausgabe, _, _ := kommandozeile(t, umgebung,
				"regelfreigabe", "magic", "sandkasten", "2", modus, ohne,
				"--json")

			return ausgabe
		}},
		{"MCP", func(modus, ohne string) string {
			return mussMCP(t, umgebung, "lesen.regelfreigabe",
				map[string]any{
					"spiel": "magic", "strenge": "sandkasten", "sitze": 2,
					"modus": modus, "abgewaehlt": strings.Split(ohne, ","),
				})
		}},
		{"REST", func(modus, ohne string) string {
			return mussREST(t, an, http.MethodGet,
				"/regelfreigabe?spiel=magic&strenge=sandkasten&sitze=2"+
					"&modus="+modus+"&ohne="+ohne, "")
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			// **Im Anfängermodus geht die Abwahl nicht — und der Satz nennt
			// den Ausweg.**
			abgewiesen := fall.fragen("anfaenger", "combat")

			if !strings.Contains(abgewiesen, "freien Modus") {
				t.Errorf("der Ausweg fehlt: %s", abgewiesen)
			}

			// **Im freien Modus geht sie — und die Matrix sagt es.**
			offen := fall.fragen("frei", "prioritaet")

			if !strings.Contains(offen, `"erlaubt":true`) &&
				!strings.Contains(offen, `"erlaubt": true`) {
				t.Errorf("die Abwahl wurde abgewiesen: %s", offen)
			}

			// **Und der Widerspruch wird erklärt statt abgewiesen** — CR 608
			// und CR 704 hängen an der Priorität.
			if !strings.Contains(offen, "stapel") ||
				!strings.Contains(offen, "CR 608") {
				t.Errorf("der Widerspruch fehlt: %s", offen)
			}

			// **Der Baum kommt an jeder Fläche mit.**
			if !strings.Contains(offen, "mechanik") {
				t.Errorf("der Baum fehlt: %s", offen)
			}
		})
	}
}
