package app

// **Teilen: vorbereiten, anbieten, verwerfen** — NET-04
// (#1708).
//
// Je Art ein Lauf über den Weg, den ihre Ausfuhr schon geht, dazu die drei
// Zusagen: was mitgeht und was fehlt steht vorher da (AK-02), ohne Fenster
// gibt es keinen Dialog (AK-03), und eine Teilung räumt sich weg (AK-04).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// teildecks gibt ein Deck her — die anderen Wege ruft das Teilen nicht.
type teildecks struct {
	Decknaht

	deck   domain.Deck
	fehler error
}

func (d teildecks) Deck(context.Context, int64) (domain.Deck, error) { return d.deck, d.fehler }

// teilbilder zeichnet eine Karte als Datei mit fünf Bytes.
type teilbilder struct {
	Bildausfuehrer

	fehler error

	// ohneDatei nennt einen Pfad und legt nichts dorthin.
	ohneDatei bool
}

func (b teilbilder) Schreibe(_ context.Context, _ int64, verzeichnis, datei string,
	_ service.Bildformat, _ int, _ string, _ render.Guete,
) (*service.Bildschreibung, error) {
	if b.fehler != nil {
		return nil, b.fehler
	}

	pfad := filepath.Join(verzeichnis, datei+".png")
	if b.ohneDatei {
		return &service.Bildschreibung{Pfad: pfad}, nil
	}

	return &service.Bildschreibung{Pfad: pfad}, os.WriteFile(pfad, []byte("12345"), 0o600)
}

// teilabzieher schreibt ein Set als Datei mit drei Bytes.
type teilabzieher struct {
	Speicherer

	fehler error
}

func (s teilabzieher) Setname(context.Context, int64) (string, error) { return "Mein Set", nil }

func (s teilabzieher) Abzug(_ context.Context, _ int64, archiv string) error {
	if s.fehler != nil {
		return s.fehler
	}

	return os.WriteFile(archiv, []byte("zip"), 0o600)
}

// teilsammlung gibt die Sammlung als CSV.
type teilsammlung struct {
	Sammlungsausfuhrquelle

	fehler error
}

func (s teilsammlung) Ausgeben(_ context.Context, form string) (string, error) {
	return "form=" + form, s.fehler
}

// eineTeilende baut eine Anwendung mit Teilstelle und gezähltem Dialog.
func eineTeilende(t *testing.T) (anwendung *App, gezeigt *[]string, wurzel string) {
	t.Helper()

	wurzel = t.TempDir()
	gezeigt = &[]string{}

	anwendung = NeueApp("").MitTeilablage(service.NeueTeilstelle(wurzel, time.Now)).MitTeilnaht(
		func(_ context.Context, pfad, text string) error {
			*gezeigt = append(*gezeigt, pfad+"|"+text)

			return nil
		})

	anwendung.MitDecknaht(teildecks{deck: domain.Deck{Name: "Rot", Plaetze: []domain.Deckplatz{
		{Printing: "bolt", Menge: 4, Bereich: domain.BereichHauptdeck, Bezeichnung: "Lightning Bolt"},
	}}})
	anwendung.MitBildausfuhr(teilbilder{})
	anwendung.MitSpeicherer(teilabzieher{})
	anwendung.MitSammlungsausfuhr(teilsammlung{})

	return anwendung, gezeigt, wurzel
}

