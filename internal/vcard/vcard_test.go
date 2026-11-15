package vcard_test

// **Eine vCard lesen, wie Apple Kontakte sie schreibt** — ID-04
// (#1703).
//
// Die Proben sind **selbst geschrieben** (§3): dieselbe Syntax wie eine
// Ausfuhr aus Kontakte — Gruppen, Faltung, Base64-Bild —, eigener Inhalt.

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vcard"
)

// crlf macht aus Zeilen eine Datei mit CRLF, wie die RFC es verlangt.
func crlf(zeilen ...string) []byte {
	return []byte(strings.Join(zeilen, "\r\n") + "\r\n")
}

func TestEineKarteWieKontakteSieSchreibt(t *testing.T) {
	t.Parallel()

	karten, befunde := vcard.Lies(crlf(
		"BEGIN:VCARD",
		"VERSION:3.0",
		"PRODID:-//Apple Inc.//macOS 27.0//EN",
		"N:Lovelace;Ada;;;",
		"FN:Ada Lovelace",
		"item1.EMAIL;type=INTERNET;type=pref:ada@example.org",
		"item1.X-ABLabel:_$!<Other>!$_",
		"TEL;type=CELL;type=VOICE:+49 170 0000000",
		"EMAIL;type=INTERNET;type=HOME:ada.privat@example.org",
		"UID:0A1B2C3D-0000-4000-8000-000000000001",
		"END:VCARD",
	))

	if len(befunde) != 0 || len(karten) != 1 {
		t.Fatalf("%d Karten, Befunde %v", len(karten), befunde)
	}

	karte := karten[0]
	if karte.Name != "Ada Lovelace" || karte.Kennung != "0A1B2C3D-0000-4000-8000-000000000001" || karte.Zeile != 1 {
		t.Errorf("%+v", karte)
	}

	if len(karte.Emails) != 2 || karte.Emails[0].Wert != "ada@example.org" ||
		!slices.Equal(karte.Emails[0].Arten, []string{"internet", "pref"}) {
		t.Errorf("E-Mails %+v", karte.Emails)
	}

	if len(karte.Telefone) != 1 || karte.Telefone[0].Wert != "+49 170 0000000" {
		t.Errorf("Telefone %+v", karte.Telefone)
	}
}

