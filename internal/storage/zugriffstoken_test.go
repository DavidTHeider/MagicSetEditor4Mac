package storage_test

// **Zugriffstoken im Bestand** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probetoken gibt eines für Profil 1.
func probetoken(bezeichnung string) domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID:    1,
		Bezeichnung: bezeichnung,
		Principal:   "Nachtrechner",
		Scope:       domain.ScopeLesen,
		Angelegt:    "2026-09-05T12:00:00Z",
	}
}

// **Ein Token geht hinein, kommt zurück — und sein Geheimnis nicht.**
//
// **Die zweite Hälfte ist die Zusage**: was in einer Liste steht, wird
// angezeigt, und ein Geheimnis, das angezeigt wird, ist keines mehr.
func TestEinTokenKommtOhneSeinGeheimnisZurueck(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Sicherungsskript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if angelegt.ID == 0 || angelegt.Fassung != 1 {
		t.Fatalf("%+v", angelegt)
	}

	gelesen, lesefehler := ablage.Token(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen) != 1 || gelesen[0].Bezeichnung != "Sicherungsskript" {
		t.Fatalf("%+v", gelesen)
	}

	if gelesen[0].Scope != domain.ScopeLesen {
		t.Errorf("der Scope ist %q", gelesen[0].Scope)
	}
}

// **Die gehütete Form gibt es über einen eigenen Weg — und der heißt danach.**
func TestDieGehueteteFormGibtEsNurUeberIhrenWeg(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gelesen, gehuetet, lesefehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gehuetet.Aktuell != "gehütet:abc" {
		t.Errorf("gelesen wurde %q", gehuetet)
	}

	if gelesen.Bezeichnung != "Skript" {
		t.Errorf("%+v", gelesen)
	}

	// **Und eines, das es nicht gibt, ist „nicht gefunden".**
	if _, _, fehler := ablage.TokenMitKennung(
		t.Context(), 4711); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Token lässt sich ändern — mit seiner Fassung.**
func TestEinTokenLaesstSichAendern(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	angelegt.Scope = domain.ScopeSchreiben
	angelegt.Ablauf = "2026-12-24T18:00:00Z"

	geaendert, aenderfehler := ablage.TokenAendern(t.Context(), angelegt)
	if aenderfehler != nil {
		t.Fatalf("ändern: %v", aenderfehler)
	}

	if geaendert.Fassung != 2 {
		t.Errorf("die Fassung ist %d", geaendert.Fassung)
	}

	// **Und das Geheimnis bleibt** — ein Schlüssel, der still seinen Bart
	// wechselt, sperrt jeden aus, der ihn hat.
	_, gehuetet, lesefehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gehuetet.Aktuell != "gehütet:abc" {
		t.Errorf("das Geheimnis wurde %q", gehuetet)
	}
}

// **Eine veraltete Fassung ist ein Konflikt, ein fehlendes Token nicht
// gefunden.**
//
// **Der Unterschied ist die ganze Zusage** (§8): das eine heißt „lies neu",
// das andere „du hast das falsche".
func TestDasAendernUnterscheidetKonfliktUndFehlen(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	veraltet := angelegt
	veraltet.Fassung = 4711

	if _, konflikt := ablage.TokenAendern(
		t.Context(), veraltet); !errors.Is(konflikt, domain.ErrConflict) {
		t.Errorf("gemeldet wurde: %v", konflikt)
	}

	fehlend := angelegt
	fehlend.ID = 4711

	if _, nichts := ablage.TokenAendern(
		t.Context(), fehlend); !errors.Is(nichts, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", nichts)
	}
}

// **Ein Token lässt sich löschen — und zweimal löschen ist „nicht gefunden".**
func TestEinTokenLaesstSichLoeschen(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if err := ablage.TokenLoeschen(t.Context(), angelegt.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if zweites := ablage.TokenLoeschen(
		t.Context(), angelegt.ID); !errors.Is(zweites, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", zweites)
	}
}

// **Zwei Token gleichen Namens an einem Profil gibt es nicht.**
//
// **Weil der Name da ist, um sie zu unterscheiden.** Zwei Zeilen
// „Sicherungsskript" sind zwei Schlüssel, von denen niemand weiß, welcher
// welcher ist — und wer einen abschalten will, trifft den falschen.
func TestZweiTokenGleichenNamensGibtEsNicht(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	if _, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc"); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, zweites := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:xyz"); zweites == nil {
		t.Error("das zweite kam durch")
	}
}

// **Ein Token ohne Profil kommt gar nicht erst in den Bestand.**
func TestEinTokenOhneProfilKommtNichtInDenBestand(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	ohne := probetoken("Skript")
	ohne.ProfilID = 0

	if _, fehler := ablage.TokenAnlegen(
		t.Context(), ohne, "gehütet:abc"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Token, das die Prüfung nicht besteht, wird auch nicht geändert.**
//
// **Die Prüfung steht vor dem Schreiben und nicht danach**: ein `UPDATE`, das
// erst hinterher merkt, dass der Scope erfunden war, hätte ihn schon
// geschrieben.
func TestEinUnbrauchbaresTokenWirdNichtGeaendert(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(
		t.Context(), probetoken("Skript"), "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	angelegt.Scope = "alles"

	if _, verbogen := ablage.TokenAendern(
		t.Context(), angelegt); !errors.Is(verbogen, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", verbogen)
	}

	// **Und es blieb, wie es war** — eine Ablehnung, die trotzdem geschrieben
	// hätte, wäre ein halber Stand (§8).
	gelesen, lesefehler := ablage.Token(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[0].Scope != domain.ScopeLesen {
		t.Errorf("der Scope ist %q", gelesen[0].Scope)
	}
}