func TestJedeArtWirdVorbereitetUndBeschrieben(t *testing.T) {
	t.Parallel()

	anwendung, _, wurzel := eineTeilende(t)

	for _, fall := range []struct {
		art, bezug, dateiname, text string
		groesse                     int64
	}{
		{"einladung", "  mse4mac-kontakt:1:abc.def  ", "Einladung.txt", "mse4mac-kontakt:1:abc.def", 26},
		{"kartenbild", "7", "Karte 7.png", "", 5},
		{"set", "3", "Mein Set.mse-set", "", 3},
		{"deck", "5", "Rot.txt", "", 22},
		{"sammlung", "", "Sammlung.csv", "", 8},
	} {
		antwort := anwendung.TeilenVorbereiten(context.Background(), fall.art, fall.bezug)
		if antwort.Fehler != nil || antwort.Teilgut == nil {
			t.Fatalf("%s: %+v", fall.art, antwort.Fehler)
		}

		gut := antwort.Teilgut
		if gut.Dateiname != fall.dateiname || gut.Text != fall.text || gut.Groesse != fall.groesse ||
			len(gut.Mitgehendes) == 0 || gut.Verluste == nil {
			t.Errorf("%s: %+v", fall.art, gut)
		}

		if _, fehler := os.Stat(filepath.Join(wurzel, gut.Kennung, gut.Dateiname)); fehler != nil {
			t.Errorf("%s: die Datei fehlt: %v", fall.art, fehler)
		}
	}
}

func TestAnbietenZeigtDenDialogUndVerwerfenRaeumt(t *testing.T) {
	t.Parallel()

	anwendung, gezeigt, wurzel := eineTeilende(t)
	ctx := context.Background()

	gut := anwendung.TeilenVorbereiten(ctx, "einladung", "code").Teilgut

	if antwort := anwendung.TeilenAnbieten(ctx, gut.Kennung); antwort.Fehler != nil {
		t.Fatalf("anbieten: %+v", antwort.Fehler)
	}

	// **Bei einer Einladung geht der Text** — Nachrichten setzt ihn ein.
	if len(*gezeigt) != 1 || !strings.HasSuffix((*gezeigt)[0], "|code") {
		t.Errorf("gezeigt: %v", *gezeigt)
	}

	if antwort := anwendung.TeilenVerwerfen(ctx, gut.Kennung); antwort.Fehler != nil {
		t.Fatalf("verwerfen: %+v", antwort.Fehler)
	}

	if _, fehler := os.Stat(filepath.Join(wurzel, gut.Kennung)); !os.IsNotExist(fehler) {
		t.Errorf("die Teilung liegt noch: %v", fehler)
	}

	// **Verworfen ist verworfen** — ein zweites Anbieten findet nichts.
	if fehler := anwendung.TeilenAnbieten(ctx, gut.Kennung).Fehler; fehler == nil || fehler.Code != CodeNichtGefunden {
		t.Errorf("nach dem Verwerfen: %+v", fehler)
	}
}

func TestWasNichtGehtLegtNichtsAb(t *testing.T) {
	t.Parallel()

	anwendung, _, wurzel := eineTeilende(t)
	ctx := context.Background()

	kaputt := errors.New("klemmt")

	for name, weg := range map[string]func() *Fehler{
		"eine fremde Art":      func() *Fehler { return anwendung.TeilenVorbereiten(ctx, "brief", "").Fehler },
		"eine leere Einladung": func() *Fehler { return anwendung.TeilenVorbereiten(ctx, "einladung", " ").Fehler },
		"eine Karte ohne Zahl": func() *Fehler { return anwendung.TeilenVorbereiten(ctx, "kartenbild", "x").Fehler },
		"ein Set ohne Zahl":    func() *Fehler { return anwendung.TeilenVorbereiten(ctx, "set", "x").Fehler },
		"ein Deck ohne Zahl":   func() *Fehler { return anwendung.TeilenVorbereiten(ctx, "deck", "x").Fehler },
		"eine fremde Kennung":  func() *Fehler { return anwendung.TeilenAnbieten(ctx, "../x").Fehler },
		"verwerfen ins Fremde": func() *Fehler { return anwendung.TeilenVerwerfen(ctx, "../x").Fehler },
		"das Bild klemmt": func() *Fehler {
			anwendung.MitBildausfuhr(teilbilder{fehler: kaputt})
			defer anwendung.MitBildausfuhr(teilbilder{})

			return anwendung.TeilenVorbereiten(ctx, "kartenbild", "7").Fehler
		},
	} {
		if weg() == nil {
			t.Errorf("%s kam durch", name)
		}
	}

	// **Kein Rest**: jede gescheiterte Vorbereitung hat ihren Ordner geräumt.
	if eintraege := ordnerinhalt(t, wurzel); len(eintraege) != 0 {
		t.Errorf("im Teilordner liegt %v", eintraege)
	}
}

