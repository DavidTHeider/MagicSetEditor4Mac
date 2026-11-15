package mcp_test

// Das In-Memory-Double des Bestands — die Vorrichtung aller Werkzeugtests.
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): dort stehen die
// Tests, hier steht, wogegen sie laufen. `server_test.go` ging über die
// Zeilengrenze, als der Lebenszyklus dazukam
// (#591).

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bestandsprobe ist ein In-Memory-Double des Bestands (§3).
type bestandsprobe struct {
	koepfe  []domain.SetHead
	karten  []*domain.CardRow
	felder  []string
	treffer domain.SearchResult
	fehler  error

	// schreibfehler scheitert **nur** beim Schreiben. Ein Double, das überall
	// gleich scheitert, kann den Weg dahinter nicht zeigen: `KarteLesen` käme
	// nie durch.
	schreibfehler error

	gesetzt []string

	// Der Lebenszyklus — was die Probe davon mitgeschrieben hat (#591).
	angelegt  []int64
	kopiert   []int64
	geloescht []int64
	getauscht [][2]int64
	naechste  int64

	// massengesetzt hält fest, was `schreiben.felder` weiterreichte (#1390).
	massengesetzt []string

	// vorgeschaut und erwartet halten fest, was die Vorschau und das Anwenden
	// mit Erwartung bekommen haben (U6, #1417).
	vorgeschaut []int64
	erwartet    []domain.Feldvorwert
}

func (b *bestandsprobe) KarteAnlegen(_ context.Context, setID int64) (int64, error) {
	if b.schreibfehler != nil {
		return 0, b.schreibfehler
	}

	b.angelegt = append(b.angelegt, setID)
	b.naechste++

	return 100 + b.naechste, nil
}

func (b *bestandsprobe) KarteKopieren(_ context.Context, _, karteID int64) (int64, error) {
	if b.schreibfehler != nil {
		return 0, b.schreibfehler
	}

	b.kopiert = append(b.kopiert, karteID)
	b.naechste++

	return 200 + b.naechste, nil
}

func (b *bestandsprobe) KarteLoeschen(_ context.Context, _, karteID int64, _ bool) error {
	if b.schreibfehler != nil {
		return b.schreibfehler
	}

	b.geloescht = append(b.geloescht, karteID)

	return nil
}

// FelderVorschau gibt je Karte eine Zeile — die zweite steht schon so (U6,
// #1417).
func (b *bestandsprobe) FelderVorschau(
	_ context.Context, _ int64, karteIDs []int64, feld, wert string,
) (domain.Massenvorschau, error) {
	if b.schreibfehler != nil {
		return domain.Massenvorschau{}, b.schreibfehler
	}

	b.vorgeschaut = append(b.vorgeschaut, karteIDs...)

	zeilen := make([]domain.Massenvorschauzeile, 0, len(karteIDs))
	for i, karteID := range karteIDs {
		zeile := domain.Massenvorschauzeile{
			KarteID: karteID, Name: "Karte " + strconv.FormatInt(karteID, 10),
			Alt: "Alt", GabAlt: true, Neu: wert,
		}
		if i == 1 {
			zeile.Alt, zeile.Ausschluss = wert, domain.AusschlussGleich
		}

		zeilen = append(zeilen, zeile)
	}

	return domain.Massenvorschau{Feld: feld, Wert: wert, Zeilen: zeilen}, nil
}

// FelderSetzenMitErwartung hält die Erwartung fest und schreibt wie
// FelderSetzen.
func (b *bestandsprobe) FelderSetzenMitErwartung(
	_ context.Context, _ int64, erwartet []domain.Feldvorwert, feld, wert string,
) (string, error) {
	if b.schreibfehler != nil {
		return "", b.schreibfehler
	}

	b.erwartet = append(b.erwartet, erwartet...)

	for _, vorwert := range erwartet {
		b.massengesetzt = append(b.massengesetzt,
			fmt.Sprintf("%d:%s=%s", vorwert.KarteID, feld, wert))
	}

	return "vorgang8", nil
}

