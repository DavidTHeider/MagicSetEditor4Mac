package app

import "context"

// ueberNutzer fragt, ob der arbeitende Mensch über `ziel` schreiben darf —
// ID-03 (#1702).
//
// # Warum an der Bindung und nicht in jedem Dienst
//
// **Weil alle vier Flächen hier durchgehen**: Kommandozeile, MCP und Dienst
// reichen über [Nutzerflaeche] an dieselben Bindungen. Die Regel selbst steht
// in `service.Nutzerverwaltung.UeberNutzer` — hier wird nur gefragt, wie
// [App.darf] fragt (§2). Die Dienste, um die es geht — Faktoren, Konten,
// Einstellungen, Personendaten, Tischbild —, kennen die Nutzerverwaltung
// nicht, und sie ihnen allen hineinzureichen hieße, dieselbe Frage an fünf
// Stellen zu verdrahten.
//
// **Gefragt wird vor jeder Wirkung** — auch vor einem Dateidialog: wer erst
// nach dem Ort fragt und dann absagt, hat jemanden umsonst suchen lassen.
//
// # Was ein `nil` hier heißt
//
// **Ohne Nutzerverwaltung gibt es keinen fremden Menschen**: eine Anwendung,
// die niemanden kennt, kennt auch niemanden, über den sie nicht schreiben
// dürfte. Beide Wurzeln setzen sie; ein Lauf, der sie weglässt, prüft etwas
// anderes.
func (a *App) ueberNutzer(ctx context.Context, ziel int64) error {
	// **Ein Token verwaltet keinen Menschen** (QA-09, #1740).
	if fehler := keinTokenAmMenschen(ctx); fehler != nil {
		return fehler
	}

	if a.nutzer == nil {
		return nil
	}

	return a.nutzer.UeberNutzer(ctx, ziel)
}

// ueberProfil fragt dasselbe für ein Profil — über seinen Eigner, und im
// gezeigten Profil nach der Rolle (der native Befund F01).
//
// **Ohne Zweig für eine fehlende Nutzerverwaltung**, anders als
// [App.ueberNutzer]: jede Bindung, die hier fragt, liest das Profil danach
// über sie. Ein Zweig, den sie nie erreichten, wäre eine Behauptung (§3).
func (a *App) ueberProfil(ctx context.Context, profilID int64) error {
	// **Ein Token wirkt nur an seinem Profil** (QA-09, #1740).
	if fehler := nurAnSeinemProfil(ctx, profilID); fehler != nil {
		return fehler
	}

	return a.nutzer.UeberProfil(ctx, profilID)
}

// nutzerVorab liest eine Nutzerkennung und fragt, **bevor** ein Dialog
// aufgeht — für die Bindungen, die erst nach einer Datei fragen.
func (a *App) nutzerVorab(ctx context.Context, nutzerID string) *Fehler {
	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return befund
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return Uebersetze(ctx, fremd)
	}

	return nil
}

// profilVorab ist dasselbe für ein Profil.
func (a *App) profilVorab(ctx context.Context, profilID string) *Fehler {
	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return befund
	}

	if fremd := a.ueberProfil(ctx, kennung); fremd != nil {
		return Uebersetze(ctx, fremd)
	}

	return nil
}
