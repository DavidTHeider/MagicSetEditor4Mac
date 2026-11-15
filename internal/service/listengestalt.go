package service

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wie aus Feldern, Wahl und Karten eine Liste wird** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil `kartenliste.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt einer Frage: dort steht, **woher** die drei Quellen
// kommen, hier, **wie** sie zusammenkommen — reine Funktionen, einzeln
// prüfbar.

// spaltenAus legt die Spalten aus Vorschlag und Wahl zusammen.
//
// **Die Wahl schlägt den Vorschlag, und ein Verbot schlägt beide.**
//
//	if (cs.visible && f->card_list_allow) { … }
//
// `src/gui/control/card_list.cpp`, Zeile 218, Fassung f9d9356 vom
// 2021-07-06.
func spaltenAus(felder []domain.Field, gewaehlt []domain.Spaltenwahl) []domain.Listenspalte {
	stellen := make(map[string]domain.Spaltenwahl, len(gewaehlt))
	for _, eine := range gewaehlt {
		stellen[eine.Feld] = eine
	}

	// **Zwei Schlüssel je Feld**: seine Stelle und sein Name. Der Name
	// entscheidet, wenn zwei Felder dieselbe Stelle tragen — sonst hinge die
	// Reihenfolge an der Laune der Sortierung.
	type mitStelle struct {
		spalte domain.Listenspalte
		stelle int
	}

	geordnet := make([]mitStelle, 0, len(felder))

	for _, feld := range felder {
		if !feld.Listenerlaubt {
			continue
		}

		spalte := domain.Listenspalte{
			Feld:        feld.Name,
			Aufschrift:  feld.Listenname,
			Sichtbar:    feld.Listensichtbar,
			Ausrichtung: feld.Listenausrichtung,
		}

		eine, gewaehltesFeld := stellen[feld.Name]
		if gewaehltesFeld {
			spalte.Sichtbar = eine.Sichtbar
		}

		stelle := eine.Stelle
		if !gewaehltesFeld {
			stelle = feld.Listenstelle
		}

		// **Ein Feld ohne Stelle steht hinten** und nicht vorn: die Null
		// bedeutet „ungesagt", nicht „ganz vorne".
		//
		// **Vor der Verschiebung darunter** und nicht danach: sonst wäre
		// `0 + hinterDerWahl` kleiner als jede echte Stelle, und ein Feld
		// ohne Angabe stünde vor allen anderen ungewählten.
		if stelle == 0 {
			stelle = ohneStelle
		}

		// **Ein Feld ohne Eintrag in der Wahl behält den Vorschlag** — und
		// steht hinter allem Gewählten, damit eine neue Spalte des Spiels
		// nicht mitten hineinrutscht.
		if !gewaehltesFeld && len(gewaehlt) > 0 {
			stelle += hinterDerWahl
		}

		geordnet = append(geordnet, mitStelle{spalte: spalte, stelle: stelle})
	}

	slices.SortStableFunc(geordnet, func(links, rechts mitStelle) int {
		if links.stelle != rechts.stelle {
			return links.stelle - rechts.stelle
		}

		return domain.Vergleiche(links.spalte.Feld, rechts.spalte.Feld)
	})

	spalten := make([]domain.Listenspalte, 0, len(geordnet))
	for _, eine := range geordnet {
		spalten = append(spalten, eine.spalte)
	}

	return spalten
}

// Zwei Stellen jenseits dessen, was ein Paket vergibt — `magic.mse-game`
// zählt bis 9999.
const (
	// hinterDerWahl schiebt ein Feld hinter alles Gewählte, sobald überhaupt
	// gewählt wurde.
	hinterDerWahl = 1_000_000

	// ohneStelle steht ganz hinten.
	ohneStelle = 2_000_000
)

// zeilenAus macht aus den Karten die Zeilen.
func zeilenAus(
	ctx context.Context, karten []*domain.CardRow,
	felder []domain.Field, spalten []domain.Listenspalte,
	spielskript string, setfelder []domain.Field, ordnung domain.Kartenordnung,
) []domain.Listenzeile {
	// **Die Umgebung des Spiels steht einmal da**, nicht je Karte.
	spiel := spielumgebung(ctx, spielskript, setfelder, felder,
		Wortliste{}, &Schluesselwortverwendung{})

	zeilen := make([]domain.Listenzeile, 0, len(karten))

	for _, karte := range karten {
		werte := make(map[string]string, len(spalten))

		// **Jede angebotene Spalte steht da, auch die leere.** Ein Feld, das
		// die Karte nicht führt, ist eine leere Zelle und kein Fehler — und
		// eine fehlende Zelle kostete jeden Leser einen Zweig.
		for _, spalte := range spalten {
			werte[spalte.Feld] = ""
		}

		for _, wert := range karte.Values {
			if _, gefragt := werte[wert.Field]; gefragt {
				werte[wert.Field] = domain.OhneAuszeichnung(wert.Text)
			}
		}

		// **Und was kein gespeicherter Wert füllt, wird gerechnet** (#1238).
		gerechneteZellen(ctx, spiel, karte, felder, spalten, werte)

		zeilen = append(zeilen, domain.Listenzeile{
			KarteID: karte.ID, Nummer: karte.Number, Werte: werte,
			Schluessel: sortierschluessel(ctx, spiel, karte, felder, ordnung.Feld, werte),
		})
	}

	return zeilen
}

// wirksameOrdnung streicht eine Sortierung nach einem Feld, das es nicht gibt.
//
// **Und sagt es damit**: ein Pfeil in einer Spaltenüberschrift, die es nicht
// gibt, wäre eine Antwort auf eine Frage, die niemand mehr stellt.
func wirksameOrdnung(
	ordnung domain.Kartenordnung, spalten []domain.Listenspalte,
) domain.Kartenordnung {
	if ordnung.Feld == "" {
		return ordnung
	}

	for _, spalte := range spalten {
		if spalte.Feld == ordnung.Feld {
			return ordnung
		}
	}

	return domain.Kartenordnung{Aufsteigend: ordnung.Aufsteigend}
}

// ordne sortiert die Zeilen — **an Ort und Stelle**, weil sie gerade entstanden
// sind und niemandem sonst gehören.
//
// **Ohne Feld bleibt die Reihenfolge des Sets.** Das ist die Reihenfolge, in
// der die Karten in der `.mse-set` stehen, und die, die beim Speichern wieder
// herauskommt.
func ordne(zeilen []domain.Listenzeile, ordnung domain.Kartenordnung) {
	if ordnung.Feld == "" {
		return
	}

	slices.SortStableFunc(zeilen, func(links, rechts domain.Listenzeile) int {
		// **Der Sortierschlüssel schlägt den Wert** — MSEs `getSortKey()`
		// (#1232). Wo kein `sort script:` steht, ist er leer und es gilt der
		// Text.
		vergleich := domain.Vergleiche(
			ordnungswert(links, ordnung.Feld), ordnungswert(rechts, ordnung.Feld))

		if ordnung.Aufsteigend {
			return vergleich
		}

		return -vergleich
	})
}

// waehlbar sagt, ob ein Feldname zu einem Feld gehört, das in die Liste darf.
func waehlbar(felder []domain.Field, name string) bool {
	for _, feld := range felder {
		if feld.Name == name {
			return feld.Listenerlaubt
		}
	}

	return false
}

// ordnungswert gibt den Sortierschlüssel einer Zeile — sonst ihren Wert.
func ordnungswert(zeile domain.Listenzeile, feld string) string {
	if schluessel, hatEinen := zeile.Schluessel[feld]; hatEinen {
		return schluessel
	}

	return zeile.Werte[feld]
}
