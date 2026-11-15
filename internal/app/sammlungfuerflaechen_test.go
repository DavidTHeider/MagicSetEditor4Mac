// **Die Sammlungsnaht für die Flächen ohne Fenster** — Werkstück 2 und 3 von
// Phase 3.4 (#1165).
//
// # Warum das ein Lauf gegen echtes Storage ist
//
// Die Naht ist der **Composition Root** dieser Fähigkeit: sie baut den
// Speicher, die Pflege und — wenn es ein Zuhause gibt — den Klienten. Gegen
// Doubles geprüft bliebe genau das ungeprüft, was sie tut.
//
// **`HOME` wird umgebogen** (§3): sonst legte dieser Lauf einen
// Zwischenspeicher im echten Zuhause dieses Rechners an, und er wäre grün,
// weil er wirklich funktioniert hat.
package app_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sammlungsnaht baut die Naht über einem frischen Bestand.
func sammlungsnaht(t *testing.T) *app.Sammlungsnaht {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	naht, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	return naht
}

// **Der ganze Weg: aufnehmen, lesen, ändern, löschen.**
func TestDieSammlungsnahtFuehrtDenGanzenWeg(t *testing.T) {
	naht := sammlungsnaht(t)

	id, err := naht.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "probe-druck", Menge: 2, Erwerbspreis: "4.17",
		Erwerbswaehrung: "EUR",
	})
	if err != nil {
		t.Fatalf("ExemplarAnlegen: %v", err)
	}

	liste, lesefehler := naht.Sammlung(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Sammlung: %v", lesefehler)
	}

	if len(liste) != 1 || liste[0].Erwerbspreis != "4.17" {
		t.Fatalf("gelesen wurde %+v", liste)
	}

	fuer, fuerfehler := naht.SammlungFuer(probekontext(t), "probe-druck")
	if fuerfehler != nil || len(fuer) != 1 {
		t.Fatalf("SammlungFuer: %v, %d", fuerfehler, len(fuer))
	}

	geaendert := liste[0]
	geaendert.Menge = 5

	if fehler := naht.ExemplarAendern(probekontext(t), geaendert); fehler != nil {
		t.Fatalf("ExemplarAendern: %v", fehler)
	}

	// **Nach dem Ändern wird neu gelesen.** Die Fassung ist eine andere, und
	// mit der alten zu löschen wäre ein Konflikt — zu Recht: dieser Lauf hätte
	// den neuen Stand nicht gesehen.
	nachher, nachlesefehler := naht.Sammlung(probekontext(t))
	if nachlesefehler != nil || len(nachher) != 1 {
		t.Fatalf("Sammlung: %v, %d", nachlesefehler, len(nachher))
	}

	if fehler := naht.ExemplarLoeschen(
		probekontext(t), id, nachher[0].Fassung); fehler != nil {
		t.Fatalf("ExemplarLoeschen: %v", fehler)
	}

	danach, danachfehler := naht.Sammlung(probekontext(t))
	if danachfehler != nil || len(danach) != 0 {
		t.Errorf("danach: %v, %d Zeilen", danachfehler, len(danach))
	}
}

// **Ein unlesbarer Preis kommt als Feldbefund zurück — auf beiden Wegen.**
func TestDieSammlungsnahtWeistEinenUnlesbarenPreisAb(t *testing.T) {
	naht := sammlungsnaht(t)

	kaputt := app.ExemplarDTO{Printing: "abc", Menge: 1, Erwerbspreis: "4,17"}

	if _, err := naht.ExemplarAnlegen(probekontext(t), kaputt); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Anlegen: %v", err)
	}

	kaputt.ID = 1

	if err := naht.ExemplarAendern(probekontext(t), kaputt); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Ändern: %v", err)
	}
}

