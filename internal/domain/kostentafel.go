package domain

import "slices"

// **Was ein Ruf kostet, steht an einer Stelle** — Folgebefund F1 des
// Lastlaufs QA-08
// (#2003),
// Marke `new`: MSE2 hat weder Dienst noch Werkzeugserver.
//
// # Warum eine Tafel und nicht zwei Regeln
//
// **Bis zum 2026-10-02 buchte jede Fläche nach ihrer eigenen Regel**: REST
// nach einer kurzen Liste von Wegen, MCP nach Wörtern im Namen eines
// Werkzeugs. Dieselbe Fähigkeit kostete deshalb je Fläche etwas anderes —
// eine Karte zeichnen an beiden 2 statt der 5, die [KostenRendern] sagt, ein
// Bogen über REST 2 und über MCP 8, und eine Vorschau der Einstellungen über
// MCP 5, weil ihr Name „vorschau" enthält. Das ist §16: dieselbe Fähigkeit,
// verschiedene Antworten.
//
// **Jetzt steht je teurer Fähigkeit eine Zeile hier**, mit ihren Wegen an
// jeder Fläche — so geschrieben wie in `docs/FAEHIGKEITEN.md`, ein REST-Weg
// als Methode und Muster mit `{}` ([RESTWeg]). Der Dienst und der
// Werkzeugserver fragen [KostenklasseFuer], und `grenzen` nennt die Tafel an
// jeder Fläche. `internal/werkzeuge/kostentafel_test.go` hält jeden Weg gegen
// den Katalog.
//
// # Was teuer ist
//
// **Was ein ganzes Set, eine Sammlung oder ein Paket herein- oder
// hinausschreibt, viele Exemplare auf einmal anfasst, ein Bild verarbeitet
// oder eine Karte rastert.** Wo die Arbeit einen Platz im Haushalt belegt
// ([Ressourcenklasse]), ist ihre Klasse dieselbe. **Eine Datei mit einem
// Gegenstand ist es nicht** — ein Turnier, ein Symbol, eine Übersetzung, die
// Einstellungen: sie kosten, was ihre Art kostet, auch wenn ihr Werkzeug
// „einlesen" oder „ausfuhr" heißt.
//
// **Was beide Flächen bisher gleich buchten, ist nicht neu verhandelt** — mit
// einer Ausnahme, die der Befund verlangt: eine Karte zeichnen kostet jetzt,
// was Rendern kostet. Ob ein Set öffnen oder ein Paket beziehen teurer sein
// sollte als ein Schreiben, ist eine Frage über Preise und nicht über
// Parität; sie steht offen in `docs/LASTMESSUNG.md`.
//
// # Und was billig ist, obwohl es POST heißt
//
// **Die Art eines Wegs liest der Dienst an seiner Methode** — und eine Vorschau,
// die einen Rumpf mitnimmt, ist ein `POST`. Über MCP ist dieselbe Vorschau ein
// lesendes Werkzeug. Einundzwanzig Paare kosteten deshalb über REST 2 und über
// MCP 1 (#2020). Sie stehen jetzt mit der Klasse Lesen in `kostenlesen.go`,
// **in derselben Tafel**: die Methode entscheidet nur noch, wo die Tafel
// schweigt. Die Erlaubnis eines Tokens folgt weiter der Methode, und das ist
// Absicht ([0163](../../docs/entscheidungen/0163-zeichnen-belegt-einen-platz-und-lesen-kostet-lesen.md)).
//
// # Und ein Abbruch kostet nichts
//
// **Er ist die eine Anfrage, die Last nimmt.** Gebucht wie ein Schreiben,
// bekam er bei leerem Guthaben ein 429, und der Auftrag lief zu Ende —
// gemessen im Lastlauf. [KostenAbbruch] kostet null, und was null kostet,
// wird nie abgewiesen.

// KostenAbbruch ist die Klasse eines Abbruchs — sie kostet nichts.
const KostenAbbruch Kostenklasse = "abbruch"

// Kostenzeile ist eine Fähigkeit, die nicht kostet, was ihre Art an dieser
// Fläche kostet: teurer als ein Lesen oder Schreiben, oder billiger, als ihre
// Methode vermuten ließe.
type Kostenzeile struct {
	// Faehigkeit ist ihr Name aus `docs/FAEHIGKEITEN.md`.
	Faehigkeit string
	Klasse     Kostenklasse

	// Wege sind die Wege dieser Klasse je Fläche — eine Fähigkeit mit
	// billigen Wegen daneben nennt nur die teuren.
	Wege map[Flaeche][]string
}

