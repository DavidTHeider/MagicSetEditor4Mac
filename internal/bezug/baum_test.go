package bezug_test

// **Der ganze Baum eines Commits** — G04
// (#2251).
//
// Der Vergleich nennt höchstens 300 Dateien. Wer mehr wissen will, fragt
// nach den Bäumen beider Commits und vergleicht sie selbst — Pfad und
// Kennung jeder Datei.

import (
	"net/http"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

func TestDerBaumNenntJedeDatei(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{"sha":"x","truncated":false,"tree":[
		{"path":"data","type":"tree","sha":"t1"},
		{"path":"data/a.mse-style/style","type":"blob","sha":"b1"},
		{"path":"data/modul","type":"commit","sha":"c1"},
		{"path":"README.md","type":"blob","sha":"b2"}]}`)

	baum, fehler := bezug.Baum(t.Context(), dienst.URL, "probe")
	if fehler != nil {
		t.Fatalf("Baum: %v", fehler)
	}

	// **Nur Dateien**: ein Verzeichnis ändert sich mit seinen Dateien, und
	// ein eingebundenes Repositorium bringt der Bezug nicht mit.
	if baum.Gekuerzt || len(baum.Dateien) != 2 || baum.Dateien["data/a.mse-style/style"] != "b1" || baum.Dateien["README.md"] != "b2" {
		t.Errorf("Baum %+v", baum)
	}
}

func TestEinGekuerzterBaumSagtEs(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{"truncated":true,"tree":[{"path":"a","type":"blob","sha":"b"}]}`)

	baum, fehler := bezug.Baum(t.Context(), dienst.URL, "probe")
	if fehler != nil || !baum.Gekuerzt {
		t.Errorf("Baum gekürzt=%v, Fehler %v", baum.Gekuerzt, fehler)
	}
}

func TestEinBaumOhneBaumBrichtLaut(t *testing.T) {
	for name, rumpf := range map[string]string{
		"ohne tree":       `{"truncated":false}`,
		"keine Gestalt":   `[]`,
		"Datei ohne Pfad": `{"tree":[{"path":"","type":"blob","sha":"b"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := bezug.Baum(t.Context(), antwortender(t, http.StatusOK, rumpf).URL, "probe"); fehler == nil {
				t.Error("eine Antwort ohne Baum ging durch")
			}
		})
	}
}

func TestEinBaumScheitertWieDieSpitze(t *testing.T) {
	if _, fehler := bezug.Baum(t.Context(), antwortender(t, http.StatusNotFound, `{}`).URL, "probe"); fehler == nil {
		t.Error("ein 404 ging durch")
	}
}

func TestZweiBaeumeErgebenDieAenderungen(t *testing.T) {
	alt := bezug.Baumstand{Dateien: map[string]string{
		"data/a.mse-style/style": "1", "data/a.mse-style/alt.png": "2", "README.md": "3",
	}}
	neu := bezug.Baumstand{Dateien: map[string]string{
		"data/a.mse-style/style": "1x", "data/b.mse-style/style": "4", "README.md": "3",
	}}

	aenderungen := bezug.Baumvergleich(alt, neu)

	// **Geordnet nach Pfad**, damit zwei Vorschauen derselben Bäume gleich
	// aussehen; eine unveränderte Datei steht nicht da.
	if !slices.Equal(aenderungen, []bezug.Aenderung{
		{Datei: "data/a.mse-style/alt.png", Art: "removed"},
		{Datei: "data/a.mse-style/style", Art: "modified"},
		{Datei: "data/b.mse-style/style", Art: "added"},
	}) {
		t.Errorf("Änderungen %+v", aenderungen)
	}
}
