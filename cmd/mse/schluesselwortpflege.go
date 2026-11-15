// Die Handlungen an einem Schlüsselwort
// (#1062).
//
// **Eine eigene Datei**, weil `schluesselwoerter.go` mit ihnen über die
// Zeilengrenze ginge (§2). Der Schnitt folgt der Frage: dort **zeigen**, hier
// **ändern**.
package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// schluesselworthandlung führt eine der drei Handlungen aus.
func schluesselworthandlung(
	ctx context.Context, an *schreiber, pflege *app.Schluesselwortweg,
	setID int64, rest []string,
) int {
	switch rest[0] {
	case "anlegen":
		return schluesselwortAnlegen(ctx, an, pflege, setID)

	case "setzen":
		return schluesselwortSetzen(ctx, an, pflege, setID, rest[1:])

	case "loeschen":
		return schluesselwortLoeschen(ctx, an, pflege, setID, rest[1:])

	// **Die Parameterhilfen** (#1342) — sie zeigen, was sich einfügen lässt.
	case "parameter":
		return schluesselwortParameter(ctx, an, pflege, setID, rest[1:])
	}

	an.klage("»%s« ist keine Handlung: anlegen, setzen, loeschen oder parameter.",
		rest[0])

	return CodeBedienung
}

// schluesselwortAnlegen hängt ein leeres ans Ende.
func schluesselwortAnlegen(
	ctx context.Context, an *schreiber, pflege *app.Schluesselwortweg, setID int64,
) int {
	neu, fehler := pflege.Anlegen(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	an.sag("%d\n", neu)

	return an.abschluss()
}

// schluesselwortSetzen schreibt ein Feld.
func schluesselwortSetzen(
	ctx context.Context, an *schreiber, pflege *app.Schluesselwortweg,
	setID int64, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("setzen braucht die Kennung, das Feld und den Wert.\n%s", hilfe)

		return CodeBedienung
	}

	wortID, gut := kennung(an, "die Schlüsselwortkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	if fehler := pflege.Setzen(ctx, setID, wortID, rest[1], rest[2]); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	an.sag("Das Feld %s steht.\n", rest[1])

	return an.abschluss()
}

// schluesselwortLoeschen nimmt eines weg.
func schluesselwortLoeschen(
	ctx context.Context, an *schreiber, pflege *app.Schluesselwortweg,
	setID int64, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("loeschen braucht die Schlüsselwortkennung.\n%s", hilfe)

		return CodeBedienung
	}

	wortID, gut := kennung(an, "die Schlüsselwortkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	if fehler := pflege.Loeschen(ctx, setID, wortID); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	an.sag("Das Schlüsselwort ist weg.\n")

	return an.abschluss()
}
