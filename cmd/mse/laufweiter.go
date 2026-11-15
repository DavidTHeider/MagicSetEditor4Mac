package main

// **Die zweite Hälfte der Befehlsverteilung.**
//
// # Warum sie eine eigene Datei ist
//
// **Weil `lauf.go` über die Zeilengrenze aus §2 ging.** Der Schnitt folgt
// keiner Frage, und das ist hier ehrlicher als eine erfundene: es ist **eine**
// Fallunterscheidung über alle Befehle, und sie in „Bestand" und „Betrieb" zu
// teilen hieße, dass jemand beim nächsten Befehl raten müsste, wohin er
// gehört.
//
// **Was der Schnitt kostet, steht im `default` der ersten Hälfte**: ein
// unbekannter Befehl geht einmal durch beide, bevor er als unbekannt gilt.
// Was er einbringt, ist die Zeilengrenze — und die gibt es, weil eine Datei
// mit vielen Zuständigkeiten unprüfbar wird.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fuehreWeiterAus setzt die Verteilung fort.
func fuehreWeiterAus(
	ctx context.Context, bestandsort string, an *schreiber,
	befehl string, rest []string, guete string,
) int {
	switch befehl {
	case "zuruecknehmen":
		return bewegungsbefehl(ctx, bestandsort, an, rest, true)

	case "wiederherstellen":
		return bewegungsbefehl(ctx, bestandsort, an, rest, false)

	case "einstellungen":
		return einstellungsbefehl(ctx, bestandsort, an, rest)

	// **Der Dienstkatalog** — Werkstück A1 von Phase 5.3 (#1440).
	case "dienstkatalog":
		return dienstkatalogbefehl(ctx, bestandsort, an, rest)
	case "grenzen":
		return grenzenbefehl(ctx, bestandsort, an, rest)

	// **Die eigene Peeridentität** — Werkstück N1 von Phase 5.4 (#1454).
	// Drei Befehle: fragen, anlegen, ablösen.
	case "peer":
		return peerbefehl(ctx, bestandsort, an, rest)
	case "peer-anlegen":
		return peeranlegenbefehl(ctx, bestandsort, an, rest)
	case "peer-rotieren":
		return peerrotierenbefehl(ctx, bestandsort, an, rest)

	// **Und die bekannten Gegenüber** — sechs Befehle, weil es sechs
	// Entscheidungen sind (N1, #1454).
	case "kontakte":
		return kontaktebefehl(ctx, bestandsort, an, rest)
	case "kontakt-aufnehmen":
		return kontaktaufnehmenbefehl(ctx, bestandsort, an, rest)
	case "kontakt-bestaetigen":
		return kontaktbestaetigenbefehl(ctx, bestandsort, an, rest)
	case "kontakt-sperren":
		return kontaktsperrenbefehl(ctx, bestandsort, an, rest)

	// **Einladen, prüfen, annehmen** (NET-01, #1705).
	case "einladung-erstellen":
		return einladungerstellenbefehl(ctx, bestandsort, an, rest)
	case "einladung-pruefen":
		return einladungpruefenbefehl(ctx, bestandsort, an, rest)
	case "einladung-annehmen":
		return einladungannehmenbefehl(ctx, bestandsort, an, rest)
	case "netzweg":
		return netzwegbefehl(ctx, bestandsort, an, rest)
	case "schnittstellen":
		return schnittstellenbefehl(ctx, an, rest)
	case "schalter":
		return schalterbefehl(ctx, bestandsort, an, rest)

	case "regelfreigabe":
		return regelfreigabebefehl(ctx, bestandsort, an, rest)
	case "nachbarn":
		return nachbarnbefehl(ctx, bestandsort, an, rest)
	case "kontakt-adresse":
		return kontaktadressebefehl(ctx, bestandsort, an, rest)
	case "kontakt-notiz":
		return kontaktnotizbefehl(ctx, bestandsort, an, rest)

	// **Die freigegebenen Sichten** (T3, #1468).
	//
	// **`kontakt-` und nicht `freigabe-`**: `freigaben` gibt es seit Phase
	// 4.12 und meint die Freigabe eines **Profils** an einen zweiten Nutzer
	// dieses Rechners. Zwei Befehle unter einem Namen wären nicht nur
	// verwirrend — der zweite wäre **toter Code**, weil der erste zuerst
	// greift.
	case "kontakt-freigaben":
		return sichtfreigabenbefehl(ctx, bestandsort, an, rest)
	case "kontakt-freigeben":
		return sichtfreigebenbefehl(ctx, bestandsort, an, rest)
	case "kontakt-widerrufen":
		return sichtfreigabewiderrufbefehl(ctx, bestandsort, an, rest)
	case "kontakt-sicht":
		return sichtvorschaubefehl(ctx, bestandsort, an, rest)
	case "sicht-anfragen":
		return sichtanfragebefehl(ctx, bestandsort, an, rest)
	case "tausch-abgleich":
		return tauschabgleichbefehl(ctx, bestandsort, an, rest)
	case "handelspost":
		return handelspostbefehl(ctx, bestandsort, an, rest)

	case "handel-festhalten":
		return handelFesthaltenbefehl(ctx, bestandsort, an, rest)

	case "handelszusagen":
		return handelszusagenbefehl(ctx, bestandsort, an, rest)

	case "horchen":
		return horchenbefehl(ctx, bestandsort, an, rest)

	case "fernhorchen":
		return fernhorchenbefehl(ctx, bestandsort, an, rest)

	case "handelszustellung":
		return handelszustellungbefehl(ctx, bestandsort, an, rest)

	case "handelspost-schicken":
		return handelsschickenbefehl(ctx, bestandsort, an, rest)

	case "handelspost-bestaetigen":
		return handelsbestaetigtbefehl(ctx, bestandsort, an, rest)

	case "handelspost-wiederaufnahme":
		return handelswiederaufnahmebefehl(ctx, bestandsort, an, rest)

	case "handel":
		return handelsverbuchungbefehl(ctx, bestandsort, an, rest)

	case "handelsjournal":
		return handelsjournalbefehl(ctx, bestandsort, an, rest)

	case "handelsbeleg":
		return handelsbelegbefehl(ctx, bestandsort, an, rest)

	case "markt-links":
		return marktlinksbefehl(ctx, bestandsort, an, rest)

	case "markt-oeffnen":
		return marktoeffnenbefehl(ctx, bestandsort, an, rest)

	case "markt-konto":
		return marktkontobefehl(ctx, bestandsort, an, rest)

	case "deck-liste":
		return deckListebefehl(ctx, bestandsort, an, rest)

	case "deck-liste-uebernehmen":
		return deckListeUebernehmenbefehl(ctx, bestandsort, an, rest)

	case "kontakt-entsperren":
		return kontaktentsperrenbefehl(ctx, bestandsort, an, rest)
	case "zweck-erlauben":
		return zweckerlaubenbefehl(ctx, bestandsort, an, rest)
	case "zweck-entziehen":
		return zweckentziehenbefehl(ctx, bestandsort, an, rest)
	case "gesundheit":
		return gesundheitsbefehl(ctx, bestandsort, an, rest)

	// **Ohne Bestand** (O9, #1433): „welche Fassung bist du" muss auch dann
	// eine Antwort haben, wenn der Bestand nicht aufgeht.
	case "meta":
		return metabefehl(ctx, an, rest)

	// **Die Verträglichkeit** (O6, #1430).
	case "vertraeglichkeit":
		return vertraeglichkeitsbefehl(ctx, bestandsort, an, rest)
	case "ruecksprung-arten":
		return ruecksprungartenbefehl(ctx, bestandsort, an, rest)
	case "ruecksprung-lage":
		return ruecksprunglagebefehl(ctx, bestandsort, an, rest)
	case "rueckkehren":
		return rueckkehrenbefehl(ctx, bestandsort, an, rest)
	case "pflege-vorschau":
		return pflegevorschaubefehl(ctx, bestandsort, an, rest)
	case "pflege":
		return pflegebefehl(ctx, bestandsort, an, rest)
	case "indexstand":
		return indexstandbefehl(ctx, bestandsort, an, rest)
	case "indexsuche":
		return indexsuchbefehl(ctx, bestandsort, an, rest)
	case "indexaufbau":
		return indexaufbaubefehl(ctx, bestandsort, an, rest)
	case "recherche":
		return recherchebefehl(ctx, bestandsort, an, rest)
	case "aehnlich":
		return aehnlichbefehl(ctx, bestandsort, an, rest)
	case "kontextpaket":
		return kontextpaketbefehl(ctx, bestandsort, an, rest)
	case "ruecksetzungsarten":
		return ruecksetzungsartenbefehl(ctx, bestandsort, an, rest)
	case "ruecksetzung-vorschau":
		return ruecksetzungsvorschaubefehl(ctx, bestandsort, an, rest)
	case "zuruecksetzen":
		return zuruecksetzenbefehl(ctx, bestandsort, an, rest)
	case "rueckkehr-quittieren":
		return rueckkehrquittierenbefehl(ctx, bestandsort, an, rest)

	// **Ohne Bestand** (#888): er liest ein Verzeichnis im Bündel.
	case "beispielsets":
		return beispielsetsbefehl(an, rest)

	case "oeffnen":
		if len(rest) != 1 {
			an.klage("oeffnen braucht genau einen Pfad.\n%s", hilfe)

			return CodeBedienung
		}

		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return setEinfahren(ctx, db, an, rest[0])
		})

	case "sichern":
		return sichernbefehl(ctx, bestandsort, an, rest)

	case "umschlag-schnueren":
		return umschlagschnuerenbefehl(ctx, bestandsort, an, rest)
	case "umschlag-schauen":
		return umschlagschauenbefehl(ctx, bestandsort, an, rest)
	case "umschlag-uebernehmen":
		return umschlaguebernehmenbefehl(ctx, bestandsort, an, rest)

	case "partie-anlegen":
		return partieanlegenbefehl(ctx, bestandsort, an, rest)
	case "partie-zeigen":
		return partiezeigenbefehl(ctx, bestandsort, an, rest)
	case "partien":
		return partienbefehl(ctx, bestandsort, an, rest)
	case "partie-verlauf":
		return partieverlaufbefehl(ctx, bestandsort, an, rest)
	case "partie-befehl":
		return partiebefehlbefehl(ctx, bestandsort, an, rest)
	case "partie-stand":
		return partiestandbefehl(ctx, bestandsort, an, rest)
	case "partie-loeschen":
		return partieloeschenbefehl(ctx, bestandsort, an, rest)
	case "partie-kopieren":
		return partiekopierenbefehl(ctx, bestandsort, an, rest)
	case "partie-ausfuhr":
		return partieausfuhrbefehl(ctx, bestandsort, an, rest)
	case "partie-einfuhr":
		return partieeinfuhrbefehl(ctx, bestandsort, an, rest)

	case "datei-ansehen":
		return dateiansehenbefehl(an, rest)
	case "bestand-sichern":
		return bestandSichernBefehl(ctx, bestandsort, an, rest)
	case "bestand-schauen":
		return bestandSchauenBefehl(ctx, bestandsort, an, rest)
	case "bestand-sicherungsstand":
		return sicherungsstandBefehl(ctx, bestandsort, an, rest)
	case "bestand-zurueckspielen":
		return bestandZurueckspielenBefehl(ctx, bestandsort, an, rest)
	case "nutzer-wechseln":
		return nutzerwechselbefehl(ctx, bestandsort, an, rest)
	case "kontext":
		return kontextbefehl(ctx, bestandsort, an, rest)

	case "bestand-verschluesselung":
		return bestandVerschluesselungBefehl(ctx, bestandsort, an, rest)

	case "zeichnen":
		return zeichnenbefehl(ctx, bestandsort, an, rest, guete)

	case "bogen":
		return bogenbefehl(ctx, bestandsort, an, rest)

	case "symbol-nachzeichnen":
		return symbolnachzeichnenbefehl(ctx, bestandsort, an, rest)

	case "mtgjson-lesen":
		return mtgjsonlesebefehl(ctx, bestandsort, an, rest)

	case "mtgjson-schreiben":
		return mtgjsonschreibbefehl(ctx, bestandsort, an, rest)

	case "vorlagen":
		return vorlagenbefehl(ctx, bestandsort, an, rest)

	case "bild-setzen":
		return bildbefehl(ctx, bestandsort, an, rest)

	case "mtgjson-spiele":
		for _, spiel := range app.MTGJSONSpiele() {
			an.sag("%s\n", spiel)
		}

		return an.abschluss()

	case "skript":
		return skriptbefehl(ctx, bestandsort, an, rest)
	case "skriptnamen":
		return namensbefehl(ctx, bestandsort, an, rest)

	case "packarten":
		return packartenbefehl(ctx, bestandsort, an, rest)

	case "pack":
		return packbefehl(ctx, bestandsort, an, rest)

	case "schluesselwoerter":
		return schluesselwoerterbefehl(ctx, bestandsort, an, rest)

	case "schluesselwort":
		return schluesselwortbefehl(ctx, bestandsort, an, rest)

	case "wortlisten":
		return wortlistenbefehl(ctx, bestandsort, an, rest)

	case "wortwahl":
		return wortwahlbefehl(ctx, bestandsort, an, rest)

	case "scryfall":
		return scryfallbefehl(ctx, bestandsort, an, rest)

	case "scryfall-karte":
		return scryfallkartebefehl(ctx, bestandsort, an, rest)

	case "scryfall-set":
		return scryfallsatzbefehl(ctx, bestandsort, an, rest)

	case "scryfall-bild":
		return scryfallbildbefehl(ctx, bestandsort, an, rest)

	// **Diese beiden brauchen keinen Bestand** (#888): sie lesen nichts aus
	// ihm. Ihn zu öffnen, nur um ihn nicht zu benutzen, machte sie von einem
	// Zustand abhängig, den sie nicht haben.
	default:
		an.klage("unbekannter Befehl %q.\n%s", befehl, hilfe)

		return CodeBedienung
	}
}
