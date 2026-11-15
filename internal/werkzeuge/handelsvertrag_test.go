package werkzeuge_test

// **Der Handelsvertrag bewegt keine Exemplare** — Werkstück `T1` von Phase
// 5.6 (#1466).
//
// # Die zweite beobachtbare Abnahme des Issues
//
// **„Jede Kopie und jede Deckleihe verändert exakt null physische Bestände;
// Marktpreise und günstige Differenz sind niemals Zulässigkeitsbedingung."**
//
// Das ist eine Aussage über den ganzen Vertrag und nicht über einen Aufruf —
// und deshalb steht hier ein Wächter und kein Lauf: **verbucht wird in T6**
// (#1471),
// und was T1 selbst schreibt, sind Revisionen und sonst nichts.
//
// # Was er hält
//
// Kein Quelltext des Vertrags nennt die Exemplartabelle, eine
// Exemplarablage oder einen Preis.
//
// # Und das Schema
//
// **Bis 0.7.0 las er auch die sechs Wanderungen der Handelstabellen** —
// `0068`, `0073`, `0074`, `0076`, `0093` und `0095`. Seit es nur die
// Grundlinie gibt (Vermerk 0189), stehen sie in ihr und nicht mehr als Datei,
// und die Grundlinie ist eingefroren (`TestDieGrundlinieIstUnveraendert`).
// Ihr Inhalt ist der von 0.7.0 — alle 97 Prüfsummen verglichen am
// 2026-10-12 —, und diesen hat jede Pipeline bis dahin hier grün gelesen.
// **Eine künftige Wanderung an einer Handelstabelle gehört wieder in die
// Liste.**
//
// # Was er nicht hält
//
//   - **Dass T6 die Zusage einhält.** Wer dort verbucht, tut es absichtlich;
//     dieser Wächter sieht nur, dass der **Vertrag** es nicht tut.
//   - **Dass ein Preis nicht über einen Umweg hereinkommt.** Er liest Namen,
//     keine Aufrufketten. Was er dagegen ausschließt, ist die naheliegende
//     Gestalt: eine Bedingung, die eine Zahl aus dem Markt liest.

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// vertragsdateien sind die Quelltexte, die den Handelsvertrag tragen.
var vertragsdateien = []string{
	"../service/handelszusagenwirkung.go",
	"../service/handelsschranke.go",
	"../service/handelspostvermerk.go",
	"../domain/handelsstandnahme.go",
	"../service/handelszusagen.go",
	"../domain/handelsangebot.go",
	"../domain/handelspruefung.go",
	"../domain/handelsmatrix.go",
	"../domain/handelsschritt.go",
	"../domain/handelsfolge.go",
	"../domain/handelsrichtlinie.go",
	"../service/handelsleitung.go",
	// **Und die Wahl eines Handels** — UX-08.AK-04
	// (#1680):
	// die vergebene Kennung und die Liste mit einem Gegenüber.
	"../service/handelswahl.go",
	"../storage/handelsvorgaenge.go",

	// **Und die Liste mit ihren Entwürfen** — TRADE-01
	// (#1715).
	"../domain/handelslage.go",
	"../domain/handelsphase.go",
	"../domain/handelsempfang.go",
	"../domain/handelsfilter.go",
	"../service/handelsverzeichnis.go",
	"../storage/handelsuebersicht.go",

	// **Und was abgeschlossen ist, bleibt** — TRADE-03.AK-01, AK-03 und AK-04
	// (#1717):
	// bedingtes Fortschreiben, das Protokoll, der Riegel gegen stilles
	// Umschreiben und die Klärung bei widersprüchlicher Post.
	"../domain/handelsprotokoll.go",
	"../storage/handelsprotokoll.go",
	"../storage/handelsfortschreiben.go",
	"../service/handelspostklaerung.go",
	"../service/handelspostausweis.go",

	// **Und was das Protokoll trägt** — Werkstück T5 von Phase 5.6
	// (#1470).
	// Der Umschlag, die Zustellung und die Wiederaufnahme sprechen über
	// denselben Vertrag; wer dort von Exemplaren spräche, hätte den Handel
	// wieder an den Bestand gebunden.
	"../domain/handelsumschlag.go",
	"../domain/handelszustellung.go",
	"../domain/handelswiederaufnahme.go",
	"../domain/handelsjournal.go",
	"../service/handelspost.go",

	// **Und der Beleg der Gegenseite** —
	// #1650.
	// Er prüft eine fremde Unterschrift gegen den **eigenen** Beleg; wer
	// dort von Exemplaren spräche, machte aus einer Urkunde eine
	// Bestandsfrage.
	"../service/handelsgegenbeleg.go",
	"../domain/fremdbeleg.go",
	"../service/handelsstrom.go",
	"../service/handelsstrecke.go",
	"../storage/handelsbriefkasten.go",
	"../storage/handelsausgang.go",
	"../domain/handelsabschluss.go",
	"../domain/belegstand.go",
	"../service/handelsverbuchung.go",
	"../service/handelsjournalsicht.go",
	// **Und welche Karte eine Zeile meint** — P10
	// (#2378):
	// nachgeschlagen beim Lesen, wie das Journal; der Vertrag führt nur die
	// Kennung.
	"../service/handelsdrucke.go",
	// **Und was ein Gegenüber zum Tausch freigibt** — P11
	// (#2379):
	// eine Auskunft über die Leitung, die nichts bucht und nichts speichert.
	"../service/handelssicht.go",
	"../storage/handelsabschluss.go",
	"../storage/handelszusagen.go",
	"../storage/handelsbestand.go",
	"../storage/handelsjournal.go",

	// **Und der Riegel davor** — Werkstück T10 von Phase 5.6
	// (#1475).
	// Er entscheidet, **wer** verbuchen darf, und spricht dabei über Vorgang,
	// Stand und Nachweis — nicht über Besitzzeilen. Stünde er nicht hier,
	// könnte er den Bestand anfassen, ohne dass es jemand merkt.
	"../service/handelsfreigabe.go",

	// **Und was den Vertrag hinausträgt** — angeschlossen mit
	// #1637.
	// Die Wahl der Nachrichtenart hängt am **Zustand** des Vorgangs, der
	// Absender baut den Umschlag daraus, und die Folge zählt je Vorgang —
	// alle drei sprechen über den Vertrag und nicht über Besitzzeilen.
	"../domain/handelsnachrichtenwahl.go",
	"../service/handelsversand.go",
	"../storage/handelsfolge.go",

	// **Und was drüben zu buchen ist** — #2495: die Zeile, mit der die
	// empfangende Seite bucht, und ihre Prüfung vor der Einigung. Sie
	// beschreibt eine Buchung und nimmt keine vor.
	"../domain/handelszugang.go",

	// **Und welche Revision ein Schritt meint** — #2496: ein Vergleich von
	// Abdrücken, nichts am Bestand.
	"../domain/handelsgesehen.go",
}

