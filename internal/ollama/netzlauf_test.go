//go:build netzlauf

// **Der Lauf gegen eine echte Ollama-Installation** — Werkstück A7 von
// Phase 5.3 (#1446).
//
// # Was er prüft, das kein anderer Lauf kann
//
// **Dass die Laufzeit die Gestalt schickt, gegen die dieses Paket gebaut ist.**
// Die Läufe daneben gehen gegen eingecheckte, selbst geschriebene Beispiele
// (§3); sie halten, dass dieses Programm die **dokumentierte** Gestalt liest,
// und können nicht sagen, ob Ollama sie noch schickt.
//
// **Und dass ein Einbettungsvertrag wirklich zu messen ist.** Das Issue
// verlangt als Abnahme einen „echten lokalen Embeddinglauf mit gepinnter
// Ollama-/Modellkonfiguration, offline mit Null-Egress". Ohne eine
// Installation ist das nicht zu haben — und dass es hier nicht gelaufen ist,
// steht in [0098](../../docs/entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)
// und in `docs/HAERTUNG.md`.
//
// # Warum er nicht in der Pipeline läuft
//
// **Er braucht eine Installation, die es hier nicht gibt.** Auf dem Rechner,
// auf dem A7 gebaut wurde, ist Ollama nicht installiert — der Lauf wäre in
// jeder Pipeline rot und sagte dabei nichts über den Quelltext.
//
// # Wie er gefahren wird
//
//	ollama serve
//	ollama pull gemma3:4b
//	ollama pull embeddinggemma
//	go test -tags netzlauf ./internal/ollama/ -v
//
// **Die Modelle werden von Hand geladen und nicht von hier** — das ist der
// Punkt der Positivliste, und ein Lauf, der sie unterliefe, prüfte das
// Gegenteil dessen, was zugesagt ist.
package ollama_test

import (
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// echterKlient baut ihn gegen die laufende Installation.
func echterKlient(t *testing.T) *ollama.Klient {
	t.Helper()

	return ollama.NeuerKlient(os.Getenv("MSE_OLLAMA"),
		netz.NeuerKlient(netz.UeblicheFristen()))
}

// **Die Laufzeit nennt ihre Fassung.**
func TestDieEchteLaufzeitNenntIhreFassung(t *testing.T) {
	fassung, fehler := echterKlient(t).Fassung(t.Context())
	if fehler != nil {
		t.Fatalf("Fassung: %v — läuft `ollama serve`?", fehler)
	}

	if fassung == "" {
		t.Error("die Fassung ist leer")
	}

	t.Logf("Ollama %s", fassung)
}

// **Die installierten Modelle kommen mit Herkunft und Fähigkeiten.**
func TestDieEchtenModelleKommenAn(t *testing.T) {
	modelle, befunde, fehler := echterKlient(t).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) == 0 {
		t.Fatal("es ist kein Modell installiert — `ollama pull gemma3:4b`")
	}

	for _, modell := range modelle {
		if fehler := modell.Validate(); fehler != nil {
			t.Errorf("%q ist ungültig: %v", modell.Alias, fehler)
		}

		// **Die Herkunft ist nie unbekannt** — `/api/tags` nennt den Digest.
		if modell.Herkunft.Unbekannt() {
			t.Errorf("%q kommt ohne Herkunft", modell.Alias)
		}

		t.Logf("%s/%s (%s) — %d belegte Fähigkeiten",
			modell.Anbieter, modell.Alias, modell.Ort, len(modell.Geprueft))
	}

	for _, befund := range befunde {
		t.Logf("Befund: %s", befund.Meldung)
	}
}

// **Ein Textmodell antwortet in der Form, die das Schema vorgibt.**
func TestDasEchteModellHaeltDieForm(t *testing.T) {
	alias := modellMit(t, domain.KIStrukturierteGenerierung)

	antwort, fehler := echterKlient(t).Rede(t.Context(), ollama.Frage{
		Alias:   alias,
		Auftrag: "Nenne den Namen der Karte »Blitzschlag«.",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"name": map[string]any{"type": "string"}},
			"required":   []string{"name"},
		},
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Abgeschnitten() {
		t.Fatalf("die Antwort brach ab: %+v", antwort)
	}

	if !strings.HasPrefix(strings.TrimSpace(antwort.Inhalt), "{") {
		t.Errorf("die Antwort ist kein JSON-Objekt: %q", antwort.Inhalt)
	}

	t.Logf("%s antwortete: %s", antwort.Modell, antwort.Inhalt)
}