// **Jeder Erzeuger, der klemmt, hinterlässt nichts.**
func TestKlemmendeErzeugerHinterlassenNichts(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("klemmt")
	ctx := context.Background()

	for name, einrichten := range map[string]func(*App){
		"set":      func(a *App) { a.MitSpeicherer(teilabzieher{fehler: kaputt}) },
		"deck":     func(a *App) { a.MitDecknaht(teildecks{fehler: kaputt}) },
		"sammlung": func(a *App) { a.MitSammlungsausfuhr(teilsammlung{fehler: kaputt}) },
	} {
		anwendung, _, wurzel := eineTeilende(t)
		einrichten(anwendung)

		if antwort := anwendung.TeilenVorbereiten(ctx, name, "3"); antwort.Fehler == nil {
			t.Errorf("%s kam durch", name)
		}

		if eintraege := ordnerinhalt(t, wurzel); len(eintraege) != 0 {
			t.Errorf("%s ließ %v liegen", name, eintraege)
		}
	}
}

// **Ohne Fenster gibt es keinen Dialog** (AK-03) — und ohne Setschreiber,
// der abziehen kann, kein Set.
func TestOhneFensterKeinTeilen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ohne := NeueApp("")

	if fehler := ohne.TeilenVorbereiten(ctx, "sammlung", "").Fehler; fehler == nil {
		t.Error("ohne Teilstelle kam eine Vorbereitung durch")
	}

	// **Verwerfen ohne Stelle ist kein Fehler** — es gibt nichts zu räumen.
	if fehler := ohne.TeilenVerwerfen(ctx, strings.Repeat("a", 32)).Fehler; fehler != nil {
		t.Errorf("verwerfen ohne Stelle: %+v", fehler)
	}

	ohneDialog := NeueApp("").MitTeilablage(service.NeueTeilstelle(t.TempDir(), time.Now))
	gut := ohneDialog.TeilenVorbereiten(ctx, "einladung", "code").Teilgut

	if fehler := ohneDialog.TeilenAnbieten(ctx, gut.Kennung).Fehler; fehler == nil {
		t.Error("ohne Dialog kam ein Anbieten durch")
	}

	ohneAbzug := NeueApp("").MitTeilablage(service.NeueTeilstelle(t.TempDir(), time.Now))
	ohneAbzug.MitSpeicherer(nurSpeicherer{})

	if fehler := ohneAbzug.TeilenVorbereiten(ctx, "set", "3").Fehler; fehler == nil {
		t.Error("ein Speicherer ohne Abzug gab ein Set her")
	}

	ohneSammlung := NeueApp("").MitTeilablage(service.NeueTeilstelle(t.TempDir(), time.Now))
	if fehler := ohneSammlung.TeilenVorbereiten(ctx, "sammlung", "").Fehler; fehler == nil {
		t.Error("ohne Sammlungsausfuhr kam eine Sammlung durch")
	}
}

// nurSpeicherer kann speichern und nicht abziehen.
type nurSpeicherer struct{ Speicherer }

// klagendeAblage klagt an einer Stelle, die ein Lauf wählt.
type klagendeAblage struct {
	Teilablage

	ordner, ablegen, verwerfen error
}

func (k klagendeAblage) Ordner() (kennung, ordner string, err error) {
	if k.ordner != nil {
		return "", "", k.ordner
	}

	return k.Teilablage.Ordner()
}