// **Ohne Preise sind die drei Lesewege leer und nicht kaputt.**
func TestDieSammlungsnahtLiestOhnePreiseLeer(t *testing.T) {
	naht := sammlungsnaht(t)

	staende, err := naht.Preise(probekontext(t), "abc")
	if err != nil || len(staende) != 0 {
		t.Errorf("Preise: %v, %d", err, len(staende))
	}

	verlauf, verlauffehler := naht.Preisverlauf(probekontext(t), probereihe())
	if verlauffehler != nil || len(verlauf) != 0 {
		t.Errorf("Preisverlauf: %v, %d", verlauffehler, len(verlauf))
	}

	zeilen, ohnePreis, rechenfehler := naht.Portfolio(probekontext(t))
	if rechenfehler != nil || len(zeilen) != 0 || ohnePreis != 0 {
		t.Errorf("Portfolio: %v, %d, %d", rechenfehler, len(zeilen), ohnePreis)
	}
}

// **Und die Übersicht zählt, was ohne Preis blieb.**
func TestDieSammlungsnahtZaehltOhnePreis(t *testing.T) {
	naht := sammlungsnaht(t)

	if _, err := naht.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "ohnepreis", Menge: 3,
	}); err != nil {
		t.Fatalf("ExemplarAnlegen: %v", err)
	}

	_, ohnePreis, err := naht.Portfolio(probekontext(t))
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if ohnePreis != 3 {
		t.Errorf("ohne Preis sind %d", ohnePreis)
	}
}

// **Ein geschlossener Bestand meldet sich auf jedem Weg.**
//
// **Und das ist der Fehlerweg, den kein Double zeigt**: er entsteht in der
// Datenbank, nicht in einer Vorrichtung.
func TestDieSammlungsnahtMeldetEinenGeschlossenenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	naht, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	for name, versuch := range map[string]func() error{
		"Sammlung": func() error {
			_, fehler := naht.Sammlung(probekontext(t))

			return fehler
		},
		"SammlungFuer": func() error {
			_, fehler := naht.SammlungFuer(probekontext(t), "abc")

			return fehler
		},
		"ExemplarAnlegen": func() error {
			_, fehler := naht.ExemplarAnlegen(probekontext(t),
				app.ExemplarDTO{Printing: "abc", Menge: 1})

			return fehler
		},
		"ExemplarAendern": func() error {
			return naht.ExemplarAendern(probekontext(t),
				app.ExemplarDTO{ID: 1, Printing: "abc", Menge: 1})
		},
		"ExemplarLoeschen": func() error {
			return naht.ExemplarLoeschen(probekontext(t), 1, 1)
		},
		"Preise": func() error {
			_, fehler := naht.Preise(probekontext(t), "abc")

			return fehler
		},
		"Preisverlauf": func() error {
			_, fehler := naht.Preisverlauf(probekontext(t), probereihe())

			return fehler
		},
		"Portfolio": func() error {
			_, _, fehler := naht.Portfolio(probekontext(t))

			return fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			if versuch() == nil {
				t.Error("der geschlossene Bestand nahm den Aufruf an")
			}
		})
	}
}

