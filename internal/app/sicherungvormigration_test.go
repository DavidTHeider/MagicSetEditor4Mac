package app

// Die Entscheidung vor der Migration — sichern, weitermachen, abbrechen.
//
// **Paketintern**, weil `sicherungVorMigration` nichts ist, was jemand von
// außen ruft: sie wird im Composition Root verdrahtet. Geprüft wird die
// Entscheidung, nicht `VACUUM INTO` — das prüft `internal/storage`.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Name nennt die Fassung, auf die migriert wird.**
func TestDerSicherungsnameNenntDieFassung(t *testing.T) {
	if name := Sicherungsname("0.2.0"); name != "bestand-vor-0.2.0.db" {
		t.Errorf("der Name ist %q", name)
	}
}

// **Gesichert wird an den Ort, den die Anwendung nennt.**
func TestDieSicherungGehtNebenDenBestand(t *testing.T) {
	var gesehen string

	fehler := sicherungVorMigration(func(_ context.Context, ziel string) error {
		gesehen = ziel

		return nil
	}, taugt, marke("gleich"), marke("gleich"), "/pfad/zum/ort", "0.2.0")(context.Background())
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if gesehen != "/pfad/zum/ort/bestand-vor-0.2.0.db" {
		t.Errorf("gesichert wurde nach %q", gesehen)
	}
}

// taugt ist die Prüfung, die alles durchwinkt.
func taugt(context.Context, string) error { return nil }

// liegtDa ist ein Sicherer, dessen Ziel schon dasteht.
func liegtDa(context.Context, string) error { return konflikt() }

// konflikt ist der Fehler, den `storage.Sichern` bei vorhandenem Ziel gibt.
func konflikt() error {
	return fmt.Errorf("%w: bestand-vor-0.2.0.db liegt schon da", domain.ErrConflict)
}

// **Eine vorhandene, brauchbare Sicherung hält die Migration nicht auf.**
//
// „Liegt schon da" heißt hier etwas anderes als sonst: ein früherer Lauf hat
// sie angelegt, die Migration danach ist gescheitert, und weil sie **eine
// Transaktion** ist (#938),
// steht der Bestand unverändert da. Die vorhandene Sicherung ist damit genau
// die richtige — **wenn sie eine ist.**
func TestEineVorhandeneSicherungHaeltNichtAuf(t *testing.T) {
	if fehler := sicherungVorMigration(liegtDa, taugt, marke("gleich"), marke("gleich"), "/ort", "0.2.0")(context.Background()); fehler != nil {
		t.Errorf("die vorhandene Sicherung hielt auf: %v", fehler)
	}
}

// **Eine vorhandene Datei, die keine brauchbare Sicherung ist, hält auf.**
//
// Angestrichen vom Zweitblick über Phase S: bis dahin galt **jeder**
// Namenskonflikt als Erfolg. Wer eine leere, beschädigte oder fremde Datei
// namens `bestand-vor-<Fassung>.db` neben den Bestand legt, ließ die Migration
// ohne Sicherung laufen — und der Kommentar daneben behauptete, die vorhandene
// sei „genau die richtige", ohne es zu messen (§10).
func TestEineUnbrauchbareVorhandeneSicherungHaeltAuf(t *testing.T) {
	taugtNicht := func(context.Context, string) error {
		return fmt.Errorf("%w: die Datei ist kein Bestand", domain.ErrSicherung)
	}

	fehler := sicherungVorMigration(liegtDa, taugtNicht, marke("gleich"), marke("gleich"), "/ort", "0.2.0")(context.Background())
	if fehler == nil {
		t.Fatal("die unbrauchbare Datei wurde als Sicherung genommen")
	}

	// **Die Meldung sagt, was zu tun ist.** Eine Abweisung ohne Ausweg ist
	// eine Sackgasse: die Anwendung startet dann gar nicht mehr.
	if !strings.Contains(fehler.Error(), "bestand-vor-0.2.0.db") {
		t.Errorf("die Meldung nennt die Datei nicht: %v", fehler)
	}
}

// **Jeder andere Fehler hält auf.**
//
// Eine Migration ohne Sicherung ist genau der Fall, für den es die Sicherung
// gibt. Wer hier weiterliefe, hätte sie ebenso gut weglassen können.
func TestEinAndererFehlerHaeltAuf(t *testing.T) {
	kaputt := errors.New("die Platte ist voll")

	fehler := sicherungVorMigration(func(context.Context, string) error {
		return kaputt
	}, taugt, marke("gleich"), marke("gleich"), "/ort", "0.2.0")(context.Background())

	if !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler %v, erwartet war die volle Platte", fehler)
	}

	if !strings.Contains(fehler.Error(), "vor der Migration sichern") {
		t.Errorf("die Meldung sagt nicht, was scheiterte: %v", fehler)
	}
}

// **Die echte Prüfung liest eine echte Datei.**
//
// `sicherungTaugt` ist die Naht zwischen der Entscheidung hier und
// `storage.PruefeSicherung`. Die Prüfung selbst ist dort belegt; hier geht es
// darum, dass die Naht **ruft** — und nicht etwa immer `nil` gibt.
func TestSicherungTaugtWeistEineFremdeDateiAb(t *testing.T) {
	ort := t.TempDir()
	fremd := filepath.Join(ort, "keine-sicherung.db")

	if err := os.WriteFile(fremd, []byte("nur Text"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if sicherungTaugt(probekontext(t), fremd) == nil {
		t.Error("eine Textdatei ging als Sicherung durch")
	}
}
