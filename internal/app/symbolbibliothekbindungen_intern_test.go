package app

// **Die Symbolbibliothek an der gebundenen Oberfläche** — DOM-02
// (#1697).
//
// **Ein innerer Lauf**, weil er `NeueBindungen` samt Dateidialog braucht —
// dieselbe Naht, die das Fenster benutzt —, und die Nähte, die im Betrieb
// nicht scheitern: die PNG-Kodierung und eine fehlende Bibliothek.

import (
	"context"
	"errors"
	"image"
	"io"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitSymbolbibliothek baut die gebundene Anwendung über einem Bestand mit
// einem Set, das ein Symbol trägt, samt Dialog, der `gewaehlt` antwortet.
func mitSymbolbibliothek(t *testing.T, gewaehlt *string, wahlfehler error) (bindungen *Bindungen, setID string) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db := probedatenbank(t)

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Kronenset", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	werkstatt, err := SymbolwerkstattUeber(db)
	if err != nil {
		t.Fatalf("die Werkstatt: %v", err)
	}

	datei, err := werkstatt.Ablegen(probekontext(t), kennung,
		SymbolGrundform(SymbolgrundformDTO{Art: "Rechteck"}).Teile, "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	if _, _, fehler := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
		probekontext(t), kennung, "symbol", datei); fehler != nil {
		t.Fatalf("das Symbol setzen: %v", fehler)
	}

	flaeche, err := SymbolbibliotheksflaecheUeber(db, domain.UrheberFenster)
	if err != nil {
		t.Fatalf("die Fläche: %v", err)
	}

	anwendung := flaeche.app
	anwendung.fenster = probekontext(t)
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return *gewaehlt, wahlfehler
	})

	return NeueBindungen(anwendung), strconv.FormatInt(kennung, 10)
}

func TestEinSymbolGehtUeberDenDialogHinausUndZurueck(t *testing.T) {
	var gewaehlt string

	bindungen, setID := mitSymbolbibliothek(t, &gewaehlt, nil)

	krone := bindungen.BibliothekssymbolAufnehmen(t.Context(), setID, "Krone")
	if krone.Fehler != nil {
		t.Fatalf("aufnehmen: %+v", krone.Fehler)
	}

	// **Ein anderer Dateiname**: „krone" hieße wie „Krone", und die Namen
	// sind ohne Rücksicht auf Groß und Klein eindeutig.
	gewaehlt = filepath.Join(t.TempDir(), "reif.mse-symbol")

	if antwort := bindungen.BibliothekssymbolAusgeben(t.Context(), krone.ID, "Krone"); antwort.Fehler != nil || antwort.Pfad != gewaehlt {
		t.Fatalf("ausgeben: %+v", antwort)
	}

	eingelesen := bindungen.BibliothekssymbolEinlesen(t.Context())
	if eingelesen.Fehler != nil || eingelesen.Name != "reif" {
		t.Fatalf("einlesen: %+v", eingelesen)
	}

	kopie := bindungen.BibliothekssymbolKopieren(t.Context(), krone.ID, "Zacke")
	umbenannt := bindungen.BibliothekssymbolUmbenennen(t.Context(), kopie.ID, "Zinne")
	eingesetzt := bindungen.BibliothekssymbolEinsetzen(t.Context(), kopie.ID, setID)
	geloescht := bindungen.BibliothekssymbolLoeschen(t.Context(), kopie.ID, true)

	for name, fehler := range map[string]*Fehler{
		"kopieren": kopie.Fehler, "umbenennen": umbenannt.Fehler,
		"einsetzen": eingesetzt.Fehler, "löschen": geloescht.Fehler,
	} {
		if fehler != nil {
			t.Errorf("%s: %+v", name, fehler)
		}
	}

	if liste := bindungen.Symbolbibliothek(t.Context(), ""); liste.Fehler != nil || len(liste.Symbole) != 2 {
		t.Errorf("danach: %+v", liste)
	}
}

// **Weggeklickt ist ein Abbruch, ein klagender Dialog ein Fehler** — bei
// beiden Wegen.
func TestDerSymboldialogBrichtAbOderKlagt(t *testing.T) {
	gewaehlt := ""
	bindungen, _ := mitSymbolbibliothek(t, &gewaehlt, nil)

	if antwort := bindungen.BibliothekssymbolEinlesen(t.Context()); !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("einlesen weggeklickt: %+v", antwort)
	}

	if antwort := bindungen.BibliothekssymbolAusgeben(t.Context(), "1", ""); !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("ausgeben weggeklickt: %+v", antwort)
	}

	klagend, _ := mitSymbolbibliothek(t, &gewaehlt, errors.New("klemmt"))

	if antwort := klagend.BibliothekssymbolEinlesen(t.Context()); antwort.Fehler == nil || antwort.Abgebrochen {
		t.Errorf("einlesen mit klagendem Dialog: %+v", antwort)
	}

	if antwort := klagend.BibliothekssymbolAusgeben(t.Context(), "1", "Krone"); antwort.Fehler == nil {
		t.Errorf("ausgeben mit klagendem Dialog: %+v", antwort)
	}
}

