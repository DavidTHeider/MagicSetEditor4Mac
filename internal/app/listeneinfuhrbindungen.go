// **Die Bindungen der Listeneinfuhr** — Werkstück 3 von Phase 4.5
// (#1292).
//
// **Eine eigene Datei neben `sammlungsbindungen.go`**, weil sie eine andere
// Quelle braucht: die Sammlung kommt aus dem Bestand, die Deutung aus
// Scryfall. Der Schnitt folgt der Frage, nicht der Zeilenzahl (§2).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Listeneinfuhrquelle ist, was das Fenster zum Einlesen einer Liste braucht.
//
// **Als Schnittstelle und nicht als `*Listeneinfuhrnaht`** (§2): so kommt
// keine Bindung ohne Zuhause zum Absturz, und ein Lauf braucht kein Netz.
type Listeneinfuhrquelle interface {
	Deuten(ctx context.Context, text string) (*Listendeutung, error)
	Einlesen(ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string) (
		*Einfuhrbericht, error)

	// Vorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	Vorschauen(ctx context.Context, zeilen []Listeneinfuhrzeile) (*Sammelvorschau, error)
}

// MitListeneinfuhr setzt die Naht.
func (a *App) MitListeneinfuhr(quelle Listeneinfuhrquelle) { a.listeneinfuhr = quelle }

// Sammlungsausfuhrquelle ist, was das Fenster zum Ausgeben braucht.
type Sammlungsausfuhrquelle interface {
	Ausgeben(ctx context.Context, form string) (string, error)
	CSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (*Einfuhrbericht, error)

	// CSVVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	CSVVorschauen(ctx context.Context, text string) (*Sammelvorschau, error)
}

// MitSammlungsausfuhr setzt die Naht.
func (a *App) MitSammlungsausfuhr(quelle Sammlungsausfuhrquelle) {
	a.sammlungsausfuhr = quelle
}

// Ausfuhrtextantwort trägt den ausgegebenen Text oder einen Fehler.
type Ausfuhrtextantwort struct {
	Form   string  `json:"form"`
	Text   string  `json:"text"`
	Fehler *Fehler `json:"fehler"`
}

// SammlungAusgeben schreibt die Sammlung als Text.
//
// **`liste` ist verlustbehaftet, `csv` nicht.** Was die Liste nicht trägt,
// steht bei `service.FormListe`.
func (a *App) SammlungAusgeben(ctx context.Context, form string) Ausfuhrtextantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungsausfuhr == nil {
		return Ausfuhrtextantwort{Fehler: Uebersetze(ctx, keineSammlungsausfuhr())}
	}

	text, err := a.sammlungsausfuhr.Ausgeben(ctx, form)
	if err != nil {
		return Ausfuhrtextantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Ausfuhrtextantwort{Form: form, Text: text}
}

// SammlungCSVEinlesen liest ein CSV zurück in den Bestand.
//
// **`quelle` sagt, woher es kam** — Werkstück 12 von Phase 4.12 (#1351): ein
// Dateiname, ein Pfad, eine Bezeichnung. Sie wird an jeder **neu**
// entstehenden Zeile notiert; leer heißt „niemand hat es gesagt", und dann
// wird nichts behauptet.
//
// **`zusammenfuehrung` sagt, was mit einer vorhandenen Zeile geschieht** —
// `addieren` oder `auslassen`, leer heißt addieren (DOM-03.AK-01, #1698).
func (a *App) SammlungCSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) Einfuhrantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungsausfuhr == nil {
		return Einfuhrantwort{Fehler: Uebersetze(ctx, keineSammlungsausfuhr())}
	}

	bericht, err := a.sammlungsausfuhr.CSVEinlesen(ctx, text, quelle, zusammenfuehrung)

	// **Dasselbe hier** (Befund 3).
	return Einfuhrantwort{Bericht: bericht, Fehler: Uebersetze(ctx, err)}
}

// Listendeutungsantwort trägt eine Deutung oder einen Fehler.
//
// **Ein Umschlag und kein zweiter Rückgabewert**: Wails nimmt bei zwei
// Rückgaben den zweiten als Fehler und verliert die Struktur.
type Listendeutungsantwort struct {
	Deutung *Listendeutung `json:"deutung"`
	Fehler  *Fehler        `json:"fehler"`
}

