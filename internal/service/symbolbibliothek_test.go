package service_test

// **Die Symbolbibliothek** — DOM-02.AK-01, AK-02 (Verwendung, Löschfolgen)
// und AK-04 (#1697).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// symbolspeicher ist die Bibliothek im Arbeitsspeicher.
type symbolspeicher struct {
	symbole  []domain.Bibliothekssymbol
	naechste int64
	fehler   error

	// schreibfehler trifft nur das Ändern — Lesen geht weiter.
	schreibfehler error
}

func (s *symbolspeicher) Symbole(context.Context) ([]domain.Bibliothekssymbol, error) {
	return slices.Clone(s.symbole), s.fehler
}

func (s *symbolspeicher) Symbol(_ context.Context, id int64) (domain.Bibliothekssymbol, error) {
	for _, symbol := range s.symbole {
		if symbol.ID == id {
			return symbol, s.fehler
		}
	}

	return domain.Bibliothekssymbol{}, errors.Join(domain.ErrNotFound, s.fehler)
}

func (s *symbolspeicher) SymbolAnlegen(_ context.Context, symbol domain.Bibliothekssymbol) (domain.Bibliothekssymbol, error) {
	if s.fehler != nil {
		return domain.Bibliothekssymbol{}, s.fehler
	}

	for _, da := range s.symbole {
		if strings.EqualFold(da.Name, symbol.Name) {
			return domain.Bibliothekssymbol{}, domain.ErrDoppelt
		}
	}

	s.naechste++
	symbol.ID = s.naechste
	s.symbole = append(s.symbole, symbol)

	return symbol, nil
}

func (s *symbolspeicher) SymbolUmbenennen(_ context.Context, id int64, name string) error {
	return s.aendere(id, func(symbol *domain.Bibliothekssymbol) { symbol.Name = name })
}

func (s *symbolspeicher) SymbolInhaltSetzen(_ context.Context, id int64, inhalt, abdruck string) error {
	return s.aendere(id, func(symbol *domain.Bibliothekssymbol) {
		symbol.Vorher, symbol.Inhalt, symbol.Abdruck = symbol.Inhalt, inhalt, abdruck
	})
}

func (s *symbolspeicher) SymbolLoeschen(_ context.Context, id int64) error {
	if s.fehler != nil {
		return s.fehler
	}

	vorher := len(s.symbole)
	s.symbole = slices.DeleteFunc(s.symbole, func(symbol domain.Bibliothekssymbol) bool { return symbol.ID == id })

	if len(s.symbole) == vorher {
		return domain.ErrNotFound
	}

	return nil
}

func (s *symbolspeicher) aendere(id int64, wie func(*domain.Bibliothekssymbol)) error {
	if s.fehler != nil {
		return s.fehler
	}

	if s.schreibfehler != nil {
		return s.schreibfehler
	}

	for i := range s.symbole {
		if s.symbole[i].ID == id {
			wie(&s.symbole[i])

			return nil
		}
	}

	return domain.ErrNotFound
}

// setwerteprobe hält die Angaben der Sets.
type setwerteprobe struct {
	werte  map[int64]map[string]string
	fehler error
}

func (p *setwerteprobe) Setwerte(_ context.Context, setID int64) ([]domain.Value, error) {
	werte := make([]domain.Value, 0, len(p.werte[setID]))
	for feld, text := range p.werte[setID] {
		werte = append(werte, domain.Value{Field: feld, Text: text})
	}

	return werte, p.fehler
}

func (p *setwerteprobe) SetwertSetzenMitVorwert(_ context.Context, setID int64, feld, wert string) (
	vorher string, hatte bool, fehler error,
) {
	if p.fehler != nil {
		return "", false, p.fehler
	}

	if p.werte[setID] == nil {
		p.werte[setID] = map[string]string{}
	}

	vorher, hatte = p.werte[setID][feld]
	p.werte[setID][feld] = wert

	return vorher, hatte, nil
}

// kroneAufgenommen nimmt das Symbol des ersten Sets als »Krone« auf — ohne
// Fehler.
func kroneAufgenommen(t *testing.T, bibliothek *service.Symbolbibliothek) domain.Bibliothekssymbol {
	t.Helper()

	symbol, err := bibliothek.Aufnehmen(t.Context(), 1, "Krone")
	if err != nil {
		t.Fatalf("aufnehmen: %v", err)
	}

	return symbol
}

