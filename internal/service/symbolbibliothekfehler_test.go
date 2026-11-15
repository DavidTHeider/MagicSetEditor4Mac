package service_test

// **Die Fehlerwege der Symbolbibliothek** — DOM-02
// (#1697):
// jeder Fehler kommt durch, und keiner lässt einen halben Stand zurück.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieBibliothekReichtIhreFehlerDurch(t *testing.T) {
	klemmt := errors.New("klemmt")

	// **Der Speicher klemmt.**
	bibliothek, speicher, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)
	speicher.fehler = klemmt

	if _, err := bibliothek.Liste(t.Context(), ""); !errors.Is(err, klemmt) {
		t.Errorf("Liste: %v", err)
	}

	// **Die Setköpfe klemmen.**
	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))
	ohneKoepfe := service.NeueSymbolbibliothek(&symbolspeicher{}, ablage,
		&setwerteprobe{werte: map[int64]map[string]string{}}, setkoepfe{fehler: klemmt})

	if _, err := ohneKoepfe.Liste(t.Context(), ""); !errors.Is(err, klemmt) {
		t.Errorf("Liste ohne Köpfe: %v", err)
	}

	// **Die Angaben klemmen.**
	ohneWerte := service.NeueSymbolbibliothek(&symbolspeicher{symbole: []domain.Bibliothekssymbol{krone}}, ablage,
		&setwerteprobe{werte: map[int64]map[string]string{}, fehler: klemmt},
		setkoepfe{koepfe: []domain.SetHead{{ID: 1, Name: "Erstes"}}})

	for name, ruf := range map[string]func() error{
		"Liste":     func() error { _, e := ohneWerte.Liste(t.Context(), ""); return e },
		"Aufnehmen": func() error { _, e := ohneWerte.Aufnehmen(t.Context(), 1, "X"); return e },
		"Löschen":   func() error { _, e := ohneWerte.Loeschen(t.Context(), krone.ID, true); return e },
		"Einsetzen": func() error { _, e := ohneWerte.Einsetzen(t.Context(), krone.ID, 1); return e },
	} {
		if err := ruf(); !errors.Is(err, klemmt) {
			t.Errorf("%s mit klemmenden Angaben: %v", name, err)
		}
	}
}

func TestEinFehlendesSymbolOderSetIstNichtGefunden(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	for name, ruf := range map[string]func() error{
		"Kopieren":           func() error { _, e := bibliothek.Kopieren(t.Context(), 99, "X"); return e },
		"Löschen":            func() error { _, e := bibliothek.Loeschen(t.Context(), 99, true); return e },
		"Einsetzen":          func() error { _, e := bibliothek.Einsetzen(t.Context(), 99, 1); return e },
		"Einsetzen ohne Set": func() error { _, e := bibliothek.Einsetzen(t.Context(), krone.ID, 9); return e },
		"Ausgeben":           func() error { return bibliothek.Ausgeben(t.Context(), 99, filepath.Join(t.TempDir(), "x")) },
		"Umbenennen":         func() error { return bibliothek.Umbenennen(t.Context(), 99, "X") },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}

	if err := bibliothek.Ausgeben(t.Context(), krone.ID, " "); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Ausgeben ohne Ort: %v", err)
	}
}

// **Ein Set, dessen Symboldatei fehlt, trägt keines** — und lässt sich
// nicht aufnehmen; die Liste scheitert an ihm nicht.
func TestEinSetMitFehlenderSymboldatei(t *testing.T) {
	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))
	werte := &setwerteprobe{werte: map[int64]map[string]string{1: {"symbol": "weg.mse-symbol"}}}
	bibliothek := service.NeueSymbolbibliothek(&symbolspeicher{}, ablage, werte,
		setkoepfe{koepfe: []domain.SetHead{{ID: 1, Name: "Erstes"}}})

	if eintraege, err := bibliothek.Liste(t.Context(), ""); err != nil || len(eintraege) != 0 {
		t.Errorf("die Liste: %+v, %v", eintraege, err)
	}

	if _, err := bibliothek.Aufnehmen(t.Context(), 1, "Weg"); err == nil {
		t.Error("eine fehlende Datei ließ sich aufnehmen")
	}
}

