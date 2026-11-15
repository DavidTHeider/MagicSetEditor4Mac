package msepaket

// **Die Fehlerwege des Archivanlegens** — Werkstück C von Phase 4.9
// (#1324).
//
// # Warum das ein interner Test ist
//
// **Weil ein Schreibfehler auf einer gesunden Platte nicht zu bestellen ist.**
// Dieselbe Lage und dieselbe Antwort wie bei [nachbardatei] nebenan: die
// Griffe sind austauschbar, und ein Lauf setzt einen ein, der scheitert. Ohne
// diese Naht bliebe „das abgebrochene Anlegen" ungeprüft (§3).

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

// stolperziel scheitert beim Schreiben, beim Schließen oder bei beidem.
type stolperziel struct {
	beimSchreiben  error
	beimSchliessen error
	geschlossen    bool
}

func (s *stolperziel) Write(p []byte) (int, error) {
	if s.beimSchreiben != nil {
		return 0, s.beimSchreiben
	}

	return len(p), nil
}

func (s *stolperziel) Close() error {
	s.geschlossen = true

	return s.beimSchliessen
}

var errStolper = errors.New("planmäßig gescheitert")

func TestEinAbgebrochenesAnlegenRaeumtAuf(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")
	gestolpert := &stolperziel{beimSchreiben: errStolper}

	weggeraeumt := ""

	err := archivAnlegen(ziel, map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen: func(string) (archivziel, error) { return gestolpert, nil },
			wegraeumen: func(ort string) error {
				weggeraeumt = ort

				return nil
			},
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}

	// **Ein halbes Archiv bleibt nicht liegen** (§8): wer eines fände, hielte
	// es für ein ganzes.
	if weggeraeumt != ziel+".tmp" {
		t.Errorf("weggeräumt wurde %q, erwartet %q", weggeraeumt, ziel+".tmp")
	}

	if !gestolpert.geschlossen {
		t.Error("der Griff blieb offen")
	}
}

