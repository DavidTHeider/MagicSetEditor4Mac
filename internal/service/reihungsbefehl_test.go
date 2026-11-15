package service_test

// **Die Reihung ausgelöster Fähigkeiten über den Sandkasten** — CR 603.3b,
// Werkstück M9
// (#1611).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Befehle, die eine Gestalt ändern** — Spielstein,
// Emblem, Kontrolle, Umwandeln, Kopieren. Hier steht der eine, der eine
// **offene Frage beantwortet**, und `gestaltbefehle_test.go` stand mit ihm
// über der Zeilengrenze aus §2.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Befehl legt die genannte Fähigkeit als Nächste auf den Stapel** —
// CR 603.3b.
func TestDerReihungsbefehlLegtDieGenannteZuerst(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einReihungsstand(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlReihung,
		Nutzlast: map[string]any{"kennung": "sitz-2/strahlung/0"},
	})
	if fehler != nil {
		t.Fatalf("reihung: %v", fehler)
	}

	// **Beide liegen, und die Frage ist beantwortet** — der Kern legt die
	// zweite von selbst nach, sobald keine Wahl mehr bleibt.
	if danach.Ausloeserreihung != "" {
		t.Errorf("es wird noch von %q geordnet", danach.Ausloeserreihung)
	}

	if anzahl := len(danach.Ausgeloest); anzahl != 0 {
		t.Errorf("%d warten noch", anzahl)
	}

	// **Die genannte liegt unten** (CR 405.2) und wird zuletzt verrechnet.
	oben := danach.StapelOben()
	if oben == nil {
		t.Fatal("auf dem Stapel liegt nichts")
	}

	if oben.Name != string(domain.AusloeserMonarchzug) {
		t.Errorf("oben liegt %q", oben.Name)
	}
}

// **Eine Nutzlast ohne Kennung wird abgewiesen** — dieselbe Zusage wie bei
// jedem Pflichtfeld.
func TestDerReihungsbefehlBrauchtEineKennung(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(einReihungsstand(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlReihung, Nutzlast: map[string]any{},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Kennung ging es durch: %v", fehler)
	}
}

// **Und wer nicht gefragt ist, kommt nicht durch** — CR 603.3b nennt den
// Beherrscher.
//
// **Die Gegenprobe zum Lauf darüber** (§10): ohne sie sagte der nur, dass
// ein Befehl **wirkt** — nicht, dass er dem Falschen verweigert wird.
func TestEineFremdeReihungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(einReihungsstand(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlReihung,
		Nutzlast: map[string]any{"kennung": "sitz-2/strahlung/0"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("sitz-1 durfte ordnen: %v", fehler)
	}

	// **Und der Satz nennt die Regel** — wer ihn liest, soll nachlesen
	// können, warum er nicht darf (§11).
	if !strings.Contains(fehler.Error(), "603.3b") {
		t.Errorf("der Satz nennt die Regel nicht: %v", fehler)
	}
}

// einReihungsstand gibt eine Partie, in der sitz-2 zwei Fähigkeiten ordnen
// muss.
//
// **Von Hand gestellt und nicht erspielt**: dass die vier eingebauten
// wirklich auslösen, misst `internal/domain` — hier geht es um den **Weg**
// vom Befehl in die Domäne.
//
// **Sitz 2 und nicht Sitz 1**: Sitz 1 ist am Zug, und eine Antwort, die nur
// dem aktiven Spieler offenstünde, sagte über CR 603.3 nichts aus.
func einReihungsstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)
	partie.Ausgeloest = []domain.Ausloeser{
		{
			Art: domain.AusloeserStrahlung, Beherrscher: "sitz-2",
			Kennung: "sitz-2/strahlung/0",
		},
		{
			Art: domain.AusloeserMonarchzug, Beherrscher: "sitz-2",
			Kennung: "sitz-2/monarch.zug/0",
		},
	}
	partie.Ausloeserreihung = "sitz-2"

	return partie
}
