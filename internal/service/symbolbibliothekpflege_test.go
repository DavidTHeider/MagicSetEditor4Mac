package service_test

// **Ein Symbol der Bibliothek bearbeiten** — DOM-02.AK-02 und AK-03
// (#1697):
// eine neue Fassung mit der vorigen daneben, die Sets, die die vorige tragen,
// und der Weg, sie nachzuziehen.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verschoben ist das Standardsymbol mit einem versetzten Punkt — eine andere
// Zeichnung mit einem anderen Abdruck.
func verschoben() msepaket.Symbolzeichnung {
	zeichnung := msepaket.Standardsymbol()
	zeichnung.Teile[0].Punkte[0].Stelle.X += 0.1

	return zeichnung
}

func TestEineNeueFassungNenntDieSetsMitDerVorigen(t *testing.T) {
	bibliothek, speicher, werte := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	aenderung, err := bibliothek.Speichern(t.Context(), krone.ID, verschoben())
	if err != nil || !slices.Equal(aenderung.Veraltet, []string{"Erstes"}) ||
		aenderung.Symbol.Abdruck == krone.Abdruck || aenderung.Symbol.Vorher != krone.Inhalt {
		t.Fatalf("speichern: %+v, %v", aenderung, err)
	}

	// **Das Set behält seine Kopie** — die Bibliothek trägt jetzt eine
	// andere Fassung, und keiner trägt sie.
	if eintraege := gelistet(t, bibliothek, ""); len(eintraege[0].Verwendet) != 0 {
		t.Errorf("nach dem Speichern: %v", eintraege[0].Verwendet)
	}

	// **Dieselbe Zeichnung noch einmal ändert nichts** — sonst wäre die
	// vorige Fassung verloren.
	if nochmal, fehler := bibliothek.Speichern(t.Context(), krone.ID, verschoben()); fehler != nil ||
		len(nochmal.Veraltet) != 0 || speicher.symbole[0].Vorher != krone.Inhalt {
		t.Errorf("dieselbe Zeichnung: %+v, %v", nochmal, fehler)
	}

	// **Nachziehen legt die neue Fassung in die Datei des Sets** — unter
	// demselben Namen, der Wert bleibt.
	datei := werte.werte[1]["symbol"]

	if nachgezogen, fehler := bibliothek.Nachziehen(t.Context(), krone.ID); fehler != nil ||
		!slices.Equal(nachgezogen, []string{"Erstes"}) {
		t.Fatalf("nachziehen: %v, %v", nachgezogen, fehler)
	}

	if werte.werte[1]["symbol"] != datei {
		t.Errorf("der Wert wechselte: %q", werte.werte[1]["symbol"])
	}

	if eintraege := gelistet(t, bibliothek, ""); !slices.Equal(eintraege[0].Verwendet, []string{"Erstes"}) {
		t.Errorf("nach dem Nachziehen: %v", eintraege[0].Verwendet)
	}

	// **Zurücknehmen tauscht die Fassungen** — und nennt das Set, das jetzt
	// die vorige trägt.
	zurueck, err := bibliothek.Zuruecknehmen(t.Context(), krone.ID)
	if err != nil || zurueck.Symbol.Inhalt != krone.Inhalt || !slices.Equal(zurueck.Veraltet, []string{"Erstes"}) {
		t.Fatalf("zurücknehmen: %+v, %v", zurueck, err)
	}
}

func TestDieZeichnungEinesSymbolsKommtZurueck(t *testing.T) {
	bibliothek, speicher, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	zeichnung, err := bibliothek.Zeichnung(t.Context(), krone.ID)
	if err != nil || len(zeichnung.Teile) != len(msepaket.Standardsymbol().Teile) {
		t.Fatalf("die Zeichnung: %+v, %v", zeichnung, err)
	}

	// **Ein Inhalt, der sich nicht mehr lesen lässt**, ist eine Validierung.
	speicher.symbole[0].Inhalt = strings.Repeat("part:\n\ttype: shape\n", msepaket.GrenzeSymbolteile+1)
	if _, err := bibliothek.Zeichnung(t.Context(), krone.ID); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unlesbar: %v", err)
	}

	if _, err := bibliothek.Zeichnung(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("fehlt: %v", err)
	}
}

func TestOhneVorigeFassungGibtEsNichtsZurueckzunehmen(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	var feldfehler *domain.FieldError
	if _, err := bibliothek.Zuruecknehmen(t.Context(), krone.ID); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolOhneVorige {
		t.Errorf("zurücknehmen: %v", err)
	}

	// **Und nachzuziehen gibt es dann auch nichts** — kein Fehler.
	if nachgezogen, err := bibliothek.Nachziehen(t.Context(), krone.ID); err != nil || nachgezogen == nil || len(nachgezogen) != 0 {
		t.Errorf("nachziehen: %v, %v", nachgezogen, err)
	}
}

