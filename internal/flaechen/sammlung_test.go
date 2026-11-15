// **Die Abnahme der Sammlung** — Phase 3.4
// (#1165).
//
// # Was diese Läufe halten
//
// Zwei Zusagen des Issues, und beide sind Aussagen über **Nähte**, die kein
// Einzeltest sieht:
//
//   - Die Fähigkeit steht an **jeder** Fläche ohne Fenster (§16) — und gibt
//     dort dieselbe Antwort.
//   - Ein Set ein- und auszufahren lässt Sammlungsdaten **unberührt**. Sie
//     gehören nicht in eine `.mse-set`
//     ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
//
// **Die dritte — dass die Sicherung sie mitnimmt — steht in
// `internal/app`**: es gibt sie nur am Fenster, und ein Lauf ohne Fenster
// könnte sie nicht rufen.
//
// **Die Erklärung, was solche Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probedruck ist die Kennung, die diese Läufe benutzen.
//
// **Eine erfundene und keine echte.** Sie geht nie ins Netz: geholt wird hier
// nichts, und eine echte Kennung lüde dazu ein, diesen Lauf mit einem Netzlauf
// zu verwechseln.
const probedruck = "abnahme-druck-0001"

// **Ein Exemplar über jede Fläche aufnehmen, wiederfinden und ändern.**
//
// # Warum das ganze Rundstück und nicht nur das Anlegen
//
// Der erste Anlauf legte je Fläche ein **minimales** Exemplar an und suchte die
// Druckkennung in der Ausgabe — eine Existenzprobe, ausgegeben als
// Flächengleichheit. Der zweite verglich sieben Werte, aber weiter als
// **Zeichenketten in unstrukturiertem Text**, und er ließ das **Ändern** ganz
// aus.
//
// **Was das gekostet hat, ist gemessen**: die Fassungsbedingung aus #1196 kam
// an MCP und REST nie an — deren Transporttypen kannten das Feld nicht —, und
// damit war `PUT /sammlung/{id}` **unmöglich**: der Dienst antwortete mit 409.
// Der Lauf hier blieb grün. Angestrichen vom Nachlauf über Phase 3.4 als
// Befund 6 und 7.
//
// **Deshalb jetzt drei Sätze statt einem:**
//
//  1. Gelesen wird die **Struktur** jeder Fläche, nicht ihr Text. Eine Fläche,
//     die ein Feld unter anderem Namen führt, fällt auf.
//  2. Verglichen wird **jedes** Feld — auch `menge`, `foil` und `fassung`.
//  3. Gefahren wird der ganze Weg: anlegen, lesen, **ändern**, wieder lesen.
func TestEinExemplarUeberJedeFlaeche(t *testing.T) {
	for name, fall := range map[string]struct {
		aufnehmen func(t *testing.T, umgebung []string)
		lesen     func(t *testing.T, umgebung []string) flaechenexemplar
		aendern   func(t *testing.T, umgebung []string, gelesen flaechenexemplar)
	}{
		"Fenster": exemplarfallFenster(),
		"Kommandozeile": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				mussKommandozeile(t, umgebung, "exemplar", probedruck, "2",
					"played", "de", "foil", "2026-08-01", "4.17", "EUR",
					"aus der Abnahme")
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				return ersteAusHuelle(t,
					mussKommandozeile(t, umgebung, "--json", "sammlung"))
			},
			aendern: func(t *testing.T, umgebung []string, gelesen flaechenexemplar) {
				t.Helper()

				// **Die Kommandozeile liest, verschmilzt und schreibt selbst.**
				// Sie nimmt die Fassung deshalb nicht entgegen — und genau das
				// ist die Ausnahme, die §16 verlangt zu begründen: wer tippt,
				// hat den vorigen Stand nicht zur Hand.
				mussKommandozeile(t, umgebung, "exemplar-aendern",
					fmt.Sprint(gelesen.ID), "3")
			},
		},
		"MCP": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "schreiben.exemplar",
					vollesExemplar())
				if istFehler {
					t.Fatalf("schreiben.exemplar: %s", text)
				}
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "lesen.sammlung",
					map[string]any{})
				if istFehler {
					t.Fatalf("lesen.sammlung: %s", text)
				}

				return ersteAusHuelle(t, text)
			},
			aendern: func(t *testing.T, umgebung []string, gelesen flaechenexemplar) {
				t.Helper()

				werte := vollesExemplar()
				werte["id"] = gelesen.ID
				werte["fassung"] = gelesen.Fassung
				werte["menge"] = 3

				text, istFehler := mcpRuf(t, umgebung,
					"schreiben.exemplaraendern", werte)
				if istFehler {
					t.Fatalf("schreiben.exemplaraendern: %s", text)
				}
			},
		},
		"Dienst": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodPost, "/sammlung",
					alsRumpf(t, vollesExemplar()))
				if status != http.StatusCreated {
					t.Fatalf("anlegen: %d (%s)", status, rumpf)
				}
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodGet, "/sammlung", "")
				if status != http.StatusOK {
					t.Fatalf("lesen: %d (%s)", status, rumpf)
				}

				return ersteAusHuelle(t, rumpf)
			},
			aendern: func(t *testing.T, umgebung []string, gelesen flaechenexemplar) {
				t.Helper()

				an := starteDienst(t, umgebung)

				werte := vollesExemplar()
				werte["id"] = gelesen.ID
				werte["fassung"] = gelesen.Fassung
				werte["menge"] = 3

				status, rumpf := an.ruf(t, http.MethodPut,
					"/sammlung/"+fmt.Sprint(gelesen.ID), alsRumpf(t, werte))
				if status != http.StatusOK {
					t.Fatalf("ändern: %d (%s)", status, rumpf)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := eigenesZuhause(t)

			fall.aufnehmen(t, umgebung)

			gelesen := fall.lesen(t, umgebung)

			// **Feld für Feld gegen dieselbe Erwartung.** Eine Fläche, die
			// eines nicht ausdrücken kann, fällt hier auf — und nicht erst,
			// wenn jemand seine Angaben verliert.
			pruefeExemplar(t, gelesen, 2)

			// **Und die Fassung ist da.** Ohne sie kann niemand ändern, ohne
			// eine fremde Änderung zu überschreiben (§7, Regel 2) — und an
			// zwei Flächen konnte man deshalb **gar nicht** ändern.
			if gelesen.Fassung < 1 {
				t.Errorf("die Fläche nennt keine Fassung: %+v", gelesen)
			}

			// **Der zweite Satz: ändern und wieder lesen.**
			fall.aendern(t, umgebung, gelesen)

			pruefeExemplar(t, fall.lesen(t, umgebung), 3)
		})
	}
}

