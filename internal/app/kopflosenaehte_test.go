// **Die Nähte der Flächen ohne Fenster** — Werkstücke
// #1324,
// #1390 und
// #1391.
//
// # Warum sie hier geprüft werden und nicht nur im Flächenlauf
//
// `internal/flaechen` fährt die **gebauten Programme** — eigene Prozesse, und
// ihre Abdeckung landet in keinem Profil dieses Pakets. Was dort grün ist,
// ist hier ungeprüft; die Zeilen sähen aus wie tot.
//
// **Mit eigenem Zuhause** (§3): die Nähte suchen sonst den Bestand dieses
// Rechners.
package app_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Skriptprüfung einer Paketdatei, kopflos.**
func TestPaketSkripteGeprueftNenntStelleUndSatz(t *testing.T) {
	zeile, spalte, meldung := app.PaketSkripteGeprueft(
		"mse version: 0.3.8\nscript: f(1, 2\n")

	if zeile != 2 || spalte != 10 {
		t.Errorf("Zeile %d, Spalte %d — erwartet 2 und 10 (%s)", zeile, spalte, meldung)
	}

	if meldung == "" {
		t.Error("kein Satz")
	}
}

// **Und eine heile Datei gibt Zeile null.**
//
// Das ist die Grenze, an der die drei Flächen „kein Befund" erkennen — sie
// steht in beiden Erklärungen, und ohne diesen Lauf wäre sie eine Behauptung.
func TestPaketSkripteGeprueftGibtOhneBefundZeileNull(t *testing.T) {
	zeile, _, meldung := app.PaketSkripteGeprueft("mse version: 0.3.8\nscript: 1 + 2\n")

	if zeile != 0 || meldung != "" {
		t.Errorf("Zeile %d, Meldung %q — erwartet null und leer", zeile, meldung)
	}
}

// symbolwerkstattMitSet baut die Naht über einem Bestand mit einem Set im
// Probeprofil — seit #1815 fragt sie, wem das Set gehört.
func symbolwerkstattMitSet(t *testing.T) (naht *app.Symbolwerkstattnaht, setID int64) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db := offeneProbedatenbank(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Symbolset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	naht, err := app.SymbolwerkstattUeber(db)
	if err != nil {
		t.Fatalf("die Naht bauen: %v", err)
	}

	return naht, setID
}

// **Die Symbolwerkstatt legt ab und liest zurück.**
func TestDieSymbolwerkstattLegtAbUndLiestZurueck(t *testing.T) {
	naht, setID := symbolwerkstattMitSet(t)

	antwort := app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Rechteck"})

	name, fehler := naht.Ablegen(probekontext(t), setID, antwort.Teile, "probe.mse-symbol")
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if name != "probe.mse-symbol" {
		t.Errorf("Name %q", name)
	}

	teile, fehler := naht.Lesen(probekontext(t), setID, name)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(teile) != 1 || teile[0].Name != "Rechteck" {
		t.Errorf("gelesen: %+v", teile)
	}
}

// **Und ein Symbol, das es nicht gibt, sagt es.**
func TestDieSymbolwerkstattWeistEinFehlendesSymbolAb(t *testing.T) {
	naht, setID := symbolwerkstattMitSet(t)

	if _, fehler := naht.Lesen(probekontext(t), setID, "gibtsnicht.mse-symbol"); fehler == nil {
		t.Error("das Lesen ging durch")
	}
}

// **Und ein Set, das es im Profil nicht gibt, ist „nicht gefunden" — beim
// Lesen wie beim Ablegen, bevor eine Datei angefasst wird** (#1815).
func TestDieSymbolwerkstattFasstKeinFremdesSetAn(t *testing.T) {
	naht, setID := symbolwerkstattMitSet(t)
	fremd := storage.MitProfil(t.Context(), 99)

	if _, fehler := naht.Lesen(fremd, setID, "probe.mse-symbol"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("lesen aus einem fremden Profil: %v", fehler)
	}

	antwort := app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Rechteck"})
	if _, fehler := naht.Ablegen(fremd, setID, antwort.Teile, ""); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ablegen aus einem fremden Profil: %v", fehler)
	}
}

