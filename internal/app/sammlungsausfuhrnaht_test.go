// **Die Sammlungsausfuhrnaht** — Werkstück 4 von Phase 4.5
// (#1292).
//
// **Eine eigene Datei**, weil `listeneinfuhrnaht_test.go` an der Zeilengrenze
// aus §2 stand — und weil der Schnitt der Frage folgt: dort **hinein**, hier
// **heraus**.
package app_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Ausfuhr braucht kein Zuhause** — das ist ihr Unterschied zur Einfuhr.
//
// §16, Regel 5: ohne Netz arbeitet die Anwendung weiter. Wer beide Nähte
// zusammenlegte, machte das Ausgeben von einem Scryfall-Zwischenspeicher
// abhängig, den es dafür nicht braucht.
func TestDieAusfuhrnahtBrauchtKeinZuhause(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	naht := app.AusfuhrUeber(frischerBestand(t))

	text, err := naht.Ausgeben(probekontext(t), "csv")
	if err != nil {
		t.Fatalf("Ausgeben: %v", err)
	}

	if !strings.Contains(text, "printing") {
		t.Errorf("die Ausgabe hat keine Kopfzeile: %q", text)
	}

	bericht, err := naht.CSVEinlesen(probekontext(t), "printing,menge\nprobe-druck,2\n", "", "")
	if err != nil {
		t.Fatalf("CSVEinlesen: %v", err)
	}

	if len(bericht.Angelegt) != 1 || bericht.Stueck != 2 {
		t.Errorf("%d angelegt, %d Stück", len(bericht.Angelegt), bericht.Stueck)
	}
}

// **Eine Form, die es nicht gibt, kommt als Fehler an der Bindung an.**
func TestEineUnbekannteFormAnDerBindung(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSammlungsausfuhr(stolpernde{})

	bindungen := app.NeueBindungen(anwendung)

	if bindungen.SammlungAusgeben(t.Context(), "csv").Fehler == nil {
		t.Error("SammlungAusgeben gab keinen Fehler")
	}

	if bindungen.SammlungCSVEinlesen(t.Context(), "x", "", "").Fehler == nil {
		t.Error("SammlungCSVEinlesen gab keinen Fehler")
	}
}

// **Ohne eingerichtete Ausfuhr sagt die Bindung es selbst.**
func TestOhneEingerichteteAusfuhr(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if anwendung.SammlungAusgeben(t.Context(), "csv").Fehler == nil {
		t.Error("SammlungAusgeben ohne Naht gab keinen Fehler")
	}

	if anwendung.SammlungCSVEinlesen(t.Context(), "x", "", "").Fehler == nil {
		t.Error("SammlungCSVEinlesen ohne Naht gab keinen Fehler")
	}
}

// **Und der geglückte Weg über die App-Methoden.**
func TestDerGeglueckteAusfuhrweg(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSammlungsausfuhr(willige{})

	gegeben := anwendung.SammlungAusgeben(t.Context(), "csv")
	if gegeben.Fehler != nil || gegeben.Text == "" {
		t.Errorf("SammlungAusgeben: %+v", gegeben)
	}

	gelesen := anwendung.SammlungCSVEinlesen(t.Context(), "x", "", "")
	if gelesen.Fehler != nil || gelesen.Bericht == nil {
		t.Errorf("SammlungCSVEinlesen: %+v", gelesen)
	}
}

// **Ein geschlossener Bestand macht aus beiden Wegen einen Fehler.**
//
// Der Zweig, den kein anderer Lauf betritt: `Einfuegen` gibt für kaputte
// **Zeilen** Befunde und keinen Fehler (§8) — einen Fehler gibt es erst, wenn
// der Bestand selbst nicht antwortet.
func TestEinGeschlossenerBestandAnDerAusfuhrnaht(t *testing.T) {
	db := frischerBestand(t)
	naht := app.AusfuhrUeber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, err := naht.Ausgeben(probekontext(t), "csv"); err == nil {
		t.Error("Ausgeben auf geschlossenem Bestand gab keinen Fehler")
	}

	if _, err := naht.CSVEinlesen(probekontext(t),
		"printing,menge\nabc,1\n", "", ""); err == nil {
		t.Error("CSVEinlesen auf geschlossenem Bestand gab keinen Fehler")
	}
}
