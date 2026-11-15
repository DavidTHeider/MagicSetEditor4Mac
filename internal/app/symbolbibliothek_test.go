package app_test

// **Die Symbolbibliothek an der Fassade** — DOM-02
// (#1697).
//
// Über die Fläche, die auch MCP, Kommandozeile und REST benutzen — und damit
// über dieselben App-Methoden wie das Fenster (§16). Mit eigenem Zuhause (§3):
// die Fläche sucht sonst das Bilderverzeichnis dieses Rechners.

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bibliotheksbestand legt einen Bestand mit zwei Sets an; das erste trägt ein
// gezeichnetes Symbol. **Mit eigenem Zuhause** — die Werkstatt legt dort ab.
func bibliotheksbestand(t *testing.T) (db *sql.DB, mitSymbol, ohne int64) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db = offeneProbedatenbank(t)
	sets := storage.NeuerSetSpeicher(db)

	mitSymbol, err := sets.SetAnlegen(probekontext(t), &domain.Set{Name: "Kronenset", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	ohne, err = sets.SetAnlegen(probekontext(t), &domain.Set{Name: "Leerset", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	werkstatt, err := app.SymbolwerkstattUeber(db)
	if err != nil {
		t.Fatalf("die Werkstatt bauen: %v", err)
	}

	datei, err := werkstatt.Ablegen(probekontext(t), mitSymbol,
		app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Rechteck"}).Teile, "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
		probekontext(t), mitSymbol, "symbol", datei); err != nil {
		t.Fatalf("das Symbol setzen: %v", err)
	}

	return db, mitSymbol, ohne
}

// bibliotheksflaeche baut die Fläche über [bibliotheksbestand].
func bibliotheksflaeche(t *testing.T) (flaeche *app.Symbolbibliotheksflaeche, mitSymbol, ohne string) {
	t.Helper()

	db, eins, zwei := bibliotheksbestand(t)

	flaeche, err := app.SymbolbibliotheksflaecheUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("die Fläche bauen: %v", err)
	}

	return flaeche, textKennung(eins), textKennung(zwei)
}

func textKennung(kennung int64) string { return strconv.FormatInt(kennung, 10) }

func TestDieBibliothekNimmtAufZeigtUndSagtWerEsTraegt(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)

	if leer := flaeche.Symbolbibliothek(ctx, ""); leer.Fehler != nil || leer.Symbole == nil || len(leer.Symbole) != 0 {
		t.Fatalf("die leere Bibliothek: %+v", leer)
	}

	aufgenommen := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, " Krone ")
	if aufgenommen.Fehler != nil || aufgenommen.ID == "" || aufgenommen.Name != "Krone" {
		t.Fatalf("aufnehmen: %+v", aufgenommen)
	}

	liste := flaeche.Symbolbibliothek(ctx, "KRO")
	if liste.Fehler != nil || len(liste.Symbole) != 1 {
		t.Fatalf("die Liste: %+v", liste)
	}

	symbol := liste.Symbole[0]
	if symbol.ID != aufgenommen.ID || symbol.Name != "Krone" || !strings.HasPrefix(symbol.Abdruck, "sha256:") {
		t.Errorf("der Eintrag: %+v", symbol)
	}

	if !strings.HasPrefix(symbol.Vorschau, "data:image/png;base64,") {
		t.Errorf("die Vorschau: %.40q", symbol.Vorschau)
	}

	if !slices.Equal(symbol.Verwendet, []string{"Kronenset"}) || symbol.Angelegt == "" || symbol.Geaendert == "" {
		t.Errorf("Träger und Zeiten: %+v", symbol)
	}

	if andere := flaeche.Symbolbibliothek(ctx, "zacke"); len(andere.Symbole) != 0 {
		t.Errorf("die Suche fand %+v", andere.Symbole)
	}
}