// **Die Massenänderung geht über die Aufsicht und nicht am Dienst vorbei.**
//
// `storage.Kartenschreiber` hat ein `FelderSetzen` **ohne** Setkennung, und
// das eingebettete verdeckte das der Aufsicht. Ohne die ausdrückliche
// Weiterreichung ginge der Aufruf an der Setgrenze vorbei
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
func TestDieSchreibnahtFuehrtDieMassenaenderungUeberDieAufsicht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	naht, err := app.Schreibnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("die Naht bauen: %v", err)
	}

	// **Eine Karte, die es nicht gibt**, in einem Set, das es nicht gibt: der
	// Weg muss scheitern, und zwar fachlich.
	_, fehler := naht.FelderSetzen(probekontext(t), 999, []int64{998}, "artist", "X")
	if fehler == nil {
		t.Fatal("der Aufruf ging durch")
	}

	if !errors.Is(fehler, domain.ErrNotFound) && !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Klasse fehlt: %v", fehler)
	}

	if strings.Contains(fehler.Error(), "SQL") {
		t.Errorf("eine rohe Speichermeldung erreicht den Aufrufer: %v", fehler)
	}

	// **Vorschau und Erwartung gehen denselben Weg** (U6, #1417) — über die
	// Aufsicht, die ein fremdes Set abweist.
	if _, vorschau := naht.FelderVorschau(probekontext(t), 999, []int64{998}, "artist", "X"); vorschau == nil {
		t.Error("die Vorschau ging durch")
	}

	if _, erwartung := naht.FelderSetzenMitErwartung(probekontext(t), 999,
		[]domain.Feldvorwert{{KarteID: 998, Wert: "x", GabEs: true}}, "artist", "X"); erwartung == nil {
		t.Error("das Anwenden mit Erwartung ging durch")
	}
}

// **Die Symbolpalette eines Kartenfeldes, kopflos.**
//
// Das Probeset bringt keine Vorlage mit Palette mit; was hier zählt, ist,
// dass der Weg trägt und eine **Liste** zurückgibt — keine `null` (§3).
func TestDieSymbolpaletteGibtEineListe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	werk, err := app.Zeichennaht(db)
	if err != nil {
		t.Fatalf("die Naht bauen: %v", err)
	}

	// **Eine Karte, die es nicht gibt**, muss fachlich scheitern — das
	// belegt, dass der Weg bis zum Bestand geht statt stillschweigend eine
	// leere Liste zu geben.
	if _, fehler := werk.Palette(probekontext(t), 999, "name"); fehler == nil {
		t.Error("die Palette einer Karte, die es nicht gibt, ging durch")
	}

	// **Und eine, die es gibt, gibt eine Liste.** Ohne Vorlage ist sie leer,
	// und genau das ist die Zusage: eine leere **Liste** und keine `null`.
	kennung, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-stil",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Eine"}}}},
	}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), kennung)
	if lesefehler != nil || len(karten) == 0 {
		t.Fatalf("Karten: %v (%d)", lesefehler, len(karten))
	}

	zeilen, palettenfehler := werk.Palette(probekontext(t), karten[0].ID, "name")
	if palettenfehler != nil {
		t.Fatalf("Palette: %v", palettenfehler)
	}

	if zeilen == nil {
		t.Error("die Palette ist nil — über die Naht wird daraus null")
	}
}

// **Ohne Zuhause gibt es kein Bilderverzeichnis** — und die Naht sagt es.
//
// Der Zweig ist zwei Zeilen und sieht aus wie eine Formalie. Ohne ihn liefe
// ein Aufruf mit kaputter Umgebung in einen Pfad, der `MagicSetEditor4Mac`
// im Wurzelverzeichnis sucht — und legte dort etwas ab.
func TestDieSymbolwerkstattBrauchtEinZuhause(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := app.SymbolwerkstattUeber(nil); err == nil {
		t.Error("ohne Zuhause ging die Naht durch")
	}
}