// gelistet gibt die Liste — ohne Fehler.
func gelistet(t *testing.T, bibliothek *service.Symbolbibliothek, suche string) []service.Symboleintrag {
	t.Helper()

	eintraege, err := bibliothek.Liste(t.Context(), suche)
	if err != nil {
		t.Fatalf("die Liste: %v", err)
	}

	return eintraege
}

// bibliotheksplatz baut die Bibliothek über zwei Sets; das erste trägt das
// Standardsymbol.
func bibliotheksplatz(t *testing.T) (bibliothek *service.Symbolbibliothek, speicher *symbolspeicher, werte *setwerteprobe) {
	t.Helper()

	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))

	name, err := ablage.Ablegen(1, msepaket.Standardsymbol(), "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	speicher = &symbolspeicher{}
	werte = &setwerteprobe{werte: map[int64]map[string]string{1: {"symbol": name}}}

	return service.NeueSymbolbibliothek(speicher, ablage, werte, setkoepfe{koepfe: []domain.SetHead{
		{ID: 1, Name: "Erstes"}, {ID: 2, Name: "Zweites"},
	}}), speicher, werte
}

func TestEinSymbolWirdAusDemSetAufgenommenUndEingesetzt(t *testing.T) {
	bibliothek, _, werte := bibliotheksplatz(t)

	krone, err := bibliothek.Aufnehmen(t.Context(), 1, "  Krone ")
	if err != nil || krone.Name != "Krone" || !strings.HasPrefix(krone.Abdruck, "sha256:") ||
		krone.Inhalt != string(msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol())) {
		t.Fatalf("aufnehmen: %+v, %v", krone, err)
	}

	eintraege, err := bibliothek.Liste(t.Context(), "")
	if err != nil || len(eintraege) != 1 || !slices.Equal(eintraege[0].Verwendet, []string{"Erstes"}) ||
		len(eintraege[0].Zeichnung.Teile) == 0 {
		t.Fatalf("die Liste: %+v, %v", eintraege, err)
	}

	datei, err := bibliothek.Einsetzen(t.Context(), krone.ID, 2)
	if err != nil || werte.werte[2]["symbol"] != datei || !strings.HasSuffix(datei, ".mse-symbol") {
		t.Fatalf("einsetzen: %q, %v", datei, err)
	}

	// **Beide Sets tragen es jetzt** — erkannt am Abdruck, nicht am Namen.
	if eintraege := gelistet(t, bibliothek, ""); !slices.Equal(eintraege[0].Verwendet, []string{"Erstes", "Zweites"}) {
		t.Errorf("nach dem Einsetzen: %+v", eintraege[0].Verwendet)
	}
}