// FelderSetzen schreibt **ein** Feld auf **mehreren** Karten (#1390).
//
// **Sie hält die Kartenkennungen fest**, nicht nur die Zahl: die Zusage des
// Werkzeugs ist, dass genau die genannten geändert werden, und eine Probe,
// die nur zählt, könnte das nicht zeigen.
func (b *bestandsprobe) FelderSetzen(
	_ context.Context, setID int64, karteIDs []int64, feld, wert string,
) (string, error) {
	if b.schreibfehler != nil {
		return "", b.schreibfehler
	}

	for _, karteID := range karteIDs {
		b.massengesetzt = append(b.massengesetzt,
			fmt.Sprintf("%d/%d/%s=%s", setID, karteID, feld, wert))
	}

	return "vorgang7", nil
}

func (b *bestandsprobe) KartenTauschen(_ context.Context, _, eineID, andereID int64) error {
	if b.schreibfehler != nil {
		return b.schreibfehler
	}

	b.getauscht = append(b.getauscht, [2]int64{eineID, andereID})

	return nil
}

func (b *bestandsprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return b.koepfe, b.fehler
}

func (b *bestandsprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return b.karten, b.fehler
}

func (b *bestandsprobe) Felder(context.Context, int64) ([]string, error) {
	return b.felder, b.fehler
}

func (b *bestandsprobe) Suche(context.Context, domain.SearchQuery) (domain.SearchResult, error) {
	return b.treffer, b.fehler
}

func (b *bestandsprobe) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	if b.fehler != nil {
		return nil, b.fehler
	}

	for _, karte := range b.karten {
		if karte.ID == karteID {
			return karte, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (b *bestandsprobe) FeldSetzenMitVorwert(
	_ context.Context, _, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	if b.schreibfehler != nil {
		return "", false, b.schreibfehler
	}

	if b.fehler != nil {
		return "", false, b.fehler
	}

	karte, err := b.karte(karteID)
	if err != nil {
		return "", false, err
	}

	vorher, hatte = karte.Value(feld)

	b.gesetzt = append(b.gesetzt, feld+"="+wert)

	// **Gesetzt und nicht angehängt** —
	// #915.
	//
	// Hier stand ein `append`, und `domain.CardRow.Value` liest immer den
	// **ersten** Treffer: ein zweiter Ruf gab damit denselben Vorwert zurück
	// wie der erste, obwohl der Betrieb den neuen gäbe. Genau daran war der
	// Idempotenztest grün — an der Vorrichtung, nicht am Weg.
	//
	// Das ist §3, wörtlich: *eine Vorrichtung, die großzügiger ist als der
	// Betrieb.* Angestrichen vom Zweitblick über Phase R.
	setzeWert(karte, feld, wert)

	return vorher, hatte, nil
}

func (b *bestandsprobe) FeldEntfernen(
	_ context.Context, _, karteID int64, feld string,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	if b.schreibfehler != nil {
		return "", false, nil, b.schreibfehler
	}

	karte, err := b.karte(karteID)
	if err != nil {
		return "", false, nil, err
	}

	vorher, hatte = karte.Value(feld)

	b.gesetzt = append(b.gesetzt, feld+" entfernt")

	uebrig := karte.Values[:0]

	// **Die Stelle wird mitgezählt** (#160): sie ist die Zusage, mit der ein
	// Gegenüber die Entfernung vollständig zurücknehmen kann.
	for nummer, wert := range karte.Values {
		if wert.Field == feld {
			gefunden := int64(nummer)
			stelle = &gefunden

			continue
		}

		uebrig = append(uebrig, wert)
	}

	karte.Values = uebrig

	return vorher, hatte, stelle, nil
}

// FeldSetzenAnStelle setzt ein Feld an eine gegebene Stelle (#160).
func (b *bestandsprobe) FeldSetzenAnStelle(
	_ context.Context, _, karteID int64, feld, wert string, stelle int64,
) (vorher string, hatte bool, fehler error) {
	if b.schreibfehler != nil {
		return "", false, b.schreibfehler
	}

	karte, err := b.karte(karteID)
	if err != nil {
		return "", false, err
	}

	vorher, hatte = karte.Value(feld)

	b.gesetzt = append(b.gesetzt,
		feld+" an "+strconv.FormatInt(stelle, 10)+" = "+wert)

	neu := domain.Value{Field: feld, Text: wert}

	if stelle < 0 || stelle > int64(len(karte.Values)) {
		karte.Values = append(karte.Values, neu)

		return vorher, hatte, nil
	}

	karte.Values = append(karte.Values[:stelle],
		append([]domain.Value{neu}, karte.Values[stelle:]...)...)

	return vorher, hatte, nil
}

// karte sucht eine Karte; fehlt sie, ist das ein ErrNotFound.
func (b *bestandsprobe) karte(karteID int64) (*domain.CardRow, error) {
	for _, karte := range b.karten {
		if karte.ID == karteID {
			return karte, nil
		}
	}

	return nil, domain.ErrNotFound
}

func probenbestand() *bestandsprobe {
	return &bestandsprobe{
		koepfe: []domain.SetHead{{ID: 1, Name: "Probeset", Game: "magic", Cards: 1}},
		karten: []*domain.CardRow{{ID: 7, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "Flammenball"},
		}}},
		felder:  []string{"name", "rule text"},
		treffer: domain.SearchResult{Hits: []domain.SearchHit{{CardID: 7, Field: "name", Value: "Flammenball"}}},
	}
}

