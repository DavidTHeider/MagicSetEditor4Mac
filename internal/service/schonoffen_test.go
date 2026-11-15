// **Eine Quelle, ein Set** — die Regel gehört in den Fachkern
// (#886).
//
// # Warum sie hierher wandert
//
// Sie stand in `internal/app/oeffnen.go` als `schonOffen` — also in der
// Fassade **des Fensters**. Die Kommandozeile ging am Importeur vorbei daran
// vorbei und legte dasselbe Set ein zweites Mal an; der REST-Dienst der
// nächsten Phase hätte dasselbe getan. Eine fachliche Regel in einer Fassade
// ist eine Regel je Fassade.
package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// setspeicherprobe zählt, wie oft wirklich angelegt wurde — und merkt sich,
// unter welcher Quelle.
type setspeicherprobe struct {
	laeufe  int
	quellen []string
}

func (s *setspeicherprobe) SetAnlegen(
	_ context.Context, set *domain.Set, _ []domain.RawBlock,
) (int64, error) {
	s.laeufe++
	s.quellen = append(s.quellen, set.Origin)

	return int64(s.laeufe), nil
}

// koepfeausprobe gibt die Setköpfe, die der Speicher angelegt hat.
//
// **Über denselben Zähler**, damit die beiden nicht auseinanderlaufen: eine
// Vorrichtung, die anderes meldet als der Speicher hält, prüfte sich selbst.
type koepfeausprobe struct{ speicher *setspeicherprobe }

func (k koepfeausprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	koepfe := make([]domain.SetHead, 0, len(k.speicher.quellen))
	for i, quelle := range k.speicher.quellen {
		koepfe = append(koepfe, domain.SetHead{ID: int64(i + 1), Origin: quelle})
	}

	return koepfe, nil
}

// neuerImporteurMitKoepfen baut den Importeur samt Kopfquelle.
func neuerImporteurMitKoepfen(t *testing.T, speicher *setspeicherprobe) *service.Importeur {
	t.Helper()

	return service.NeuerImporteur(speicher).MitKoepfen(koepfeausprobe{speicher: speicher})
}

// neuerImporteurOhneKoepfe baut ihn ohne — wie bisher.
func neuerImporteurOhneKoepfe(speicher *setspeicherprobe) *service.Importeur {
	return service.NeuerImporteur(speicher)
}

// **Zweimal dieselbe Datei einfahren legt ein Set an, nicht zwei.**
func TestZweimalEinfahrenLegtEinSetAn(t *testing.T) {
	t.Parallel()

	speicher := &setspeicherprobe{}
	importeur := neuerImporteurMitKoepfen(t, speicher)

	pfad := filepath.Join(testdaten, "probe.mse-set")

	erst, err := importeur.Set(t.Context(), pfad)
	if err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	zweit, err := importeur.Set(t.Context(), pfad)
	if err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if zweit.SetID != erst.SetID {
		t.Errorf("der zweite Lauf gab Set %d, der erste %d", zweit.SetID, erst.SetID)
	}

	if len(zweit.Angelegt) != 0 {
		t.Errorf("der zweite Lauf hat %d Sets angelegt, erwartet war keines", len(zweit.Angelegt))
	}

	if speicher.laeufe != 1 {
		t.Errorf("der Speicher lief %d-mal, erwartet war einmal", speicher.laeufe)
	}
}

// **Ohne Kopfquelle bleibt es beim alten Verhalten.**
//
// Sie ist eine Naht mit `nil`-Vorgabe: ein Test, der nur den Weg durch das
// Paket prüft, braucht keinen Bestand — und ein Aufrufer, der die Regel nicht
// will, sagt das, statt sie zu erben.
func TestOhneKopfquelleWirdZweimalAngelegt(t *testing.T) {
	t.Parallel()

	speicher := &setspeicherprobe{}
	importeur := neuerImporteurOhneKoepfe(speicher)

	pfad := filepath.Join(testdaten, "probe.mse-set")

	if _, err := importeur.Set(t.Context(), pfad); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	if _, err := importeur.Set(t.Context(), pfad); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if speicher.laeufe != 2 {
		t.Errorf("der Speicher lief %d-mal, erwartet waren zwei", speicher.laeufe)
	}
}

// **Ein Alias auf dieselbe Datei ist dieselbe Datei.**
//
// Ein Vergleich als Zeichenkette legte dieselben Karten zweimal an — und wer
// im einen Set ändert, fände die Änderung im anderen nicht. Angestrichen vom
// Nachlauf des Zweitblicks über Phase 9; **hierher gewandert mit der Regel**
// (#886).
func TestEinAliasAufDieselbeDateiIstDieselbeDatei(t *testing.T) {
	t.Parallel()

	verzeichnis := t.TempDir()

	echt := filepath.Join(verzeichnis, "echt.mse-set")
	if err := kopiere(t, filepath.Join(testdaten, "probe.mse-set"), echt); err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	verweis := filepath.Join(verzeichnis, "verweis.mse-set")
	if err := os.Symlink(echt, verweis); err != nil {
		t.Fatalf("verweisen: %v", err)
	}

	speicher := &setspeicherprobe{}
	importeur := neuerImporteurMitKoepfen(t, speicher)

	if _, err := importeur.Set(t.Context(), echt); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	if _, err := importeur.Set(t.Context(), verweis); err != nil {
		t.Fatalf("zweiter Lauf über den Verweis: %v", err)
	}

	if speicher.laeufe != 1 {
		t.Errorf("der Speicher lief %d-mal — der Verweis galt als zweite Datei", speicher.laeufe)
	}
}