func (k klagendeAblage) Ablegen(
	kennung, name string, inhalt []byte,
) (pfad string, groesse int64, err error) {
	if k.ablegen != nil {
		return "", 0, k.ablegen
	}

	return k.Teilablage.Ablegen(kennung, name, inhalt)
}

func (k klagendeAblage) Verwerfen(kennung string) error {
	if k.verwerfen != nil {
		return k.verwerfen
	}

	return k.Teilablage.Verwerfen(kennung)
}

// **Was an der Ablage oder am Dialog klemmt, kommt als Fehler zurück.**
func TestKlemmendeAblageUndDialog(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("klemmt")
	ctx := context.Background()
	echt := service.NeueTeilstelle(t.TempDir(), time.Now)

	if fehler := NeueApp("").MitTeilablage(klagendeAblage{Teilablage: echt, ordner: kaputt}).
		TeilenVorbereiten(ctx, "einladung", "code").Fehler; fehler == nil {
		t.Error("ohne Ordner kam eine Teilung zustande")
	}

	if fehler := NeueApp("").MitTeilablage(klagendeAblage{Teilablage: echt, ablegen: kaputt}).
		TeilenVorbereiten(ctx, "einladung", "code").Fehler; fehler == nil {
		t.Error("ohne Ablegen kam eine Teilung zustande")
	}

	// **Ein klemmendes Verwerfen kommt an — und lässt sich wiederholen**
	// (Nachlauf über Phase 6.7, N4): die Teilung bleibt, bis sie geräumt ist.
	verwerfend := NeueApp("").MitTeilablage(klagendeAblage{Teilablage: echt, verwerfen: kaputt})

	vorbereitet := verwerfend.TeilenVorbereiten(ctx, "einladung", "code").Teilgut
	if vorbereitet == nil {
		t.Fatal("die Teilung kam nicht zustande")
	}

	for versuch := range 2 {
		if fehler := verwerfend.TeilenVerwerfen(ctx, vorbereitet.Kennung).Fehler; fehler == nil {
			t.Errorf("Versuch %d: ein klemmendes Verwerfen kam durch", versuch)
		}
	}

	// **Eine unbekannte Kennung ist kein Fehler** — es gibt nichts zu räumen.
	if fehler := verwerfend.TeilenVerwerfen(ctx, strings.Repeat("a", 32)).Fehler; fehler != nil {
		t.Errorf("eine unbekannte Kennung: %+v", fehler)
	}

	klagend := NeueApp("").MitTeilablage(echt).MitTeilnaht(func(context.Context, string, string) error { return kaputt })
	gut := klagend.TeilenVorbereiten(ctx, "einladung", "code").Teilgut

	if fehler := klagend.TeilenAnbieten(ctx, gut.Kennung).Fehler; fehler == nil {
		t.Error("ein klemmender Dialog kam durch")
	}
}

// **Ein Deck ohne Namen heißt nach seiner Kennung**, und ein Bild, das nicht
// dasteht, wird nicht geteilt.
func TestNamenloseDecksUndFehlendeBilder(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := eineTeilende(t)
	ctx := context.Background()

	anwendung.MitDecknaht(teildecks{deck: domain.Deck{}})

	if gut := anwendung.TeilenVorbereiten(ctx, "deck", "5").Teilgut; gut == nil || gut.Dateiname != "Deck 5.txt" {
		t.Errorf("ein namenloses Deck: %+v", gut)
	}

	anwendung.MitBildausfuhr(teilbilder{ohneDatei: true})

	if fehler := anwendung.TeilenVorbereiten(ctx, "kartenbild", "7").Fehler; fehler == nil {
		t.Error("ein Bild ohne Datei wurde geteilt")
	}
}

// ordnerinhalt liest einen Ordner — ein Lesefehler bricht den Lauf ab, statt
// als „leer" durchzugehen.
func ordnerinhalt(t *testing.T, ordner string) []os.DirEntry {
	t.Helper()

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("den Teilordner lesen: %v", fehler)
	}

	return eintraege
}