// **Der Vorschlag ist der Name, ohne Trenner** — und ohne Namen die Vorgabe.
func TestDerSymbolvorschlagIstEinDateiname(t *testing.T) {
	for name, erwartet := range map[string]string{
		"Krone":      "Krone.mse-symbol",
		" a/b:c\\d ": "a-b-c-d.mse-symbol",
		"  ":         "symbol.mse-symbol",
	} {
		if vorschlag := symbolvorschlag(name, "symbol.mse-symbol"); vorschlag != erwartet {
			t.Errorf("%q gab %q, erwartet %q", name, vorschlag, erwartet)
		}
	}
}

// **Ohne Bibliothek sagt jeder Weg es**, statt an `nil` zu stolpern.
func TestOhneBibliothekSagtJederWegEs(t *testing.T) {
	anwendung := NeueApp("")
	ctx := probekontext(t)

	fehler := map[string]*Fehler{
		"Liste":      anwendung.Symbolbibliothek(ctx, "").Fehler,
		"Aufnehmen":  anwendung.BibliothekssymbolAufnehmen(ctx, "1", "X").Fehler,
		"Kopieren":   anwendung.BibliothekssymbolKopieren(ctx, "1", "X").Fehler,
		"Umbenennen": anwendung.BibliothekssymbolUmbenennen(ctx, "1", "X").Fehler,
		"Löschen":    anwendung.BibliothekssymbolLoeschen(ctx, "1", true).Fehler,
		"Einsetzen":  anwendung.BibliothekssymbolEinsetzen(ctx, "1", "1").Fehler,
		"Ausgeben":   anwendung.BibliothekssymbolAusgabeSchreiben(ctx, "1", "/x").Fehler,
		"Einlesen":   anwendung.BibliothekssymbolEinlesenAus(ctx, "/x", "").Fehler,
	}

	for name, befund := range fehler {
		if befund == nil || befund.Code != CodeValidierung {
			t.Errorf("%s: %+v", name, befund)
		}
	}

	if liste := anwendung.Symbolbibliothek(ctx, ""); liste.Symbole == nil {
		t.Error("die Liste ist nil")
	}
}

// klemmendeBibliothek scheitert an der Liste.
type klemmendeBibliothek struct{ Symbolbibliotheksfuehrer }

func (klemmendeBibliothek) Liste(context.Context, string) ([]service.Symboleintrag, error) {
	return nil, errors.New("klemmt")
}

func TestEineKlemmendeListeUndEinePackfehlerhafteVorschau(t *testing.T) {
	var gewaehlt string

	bindungen, setID := mitSymbolbibliothek(t, &gewaehlt, nil)
	anwendung := bindungen.app(t.Context())
	bindungen.BibliothekssymbolAufnehmen(t.Context(), setID, "Krone")

	// **Eine Vorschau, die sich nicht packen lässt, ist leer** — der Eintrag
	// bleibt.
	anwendung.packen = func(io.Writer, image.Image) error { return errors.New("voll") }

	if liste := bindungen.Symbolbibliothek(t.Context(), ""); liste.Fehler != nil || len(liste.Symbole) != 1 ||
		liste.Symbole[0].Vorschau != "" {
		t.Errorf("mit Packfehler: %+v", liste)
	}

	anwendung.MitSymbolbibliothek(klemmendeBibliothek{})

	if liste := bindungen.Symbolbibliothek(t.Context(), ""); liste.Fehler == nil || liste.Symbole == nil {
		t.Errorf("mit klemmender Liste: %+v", liste)
	}
}

// **Ein Zeitpunkt, den niemand kennt, ist leer** und nicht das Jahr 1.
func TestEinUnbekannterZeitpunktIstLeer(t *testing.T) {
	if text := zeitpunktText(service.Symboleintrag{}.Symbol.Angelegt); text != "" {
		t.Errorf("der Nullwert gab %q", text)
	}
}

// **Auch die Textwege sagen, dass die Bibliothek fehlt.**
func TestOhneBibliothekSagenAuchDieTextwegeEs(t *testing.T) {
	anwendung := NeueApp("")
	ctx := probekontext(t)

	for name, befund := range map[string]*Fehler{
		"Text":     anwendung.BibliothekssymbolText(ctx, "1").Fehler,
		"aus Text": anwendung.BibliothekssymbolAusText(ctx, "x", "X").Fehler,
	} {
		if befund == nil || befund.Code != CodeValidierung {
			t.Errorf("%s: %+v", name, befund)
		}
	}
}
