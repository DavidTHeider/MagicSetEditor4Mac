package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Setangaben kommen im Fenster an** —
// #567.
//
// # Warum Felder und Werte in **einer** Antwort
//
// Anders als bei den Karten: dort ändern sich die Werte bei jedem Tastendruck
// und die Beschreibungen nie, also sind es zwei Aufrufe. Ein Set ist **ein**
// Gegenstand — es gibt keine Liste, durch die die 89 Beschreibungen mitreisen
// müssten, und zwei Aufrufe für eine Seite wären zwei Läufe, die dieselbe
// Frage stellen.
func TestDieSetangabenKommenImFensterAn(t *testing.T) {
	waehler := &stilwaehler{setfelder: []domain.Field{
		{Name: "title", Kind: domain.KindText},
		{Name: "set code", Kind: domain.KindText},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(waehler)
	anwendung.MitSetwerten(&setwerteprobe{werte: []domain.Value{
		{Field: "title", Text: "Aus der Saat"},
	}})

	antwort := anwendung.Setangaben(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Felder) != 2 || antwort.Felder[0].Name != "title" {
		t.Errorf("%+v", antwort.Felder)
	}

	// **Die Aufschrift kommt mit** — MSE beschriftet ein Setfeld genauso wie
	// ein Kartenfeld (`src/data/field.cpp`, Zeile 69, Fassung f9d9356 vom
	// 2021-07-06).
	if antwort.Felder[1].Aufschrift != "Set code" {
		t.Errorf("Aufschrift %q", antwort.Felder[1].Aufschrift)
	}

	if len(antwort.Werte) != 1 || antwort.Werte[0].Wert != "Aus der Saat" {
		t.Errorf("%+v", antwort.Werte)
	}
}

// **Leere Listen und kein `nil`.** Über die Wails-Bindung stünde dort sonst
// `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
func TestDieSetangabenKommenAlsListenAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{fehler: errors.New("kein Spielpaket")})
	anwendung.MitSetwerten(&setwerteprobe{})

	antwort := anwendung.Setangaben(t.Context(), "1")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Felder == nil || antwort.Werte == nil {
		t.Errorf("Felder %v, Werte %v", antwort.Felder, antwort.Werte)
	}
}

// **Ein Fehler beim Lesen der Werte kommt durch** (§8).
func TestDieSetangabenMeldenEinenLesefehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{setfelder: []domain.Field{{Name: "title"}}})
	anwendung.MitSetwerten(&setwerteprobe{fehler: errors.New("die Platte ist voll")})

	if antwort := anwendung.Setangaben(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("der Lesefehler ging verloren")
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — dieselbe Wache wie überall.
func TestDieSetangabenWeisenEineUnlesbareKennungAb(t *testing.T) {
	anwendung := app.NeueApp("")

	if antwort := anwendung.Setangaben(t.Context(), "keine zahl"); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// **Und eine Angabe lässt sich schreiben.**
func TestEineSetangabeLaesstSichSchreiben(t *testing.T) {
	schreiber := &setwerteprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitSetwerten(schreiber)

	if antwort := anwendung.SetwertSetzen(t.Context(), "7", "set code", "PRB"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.setID != 7 || schreiber.feld != "set code" || schreiber.wert != "PRB" {
		t.Errorf("%+v", schreiber)
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — dieselbe Wache wie überall.
func TestSetwertSetzenWeistEineUnlesbareKennungAb(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSetwerten(&setwerteprobe{})

	if antwort := anwendung.SetwertSetzen(
		probekontext(t), "keine zahl", "set code", "PRB"); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// **Ein leerer Feldname wird abgewiesen** — er ergäbe eine Zeile, die niemand
// wiederfindet.
func TestEineSetangabeOhneFeldnamenWirdAbgewiesen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSetwerten(&setwerteprobe{})

	if antwort := anwendung.SetwertSetzen(t.Context(), "7", "  ", "PRB"); antwort.Fehler == nil {
		t.Error("ein leerer Feldname kam durch")
	}
}

// **Und ein Schreibfehler kommt durch.**
func TestEineSetangabeMeldetEinenSchreibfehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSetwerten(&setwerteprobe{fehler: errors.New("die Platte ist voll")})

	if antwort := anwendung.SetwertSetzen(
		probekontext(t), "7", "set code", "PRB"); antwort.Fehler == nil {
		t.Error("der Schreibfehler ging verloren")
	}
}

// setwerteprobe steht für Lesen und Schreiben der Setangaben.
type setwerteprobe struct {
	werte  []domain.Value
	fehler error

	setID int64
	feld  string
	wert  string

	// kopf ist, was `Setkopf` zurückgibt (#1325).
	kopf domain.Setkopf

	// kopffehler ist der Fehler **nur** dieses Wegs.
	//
	// **Getrennt von `fehler`**: sonst ließe sich nicht prüfen, dass ein
	// Fehler beim Kopf die Angaben nicht mitreißt — und umgekehrt.
	kopffehler error
}

func (s *setwerteprobe) Setkopf(context.Context, int64) (domain.Setkopf, error) {
	return s.kopf, s.kopffehler
}

func (s *setwerteprobe) Setwerte(context.Context, int64) ([]domain.Value, error) {
	return s.werte, s.fehler
}

func (s *setwerteprobe) SetwertSetzenMitVorwert(
	_ context.Context, setID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	s.setID, s.feld, s.wert = setID, feld, wert

	return "", false, s.fehler
}

// **Der Setkopf kommt mit** —
// #1325.
//
// Die fünf Formatangaben standen im Kasten „Nicht gedeutet"; seit dem
// 2026-09-03 stehen sie im Bestand und kommen mit derselben Antwort.
func TestDerSetkopfKommtMitDenAngaben(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{setfelder: []domain.Field{{Name: "title"}}})
	anwendung.MitSetwerten(&setwerteprobe{
		kopf: domain.Setkopf{MSEFassung: "2.1.2", Versionskontrolle: "none"},
	})

	antwort := anwendung.Setangaben(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler %v", antwort.Fehler)
	}

	if antwort.Kopf.MSEFassung != "2.1.2" {
		t.Errorf("die MSE-Fassung ist %q", antwort.Kopf.MSEFassung)
	}

	if antwort.Kopf.Versionskontrolle != "none" {
		t.Errorf("die Versionskontrolle ist %q", antwort.Kopf.Versionskontrolle)
	}
}

// **Und ein Fehler dabei nimmt die Angaben mit** (§8).
//
// **Nicht „die Angaben stehen trotzdem"**: eine Antwort, die zur Hälfte
// stimmt, ist schwerer zu deuten als eine, die sagt, dass etwas nicht ging.
// Wer die Formatangaben nicht lesen kann, hat ein Problem mit dem Set, nicht
// mit einem Feld.
func TestEinFehlerAmSetkopfKommtDurch(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{setfelder: []domain.Field{{Name: "title"}}})
	anwendung.MitSetwerten(&setwerteprobe{
		kopffehler: errors.New("der Kopf ist unlesbar"),
	})

	antwort := anwendung.Setangaben(t.Context(), "1")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	// **Leere Listen und kein `nil`** — dieselbe Zusage wie oben.
	if antwort.Felder == nil || antwort.Werte == nil {
		t.Errorf("Felder %v, Werte %v", antwort.Felder, antwort.Werte)
	}
}
