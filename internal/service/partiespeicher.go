package service

// **Was die Partieleitung zum Lesen und Schreiben braucht** — Werkstück G1
// von Phase 5.5
// (#1458).
//
// # Eine eigene Datei
//
// **Die Zeilengrenze** (§2): `partieleitung.go` stand mit ihr bei 402
// Zeilen. Der Schnitt folgt der Frage — dort steht, **was** die Leitung tut,
// hier, **was sie dafür braucht**.
//
// **Die Schnittstelle steht in `service` und wird in `storage` erfüllt** —
// nie umgekehrt (§2). Dass `storage` sie ohne Import erfüllt, ist eine Folge
// von Gos struktureller Typisierung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partiespeicher ist, was die Leitung zum Lesen und Schreiben braucht.
type Partiespeicher interface {
	Anlegen(ctx context.Context, partie domain.Partie) (domain.Partie, error)
	Partie(ctx context.Context, kennung string) (domain.Partie, error)

	// Partien gibt die Partien dieses Profils, zuletzt geänderte zuerst
	// (#1600).
	//
	// **Die Ablage hatte sie seit G1 und niemand rief sie** — genau die
	// Gestalt aus §10, und sie kostete den Wirt den Weg zurück in seine
	// eigene Partie.
	Partien(ctx context.Context) ([]domain.Partie, error)

	// Anwenden schreibt Zustand und Ereignisse in **einer** Transaktion.
	//
	// **Eine und nicht zwei**: ein Zustand ohne seine Ereignisse ließe sich
	// nicht mehr erklären, und Ereignisse ohne ihren Zustand wären ein
	// Strom, der ins Leere führt (#1458: „atomar persistieren").
	Anwenden(ctx context.Context, erwartet int64, befehlID string,
		zustand domain.Partie, ereignisse []domain.Ereignis) error

	// Befehlsergebnis gibt die Ereignisse eines schon angewendeten Befehls.
	//
	// **Der zweite Rückgabewert sagt, ob es ihn gab** — und nicht eine leere
	// Liste, die wie „hat nichts bewirkt" aussähe.
	Befehlsergebnis(ctx context.Context, partieID, befehlID string) (
		[]domain.Ereignis, bool, error)

	// LetzterBefehl gibt die Ereignisse des zuletzt angewendeten Befehls —
	// keine bei einer Partie ohne einen (P8, #2376).
	LetzterBefehl(ctx context.Context, partieID string) ([]domain.Ereignis, error)

	// LetzteKette gibt den Abdruck des letzten Ereignisses — leer bei einer
	// Partie ohne eines.
	LetzteKette(ctx context.Context, partieID string) (string, int64, error)

	// Standsetzen schreibt Pause, Fortsetzung oder Ende.
	Standsetzen(ctx context.Context, partieID string,
		stand domain.Partiestand) (domain.Partie, error)

	// Ereignisse gibt den Strom einer Partie — **ungefiltert**.
	//
	// **Und das ist Absicht** (G5,
	// #1462):
	// der Strom ist die Wahrheit, aus der sich jeder Stand nachrechnen lässt,
	// und deshalb trägt er alles. Wer ihn **sehen** darf und wie viel davon,
	// entscheidet [Partieleitung.Verlauf] — an **einer** Stelle, nicht in
	// jedem Speicher noch einmal.
	//
	// **Null und null bedeuten: alles.**
	Ereignisse(ctx context.Context, partieID string, von, bis int64) (
		[]domain.Ereignis, error)

	// Freigaben gibt die Einwilligungen einer Partie (G5, #1462).
	Freigaben(ctx context.Context, partieID string) (
		[]domain.Offenlegung, error)

	// Freigeben schreibt die Einwilligung **eines** Sitzes — und ersetzt
	// die alte.
	Freigeben(ctx context.Context, partieID string,
		gabe domain.Offenlegung) error

	// Freigabeloeschen nimmt sie zurück.
	Freigabeloeschen(ctx context.Context, partieID, von string) error

	Partieverwaltungsspeicher // die drei Verwaltungswege, nebenan (#1711)

	// Vergessen löscht eine Partie samt Strom, Befehlen und Freigaben.
	Vergessen(ctx context.Context, partieID string) error
}
