package main

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135. Dieselbe Naht wie Fenster, Dienst und Werkzeugserver
// (`app.ObjektmarkenUeber`); hier steht nur, wie die Argumente hereinkommen
// und die Antwort hinausgeht.

import (
	"context"
	"database/sql"
	"slices"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitObjektmarken öffnet den Bestand und reicht die Naht weiter.
func mitObjektmarken(ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Objektmarkennaht) error,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		if fehler := arbeit(ctx, app.ObjektmarkenUeber(db)); fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return an.abschluss()
	})
}

// objektmarkebefehl liest, setzt oder nimmt die Marken eines Objekts ab.
func objektmarkebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("objektmarke braucht eine Art und eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	art, kennung := rest[0], rest[1]

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		if len(rest) == 2 {
			marken, fehler := naht.ObjektmarkenAn(ctx, art, kennung)
			if fehler == nil {
				wortliste(an, "marken", marken)
			}

			return fehler
		}

		weg := slices.Contains(rest[3:], "--weg")
		tun, was := naht.ObjektmarkeSetzen, "gesetzt"

		if weg {
			tun, was = naht.ObjektmarkeEntfernen, "entfernt"
		}

		fehler := tun(ctx, art, kennung, rest[2])
		if fehler == nil {
			markenmeldung(an, map[string]any{"marke": domain.Markennormalform(rest[2]), was: true})
		}

		return fehler
	})
}

// objektmarkenbefehl nennt die Marken jedes Objekts einer Art.
func objektmarkenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("objektmarken braucht eine Art.\n%s", hilfe)

		return CodeBedienung
	}

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		je, fehler := naht.MarkenJeObjekt(ctx, rest[0])
		if fehler != nil {
			return fehler
		}

		if an.alsJSON {
			alsJSONSchreiben(an, map[string]any{"marken": je})

			return nil
		}

		kennungen := make([]string, 0, len(je))
		for kennung := range je {
			kennungen = append(kennungen, kennung)
		}

		sort.Strings(kennungen)

		for _, kennung := range kennungen {
			an.sag("%s\t%s\n", kennung, strings.Join(je[kennung], ", "))
		}

		return nil
	})
}

// markenfilterbefehl nennt die Objekte, die die Marken tragen.
func markenfilterbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	verknuepfung := string(domain.VerknuepfungUnd)
	if slices.Contains(rest, "--oder") {
		verknuepfung = string(domain.VerknuepfungOder)
	}

	rest = slices.DeleteFunc(slices.Clone(rest), func(wort string) bool { return wort == "--oder" })
	if len(rest) < 2 {
		an.klage("marken-filtern braucht eine Art und wenigstens eine Marke.\n%s", hilfe)

		return CodeBedienung
	}

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		kennungen, fehler := naht.MarkenFiltern(ctx, rest[0], rest[1:], verknuepfung)
		if fehler == nil {
			wortliste(an, "kennungen", kennungen)
		}

		return fehler
	})
}

// markevergabebefehl hängt eine Marke an mehrere Objekte.
func markevergabebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 3 {
		an.klage("marke-vergeben braucht eine Art, eine Marke und wenigstens eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		fehler := naht.MarkeVergeben(ctx, rest[0], rest[2:], rest[1])
		if fehler == nil {
			markenmeldung(an, map[string]any{"marke": domain.Markennormalform(rest[1]), "vergeben": len(rest) - 2})
		}

		return fehler
	})
}

// markenkatalogbefehl zählt die Marken des Profils.
func markenkatalogbefehl(ctx context.Context, bestandsort string, an *schreiber) int {
	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		katalog, fehler := naht.Markenkatalog(ctx)
		if fehler != nil {
			return fehler
		}

		if an.alsJSON {
			alsJSONSchreiben(an, map[string]any{"marken": katalog})

			return nil
		}

		for _, posten := range katalog {
			an.sag("%s\t%d\n", posten.Marke, posten.Anzahl)
		}

		return nil
	})
}

// markeumbenennbefehl gibt einer Marke einen neuen Namen.
func markeumbenennbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	zusammen := slices.Contains(rest, "--zusammenfuehren")
	rest = slices.DeleteFunc(slices.Clone(rest), func(wort string) bool { return wort == "--zusammenfuehren" })

	if len(rest) != 2 {
		an.klage("marke-umbenennen braucht den alten und den neuen Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		zahl, fehler := naht.MarkeUmbenennen(ctx, rest[0], rest[1], zusammen)
		if fehler == nil {
			markenmeldung(an, map[string]any{"marke": domain.Markennormalform(rest[1]), "anzahl": zahl})
		}

		return fehler
	})
}

// marketilgbefehl nimmt eine Marke von jedem Objekt des Profils.
func marketilgbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("marke-tilgen braucht eine Marke.\n%s", hilfe)

		return CodeBedienung
	}

	return mitObjektmarken(ctx, bestandsort, an, func(ctx context.Context, naht *app.Objektmarkennaht) error {
		zahl, fehler := naht.MarkeUeberallEntfernen(ctx, rest[0])
		if fehler == nil {
			markenmeldung(an, map[string]any{"marke": domain.Markennormalform(rest[0]), "anzahl": zahl})
		}

		return fehler
	})
}

// wortliste schreibt eine Liste — als JSON unter ihrem Namen, sonst je Zeile.
func wortliste(an *schreiber, name string, woerter []string) {
	if woerter == nil {
		woerter = []string{}
	}

	if an.alsJSON {
		alsJSONSchreiben(an, map[string]any{name: woerter})

		return
	}

	for _, wort := range woerter {
		an.sag("%s\n", wort)
	}
}

// markenmeldung schreibt, was geschah — als JSON, sonst die Marke.
func markenmeldung(an *schreiber, meldung map[string]any) {
	if an.alsJSON {
		alsJSONSchreiben(an, meldung)

		return
	}

	an.sag("%s\n", meldung["marke"])
}
