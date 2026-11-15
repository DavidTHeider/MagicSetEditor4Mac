package domain_test

// **Was eine Kontrollnachricht meint** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass „Kontrolle" in beide Richtungen nicht dasselbe heißt.** Vom Gast zum
// Wirt ist sie ein Wunsch, vom Wirt zum Gast das Tischbild. Ohne die Art
// müsste jede Seite am Inhalt raten — und die erste neue Form ginge still als
// die alte durch (§8).

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Es gibt genau sechs Arten, und alle bestehen ihre Prüfung.**
//
// **Zwei je Richtung**: der Gast schickt einen **Wunsch** und eine
// **Rückkehr**, der Wirt ein **Tischbild** und einen **Abzug**. **Und zwei in
// beide** (P9, #2377): wer ein Bild nicht zeichnen kann, schickt einen
// **Bildwunsch** an den, dem das Deck gehört, und bekommt ein **Bild**.
func TestJedeKontrollartBestehtIhrePruefung(t *testing.T) {
	t.Parallel()

	arten := domain.Kontrollarten()
	if len(arten) != 6 {
		t.Fatalf("%d Arten, erwartet 6", len(arten))
	}

	for _, art := range arten {
		if fehler := art.Pruefe(); fehler != nil {
			t.Errorf("%q: %v", art, fehler)
		}
	}
}

// **Eine erfundene Art wird abgewiesen — und die Meldung nennt die echten.**
func TestEineErfundeneKontrollartWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := domain.Kontrollart("unfug").Pruefe()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if !containsAlle(fehler.Error(), "wunsch", "tischbild") {
		t.Errorf("die Meldung nennt die echten nicht: %v", fehler)
	}
}

// **Verpacken gibt Bytes, die sich wieder lesen lassen.**
func TestVerpackenUndLesenGehenZusammen(t *testing.T) {
	t.Parallel()

	roh, fehler := domain.VerpackeKontrolle(
		domain.KontrolleWunsch, map[string]string{"art": "objekt.tappen"})
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	// **Der Inhalt steht lesbar da und nicht als Base64** — dasselbe wie bei
	// der Nachricht darum herum.
	if !containsAlle(string(roh), `"art":"wunsch"`, `"objekt.tappen"`) {
		t.Errorf("die Gestalt: %s", roh)
	}

	art, inhalt, fehler := domain.LiesKontrolle(
		roh, "gast", domain.KontrolleWunsch)
	if fehler != nil {
		t.Fatalf("LiesKontrolle: %v", fehler)
	}

	if art != domain.KontrolleWunsch {
		t.Errorf("die Art kam als %q zurück", art)
	}

	if !containsAlle(string(inhalt), `"objekt.tappen"`) {
		t.Errorf("der Inhalt: %s", inhalt)
	}
}

// **Wer die falsche Art bekommt, bekommt keinen Formfehler, sondern eine
// Absage.**
//
// **Ein Wunsch an einen Gast oder ein Tischbild an den Wirt ist eine
// Anmaßung**: der eine hat keinen Vollzustand, der andere rechnet selbst. Die
// Meldung nennt beide Arten, damit der Leser sieht, was er statt dessen
// bekommen hat.
func TestEineUnerwarteteArtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	roh, fehler := domain.VerpackeKontrolle(
		domain.KontrolleTischbild, map[string]string{"partie": "p1"})
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	_, _, fehler = domain.LiesKontrolle(roh, "wirt", domain.KontrolleWunsch)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if !containsAlle(fehler.Error(), "tischbild", "wunsch", "wirt") {
		t.Errorf("die Meldung sagt zu wenig: %v", fehler)
	}
}

// **Was keine Kontrolle ist, kommt nicht durch** (§8).
func TestWasKeineKontrolleIstKommtNichtDurch(t *testing.T) {
	t.Parallel()

	if _, _, fehler := domain.LiesKontrolle(
		[]byte("{kein json"), "gast", domain.KontrolleWunsch,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und eine Kontrolle, die ihre Gestalt nicht hält, auch nicht.**
func TestEineKrummeKontrolleKommtNichtDurch(t *testing.T) {
	t.Parallel()

	roh, fehler := json.Marshal(domain.Kontrolle{
		Art: domain.Kontrollart("unfug"), Inhalt: []byte(`{"a":1}`),
	})
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	if _, _, fehler := domain.LiesKontrolle(
		roh, "gast", domain.KontrolleWunsch,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine erfundene Art wird schon beim Verpacken abgewiesen.**
//
// **Vor dem Verpacken und nicht danach** (§8): was die Gestalt nicht einhält,
// entsteht gar nicht erst — sonst stünde die Meldung beim Empfänger, und der
// kann nichts dafür.
func TestEineErfundeneArtWirdNichtVerpackt(t *testing.T) {
	t.Parallel()

	if _, fehler := domain.VerpackeKontrolle(
		domain.Kontrollart("unfug"), map[string]string{},
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Was sich nicht verpacken lässt, wird gemeldet.**
func TestEineUnverpackbareNutzlastWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, fehler := domain.VerpackeKontrolle(
		domain.KontrolleWunsch, map[string]any{"k": make(chan int)},
	); fehler == nil {
		t.Error("die Nutzlast kam durch")
	}
}

// **Eine Kontrolle ohne Inhalt sagt nichts.**
func TestEineKontrolleOhneInhaltWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Kontrolle{Art: domain.KontrolleWunsch}).Pruefe(); !errors.Is(
		fehler, domain.ErrValidation,
	) {
		t.Errorf("es kam %v", fehler)
	}
}

// containsAlle sagt, ob jeder Teil vorkommt.
func containsAlle(text string, teile ...string) bool {
	for _, teil := range teile {
		if !strings.Contains(text, teil) {
			return false
		}
	}

	return true
}

// **Und eine Kontrolle mit erfundener Art auch dann, wenn sie Inhalt hat.**
//
// **Die Gegenprobe zum Lauf darüber** (§10): dort fehlt der Inhalt, hier die
// Art. Ein Lauf, der nur einen von beiden prüfte, ließe offen, ob die Prüfung
// überhaupt zweimal hinsieht.
func TestEineKontrolleMitErfundenerArtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kontrolle := domain.Kontrolle{
		Art: domain.Kontrollart("unfug"), Inhalt: []byte(`{"a":1}`),
	}

	if fehler := kontrolle.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}
