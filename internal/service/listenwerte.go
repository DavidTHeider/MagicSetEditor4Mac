package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Berechnete Zellen der Kartenliste** —
// #1238.
//
// # Warum es das gibt
//
// `zeilenAus` füllte die Zellen ausschließlich aus `CardRow.Values`. In
// `magic.mse-game` sind zahlreiche sichtbare Listenfelder aber **nicht
// gespeichert, sondern gerechnet** — `full name` (Zeile 337), `cmc` (564),
// `type` (961), `pt` (2306), `card number` (2958). Sie tragen
// `save value: false` und entstehen erst aus ihrem `script:`.
//
// Eine Spalte ließ sich damit wählen und blieb leer, und eine Sortierung
// danach ordnete nach lauter leeren Schlüsseln.
//
// # Was hier **nicht** geschieht
//
// **Der Planbau.** `umgebungAus` rechnet dieselben Werte längst — mit Stil,
// Schriften, Bildern und Vorgabeskripten. Das kostet je Karte Millisekunden;
// bei den 3.788 Karten von `Black.mse-set` wäre die Liste damit unbenutzbar.
//
// Hier steht die **magere** Fassung: die Werte der Karte, ihre Abbildung als
// `card`, und das Wertskript des Feldes. Kein Stil, keine Bilder, keine
// Schriften.
//
// # Der Preis, und er gehört gesagt
//
// **Das `init script` des Spiels läuft einmal je Liste** ([spielumgebung],
// #1238) — was es anlegt (`card_full_name`, `cmc`), steht jeder Zelle
// bereit. **Der Stil der Karte nicht**: ein Feld, das ihn befragt (`shape`
// rechnet `stylesheet; card_shape()`), zeigt hier den Wert des Spiels. Die
// Ausfuhr rechnet seit VK-01b im Stil ([ausfuhrkontexte]); Liste und
// Schnellfilter tun es nicht, denn ein Kontext je Stil kostete je Anfrage die
// Initskripte jedes Stils im Set.
//
// Hier stand bis WART-01 (#2172), das `init script` stehe in dieser Umgebung
// nicht — das galt vor #1238.

// sortierschluessel rechnet die `sort script:` einer Zeile.
//
//	inline String getSortKey() const {
//	  return fieldP->sort_script ? sort_value : toString();
//	}
//
// `src/data/field.hpp`, Zeilen 234–237, Fassung f9d9356 vom 2021-07-06 —
// #1232.
//
// **Nur für die Spalte, nach der geordnet wird.** Alle fünfzehn zu rechnen
// hieße, für jede Karte vierzehn Ergebnisse wegzuwerfen.
func sortierschluessel(
	ctx context.Context, spiel *msescript.Umgebung, karte *domain.CardRow,
	felder []domain.Field, nach string, werte map[string]string,
) map[string]string {
	if nach == "" {
		return nil
	}

	for i := range felder {
		if felder[i].Name != nach || strings.TrimSpace(felder[i].Sortierskript) == "" {
			continue
		}

		eigene := listenumgebung(spiel, karte, felder).Darueber()
		eigene.Setze("value", msescript.Text(werte[nach]))

		wert, err := msescript.AuswertenQuelle(ctx, felder[i].Sortierskript, eigene)
		if err != nil || wert.IstFehler() {
			// **Ein gescheitertes Sortierskript ordnet nach dem Wert** — das
			// ist MSEs Verhalten ohne Skript und die harmlose Richtung.
			return nil
		}

		return map[string]string{nach: wert.AlsText()}
	}

	return nil
}

// gerechneteZellen füllt die Zellen, die kein gespeicherter Wert füllt.
//
// **Nur sichtbare Spalten.** Das ist die Sparsamkeitszusage: `magic.mse-game`
// führt 142 Wertskripte, und eine Liste zeigt ein halbes Dutzend Spalten.
//
// **Und nur leere Zellen.** Wer etwas eingetippt hat, soll es sehen —
// dieselbe Regel wie beim Vorgabewert (`kartenvorgaben.go`).
func gerechneteZellen(
	ctx context.Context, spiel *msescript.Umgebung, karte *domain.CardRow,
	felder []domain.Field, spalten []domain.Listenspalte, werte map[string]string,
) {
	zuRechnen := offeneSkripte(felder, spalten, werte)
	if len(zuRechnen) == 0 {
		return
	}

	umgebung := listenumgebung(spiel, karte, felder)

	for _, feld := range zuRechnen {
		// **Der Abbruch wirkt je Feld.** Wer während des Ladens das Set
		// wechselt, wartet sonst darauf, dass die alte Liste fertig gerechnet
		// wird — dieselbe Überlegung wie in `mitVorgaben`.
		if ctx.Err() != nil {
			return
		}

		// **`value` ist der bisherige Wert dieses Feldes** — MSE-Skripte kennen
		// keine Stellungsargumente, sie lesen die Kontextvariable
		// (`src/data/field/text.cpp`, Zeile 179, Fassung f9d9356 vom 2021-07-06). `casting cost` heißt `script: mana_filter(value)`.
		eigene := umgebung.Darueber()
		eigene.Setze("value", msescript.Text(werte[feld.Name]))

		wert, err := msescript.AuswertenQuelle(ctx, feld.Skript, eigene)
		if err != nil || wert.IstFehler() {
			// **Ein gescheitertes Skript lässt die Zelle leer und hält die
			// Liste nicht auf** (§8). Es zu melden hieße, für ein
			// Vorlagenpaket mit 142 Skripten eine Befundliste je Karte zu
			// bauen — und die Liste ist eine Übersicht, keine Diagnose.
			continue
		}

		werte[feld.Name] = domain.OhneAuszeichnung(wert.AlsText())
	}
}