func TestDieBearbeitungReichtIhreFehlerDurch(t *testing.T) {
	bibliothek, speicher, werte := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	// **Zu viele Teile** werden vor dem Schreiben abgewiesen.
	viele := msepaket.Symbolzeichnung{Teile: make([]msepaket.Symbolteil, msepaket.GrenzeSymbolteile+1)}
	if _, err := bibliothek.Speichern(t.Context(), krone.ID, viele); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("zu viele Teile: %v", err)
	}

	for name, ruf := range map[string]func() error{
		"Speichern":    func() error { _, e := bibliothek.Speichern(t.Context(), 99, verschoben()); return e },
		"Zurücknehmen": func() error { _, e := bibliothek.Zuruecknehmen(t.Context(), 99); return e },
		"Nachziehen":   func() error { _, e := bibliothek.Nachziehen(t.Context(), 99); return e },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}

	if _, err := bibliothek.Speichern(t.Context(), krone.ID, verschoben()); err != nil {
		t.Fatalf("speichern: %v", err)
	}

	// **Klemmende Angaben** kommen durch — beim Speichern und beim
	// Nachziehen.
	klemmt := errors.New("klemmt")
	werte.fehler = klemmt

	if _, err := bibliothek.Nachziehen(t.Context(), krone.ID); !errors.Is(err, klemmt) {
		t.Errorf("nachziehen mit klemmenden Angaben: %v", err)
	}

	if _, err := bibliothek.Zuruecknehmen(t.Context(), krone.ID); !errors.Is(err, klemmt) {
		t.Errorf("zurücknehmen mit klemmenden Angaben: %v", err)
	}

	werte.fehler = nil

	// **Ein klemmender Speicher** beim Lesen und beim Schreiben der neuen
	// Fassung.
	speicher.fehler = klemmt
	if _, err := bibliothek.Speichern(t.Context(), krone.ID, msepaket.Standardsymbol()); !errors.Is(err, klemmt) {
		t.Errorf("speichern mit klemmendem Speicher: %v", err)
	}

	speicher.fehler, speicher.schreibfehler = nil, klemmt
	if _, err := bibliothek.Speichern(t.Context(), krone.ID, msepaket.Standardsymbol()); !errors.Is(err, klemmt) {
		t.Errorf("speichern mit klemmendem Schreiben: %v", err)
	}

	speicher.schreibfehler = nil

	// **Ein unlesbarer Inhalt** lässt sich nicht nachziehen.
	speicher.symbole[0].Inhalt = strings.Repeat("part:\n\ttype: shape\n", msepaket.GrenzeSymbolteile+1)
	if _, err := bibliothek.Nachziehen(t.Context(), krone.ID); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("nachziehen unlesbar: %v", err)
	}
}

// **Ein Set, in dessen Ordner nicht geschrieben werden darf**, lässt das
// Nachziehen scheitern — und die Antwort nennt, was davor schon nachgezogen
// war: hier nichts.
func TestNachziehenInEinenGesperrtenOrdnerSagtEs(t *testing.T) {
	wurzel := t.TempDir()
	ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

	datei, err := ablage.Ablegen(1, msepaket.Standardsymbol(), "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	speicher := &symbolspeicher{}
	werte := &setwerteprobe{werte: map[int64]map[string]string{1: {"symbol": datei}}}
	bibliothek := service.NeueSymbolbibliothek(speicher, ablage, werte,
		setkoepfe{koepfe: []domain.SetHead{{ID: 1, Name: "Erstes"}}})
	krone := kroneAufgenommen(t, bibliothek)

	if _, err := bibliothek.Speichern(t.Context(), krone.ID, verschoben()); err != nil {
		t.Fatalf("speichern: %v", err)
	}

	sperreOrdnerVon(t, wurzel, datei)

	if nachgezogen, err := bibliothek.Nachziehen(t.Context(), krone.ID); err == nil || len(nachgezogen) != 0 {
		t.Errorf("gesperrt: %v, %v", nachgezogen, err)
	}
}

// sperreOrdnerVon nimmt dem Ordner einer Datei das Schreibrecht — und gibt es
// nach dem Lauf zurück, damit er sich aufräumen lässt.
func sperreOrdnerVon(t *testing.T, wurzel, name string) {
	t.Helper()

	var ordner string

	_ = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, _ error) error { //nolint:errcheck // gesucht wird, gefunden oder nicht
		if eintrag != nil && eintrag.Name() == name {
			ordner = filepath.Dir(pfad)
		}

		return nil
	})

	if ordner == "" {
		t.Fatalf("%s liegt nicht unter %s", name, wurzel)
	}

	if err := os.Chmod(ordner, 0o500); err != nil { //nolint:gosec // ein Ordner dieses Laufs
		t.Fatalf("sperren: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(ordner, 0o700) }) //nolint:errcheck,gosec // dito
}
