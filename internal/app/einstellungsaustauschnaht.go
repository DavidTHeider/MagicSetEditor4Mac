// **Der Austausch der Einstellungen an den drei Flächen ohne Fenster** —
// Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum sie am [Einstellungsweg] hängen und nicht an der App
//
// **Weil `App` ein Fenster ist.** Sie hält den Profilzeiger, die Umschaltung,
// die Menüleiste und sechzehn Wege durch einen Dateidialog; die
// Kommandozeile, der MCP-Server und der Dienst brauchen davon nichts. Der Weg
// ist die schmale Naht, die sie stattdessen nehmen — dieselbe wie beim Lesen
// und Setzen einer Einstellung.
//
// # Und warum sie einen `error` geben statt eines `Fehler`
//
// **Weil dort ein Rückgabewert daraus wird**: die Kommandozeile macht einen
// Abbruchcode, der Dienst einen Status, der MCP-Server eine Fehlermeldung im
// Werkzeugergebnis. Ein übersetzter `*Fehler` wäre für alle drei die falsche
// Gestalt — und für das Fenster ist er die richtige, deshalb steht die
// Übersetzung dort.

package app

import (
	"context"
	"fmt"
)

// Ausfuhr gibt den Austausch als Text.
func (w *Einstellungsweg) Ausfuhr(
	ctx context.Context, abschnitte []string,
) (string, error) {
	austausch, fehler := w.austausch.Ausfuhr(ctx, abschnitte)
	if fehler != nil {
		return "", fehler
	}

	roh, packfehler := auskunftPacken(austausch)
	if packfehler != nil {
		return "", fmt.Errorf("den Austausch packen: %w", packfehler)
	}

	return string(roh), nil
}

// Einfuhrvorschau sagt, was ein Text ändern würde — und ändert nichts.
func (w *Einstellungsweg) Einfuhrvorschau(
	ctx context.Context, roh []byte,
) (Einstellungssicht, error) {
	sicht, fehler := einstellungssichtAus(w.austausch.Vorschau(ctx, roh))
	if fehler != nil {
		return sicht, fehler
	}

	sicht.Inhalt = string(roh)

	return sicht, nil
}

// Einfuhruebernahme schreibt die gewählten Zeilen.
func (w *Einstellungsweg) Einfuhruebernahme(
	ctx context.Context, roh []byte, revision string, nur []string,
) (Einstellungsbericht, error) {
	return einstellungsberichtAus(
		w.austausch.Uebernehmen(ctx, roh, revision, nur))
}

// **Die drei Bindungen des Fensters** — Werkstück O3 von Phase 5.2 (#1427).
//
// **Hier und nicht in `bindungen.go`**, weil jene Datei mit ihnen über 400
// Zeilen ginge (§2) — derselbe Schnitt wie bei den Nutzerbindungen daneben.

// EinstellungenSichern fragt nach einem Ziel und schreibt den Austausch.
func (b *Bindungen) EinstellungenSichern(ctx context.Context, abschnitte []string) SchreibAntwort {
	return b.app(ctx).EinstellungenSichern(b.ruf(ctx), abschnitte)
}

// EinstellungenEinlesen fragt nach einer Datei und zeigt, was sie änderte.
func (b *Bindungen) EinstellungenEinlesen(ctx context.Context) EinfuhrvorschauAntwort {
	return b.app(ctx).EinstellungenEinlesen(b.ruf(ctx))
}

// EinstellungenUebernehmen schreibt die gewählten Zeilen.
func (b *Bindungen) EinstellungenUebernehmen(
	ctx context.Context,
	inhalt, revision string, nur []string,
) EinfuhrAntwort {
	return b.app(ctx).EinstellungenUebernehmen(b.ruf(ctx), inhalt, revision, nur)
}
