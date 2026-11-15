// **Die Durchsetzung sitzt, wo die Matrix es sagt** — Werkstück O11 von
// Phase 5.2 (#1435).
//
// # Der Befund, für den es diesen Lauf gibt
//
// `internal/werkzeuge/autorisierung_test.go` hält `docs/AUTORISIERUNG.md`
// gegen `FAEHIGKEITEN.md` und gegen `domain` — und sagt in seinem eigenen Kopf,
// was er **nicht** hält: „ob die Durchsetzung dort sitzt, wo die Zeile es
// sagt". Ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar
// (§10); dieser Lauf ist die andere Hälfte.
//
// # Was er fährt
//
// Einen **gestarteten Dienst** über einem echten Bestand, mit zwei Token
// desselben Profils: eines darf nur lesen, eines darf schreiben. Jeder Fall
// geht **zweimal** hinaus — mit dem Lesetoken und mit dem schreibenden. Ohne
// die zweite Hälfte hieße ein Nein nur, dass der Weg überhaupt nicht ging
// (§10: eine Gegenprobe, die nichts ändert, ist keine).
//
// # Was er nicht fährt
//
//   - **Nicht jeden Endpunkt.** Er fährt einen Fall je Zusage, nicht die
//     ganze Liste: dass **jede** Fähigkeit in der Matrix steht, hält der
//     Wächter daneben, und die Rollen selbst prüft `domain.Akteur`.
//   - **`verbindlich` gar nicht** — die Matrix zählt heute **null** solche
//     Fähigkeiten; sie kommen mit Phase 5.6. Ein Lauf darüber prüfte eine
//     leere Menge und sähe aus wie einer.
//   - **`externSenden` nur an seiner Abweisung**: was wirklich hinausgeht,
//     steht hinter dem Dienstschalter (§16, Regel 6) und hinter dem Build-Tag
//     `netzlauf`.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// **Ein Lesetoken wird abgewiesen — und der Bestand bleibt, wie er war.**
func TestEinLesetokenSchreibtAnKeinerStelle(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Matrixlauf", "--json"))

	nurLesen := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Leser","scope":"lesen"}`,
		"token-anlegen", profilID))

	darfSchreiben := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Setzer","scope":"schreiben"}`,
		"token-anlegen", profilID))

	an := starteDienst(t, umgebung)

	// **Ein Exemplar, an dem sich Ändern und Löschen prüfen lassen.**
	angelegt := mussREST(t, an.mitNachweis(darfSchreiben), http.MethodPost,
		"/sammlung", `{"printing":"probe-druck","menge":2}`)
	kennung := kennungAusAntwort(t, angelegt)

	for name, fall := range map[string]struct {
		methode, weg, rumpf string
		erfolg              int
	}{
		"aufnehmen": {
			http.MethodPost, "/sammlung",
			`{"printing":"zweiter-druck","menge":1}`, http.StatusCreated,
		},
		"ändern": {
			http.MethodPut, "/sammlung/" + kennung,
			`{"printing":"probe-druck","menge":5,"fassung":1}`, http.StatusOK,
		},
		"herausnehmen": {
			http.MethodDelete, "/sammlung/" + kennung + "?fassung=1", "",
			http.StatusOK,
		},
	} {
		t.Run(name, func(t *testing.T) {
			status, antwort := an.mitNachweis(nurLesen).ruf(
				t, fall.methode, fall.weg, fall.rumpf)

			if status != http.StatusForbidden {
				t.Fatalf("Status %d, erwartet 403 — %s", status, antwort)
			}

			// **Die Abweisung sagt, woran es lag** — wer ein gültiges Token
			// hat, ist kein Fremder und suchte den Fehler sonst bei sich.
			if !strings.Contains(antwort, "erlaubnis") {
				t.Errorf("die Meldung nennt die Klasse nicht: %s", antwort)
			}
		})
	}

	// **Und danach steht der Bestand unverändert da** — die Zusage hinter der
	// Zusage: ein 403 ist wertlos, wenn vorher schon geschrieben wurde.
	nachher := mussREST(t, an.mitNachweis(nurLesen), http.MethodGet,
		"/sammlung", "")

	if strings.Count(nachher, `"printing"`) != 1 ||
		!strings.Contains(nachher, `"menge":2`) {
		t.Errorf("der Bestand hat sich geändert: %s", nachher)
	}

	// **Die Gegenprobe**: mit dem schreibenden Token geht derselbe Weg.
	mussREST(t, an.mitNachweis(darfSchreiben), http.MethodPost, "/sammlung",
		`{"printing":"zweiter-druck","menge":1}`)
}

