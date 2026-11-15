package service

// **Ein Turnier ausgeben und einlesen** — DOM-04.AK-03
// (#1699),
// Vermerk [0140](../../docs/entscheidungen/0140-ein-turnier-als-datei.md).
//
// # Warum das Format über seinen Namen reist
//
// **Weil eine Kennung nur in ihrem Bestand gilt.** Die Datei nennt „Pauper";
// dieser Dienst sucht es unter den Formaten desselben Spiels im Profil. Gibt
// es keines, kommt das Turnier **ohne** Format — und der Befund sagt es
// (§8). Ein Nein wäre zu streng: das Format ist eine Angabe und kein Zwang
// (Vermerk 0138), und wer eine Datei einliest, will den Verlauf.
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausgeben schreibt ein Turnier samt Verlauf als Datei — ohne eine Kennung
// dieses Bestands.
func (l *Turnierleitung) Ausgeben(ctx context.Context, turnierID int64) ([]byte, error) {
	turnier, format, err := l.mitFormatnamen(ctx, turnierID)
	if err != nil {
		return nil, err
	}

	return domain.FormeTurnier(domain.AusgabeVon(turnier, format)), nil
}

// Bericht schreibt den Abschlussbericht — in den Wörtern, die die Fläche
// mitgibt ([domain.Turnierbericht]).
func (l *Turnierleitung) Bericht(
	ctx context.Context, turnierID int64, texte domain.Berichttexte,
) (string, error) {
	turnier, format, err := l.mitFormatnamen(ctx, turnierID)
	if err != nil {
		return "", err
	}

	return domain.Turnierbericht(turnier, format, texte), nil
}

// mitFormatnamen liest ein Turnier samt dem Namen seines Formats — leer,
// wenn keines gewählt ist.
func (l *Turnierleitung) mitFormatnamen(ctx context.Context, turnierID int64) (
	domain.Turnier, string, error,
) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil || turnier.FormatID == 0 {
		return turnier, "", err
	}

	if nahtfehler := l.formatnahtDa(); nahtfehler != nil {
		return domain.Turnier{}, "", nahtfehler
	}

	format, err := l.formate.Format(ctx, turnier.FormatID)
	if err != nil {
		return domain.Turnier{}, "", err
	}

	return turnier, format.Name, nil
}

// Einlesen legt ein Turnier aus einer Datei an — samt Verlauf, **in einem
// Zug** (`storage.Turnierablage.Einlesen`).
func (l *Turnierleitung) Einlesen(ctx context.Context, roh []byte) (
	domain.Turnier, []domain.Einfuhrbefund, error,
) {
	austausch, befunde, err := domain.ZerlegeTurnier(roh)
	if err != nil {
		return domain.Turnier{}, nil, err
	}

	turnier := austausch.Turnier.AlsTurnier()

	if name := austausch.Turnier.Format; name != "" {
		kennung, suchfehler := l.formatSuchen(ctx, turnier.Spiel, name)
		if suchfehler != nil {
			return domain.Turnier{}, nil, suchfehler
		}

		if kennung == 0 {
			befunde = append(befunde, domain.Einfuhrbefund{
				Schluessel: "format",
				Grund: fmt.Sprintf("Das Format »%s« gibt es hier für »%s« nicht; "+
					"das Turnier kommt ohne Format.", name, turnier.Spiel),
			})
		}

		turnier.FormatID = kennung
	}

	eingelesen, err := l.turniere.Einlesen(ctx, turnier)
	if err != nil {
		return domain.Turnier{}, nil, err
	}

	return eingelesen, befunde, nil
}

// formatSuchen gibt die Kennung des Formats mit diesem Namen — **null, wenn
// es keines gibt**. Verglichen wird der Name genau so, wie er dasteht.
func (l *Turnierleitung) formatSuchen(ctx context.Context, spiel, name string) (int64, error) {
	if err := l.formatnahtDa(); err != nil {
		return 0, err
	}

	formate, err := l.formate.Formate(ctx, spiel)
	if err != nil {
		return 0, err
	}

	for _, format := range formate {
		if format.Name == name {
			return format.ID, nil
		}
	}

	return 0, nil
}

// formatnahtDa sagt, ob sich hier Formate nachschlagen lassen — **ohne Naht
// wird keines angenommen**, wie in [Turnierleitung.formatPruefen].
func (l *Turnierleitung) formatnahtDa() error {
	if l.formate == nil {
		return fmt.Errorf("%w: hier lassen sich keine Formate nachschlagen", domain.ErrValidation)
	}

	return nil
}
