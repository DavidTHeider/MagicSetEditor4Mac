package app

// **Ohne Schreibrecht legt keine Symbolnaht ab** — C04
// (#2229).
//
// Die Kommandozeile schützt einen Leser über eine Verbindung, die nichts
// annimmt (`query_only`). Ein Symbol liegt aber als **Datei** neben dem Set
// und nicht in der Datenbank: über `mse symbol-setzen` und
// `mse symbol-nachzeichnen` schrieb ein Leser — und einer ganz ohne Rolle —
// in ein fremdes Profil. MCP und Dienst wiesen es ab, weil sie eine eigene
// Wache haben. Gefunden vom Rollenlauf über die übrigen Schreibklassen
// (`internal/flaechen/rollenuebrigeklassen_test.go`).

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestOhneSchreibrechtLegtKeineSymbolnahtAb(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := probedatenbank(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Fremdes Set", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	// **Bob arbeitet, das Profil zeigt auf das von jemand anderem** — und
	// eine Freigabe hat er dort nicht.
	nutzer := storage.NeueNutzerablage(db)

	bob, fehler := nutzer.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Bob"})
	if fehler != nil {
		t.Fatalf("Bob anlegen: %v", fehler)
	}

	if wechsel := nutzer.NutzerzeigerSetzen(probekontext(t), bob.ID); wechsel != nil {
		t.Fatalf("zu Bob wechseln: %v", wechsel)
	}

	werkstatt, fehler := SymbolwerkstattUeber(db)
	if fehler != nil {
		t.Fatalf("die Werkstatt bauen: %v", fehler)
	}

	teile := SymbolGrundform(SymbolgrundformDTO{Art: "Rechteck"}).Teile
	if _, abgelegt := werkstatt.Ablegen(probekontext(t), setID, teile, ""); !errors.Is(abgelegt, domain.ErrErlaubnis) {
		t.Errorf("ablegen ohne Rolle: %v, erwartet ErrErlaubnis", abgelegt)
	}

	nachzeichnung, fehler := SymbolnachzeichnungUeber(db)
	if fehler != nil {
		t.Fatalf("die Nachzeichnung bauen: %v", fehler)
	}

	if _, gezeichnet := nachzeichnung.Nachzeichnen(probekontext(t), setID, kreisdatei(t)); !errors.Is(gezeichnet, domain.ErrErlaubnis) {
		t.Errorf("nachzeichnen ohne Rolle: %v, erwartet ErrErlaubnis", gezeichnet)
	}

	// **An der Wirkung**: neben dem Set liegt keine Zeichnung.
	bilder, fehler := Bilderverzeichnis()
	if fehler != nil {
		t.Fatalf("das Bilderverzeichnis: %v", fehler)
	}

	_ = filepath.WalkDir(bilder, func(pfad string, eintrag fs.DirEntry, _ error) error { //nolint:errcheck // ein fehlendes Verzeichnis heißt: nichts abgelegt
		if eintrag != nil && strings.HasSuffix(eintrag.Name(), ".mse-symbol") {
			t.Errorf("abgelegt ohne Rolle: %s", pfad)
		}

		return nil
	})
}

// **Und wer schreiben darf, schreibt trotzdem nur ins gezeigte Profil** —
// die Rolle steht jetzt vor dem Set, und ohne diesen Fall erreichte kein
// Lauf mehr die Setwache dahinter (#1815): ein Profil, das es nicht gibt,
// scheitert schon an der Rolle.
func TestEinZweitesEigenesProfilSiehtDasSetDesErstenNicht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := probedatenbank(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Set im ersten Profil", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	zweites, fehler := storage.NeueNutzerablage(db).ProfilAnlegen(probekontext(t),
		domain.Profil{NutzerID: 1, Name: "Zweites"})
	if fehler != nil {
		t.Fatalf("das zweite Profil anlegen: %v", fehler)
	}

	imZweiten := storage.MitProfil(t.Context(), zweites.ID)

	werkstatt, fehler := SymbolwerkstattUeber(db)
	if fehler != nil {
		t.Fatalf("die Werkstatt bauen: %v", fehler)
	}

	teile := SymbolGrundform(SymbolgrundformDTO{Art: "Rechteck"}).Teile
	if _, abgelegt := werkstatt.Ablegen(imZweiten, setID, teile, ""); !errors.Is(abgelegt, domain.ErrNotFound) {
		t.Errorf("ablegen ins Set des ersten Profils: %v, erwartet ErrNotFound", abgelegt)
	}

	nachzeichnung, fehler := SymbolnachzeichnungUeber(db)
	if fehler != nil {
		t.Fatalf("die Nachzeichnung bauen: %v", fehler)
	}

	if _, fehler := nachzeichnung.Nachzeichnen(imZweiten, setID, kreisdatei(t)); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("nachzeichnen ins Set des ersten Profils: %v, erwartet ErrNotFound", fehler)
	}
}
