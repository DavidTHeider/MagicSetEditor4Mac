package storage_test

// **Der Partiespeicher** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die dritte Abnahme des Werkstücks
//
// „Neuaufbau nach Prozessabbruch rekonstruiert den letzten atomar
// bestätigten Zustand; beschädigte Hashkette oder falscher Snapshot wird
// abgewiesen und **nie still repariert**."
//
// **Geprüft wird gegen eine echte Datenbank** (§3), denn die Zusage ist eine
// über Transaktionen — und die hält kein Doppel im Arbeitsspeicher.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// partieablage gibt eine frische Ablage samt Datenbank.
func partieablage(t *testing.T) *storage.Partieablage {
	t.Helper()

	return storage.NeuePartieablage(frischeDatenbank(t))
}

// einePartie ist die kleinste, die der Speicher hält.
func einePartie() domain.Partie {
	abzug := domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 40},
		},
	}

	return domain.Partie{
		ID:    "p1",
		Spiel: "magic",
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt:  "sitz-1",
		Stand: domain.StandLaeuft,
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David", Deck: abzug},
			{Kennung: "sitz-2", Name: "Gegenüber", Deck: abzug},
		},
		Objekte: []domain.Spielobjekt{
			{
				Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
				Eigner: "sitz-1", Zone: domain.ZoneHand, Karte: "k1",
			},
		},
	}
}

// **Eine Partie lässt sich anlegen und ganz wiederfinden.**
func TestEinePartieLaesstSichAnlegenUndGanzWiederfinden(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	angelegt, fehler := ablage.Anlegen(ctx, einePartie())
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if angelegt.Fassung != 1 {
		t.Errorf("die neue Partie hat die Fassung %d", angelegt.Fassung)
	}

	gelesen, fehler := ablage.Partie(ctx, "p1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Der ganze Baum kommt zurück** — Sitze, Deckabzüge, Objekte. Ein
	// Speicher, der nur die Zeile hielte, gäbe eine Partie ohne Tisch.
	if len(gelesen.Sitze) != 2 || len(gelesen.Objekte) != 1 {
		t.Fatalf("die Partie kam unvollständig zurück: %+v", gelesen)
	}

	if gelesen.Sitze[0].Deck.Abdruck != "abcd" ||
		gelesen.Objekte[0].Karte != "k1" {
		t.Errorf("der Inhalt stimmt nicht: %+v", gelesen)
	}

	if gelesen.Regelwerk.Strenge != domain.StrengeSandkasten {
		t.Errorf("das Regelwerk fehlt: %+v", gelesen.Regelwerk)
	}
}

// **Eine Partie aus einem fremden Profil gibt es nicht.**
func TestEinePartieAusEinemFremdenProfilGibtEsNicht(t *testing.T) {
	ablage := partieablage(t)

	if _, fehler := ablage.Anlegen(probekontext(t), einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Ohne Profil im Kontext geht gar nichts** — die Klemme aus #1351.
	if _, fehler := ablage.Partie(t.Context(), "p1"); fehler == nil {
		t.Error("ohne Profil kam eine Partie zurück")
	}

	if _, fehler := ablage.Partie(
		probekontext(t), "p9"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("eine erfundene Partie gab %v", fehler)
	}
}

// **Zustand, Strom und Befehl entstehen zusammen — oder gar nicht.**
func TestZustandStromUndBefehlEntstehenZusammen(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	danach := einePartie()
	danach.Fassung = 2
	danach.Objekte[0].Getappt = true

	ereignisse := []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Art: "objekt.tappen",
		Oeffentlich: map[string]any{"objekt": "o1", "sitz": "sitz-1"},
		Privat: []domain.Ergaenzung{{
			Empfaenger: "sitz-1",
			Inhalt:     map[string]any{"karte": "k1"},
		}},
		Kette: "abdruck-1",
	}}

	if fehler := ablage.Anwenden(ctx, 1, "b1", danach, ereignisse); fehler != nil {
		t.Fatalf("anwenden: %v", fehler)
	}

	gelesen, fehler := ablage.Partie(ctx, "p1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if !gelesen.Objekte[0].Getappt || gelesen.Fassung != 2 {
		t.Errorf("der Zustand kam nicht an: %+v", gelesen)
	}

	// **Und der Strom steht daneben**, samt seiner privaten Ergänzung.
	strom, fehler := ablage.Ereignisse(ctx, "p1", 1, 99)
	if fehler != nil {
		t.Fatalf("Ereignisse: %v", fehler)
	}

	if len(strom) != 1 || strom[0].Kette != "abdruck-1" {
		t.Fatalf("der Strom: %+v", strom)
	}

	if len(strom[0].Privat) != 1 ||
		strom[0].Privat[0].Inhalt["karte"] != "k1" {
		t.Errorf("die Ergänzung fehlt: %+v", strom[0].Privat)
	}

	// **Und `0, 0` heißt alles** — so steht es am Vertrag in
	// `service.Partiespeicher`, und hier wird es eingelöst. Der erste Anlauf
	// setzte beide Grenzen wörtlich ein: `folge >= 0 AND folge <= 0` trifft
	// nie etwas, denn die Folge fängt bei eins an. Die Rückschau war damit an
	// **jeder** Fläche leer, und zwar ohne Fehler — angestrichen hat es der
	// Lauf in `internal/flaechen/partieverlauf_test.go` (§10: eine Zusage,
	// die nur im Kopf steht, hält niemand).
	ganz, fehler := ablage.Ereignisse(ctx, "p1", 0, 0)
	if fehler != nil {
		t.Fatalf("Ereignisse ohne Grenzen: %v", fehler)
	}

	if len(ganz) != len(strom) {
		t.Errorf("»0, 0« gab %d Ereignisse, der ganze Strom hat %d",
			len(ganz), len(strom))
	}

	// **Und der Befehl ist gemerkt** — das ist der Schutz gegen die
	// doppelte Wirkung.
	gemerkt, gab, fehler := ablage.Befehlsergebnis(ctx, "p1", "b1")
	if fehler != nil {
		t.Fatalf("Befehlsergebnis: %v", fehler)
	}

	if !gab || len(gemerkt) != 1 {
		t.Errorf("der Befehl ist nicht gemerkt: %v, %+v", gab, gemerkt)
	}

	// **Ein Befehl, den es nicht gab, sagt das** — und gibt keine leere
	// Liste, die wie „hat nichts bewirkt" aussähe.
	if _, gab, fehler := ablage.Befehlsergebnis(
		ctx, "p1", "b9"); fehler != nil || gab {
		t.Errorf("ein unbekannter Befehl gab %v, %v", gab, fehler)
	}
}

