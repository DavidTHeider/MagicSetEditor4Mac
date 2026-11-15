// **Der Dateibaum der Werkstatt** — FMT-02.AK-01
// (#1691).
//
// # Was hier entschieden wird
//
// Wie und mit welcher Sicherung geschrieben wird, steht in
// `msepaket/dateibaum.go`. Hier steht, **was** erlaubt ist:
//
//   - **Angelegt wird eine Vorlage** — eine Textdatei mit Inhalt. Ein leeres
//     `rahmen.png` wäre kein Bild, sondern ein Fehler, der erst beim Zeichnen
//     auffiele.
//   - **Eingefügt wird eine Datei von der Platte** — ein Bild, eine Schrift.
//     **Nur am Fenster und an der Kommandozeile**: beide handeln mit den
//     Rechten dessen, der davor sitzt. Über REST oder MCP wäre es ein Weg,
//     eine beliebige Datei des Rechners in ein Paket zu legen und dann über
//     `GET …/dateien` wieder hinauszulesen (`docs/FAEHIGKEITEN.md`).
//   - **Umbenannt und entfernt wird jede Datei**, Vorlage oder Beiwerk —
//     **eine Vorlage nur mit ihrem Stand**: ohne ihn wäre nicht zu erkennen,
//     ob inzwischen jemand anderes geschrieben hat (§7, Regel 2). Beiwerk
//     liest die Werkstatt nicht und hat deshalb keinen; dass es sich
//     trotzdem zurückholen lässt, sagt die Sicherung.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Anlegen legt eine neue Vorlage an und gibt ihren Stand.
func (v *Paketwerkstatt) Anlegen(ctx context.Context, paket, datei, inhalt string) (string, error) {
	if !Bearbeitbar(datei) {
		return "", nichtBearbeitbar(datei)
	}

	err := v.amPaket(ctx, paket, func(ort string) error {
		return msepaket.DateiAnlegen(ctx, ort, datei, []byte(inhalt))
	})
	if err != nil {
		return "", err
	}

	return Stand([]byte(inhalt)), nil
}

// Einfuegen legt eine Datei von der Platte als `datei` ins Paket.
func (v *Paketwerkstatt) Einfuegen(ctx context.Context, paket, quelle, datei string) error {
	return v.amPaket(ctx, paket, func(ort string) error {
		inhalt, err := msepaket.LeseDateiBegrenzt(quelle)
		if err != nil {
			return domain.MitGrund(domain.ErrValidation, err,
				"%q ließ sich nicht lesen — prüfe, ob die Datei noch dort liegt", quelle)
		}

		return msepaket.DateiAnlegen(ctx, ort, datei, inhalt)
	})
}

// Umbenennen gibt einer Datei des Pakets einen neuen Namen.
func (v *Paketwerkstatt) Umbenennen(ctx context.Context, paket, alt, neu, erwartet string) error {
	if err := mitStand(alt, erwartet); err != nil {
		return err
	}

	return v.amPaket(ctx, paket, func(ort string) error {
		return msepaket.DateiUmbenennen(ctx, ort, alt, neu, erwartet)
	})
}

// Entfernen nimmt eine Datei aus dem Paket — in die Sicherung neben ihm.
func (v *Paketwerkstatt) Entfernen(ctx context.Context, paket, datei, erwartet string) error {
	if err := mitStand(datei, erwartet); err != nil {
		return err
	}

	return v.amPaket(ctx, paket, func(ort string) error {
		return msepaket.DateiEntfernen(ctx, ort, datei, erwartet)
	})
}

// mitStand verlangt den Stand einer Vorlage.
func mitStand(datei, erwartet string) error {
	if Bearbeitbar(datei) && erwartet == "" {
		return domain.NewFieldError("stand", domain.SchluesselValidierung,
			"%s ist eine Vorlage — ohne ihren Stand ließe sich nicht erkennen, ob "+
				"inzwischen jemand anderes geschrieben hat. Lies sie zuerst.", datei)
	}

	return nil
}

// amPaket führt einen Handgriff am Paket aus — unter der geteilten Sperre am
// Bezugsbaum, wie [Paketwerkstatt.Schreiben] (#1190): ein Bezug tauscht den
// Baum aus, und ein Umbenennen darin sähe er sonst nicht.
func (v *Paketwerkstatt) amPaket(ctx context.Context, paket string, handgriff func(ort string) error) error {
	ort, err := v.ort(paket)
	if err != nil {
		return err
	}

	fertig, sperrfehler := GeteiltAmBaum(ctx, bezugsbaum(v.wurzel, ort))
	if sperrfehler != nil {
		return fmt.Errorf("das Paket sperren: %w", sperrfehler)
	}

	defer fertig()

	return handgriff(ort)
}
