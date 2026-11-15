package netz_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Was einmal geholt wurde, wird 24 Stunden lang nicht neu geholt**
// (#1002).
//
// Nachgelesen am 2026-08-23 unter `https://scryfall.com/docs/api/rate-limits`:
// „We encourage you to cache the data you download from Scryfall or process it
// locally in your own system, **at least for 24 hours**."
func TestWasFrischIstWirdNichtNeuGeholt(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), 24*time.Hour, func() time.Time { return jetzt })

	geholt := 0
	holen := func() ([]byte, error) { geholt++; return []byte(`{"a":1}`), nil }

	erst, alter, fehler := speicher.Hole(t.Context(), "karte/x", holen)
	if fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	if geholt != 1 || string(erst) != `{"a":1}` {
		t.Fatalf("erstes Holen: %d Aufrufe, %q", geholt, erst)
	}

	// **Frisch geholt heißt Alter null.**
	if alter != 0 {
		t.Errorf("das frisch Geholte ist %v alt", alter)
	}

	jetzt = jetzt.Add(23 * time.Hour)

	zweit, alter, fehler := speicher.Hole(t.Context(), "karte/x", holen)
	if fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if geholt != 1 {
		t.Errorf("es wurde %dmal geholt — innerhalb der Frist wird nicht gefragt", geholt)
	}

	if string(zweit) != `{"a":1}` {
		t.Errorf("aus dem Speicher kam %q", zweit)
	}

	// **Und der Stand sagt, wie alt er ist** (§16, Regel 5).
	if alter != 23*time.Hour {
		t.Errorf("Alter %v, erwartet waren 23h", alter)
	}
}

// **Nach der Frist wird neu geholt.**
//
// Ohne diesen Fall wäre auch ein Speicher grün, der nie wieder fragt — und der
// zeigte in einem Jahr noch den Stand von heute.
func TestNachDerFristWirdNeuGeholt(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), 24*time.Hour, func() time.Time { return jetzt })

	inhalt := `{"a":1}`
	geholt := 0
	holen := func() ([]byte, error) { geholt++; return []byte(inhalt), nil }

	if _, _, fehler := speicher.Hole(t.Context(), "karte/x", holen); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	jetzt = jetzt.Add(25 * time.Hour)
	inhalt = `{"a":2}`

	neu, alter, fehler := speicher.Hole(t.Context(), "karte/x", holen)
	if fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if geholt != 2 {
		t.Errorf("es wurde %dmal geholt — nach der Frist wird gefragt", geholt)
	}

	if string(neu) != `{"a":2}` || alter != 0 {
		t.Errorf("es kam %q mit Alter %v", neu, alter)
	}
}

// **Und wenn das Holen scheitert, bleibt der letzte gute Stand** (§16, Regel 5).
//
// # Warum das die wichtigste Zusage dieses Speichers ist
//
// „Ohne Netz arbeitet die Anwendung weiter. Ein Dienst, dessen Ausfall das
// Bearbeiten anhält, ist falsch eingebaut. Der letzte gute Stand bleibt, und
// er sagt, wie alt er ist."
func TestBeiEinemFehlerBleibtDerLetzteGuteStand(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), time.Hour, func() time.Time { return jetzt })

	if _, _, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return []byte(`{"gut":true}`), nil }); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	jetzt = jetzt.Add(48 * time.Hour)

	alt, alter, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return nil, os.ErrDeadlineExceeded })
	if fehler != nil {
		t.Fatalf("der Ausfall hat den Aufruf scheitern lassen: %v", fehler)
	}

	if string(alt) != `{"gut":true}` {
		t.Errorf("der letzte gute Stand kam nicht: %q", alt)
	}

	if alter != 48*time.Hour {
		t.Errorf("Alter %v, erwartet waren 48h — er soll sagen, wie alt er ist", alter)
	}
}

// **Ohne Netz und ohne Stand scheitert es — und sagt beides.**
//
// Die Gegenprobe zum Fall darüber: ohne sie wäre auch ein Speicher grün, der
// bei einem Ausfall stillschweigend nichts zurückgibt.
func TestOhneStandUndOhneNetzScheitertEs(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), time.Hour, func() time.Time { return jetzt })

	_, _, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return nil, os.ErrDeadlineExceeded })
	if fehler == nil {
		t.Fatal("ohne Stand und ohne Netz kam etwas zurück")
	}

	if !strings.Contains(fehler.Error(), "keinen Stand") {
		t.Errorf("die Meldung sagt nicht, dass auch kein Stand da ist: %v", fehler)
	}
}

