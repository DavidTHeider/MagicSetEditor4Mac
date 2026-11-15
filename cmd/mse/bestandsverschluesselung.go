// **Die Verschlüsselung des Bestands an der Kommandozeile** — Werkstück 14 von
// Phase 4.12 (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// **Dieselbe Fähigkeit wie am Fenster** (§16): was das Fenster kann, kann auch
// diese Fläche.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// bestandVerschluesselungBefehl zeigt den Stand oder schaltet um.
//
// # Ohne Argument zeigen, mit Argument schalten
//
// **Weil das Zeigen der häufigere und der harmlosere Ruf ist.** Wer
// `bestand-verschluesselung` tippt, will wissen, wie es steht — nicht es
// ändern. Das Ändern verlangt ein Wort, und zwar ein eindeutiges.
func bestandVerschluesselungBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) > 1 {
		an.klage("bestand-verschluesselung nimmt höchstens ein Wort: an oder aus.\n%s", hilfe)

		return CodeBedienung
	}

	var (
		schalten bool
		ziel     bool
	)

	if len(rest) == 1 {
		switch rest[0] {
		case "an":
			schalten, ziel = true, true
		case "aus":
			schalten, ziel = true, false
		default:
			an.klage("bestand-verschluesselung kennt nur an und aus, nicht %q.\n%s",
				rest[0], hilfe)

			return CodeBedienung
		}
	}

	return mitBestandsgriff(ctx, bestandsort, an, func(ctx context.Context, db **sql.DB) int {
		naht := app.BestandsnahtUeber(db, bestandsort)

		if schalten {
			// **Der Stand kommt aus dem Umschalten selbst.** Ein zweiter Ruf
			// danach wäre eine zweite Frage an eine Datei, die gerade
			// getauscht wurde — und ein Fehlerweg, den kein Fall betritt.
			stand, fehler := naht.VerschluesselungUmschalten(ctx, ziel)
			if fehler != nil {
				an.klageUeber(fehler)

				return CodeFachlich
			}

			an.sag("%s\n", verschluesselungssatz(stand))

			return CodeGut
		}

		stand, fehler := naht.Verschluesselungsstand()
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		an.sag("%s\n", verschluesselungssatz(stand))

		return CodeGut
	})
}

// verschluesselungssatz sagt in einem Satz, wie der Bestand liegt.
//
// **Der Hinweis auf den Schlüsselbund gehört dazu**, und nicht nur an die
// Oberfläche: wer hier verschlüsselt und die Bedeutung des Schlüsselbunds
// nicht kennt, verliert seine Daten beim nächsten Rechnerwechsel.
func verschluesselungssatz(an bool) string {
	if an {
		return "Der Bestand ist verschlüsselt. Der Schlüssel liegt im " +
			"Schlüsselbund dieses Rechners und lässt sich nicht neu berechnen — " +
			"ohne ihn kommt niemand mehr an die Daten."
	}

	return "Der Bestand liegt im Klartext."
}
