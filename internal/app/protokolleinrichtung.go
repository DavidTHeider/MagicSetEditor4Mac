package app

// **Die Weiche nach den Einstellungen richten und in der Diagnose zeigen** —
// OPS-04 (#1721).

import (
	"context"
	"log/slog"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitProtokollweiche gibt der Anwendung die Weiche, die `main` als Standard
// gesetzt hat — sie richtet sie ein, sobald der Bestand offen ist.
func (a *App) MitProtokollweiche(weiche *Protokollweiche) *App {
	a.protokollweiche = weiche

	return a
}

// protokollEinrichten richtet die Weiche nach den Einstellungen.
//
// **Ein Lesefehler lässt die Vorgabe stehen** — das Protokoll schreibt
// weiter, und die Zeile darüber sagt, warum es die Einstellungen nicht kennt.
func (a *App) protokollEinrichten(ctx context.Context) {
	if a.protokollweiche == nil {
		return
	}

	werte, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		slog.WarnContext(ctx, "die Einstellungen des Protokolls ließen sich nicht lesen", "grund", fehler.Error())

		return
	}

	a.protokollweiche.Einrichten(domain.ProtokollregelAus(werte))
}

// protokolldatei gibt die Datei, in die geschrieben wird — ohne Weiche oder
// ohne Dateikanal die neben dem Bestand.
func (a *App) protokolldatei() (string, error) {
	if a.protokollweiche != nil {
		if datei := a.protokollweiche.Stand().Datei; datei != "" {
			return datei, nil
		}
	}

	return Protokollpfad()
}

// protokollzeilen sagt in der Diagnose, was das Protokoll tut — leer ohne
// Weiche, denn dann gibt es nur die Datei, und die steht schon da.
func (a *App) protokollzeilen(texte Dialogtexte) string {
	if a.protokollweiche == nil {
		return ""
	}

	stand := a.protokollweiche.Stand()
	namen := map[string]string{
		KanalDatei: texte.DiagnoseKanalDatei, KanalSystem: texte.DiagnoseKanalSystem,
		KanalAusgabe: texte.DiagnoseKanalAusgabe,
	}

	teile := make([]string, 0, len(stand.Kanaele))

	for _, kanal := range stand.Kanaele {
		zustand := texte.DiagnoseKanalAus

		switch {
		case kanal.Aktiv:
			zustand = texte.DiagnoseKanalAn
		case kanal.Grund != "":
			zustand = texte.DiagnoseKanalFehlt + " (" + kanal.Grund + ")"
		}

		teile = append(teile, namen[kanal.Name]+" "+zustand)
	}

	zeilen := texte.DiagnoseKanaele + ": " + strings.Join(teile, ", ") + " — " +
		texte.DiagnoseStufe + " " + stand.Stufe + ", " + stand.Format + "\n"

	if stand.Ausfuehrlich {
		zeilen += texte.DiagnoseAusfuehrlich + "\n"
	}

	return zeilen
}