func TestKopieUmbenennenUndLoeschenMitFolgen(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)

	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	kopie := flaeche.BibliothekssymbolKopieren(ctx, krone.ID, "Krone 2")
	if kopie.Fehler != nil || kopie.ID == krone.ID || kopie.Name != "Krone 2" {
		t.Fatalf("kopieren: %+v", kopie)
	}

	if umbenannt := flaeche.BibliothekssymbolUmbenennen(ctx, kopie.ID, "Zacke"); umbenannt.Fehler != nil ||
		umbenannt.ID != kopie.ID || umbenannt.Name != "Zacke" {
		t.Fatalf("umbenennen: %+v", umbenannt)
	}

	// **Ohne Bestätigung geht ein verwendetes Symbol nicht** — und die
	// Absage nennt, wer es trägt.
	abgelehnt := flaeche.BibliothekssymbolLoeschen(ctx, krone.ID, false)
	if abgelehnt.Fehler == nil || abgelehnt.Fehler.Schluessel != domain.SchluesselSymbolVerwendet ||
		abgelehnt.Fehler.Code != app.CodeKonflikt || !slices.Equal(abgelehnt.Sets, []string{"Kronenset"}) {
		t.Fatalf("ohne Bestätigung: %+v", abgelehnt)
	}

	if abgelehnt.Fehler.Werte["sets"] != "Kronenset" {
		t.Errorf("die Werte: %v", abgelehnt.Fehler.Werte)
	}

	geloescht := flaeche.BibliothekssymbolLoeschen(ctx, krone.ID, true)
	if geloescht.Fehler != nil || !slices.Equal(geloescht.Sets, []string{"Kronenset"}) {
		t.Fatalf("mit Bestätigung: %+v", geloescht)
	}

	if liste := flaeche.Symbolbibliothek(ctx, ""); len(liste.Symbole) != 1 || liste.Symbole[0].Name != "Zacke" {
		t.Errorf("danach: %+v", liste.Symbole)
	}
}

func TestEinsetzenAusgebenUndEinlesen(t *testing.T) {
	flaeche, mitSymbol, ohne := bibliotheksflaeche(t)
	ctx := probekontext(t)

	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	eingesetzt := flaeche.BibliothekssymbolEinsetzen(ctx, krone.ID, ohne)
	if eingesetzt.Fehler != nil || !strings.HasSuffix(eingesetzt.Datei, ".mse-symbol") {
		t.Fatalf("einsetzen: %+v", eingesetzt)
	}

	// **In der Reihenfolge der Setliste** — zuletzt geöffnet zuerst, und
	// geöffnet ist keines; dann das jüngere.
	if liste := flaeche.Symbolbibliothek(ctx, ""); !slices.Equal(liste.Symbole[0].Verwendet, []string{"Leerset", "Kronenset"}) {
		t.Errorf("die Träger danach: %v", liste.Symbole[0].Verwendet)
	}

	ziel := filepath.Join(t.TempDir(), "zacke.mse-symbol")

	ausgegeben := flaeche.BibliothekssymbolAusgabeSchreiben(ctx, krone.ID, ziel)
	if ausgegeben.Fehler != nil || ausgegeben.Pfad != ziel || ausgegeben.ID != krone.ID {
		t.Fatalf("ausgeben: %+v", ausgegeben)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Fatalf("die Datei: %v", err)
	}

	eingelesen := flaeche.BibliothekssymbolEinlesenAus(ctx, ziel, "")
	if eingelesen.Fehler != nil || eingelesen.Name != "zacke" || eingelesen.Pfad != ziel {
		t.Fatalf("einlesen: %+v", eingelesen)
	}
}