// **Eine überholte Fassung schreibt nichts — und hinterlässt nichts.**
//
// **Das ist die Atomarität, und sie wird hier wirklich gemessen**: nach dem
// abgewiesenen Ruf darf weder der Zustand noch ein Ereignis noch ein
// gemerkter Befehl dastehen.
func TestEineUeberholteFassungHinterlaesstNichts(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	danach := einePartie()
	danach.Fassung = 2

	ereignisse := []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Art: "objekt.tappen",
		Oeffentlich: map[string]any{"objekt": "o1", "sitz": "sitz-1"},
		Kette:       "abdruck-1",
	}}

	if fehler := ablage.Anwenden(
		ctx, 9, "b1", danach, ereignisse); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	gelesen, fehler := ablage.Partie(ctx, "p1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Fassung != 1 {
		t.Errorf("die Fassung ist %d", gelesen.Fassung)
	}

	strom, fehler := ablage.Ereignisse(ctx, "p1", 1, 99)
	if fehler != nil {
		t.Fatalf("Ereignisse: %v", fehler)
	}

	if len(strom) != 0 {
		t.Errorf("es stehen %d Ereignisse da", len(strom))
	}

	_, gab, fehler := ablage.Befehlsergebnis(ctx, "p1", "b1")
	if fehler != nil {
		t.Fatalf("Befehlsergebnis: %v", fehler)
	}

	if gab {
		t.Error("der Befehl ist gemerkt, obwohl er nicht gewirkt hat")
	}
}

