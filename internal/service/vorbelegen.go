package service

// Das Vorbelegen einer Umgebung aus einem `init script:` — und was dabei
// nur mit Vorbehalt gelang (#216).

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// vorbelegt führt ein Initskript aus und macht aus einem Fehler einen Befund.
//
// **Ein kaputtes Initskript hält den Plan nicht auf** (§8). Eine Vorlage von
// 2007 mit einer Zeile, die unser Zerteiler nicht liest, ergibt eine Karte mit
// den Feldern, die ohne Vorbelegung auskommen — und einen Befund, der sagt,
// warum die anderen fehlen.
func vorbelegt(
	ctx context.Context, skript, ort string, umgebung *msescript.Umgebung,
) []domain.Befund {
	if skript == "" {
		return nil
	}

	vorbehalte, err := msescript.VorbelegenMitVorbehalten(ctx, skript, umgebung)
	if err != nil {
		return []domain.Befund{{
			Ort:     ort,
			Meldung: "das " + ort + " ließ sich nicht ausführen: " + err.Error(),
		}}
	}

	// **Was weitergerechnet wurde, wird trotzdem genannt** (#216).
	//
	// Ein Vorbehalt ist kein Abbruch, aber auch kein Nichts: an dieser Stelle
	// hat das Skript etwas anderes getan, als die Vorlage sagt. Wer die
	// Meldung nicht bekäme, hätte den stillen Weg — genau den, vor dem der
	// Kopf von [msescript.Vorbehalt] warnt.
	befunde := make([]domain.Befund, 0, len(vorbehalte))
	for _, vorbehalt := range vorbehalte {
		befunde = append(befunde, domain.Befund{
			Ort:     ort + ", Zeile " + strconv.Itoa(vorbehalt.Zeile),
			Meldung: vorbehalt.Meldung,
		})
	}

	return befunde
}