// **Eine Kennung, die keine Zahl ist, steht am Feld**; ein Symbol, das es
// nicht gibt, ist nicht gefunden — an jedem Weg.
func TestDieBibliothekWeistKaputteKennungenAmFeldAb(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)
	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	for name, fall := range map[string]struct {
		fehler *app.Fehler
		feld   string
	}{
		"Aufnehmen":  {flaeche.BibliothekssymbolAufnehmen(ctx, "x", "X").Fehler, "setID"},
		"Kopieren":   {flaeche.BibliothekssymbolKopieren(ctx, "x", "X").Fehler, "id"},
		"Umbenennen": {flaeche.BibliothekssymbolUmbenennen(ctx, "x", "X").Fehler, "id"},
		"Löschen":    {flaeche.BibliothekssymbolLoeschen(ctx, "x", true).Fehler, "id"},
		"Einsetzen":  {flaeche.BibliothekssymbolEinsetzen(ctx, "x", mitSymbol).Fehler, "id"},
		"Einsetzen ins Set": {
			flaeche.BibliothekssymbolEinsetzen(ctx, krone.ID, "x").Fehler, "setID",
		},
		"Ausgeben": {flaeche.BibliothekssymbolAusgabeSchreiben(ctx, "x", "/y").Fehler, "id"},
	} {
		if fall.fehler == nil || fall.fehler.Feld != fall.feld {
			t.Errorf("%s: %+v", name, fall.fehler)
		}
	}

	for name, fehler := range map[string]*app.Fehler{
		"Kopieren":   flaeche.BibliothekssymbolKopieren(ctx, "99", "X").Fehler,
		"Umbenennen": flaeche.BibliothekssymbolUmbenennen(ctx, "99", "X").Fehler,
		"Löschen":    flaeche.BibliothekssymbolLoeschen(ctx, "99", true).Fehler,
		"Einsetzen":  flaeche.BibliothekssymbolEinsetzen(ctx, "99", mitSymbol).Fehler,
		"Ausgeben":   flaeche.BibliothekssymbolAusgabeSchreiben(ctx, "99", filepath.Join(t.TempDir(), "x")).Fehler,
		"Einlesen":   flaeche.BibliothekssymbolEinlesenAus(ctx, filepath.Join(t.TempDir(), "weg.mse-symbol"), "").Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeNichtGefunden {
			t.Errorf("%s: %+v", name, fehler)
		}
	}

	// **Und ein Löschen, das scheitert, gibt trotzdem eine Liste** — leer,
	// nicht `null` (`leerelisten_test.go`).
	if antwort := flaeche.BibliothekssymbolLoeschen(ctx, "x", true); antwort.Sets == nil {
		t.Error("die Folgen sind nil")
	}
}

// **Ohne Zuhause gibt es kein Bilderverzeichnis** — und die Fläche sagt es,
// statt die Kopien ins Wurzelverzeichnis zu legen.
func TestDieBibliotheksflaecheBrauchtEinZuhause(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := app.SymbolbibliotheksflaecheUeber(nil, domain.UrheberKommandozeile); err == nil {
		t.Error("ohne Zuhause ging die Fläche durch")
	}
}

// **Als Text hinaus und herein** — für MCP und REST, die an keiner Platte
// stehen, die ihr Gegenüber sieht.
func TestEinSymbolGehtAlsTextUeberDieFlaeche(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)
	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	text := flaeche.BibliothekssymbolText(ctx, krone.ID)
	if text.Fehler != nil || text.ID != krone.ID || text.Name != "Krone" || !strings.Contains(text.Inhalt, "mse_version:") {
		t.Fatalf("als Text: %+v", text)
	}

	reif := flaeche.BibliothekssymbolAusText(ctx, text.Inhalt, "Reif")
	if reif.Fehler != nil || reif.Name != "Reif" || reif.ID == krone.ID {
		t.Fatalf("aus Text: %+v", reif)
	}

	for name, fehler := range map[string]*app.Fehler{
		"Text ohne Kennung":   flaeche.BibliothekssymbolText(ctx, "x").Fehler,
		"Text fehlt":          flaeche.BibliothekssymbolText(ctx, "99").Fehler,
		"aus Text ohne Namen": flaeche.BibliothekssymbolAusText(ctx, text.Inhalt, "").Fehler,
	} {
		if fehler == nil {
			t.Errorf("%s ging durch", name)
		}
	}
}
