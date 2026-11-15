// **Die Werkzeuge der Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// rufeSammlungswerkzeug sucht ein Werkzeug und ruft es.
func rufeSammlungswerkzeug(
	t *testing.T, sammlung mcp.Sammlungswerkstatt, preise mcp.Preiswerkstatt,
	name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand:  neueBestandsprobeWeit(),
		Sammlung: sammlung, Preise: preise,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// **Die Sammlung kommt als Liste.**
func TestDasWerkzeugSammlungGibtDieExemplare(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(), neuePreisprobe(),
		"lesen.sammlung", `{}`)
	if err != nil {
		t.Fatalf("lesen.sammlung: %v", err)
	}

	if !strings.Contains(antwort, "probe-druck") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Und die eines Drucks ebenso.**
func TestDasWerkzeugSammlungFuerFiltert(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(), neuePreisprobe(),
		"lesen.sammlungfuer", `{"printing":"gibtesnicht"}`)
	if err != nil {
		t.Fatalf("lesen.sammlungfuer: %v", err)
	}

	// **Eine leere Liste und kein `null`** (§3).
	if !strings.Contains(antwort, `"Exemplare": []`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Ein Exemplar wird angelegt und trägt seine Kennung zurück.**
func TestDasWerkzeugExemplarLegtAn(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	antwort, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplar", `{"printing":"neuer-druck","menge":2}`)
	if err != nil {
		t.Fatalf("schreiben.exemplar: %v", err)
	}

	if !strings.Contains(antwort, `"ID": 2`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Dasselbe zweimal ist ein Konflikt** — so streng wie der Bestand.
func TestDasWerkzeugExemplarWeistDasZweiteAb(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	_, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplar",
		`{"printing":"probe-druck","menge":1,"zustand":"nearMint","sprache":"en"}`)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Aber ein WUNSCH auf denselben Druck ist keiner.**
//
// Die fünfte Spalte der Eindeutigkeit ist `wunsch` (0057): wer eine gespielte
// Fassung hat und eine schöne will, hat beides. Ohne diesen Lauf wäre die
// Strenge der Vorrichtung daneben unbewacht — der Nachlauf über Phase 4.5 hat
// genau das angestrichen.
func TestDasWerkzeugExemplarNimmtDenWunschZumBesitz(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	antwort, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplar",
		`{"printing":"probe-druck","menge":1,"zustand":"nearMint",`+
			`"sprache":"en","wunsch":true}`)
	if err != nil {
		t.Fatalf("ein Wunsch auf einen besessenen Druck wurde abgewiesen: %v", err)
	}

	if !strings.Contains(antwort, `"ID": 2`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Und der Lagerort kommt an.**
func TestDasWerkzeugExemplarNimmtDenLagerort(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplar",
		`{"printing":"neuer-druck","menge":2,"lagerort":"Kiste 4"}`); err != nil {
		t.Fatalf("schreiben.exemplar: %v", err)
	}

	antwort, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"lesen.sammlung", `{}`)
	if err != nil {
		t.Fatalf("lesen.sammlung: %v", err)
	}

	if !strings.Contains(antwort, "Kiste 4") {
		t.Errorf("der Lagerort steht nicht in der Antwort: %s", antwort)
	}
}

// **Ändern und Löschen treffen die Zeile, die gemeint ist.**
func TestDieWerkzeugeAendernUndLoeschen(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	// **Fassung 1, wie sie aus `lesen.sammlung` käme.** Ohne sie meldet die
	// Vorrichtung einen Konflikt — so streng wie der Bestand (§3).
	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplaraendern",
		`{"id":1,"fassung":1,"printing":"probe-druck","menge":9}`); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	// **Und danach ist es Fassung 2.** Mit der alten zu löschen wäre ein
	// Konflikt — der Fall steht darunter.
	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplarloeschen", `{"id":1,"fassung":2}`); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplarloeschen",
		`{"id":1,"fassung":2}`); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der zweite Fehler ist %v — will ErrNotFound", err)
	}
}

// **Kaputte Argumente kommen als Fehler zurück, nicht als Antwort.**
func TestKaputteArgumenteDerSammlungMeldenSich(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.sammlungfuer":         `{"printing":`,
		"schreiben.exemplar":         `{"menge":`,
		"schreiben.exemplaraendern":  `{"id":`,
		"schreiben.exemplarloeschen": `{"id":`,
		"lesen.preise":               `{"printing":`,
		"lesen.preisverlauf":         `{"printing":`,
		"schreiben.preiseholen":      `{"printing":`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(),
				neuePreisprobe(), name, argumente); err == nil {
				t.Error("die kaputten Argumente kamen durch")
			}
		})
	}
}