// Einfuhrantwort trägt einen Bericht oder einen Fehler.
type Einfuhrantwort struct {
	Bericht *Einfuhrbericht `json:"bericht"`
	Fehler  *Fehler         `json:"fehler"`
}

// ListeDeuten liest eine Kartenliste und ordnet jede Zeile zu.
//
// **Sie schreibt nichts** (§16, Regel 6). Was daraus wird, entscheidet, wer
// hinsieht — „Counterspell“ ohne Setangabe passt auf 72 Drucke.
func (a *App) ListeDeuten(ctx context.Context, text string) Listendeutungsantwort {
	ctx = a.imProfil(ctx)

	if a.listeneinfuhr == nil {
		return Listendeutungsantwort{Fehler: Uebersetze(ctx, keineListeneinfuhr())}
	}

	deutung, err := a.listeneinfuhr.Deuten(ctx, text)
	if err != nil {
		return Listendeutungsantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Listendeutungsantwort{Deutung: deutung}
}

// ListeEinlesen nimmt die entschiedenen Zeilen in die Sammlung auf.
func (a *App) ListeEinlesen(
	ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string,
) Einfuhrantwort {
	ctx = a.imProfil(ctx)

	if a.listeneinfuhr == nil {
		return Einfuhrantwort{Fehler: Uebersetze(ctx, keineListeneinfuhr())}
	}

	bericht, err := a.listeneinfuhr.Einlesen(ctx, zeilen, zusammenfuehrung)

	// **Bericht und Fehler zugleich** (§7, Regel 3, Befund 3 des Zweitblicks
	// über Phase 4.5): was geschrieben ist, ist geschrieben — und das Fenster
	// muss es sehen, sonst schreibt ein zweiter Versuch dasselbe noch einmal.
	return Einfuhrantwort{Bericht: bericht, Fehler: Uebersetze(ctx, err)}
}

// SammlungAusgeben schreibt die Sammlung als Text.
func (b *Bindungen) SammlungAusgeben(ctx context.Context, form string) Ausfuhrtextantwort {
	return b.app(ctx).SammlungAusgeben(b.ruf(ctx), form)
}

// SammlungCSVEinlesen liest ein CSV zurück in den Bestand.
//
// **`quelle` sagt, woher es kam** — Werkstück 12 (#1351). Das Fenster gibt
// den Dateinamen weiter; wer aus der Zwischenablage einfügt, gibt nichts, und
// dann steht auch nichts da.
func (b *Bindungen) SammlungCSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) Einfuhrantwort {
	return b.app(ctx).SammlungCSVEinlesen(b.ruf(ctx), text, quelle, zusammenfuehrung)
}

// ListeDeuten gibt zu einer Kartenliste die passenden Drucke.
func (b *Bindungen) ListeDeuten(ctx context.Context, text string) Listendeutungsantwort {
	return b.app(ctx).ListeDeuten(b.ruf(ctx), text)
}

// ListeEinlesen nimmt die entschiedenen Zeilen in die Sammlung auf.
func (b *Bindungen) ListeEinlesen(ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string) Einfuhrantwort {
	return b.app(ctx).ListeEinlesen(b.ruf(ctx), zeilen, zusammenfuehrung)
}

// keineListeneinfuhr ist der Fehler ohne eingerichtete Naht.
//
// **Sie kann fehlen**, und das ist kein Programmfehler: `ListeneinfuhrUeber`
// braucht ein Zuhause für den Zwischenspeicher, und ohne das gibt es sie
// nicht. Die Bindung sagt es dann selbst, statt dass die Seite leer bleibt
// (§8) — dieselbe Gestalt wie bei den Preisen.
func keineSammlungsausfuhr() error {
	return fmt.Errorf("%w: die Sammlungsausfuhr ist nicht eingerichtet",
		domain.ErrValidation)
}

func keineListeneinfuhr() error {
	return fmt.Errorf("%w: die Listeneinfuhr ist nicht eingerichtet — sie "+
		"braucht ein Zuhause für den Zwischenspeicher von Scryfall",
		domain.ErrValidation)
}