// zeile baut eine Zeile mit je einem Weg an REST, MCP und der
// Kommandozeile — leer heißt: keiner an dieser Fläche.
func zeile(faehigkeit string, klasse Kostenklasse, rest, mcp, cli string) Kostenzeile {
	wege := map[Flaeche][]string{}

	for flaeche, weg := range map[Flaeche]string{FlaecheREST: rest, FlaecheMCP: mcp, FlaecheCLI: cli} {
		if weg != "" {
			wege[flaeche] = []string{weg}
		}
	}

	return Kostenzeile{Faehigkeit: faehigkeit, Klasse: klasse, Wege: wege}
}

// teureWege sind die Fähigkeiten, die mehr kosten als ihre Art, in der
// Reihenfolge der Klassen.
var teureWege = []Kostenzeile{
	zeile("Kartenbilder schreiben", KostenRendern,
		"POST /karten/{}/bild", "schreiben.karte.zeichnen", "mse zeichnen"),

	zeile("Einen Druckbogen legen", KostenAusfuhr,
		"POST /sets/{}/bogen", "schreiben.set.bogen", "mse bogen"),
	zeile("MTG JSON schreiben", KostenAusfuhr,
		"POST /sets/{}/mtgjson", "schreiben.mtgjson.schreiben", "mse mtgjson-schreiben"),
	zeile("Die Sammlung ausgeben", KostenAusfuhr,
		"GET /sammlung/ausgeben", "lesen.sammlung.ausgeben", "mse sammlung-ausgeben"),

	zeile("MTG JSON lesen", KostenEinfuhr,
		"POST /mtgjson", "schreiben.mtgjson.lesen", "mse mtgjson-lesen"),
	zeile("Ein Paket als `.mse-installer` weitergeben und einlesen", KostenEinfuhr,
		"POST /pakete/eingelesen", "schreiben.paket.einlesen", "mse installer-einlesen"),
	zeile("Eine CSV-Sammlung zurücklesen", KostenEinfuhr,
		"POST /sammlung/einlesen", "schreiben.sammlung.einlesen", "mse sammlung-einlesen"),
	zeile("Eine Kartenliste einlesen", KostenEinfuhr,
		"POST /sammlung/liste/einlesen", "schreiben.liste.einlesen", "mse liste-einlesen"),

	zeile("Viele Exemplare auf einmal ändern", KostenMasse,
		"POST /sammlung/masse", "schreiben.sammlung.masse", "mse masse-aendern"),
	zeile("Viele Exemplare auf einmal löschen", KostenMasse,
		"POST /sammlung/masse/loeschen", "schreiben.sammlung.masseloeschen", "mse masse-loeschen"),
	zeile("Preise für die ganze Sammlung", KostenMasse,
		"POST /sammlung/masse/preise", "schreiben.sammlung.preisschub", "mse preise-alle"),
	// **Lesend und trotzdem teuer** (PERF02, #2352): die Prüfung zeichnet
	// jede Karte eines Sets.
	zeile("Die Karten eines Sets auf Probleme prüfen", KostenMasse,
		"POST /sets/{}/probleme", "lesen.set.probleme.pruefen", "mse probleme"),

	zeile("Ein Kartenbild einsetzen", KostenBild,
		"PUT /sets/{}/karten/{}/felder/{}/bild", "schreiben.karte.bild", "mse bild-setzen"),
	zeile("Ein Kartenbild von Scryfall holen", KostenBild,
		"POST /sets/{}/karten/{}/scryfall/bild", "schreiben.scryfall.bild", "mse scryfall-bild"),

	// **Die Abbrüche gibt es nur am Dienst**: über MCP ist der Abbruch eine
	// Benachrichtigung und kein Werkzeug, auf der Kommandozeile ⌃C.
	zeile("Hintergrundaufgaben — was gerade geholt wird, mit Stand, Fortschritt und Dauer",
		KostenAbbruch, "DELETE /auftraege/{}", "", ""),
	zeile("Einen laufenden Indexaufbau abbrechen", KostenAbbruch, "DELETE /index/aufbau", "", ""),
}

