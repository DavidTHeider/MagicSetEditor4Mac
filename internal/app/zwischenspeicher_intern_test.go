package app

// **Begrenzte, beobachtbare und richtig geleerte Zwischenspeicher** — OPS-08
// (#1725).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Zwischenspeicher liegt, wo macOS ihn erwartet** (AK-03): unter
// `Caches` und nicht neben dem Bestand — wer den Ordner der Anwendung
// ansieht, soll ihn nicht für Daten halten, und eine Sicherung nimmt ihn
// nicht mit.
func TestDerZwischenspeicherLiegtNichtNebenDemBestand(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	bestand, _ := Datenbankpfad() //nolint:errcheck // dasselbe Zuhause wie eine Zeile darüber
	if !strings.HasPrefix(ort, filepath.Join(heim, "Library", "Caches")) ||
		strings.HasPrefix(ort, filepath.Dir(bestand)) {
		t.Errorf("%q neben %q", ort, bestand)
	}
}

// **Ein alter Speicher neben dem Bestand zieht einmal um** — die Antworten
// und Bilder, die schon da sind, gehen nicht verloren.
func TestEinAlterZwischenspeicherZiehtUm(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)
	t.Setenv(Bestandsumgebung, "")

	bestand, fehler := Datenbankpfad()
	if fehler != nil {
		t.Fatal(fehler)
	}

	alt := filepath.Join(filepath.Dir(bestand), "scryfall")
	if anlegefehler := os.MkdirAll(filepath.Join(alt, "bilder"), 0o700); anlegefehler != nil {
		t.Fatal(anlegefehler)
	}

	if schreibfehler := os.WriteFile(filepath.Join(alt, "bilder", "karte.png"), []byte("png"), 0o600); schreibfehler != nil {
		t.Fatal(schreibfehler)
	}

	ort, fehler := scryfallverzeichnis(true)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ort, "bilder", "karte.png")); fehler != nil {
		t.Errorf("das Bild zog nicht mit: %v", fehler)
	}

	if _, fehler := os.Stat(alt); !os.IsNotExist(fehler) {
		t.Errorf("der alte Ordner steht noch: %v", fehler)
	}
}

// **Ein umgelenkter Bestand zieht nicht um** — eine Probe soll den Speicher
// der installierten Anwendung nicht wegtragen.
func TestEinUmgelenkterBestandZiehtNichtUm(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)
	t.Setenv(Bestandsumgebung, filepath.Join(heim, "probe", "bestand.db"))

	bestand, fehler := Datenbankpfad()
	if fehler != nil {
		t.Fatal(fehler)
	}

	alt := filepath.Join(filepath.Dir(bestand), "scryfall")
	if fehler := os.MkdirAll(alt, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := scryfallverzeichnis(true); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(alt); fehler != nil {
		t.Errorf("der alte Ordner ist fort: %v", fehler)
	}

	// **Und ein Testlauf zieht nie um** — auch ohne umgelenkten Bestand.
	t.Setenv(Bestandsumgebung, "")

	if umzugErlaubt {
		t.Error("ein Testlauf darf umziehen")
	}

	if _, fehler := Scryfallverzeichnis(); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(alt); fehler != nil {
		t.Errorf("der Testlauf trug den alten Ordner weg: %v", fehler)
	}
}

// **Die Auskunft nennt jeden Speicher mit Ort, Zahlen und Grenze** (AK-01),
// und **Leeren leert alle** (AK-03) — die Platte und den Arbeitsspeicher.
func TestDieZwischenspeicherLassenSichSehenUndLeeren(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, nil)

	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	antwort := filepath.Join(ort, strings.Repeat("a", 64)+".json")
	if fehler := os.WriteFile(antwort, []byte("{}"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung.bildfach.Hinterlege([]byte("vorschau"))

	stand := anwendung.Zwischenspeicher(t.Context())
	if stand.Fehler != nil {
		t.Fatalf("%+v", stand.Fehler)
	}

	namen := map[string]ZwischenspeicherDTO{}
	for _, eintrag := range stand.Speicher {
		namen[eintrag.Name] = eintrag
	}

	for _, name := range []string{
		SpeicherAntworten, SpeicherBilder, SpeicherSammeldaten, SpeicherEbenen,
		SpeicherPaketbilder, SpeicherSkripte, SpeicherVorschau, SpeicherTexturen,
	} {
		if _, da := namen[name]; !da {
			t.Errorf("%s fehlt in %+v", name, stand.Speicher)
		}
	}

	// **Und die beiden Speicher, die bis PERF03 nur zählten, nennen ihre
	// Bytegrenze** (#2353, Befund A-19).
	if namen[SpeicherSkripte].GrenzeBytes != msescript.GrenzeBaumbytes ||
		namen[SpeicherPaketbilder].GrenzeBytes != msepaket.GrenzeBildbytes {
		t.Errorf("ohne Bytegrenze: %+v, %+v", namen[SpeicherSkripte], namen[SpeicherPaketbilder])
	}

	if namen[SpeicherAntworten].Eintraege != 1 || namen[SpeicherAntworten].Ort != OrtPlatte ||
		namen[SpeicherVorschau].Eintraege != 1 || namen[SpeicherVorschau].Ort != OrtArbeitsspeicher {
		t.Errorf("%+v, %+v", namen[SpeicherAntworten], namen[SpeicherVorschau])
	}

	geleert := anwendung.ZwischenspeicherLeeren(t.Context())
	if geleert.Fehler != nil {
		t.Fatalf("%+v", geleert.Fehler)
	}

	if _, fehler := os.Stat(antwort); !os.IsNotExist(fehler) || anwendung.bildfach.Zahlen().Eintraege != 0 {
		t.Errorf("nach dem Leeren: %v, %+v", fehler, anwendung.bildfach.Zahlen())
	}

	// **Die Antwort sagt, was vorher belegt war** — wer leert, soll sehen,
	// was es gebracht hat.
	if len(geleert.Speicher) != len(stand.Speicher) {
		t.Errorf("%+v", geleert.Speicher)
	}
}

// **Ein Profilwechsel leert die Vorschauen** (AK-02) — dieselbe Regel wie für
// die Spielbilder: sie gehören dem Profil, das sie gezeichnet hat.
func TestEinProfilwechselLeertDieVorschauen(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, nil)
	anwendung.bildfach.Hinterlege([]byte("vorschau"))

	anwendung.profilUmstellen(t.Context(), anwendung.profil.Load())

	if zahl := anwendung.bildfach.Zahlen().Eintraege; zahl != 0 {
		t.Errorf("%d Vorschauen nach dem Wechsel", zahl)
	}
}