// **Ein abgebrochenes Schreiben lässt den alten Stand stehen.**
//
// # Warum das nicht nur Sorgfalt ist
//
// Ein halb geschriebener Zwischenspeicher ist schlimmer als keiner: er sieht
// aus wie ein gültiger Stand und ist keiner. Geschrieben wird deshalb
// **daneben** und dann umbenannt — auf demselben Dateisystem ist das atomar.
func TestEinHalbGeschriebenerStandEntstehtNicht(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(ort, time.Hour, func() time.Time { return jetzt })

	if _, _, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return []byte("gut"), nil }); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	// **Kein Zwischenstand bleibt liegen.** Ein `.teil` neben der Datei wäre
	// der Rückstand eines abgebrochenen Schreibens.
	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		if strings.Contains(eintrag.Name(), "teil") || strings.HasSuffix(eintrag.Name(), "~") {
			t.Errorf("ein Zwischenstand blieb liegen: %s", eintrag.Name())
		}
	}
}

// **Ein Schlüssel wird zu einem Dateinamen, der nichts kaputt macht.**
//
// # Warum das hier steht
//
// Ein Schlüssel kommt aus einer Adresse — `cards/named?exact=Lightning Bolt`.
// Wer den ungeprüft als Dateinamen nähme, schriebe bei `../` außerhalb seines
// Verzeichnisses. Das ist dieselbe Klasse wie das Auspacken in
// `internal/bezug`.
func TestEinSchluesselKommtNichtAusSeinemVerzeichnisHeraus(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(ort, time.Hour, func() time.Time { return jetzt })

	// **Ohne `t.Parallel()` in den Unterfällen**, und das ist der ganze Punkt.
	//
	// Der erste Anlauf hatte es, und die Prüfung darunter lief **vor** ihnen:
	// parallele Unterfälle beginnen erst, wenn der äußere Rumpf durch ist. Der
	// Test war grün, auch als der Pfad ohne Prüfsumme gebaut wurde und
	// `../ausbruch` wirklich ausbrach — genau die „Zusicherung, die schon vor
	// der Antwort zutrifft" aus AGENTS.md §3.
	for _, schluessel := range []string{
		"../ausbruch", "a/../../ausbruch", "cards/named?exact=Lightning Bolt", "/absolut",
	} {
		if _, _, fehler := speicher.Hole(t.Context(), schluessel,
			func() ([]byte, error) { return []byte("x"), nil }); fehler != nil {
			t.Fatalf("holen (%q): %v", schluessel, fehler)
		}
	}

	// **Nichts liegt außerhalb.** Der Elternordner enthält nur den
	// Speicherort selbst.
	eltern := filepath.Dir(ort)

	eintraege, err := os.ReadDir(eltern)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		if strings.Contains(eintrag.Name(), "ausbruch") {
			t.Errorf("ein Schlüssel ist ausgebrochen: %s", filepath.Join(eltern, eintrag.Name()))
		}
	}
}

// **Ein kaputter Eintrag gilt als keiner.**
//
// # Warum er nicht gemeldet wird
//
// Er wird beim nächsten geglückten Holen überschrieben. Eine Meldung darüber
// hülfe niemandem: der Aufrufer hat nichts falsch gemacht, und tun kann er
// nichts — außer dem, was ohnehin geschieht.
func TestEinKaputterEintragGiltAlsKeiner(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(ort, time.Hour, func() time.Time { return jetzt })

	if _, _, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return []byte("gut"), nil }); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	// Den geschriebenen Eintrag verderben.
	eintraege := nurEintraege(t, ort)
	if len(eintraege) != 1 {
		t.Fatalf("es liegen %d Einträge", len(eintraege))
	}

	if err := os.WriteFile(filepath.Join(ort, eintraege[0]),
		[]byte("das ist kein JSON"), 0o600); err != nil {
		t.Fatalf("verderben: %v", err)
	}

	geholt := 0

	inhalt, _, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { geholt++; return []byte("wieder gut"), nil })
	if fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if geholt != 1 {
		t.Error("der kaputte Eintrag wurde für gültig gehalten")
	}

	if string(inhalt) != "wieder gut" {
		t.Errorf("es kam %q", inhalt)
	}
}