// **Ohne Zuhause gibt es keine Naht.**
//
// Preise brauchen einen Zwischenspeicher; ohne ihn scheitert der Aufbau, statt
// eine halbe Naht herauszugeben, deren Preiswege abstürzen.
//
// **Das Zuhause ist eine Datei und kein fehlendes Verzeichnis.** Ein fehlendes
// legt `MkdirAll` einfach an — der Lauf wäre grün, ohne den Zweig zu betreten.
// Unter einer Datei geht das nicht.
func TestOhneZuhauseGibtEsKeineSammlungsnaht(t *testing.T) {
	statteinsZuhause := filepath.Join(t.TempDir(), "keinordner")

	if err := os.WriteFile(statteinsZuhause, []byte("x"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Setenv("HOME", statteinsZuhause)

	db := frischerBestand(t)

	if _, err := app.SammlungsnahtUeber(db); err == nil {
		t.Error("ohne Zuhause kam eine Naht heraus")
	}
}

// **Ein Abruf, der scheitert, kommt als Fehler durch — samt Befunden.**
func TestEinGescheiterterPreisabrufKommtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, err := app.SammlungsnahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	if _, _, holfehler := naht.PreiseHolen(probekontext(t), "gibtesnicht"); holfehler == nil {
		t.Error("der unbekannte Druck kam durch")
	}
}

// **Preise holen geht wirklich hinaus — an eine umgelenkte Adresse.**
//
// # Zwei Dinge gehören an so einen Lauf (§3)
//
// `t.Setenv("HOME", …)` **und** eine umgelenkte Adresse. Eines allein reicht
// nicht: ohne die Adresse ginge die Anfrage an Scryfall, und ohne das Zuhause
// läge im echten Zwischenspeicher ein Stand, der beim nächsten Lauf gefunden
// wird, ohne dass jemand hinausgegangen wäre.
func TestDieSammlungsnahtHoltPreise(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"id":"probe-druck","prices":{"eur":"4.17","usd":"5,00"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, err := app.SammlungsnahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	staende, befunde, holfehler := naht.PreiseHolen(probekontext(t), "probe-druck")
	if holfehler != nil {
		t.Fatalf("PreiseHolen: %v", holfehler)
	}

	if len(staende) != 1 || staende[0].Wert != "4.17" {
		t.Errorf("es kam %+v", staende)
	}

	// **Der unlesbare Preis ist ein Befund und bricht nichts** (§8).
	if len(befunde) != 1 {
		t.Errorf("es kamen %d Befunde: %v", len(befunde), befunde)
	}

	// **Und er liegt danach im Bestand.** Ohne diese Zusicherung prüfte der
	// Lauf den Abruf und nicht das Ablegen.
	abgelegt, lesefehler := naht.Preise(probekontext(t), "probe-druck")
	if lesefehler != nil || len(abgelegt) != 1 {
		t.Errorf("Preise: %v, %d", lesefehler, len(abgelegt))
	}
}

// **Ohne Netz zeigt die Übersicht den letzten Stand samt Alter** (§16, Satz 5).
//
// # Wie „ohne Netz" hier hergestellt wird
//
// Die Adresse zeigt auf einen Port, an dem niemand horcht. Ein Abruf scheitert
// damit sicher — und die drei **lesenden** Wege dürfen das nicht merken, weil
// sie gar nicht hinausgehen.
//
// **Das ist die Zusage, und sie ist mehr als „es geht".** Wer die Preise beim
// Lesen nachschlüge, hätte eine Anwendung, die ohne Netz keine Sammlung mehr
// zeigt.
func TestOhneNetzKommenDieAbgelegtenPreise(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"id":"probe-druck","prices":{"eur":"4.17"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, err := app.SammlungsnahtUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	if _, err := naht.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "probe-druck", Menge: 2,
	}); err != nil {
		t.Fatalf("ExemplarAnlegen: %v", err)
	}

	if _, _, holfehler := naht.PreiseHolen(probekontext(t), "probe-druck"); holfehler != nil {
		t.Fatalf("PreiseHolen: %v", holfehler)
	}

	// **Jetzt ist das Netz weg.**
	gegenstelle.Close()

	staende, lesefehler := naht.Preise(probekontext(t), "probe-druck")
	if lesefehler != nil || len(staende) != 1 {
		t.Fatalf("Preise: %v, %d", lesefehler, len(staende))
	}

	// **Und der Stand sagt, wie alt er ist.** Ohne diese Angabe wäre er eine
	// Zahl, die frisch aussieht und es nicht sein muss.
	if staende[0].Erhoben == "" {
		t.Error("der Stand nennt keinen Zeitpunkt")
	}

	zeilen, ohnePreis, rechenfehler := naht.Portfolio(probekontext(t))
	if rechenfehler != nil {
		t.Fatalf("Portfolio: %v", rechenfehler)
	}

	if len(zeilen) != 1 || zeilen[0].Wert != "8.34" || ohnePreis != 0 {
		t.Errorf("die Übersicht ist %+v, ohne Preis %d", zeilen, ohnePreis)
	}

	// **Und ein Abruf scheitert jetzt wirklich.** Ohne diese Zeile bewiese der
	// Lauf, dass die Lesewege ohne Netz gehen — nicht, dass es weg war.
	if _, _, holfehler := naht.PreiseHolen(probekontext(t), "anderer-druck"); holfehler == nil {
		t.Error("der Abruf kam durch, obwohl die Gegenstelle zu ist")
	}
}