func TestDieBibliothekSuchtKopiertUndBenenntUm(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)

	krone := kroneAufgenommen(t, bibliothek)

	kopie, err := bibliothek.Kopieren(t.Context(), krone.ID, "Anker")
	if err != nil || kopie.ID == krone.ID || kopie.Abdruck != krone.Abdruck {
		t.Fatalf("kopieren: %+v, %v", kopie, err)
	}

	if err := bibliothek.Umbenennen(t.Context(), kopie.ID, "Diadem"); err != nil {
		t.Errorf("umbenennen: %v", err)
	}

	// **Gesucht wird im Namen, ohne Rücksicht auf Groß und Klein.**
	if gefunden := gelistet(t, bibliothek, "DIA"); len(gefunden) != 1 || gefunden[0].Symbol.Name != "Diadem" {
		t.Errorf("die Suche: %+v", gefunden)
	}

	for name, ruf := range map[string]func() error{
		"leerer Name beim Kopieren":   func() error { _, e := bibliothek.Kopieren(t.Context(), krone.ID, " "); return e },
		"leerer Name beim Umbenennen": func() error { return bibliothek.Umbenennen(t.Context(), krone.ID, "") },
		"leerer Name beim Aufnehmen":  func() error { _, e := bibliothek.Aufnehmen(t.Context(), 1, ""); return e },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// **Ein Symbol, das Sets tragen, wird nur mit Bestätigung gelöscht** — und
// die Absage nennt sie (AK-02).
func TestEinVerwendetesSymbolZeigtVorDemLoeschenSeineFolgen(t *testing.T) {
	bibliothek, speicher, _ := bibliotheksplatz(t)

	krone := kroneAufgenommen(t, bibliothek)

	var feldfehler *domain.FieldError

	folgen, err := bibliothek.Loeschen(t.Context(), krone.ID, false)
	if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolVerwendet || feldfehler.Werte["sets"] != "Erstes" ||
		!slices.Equal(folgen, []string{"Erstes"}) || len(speicher.symbole) != 1 {
		t.Fatalf("ohne Bestätigung: %v, %v", folgen, err)
	}

	if bestaetigt, fehler := bibliothek.Loeschen(t.Context(), krone.ID, true); fehler != nil || len(bestaetigt) != 1 || len(speicher.symbole) != 0 {
		t.Errorf("mit Bestätigung: %v, %v", bestaetigt, fehler)
	}

	// **Ein ungenutztes geht ohne Frage.**
	frei, err := speicher.SymbolAnlegen(t.Context(), domain.Bibliothekssymbol{Name: "Frei", Inhalt: "x", Abdruck: "sha256:frei"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}
	if _, err := bibliothek.Loeschen(t.Context(), frei.ID, false); err != nil {
		t.Errorf("ungenutzt: %v", err)
	}
}

func TestEinSymbolGehtHinausUndKommtWieder(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	ziel := filepath.Join(t.TempDir(), "krone.mse-symbol")
	if err := bibliothek.Ausgeben(t.Context(), krone.ID, ziel); err != nil {
		t.Fatalf("ausgeben: %v", err)
	}

	// **Ohne Namen heißt es wie die Datei.**
	gelesen, err := bibliothek.Einlesen(t.Context(), ziel, "")
	if !errors.Is(err, domain.ErrDoppelt) {
		t.Errorf("derselbe Name: %+v, %v", gelesen, err)
	}

	gelesen, err = bibliothek.Einlesen(t.Context(), ziel, "Krone, eingelesen")
	if err != nil || gelesen.Abdruck != krone.Abdruck {
		t.Errorf("eingelesen: %+v, %v", gelesen, err)
	}

	anders := filepath.Join(t.TempDir(), "anker.mse-symbol")
	if err := os.WriteFile(anders, msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol()), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if gelesen, err := bibliothek.Einlesen(t.Context(), anders, ""); err != nil || gelesen.Name != "anker" {
		t.Errorf("mit dem Namen der Datei: %+v, %v", gelesen, err)
	}
}

// **Schadhafte und zu große Dateien bekommen einen verständlichen Fehler**
// (AK-04) — mit eigenem Schlüssel, nicht als „ungültig".
func TestEinSchadhaftesSymbolWirdVerstaendlichAbgewiesen(t *testing.T) {
	bibliothek, speicher, _ := bibliotheksplatz(t)
	ort := t.TempDir()

	gross := filepath.Join(ort, "gross.mse-symbol")
	if err := os.WriteFile(gross, make([]byte, msepaket.GrenzeSymboldatei+1), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var feldfehler *domain.FieldError
	if _, err := bibliothek.Einlesen(t.Context(), gross, ""); !errors.Is(err, domain.ErrValidation) ||
		!errors.As(err, &feldfehler) || feldfehler.Key != domain.SchluesselSymbolZuGross || feldfehler.Werte["grenze"] != "1 MiB" {
		t.Errorf("zu groß: %v", err)
	}

	if _, err := bibliothek.Einlesen(t.Context(), filepath.Join(ort, "fehlt.mse-symbol"), ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("eine fehlende Datei: %v", err)
	}

	if len(speicher.symbole) != 0 {
		t.Errorf("etwas kam doch hinein: %+v", speicher.symbole)
	}
}

func TestEinSetOhneSymbolLaesstSichNichtAufnehmen(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)

	var feldfehler *domain.FieldError
	if _, err := bibliothek.Aufnehmen(t.Context(), 2, "Leer"); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSetOhneSymbol {
		t.Errorf("ohne Symbol: %v", err)
	}

	if _, err := bibliothek.Aufnehmen(t.Context(), 9, "Fremd"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Set: %v", err)
	}
}

// **Die Vorschau zeichnet der Rasterer** — dieselbe Rechnung wie auf der
// Karte, keine zweite im Fenster.
func TestDieVorschauIstEinGerastertesBild(t *testing.T) {
	bild := service.Symbolvorschau(msepaket.Standardsymbol(), 64)
	if bild.Bounds().Dx() == 0 || bild.Bounds().Dx() > 64 || bild.Bounds().Dy() > 64 {
		t.Fatalf("die Maße: %v", bild.Bounds())
	}

	gedeckt := 0
	for stelle := 3; stelle < len(bild.Pix); stelle += 4 {
		if bild.Pix[stelle] > 0 {
			gedeckt++
		}
	}

	if gedeckt == 0 {
		t.Error("die Vorschau ist leer")
	}
}