// **Auch ein Weg nach draußen wird an der Erlaubnis abgewiesen — nicht am
// Dienstschalter.**
//
// # Warum das eine eigene Zusage ist
//
// Der Schalter für den Netzweg steht auf **aus** (§16, Regel 6), und ein
// Aufruf, der daran scheitert, sieht aus wie eine Abweisung. Käme sie **vor**
// der Erlaubnisprüfung, erführe ein Rufender, der gar nichts darf, wie dieser
// Rechner eingestellt ist — und suchte den Fehler an der falschen Stelle.
func TestEinLesetokenScheitertAnDerErlaubnisUndNichtAmSchalter(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Matrixlauf", "--json"))

	nurLesen := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Leser","scope":"lesen"}`,
		"token-anlegen", profilID))

	an := starteDienst(t, umgebung)

	status, antwort := an.mitNachweis(nurLesen).ruf(t, http.MethodPost,
		"/drucke/probe-druck/preise", "")

	if status != http.StatusForbidden {
		t.Fatalf("Status %d, erwartet 403 — %s", status, antwort)
	}

	if !strings.Contains(antwort, "erlaubnis") {
		t.Errorf("die Meldung nennt die Klasse nicht: %s", antwort)
	}
}

// kennungAusAntwort liest die Kennung aus einer Anlegeantwort.
//
// **Eine eigene und nicht [ausRest]**: jene gibt nur Zeichenketten, und eine
// Kennung kommt als Zahl herein.
func kennungAusAntwort(t *testing.T, rumpf string) string {
	t.Helper()

	var gelesen struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal([]byte(rumpf), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, rumpf)
	}

	if gelesen.ID == 0 {
		t.Fatalf("die Antwort nennt keine Kennung: %s", rumpf)
	}

	return strconv.FormatInt(gelesen.ID, 10)
}

// **Und kein einziger schreibender Weg des Dienstes nimmt ein Lesetoken an.**
//
// # Warum das der Lauf ist, den die Zusage braucht
//
// Befund 7 des Zweitblicks über Phase 5.2: der Lauf darüber fährt drei
// Sammlungswege und eine Preisablehnung, und `docs/HAERTUNG.md` behauptete
// daraufhin die Durchsetzung „an jedem schreibenden Weg". Ein Kopf, der mehr
// verspricht als sein Rumpf hält, ist schlimmer als gar keiner (§10).
//
// **Die Liste kommt aus dem Quelltext des Dienstes**, nicht aus einer Kopie
// hier: `weg.HandleFunc("METHODE /pfad", …)` ist dieselbe Stelle, die auch der
// Wächter über die REST-Spalte liest. Ein 125. Weg, den jemand einträgt, steht
// damit sofort in diesem Lauf.
//
// # Was er zusagt — und was seine Gegenprobe wirklich zeigt
//
// **Zugesagt ist zweierlei**: kein 2xx, und **nichts hat sich geändert**. Das
// zweite ist die eigentliche Frage — ein Weg, der erst schreibt und danach
// 409 meldet, bliebe ohne sie grün.
//
// **Die Gegenprobe bittet um Genauigkeit.** Hier stand, sie bleibe grün. Das
// ist falsch, und der Nachlauf zum Zweitblick über Phase 5.2 hat es
// nachgemessen: wer **beide** wirksamen Tokenprüfungen umgeht — die Reichweite
// in `domain.Akteur.Darf` **und** den Scopevergleich in
// `service.Tokenverwaltung.Pruefe` —, bekommt auf `DELETE /token/{tokenID}`
// ein 200, und dieser Lauf wird rot.
//
// **Wer nur eine von beiden entschärft, sieht nichts** — und das ist keine
// Nachlässigkeit des Laufs, sondern eine Aussage über die Anwendung: die
// Reichweite eines Tokens wird an **zwei** Stellen geprüft, und eine allein
// genügt. Für die Gegenprobe heißt es: sie muss beide anfassen, sonst misst
// sie nichts (§10).
//
// **Und die Erlaubnis ist der Grund** — 403 und `erlaubnis` an jedem Weg,
// gemessen am 2026-10-10. Hier stand, die meisten Wege scheiterten schon an
// Rumpf oder Kennung; gemessen war das nicht, und verdeckt hatte es die
// Drossel (unten).
func TestKeinSchreibenderWegNimmtEinLesetokenAn(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Matrixlauf", "--json"))

	nurLesen := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Leser","scope":"lesen"}`,
		"token-anlegen", profilID))

	an := starteDienst(t, umgebung).mitNachweis(nurLesen)

	wege := schreibendeWege(t)
	if len(wege) < 100 {
		t.Fatalf("nur %d schreibende Wege gefunden — der Leser greift daneben",
			len(wege))
	}

	// **Das Lagebild vorher** — gegen das nachher geprüft wird.
	vorher := lagebild(t, an)

	for _, weg := range wege {
		rumpf := "{}"
		if weg.methode == http.MethodDelete {
			rumpf = ""
		}

		// **Die Drossel ist kein Nein des Tokens** (Zweitblick über Phase C,
		// Befund 3): ohne Warten antworteten 101 der Wege mit 429, und „ab
		// 400" war auch damit erfüllt — für sie belegte der Lauf nichts. Er
		// wartet deshalb, wie es die Antwort sagt, und verlangt danach **403
		// und `erlaubnis`** an jedem Weg: der Dienst prüft die Reichweite vor
		// Rumpf und Kennung.
		status, antwort := an.ruf(t, weg.methode, ausgefuellt(weg.pfad), rumpf)
		for status == http.StatusTooManyRequests {
			time.Sleep(1100 * time.Millisecond)
			status, antwort = an.ruf(t, weg.methode, ausgefuellt(weg.pfad), rumpf)
		}

		if status != http.StatusForbidden || !strings.Contains(antwort, `"code":"erlaubnis"`) {
			t.Errorf("%s %s nahm das Lesetoken an oder wies es anders ab: %d — %s",
				weg.methode, weg.pfad, status, antwort)
		}
	}

	// **Und nichts hat sich geändert.** Ein Weg, der erst schriebe und danach
	// 409 meldete, bliebe ohne diese Zeilen grün — angestrichen vom Nachlauf
	// zum Zweitblick über Phase 5.2.
	for weg, davor := range vorher {
		if danach := abgefragt(t, an, weg); danach != davor {
			t.Errorf("%s hat sich geändert:\nvorher  %s\nnachher %s",
				weg, davor, danach)
		}
	}
}