// verbotenImVertrag ist, was dort nicht vorkommen darf.
//
// **Die Tabelle heißt `sammlung` und nicht `exemplare`** — der erste Anlauf
// dieses Wächters verbot `exemplare` und `Exemplarablage` und traf damit den
// Schreibweg nicht, den er meinte. **Ein Wächter, der die halbe Zusage hält,
// macht die andere unsichtbar** (§10); gefunden beim Lesen von T2 am selben
// Tag.
//
// **Nicht das Wort „Exemplar"**: der Vertrag spricht in seinen Kommentaren
// über Exemplare, und das soll er — er bewegt sie nur nicht.
var verbotenImVertrag = map[string]string{
	"sammlung":          "die Sammlungstabelle — verbucht wird in der Verbuchung",
	"Sammlungsablage":   "die Sammlungsablage — verbucht wird in der Verbuchung",
	"Sammlungsspeicher": "der Sammlungsspeicher — verbucht wird in der Verbuchung",
	"exemplare":         "eine Exemplarliste — verbucht wird in der Verbuchung",
}

// verbotenImHandel ist, was **nirgends** im Handel vorkommen darf.
//
// **Ein Preis ist nie eine Zulässigkeitsbedingung** — und er ist erst recht
// keine Kostenbasis für ein empfangenes Stück (T6, #1471). Diese Wörter sind
// deshalb auch in der Verbuchung verboten, die die Sammlung anfassen darf.
var verbotenImHandel = map[string]string{
	"Preisstand":    "ein Preis ist nie eine Zulässigkeitsbedingung",
	"Preisablage":   "ein Preis ist nie eine Zulässigkeitsbedingung",
	"Marktpreis":    "ein Preis ist nie eine Zulässigkeitsbedingung",
	"mtgjsonpreise": "ein Preis ist nie eine Zulässigkeitsbedingung",
}

// verbuchungsdateien sind die Stellen, die den Bestand bewegen **dürfen**.
//
// # Warum es diese zweite Liste gibt
//
// **Weil T6 genau das tut, was der Wächter bis dahin verbot** (#1471): ein
// Abschluss bucht, und dafür fasst er `sammlung` an. Die Zusage bleibt
// dieselbe — der **Vertrag** bewegt keine Exemplare —, aber sie gilt jetzt
// für eine Liste und nicht mehr für alle `handels*.go`.
//
// **Eine Aufzählung und kein Vorsatz**: eine vierte Datei, die den Bestand
// anfasst, ist eine Entscheidung und kein Nebeneffekt. Wer sie einträgt, sagt
// dazu, wogegen sie prüft — und `internal/werkzeuge/bestandsschreibwege_test.go`
// hält daneben, dass sie überhaupt schreiben darf.
var verbuchungsdateien = []string{
	"../storage/handelsabschluss.go",
	"../storage/handelsbestand.go",
	"../storage/handelsjournal.go",
}