// **Zwei Sets mit demselben Namen stehen einmal** unter den Trägern.
func TestZweiGleichnamigeSetsStehenEinmal(t *testing.T) {
	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))
	eins, err := ablage.Ablegen(1, msepaket.Standardsymbol(), "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	zwei, err := ablage.Ablegen(2, msepaket.Standardsymbol(), "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	werte := &setwerteprobe{werte: map[int64]map[string]string{1: {"symbol": eins}, 2: {"symbol": zwei}}}
	bibliothek := service.NeueSymbolbibliothek(&symbolspeicher{}, ablage, werte,
		setkoepfe{koepfe: []domain.SetHead{{ID: 1, Name: "Probe"}, {ID: 2, Name: "Probe"}}})

	kroneAufgenommen(t, bibliothek)

	if eintraege := gelistet(t, bibliothek, ""); len(eintraege[0].Verwendet) != 1 {
		t.Errorf("die Träger: %v", eintraege[0].Verwendet)
	}
}

func TestEinsetzenLaesstBeimScheiternKeinenHalbenStand(t *testing.T) {
	bibliothek, speicher, werte := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	// **Ein Inhalt, der die Grenzen nicht mehr hält**, wird nicht eingesetzt.
	speicher.symbole[0].Inhalt = strings.Repeat("part:\n\ttype: shape\n", msepaket.GrenzeSymbolteile+1)

	if _, err := bibliothek.Einsetzen(t.Context(), krone.ID, 2); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("zu viele Teile: %v", err)
	}

	// **Ohne Bildverzeichnis** gibt es keinen Ort für die Kopie.
	ohneOrt := service.NeueSymbolbibliothek(speicher, service.NeueSymbolablage(service.NeueBildablage("")), werte,
		setkoepfe{koepfe: []domain.SetHead{{ID: 2, Name: "Zweites"}}})
	speicher.symbole[0].Inhalt = krone.Inhalt

	if _, err := ohneOrt.Einsetzen(t.Context(), krone.ID, 2); err == nil {
		t.Error("ohne Bildverzeichnis eingesetzt")
	}

	// **Der Wert lässt sich nicht schreiben** — die Datei liegt dann schon,
	// und das ist ein Nachbar mehr, kein falsches Symbol.
	werte.fehler = errors.New("klemmt")
	if _, err := bibliothek.Einsetzen(t.Context(), krone.ID, 2); err == nil || werte.werte[2]["symbol"] != "" {
		t.Errorf("ein klemmender Wert: %v, %q", err, werte.werte[2]["symbol"])
	}
}

func TestEinlesenSagtWasMitDerDateiNichtStimmt(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	ort := t.TempDir()

	// **Zu viele Teile** kommen mit ihrem eigenen Schlüssel an.
	viel := filepath.Join(ort, "viel.mse-symbol")
	if err := os.WriteFile(viel, []byte(strings.Repeat("part:\n\ttype: shape\n", msepaket.GrenzeSymbolteile+1)), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var feldfehler *domain.FieldError
	if _, err := bibliothek.Einlesen(t.Context(), viel, ""); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolZuVieleTeile {
		t.Errorf("zu viele Teile: %v", err)
	}

	// **Ein Verzeichnis ist keine Zeichnung.**
	verzeichnis := filepath.Join(ort, "ordner.mse-symbol")
	if err := os.Mkdir(verzeichnis, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := bibliothek.Einlesen(t.Context(), verzeichnis, ""); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ein Verzeichnis: %v", err)
	}

	// **Eine Datei ohne Namen** braucht einen genannten.
	if _, err := bibliothek.Einlesen(t.Context(), filepath.Join(ort, ".mse-symbol"), ""); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ohne Namen: %v", err)
	}
}
