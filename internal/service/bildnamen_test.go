package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// karten baut einen Bestand aus Kartennamen.
func karten(namen ...string) leserDouble {
	bestand := leserDouble{koepfe: []domain.SetHead{{ID: 7, Name: "Probeset"}}}

	for nummer, name := range namen {
		werte := []domain.Value{}
		if name != "" {
			werte = append(werte, domain.Value{Field: "name", Text: name})
		}

		bestand.karten = append(bestand.karten, &domain.CardRow{
			ID: int64(nummer + 1), Number: nummer + 1, Values: werte,
		})
	}

	return bestand
}

// namen ruft die Ausfuhr und gibt nur die Dateinamen.
func namen(t *testing.T, bestand leserDouble, vorlage string) []string {
	t.Helper()

	gefunden, err := service.NeueBildausfuhr(kartenDouble{leserDouble: bestand}, nil, nil).
		Namen(t.Context(), 7, vorlage, service.FormatPNG, domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Namen: %v", err)
	}

	nur := make([]string, 0, len(gefunden))
	for _, name := range gefunden {
		nur = append(nur, name.Datei)
	}

	return nur
}

// **Der Dateiname folgt einer Vorlage, nicht einer Regel im Code.**
//
// Wer 300 Karten ausgibt, will sie hinterher wiederfinden — und zwar in der
// Ordnung, die er selbst wählt. Die Nummer steht dreistellig, damit die
// alphabetische Ordnung des Dateibrowsers die des Sets ist.
func TestBildnamenFolgenDerVorlage(t *testing.T) {
	faelle := []struct {
		name     string
		vorlage  string
		erwartet []string
	}{
		{
			name:     "Nummer und Name",
			vorlage:  "{nummer}-{name}",
			erwartet: []string{"001-Feuerball.png", "002-Eisball.png"},
		},
		{
			name:     "nur der Name",
			vorlage:  "{name}",
			erwartet: []string{"Feuerball.png", "Eisball.png"},
		},
		{
			name:     "mit festem Text davor",
			vorlage:  "Probeset {nummer}",
			erwartet: []string{"Probeset 001.png", "Probeset 002.png"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gefunden := namen(t, karten("Feuerball", "Eisball"), fall.vorlage)

			if strings.Join(gefunden, ", ") != strings.Join(fall.erwartet, ", ") {
				t.Errorf("%v, erwartet %v", gefunden, fall.erwartet)
			}
		})
	}
}

// **Zwei Karten mit demselben Namen überschreiben einander nicht.**
//
// Der Fall ist keiner am Rand: ein Set führt dieselbe Karte mehrfach, und eine
// Vorlage ohne `{nummer}` ist die, die jemand zuerst eintippt. Wer hier nichts
// tut, verliert Karten **stillschweigend** — die Ausgabe sieht vollständig aus.
//
// # Die Zählweise hat sich mit #491 geändert, und das ist Absicht
//
// Bis dahin hieß der zweite `Feuerball-2.png` — eine **erfundene** Zählung,
// die zudem **immer** lief, gleich was jemand wollte. MSE hängt die Zahl vor
// die Endung:
//
//	while (used.find(fn.GetFullPath()) != used.end())
//	  fn.SetExt(String() << ++i << _(".") << ext);
//
// `src/util/file_utils.cpp`, Zeilen 88–94, Fassung f9d9356 vom 2021-07-06.
// `SetExt` **ersetzt** die Endung: `Feuerball.1.png`. Und der Zähler beginnt
// bei 1, nicht bei 2.
//
// **Und sie läuft nur bei `nummerierenneu`** — MSEs Vorgabe. Wer
// `ueberschreiben` wählt, bekommt eine Datei; das hat er so gewählt.
func TestBildnamenKollidierenNicht(t *testing.T) {
	gefunden := namen(t, karten("Feuerball", "Feuerball", "Feuerball"), "{name}")

	erwartet := []string{"Feuerball.png", "Feuerball.1.png", "Feuerball.2.png"}
	if strings.Join(gefunden, ", ") != strings.Join(erwartet, ", ") {
		t.Errorf("%v, erwartet %v", gefunden, erwartet)
	}
}

// **Ein Kartenname ist eine Eingabe, kein Dateiname.**
//
// Er kommt aus einem fremden Paket (§8) und darf alles enthalten: Schrägstriche,
// Doppelpunkte, Steuerzeichen, `..`. Wer ihn ungeprüft zusammensetzt, hat einen
// Schreibweg, der aus dem gewählten Verzeichnis herausführt.
func TestEinKartennameFuehrtNichtAusDemVerzeichnis(t *testing.T) {
	faelle := map[string]string{
		"Feuer/ball":    "001-Feuer-ball.png",
		"../../oben":    "001-oben.png",
		"C:\\Windows":   "001-C-Windows.png",
		"Feuer\x00ball": "001-Feuer-ball.png",
		"  Rand  ":      "001-Rand.png",
		".versteckt":    "001-versteckt.png",
	}

	for name, erwartet := range faelle {
		t.Run(name, func(t *testing.T) {
			gefunden := namen(t, karten(name), "{nummer}-{name}")

			if gefunden[0] != erwartet {
				t.Errorf("%q, erwartet %q", gefunden[0], erwartet)
			}
		})
	}
}