// **Ein Ort, an den nicht geschrieben werden kann, macht die Antwort nicht
// kaputt.**
//
// # Warum das die richtige Seite des Handels ist
//
// Die Antwort ist da. Dass sie nicht auf die Platte kam, kostet die nächste
// Anfrage einen Netzgang — mehr nicht. Sie deshalb scheitern zu lassen hieße,
// einen Zwischenspeicher zur Voraussetzung zu machen, und er ist eine
// Bequemlichkeit.
//
// **Als `root` übersprungen**: dort greift keine Rechteprüfung.
func TestEinUnbeschreibbarerOrtMachtDieAntwortNichtKaputt(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("als root greift keine Rechteprüfung")
	}

	verschlossen := filepath.Join(t.TempDir(), "zu")
	if err := os.Mkdir(verschlossen, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() {
		// #nosec G302 -- sonst räumt t.TempDir() nicht auf
		if err := os.Chmod(verschlossen, 0o700); err != nil {
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(filepath.Join(verschlossen, "speicher"),
		time.Hour, func() time.Time { return jetzt })

	inhalt, alter, fehler := speicher.Hole(t.Context(), "karte/x",
		func() ([]byte, error) { return []byte("gut"), nil })
	if fehler != nil {
		t.Fatalf("das Holen scheiterte am Speicherort: %v", fehler)
	}

	if string(inhalt) != "gut" || alter != 0 {
		t.Errorf("es kam %q mit Alter %v", inhalt, alter)
	}
}

// **Beide Fehlerwege des Schreibens werden betreten.**
//
// # Wie sie hergestellt werden
//
// Ein **Verzeichnis** am Zwischennamen lässt das Schreiben scheitern, eines am
// **Ziel** das Umbenennen. Beides sind Dateisystemzustände, die sich herstellen
// lassen — und genau deshalb entsteht der Zwischenname aus dem Ziel und nicht
// aus `os.CreateTemp`.
//
// **Geprüft wird, dass die Antwort trotzdem ankommt.** Ein Zwischenspeicher
// ist eine Bequemlichkeit; sein Scheitern darf den Aufruf nicht scheitern
// lassen.
func TestBeideFehlerwegeDesSchreibensSindHarmlos(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"ein Verzeichnis am Zwischennamen": ".teil",
		"ein Verzeichnis am Ziel":          "",
	}

	for name, endung := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := t.TempDir()
			jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
			speicher := netz.NeuerSpeicher(ort, time.Hour, func() time.Time { return jetzt })

			// **Den Pfad so bilden wie der Speicher.** Er ist eine Prüfsumme;
			// ihn hier nachzurechnen ist die einzige Art, das Verzeichnis an
			// die richtige Stelle zu setzen.
			im.Weg(t, ort, "karte/x", endung)

			inhalt, alter, fehler := speicher.Hole(t.Context(), "karte/x",
				func() ([]byte, error) { return []byte("gut"), nil })
			if fehler != nil {
				t.Fatalf("das Holen scheiterte am Speicher: %v", fehler)
			}

			if string(inhalt) != "gut" || alter != 0 {
				t.Errorf("es kam %q mit Alter %v", inhalt, alter)
			}
		})
	}
}

// nurEintraege gibt die Zwischenspeichereinträge — **ohne** die Nachbardateien.
//
// **`.laedt` gehört nicht dazu**: sie ist die Absprache darüber, wer holt
// (#1163),
// und sie bleibt liegen. Eine Sperrdatei zu löschen, während ein anderer sie
// hält, gäbe jedem einen eigenen Inode.
func nurEintraege(t *testing.T, ort string) []string {
	t.Helper()

	gelesen, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	var namen []string

	for _, eintrag := range gelesen {
		if !strings.HasPrefix(eintrag.Name(), ".") {
			namen = append(namen, eintrag.Name())
		}
	}

	return namen
}
