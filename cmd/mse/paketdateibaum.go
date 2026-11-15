// **Der Dateibaum eines Pakets an der Kommandozeile** — FMT-02.AK-01
// (#1691).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort — und **Einfügen von der
// Platte gibt es hier**, anders als über REST und MCP: wer an dieser
// Kommandozeile sitzt, liest die Datei ohnehin mit seinen eigenen Rechten.
package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// paketdateianlegenbefehl legt eine neue Vorlage an — leer oder mit dem
// Inhalt einer lokalen Textdatei.
func paketdateianlegenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paketdatei-anlegen braucht Paket und Datei.\n%s", hilfe)

		return CodeBedienung
	}

	inhalt := ""

	if len(rest) > 2 {
		gelesen, lesefehler := os.ReadFile(rest[2]) // #nosec G304,G703 -- der Aufrufer sitzt an dieser Kommandozeile
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		inhalt = string(gelesen)
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	stand, fehler := naht.Anlegen(ctx, rest[0], rest[1], inhalt)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"stand": stand})
	}

	an.sag("%s\n", stand)

	return an.abschluss()
}

// paketdateieinfuegenbefehl legt eine Datei von der Platte ins Paket — ein
// Bild, eine Schrift. Ohne Ziel heißt sie wie die Quelle, oben im Paket.
func paketdateieinfuegenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paketdatei-einfuegen braucht Paket und Quelle.\n%s", hilfe)

		return CodeBedienung
	}

	ziel := filepath.Base(rest[1])
	if len(rest) > 2 {
		ziel = rest[2]
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if fehler := naht.Einfuegen(ctx, rest[0], rest[1], ziel); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return meldeDatei(an, "datei", ziel)
}

// paketdateiumbenennenbefehl gibt einer Datei einen neuen Namen. Der Stand
// ist bei einer Vorlage Pflicht und bei Beiwerk leer — das entscheidet der
// Dienst, nicht diese Fläche.
func paketdateiumbenennenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 3 {
		an.klage("paketdatei-umbenennen braucht Paket, Datei und neuen Namen.\n%s", hilfe)

		return CodeBedienung
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if fehler := naht.Umbenennen(ctx, rest[0], rest[1], rest[2], optional(rest, 3)); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return meldeDatei(an, "datei", rest[2])
}

// paketdateientfernenbefehl nimmt eine Datei aus dem Paket — in die Sicherung.
func paketdateientfernenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paketdatei-entfernen braucht Paket und Datei.\n%s", hilfe)

		return CodeBedienung
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if fehler := naht.Entfernen(ctx, rest[0], rest[1], optional(rest, 2)); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return meldeDatei(an, "entfernt", rest[1])
}

// optional gibt ein Argument, das fehlen darf.
func optional(rest []string, stelle int) string {
	if len(rest) > stelle {
		return rest[stelle]
	}

	return ""
}

// meldeDatei sagt, welche Datei es betraf — als Zeile oder als JSON.
func meldeDatei(an *schreiber, schluessel, datei string) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{schluessel: datei})
	}

	an.sag("%s\n", datei)

	return an.abschluss()
}