// **Die Grenze auf der Platte gilt ab dem Start** — die ältesten Bilder
// gehen, bis sie eingehalten ist.
func TestDieGrenzeGiltAbDemStart(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, map[string]string{domain.ZwischenspeicherGrenze: "1"})

	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	bilder := filepath.Join(ort, "bilder")
	if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	for stelle, name := range []string{"alt.png", "neu.png"} {
		pfad := filepath.Join(bilder, name)
		if fehler := os.WriteFile(pfad, make([]byte, 700<<10), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		zeit := time.Now().Add(time.Duration(stelle-2) * time.Hour)
		if fehler := os.Chtimes(pfad, zeit, zeit); fehler != nil {
			t.Fatal(fehler)
		}
	}

	// **Über einen echten Neustart** — ein Ruf auf die Pflege selbst bliebe
	// grün, auch wenn der Start sie nicht riefe.
	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(filepath.Join(bilder, "alt.png")); !os.IsNotExist(fehler) {
		t.Errorf("das ältere Bild steht noch: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(bilder, "neu.png")); fehler != nil {
		t.Errorf("das neuere ist fort: %v", fehler)
	}
}

// **Ein Neubeginn leert die Vorschauen** — nach einem Zurückspielen startet
// die Anwendung über denselben Weg neu, und dann ist es ein anderer Bestand.
func TestEinNeubeginnLeertDieVorschauen(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, nil)
	anwendung.bildfach.Hinterlege([]byte("vorschau"))

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	if zahl := anwendung.bildfach.Zahlen().Eintraege; zahl != 0 {
		t.Errorf("%d Vorschauen nach dem Neubeginn", zahl)
	}
}

// **Ohne Ort für Zwischenspeicher gibt es keine Auskunft** — und einen Satz
// statt einer leeren Liste, die aussähe wie ein leerer Speicher.
func TestOhneOrtGibtEsKeineAuskunft(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, nil)

	datei := filepath.Join(t.TempDir(), "keinzuhause")
	if fehler := os.WriteFile(datei, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	t.Setenv("HOME", datei)

	if antwort := anwendung.Zwischenspeicher(t.Context()); antwort.Fehler == nil || antwort.Speicher == nil {
		t.Errorf("Auskunft: %+v", antwort)
	}

	if antwort := anwendung.ZwischenspeicherLeeren(t.Context()); antwort.Fehler == nil {
		t.Errorf("Leeren: %+v", antwort)
	}

	// **Und die Pflege beim Start schweigt** — sie hält nichts auf.
	anwendung.zwischenspeicherPflegen(t.Context())
}

// **Ohne Einstellungen gilt die Vorgabe** — eine Anwendung ohne Bestand und
// eine, deren Bestand nicht antwortet.
func TestOhneEinstellungenGiltDieVorgabe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if grenze := NeueApp("").zwischenspeichergrenze(t.Context()); grenze != domain.ZwischenspeichergrenzeAus("") {
		t.Errorf("ohne Bestand: %d", grenze)
	}

	anwendung, _ := gestartetMitWeiche(t, map[string]string{domain.ZwischenspeicherGrenze: "64"})
	ctx := anwendung.imProfil(t.Context())

	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if grenze := anwendung.zwischenspeichergrenze(ctx); grenze != domain.ZwischenspeichergrenzeAus("") {
		t.Errorf("ohne Antwort: %d", grenze)
	}
}

// **Ohne Ort für Zwischenspeicher startet die Anwendung trotzdem** — Scryfall
// fehlt dann, und die Wege, die ihn brauchen, sagen es. Hier stand bis OPS-08
// ein Test, der denselben Ausfall über den Ordner des Bestands herbeiführte;
// seit der Speicher unter `Caches` liegt, trifft er ihn nicht mehr.
func TestEinStartOhneZwischenspeicherortGehtDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	sperre := filepath.Join(zuhause, "Library", "Caches", "MagicSetEditor4Mac")
	if fehler := os.MkdirAll(filepath.Dir(sperre), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(sperre, []byte("kein Verzeichnis"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Error(fehler)
		}
	})

	if _, fehler := Scryfallverzeichnis(); fehler == nil {
		t.Error("ein versperrter Ort gab einen Speicher")
	}
}

// **Das Fenster kennt die Spur, wie sie eingestellt ist** — Befund 8 des
// Zweitblicks über Phase 6.4: es sagte immer „aus", weil nur Dienst und
// Werkzeugserver sie beim Start einrichteten.
func TestDasFensterKenntDieEingestellteSpur(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	gestartetMitWeiche(t, map[string]string{domain.MessungSpur: "an"})

	if !service.Spurstand().An {
		t.Error("die Spur heißt am Fenster aus")
	}
}