// sprich schickt Zeilen an den Server und gibt die Antwortzeilen.
func sprich(t *testing.T, bestand *bestandsprobe, zeilen ...string) []string {
	t.Helper()

	server := mcp.NeuerServer(mcp.Werkzeuge(probenaehte(bestand)))

	var aus strings.Builder

	if err := server.Bediene(t.Context(), strings.NewReader(strings.Join(zeilen, "\n")+"\n"), &aus); err != nil {
		t.Fatalf("bedienen: %v", err)
	}

	ausgabe := strings.TrimSpace(aus.String())
	if ausgabe == "" {
		return nil
	}

	return strings.Split(ausgabe, "\n")
}

// setzeWert setzt einen Feldwert im Double — oder hängt ihn an, wenn es ihn
// noch nicht gibt.
//
// **`domain.CardRow` hat kein `SetValue`**, und es soll auch keins bekommen:
// im Betrieb schreibt die Datenbank, nicht die Zeile. Diese Vorrichtung bildet
// dasselbe Verhalten nach.
func setzeWert(karte *domain.CardRow, feld, wert string) {
	for i := range karte.Values {
		if karte.Values[i].Field == feld {
			karte.Values[i].Text = wert

			return
		}
	}

	karte.Values = append(karte.Values, domain.Value{Field: feld, Text: wert})
}

// ruhendeAusfuhr tut nichts und meldet nichts.
//
// **Für Tests, die die Ausfuhr nur brauchen, weil die Liste sie verlangt.**
// Wer sie prüft, baut sich ihre Teile im Test (`ablauf_test.go`).
type ruhendeAusfuhr struct{}

func (ruhendeAusfuhr) Einfahren(context.Context, string) (mcp.Einfahrt, error) {
	return mcp.Einfahrt{Befunde: []string{}}, nil
}

func (ruhendeAusfuhr) Sichern(context.Context, int64, string) ([]string, error) {
	return []string{}, nil
}

func (ruhendeAusfuhr) Zeichnen(context.Context, int64, string, string) (
	breite, hoehe int, befunde []string, fehler error,
) {
	return 0, 0, nil, nil
}

func (ruhendeAusfuhr) Bogen(context.Context, int64, string) (
	seiten, karten int, befunde []string, fehler error,
) {
	return 0, 0, nil, nil
}

func (ruhendeAusfuhr) MTGJSONLesen(context.Context, string, string) (
	mcp.MTGJSONEinfahrt, error,
) {
	return mcp.MTGJSONEinfahrt{Angelegt: []int64{}, Befunde: []string{}}, nil
}

func (ruhendeAusfuhr) MTGJSONSchreiben(context.Context, int64, string, string) (
	karten int, befunde []string, fehler error,
) {
	return 0, []string{}, nil
}

func (ruhendeAusfuhr) MTGJSONSpiele() []string { return []string{"magic"} }
