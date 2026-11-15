package app

// **Der Start liest das gezeigte Profil** — Werkstück 2 von Phase 4.12
// (#1351).
//
// **Paketintern**, weil `profil` und `profilLaden` es sind: von außen ließe
// sich nur beobachten, dass Abfragen glücken — und das prüfen die Bindungen
// ohnehin. Hier steht, **welche** Zahl ankommt.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Nach dem Start trägt jeder Kontext das gezeigte Profil.**
func TestDerStartLiestDenProfilzeiger(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	gesehen, fehler := storage.ProfilAus(anwendung.imProfil(t.Context()))
	if fehler != nil {
		t.Fatalf("das Profil aus dem Kontext: %v", fehler)
	}

	if gesehen != 1 {
		t.Errorf("Profil %d, erwartet war das erste", gesehen)
	}
}

// **Ein Zeiger, der sich nicht lesen lässt, hält den Start nicht auf.**
//
// Er ist eine Bequemlichkeit: [storage.Nutzerablage.Profilzeiger] fällt auf
// das erste Profil zurück, und gibt es gar keines, sagt das die erste Abfrage
// deutlicher als eine Meldung beim Anlaufen. Was bleibt, ist die Null — und
// die überschreibt [App.imProfil] ausdrücklich **nicht**.
func TestEinUnlesbarerZeigerHaeltDenStartNichtAuf(t *testing.T) {
	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	anwendung := &App{db: db}
	anwendung.profilLaden(t.Context())

	if gesehen := anwendung.profil.Load(); gesehen != 0 {
		t.Errorf("Profil %d, erwartet war die Null", gesehen)
	}

	// **Und der nackte Kontext bleibt nackt.** Eine Null einzuhängen sähe aus
	// wie ein Profil und wäre keines.
	if _, profilfehler := storage.ProfilAus(
		anwendung.imProfil(t.Context())); profilfehler == nil {
		t.Error("der Kontext trug ein Profil, das es nicht gibt")
	}
}

// **Ein Kontext, der schon eines trägt, behält es.**
//
// Der Fall der drei Flächen ohne Fenster: ihr Wurzelkontext ist ausgestattet
// ([storage.MitGezeigtemProfil]), und `a.profil` ist bei ihnen nie gesetzt.
func TestEinAusgestatteterKontextBehaeltSeinProfil(t *testing.T) {
	anwendung := &App{}

	ausgestattet := storage.MitProfil(t.Context(), 7)

	gesehen, fehler := storage.ProfilAus(anwendung.imProfil(ausgestattet))
	if fehler != nil {
		t.Fatalf("das Profil aus dem Kontext: %v", fehler)
	}

	if gesehen != 7 {
		t.Errorf("Profil %d, erwartet war 7", gesehen)
	}
}

// **Ein Profilwechsel räumt die gezeichneten Spielbilder weg.**
//
// # Warum Speicher der Grund ist und nicht Verwechslung
//
// **Weil ein falsches Bild gar nicht herauskäme.** Der Schlüssel steht über
// dem **Inhalt** — dieselbe Karte in einem anderen Profil ist ein anderer
// Schlüssel, und eine Adresse aus dem alten Profil trifft im neuen nichts.
//
// **Was bliebe, ist der Platz.** Das Fach hält bis zu 128 MB, und über den
// Wechsel hinweg hielte es sie für Karten, die niemand mehr sieht. #1459
// nennt den Profilwechsel neben dem Speicherbudget, und das ist derselbe
// Satz.
//
// **Wer eine Verwechslung als Grund schriebe, hätte eine Begründung, die
// nicht trägt** (§10) — und beim nächsten Umbau des Schlüssels fiele sie
// unbemerkt weg.
func TestEinProfilwechselRaeumtDieSpielbilder(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	fach := vorschauweg.NeuesTexturfach(1 << 20)
	fach.Hinterlege("abc", []byte("ein Bild"), 375, 523)

	anwendung.spielbilder = fach

	if fach.Belegt() == 0 {
		t.Fatal("das Fach ist schon leer — dieser Lauf misst nichts")
	}

	anwendung.spielbilderRaeumen()

	if fach.Belegt() != 0 {
		t.Errorf("%d Bytes bleiben liegen", fach.Belegt())
	}
}

// **Und eine Anwendung ohne Fach stürzt darüber nicht.**
//
// **Jede Fläche ohne Fenster hat keines**: `mse`, `mse-mcp` und `mse-rest`
// bauen kein Zeichenwerk, und ein Profilwechsel geschieht dort trotzdem.
func TestEinProfilwechselOhneFachGehtDurch(t *testing.T) {
	NeueApp(filepath.Join(t.TempDir(), "bestand.db")).spielbilderRaeumen()
}