// **Eine Karte ohne Namen bekommt trotzdem einen.**
//
// Sonst hieße die Datei `.png` — ein leerer Name ist unter macOS zwar erlaubt,
// aber niemand findet sie wieder, und die zweite namenlose Karte überschriebe
// die erste.
func TestEineKarteOhneNamenBekommtEinen(t *testing.T) {
	gefunden := namen(t, karten("", ""), "{name}")

	erwartet := []string{"karte-001.png", "karte-002.png"}
	if strings.Join(gefunden, ", ") != strings.Join(erwartet, ", ") {
		t.Errorf("%v, erwartet %v", gefunden, erwartet)
	}
}

// **Ein sehr langer Name wird gekürzt, und zwar an einer Zeichengrenze.**
//
// macOS lässt 255 Bytes je Dateiname zu. Ein Kartenname darüber — ein
// Regeltext, der versehentlich im Namensfeld steht — ließe das Schreiben
// scheitern, und zwar erst nach dem Rastern, also nach der teuren Hälfte.
// Gekürzt wird nach **Bytes**, geschnitten an Zeichen: ein halbes „ä" im
// Dateinamen ist kein Name.
func TestEinSehrLangerNameWirdGekuerzt(t *testing.T) {
	// **Zwei und drei Bytes je Zeichen**, und das ist der Punkt: bei „ä" fällt
	// die Grenze auf eine Zeichengrenze, bei „€" mitten hinein. Ein Prüfstück
	// mit nur dem ersten Fall ließe die Schleife ungeprüft, die den Schnitt
	// zurücknimmt (§10).
	for _, zeichen := range []string{"ä", "€"} {
		t.Run(zeichen, func(t *testing.T) {
			gefunden := namen(t, karten(strings.Repeat(zeichen, 300)), "{name}")

			if len(gefunden[0]) > 255 {
				t.Errorf("%d Bytes lang", len(gefunden[0]))
			}

			if !strings.HasSuffix(gefunden[0], ".png") {
				t.Errorf("die Endung ist abgeschnitten: %q", gefunden[0])
			}

			// **Kein halbes Zeichen.** Ein Schnitt mitten in einer UTF-8-Folge
			// ergäbe ein Ersatzzeichen, und das ist im Dateinamen nicht mehr zu
			// tilgen.
			if strings.ContainsRune(gefunden[0], '\uFFFD') {
				t.Errorf("mitten in einem Zeichen geschnitten: %q", gefunden[0])
			}
		})
	}
}

// Ein Format, das keines ist, wird abgewiesen — nicht geraten.
//
// **Das Gegenbeispiel war `tiff`, bis es eines wurde** (#826). Ein Test, dessen
// „gibt es nicht" irgendwann gebaut wird, prüft danach das Gegenteil dessen,
// was er sagt — hier fiel es sofort auf, weil er rot wurde. `webp` steht
// stattdessen: kein Kodierer dafür ist Abhängigkeit, und MSE2 bietet es nicht
// an.
func TestEinUnbekanntesBildformatWirdAbgewiesen(t *testing.T) {
	_, err := service.NeueBildausfuhr(kartenDouble{leserDouble: karten("Feuerball")}, nil, nil).
		Namen(t.Context(), 7, "{name}", service.Bildformat("webp"), domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{})
	if err == nil {
		t.Fatal("kein Fehler — unter welcher Endung wäre geschrieben worden?")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler: %v, erwartet ErrValidation", err)
	}
}

// Eine leere Vorlage ist keine Eingabe, sondern die Vorgabe.
func TestEineLeereVorlageIstDieVorgabe(t *testing.T) {
	gefunden := namen(t, karten("Feuerball"), "   ")

	if gefunden[0] != "001-Feuerball.png" {
		t.Errorf("%q", gefunden[0])
	}
}

// Ein Bestand, der nicht antwortet, wird gemeldet und nicht übergangen.
func TestNamenMeldenEinenLesefehler(t *testing.T) {
	bestand := karten("Feuerball")
	bestand.kartenfehler = errors.New("die Datenbank ist zu")

	if _, err := service.NeueBildausfuhr(kartenDouble{leserDouble: bestand}, nil, nil).
		Namen(t.Context(), 7, "{name}", service.FormatPNG, domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{}); err == nil {
		t.Fatal("kein Fehler")
	}
}