// kostentafel ist die ganze Tafel: erst, was mehr kostet als seine Art, dann
// die lesenden Wege, die `POST` heißen (`kostenlesen.go`).
var kostentafel = slices.Concat(teureWege, lesendePostwege)

// kostenwege ordnet je Fläche jedem Weg der Tafel seine Klasse zu.
var kostenwege = func() map[Flaeche]map[string]Kostenklasse {
	je := map[Flaeche]map[string]Kostenklasse{}

	for _, zeile := range kostentafel {
		for flaeche, liste := range zeile.Wege {
			if je[flaeche] == nil {
				je[flaeche] = map[string]Kostenklasse{}
			}

			for _, weg := range liste {
				je[flaeche][weg] = zeile.Klasse
			}
		}
	}

	return je
}()

// Kostentafel gibt die Zeilen — eine Abschrift, die niemand verändert.
func Kostentafel() []Kostenzeile {
	abschrift := make([]Kostenzeile, 0, len(kostentafel))

	for _, zeile := range kostentafel {
		wege := make(map[Flaeche][]string, len(zeile.Wege))
		for flaeche, liste := range zeile.Wege {
			wege[flaeche] = append([]string(nil), liste...)
		}

		abschrift = append(abschrift, Kostenzeile{Faehigkeit: zeile.Faehigkeit, Klasse: zeile.Klasse, Wege: wege})
	}

	return abschrift
}

// Kostenauskunft ist eine Zeile der Tafel, wie jede Fläche sie in `grenzen`
// nennt — an allen vier dieselbe, und deshalb hier und nicht viermal.
type Kostenauskunft struct {
	Faehigkeit string   `json:"faehigkeit"`
	Klasse     string   `json:"klasse"`
	Kosten     int      `json:"kosten"`
	REST       []string `json:"rest"`
	MCP        []string `json:"mcp"`
	CLI        []string `json:"cli"`
}

// KostentafelAlsAuskunft gibt die Tafel als Auskunft — **leere Listen statt
// keiner**: eine Fläche ohne Weg sagt `[]`, nicht `null` (§3).
func KostentafelAlsAuskunft() []Kostenauskunft {
	auskunft := make([]Kostenauskunft, 0, len(kostentafel))

	for _, zeile := range kostentafel {
		auskunft = append(auskunft, Kostenauskunft{
			Faehigkeit: zeile.Faehigkeit, Klasse: string(zeile.Klasse), Kosten: zeile.Klasse.Kosten(),
			REST: append([]string{}, zeile.Wege[FlaecheREST]...),
			MCP:  append([]string{}, zeile.Wege[FlaecheMCP]...),
			CLI:  append([]string{}, zeile.Wege[FlaecheCLI]...),
		})
	}

	return auskunft
}

// KostenklasseFuer sagt, was ein Weg an einer Fläche kostet: die Klasse
// seiner Zeile, sonst die seiner Art — Lesen oder Schreiben.
//
// **Nur der Weg entscheidet, nie ein Wort darin.** Ein REST-Weg kommt als
// Methode und Muster ([RESTWeg]), ein Werkzeug mit seinem Namen, ein Befehl,
// wie er in der Wegetafel steht.
func KostenklasseFuer(flaeche Flaeche, weg string, schreibend bool) Kostenklasse {
	return KostenklasseFuerSchreibweisen(flaeche, schreibend, weg)
}

// KostenklasseFuerSchreibweisen fragt die Tafel mit mehreren Schreibweisen
// **desselben** Wegs — **die erste, die sie kennt, entscheidet**; kennt sie
// keine, entscheidet die Art.
//
// **Wofür** (#2020): der Dienst trägt die Rechnungen des Symboleditors als ein
// Muster ein, `POST /symbol/{rechnung}`, während `FAEHIGKEITEN.md` — und damit
// die Tafel — jede mit ihrem Wert nennt, `POST /symbol/grundform`. Der Dienst
// fragt deshalb mit dem Muster **und** mit dem Pfad, den die Anfrage trägt.
func KostenklasseFuerSchreibweisen(flaeche Flaeche, schreibend bool, wege ...string) Kostenklasse {
	for _, weg := range wege {
		if klasse, steht := kostenwege[flaeche][weg]; steht {
			return klasse
		}
	}

	if schreibend {
		return KostenSchreiben
	}

	return KostenLesen
}
