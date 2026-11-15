package storage_test

// **Rotation, letzte Nutzung und Sammelwiderruf im Bestand** — Werkstück O2
// von Phase 5.2 (#1426).
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Rotation setzt ein neues Geheimnis und hebt das alte auf — genau eines.**
func TestRotationHebtGenauEinVorigesGeheimnisAuf(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(t.Context(), probetoken("Nachtlauf"), "erstes")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	rotiert, rotationsfehler := ablage.TokenRotieren(
		t.Context(), angelegt.ID, "zweites", "2026-09-08T13:00:00Z")
	if rotationsfehler != nil {
		t.Fatalf("rotieren: %v", rotationsfehler)
	}

	if rotiert.VorgaengerBis != "2026-09-08T13:00:00Z" || rotiert.Fassung != 2 {
		t.Errorf("%+v", rotiert)
	}

	_, gehuetet, lesefehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gehuetet.Aktuell != "zweites" || gehuetet.Vorgaenger != "erstes" {
		t.Fatalf("%+v", gehuetet)
	}

	// **Und ein zweiter Lauf schiebt das erste heraus, statt es zu behalten.**
	if _, fehler := ablage.TokenRotieren(
		t.Context(), angelegt.ID, "drittes", "2026-09-08T14:00:00Z"); fehler != nil {
		t.Fatalf("zweite Rotation: %v", fehler)
	}

	_, danach, danachfehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if danachfehler != nil {
		t.Fatalf("lesen: %v", danachfehler)
	}

	if danach.Aktuell != "drittes" || danach.Vorgaenger != "zweites" {
		t.Errorf("nach zwei Rotationen %+v — »erstes« hängt noch dran", danach)
	}
}

// **Ein Token, das es nicht gibt, rotiert nicht.**
func TestRotationEinesUnbekanntenTokensScheitert(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	_, fehler := ablage.TokenRotieren(t.Context(), 404, "neu", "")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Rotation eines Unbekannten: %v", fehler)
	}
}

// **Die letzte Nutzung wird vermerkt — und nur sie.**
func TestDieLetzteNutzungWirdVermerkt(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(t.Context(), probetoken("Nachtlauf"), "erstes")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if angelegt.LetzteNutzung != "" {
		t.Errorf("frisch angelegt trägt schon eine Nutzung: %q", angelegt.LetzteNutzung)
	}

	if fehler := ablage.TokenNutzungVermerken(
		t.Context(), angelegt.ID, "2026-09-08T12:30:00Z"); fehler != nil {
		t.Fatalf("vermerken: %v", fehler)
	}

	gelesen, _, lesefehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen.LetzteNutzung != "2026-09-08T12:30:00Z" {
		t.Errorf("die letzte Nutzung ist %q", gelesen.LetzteNutzung)
	}

	// **Die Fassung bleibt.** Ein Vermerk über eine Benutzung ist keine
	// Änderung, die jemandes gelesenen Stand veraltet — §7, Regel 2 hinge
	// sonst an einer Uhr.
	if gelesen.Fassung != 1 {
		t.Errorf("der Vermerk hat die Fassung auf %d gehoben", gelesen.Fassung)
	}
}

// **Ein Vermerk auf ein Token, das es nicht gibt, sagt es.**
func TestEinVermerkAufEinUnbekanntesTokenSagtEs(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	if fehler := ablage.TokenNutzungVermerken(
		t.Context(), 404, "2026-09-08T12:30:00Z"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Vermerk auf ein Unbekanntes: %v", fehler)
	}
}

// **Der Sammelwiderruf nimmt alle Token eines Profils — und nur dessen.**
func TestDerSammelwiderrufTrifftNurEinProfil(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueZugriffstokenablage(db)

	zweiteszeile, profilfehler := storage.NeueNutzerablage(db).ProfilAnlegen(
		t.Context(), domain.Profil{NutzerID: 1, Name: "Zweites"})
	if profilfehler != nil {
		t.Fatalf("Profil anlegen: %v", profilfehler)
	}

	zweites := zweiteszeile.ID

	for _, name := range []string{"Nachtlauf", "Bogenbau"} {
		if _, fehler := ablage.TokenAnlegen(t.Context(), probetoken(name), "x"); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	fremd := probetoken("Fremdes")
	fremd.ProfilID = zweites

	if _, fehler := ablage.TokenAnlegen(t.Context(), fremd, "y"); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gezogen, fehler := ablage.TokenAlleLoeschen(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("sammelwiderruf: %v", fehler)
	}

	if gezogen != 2 {
		t.Errorf("%d Token gezogen, erwartet 2", gezogen)
	}

	uebrig, lesefehler := ablage.Token(t.Context(), zweites)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(uebrig) != 1 {
		t.Errorf("im zweiten Profil stehen %d Token", len(uebrig))
	}
}

// **Ein Sammelwiderruf über ein Profil ohne Token ist kein Fehler.**
//
// **Weil er sein Ziel erreicht hat**: danach gibt es dort keines. Ein
// `ErrNotFound` hieße, der Rufer müsse etwas tun — und er muss nichts.
func TestEinSammelwiderrufOhneToken(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	gezogen, fehler := ablage.TokenAlleLoeschen(t.Context(), 1)
	if fehler != nil || gezogen != 0 {
		t.Errorf("%d Token, Fehler %v", gezogen, fehler)
	}
}

// **Principal und letzte Nutzung stehen in jeder Liste.**
func TestPrincipalUndNutzungStehenInDerListe(t *testing.T) {
	ablage := storage.NeueZugriffstokenablage(frischeDatenbank(t))

	angelegt, fehler := ablage.TokenAnlegen(t.Context(), probetoken("Nachtlauf"), "x")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.TokenNutzungVermerken(
		t.Context(), angelegt.ID, "2026-09-08T12:30:00Z"); fehler != nil {
		t.Fatalf("vermerken: %v", fehler)
	}

	gelesen, lesefehler := ablage.Token(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[0].Principal != "Nachtrechner" ||
		gelesen[0].LetzteNutzung != "2026-09-08T12:30:00Z" {
		t.Errorf("%+v", gelesen[0])
	}
}
