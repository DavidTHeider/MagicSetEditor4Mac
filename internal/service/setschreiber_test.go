package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ortsschreiber merkt sich, wohin der Arbeitsort gesetzt wurde (#226).
type ortsschreiber struct {
	ort    string
	fehler error
}

func (o *ortsschreiber) HerkunftSetzen(_ context.Context, _ int64, ort string) error {
	if o.fehler != nil {
		return o.fehler
	}

	o.ort = ort

	return nil
}

// quellenDouble gibt einen Bestand samt Quelle her.
type quellenDouble struct {
	leserDouble

	quelle        *domain.Quelle
	quellenfehler error
}

func (q quellenDouble) Quelle(context.Context, int64) (*domain.Quelle, error) {
	return q.quelle, q.quellenfehler
}

// probenquelle stellt eine `.mse-set` nach, wie sie wirklich aussieht.
//
// Mit BOM, doppeltem Abstand hinter einem Doppelpunkt, einem Kommentar, einer
// Leerzeile und einem unverstandenen Block — genau das, was eine Rekonstruktion
// aus `card_values` verlöre.
//
// **Der Abstand steht auch an einem Kartenfeld**, und das ist Absicht: die
// erste Fassung hatte ihn nur an `game:`, und der Schreiber fasst nur
// Kartenfelder an. Die Gegenprobe — alle Rohzeilen verwerfen — blieb damit
// grün, weil `\tname: Feuerball` sich zufällig genau so wieder zusammensetzt.
// Ein Prüfstück, das seinen Fall nicht enthält, prüft ihn nicht.
const probenquelle = "\ufeffmse version: 2.0.0\r\n" +
	"game:  probe\n" +
	"# ein Kommentar\n" +
	"\n" +
	"set info:\n" +
	"\ttitle: Mein Probeset\n" +
	"card:\n" +
	"\tname:   Feuerball\n" +
	"unverstandener block:\n" +
	"\tirgendwas: bleibt liegen"

func mitQuelle(werte ...domain.Value) quellenDouble {
	return quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Mein Probeset", Game: "probe"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: append(
				[]domain.Value{{Field: "name", Text: "Feuerball"}}, werte...)}},
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(probenquelle)},
	}
}

// **Ein unverändertes Set kommt Byte für Byte zurück** (#48).
//
// Nicht, weil wir die Datei gut genug nachbauen — sondern weil es dieselbe ist
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
func TestEinUnveraendertesSetKommtByteFuerByteZurueck(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(mitQuelle(), &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein `.mse-set` ist ein Verzeichnis** mit der Hauptdatei `set` darin
	// (docs/FORMATE.md). Geschrieben wird die Gestalt, die gelesen wurde.
	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(gelesen) != probenquelle {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, probenquelle)
	}
}