// **Und ein Einbetter lässt sich vermessen.**
//
// **Das ist die Abnahme aus der Ergänzung vom 2026-09-10**: „echter lokaler
// Embeddinglauf mit gepinnter Ollama-/Modellkonfiguration".
func TestDerEchteEinbetterLaesstSichVermessen(t *testing.T) {
	alias := einbetterAus(t)

	klient := echterKlient(t)

	vertrag, befunde, fehler := klient.Vermessen(t.Context(), alias)
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	for _, befund := range befunde {
		t.Logf("Befund: %s", befund.Meldung)
	}

	if len(befunde) > 0 {
		t.Skipf("%q trägt keinen Vertrag — der Grund steht darüber", alias)
	}

	t.Logf("Vertrag: Dimension %d, %s, normalisiert %t, Zerteiler %q, "+
		"Fingerabdruck %s", vertrag.Dimension, vertrag.Pooling,
		vertrag.Normalisiert, vertrag.Zerteiler, vertrag.Fingerabdruck())

	// **Derselbe Text gibt denselben Vektor** — sonst zeigte ein Index
	// morgen woandershin.
	einbetter := klient.Einbetter(domain.Vektorraum{Modell: alias, Vertrag: vertrag})

	erste, fehler := einbetter.Einbetten(t.Context(), []string{"Blitzschlag"})
	if fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	zweite, fehler := einbetter.Einbetten(t.Context(), []string{"Blitzschlag"})
	if fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	for i := range erste[0] {
		if erste[0][i] != zweite[0][i] {
			t.Fatalf("zweimal derselbe Text gab zwei verschiedene Vektoren "+
				"(Stelle %d: %v gegen %v)", i, erste[0][i], zweite[0][i])
		}
	}
}

// **Ein zu langer Text wird abgewiesen und nicht gekürzt.**
func TestEinZuLangerTextWirdEchtAbgewiesen(t *testing.T) {
	alias := einbetterAus(t)

	klient := echterKlient(t)

	vertrag, befunde, fehler := klient.Vermessen(t.Context(), alias)
	if fehler != nil || len(befunde) > 0 {
		t.Skipf("%q trägt keinen Vertrag", alias)
	}

	// **Weit über jedes Kontextfenster** — 200.000 Wörter.
	lang := strings.Repeat("Blitzschlag ", 200_000)

	if _, fehler := klient.Einbetter(domain.Vektorraum{Modell: alias, Vertrag: vertrag}).
		Einbetten(t.Context(), []string{lang}); fehler == nil {
		t.Error("der zu lange Text ging durch — wurde er still gekürzt?")
	}
}

// modellMit sucht ein installiertes Modell mit dieser Fähigkeit.
func modellMit(t *testing.T, faehigkeit domain.KIFaehigkeit) string {
	t.Helper()

	modelle, _, fehler := echterKlient(t).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, modell := range modelle {
		if modell.Kann(faehigkeit) {
			return modell.Alias
		}
	}

	t.Skipf("kein installiertes Modell belegt %q", string(faehigkeit))

	return ""
}

// einbetterAus sucht ein installiertes Einbettungsmodell.
//
// **Über `/api/show` und nicht über den Katalog**: der Einbetter steht dort
// **ohne** Fähigkeit, solange ihn niemand vermessen hat — und genau das
// Vermessen ist der Lauf darüber.
func einbetterAus(t *testing.T) string {
	t.Helper()

	modelle, befunde, fehler := echterKlient(t).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Einbettung") {
			return befund.Datei
		}
	}

	if len(modelle) == 0 {
		t.Fatal("es ist kein Modell installiert")
	}

	t.Skip("kein installiertes Modell meldet die Einbettung — " +
		"`ollama pull embeddinggemma`")

	return ""
}