// **Die letzte Kette, und der Anfang, an dem es keine gibt.**
func TestDieLetzteKette(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Eine Partie ohne Ereignis ist kein Fehler, sondern der Anfang.**
	kette, folge, fehler := ablage.LetzteKette(ctx, "p1")
	if fehler != nil || kette != "" || folge != 0 {
		t.Fatalf("der Anfang gab %q, %d, %v", kette, folge, fehler)
	}

	danach := einePartie()
	danach.Fassung = 2

	if angewendet := ablage.Anwenden(ctx, 1, "b1", danach, []domain.Ereignis{
		{
			PartieID: "p1", Folge: 1, Art: "a",
			Oeffentlich: map[string]any{"sitz": "sitz-1"}, Kette: "k1",
		},
		{
			PartieID: "p1", Folge: 2, Art: "b",
			Oeffentlich: map[string]any{"sitz": "sitz-1"}, Kette: "k2",
		},
	}); angewendet != nil {
		t.Fatalf("anwenden: %v", angewendet)
	}

	kette, folge, fehler = ablage.LetzteKette(ctx, "p1")
	if fehler != nil || kette != "k2" || folge != 2 {
		t.Errorf("die letzte Kette ist %q bei Folge %d (%v)", kette, folge, fehler)
	}
}

// **Der letzte Befehl, und der Anfang, an dem es keinen gibt** — P8 (#2376).
// Zwei Befehle; der zweite schreibt zwei Ereignisse, und beide kommen.
func TestDerLetzteBefehl(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if letzte, fehler := ablage.LetzterBefehl(ctx, "p1"); fehler != nil || len(letzte) != 0 {
		t.Fatalf("der Anfang gab %v, %v", letzte, fehler)
	}

	ereignis := func(folge int64, art string) domain.Ereignis {
		return domain.Ereignis{
			PartieID: "p1", Folge: folge, Art: art,
			Oeffentlich: map[string]any{"sitz": "sitz-1"}, Kette: art,
		}
	}

	danach := einePartie()
	for fassung, schritt := range []struct {
		kennung    string
		ereignisse []domain.Ereignis
	}{
		{"b1", []domain.Ereignis{ereignis(1, "a")}},
		{"b2", []domain.Ereignis{ereignis(2, "b"), ereignis(3, "c")}},
	} {
		danach.Fassung = int64(fassung) + 2
		if fehler := ablage.Anwenden(ctx, int64(fassung)+1, schritt.kennung, danach, schritt.ereignisse); fehler != nil {
			t.Fatalf("anwenden %s: %v", schritt.kennung, fehler)
		}
	}

	letzte, fehler := ablage.LetzterBefehl(ctx, "p1")
	if fehler != nil || len(letzte) != 2 || letzte[0].Art != "b" || letzte[1].Art != "c" {
		t.Errorf("der letzte Befehl gab %+v (%v)", letzte, fehler)
	}
}

// **Pausieren, fortsetzen, beenden — und die Liste.**
func TestDerStandUndDieListe(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gesetzt, fehler := ablage.Standsetzen(ctx, "p1", domain.StandPausiert)
	if fehler != nil {
		t.Fatalf("Standsetzen: %v", fehler)
	}

	if gesetzt.Stand != domain.StandPausiert || gesetzt.Fassung != 2 {
		t.Errorf("der Stand: %+v", gesetzt)
	}

	// **Und er steht auch im gelesenen Zustand** — nicht nur in der Spalte.
	gelesen, fehler := ablage.Partie(ctx, "p1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Stand != domain.StandPausiert {
		t.Errorf("der gelesene Stand ist %q", gelesen.Stand)
	}

	if _, erfunden := ablage.Standsetzen(
		ctx, "p9", domain.StandBeendet); !errors.Is(
		erfunden, domain.ErrNotFound) {
		t.Errorf("eine erfundene Partie gab %v", erfunden)
	}

	partien, fehler := ablage.Partien(ctx)
	if fehler != nil {
		t.Fatalf("Partien: %v", fehler)
	}

	if len(partien) != 1 || partien[0].ID != "p1" {
		t.Errorf("die Liste: %+v", partien)
	}

	// **Ohne Profil geht auch die Liste nicht.**
	if _, fehler := ablage.Partien(t.Context()); fehler == nil {
		t.Error("ohne Profil kam eine Liste zurück")
	}
}
