// **Die Vorschau einer ungespeicherten Vorlage** — Werkstück 3 von Phase 3.3
// (#1164).
//
// # Wofür
//
// §6 verlangt: „Eine Änderung an einem Feld ist in der Vorschau zu sehen, ohne
// dass jemand etwas bestätigt." Für eine Vorlage wiegt das schwerer als für
// ein Kartenfeld: Speichern ändert sie **für jedes Set**, das dieses Paket
// benutzt. Wer erst speichern muss, um zu sehen, was er getan hat, probiert an
// allen Sets zugleich.
//
// # Welche Karte gezeichnet wird
//
// **Die erste Karte des ersten Sets, das diesen Stil benutzt.** Eine beliebige
// Karte zeigte die Änderung nicht — sie hinge an einer anderen Vorlage —, und
// die Wahl dem Fenster zu überlassen hieße, dieselbe Suche dort noch einmal zu
// schreiben.
//
// **Was das bis Phase 5.1 nicht war**: eine Auswahl. Seit Werkstück U7
// (#1418)
// darf die Werkstatt eine Karte **nennen** — eine Testkarte mit langem Text,
// fehlendem Bild oder zweiter Seite zeigt, was die erste Karte des ersten
// Sets nicht zeigt. Ohne Nennung bleibt es bei der Probekarte: wiederholbar,
// und ohne dass das Fenster die Suche nachbauen müsste.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Vorlagenzeichner baut ein Zeichenwerk, in dem eine Paketdatei ersetzt ist.
//
// **Als Naht und nicht als Aufruf**: ein Zeichenwerk hängt an Datenbank und
// Paketverzeichnis, und beides kennt `App` nicht.
type Vorlagenzeichner interface {
	// Probekarte gibt eine Karte, die diesen Stil benutzt — oder mit einer
	// Kennung über 0 genau die genannte, gleich welchen Stil sie trägt.
	Probekarte(ctx context.Context, paket string, karteID int64) (*domain.CardRow, error)

	// Probekarten gibt die Karten des ersten Sets, das diesen Stil benutzt —
	// die Wahl der Werkstatt (U7, #1418).
	Probekarten(ctx context.Context, paket string) ([]*domain.CardRow, error)

	// MitErsatz gibt Planbauer und Rasterer, die die ersetzte Datei sehen.
	MitErsatz(ersatz service.Paketersatz) (*service.Planbauer, Rasterer)
}

// MitVorlagenzeichner setzt die Naht.
func (a *App) MitVorlagenzeichner(zeichner Vorlagenzeichner) { a.vorlagenzeichner = zeichner }

// Vorlagenvorschau zeichnet eine Karte, als stünde die Änderung schon im Paket.
//
// **Der Abbruch gilt** (§7): der Aufruf teilt sich den Lauf mit
// [App.Vorschau], und ein neuer Tastendruck bricht den vorigen ab. Eine
// Vorlage zu zeichnen kostet mehr als ein Kartenfeld — vier Läufe beim Tippen
// zeigten sonst am Ende den drittletzten Stand.
//
// **`karteID` ist leer oder eine Kennung.** Leer heißt: die Probekarte des
// Pakets. Eine unlesbare Kennung ist ein Validierungsfehler am Feld `id`,
// wie bei jeder anderen Bindung, die eine Kennung entgegennimmt.
func (a *App) Vorlagenvorschau(
	ctx context.Context, paket, datei, inhalt string, vergroesserung float64, karteID string,
) VorschauAntwort {
	ctx = a.imProfil(ctx)

	ctx, fertig := a.vorschaulauf.beginne(ctx)
	defer fertig()

	if a.vorlagenzeichner == nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, keinVorlagenzeichner())}
	}

	gewuenscht, kennungsfehler := probekartenkennung(karteID)
	if kennungsfehler != nil {
		return VorschauAntwort{Fehler: kennungsfehler}
	}

	karte, err := a.vorlagenzeichner.Probekarte(ctx, paket, gewuenscht)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	planbauer, rasterer := a.vorlagenzeichner.MitErsatz(service.Paketersatz{
		Paket: paket, Datei: datei, Inhalt: []byte(inhalt),
	})

	// **Ein eigener Planbauer für diesen Lauf.** Der gemeinsame merkt sich
	// Stile; ein Stand, den niemand geschrieben hat, gehört dort nicht hinein.
	//
	// **Der Rasterer bleibt derselbe** — er kennt keine Pakete.
	return a.zeichneVorschau(ctx, karte, vergroesserung, planbauer, rasterer)
}

// keinVorlagenzeichner meldet die fehlende Naht.
func keinVorlagenzeichner() error {
	return fmt.Errorf("%w: der Vorlagenzeichner ist nicht eingerichtet", domain.ErrValidation)
}

// probekartenkennung liest die gewünschte Karte: leer ist keine Wahl.
func probekartenkennung(karteID string) (int64, *Fehler) {
	if karteID == "" {
		return 0, nil
	}

	return kennungAus(karteID)
}