// **Der Vertrag nennt weder eine Exemplartabelle noch einen Preis.**
func TestDerHandelsvertragBewegtKeineExemplare(t *testing.T) {
	t.Parallel()

	for _, pfad := range vertragsdateien {
		t.Run(filepath.Base(pfad), func(t *testing.T) {
			t.Parallel()

			// **Die Verbuchung bucht — sie ist der Zweck der Ausnahme.**
			verboten := verbotenImHandel

			if !slices.Contains(verbuchungsdateien, pfad) {
				verboten = map[string]string{}
				maps.Copy(verboten, verbotenImHandel)
				maps.Copy(verboten, verbotenImVertrag)
			}

			roh, fehler := os.ReadFile(pfad) // #nosec G304 -- feste Liste im Baum
			if fehler != nil {
				t.Fatalf("%q ist nicht lesbar: %v", pfad, fehler)
			}

			inhalt := string(roh)

			for wort, warum := range verboten {
				if strings.Contains(inhalt, wort) {
					t.Errorf("%q nennt %q — %s", pfad, wort, warum)
				}
			}
		})
	}
}

// **Und die Ausnahmeliste ist keine Sammelstelle.**
//
// Jede Datei darin steht auch in `vertragsdateien` — sonst wäre sie eine
// Erlaubnis für etwas, das der Wächter gar nicht liest, und niemand sähe es.
func TestJedeVerbuchungsdateiStehtImVertrag(t *testing.T) {
	t.Parallel()

	for _, pfad := range verbuchungsdateien {
		if !slices.Contains(vertragsdateien, pfad) {
			t.Errorf("%q darf buchen, wird aber gar nicht gelesen", pfad)
		}
	}
}

// **Und die Aufzählung oben ist vollständig.**
//
// **Ohne diesen Lauf wäre der Wächter eine Zusage über eine Liste**, die beim
// nächsten `handels*.go` still zu kurz wird — dieselbe Gestalt wie bei der
// Seitenliste der Barrierefreiheit (§3).
func TestDieVertragsdateienSindVollstaendig(t *testing.T) {
	t.Parallel()

	for _, ort := range []string{"../domain", "../service", "../storage"} {
		gefunden, fehler := filepath.Glob(filepath.Join(ort, "handels*.go"))
		if fehler != nil {
			t.Fatalf("%q: %v", ort, fehler)
		}

		for _, pfad := range gefunden {
			if strings.HasSuffix(pfad, "_test.go") {
				continue
			}

			if !slices.Contains(vertragsdateien, filepath.ToSlash(pfad)) {
				t.Errorf("%q steht nicht in vertragsdateien", pfad)
			}
		}
	}
}

// **Und jedes Verbotswort gibt es wirklich.**
//
// # Der Befund, für den dieser Lauf steht
//
// **Gesehen am 2026-09-18, an diesem Wächter selbst**: er verbot
// `Exemplarablage` und `exemplare` als „die Exemplartabelle" — und die
// Tabelle heißt `sammlung`. Zwei der sechs Wörter trafen damit nichts, was
// der Vertrag je berühren könnte, und der Wächter war grün, ohne etwas zu
// halten.
//
// **Ein Verbot über ein Wort, das es nicht gibt, ist keines.** Dieser Lauf
// hält es: jedes Wort kommt **irgendwo** unter `internal/` vor — nur eben
// nicht im Vertrag.
//
// # Was er nicht hält
//
// **Dass das Wort den richtigen Schreibweg trifft.** Er sieht, dass es
// existiert, nicht dass es die Tabelle meint, die Bestände hält. Das bleibt
// Handarbeit — und es ist genau die, die hier einmal gefehlt hat.
func TestJedesVerbotswortGibtEsWirklich(t *testing.T) {
	t.Parallel()

	gefunden := map[string]bool{}

	fehler := filepath.WalkDir("..", func(
		pfad string, eintrag os.DirEntry, fehler error,
	) error {
		if fehler != nil {
			return fehler
		}

		// **Diese Datei zählt nicht mit.** Sie führt jedes Verbotswort in
		// ihrem Literal — wer sie mitläse, fände jedes Wort in ihr selbst,
		// und der Lauf wäre grün, gleich was in der Liste steht. Gemessen
		// am 2026-09-18 mit einem erfundenen Wort: er blieb grün.
		if eintrag.IsDir() || !passendeEndung(pfad) ||
			strings.HasSuffix(pfad, "handelsvertrag_test.go") ||
			slices.Contains(vertragsdateien, filepath.ToSlash(pfad)) {
			return nil
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Weg aus dem Baumgang
		if lesefehler != nil {
			return lesefehler
		}

		for wort := range verbotenImVertrag {
			if strings.Contains(string(roh), wort) {
				gefunden[wort] = true
			}
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("der Baum ist nicht lesbar: %v", fehler)
	}

	for wort, warum := range verbotenImVertrag {
		if !gefunden[wort] {
			t.Errorf("%q kommt nirgends vor — ein Verbot über ein Wort, "+
				"das es nicht gibt, hält nichts (%s)", wort, warum)
		}
	}
}

// passendeEndung sagt, ob diese Datei Quelltext ist.
func passendeEndung(pfad string) bool {
	return strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, ".sql")
}
