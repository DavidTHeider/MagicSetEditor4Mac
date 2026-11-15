package service

// **Die Symbolbibliothek** — DOM-02
// (#1697):
// „Eine Bibliothek bietet Vorschau, Suche, Tags, Kopie, Import/Export sowie
// klare Unterscheidung zwischen Setsymbol und Symbolschrift" (AK-01), dazu
// Verwendungsnachweise und die Folgen vor dem Löschen (AK-02) und
// begrenzte, verständlich abgewiesene Dateien (AK-04).
//
// # Kopie und Abdruck statt Verweis
//
// **Ein Set trägt nach dem Einsetzen eine Kopie** — die Datei liegt neben
// dem Set, wie jedes Symbol, das dort gezeichnet wurde, und reist beim
// Sichern mit (`Symbolablage`). Ein Verweis in die Bibliothek wäre in einer
// weitergegebenen `.mse-set` ins Leere gegangen. **Welches Set welches
// Symbol trägt**, sagt der Abdruck: die Prüfsumme der Zeichnung in ihrer
// geschriebenen Form. Zwei gleiche Zeichnungen sind dasselbe Symbol, gleich
// woher sie kamen.
//
// **Symbolschriften führt die Bibliothek nicht.** Sie sind Pakete, gehören
// dem Rechner und stehen im Paketkatalog; das Fenster zeigt sie neben der
// Bibliothek, getrennt und nur lesend.
//
// Portierungsmarke **new** — MSE2 hat keine Bibliothek; gesucht in
// `src/gui/symbol/` und `src/data/symbol.cpp` (Fassung f9d9356 vom
// 2021-07-06).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Symbolspeicher hält die Symbole eines Profils.
type Symbolspeicher interface {
	Symbole(ctx context.Context) ([]domain.Bibliothekssymbol, error)
	Symbol(ctx context.Context, id int64) (domain.Bibliothekssymbol, error)
	SymbolAnlegen(ctx context.Context, symbol domain.Bibliothekssymbol) (domain.Bibliothekssymbol, error)
	SymbolUmbenennen(ctx context.Context, id int64, name string) error
	SymbolInhaltSetzen(ctx context.Context, id int64, inhalt, abdruck string) error
	SymbolLoeschen(ctx context.Context, id int64) error
}

// Setwertzugang liest und schreibt die Angaben eines Sets — hier nur
// `symbol`.
type Setwertzugang interface {
	Setwerte(ctx context.Context, setID int64) ([]domain.Value, error)
	SetwertSetzenMitVorwert(ctx context.Context, setID int64, feld, wert string) (
		vorher string, hatte bool, fehler error,
	)
}

// Symbolbibliothek führt die Symbole eines Profils.
type Symbolbibliothek struct {
	speicher  Symbolspeicher
	ablage    *Symbolablage
	setwerte  Setwertzugang
	setkoepfe Setkopfleser

	// fassungen nennt die Fassungen des Setsymbols — nil heißt: keine
	// Vorschau in Fassungen ([Symbolbibliothek.MitFassungen]).
	fassungen Symbolfassungsquelle

	// **Einsetzen, Nachziehen und Ausgeben schreiben Dateien** — die hält
	// eine lesende Verbindung nicht auf (Vermerk 0078,
	// [Symbolbibliothek.MitPruefer]).
	wache
}

// Symboleintrag ist ein Symbol samt Zeichnung und den Sets, die es tragen.
type Symboleintrag struct {
	Symbol    domain.Bibliothekssymbol
	Zeichnung msepaket.Symbolzeichnung
	Verwendet []string
}

// NeueSymbolbibliothek baut die Bibliothek.
func NeueSymbolbibliothek(
	speicher Symbolspeicher, ablage *Symbolablage, setwerte Setwertzugang, setkoepfe Setkopfleser,
) *Symbolbibliothek {
	return &Symbolbibliothek{speicher: speicher, ablage: ablage, setwerte: setwerte, setkoepfe: setkoepfe}
}

// setsymbolfeld ist das Setfeld, das in jedem Spielpaket das Symbol trägt.
const setsymbolfeld = "symbol"

// Liste gibt die Symbole, deren Name `suche` enthält — ohne Rücksicht auf
// Groß und Klein; eine leere Suche gibt alle.
func (b *Symbolbibliothek) Liste(ctx context.Context, suche string) ([]Symboleintrag, error) {
	symbole, err := b.speicher.Symbole(ctx)
	if err != nil {
		return nil, err
	}

	traeger, err := b.traegerJeAbdruck(ctx)
	if err != nil {
		return nil, err
	}

	frage := strings.ToLower(strings.TrimSpace(suche))
	eintraege := []Symboleintrag{}

	for _, symbol := range symbole {
		if !strings.Contains(strings.ToLower(symbol.Name), frage) {
			continue
		}

		// **Was im Bestand steht, war beim Hineinlegen geprüft**; eine
		// Zeichnung, die sich jetzt nicht mehr lesen ließe, zeigt eine leere
		// Vorschau statt die Liste zu verweigern (§8).
		zeichnung, _, _ := msepaket.SymbolzeichnungAus(symbol.Name, []byte(symbol.Inhalt)) //nolint:errcheck // siehe oben

		eintraege = append(eintraege, Symboleintrag{
			Symbol: symbol, Zeichnung: zeichnung,
			Verwendet: traegernamen(traeger[symbol.Abdruck]),
		})
	}

	return eintraege, nil
}

