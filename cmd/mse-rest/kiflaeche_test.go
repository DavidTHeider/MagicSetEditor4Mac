package main

// **Die KI-Naht des Dienstes reicht ihre Fehler durch** — angestrichen beim
// Bau von A13.4
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass eine Naht ohne Bestand einen Fehler gibt und keine leere Sicht.**
// Der Unterschied zählt: „es gibt kein Modell" und „die Frage ging schief"
// sind zwei Auskünfte, und ein Dienst, der die erste gäbe, wo die zweite
// gilt, sagte, es sei alles in Ordnung.
//
// **Und was er nicht deckt**: den Erfolgsfall. Der braucht einen Bestand,
// eine Laufzeit und ein Modell; was er belegte, steht in `internal/flaechen`
// und hinter `netzlauf`.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// aufLeeremBestand gibt die Naht über einem **offenen, aber leeren** Bestand.
//
// # Warum nicht über `nil`
//
// **Weil `nil` kein Zustand dieses Dienstes ist.** `dienst.go` legt die Naht
// mit der laufenden Verbindung an; eine nil-Verbindung gibt es nur im Test —
// und `database/sql` stürzt daran ab, statt einen Fehler zu geben. Ein Lauf
// darüber prüfte die Vorrichtung und nicht den Weg (§3).
//
// **Ein leerer Bestand ist der Fall, den es gibt**: geöffnet, migriert,
// ohne die Karte, nach der gefragt wird.
func aufLeeremBestand(t *testing.T) *preise {
	t.Helper()

	db, fehler := sql.Open("sqlite3", ":memory:")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return &preise{db: db}
}

// **Jeder KI-Weg des Dienstes meldet sich über einem leeren Bestand.**
//
// **Tabellengetrieben**, weil vier Wege dieselbe Zusage haben — und eine
// Schleife über sie hält sie auch für den fünften, der dazukommt.
func TestJedeKINahtDesDienstesMeldetSichAufLeeremBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := aufLeeremBestand(t)

	wege := map[string]func() (any, error){
		"lesen.ki.modelle": func() (any, error) {
			return naht.KIModelle(probekontext(t))
		},
		"schreiben.ki.vorschlag": func() (any, error) {
			return naht.KIVorschlag(probekontext(t), rest.KIVorschlagsbezug{
				KarteID: 4, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "Schreibe.",
			})
		},
		"schreiben.ki.einschaetzung": func() (any, error) {
			return naht.KIEinschaetzung(probekontext(t),
				rest.KIEinschaetzungsbezug{
					KarteID: 4, Art: "regel", Anbieter: "ollama",
					Alias: "gemma4:8b", Frage: "Wie blockt sie?",
				})
		},
		"schreiben.ki.erkennung": func() (any, error) {
			return naht.KIErkennung(probekontext(t), rest.KIErkennungsbezug{
				Pfad: "/gibtesnicht.png", Anbieter: "ollama",
				Alias: "gemma4:8b",
			})
		},
	}

	for name, fahren := range wege {
		t.Run(name, func(t *testing.T) {
			if _, fehler := fahren(); fehler == nil {
				t.Error("über einem leeren Bestand kam kein Fehler")
			}
		})
	}
}

// **Und der Vorschlag reicht seine Belegfrage weiter.**
//
// **Ohne diesen Lauf sagte der obige nur, dass etwas scheitert** — nicht,
// dass die zwei Felder aus A13.4 überhaupt an der Anwendung ankommen. Ein
// Feld, das die Naht fallen ließe, sähe an der Absage genauso aus.
func TestDerVorschlagDesDienstesTraegtSeineBelegfrage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := aufLeeremBestand(t)

	// **Er scheitert weiter am leeren Bestand** — geprüft wird, dass er dabei
	// nicht schon an der Übersetzung des Verbunds stolpert.
	if _, fehler := naht.KIVorschlag(probekontext(t), rest.KIVorschlagsbezug{
		KarteID: 4, Anbieter: "ollama", Alias: "gemma4:8b",
		Auftrag: "Schreibe.", Belegfrage: "trampelschaden",
		Freigegeben: []string{"eigenekarte"},
	}); fehler == nil {
		t.Error("über einem leeren Bestand kam kein Fehler")
	}
}