// offeneSkripte sammelt die Felder, deren Zelle leer ist und die ein
// Wertskript tragen.
func offeneSkripte(
	felder []domain.Field, spalten []domain.Listenspalte, werte map[string]string,
) []domain.Field {
	gefragt := make(map[string]bool, len(spalten))

	for _, spalte := range spalten {
		if spalte.Sichtbar {
			gefragt[spalte.Feld] = true
		}
	}

	offen := make([]domain.Field, 0, len(gefragt))

	for i := range felder {
		if !gefragt[felder[i].Name] || strings.TrimSpace(felder[i].Skript) == "" {
			continue
		}

		if werte[felder[i].Name] == "" {
			offen = append(offen, felder[i])
		}
	}

	return offen
}

// listenumgebung baut die magere Umgebung einer Karte.
//
// **`card` und die Felder einzeln**, denn beide Formen kommen in Wertskripten
// vor: `magic.mse-game` schreibt `card.name` ebenso wie das bloße `name`.
func listenumgebung(
	spiel *msescript.Umgebung, karte *domain.CardRow, felder []domain.Field,
) *msescript.Umgebung {
	eintraege := make([]msescript.Eintrag, 0, len(karte.Values))

	for _, wert := range karte.Values {
		eintraege = append(eintraege, msescript.Eintrag{
			Name: umgebungsname(wert.Field),
			Wert: msescript.Text(wert.Text),
		})
	}

	// **Auch die Felder ohne Wert stehen da** — mit dem leeren Text. Ein
	// Skript, das `card.name_2` liest, scheitert sonst am unbekannten Namen,
	// obwohl das Feld dem Spiel bekannt ist und nur an dieser Karte leer.
	da := make(map[string]bool, len(eintraege))
	for i := range eintraege {
		da[eintraege[i].Name] = true
	}

	for i := range felder {
		if name := umgebungsname(felder[i].Name); !da[name] {
			eintraege = append(eintraege, msescript.Eintrag{
				Name: name, Wert: msescript.Text(""),
			})
			da[name] = true
		}
	}

	// **Und `notes` steht an jeder Karte, auch an einer ohne.**
	//
	// Die Regel und ihre Fundstelle stehen bei [mitNotizen];
	// [kartenumgebungVorbereiten] tut dasselbe. **Hier fehlte sie bis zum
	// 2026-08-31**, und aufgefallen ist es an den Auszählungsdimensionen:
	// drei der fünfzig von `magic.mse-game` scheiterten mit „die Abbildung
	// führt kein Glied `notes`" (#1338).
	//
	// **Es traf die Kartenliste genauso** — eine gerechnete Spalte, die
	// `card.notes` liest, scheiterte still und blieb leer.
	eintraege = mitNotizen(eintraege, karte)

	// **Die Feldnamen stehen in `card` und **nicht** bloß daneben.**
	//
	// Der erste Anlauf setzte jeden Feldnamen zusätzlich als eigenen Namen —
	// und `magic.mse-game` führt ein Feld `cmc` **und** eine Funktion `cmc()`.
	// Der Name verdeckte die Funktion, und das Skript `cmc(card.casting_cost)`
	// meldete „hier wird Zeichenkette gerufen".
	//
	// MSE macht es genauso: die Werte einer Karte sind `card.x`, der eigene
	// Wert ist `value` (`src/data/field/text.cpp`, Zeile 179, Fassung f9d9356 vom 2021-07-06).
	umgebung := spiel.Darueber()
	// **Die Kennung macht `card` zu einem Gegenstand** (#1366). MSE bindet
	// sie als `ScriptObject<CardP>` (`src/script/script_manager.cpp`, Zeilen
	// 30–60, Fassung `f9d9356` vom 2021-07-06), und der Vergleich läuft über
	// `COMPARE_AS_POINTER`: zwei Karten sind gleich, wenn es **dieselbe**
	// Karte ist — nicht, wenn sie dieselben Felder tragen.
	umgebung.Setze("card", msescript.Wert{
		Art: msescript.ArtAbbildung, Eintraege: eintraege,
		Kennung: Skriptkartenkennung(karte),
	})

	return umgebung
}