// flaechenexemplar ist ein Exemplar, wie **jede** Fläche es herausgibt.
//
// **Eine Gestalt für alle drei.** Sie ist die Zusage aus §16 in Typform: „was
// das Fenster kann, können Kommandozeile und MCP" heißt auch, dass sie
// dieselben Felder unter denselben Namen führen. Eine Fläche, die eines
// umbenennt, lässt hier eine Null stehen.
type flaechenexemplar struct {
	ID       int64  `json:"id"`
	Fassung  int64  `json:"fassung"`
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Foil ist der **alte** Wahrheitswert, abgeleitet aus [flaechenexemplar.Finish]
	// — abgekündigt seit T2 (0119). Er steht hier, weil ein Gegenüber, das
	// diese Fassung nicht kennt, ihn weiter liest und schreibt.
	Foil bool `json:"foil"`

	// Finish ist die Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish          string `json:"finish"`
	Erwerbsdatum    string `json:"erwerbsdatum"`
	Erwerbspreis    string `json:"erwerbspreis"`
	Erwerbswaehrung string `json:"erwerbswaehrung"`
	Notiz           string `json:"notiz"`
}

// pruefeExemplar hält ein gelesenes Exemplar gegen die Erwartung.
func pruefeExemplar(t *testing.T, gelesen flaechenexemplar, menge int) {
	t.Helper()

	erwartet := flaechenexemplar{
		ID: gelesen.ID, Fassung: gelesen.Fassung,
		Printing: probedruck, Menge: menge, Zustand: "played", Sprache: "de",
		Foil: true, Finish: "foil",
		Erwerbsdatum: "2026-08-01", Erwerbspreis: "4.17",
		Erwerbswaehrung: "EUR", Notiz: "aus der Abnahme",
	}

	if gelesen != erwartet {
		t.Errorf("die Fläche gibt\n  %+v\nerwartet\n  %+v", gelesen, erwartet)
	}
}