// **Und auch der umgekehrte Verweis wird erkannt.**
//
// Der Bestand kennt den Verweis, geöffnet wird die echte Datei. Wer nur den
// gesuchten Pfad auflöste, sähe zwei.
func TestAuchDerUmgekehrteVerweisWirdErkannt(t *testing.T) {
	t.Parallel()

	verzeichnis := t.TempDir()

	echt := filepath.Join(verzeichnis, "echt.mse-set")
	if err := kopiere(t, filepath.Join(testdaten, "probe.mse-set"), echt); err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	verweis := filepath.Join(verzeichnis, "verweis.mse-set")
	if err := os.Symlink(echt, verweis); err != nil {
		t.Fatalf("verweisen: %v", err)
	}

	speicher := &setspeicherprobe{}
	importeur := neuerImporteurMitKoepfen(t, speicher)

	if _, err := importeur.Set(t.Context(), verweis); err != nil {
		t.Fatalf("erster Lauf über den Verweis: %v", err)
	}

	if _, err := importeur.Set(t.Context(), echt); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if speicher.laeufe != 1 {
		t.Errorf("der Speicher lief %d-mal — die echte Datei galt als zweite", speicher.laeufe)
	}
}

// kopiere legt ein Setverzeichnis an einem zweiten Ort ab.
//
// **Ein `.mse-set` ist ein Verzeichnis oder ein Archiv**; hier reicht ein
// flacher Abzug, denn geprüft wird der Pfadvergleich und nicht der Inhalt.
func kopiere(t *testing.T, quelle, ziel string) error {
	t.Helper()

	eintraege, err := os.ReadDir(quelle)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(ziel, 0o700); err != nil {
		return err
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			continue
		}

		roh, err := os.ReadFile(filepath.Join(quelle, eintrag.Name())) // #nosec G304 -- Testdaten
		if err != nil {
			return err
		}

		// #nosec G703 -- `ziel` kommt aus t.TempDir(), der Name aus dem
		// gelesenen Verzeichnis; es gibt keine Eingabe.
		if err := os.WriteFile(filepath.Join(ziel, eintrag.Name()), roh, 0o600); err != nil {
			return err
		}
	}

	return nil
}

// **Ein Fehler beim Nachsehen hält den Import an.**
//
// Wer ihn überginge, legte ein Set zum zweiten Mal an, weil der Bestand gerade
// nicht lesbar war.
func TestEinFehlerBeimNachsehenHaeltAn(t *testing.T) {
	t.Parallel()

	speicher := &setspeicherprobe{}
	importeur := service.NeuerImporteur(speicher).MitKoepfen(kaputteKoepfe{})

	if _, err := importeur.Set(t.Context(), filepath.Join(testdaten, "probe.mse-set")); err == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if speicher.laeufe != 0 {
		t.Error("es wurde trotzdem angelegt")
	}
}

// **Und dieselbe Regel gilt für MTG JSON.**
//
// Sie hier wegzulassen hieße: wer eine JSON-Datei zweimal öffnet, führt ihre
// Karten doppelt — und wer in einem Satz ändert, findet die Änderung im
// anderen nicht.
func TestZweimalMTGJSONLegtEinSetAn(t *testing.T) {
	t.Parallel()

	speicher := &setspeicherprobe{}
	importeur := neuerImporteurMitKoepfen(t, speicher)

	pfad := filepath.Join(testdaten, "mtgjson", "probe.json")

	if _, err := importeur.MTGJSON(t.Context(), pfad, "magic"); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	if _, err := importeur.MTGJSON(t.Context(), pfad, "magic"); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if speicher.laeufe != 1 {
		t.Errorf("der Speicher lief %d-mal, erwartet war einmal", speicher.laeufe)
	}
}

// **Ein Set ohne Quelle zählt nicht als dieselbe Datei.**
//
// Ein in der Anwendung entstandenes Set trägt keinen Herkunftspfad; ein leerer
// Vergleich träfe sonst jede Datei.
func TestEinSetOhneQuelleTrifftNichts(t *testing.T) {
	t.Parallel()

	speicher := &setspeicherprobe{}
	importeur := service.NeuerImporteur(speicher).MitKoepfen(ohneQuelle{})

	if _, err := importeur.Set(t.Context(), filepath.Join(testdaten, "probe.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.laeufe != 1 {
		t.Errorf("der Speicher lief %d-mal, erwartet war einmal", speicher.laeufe)
	}
}

// kaputteKoepfe meldet beim Nachsehen einen Fehler.
type kaputteKoepfe struct{}

func (kaputteKoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand ist nicht lesbar")
}

// ohneQuelle gibt ein Set, das keinen Herkunftspfad hat.
type ohneQuelle struct{}

func (ohneQuelle) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Name: "In der Anwendung entstanden"}}, nil
}
