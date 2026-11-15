// Der Composition Root der Kommandozeile.
//
// **Der dritte** — nach dem Fenster (`internal/app/app.go`) und dem MCP-Server
// (`cmd/mse-mcp/server.go`). Was ein weiterer kostet, hat der Zweitblick über
// Phase S gezeigt: der Server migrierte ohne Sicherung, weil sein Root `nil`
// übergab. Dieser reicht dieselbe Vorbereitung durch wie die anderen beiden.

package main

import (
	"context"
	"database/sql"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Die Rückgabewerte, die ein Skript auswerten kann.
//
// **Drei und nicht zwei.** „Es hat nicht geklappt" ist für ein Skript zu
// wenig: ein Tippfehler im Aufruf will anders behandelt werden als ein Set,
// das es nicht gibt. Ein Skript, das beides gleich behandelt, wiederholt den
// Aufruf, der nie funktionieren wird.
const (
	// CodeGut heißt: getan.
	CodeGut = 0

	// CodeFachlich heißt: der Aufruf war richtig, die Antwort ist nein — das
	// Set gibt es nicht, der Bestand ist belegt, die Platte ist voll.
	CodeFachlich = 1

	// CodeBedienung heißt: der Aufruf selbst stimmt nicht — unbekannter
	// Befehl, fehlendes Argument.
	CodeBedienung = 2

	// CodeAbgebrochen heißt: jemand hat ⌃C gedrückt oder SIGTERM geschickt.
	//
	// **128 plus die Signalnummer** — die Übereinkunft der Shells, und SIGINT
	// ist 2. Ein Skript, das `until mse …; do` schleift, soll nach einem
	// Abbruch **nicht** weiterprobieren; ohne eigenen Wert sähe der Abbruch
	// aus wie „die Antwort ist nein" und die Schleife liefe weiter.
	CodeAbgebrochen = 130
)

// lauf führt einen Aufruf aus und gibt seinen Rückgabewert.
//
// **Mit Strömen, Argumenten und dem Bestandsort als Parameter**, obwohl `main`
// immer dieselben übergibt: ohne das bliebe der Composition Root ungeprüft —
// dort entstand in Phase 5 der Fehler, den kein Test sah (#130).
//
// **Der Ort ist der wichtigste davon**: ein Test, der ihn nicht setzen kann,
// führe gegen den echten Bestand der Nutzenden.
// **`ablage-einfuegen` liest von der Standardeingabe** (#1168) — ohne Fenster
// ist sie die Zwischenablage, und sie steht hier aus demselben Grund wie die
// Ströme darüber: ein Test, der sie nicht setzen kann, läse, was auf dem
// Terminal des Prüfenden ansteht.
func lauf(
	ctx context.Context, argumente []string, ein io.Reader,
	aus, fehlerstrom io.Writer, bestandsort string,
) int {
	gelesen := zerlege(argumente)

	an := &schreiber{
		ziel: aus, fehler: fehlerstrom, alsJSON: gelesen.alsJSON, eingabe: ein,
		erwartet: gelesen.erwartet,
	}

	// **`--profil` reist im Kontext**, wie das Profil selbst — und
	// `mitBestand` sieht ihn dann schon ausgestattet, statt den Zeiger zu
	// lesen. Ein zweiter Weg durch jede Befehlssignatur wäre derselbe Wert an
	// zwei Stellen.
	ctx = mitWunschprofil(ctx, gelesen.profil)

	// **Und hier entsteht die Vorgangskennung** — Werkstück O9 von Phase 5.2
	// (#1433). Vor der Verteilung und nicht in `mitBestand`: die Befehle ohne
	// Bestand — `beispielsets`, `scryfall-regeln`, die Werkstatt — sollen
	// dieselbe Auskunft geben wie die anderen.
	ctx = app.MitNeuemVorgang(ctx)
	an.vorgang = app.VorgangAus(ctx)

	// **Und wer an einem Terminal sitzt, sieht den Stand** (FMT-03).
	if melder, da := terminalmelder(fehlerstrom); da {
		ctx = app.MitFortschritt(ctx, melder)
	}

	// **Und vor dem Befehl das Tor der Schalter** — OPS-10 (#1908).
	code, abgewiesen := schaltertor(ctx, gelesen.befehl, gelesen.rest, bestandsort, an)
	if !abgewiesen {
		code = abgebrochenOder(ctx,
			fuehreAus(ctx, gelesen.befehl, gelesen.rest, bestandsort, an, hilfe, gelesen.guete))
	}

	// **Der Griff ins Protokoll, und nur wenn etwas schiefging.** Bei
	// `CodeBedienung` gibt es nichts nachzusehen — der Satz darüber sagt schon
	// alles —, und auf dem geglückten Weg wäre die Zeile Rauschen. Wer sie
	// bekommt, sucht damit in `mse.log` (`docs/ADMINISTRATION.md`).
	if code == CodeFachlich {
		an.klage("Vorgang: %s\n", app.VorgangAus(ctx))
	}

	return code
}

// abgebrochenOder gibt den Abbruchwert, wenn der Kontext abgebrochen ist.
//
// **Nach dem Lauf und nicht davor**: ein Befehl, der schon fertig war, als das
// Signal kam, hat getan, was er sollte — und ein Skript, dem wir dann 130
// meldeten, räumte etwas auf, das steht.
//
// **Und nur, wenn er ohnehin nicht gelang.** Wer abbricht, nachdem geschrieben
// wurde, bekommt den Erfolg gemeldet; geschrieben ist geschrieben (§7, Regel 3).
func abgebrochenOder(ctx context.Context, code int) int {
	if code != CodeGut && ctx.Err() != nil {
		return CodeAbgebrochen
	}

	return code
}

// fuehreAus wählt den Befehl.
func fuehreAus(
	ctx context.Context, befehl string, rest []string, bestandsort string,
	an *schreiber, hilfe, guete string,
) int {
	// **Die Sammlung zuerst** (#1165). Sie steht in einer eigenen Verteilung,
	// weil diese hier an der Zeilengrenze aus §2 stand — und weil der Schnitt
	// der Frage folgt: dort das **Set**, hier der **Besitz**.
	if code, getroffen := sammlungsbefehle(ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	// **Und die Paketwerkstatt ohne Bestand** (#1164): eine Paketdatei liegt
	// auf der Platte und nicht in der Datenbank — sie braucht keinen der
	// Werte, die diese Verteilung sonst herumreicht.
	if code, getroffen := werkstattbefehle(ctx, befehl, rest, an); getroffen {
		return code
	}

	// **Und die vier ohne Bestand** — dieselbe Aussage: sie fragen einen
	// fremden Dienst und fassen den Bestand nicht an.
	if code, getroffen := fremdbefehle(
		ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	// **Und die Befehle über Menschen** — Nutzer, Profile, Konten, zweite
	// Faktoren und Zugriffstoken. Herausgelöst am 2026-09-05 (§2).
	if code, getroffen := leutebefehle(
		ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	// **Und die drei über Geld** (O12, #1439) — eigene Datei, §2.
	if code, getroffen := kostenbefehle(
		ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	// **Und Plan, Ersetzen und Zurück** (DOM-01, #1696) — eigene Datei, §2.
	if code, getroffen := paketplanbefehle(
		ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	// **Und die Stände der bezogenen Bäume** (G04, #2251) — eigene Datei, §2.
	if code, getroffen := paketbezugbefehle(
		ctx, befehl, rest, bestandsort, an); getroffen {
		return code
	}

	switch befehl {
	case "":
		an.klage("kein Befehl genannt.\n%s", hilfe)

		return CodeBedienung

	case "hilfe":
		an.sag("%s", hilfe)

		return an.abschluss()

	case "sets":
		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return setsAuflisten(ctx, db, an)
		})

	case "set-neu":
		return setneubefehl(ctx, bestandsort, an, rest)

	case "set-loeschen":
		return setloeschbefehl(ctx, bestandsort, an, rest)

	case "karten":
		if len(rest) != 1 {
			an.klage("karten braucht genau eine Setkennung.\n%s", hilfe)

			return CodeBedienung
		}

		setID, gut := kennung(an, "die Setkennung", rest[0])
		if !gut {
			return CodeBedienung
		}

		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return kartenAuflisten(ctx, db, an, setID)
		})

	case "feld":
		return feldbefehl(ctx, bestandsort, an, rest)

	case "felder":
		return felderbefehl(ctx, bestandsort, an, rest)

	case "felder-setzen":
		return felderSetzenBefehl(ctx, bestandsort, an, rest)

	case "felder-vorschau":
		return felderVorschauBefehl(ctx, bestandsort, an, rest)

	case "symbolpalette":
		return symbolpaletteBefehl(ctx, bestandsort, an, rest)

	// **Das Symbol eines Sets geht über den Bestand** (#1815): es liegt als
	// Datei neben dem Set, aber wem das Set gehört, weiß nur der Bestand.
	case "symbol":
		return symbolLesenBefehl(ctx, bestandsort, an, rest)

	case "symbol-setzen":
		return symbolSetzenBefehl(ctx, bestandsort, an, rest)

	case "karte":
		return kartenbefehl(ctx, bestandsort, an, rest)

	case "suche":
		return suchbefehl(ctx, bestandsort, an, rest)

	case "setangaben":
		return setangabenbefehl(ctx, bestandsort, an, rest)

	case "vorbehalte":
		return vorbehaltebefehl(ctx, bestandsort, an, rest)

	case "probleme":
		return problemebefehl(ctx, bestandsort, an, rest)

	case "stile":
		return stilebefehl(ctx, bestandsort, an, rest)

	case "stiloption":
		return stiloptionbefehl(ctx, bestandsort, an, rest)

	case "zusatzfeld":
		return zusatzfeldbefehl(ctx, bestandsort, an, rest)

	case "pakete":
		return paketebefehl(ctx, bestandsort, an, rest)

	case "paket-beziehen":
		return paketbezugbefehl(ctx, bestandsort, an, rest)

	case "paket-installieren":
		return paketeinbindebefehl(ctx, bestandsort, an, rest)

	case "paket-anlegen":
		return paketanlegebefehl(ctx, bestandsort, an, rest)

	case "paket-entfernen":
		return paketentfernbefehl(ctx, bestandsort, an, rest)

	case "paket-kopieren":
		return paketkopierbefehl(ctx, bestandsort, an, rest)

	case "paket-weitergeben":
		return paketweitergabebefehl(ctx, bestandsort, an, rest)

	case "installer-einlesen":
		return installereinlesebefehl(ctx, bestandsort, an, rest)

	case "mse2-orte":
		if len(rest) != 0 {
			an.klage("mse2-orte nimmt keine Argumente.\n%s", hilfe)

			return CodeBedienung
		}

		return mse2Orte(an)

	case "mse2-uebernehmen":
		return uebernahmebefehl(ctx, bestandsort, an, rest)

	case "statistik":
		return statistikbefehl(ctx, bestandsort, an, rest)

	case "verlauf":
		return verlaufbefehl(ctx, bestandsort, an, rest)

	default:
		// **Die zweite Hälfte steht in `laufweiter.go`** — dieselbe
		// Fallunterscheidung, nur in einer zweiten Datei: zusammen ging sie
		// über die Zeilengrenze aus §2. Ein `default`, der weiterreicht,
		// hält die Reihenfolge der Fälle **und** die eine Stelle, an der ein
		// unbekannter Befehl endet (unten).
		return fuehreWeiterAus(ctx, bestandsort, an, befehl, rest, guete)
	}
}

// ohneBestandsort meldet, dass sich der Ort des Bestands nicht bestimmen ließ
// — **vor** dem Lauf, aber mit `--json` trotzdem mit der Form (Nachlauf über
// Phase 1 von 0.9.0, Befund 3). Bis dahin schrieb `main` den Satz selbst.
func ohneBestandsort(argumente []string, aus, fehlerstrom io.Writer, fehler error) int {
	an := &schreiber{ziel: aus, fehler: fehlerstrom, alsJSON: zerlege(argumente).alsJSON}
	an.vorgang = app.VorgangAus(app.MitNeuemVorgang(context.Background()))

	an.klageUeber(fehler)

	return CodeFachlich
}