// **Ein geänderter Wert ändert seine Zeile — und sonst nichts.**
//
// Der Kommentar bleibt, die Leerzeile bleibt, der doppelte Abstand hinter
// `game:` bleibt, der unverstandene Block bleibt. Wer stattdessen die ganze
// Datei neu zusammensetzt, hat sie angefasst, ohne dass jemand es angeordnet
// hat.
func TestEinGeaenderterWertAendertNurSeineZeile(t *testing.T) {
	bestand := mitQuelle()
	bestand.karten[0].Values[0].Text = "Eisball"

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein `.mse-set` ist ein Verzeichnis** mit der Hauptdatei `set` darin
	// (docs/FORMATE.md). Geschrieben wird die Gestalt, die gelesen wurde.
	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	erwartet := strings.Replace(probenquelle, "\tname:   Feuerball", "\tname: Eisball", 1)

	if string(gelesen) != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// Was beim Zurückschreiben schiefgehen kann, wird gemeldet.
func TestWasBeimZurueckschreibenSchiefgehtWirdGemeldet(t *testing.T) {
	kaputt := mitQuelle()
	// **Ein leerer Schlüssel**, seit #170 der Fall, der wirklich bleibt: eine
	// zu tiefe Einrückung ist kein Fehler mehr, sondern überflüssig — echte
	// Pakete rücken so ein.
	kaputt.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte("game: probe\n: ohne Schlüssel\n")}

	faelle := map[string]struct {
		bestand quellenDouble
		ziel    string
	}{
		"die Quelle lässt sich nicht lesen": {
			bestand: quellenDouble{quellenfehler: errors.New("nicht lesbar")},
			ziel:    filepath.Join(t.TempDir(), "wieder.mse-set"),
		},
		"die Quelle ist kein Baum": {
			bestand: kaputt,
			ziel:    filepath.Join(t.TempDir(), "wieder.mse-set"),
		},
		"die Karten lassen sich nicht lesen": {
			bestand: kartenfehlerBestand(),
			ziel:    filepath.Join(t.TempDir(), "wieder.mse-set"),
		},
		"die Setangaben lassen sich nicht lesen": {
			bestand: wertefehlerBestand(),
			ziel:    filepath.Join(t.TempDir(), "wieder.mse-set"),
		},
		"die Schlüsselwörter lassen sich nicht lesen": {
			bestand: woertfehlerBestand(),
			ziel:    filepath.Join(t.TempDir(), "wieder.mse-set"),
		},
		"das Ziel lässt sich nicht anlegen": {
			bestand: mitQuelle(),
			ziel:    filepath.Join(t.TempDir(), "datei", "wieder.mse-set"),
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if name == "das Ziel lässt sich nicht anlegen" {
				if err := os.WriteFile(filepath.Dir(fall.ziel), []byte("x"), 0o600); err != nil {
					t.Fatalf("schreiben: %v", err)
				}
			}

			if _, err := service.NeuerSetschreiber(fall.bestand, &ortsschreiber{}).
				Schreibe(t.Context(), 7, fall.ziel); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// woertfehlerBestand gibt einen Bestand, dessen Schlüsselwörter nicht lesbar
// sind (#1062).
//
// **Ein eigener Fehler und nicht `fehler`**: ein Double, das überall gleich
// scheitert, kann diesen Weg nicht zeigen — die Quelle und die Karten kämen
// nicht durch, und der Lauf endete vorher.
func woertfehlerBestand() quellenDouble {
	bestand := mitQuelle()
	bestand.woertfehler = errors.New("die Schlüsselwörter sind nicht lesbar")

	return bestand
}

// wertefehlerBestand gibt einen Bestand, dessen Setangaben nicht lesbar sind.
//
// **Ein eigener Fall neben den Karten** (#890): der Schreiber liest sie seit
// dem 2026-08-25 mit, und ein Lesefehler dort muss dieselbe Antwort geben —
// nicht ein Set, dem stillschweigend die Angaben fehlen.
func wertefehlerBestand() quellenDouble {
	bestand := mitQuelle()
	bestand.wertefehler = errors.New("die Setangaben sind nicht lesbar")

	return bestand
}

// kartenfehlerBestand gibt einen Bestand, dessen Karten nicht lesbar sind.
func kartenfehlerBestand() quellenDouble {
	bestand := mitQuelle()
	bestand.kartenfehler = errors.New("die Karten sind nicht lesbar")

	return bestand
}

// **Ein Feld, das nur im Bestand steht, bekommt seine Zeile.**
//
// Hier stand bis #191
// das Gegenteil: „ändert die Datei nicht — noch nicht". Die Begründung lautete,
// die Reihenfolge entscheide die Datei und nicht der Bestand. **Sie entscheidet
// beides**: die Datei sagt, wo die vorhandenen Zeilen stehen, der Bestand, wohin
// die neue gehört — hinter die ihres Vorgängers.
//
// **Geschrieben wird kanonisch**: aus dem Feld `rule text` wird die Zeile
// `rule_text:`. So schreibt MSE selbst — `canonical_name_form` wandelt beim
// Schreiben jedes Leerzeichen in einen Unterstrich (`src/util/string.cpp`,
// Zeilen 121–126, Fassung f9d9356 vom 2021-07-06), und sein Leser macht es
// wieder rückgängig. Hier stand bis
// #396
// `rule text:` — eine Schreibweise, die MSE zwar liest, aber selbst nie
// erzeugt.
func TestEinNurImBestandStehendesFeldBekommtSeineZeile(t *testing.T) {
	bestand := mitQuelle(domain.Value{Field: "rule text", Text: "Wachsamkeit"})

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	erwartet := strings.Replace(probenquelle,
		"\tname:   Feuerball\n", "\tname:   Feuerball\n\trule_text: Wachsamkeit\n", 1)

	if string(gelesen) != erwartet {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, erwartet)
	}
}

// Ein Ziel, in das sich nicht schreiben lässt, wird gemeldet.
func TestEinUnschreibbaresZielWirdBeimSetSchreibenGemeldet(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root lässt sich überall schreiben")
	}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")
	if err := os.Mkdir(ort, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(ort, 0o700); err != nil { // #nosec G302 -- Verzeichnis
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	if _, err := service.NeuerSetschreiber(mitQuelle(), &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err == nil {
		t.Error("kein Fehler")
	}
}

// Ein Kommentar in einem Kartenblock bleibt, wie er dastand.
//
// **Die Schleife über die Kinder übergeht ihn** — und das ist der Fall, den ein
// Prüfstück ohne Kommentar in einer Karte nie ausführt.
func TestEinKommentarInEinerKarteBleibt(t *testing.T) {
	mitKommentar := "game: probe\n" +
		"card:\n" +
		"\t# hierzu ist noch etwas zu sagen\n" +
		"\tname:   Feuerball\n"

	bestand := mitQuelle()
	bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(mitKommentar)}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(gelesen) != mitKommentar {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, mitKommentar)
	}
}