// lesewege sind die Auskünfte, an denen sich eine Wirkung zeigen würde.
//
// **Vier, nicht vierzig**: Sets, Sammlung, Profile und Token decken die
// Tafeln ab, die die schreibenden Wege dieser Fläche anfassen. Eine
// vollständige Abschrift des Bestands wäre ein zweiter Bestand.
var lesewege = []string{"/sets", "/sammlung", "/nutzer/1/profile", "/profile/1/token"}

// lagebild liest jede Auskunft einmal.
func lagebild(t *testing.T, an dienst) map[string]string {
	t.Helper()

	bild := make(map[string]string, len(lesewege))
	for _, weg := range lesewege {
		bild[weg] = abgefragt(t, an, weg)
	}

	return bild
}

// abgefragt gibt die Antwort einer Auskunft — samt ihrem Status.
//
// **Der Status gehört dazu**: eine Auskunft, die vorher 200 gab und nachher
// 404, hat sich geändert, auch wenn beide Rümpfe leer sind.
func abgefragt(t *testing.T, an dienst, weg string) string {
	t.Helper()

	status, antwort := an.ruf(t, http.MethodGet, weg, "")

	return strconv.Itoa(status) + " " + antwort
}

// dienstweg ist ein Eintrag aus der Wegetafel.
type dienstweg struct{ methode, pfad string }

// eingetragenerWeg trifft `weg.HandleFunc("POST /sets", …)` — auch über einen
// Zeilenumbruch hinweg, wie der Wächter über die REST-Spalte.
var eingetragenerWeg = regexp.MustCompile(`weg\.HandleFunc\(\s*"([A-Z]+) ([^"]+)"`)

// schreibendeWege liest jeden Weg des Dienstes, der etwas ändern könnte.
func schreibendeWege(t *testing.T) []dienstweg {
	t.Helper()

	wege := []dienstweg{}

	ordner := filepath.Join("..", "rest")

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("internal/rest lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".go") ||
			strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		inhalt, lesefehler := os.ReadFile(filepath.Join(ordner, eintrag.Name())) //nolint:gosec // Pfad aus dem Baum selbst
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		for _, treffer := range eingetragenerWeg.FindAllStringSubmatch(string(inhalt), -1) {
			if treffer[1] == http.MethodGet {
				continue
			}

			wege = append(wege, dienstweg{methode: treffer[1], pfad: treffer[2]})
		}
	}

	return wege
}

// platzhalter trifft `{setID}` und `{id...}`.
var platzhalter = regexp.MustCompile(`\{[^}]+\}`)

// ausgefuellt setzt für jeden Platzhalter eine Eins ein.
//
// **Eine Kennung, die es geben könnte** — und meistens nicht gibt. Das ist
// gewollt: geprüft wird die Erlaubnis, nicht der Datensatz.
func ausgefuellt(pfad string) string {
	return platzhalter.ReplaceAllString(pfad, "1")
}
