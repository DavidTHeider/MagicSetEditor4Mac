package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Wert wird links beschnitten, nicht beidseitig** —
// #565.
//
//	value = trim_left(substr(line, pos+1));
//
// `Reader::readLine`, `src/util/io/reader.cpp`, Zeile 224, Fassung f9d9356 vom
// 2021-07-06. Der **Schlüssel** dagegen wird beidseitig beschnitten
// (`key = canonical_name_form(trim(key))`, Zeile 217) — zwei verschiedene
// Aufrufe in zwei aufeinanderfolgenden Zeilen, und das ist kein Zufall.
//
// # Wo das zählt
//
// Gemessen über 1.102 Textdateien in `full-magic-pack/data`: **1.432 Werte**
// tragen Leerraum am Ende. Der sichtbarste Fall ist `auto_replace`:
//
//	replace: •_
//
// — das Leerzeichen dahinter gehört zur Ersetzung. Wer es wegschneidet, tippt
// beim Nutzer einen Aufzählungspunkt ohne Abstand.
//
// # Und was er **nicht** darf
//
// Die Rundlaufzusage hängt nicht daran: `Knoten.Roh` trägt die Zeile
// unverändert ([0002](../../docs/entscheidungen/0002-rundlauf-byte-fuer-byte.md)).
func TestEinWertBehaeltSeinenLeerraumAmEnde(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		zeile    string
		erwartet string
	}{
		"ein Leerzeichen am Ende":      {"replace: • ", "• "},
		"mehrere am Ende":              {"replace: •   ", "•   "},
		"ein Tabulator am Ende":        {"replace: •\t", "•\t"},
		"links wird geschnitten":       {"replace:    •", "•"},
		"beides zugleich":              {"replace:   • ", "• "},
		"ohne Leerraum":                {"replace: •", "•"},
		"ein Wert aus Leerraum allein": {"replace:   ", ""},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			baum, fehler := msepaket.Zerteile("probe", []byte(fall.zeile))
			if fehler != nil {
				t.Fatalf("zerteilen: %v", fehler)
			}

			knoten := baum.Ersten("replace")
			if knoten == nil {
				t.Fatal("kein Knoten `replace`")
			}

			if knoten.Wert != fall.erwartet {
				t.Errorf("Wert %q, erwartet %q", knoten.Wert, fall.erwartet)
			}
		})
	}
}