// ersteAusHuelle liest das erste Exemplar aus einer Antwort.
//
// **Über die Struktur und nicht über einen Textfund.** Jede der drei Flächen
// verpackt anders — die eine in `{"exemplare": […]}`, die andere gibt dieselbe
// Hülle im Werkzeugtext zurück —, und der gemeinsame Nenner ist, dass ein
// **Feld** `exemplare` darin steht. Wer stattdessen nach Zeichenketten sucht,
// prüft, ob ein Wert **irgendwo** vorkommt, und nicht, ob er im richtigen Feld
// steht.
func ersteAusHuelle(t *testing.T, roh string) flaechenexemplar {
	t.Helper()

	var huelle struct {
		Exemplare []flaechenexemplar `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	if len(huelle.Exemplare) != 1 {
		t.Fatalf("es sind %d Exemplare: %s", len(huelle.Exemplare), roh)
	}

	return huelle.Exemplare[0]
}

// **Ein Set ein- und auszufahren lässt die Sammlung unberührt.**
//
// # Warum das die wichtigste Zusage dieser Phase ist
//
// Sammlungs- und Preisdaten sind **persönlich**; eine `.mse-set` gibt man
// weiter. Landeten sie darin, bekäme sie jeder, dem man ein Set schickt — und
// niemand bekäme sie je wieder heraus.
//
// **Geprüft wird in beide Richtungen**: die geschriebene Datei nennt sie
// nicht, und der Bestand hat sie danach noch.
func TestEinSetrundlaufLaesstDieSammlungUnberuehrt(t *testing.T) {
	umgebung := eigenesZuhause(t)

	mussKommandozeile(t, umgebung, "exemplar", probedruck, "3")
	mussKommandozeile(t, umgebung, "oeffnen", probeset(t))

	setID := ersteSetkennung(t, mussKommandozeile(t, umgebung, "sets"))

	ziel := filepath.Join(t.TempDir(), "hinaus.mse-set")

	mussKommandozeile(t, umgebung, "sichern", setID, ziel)

	// **Gesucht wird in jeder Datei des Pakets.** Ein `.mse-set` ist ein
	// **Verzeichnis** und keine Datei — der erste Anlauf las es als eine und
	// scheiterte an „is a directory". Eine Suche in einer Datei hätte über die
	// anderen nichts gesagt (§10).
	if fundstelle := durchsucht(t, ziel, probedruck); fundstelle != "" {
		t.Errorf("%s trägt eine Kennung aus der Sammlung", fundstelle)
	}

	// **Und die Gegenprobe: die Suche kann etwas finden.**
	//
	// Ein Lauf, der nichts findet, hat zwei Lesarten — die Sammlung steht
	// nicht drin, oder die Suche sieht nichts. Erst diese Zeile trennt sie:
	// derselbe Text als **Kartenwert** geschrieben, und er steht in der Datei.
	karteID := ersteSetkennung(t, mussKommandozeile(t, umgebung, "karten", setID))

	mussKommandozeile(t, umgebung, "feld", setID, karteID, "name", probedruck)

	gegenprobe := filepath.Join(t.TempDir(), "gegenprobe.mse-set")

	mussKommandozeile(t, umgebung, "sichern", setID, gegenprobe)

	if durchsucht(t, gegenprobe, probedruck) == "" {
		t.Error("die Suche findet den Text auch dann nicht, wenn er dasteht")
	}

	// **Und nach dem Rundlauf ist sie noch da.**
	mussKommandozeile(t, umgebung, "oeffnen", ziel)

	if gelesen := mussKommandozeile(t, umgebung, "sammlung"); !strings.Contains(
		gelesen, probedruck) {
		t.Errorf("nach dem Rundlauf fehlt das Exemplar: %s", gelesen)
	}
}

// ersteSetkennung liest die Kennung aus der ersten Zeile von `mse sets`.
func ersteSetkennung(t *testing.T, ausgabe string) string {
	t.Helper()

	zeilen := strings.Split(strings.TrimSpace(ausgabe), "\n")
	if len(zeilen) == 0 || zeilen[0] == "" {
		t.Fatalf("keine Sets: %q", ausgabe)
	}

	kennung, _, _ := strings.Cut(zeilen[0], "\t")

	return kennung
}

// durchsucht sucht einen Text in jeder Datei unter einem Ort.
//
// **Gibt die Fundstelle zurück und nicht nur ja oder nein**: bei einem Fund
// soll dastehen, welche Datei ihn trägt.
func durchsucht(t *testing.T, ort, gesucht string) string {
	t.Helper()

	fundstelle := ""

	fehler := filepath.WalkDir(ort, func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() {
			return err
		}

		// #nosec G304,G122 -- ein Verzeichnis aus t.TempDir(), das dieser Lauf
		// selbst hat schreiben lassen
		roh, lesefehler := os.ReadFile(pfad)
		if lesefehler != nil {
			return lesefehler
		}

		if strings.Contains(string(roh), gesucht) {
			fundstelle = pfad
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchsuchen: %v", ort, fehler)
	}

	return fundstelle
}

// vollesExemplar ist ein Exemplar mit **jedem** Feld.
//
// **Einmal aufgeschrieben und von jeder Fläche benutzt**: zwei Erwartungen
// daneben liefen an dem Tag auseinander, an dem eine ein Feld dazubekommt.
func vollesExemplar() map[string]any {
	return map[string]any{
		"printing": probedruck, "menge": 2,
		"zustand": "played", "sprache": "de", "foil": true,
		"erwerbsdatum": "2026-08-01", "erwerbspreis": "4.17",
		"erwerbswaehrung": "EUR", "notiz": "aus der Abnahme",
	}
}

// alsRumpf macht aus einer Tabelle einen JSON-Rumpf.
//
// **Nicht `alsJSONText`** — das gibt es schon in `restablauf_test.go` und nimmt
// eine fertige Zeichenkette entgegen.
func alsRumpf(t *testing.T, werte map[string]any) string {
	t.Helper()

	roh, err := json.Marshal(werte)
	if err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	return string(roh)
}
