// Der Wächter über die gebundene Oberfläche.
//
// **Er hält die Naht, die kein anderer Test betritt.** Die Go-Tests rufen `App`
// unmittelbar, die Frontend-Tests legen eine eigene Fassade unter
// `globalThis.go`; dass Wails dazwischen etwas anderes verlangt, sah keiner von
// beiden. Der Preis war eine Anwendung, deren Oberfläche nie eine Antwort bekam.
package app

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// **Jede gebundene Methode nimmt den Kontext ihres Rufs — vorn, und nur
// dort.**
//
// Unter Wails 2 reichte Wails keinen hinein: `ParseArgs` verlangte genau so
// viele Argumente, wie die Signatur führt, und eine Methode mit Kontext
// antwortete dem Frontend **gar nicht** — deshalb hielt dieser Wächter bis
// BK-01 das Gegenteil. Wails 3 füllt einen Kontext an **erster** Stelle mit
// dem des Rufs (`needsContext`, `bindings.go` in v3.0.0-beta.23) und bricht
// ihn ab, wenn die Oberfläche den Ruf aufgibt oder das Fenster sich schließt
// (#2109, Vermerk 0170). An jeder anderen Stelle bliebe es ein Argument, das
// das Frontend schicken müsste und nicht kann.
//
// **Was er nicht hält**: dass der Kontext auch weitergegeben wird. Das hält
// der Übersetzer — `App.kontext` gibt es nicht mehr — und
// `rufabbruch_intern_test.go` an einem Ruf, der wartet.
func TestJedeGebundeneMethodeNimmtDenKontextIhresRufs(t *testing.T) {
	kontexttyp := reflect.TypeFor[context.Context]()
	typ := reflect.TypeFor[*Bindungen]()

	if typ.NumMethod() < 400 {
		t.Fatalf("nur %d gebundene Methoden — der Wächter greift daneben", typ.NumMethod())
	}

	for i := range typ.NumMethod() {
		methode := typ.Method(i)

		t.Run(methode.Name, func(t *testing.T) {
			if methode.Type.NumIn() < 2 || methode.Type.In(1) != kontexttyp {
				t.Errorf("%s nimmt keinen context.Context an erster Stelle — "+
					"ein Abbruch des Fensters erreichte sie nicht", methode.Name)
			}

			for arg := 2; arg < methode.Type.NumIn(); arg++ {
				if methode.Type.In(arg) == kontexttyp {
					t.Errorf("%s nimmt einen weiteren context.Context — "+
						"das Frontend kann keinen schicken", methode.Name)
				}
			}
		})
	}
}

// **Und die Hülle bindet wirklich diese Fassade** — seit ENG-02 in
// `internal/fenster/wails.go` und nicht mehr in `main.go`.
//
// Der erste Anlauf spiegelte nur den Typ `Bindungen` und verglich seine
// Methoden. Wer wieder `anwendung` bindet, bekäme davon kein rotes Licht — und
// stünde wieder da, wo Phase 5 stand: mit `Starten`, `Beenden` und sieben
// `Mit…`-Nähten im Fenster. Der Nachlauf des Zweitblicks hat es angestrichen.
func TestDieHuelleBindetDieFassade(t *testing.T) {
	quelle, err := os.ReadFile("../fenster/wails.go")
	if err != nil {
		t.Fatalf("internal/fenster/wails.go nicht lesbar: %v", err)
	}

	inhalt := ohneKommentare(string(quelle))

	if !strings.Contains(inhalt, "application.NewService(app.NeueBindungen(") {
		t.Error("die Hülle bindet nicht app.NeueBindungen(…) — steht die Anwendung selbst im Fenster?")
	}

	if strings.Contains(inhalt, "application.NewService(anwendung)") {
		t.Error("die Hülle bindet die Anwendung unmittelbar — damit stehen ihre Nähte im Fenster")
	}

	pruefeDienstliste(t, inhalt)
}

// dienstliste fängt den Inhalt von `Services: []application.Service{…}`.
var dienstliste = regexp.MustCompile(`Services:\s*\[\]application\.Service\{([^}]*)\}`)

// pruefeDienstliste hält die Liste der Dienste auf **genau einen** Eintrag.
//
// # Warum das eine eigene Zusage ist
//
// Wails nimmt die exportierten Methoden **jedes** Dienstes auf (`bindings.go`,
// v3.0.0-beta.23; unter Wails 2 jedes Eintrags in `Bind`). Die Zeile darüber
// sucht **irgendwo** nach `app.NeueBindungen(` und verbietet nur die genaue
// Gestalt mit `anwendung`; eine Liste mit der Fassade **und** der Anwendung
// liefe an ihr vorbei — und hätte `Starten`, `Beenden` und die `Mit…`-Nähte
// wieder ins Fenster gestellt. Angestrichen vom Nachlauf über Phase 4.0
// (Befund 3), damals an `Bind: []any{…}`.
//
// **Was er nicht hält**: was der eine Eintrag *ist*. Dass er
// `app.NeueBindungen(…)` heißt, prüft die Zeile darüber; dass sein Typ nichts
// einbettet, prüft `bindungszaehler_test.go` im Werkzeugpaket.
func pruefeDienstliste(t *testing.T, inhalt string) {
	t.Helper()

	fund := dienstliste.FindStringSubmatch(inhalt)
	if fund == nil {
		t.Fatal("die Hülle hat keine Dienstliste `Services: []application.Service{…}` — der Leser greift daneben")
	}

	if dienste := strings.Count(fund[1], "application.NewService("); dienste != 1 {
		t.Errorf("die Hülle bindet %d Dienste — Wails nimmt die exportierten Methoden "+
			"**jedes** davon auf. Es darf genau einer sein: %q", dienste, fund[1])
	}
}

// ohneKommentare wirft Zeilenkommentare weg.
//
// **Damit ein auskommentierter Aufruf nicht als Beleg durchgeht.** Der
// Fassadenwächter suchte mit `strings.Contains`, und ein Name in einem
// Kommentar erfüllte ihn — auch das ein Befund des Nachlaufs.
func ohneKommentare(quelle string) string {
	var gebaut strings.Builder

	for zeile := range strings.SplitSeq(quelle, "\n") {
		if schnitt := strings.Index(zeile, "//"); schnitt >= 0 {
			zeile = zeile[:schnitt]
		}

		gebaut.WriteString(zeile)
		gebaut.WriteByte('\n')
	}

	return gebaut.String()
}