func TestDerLeserFolgtDerRFC(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		zeilen []string
		pruefe func(t *testing.T, karte vcard.Karte)
	}{
		// RFC 6350 §3.2: CRLF und **ein** Leerraum werden entfernt.
		"gefaltete Zeilen werden zusammengesetzt": {
			[]string{"BEGIN:VCARD", "FN:Ada Love", " lace", "EMAIL:ada@exa", "\tmple.org", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Ada Lovelace" || karte.Emails[0].Wert != "ada@example.org" {
					t.Errorf("%+v", karte)
				}
			},
		},
		// RFC 6350 §3.4: Komma, Semikolon, Backslash und Zeilenwechsel.
		"maskierte Zeichen werden gelesen": {
			[]string{"BEGIN:VCARD", `FN:Lovelace\, Ada \; Gräfin\\Byron\nLondon`, "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Lovelace, Ada ; Gräfin\\Byron\nLondon" {
					t.Errorf("%q", karte.Name)
				}
			},
		},
		"ohne FN setzt sich der Name aus N zusammen": {
			[]string{"BEGIN:VCARD", "N:Lovelace;Ada;;Lady;", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Lady Ada Lovelace" {
					t.Errorf("%q", karte.Name)
				}
			},
		},
		"ohne UID gilt X-ABUID": {
			[]string{"BEGIN:VCARD", "FN:Ada", "X-ABUID:1234:ABPerson", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Kennung != "1234:ABPerson" {
					t.Errorf("%q", karte.Kennung)
				}
			},
		},
		"Eigenschaften sind unabhängig von der Schreibung": {
			[]string{"begin:vcard", "fn:Ada", "email;TYPE=work:ada@example.org", "end:vcard"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Ada" || !slices.Equal(karte.Emails[0].Arten, []string{"work"}) {
					t.Errorf("%+v", karte)
				}
			},
		},
		"TYPE mit Kommaliste zählt jede Art": {
			[]string{"BEGIN:VCARD", "FN:Ada", `TEL;TYPE="cell,voice":1`, "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if !slices.Equal(karte.Telefone[0].Arten, []string{"cell", "voice"}) {
					t.Errorf("%+v", karte.Telefone)
				}
			},
		},
		"ein kurzes N reicht": {
			[]string{"BEGIN:VCARD", "N:Lovelace;Ada", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Ada Lovelace" {
					t.Errorf("%q", karte.Name)
				}
			},
		},
		"ein maskiertes Semikolon in N trennt nicht": {
			[]string{"BEGIN:VCARD", `N:Love\;lace;Ada;;;`, "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != "Ada Love;lace" {
					t.Errorf("%q", karte.Name)
				}
			},
		},
		"ein Parameter ohne Gleichheitszeichen ist eine Art wie in 2.1": {
			[]string{"BEGIN:VCARD", "FN:Ada", "TEL;CELL:1", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if !slices.Equal(karte.Telefone[0].Arten, []string{"cell"}) {
					t.Errorf("%+v", karte.Telefone)
				}
			},
		},
		"ein unbekannter Backslash bleibt stehen": {
			[]string{"BEGIN:VCARD", `FN:C:\Pfad\`, "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if karte.Name != `C:\Pfad\` {
					t.Errorf("%q", karte.Name)
				}
			},
		},
		"ein leerer Kontaktweg zählt nicht": {
			[]string{"BEGIN:VCARD", "FN:Ada", "EMAIL:", "TEL: ", "END:VCARD"},
			func(t *testing.T, karte vcard.Karte) {
				t.Helper()

				if len(karte.Emails)+len(karte.Telefone) != 0 {
					t.Errorf("%+v", karte)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karten, befunde := vcard.Lies(crlf(fall.zeilen...))
			if len(karten) != 1 || len(befunde) != 0 {
				t.Fatalf("%d Karten, Befunde %v", len(karten), befunde)
			}

			fall.pruefe(t, karten[0])
		})
	}
}

// **Das Bild kommt mit, als Bytes und mit seiner Art** — RFC 2426 §3.1.4
// (`ENCODING=b`, Apple Kontakte) und RFC 6350 §6.2.4 (`data:`-Adresse).
func TestEinBildKommtMit(t *testing.T) {
	t.Parallel()

	punkt := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

	for name, zeile := range map[string]string{
		"Base64 nach 3.0":       "PHOTO;ENCODING=b;TYPE=PNG:iVBORw0KGgo=",
		"data-Adresse nach 4.0": "PHOTO:data:image/png;base64,iVBORw0KGgo=",
	} {
		karten, befunde := vcard.Lies(crlf("BEGIN:VCARD", "FN:Ada", zeile, "END:VCARD"))
		if len(befunde) != 0 || len(karten) != 1 {
			t.Fatalf("%s: %v", name, befunde)
		}

		if !bytes.Equal(karten[0].Bild, punkt) || karten[0].Bildart != "image/png" {
			t.Errorf("%s: %v %q", name, karten[0].Bild, karten[0].Bildart)
		}
	}
}

// **Was nicht gelesen wird, wird genannt** (§8): mit Zeile, und die übrigen
// Karten bleiben.
func TestWasNichtGehtWirdBenannt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		zeilen  []string
		karten  int
		befund  string
		zeileNr int
	}{
		"ein Bild als Adresse wird nicht geholt": {
			[]string{"BEGIN:VCARD", "FN:Ada", "PHOTO;VALUE=uri:https://example.org/ada.jpg", "END:VCARD"},
			1, "Adresse", 3,
		},
		"ein unlesbares Bild bleibt weg": {
			[]string{"BEGIN:VCARD", "FN:Ada", "PHOTO;ENCODING=b;TYPE=JPEG:%%%", "END:VCARD"},
			1, "Bild", 3,
		},
		"Quoted-Printable aus 2.1 wird nicht gedeutet": {
			[]string{"BEGIN:VCARD", "VERSION:2.1", "FN;ENCODING=QUOTED-PRINTABLE:Ad=C3=A4", "END:VCARD"},
			0, "Quoted-Printable", 3,
		},
		"eine Karte ohne Namen wird nicht übernommen": {
			[]string{"BEGIN:VCARD", "EMAIL:ada@example.org", "END:VCARD"},
			0, "Namen", 1,
		},
		"eine Karte ohne Ende": {
			[]string{"BEGIN:VCARD", "FN:Ada"},
			0, "END:VCARD", 1,
		},
		"eine Zeile außerhalb einer Karte": {
			[]string{"FN:Niemand", "BEGIN:VCARD", "FN:Ada", "END:VCARD"},
			1, "außerhalb", 1,
		},
		"eine Zeile ohne Doppelpunkt": {
			[]string{"BEGIN:VCARD", "FN:Ada", "KAPUTT", "END:VCARD"},
			1, "Doppelpunkt", 3,
		},
		"eine Karte in einer Karte": {
			[]string{"BEGIN:VCARD", "FN:Ada", "BEGIN:VCARD", "FN:Bo", "END:VCARD"},
			1, "offen", 1,
		},
		"gar keine vCard": {
			[]string{"das ist kein Adressbuch"},
			0, "außerhalb", 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karten, befunde := vcard.Lies(crlf(fall.zeilen...))
			if len(karten) != fall.karten {
				t.Errorf("%d Karten, erwartet %d", len(karten), fall.karten)
			}

			if len(befunde) == 0 || !strings.Contains(befunde[0].Satz, fall.befund) || befunde[0].Zeile != fall.zeileNr {
				t.Errorf("Befunde %+v — erwartet %q in Zeile %d", befunde, fall.befund, fall.zeileNr)
			}
		})
	}
}

func TestEineLeereDateiIstKeineVCard(t *testing.T) {
	t.Parallel()

	karten, befunde := vcard.Lies(nil)
	if len(karten) != 0 || len(befunde) != 1 || !strings.Contains(befunde[0].Satz, "keine vCard") {
		t.Errorf("%v %v", karten, befunde)
	}
}

func TestMehrereKartenUndNurLF(t *testing.T) {
	t.Parallel()

	daten := []byte("BEGIN:VCARD\nFN:Ada\nEND:VCARD\nBEGIN:VCARD\nFN:Bo\nEND:VCARD\n")

	karten, befunde := vcard.Lies(daten)
	if len(befunde) != 0 || len(karten) != 2 || karten[1].Name != "Bo" || karten[1].Zeile != 4 {
		t.Errorf("%+v %v", karten, befunde)
	}
}
