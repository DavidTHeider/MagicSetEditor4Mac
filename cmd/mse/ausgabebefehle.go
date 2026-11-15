// **Die Wachen der ausgebenden Befehle** — sichern, zeichnen, den Bogen legen,
// MTG JSON lesen und schreiben.
//
// **Eine eigene Datei**, seit `lauf.go` mit dem Druckbogen und MTG JSON an der
// Zeilengrenze aus §2 stand. Der Schnitt folgt der Frage: dort steht, **welcher**
// Befehl gemeint ist, hier, **ob sein Aufruf stimmt**.

package main

import (
	"context"
	"database/sql"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sichernbefehl prüft die Argumente von `sichern` und führt ihn aus.
func sichernbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("sichern braucht die Setkennung und ein Ziel.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return setSchreiben(ctx, db, an, setID, rest[1])
	})
}

// bogenbefehl legt die Karten eines Sets auf Papier (#1291).
//
// **Mit `--probe` rechnet er nur** (#2015): derselbe Aufruf, und es wird
// nichts geschrieben.
func bogenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if uebrig, probe := probeAus(rest); probe {
		return bogenprobebefehl(ctx, bestandsort, an, uebrig)
	}

	if len(rest) != 2 {
		an.klage("bogen braucht die Setkennung und ein Ziel.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return bogenLegen(ctx, db, an, setID, rest[1])
	})
}

// mtgjsonlesebefehl fährt eine MTG-JSON-Quelle ein (#1291).
func mtgjsonlesebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("mtgjson-lesen braucht den Pfad und das Spiel.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return mtgjsonLesen(ctx, db, an, rest[0], rest[1])
	})
}

// mtgjsonschreibbefehl schreibt ein Set als MTG JSON (#1291).
func mtgjsonschreibbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 3 {
		an.klage("mtgjson-schreiben braucht die Setkennung, das Spiel und ein Ziel.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return mtgjsonSchreiben(ctx, db, an, setID, rest[1], rest[2])
	})
}

// zeichnenbefehl prüft die Argumente von `zeichnen` und führt ihn aus.
func zeichnenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
	guete string,
) int {
	if len(rest) != 2 {
		an.klage("zeichnen braucht die Kartenkennung und ein Ziel.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return karteZeichnen(ctx, db, an, karteID, rest[1], guete)
	})
}

// paketeinbindebefehl bindet ein Paket von der Platte ein (#1291).
func paketeinbindebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("paket-installieren braucht die Quelle.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return paketInstallieren(ctx, db, an, rest[0])
	})
}

// paketanlegebefehl legt ein neues Paket aus einem Gerüst an (#1350).
//
// # Die Wortstellung hängt von der Art ab, und das ist Absicht
//
//	paket-anlegen mse-game            <name> [vollname]
//	paket-anlegen mse-include         <name> [vollname]
//	paket-anlegen mse-style           <name> <spiel> [vollname]
//	paket-anlegen mse-export-template <name> <spiel> [vollname]
//
// **Welche Art ein Spiel braucht, sagt `app.Anlegbar`** und nicht diese
// Fläche — hier stand `art == "mse-style"`, und die Exportvorlage (FMT-02,
// #1691) nahm ihr Spiel als Vollnamen.
//
// **Weil die Begriffe es tun: ein Spiel hat kein Spiel.** Hier stand
// `<art> <name> [voll] [spiel]` — beide letzten als wahlfrei —, und das dritte
// Wort war trotzdem **immer** der Vollname. Wer `mse paket-anlegen mse-style
// schlicht eigenes` schrieb, legte „eigenes" als Vollnamen ab und scheiterte
// am fehlenden Spiel. Befund 1 des Zweitblicks über Phase 4.10.
//
// **Zwei Zeilen, jede eindeutig**, sind ehrlicher als eine mit zwei wahlfreien
// Wörtern, von denen man das erste nicht auslassen kann.
func paketanlegebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paket-anlegen braucht Art und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	art, name, uebrige := rest[0], rest[1], rest[2:]

	vollname, spiel := "", ""

	if brauchtSpiel(art) {
		if len(uebrige) < 1 || len(uebrige) > 2 {
			an.klage("paket-anlegen %s braucht Namen und Spiel "+
				"und darf einen Vollnamen tragen.\n%s", art, hilfe)

			return CodeBedienung
		}

		spiel = uebrige[0]
		uebrige = uebrige[1:]
	}

	if len(uebrige) > 1 {
		an.klage("paket-anlegen: zu viele Wörter.\n%s", hilfe)

		return CodeBedienung
	}

	if len(uebrige) == 1 {
		vollname = uebrige[0]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return paketAnlegen(ctx, db, an, art, name, vollname, spiel)
	})
}

// paketentfernbefehl nimmt ein Paket wieder heraus (#1291).
func paketentfernbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("paket-entfernen braucht Art und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return paketEntfernen(ctx, db, an, rest[0], rest[1])
	})
}

// uebernahmebefehl holt die Vorlagen aus einem Verzeichnis (#1291).
func uebernahmebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("mse2-uebernehmen braucht das Verzeichnis.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return vorlagenUebernehmen(ctx, db, an, rest[0])
	})
}

// bildbefehl setzt ein Kartenbild aus einer Datei (#1291).
func bildbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 4 {
		an.klage("bild-setzen braucht Setkennung, Kartenkennung, Feld und Pfad.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[1])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return bildSetzen(ctx, db, an, setID, karteID, rest[2], rest[3])
	})
}

// vorlagenbefehl verteilt die drei Züge rund um Exportvorlagen (#1291).
//
// **Ein Befehl mit drei Zügen und nicht drei Befehle**, wie `stile` und
// `stiloption`: sie gehören zusammen — wer eine Vorlage ausführen will, sieht
// erst nach, welche es gibt.
//
// **`ausfuehren` nimmt `--probe`** (#2015): dieselbe Ausführung ohne das
// Schreiben.
func vorlagenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	rest, probe := probeAus(rest)

	if len(rest) == 0 {
		an.klage("vorlagen braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	switch {
	case probe:
		return vorlagenprobebefehl(ctx, bestandsort, an, setID, rest)

	case len(rest) == 1:
		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return vorlagenZeigen(ctx, db, an, setID)
		})

	case rest[1] == "ausfuehren" && len(rest) >= 4:
		wahl, brauchbar := wahlAus(an, rest[4:])
		if !brauchbar {
			return CodeBedienung
		}

		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return vorlageAusfuehren(ctx, db, an, setID, rest[2], rest[3], wahl)
		})

	case rest[1] == "option" && len(rest) == 5:
		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return vorlagenoptionSetzen(ctx, db, an, rest[2], rest[3], rest[4])
		})
	}

	an.klage("vorlagen kennt »ausfuehren« und »option«.\n%s", hilfe)

	return CodeBedienung
}

// brauchtSpiel sagt, ob eine Art beim Anlegen ein Spiel verlangt — aus
// `app.Anlegbar`, der einen Liste für jede Fläche.
func brauchtSpiel(art string) bool {
	return slices.ContainsFunc(app.Anlegbar(), func(eintrag app.Anlegbarkeit) bool {
		return eintrag.Art == art && eintrag.BrauchtSpiel
	})
}
