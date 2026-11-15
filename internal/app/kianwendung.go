package app

// **Wo eine KI-Anwendung ohne Fenster herkommt** — der Composition Root für
// MCP, REST und die Kommandozeile.
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet.** Nebenan steht, *was* der
// Katalog ist — Modelle, Belege, Herkunft, und wie sie als DTO aussehen.
// Hier steht, *woraus eine Anwendung besteht*, die ihn beantworten kann: die
// Laufzeiten, das Kostenjournal, die Recherche, die Bilderkennung, der
// Schreibweg.
//
// **Und weil es ein Composition Root ist** — §2 nennt sie ausdrücklich als
// benannte Ausnahme von „`internal/app` entscheidet nichts". Eine Ausnahme
// gehört sichtbar an eine Stelle und nicht zwischen die DTOs.
//
// Der Schnitt folgt der Frage und nicht der Zeilenzahl (§2) — die 400 waren
// der Anlass.

import (
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIAnwendungUeber baut eine Anwendung, die den Katalog kennt.
//
// **Mit den Einstellungen** — ohne sie gäbe es keine Ortsregel, und der
// Katalog sagte an einer Fläche etwas anderes als an der nächsten (§16).
func KIAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app
	anwendung.db = db

	anwendung.MitEinstellungen(einstellungsverwalterUeber(db))

	// **Und die Laufzeit dazu** (A7, #1446). Ohne diese Zeile stünde der
	// Katalog an den Flächen ohne Fenster leer da, während er am Fenster
	// gefüllt ist — dieselbe Frage, zwei Antworten (§16).
	anwendung.MitKIAnbietern(anwendung.ollamalaufzeit, anwendung.cloudnaehte())

	// **Und die Recherche** (A13.2a, #1452): eine Einschätzung stützt sich
	// auf Belege, und die kommen aus dem Kontextdienst von A16.
	//
	// **Ohne diese Zeile gäbe es die Einschätzung an MCP, REST und der
	// Kommandozeile nicht** — sie scheiterte mit „es ist kein Bestand
	// offen", während sie am Fenster liefe. Dieselbe Gestalt wie
	// #1539:
	// gebaut und nicht angeschlossen. Gefunden hat es der Flächenlauf, nicht
	// die Abdeckung.
	//
	// **Es kostet nichts, sie hier zu bauen**: [App.rechercheVerdrahten]
	// setzt Schnittstellen zusammen und öffnet nichts; der Einbetter entsteht
	// je Ruf und fragt erst, wenn jemand sucht.
	anwendung.rechercheVerdrahten()

	// **Und die Bilderkennung** (A13.3, #1452) — aus demselben Grund: ohne
	// diese Zeile gäbe es sie an MCP, REST und der Kommandozeile nicht,
	// während sie am Fenster liefe.
	anwendung.bilderkennungVerdrahten()

	// **Und das Kostenjournal** (A8, #1447): ein Modell, das anderswo
	// rechnet, kostet Geld, und ohne Journal geht kein solcher Auftrag
	// hinaus ([kibudget]).
	anwendung.MitKosten(KostenwacheUeber(
		func() *sql.DB { return db }, anwendung.autorisierung, anwendung.uhr))

	// **Mit Kartenleser, -schreiber und Stilwahl** — ein Vorschlag liest die
	// Zielkarte, holt ihr Schema aus dem Paket und schreibt sie, wenn jemand
	// es verlangt.
	//
	// **Über [Schreibnaht] und nicht über den rohen Schreiber**: wer den
	// nähme, legte einen rohen `text`-Wert ab, den kein Paket liest — der
	// Befund aus #710,
	// und für eine KI gilt er genauso wie für die Kommandozeile.
	//
	// **Der Urheber ist der Server**, weil dieser Bauweg die Flächen **ohne
	// Fenster** bedient — MCP, REST und Kommandozeile. Er sagt, **wer**
	// geschrieben hat; **woher der Wert kam**, steht daneben als Herkunft
	// `ki` ([kiherkunft]), und das ist die Angabe, auf die es hier ankommt.
	//
	// **Ein Fehler dabei bleibt stehen** — er wird gemerkt und beim ersten
	// Ruf gemeldet, der ihn braucht. Dieser Bauweg hat keine Rückgabe für
	// einen, und ihn zu verschlucken hieße, den Vorschlag später mit
	// „kein Bestand offen" abzuweisen statt mit dem, was wirklich fehlt.
	pakete, paketfehler := Paketverzeichnis()
	bilder, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		anwendung.kinahtfehler = fehler

		return anwendung
	}

	stile := stilwaehlerMit(db, pakete)
	naht := schreibnahtMit(db, domain.UrheberServer, stile, nachzugswerk(db, pakete, bilder))

	anwendung.MitEditor(naht, storage.NeuerSetLeser(db), nil, nil)
	anwendung.MitStilwahl(stile)

	return anwendung
}

// MitKIModellen setzt den Katalog — bis A7 den ersten Anbieter bringt, ist er
// leer, und ein Lauf braucht ihn gefüllt.
func (a *App) MitKIModellen(modelle ...domain.KIModell) *App {
	a.kimodelle = modelle

	return a
}