// Aufnehmen legt das Symbol eines Sets als neues Symbol in die Bibliothek.
func (b *Symbolbibliothek) Aufnehmen(ctx context.Context, setID int64, name string) (domain.Bibliothekssymbol, error) {
	geprueft, err := domain.PruefeSymbolname(name)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	if fehlt := SetDaSein(ctx, b.setkoepfe, setID); fehlt != nil {
		return domain.Bibliothekssymbol{}, fehlt
	}

	datei, err := b.symbolwert(ctx, setID)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	if datei == "" {
		return domain.Bibliothekssymbol{}, domain.NewFieldError(setsymbolfeld, domain.SchluesselSetOhneSymbol,
			"Dieses Set hat noch kein Symbol, das sich aufnehmen ließe.")
	}

	zeichnung, err := b.ablage.Gestalt(setID, datei)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	return b.anlegen(ctx, geprueft, zeichnung)
}

// anlegen schreibt eine Zeichnung als neues Symbol — in ihrer geschriebenen
// Form, damit der Abdruck von der Herkunft unabhängig ist.
func (b *Symbolbibliothek) anlegen(
	ctx context.Context, name string, zeichnung msepaket.Symbolzeichnung,
) (domain.Bibliothekssymbol, error) {
	inhalt := msepaket.SchreibeSymbolzeichnung(zeichnung)

	return b.speicher.SymbolAnlegen(ctx, domain.Bibliothekssymbol{
		Name: name, Inhalt: string(inhalt), Abdruck: symbolabdruck(inhalt),
	})
}

// symbolabdruck ist die Prüfsumme einer geschriebenen Zeichnung.
func symbolabdruck(inhalt []byte) string {
	summe := sha256.Sum256(inhalt)

	return "sha256:" + hex.EncodeToString(summe[:])
}

// symbolwert liest den Wert des Symbolfeldes eines Sets — leer, wenn es
// keines trägt.
func (b *Symbolbibliothek) symbolwert(ctx context.Context, setID int64) (string, error) {
	werte, err := b.setwerte.Setwerte(ctx, setID)
	if err != nil {
		return "", fmt.Errorf("die Angaben des Sets %d lesen: %w", setID, err)
	}

	for _, wert := range werte {
		if wert.Field == setsymbolfeld {
			return strings.TrimSpace(wert.Text), nil
		}
	}

	return "", nil
}

// symboltraeger ist ein Set, das ein Symbol trägt, samt der Datei, in der es
// liegt.
type symboltraeger struct {
	setID int64
	name  string
	datei string
}

// traegerJeAbdruck nennt je Abdruck die Sets, die ein Symbol mit ihm tragen —
// in der Reihenfolge der Setliste.
//
// **Ein Set, dessen Symboldatei fehlt oder sich nicht lesen lässt, trägt
// keines** — es ist kein Nachweis einer Verwendung, und die Liste der
// Bibliothek soll an einem kaputten Set nicht scheitern.
func (b *Symbolbibliothek) traegerJeAbdruck(ctx context.Context) (map[string][]symboltraeger, error) {
	koepfe, err := b.setkoepfe.Setkoepfe(ctx)
	if err != nil {
		return nil, err
	}

	traeger := map[string][]symboltraeger{}

	for _, kopf := range koepfe {
		datei, err := b.symbolwert(ctx, kopf.ID)
		if err != nil {
			return nil, err
		}

		if datei == "" {
			continue
		}

		zeichnung, lesefehler := b.ablage.Gestalt(kopf.ID, datei)
		if lesefehler != nil {
			continue
		}

		abdruck := symbolabdruck(msepaket.SchreibeSymbolzeichnung(zeichnung))
		traeger[abdruck] = append(traeger[abdruck], symboltraeger{setID: kopf.ID, name: kopf.Name, datei: datei})
	}

	return traeger, nil
}

// traegernamen gibt die Namen der Träger — **jeder einmal**, auch wenn zwei Sets
// gleich heißen.
func traegernamen(traeger []symboltraeger) []string {
	namen := []string{}

	for _, einer := range traeger {
		if !slices.Contains(namen, einer.name) {
			namen = append(namen, einer.name)
		}
	}

	return namen
}