// **Ein Schreibfehler mitten im Packen wird gemeldet.**
//
// # Warum der Inhalt groß sein muss
//
// **`zip` puffert.** Ein kleiner Eintrag landet im Puffer des Verdichters und
// erreicht das Ziel erst beim Abschluss — dann meldet sich der Fehler beim
// Schließen und nicht beim Schreiben, und der Zweig hier bliebe unbetreten.
//
// **Und der Inhalt ist wirklich zufällig.** Der erste Anlauf nahm ein
// gerechnetes Muster — `zip` verdichtete das eine Megabyte auf wenige
// Kilobytes, und die blieben im Puffer. Ein unverdichtbarer Strom fließt
// durch; der Samen steht fest, damit der Lauf sich wiederholen lässt.
func TestEinSchreibfehlerMittenImPackenWirdGemeldet(t *testing.T) {
	t.Parallel()

	gross := make([]byte, 1<<20)

	wuerfel := rand.New(rand.NewPCG(4, 9)) // #nosec G404 -- Prüfdaten, kein Geheimnis
	for i := range gross {
		gross[i] = byte(wuerfel.UintN(256))
	}

	err := archivAnlegen(
		filepath.Join(t.TempDir(), "Probe.mse-set"),
		map[string][]byte{"set": gross},
		archivhilfen{
			anlegen: func(string) (archivziel, error) {
				return &stolperziel{beimSchreiben: errStolper}, nil
			},
			wegraeumen: func(string) error { return nil },
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}
}

// **Auch ein Fehler beim Schließen wird gemeldet.**
//
// **Und zwar erst dort**: ein Packer schreibt gepuffert, und was beim
// Abschluss schiefgeht, erfährt man beim Schließen — nicht beim Schreiben.
func TestEinFehlerBeimSchliessenWirdGemeldet(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	err := archivAnlegen(ziel, map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen: func(string) (archivziel, error) {
				return &stolperziel{beimSchliessen: errStolper}, nil
			},
			wegraeumen: func(string) error { return nil },
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}
}

// **Und ein Ziel, das sich gar nicht anlegen lässt.**
func TestEinZielDasSichNichtAnlegenLaesst(t *testing.T) {
	t.Parallel()

	err := archivAnlegen(
		filepath.Join(t.TempDir(), "Probe.mse-set"),
		map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen:    func(string) (archivziel, error) { return nil, errStolper },
			wegraeumen: func(string) error { return nil },
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}
}

// **Und ein Fehler beim Schließen lässt das Ziel nicht liegen.**
//
// Befund 5 des Zweitblicks über die Phasen 4.0 bis 4.9. Der Zweig meldete den
// Fehler und ließ die Datei stehen — während der Kommentar zwei Zeilen
// darunter zusagt, ein halbes Archiv bleibe nicht liegen. **Ein Kopf, der
// mehr verspricht als sein Rumpf hält** (§10).
//
// **Was daran materiell ist**: nach einer gemeldeten fehlgeschlagenen Ausfuhr
// liegt am endgültigen Namen ein `.mse-set`, das niemand geschlossen hat —
// und das später für ein fertiges gehalten wird.
func TestEinFehlerBeimSchliessenRaeumtDasZielWeg(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	weggeraeumt := ""

	err := archivAnlegen(ziel, map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen: func(string) (archivziel, error) {
				return &stolperziel{beimSchliessen: errStolper}, nil
			},
			wegraeumen: func(pfad string) error {
				weggeraeumt = pfad

				return nil
			},
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}

	// **Weggeräumt wird die Datei daneben** — seit dem 2026-09-03 schreibt
	// der Lauf nach `<ziel>.tmp` und tauscht erst am Ende (Befund 2 des
	// Nachlaufs). Am Ziel selbst gibt es nichts wegzuräumen; das ist der
	// Ertrag der Änderung und nicht ihr Nebeneffekt.
	if weggeraeumt != ziel+".tmp" {
		t.Errorf("weggeräumt wurde %q, erwartet war %q", weggeraeumt, ziel+".tmp")
	}
}

// **Ein vorhandenes Archiv überlebt einen gescheiterten Lauf** — Befund 2 des
// Nachlaufs über die Phasen 4.0 bis 4.9.
//
// # Was der erste Anlauf getroffen hat, und was nicht
//
// Er räumte nach einem gescheiterten `Close` das Ziel weg. **Das ist zu
// wenig**: geöffnet wurde unmittelbar am endgültigen Namen, und `os.Create`
// schneidet eine vorhandene Datei zuvor ab. Wer über ein fertiges Archiv
// schreibt und dabei scheitert, verliert es — der Fehlerweg löscht danach
// auch noch den Rest.
//
// # Wie MSE2 es macht
//
//	String tempFile = saveAs + _(".tmp");
//	…
//	if (wxFileExists(saveAs)) {
//	  remove_file(saveAs + _(".bak"));
//	  wxRenameFile(saveAs, saveAs + _(".bak"));
//	}
//	wxRenameFile(tempFile, saveAs);
//
// `src/util/io/package.cpp`, Zeilen 446–489, Fassung f9d9356 vom 2021-07-06.
// Es schreibt neben das Ziel und **tauscht erst nach dem Gelingen**.
//
// **Hier ohne `.bak`**: MSE2 hält damit die vorige Fassung eines Sets, an dem
// jemand arbeitet. Dieses Archiv ist eine **Ausfuhr** — was vorher dalag,
// gehört dem, der es dorthin gelegt hat, und eine stille zweite Datei daneben
// wäre eine Zugabe, um die niemand gebeten hat.
func TestEinVorhandenesArchivUeberlebtEinenFehlschlag(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	if err := os.WriteFile(ziel, []byte("das alte Archiv"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	err := archivAnlegen(ziel, map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen: func(pfad string) (archivziel, error) {
				// **Der Lauf schreibt woanders hin.** Ohne diese Zusicherung
				// wäre der Test auch dann grün, wenn das Aufräumen nur
				// zufällig gelänge.
				if pfad == ziel {
					t.Errorf("geschrieben wird nach %q — das ist das Ziel selbst",
						pfad)
				}

				return &stolperziel{beimSchliessen: errStolper}, nil
			},
			wegraeumen: os.Remove,
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}

	inhalt, lesefehler := os.ReadFile(ziel) //nolint:gosec // Wegwerfverzeichnis
	if lesefehler != nil {
		t.Fatalf("das alte Archiv ist fort: %v", lesefehler)
	}

	if string(inhalt) != "das alte Archiv" {
		t.Errorf("das alte Archiv lautet %q", inhalt)
	}
}

// **Und ein geglückter Lauf tauscht es aus.**
//
// Ohne diesen Fall hieße die Korrektur „es wird nie geschrieben" — die
// Gegenprobe zur Zusicherung darüber.
func TestEinGeglueckterLaufErsetztDasVorhandeneArchiv(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	if err := os.WriteFile(ziel, []byte("das alte Archiv"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := ArchivAnlegen(ziel, map[string][]byte{"set": []byte("x")}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt, lesefehler := os.ReadFile(ziel) //nolint:gosec // Wegwerfverzeichnis
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(inhalt) == "das alte Archiv" {
		t.Error("das alte Archiv steht noch da")
	}

	// **Und nichts daneben.** Eine `.tmp`, die liegen bleibt, ist derselbe
	// halbe Stand an einem anderen Namen.
	eintraege, err := os.ReadDir(filepath.Dir(ziel))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 1 {
		namen := make([]string, 0, len(eintraege))
		for _, eintrag := range eintraege {
			namen = append(namen, eintrag.Name())
		}

		t.Errorf("im Verzeichnis liegen %v, erwartet nur das Archiv", namen)
	}
}

// **Ein gescheiterter Tausch räumt die Datei daneben weg.**
//
// Der letzte Schritt kann ebenso scheitern wie die davor — ein volles
// Dateisystem, ein Ziel, das jemand anderes hält. Bliebe die `.tmp` liegen,
// wäre das genau der halbe Stand an einem anderen Namen (§8).
func TestEinGescheiterterTauschRaeumtDieDateiDanebenWeg(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	weggeraeumt := ""

	err := archivAnlegen(ziel, map[string][]byte{"set": []byte("x")},
		archivhilfen{
			anlegen:    func(string) (archivziel, error) { return &stolperziel{}, nil },
			wegraeumen: func(pfad string) error { weggeraeumt = pfad; return nil },
			tauschen:   func(string, string) error { return errStolper },
		})

	if !errors.Is(err, errStolper) {
		t.Fatalf("der Fehler ist %v, erwartet den gestellten", err)
	}

	if weggeraeumt != ziel+".tmp" {
		t.Errorf("weggeräumt wurde %q, erwartet %q", weggeraeumt, ziel+".tmp")
	}
}