// spielumgebung fährt das `init script:` des Spiels — **einmal je Liste**.
//
// # Warum einmal und nicht je Karte
//
// `magic.mse-game` legt darin die Funktionen an, die seine Feldskripte rufen:
// `card_full_name()`, `cmc(…)`, `card_type()`. Bei 3.788 Karten wäre ein Lauf
// je Karte das Vielfache — und das Ergebnis wäre jedes Mal dasselbe.
//
// # Ohne Skript die bloße Bibliothek
//
// Ein Set ohne Spielpaket bekommt die Standardbibliothek, und die Feldskripte
// rechnen, was sich damit rechnen lässt. Ein Fehler im Spielskript hält die
// Liste **nicht** auf (§8): sie ist eine Übersicht, keine Diagnose.
// **Die Schlüsselwortnaht steht vor dem Skript** — Befund 5 des Zweitblicks
// über Phase 4.11. `text_filter` friert den Funktionswert ein, den
// `expand_keywords` beim Bauen der Kette hatte; die Begründung steht in
// `schluesselwortnaht.go`.
func spielumgebung(
	ctx context.Context, skript string, setfelder, kartenfelder []domain.Field,
	woerter Wortliste, vermerk *Schluesselwortverwendung,
) *msescript.Umgebung {
	umgebung := msescript.Bibliothek().Darueber()

	belegeSchluesselwortnaht(umgebung, woerter, vermerk)

	if strings.TrimSpace(skript) == "" {
		return umgebung
	}

	// **`set` und `card` stehen da, bevor das Skript sie liest.**
	//
	// Das ist dieselbe Regel wie in `planumgebung.go`, und sie ist dort teuer
	// bezahlt: das `init script` von `magic.mse-game` greift in **Zeile
	// 10583** auf `set` zu. Ohne den Namen bricht es dort ab — und alles
	// danach fehlt, also auch `card_full_name` und `cmc`.
	//
	// Gemessen am 2026-08-28: ohne diese zwei Zeilen meldet der Lauf
	// `der Name "set" ist hier nicht bekannt`, und beide Feldskripte scheitern
	// danach am unbekannten Namen ihrer eigenen Funktion.
	//
	// **Und `set` führt seine Felder, nicht nichts.** Zeile 10583 liest
	// `set.custom_super_types` — ein fehlendes **Glied** bricht MSE-Script ab,
	// und anders als ein fehlender Name lässt es sich nicht mit `or else`
	// fangen. `magic.mse-game` führt 89 `set field:`-Blöcke; leer genügen sie,
	// denn gelesen wird hier nur die Gestalt.
	//
	// **Und `card` führt seine Felder aus demselben Grund**, seit dem
	// 2026-08-31 (#1338).
	//
	// Hier stand „**`card` bleibt leer**: was die Karte trägt, setzt
	// [listenumgebung] eine Ebene darüber, je Karte." Für die **Werte** ist
	// das richtig; für die **Gestalt** war es falsch, und es hat einen Namen
	// gekostet.
	//
	// `magic.mse-game` legt `land_to_color` als Block mit `@`-Bindung an:
	//
	//	land_to_color := {
	//	  …
	//	}@(rules:card.rule_text)
	//
	// Zeile 13768 des zusammengesetzten Skripts. Eine `@`-Bindung wird beim
	// **Anlegen** ausgewertet — ein fehlendes Glied bricht ab, und der Name
	// entsteht nie. Anders als bei einem gescheiterten Satz nimmt sie den
	// Namen mit, zu dem sie gehört.
	//
	// **Gemessen am 2026-08-31**: sechs von 3.761 Karten aus `White.mse-set`
	// bekamen für ihre Rückseite `Colorless` statt `Land`, weil
	// `card_color(…)` ohne `land_to_color` in seinen Ersatzwert fiel. Mit
	// den erklärten Feldern fallen die Vorbehalte des Spielskripts von drei
	// auf zwei.
	umgebung.Setze("set", leereAbbildung(setfelder))
	umgebung.Setze("card", leereAbbildung(kartenfelder))

	// **`VorbelegenMitVorbehalten` und nicht `AuswertenQuelle`.**
	//
	// Jenes bricht an der **ersten** Stelle ab, die nicht aufgeht; dieses
	// macht weiter und sammelt Vorbehalte. Bei einem Spielskript von 583.549
	// Zeichen ist das der Unterschied zwischen „alle Funktionen stehen da"
	// und „keine". Es ist derselbe Weg, den `planumgebung.go` seit #240 geht
	// — und diese Stelle hat ihn beim ersten Anlauf **nicht** genommen, mit
	// genau der Folge: `card_full_name` blieb unbekannt.
	//
	// **Die Vorbehalte werden hier nicht gemeldet.** Die Liste ist eine
	// Übersicht, keine Diagnose; wer sie sehen will, öffnet die Karte.
	//nolint:errcheck // ein gescheitertes Spielskript lässt die Zellen leer, siehe darüber
	_, _ = msescript.VorbelegenMitVorbehalten(ctx, skript, umgebung)

	return umgebung
}

// leereAbbildung gibt eine Abbildung mit allen Feldnamen und leerem Text.
func leereAbbildung(felder []domain.Field) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(felder))

	for i := range felder {
		eintraege = append(eintraege, msescript.Eintrag{
			Name: umgebungsname(felder[i].Name), Wert: msescript.Text(""),
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}
