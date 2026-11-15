package service

// **Die Karten der Ausfuhr tragen ihre gerechneten Werte** —
// #2006.
//
// # Was MSE2 tut, nachgelesen
//
// `export_set` wertet das Vorlagenskript im Kontext des **Sets** aus —
//
//	Context& ctx = set->getContext();
//	ctx.setVariable(_("cards"), to_script(&cards));
//	ScriptValueP result = exp->script.invoke(ctx);
//
// `src/gui/html_export_window.cpp`, Zeilen 55–90, Fassung f9d9356 vom
// 2021-07-06 —, und die Karten darin tragen ihre berechneten Werte, weil
// `SetScriptManager::updateAll` jedes Feld jeder Karte beim Laden rechnet
// (`src/script/script_manager.cpp`, Zeilen 265–300): erst alle Felder jeder
// Karte, dann `updateAllDependend` für alles, was von der Kartenliste abhängt.
// `type` und `card number` von `magic.mse-game` tragen `save value: false` und
// ein `script:` — in der Datei steht von beiden nichts, und eine Vorlage, die
// nach `card.card_number` ordnet, ordnete nach lauter leeren Schlüsseln.
//
// # Was hier geschieht
//
// **Zwei Runden über denselben Weg wie die Auszählung** ([auszaehlumgebung]):
// das `init script:` des Spiels einmal, je Karte die Vorgabewerte, die
// Vorgabeskripte und die Wertskripte.
//
//  1. Die erste Runde rechnet ohne Kartenliste. Was nicht von ihr abhängt, ist
//     danach fertig; `card_number` ist es nicht, denn es zählt die Karte im
//     **Set** — nach Werten, die erst die Runde berechnet hat.
//  2. Die zweite Runde rechnet mit der Liste der Ergebnisse der ersten als
//     `set` und einem leeren Speicher der Ordnungen ([Umgebung.OrdnungenVergessen],
//     MSEs `clearOrderCache`).
//
// **adapted**: MSE ordnet die zweite Runde nach einem Abhängigkeitsgraphen
// (`dependent_scripts_cards`) und rechnet nur, was von der Liste abhängt; dieses
// Programm führt keinen (siehe `kartenlesezugriff.go`) und rechnet beide Male
// alles. Das kostet das Doppelte einer Runde; der Weg, es zu sparen, wäre ein
// solcher Graph — oder die Beschränkung der zweiten Runde auf die Felder, die
// sich gegenüber der ersten ändern.
//
// # Im Stil der Karte
//
// MSE rechnet jede Karte im Kontext ihres Stilbogens (`getContext(card)`), und
// ein Feld, das den Stil befragt (`shape` rechnet `stylesheet; card_shape()`),
// bekommt dessen Wert. Bis VK-01b
// (#2182)
// stand hier ein leerer Stil; die Kontexte stehen in `ausfuhrkontext.go`.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// ausfuhrkarten rechnet die Felder der Karten für eine Ausfuhr.
//
// **Je Karte die Felder, die sie nach den Skripten trägt**, als Text — und
// nichts, wo es nichts zu rechnen gibt: ohne erklärte Kartenfelder bleibt die
// Liste, wie [kartenliste] sie baut.
//
// **Ein Abbruch gibt nichts zurück**, kein halbes Ergebnis; die Vorlage selbst
// hört ihn beim ersten Knoten.
func ausfuhrkarten(
	ctx context.Context, vorlage msepaket.Exportvorlage, karten []*domain.CardRow,
	setwerte []domain.Value, stile Ausfuhrstile,
) ([][]msescript.Eintrag, []domain.Befund) {
	if len(vorlage.Kartenfelder) == 0 || len(karten) == 0 {
		return nil, nil
	}

	modell := Vorlage{
		Kartenfelder: vorlage.Kartenfelder, Setfelder: vorlage.Setfelder,
		Setangaben: setwerte, Spielskript: vorlage.Spielskript,
		Symbolfassungen: vorlage.Symbolfassungen,
	}

	kontexte := neueAusfuhrkontexte(modell, stile, karten)

	// **Zwei Runden, und die zweite liest die Ergebnisse der ersten als
	// `set`.** Vor jeder wird der Speicher der Ordnungen geleert — MSEs
	// `clearOrderCache` am Anfang einer Runde.
	var (
		liste   []msescript.Eintrag
		letzte  []gerechnetekarte
		befunde []domain.Befund
	)

	for durchgang := range 2 {
		kontexte.ordnungenVergessen()

		runde, hinweise, ok := ausfuhrrunde(ctx, kontexte, karten, liste, durchgang == 1)
		if !ok {
			return nil, nil
		}

		letzte, befunde = runde, hinweise

		liste = make([]msescript.Eintrag, 0, len(runde))
		for _, karte := range runde {
			liste = append(liste, msescript.Eintrag{Name: karte.name, Wert: karte.abbildung})
		}
	}

	gerechnet := make([][]msescript.Eintrag, 0, len(letzte))
	for _, karte := range letzte {
		gerechnet = append(gerechnet, karte.felder)
	}

	return gerechnet, entdoppelt(befunde)
}

// gerechnetekarte ist eine Karte nach einer Runde.
type gerechnetekarte struct {
	// name ist ihre Stelle in der Liste, als Name eines Eintrags.
	name string

	// abbildung ist `card`, wie ein Skript sie sieht — mit Kennung, Farben und
	// Bildern. Die zweite Runde steht damit als `set.cards` da. **Nur nach der
	// ersten Runde gefüllt.**
	abbildung msescript.Wert

	// felder sind die Felder als Text, bevor Farben und Bilder umgewandelt sind.
	// **Nur nach der letzten Runde gefüllt.**
	felder []msescript.Eintrag
}

// ausfuhrrunde rechnet jede Karte einmal.
//
// **Jede Runde behält nur, was die nächste Stufe braucht**: die erste die
// Abbildungen für `set`, die letzte die Felder für die Vorlage. Beides zu
// halten kostete bei den 3.788 Karten von `Black.mse-set` rund 0,4 GB mehr
// (Höchststand 1,41 statt 0,97 GB, gemessen am 2026-10-03; ohne Rechnung der
// Felder 0,39 GB).
//
// **`false` heißt: abgebrochen.** Ein Ergebnis, dem Karten fehlen, wäre eines,
// das für vollständig gehalten wird.
func ausfuhrrunde(
	ctx context.Context, kontexte *ausfuhrkontexte, karten []*domain.CardRow,
	liste []msescript.Eintrag, letzte bool,
) ([]gerechnetekarte, []domain.Befund, bool) {
	runde := make([]gerechnetekarte, 0, len(karten))

	var befunde []domain.Befund

	for stelle, karte := range karten {
		if ctx.Err() != nil {
			return nil, nil, false
		}

		kontext := kontexte.fuer(ctx, karte)

		umgebung, felder, hinweise := auszaehlumgebung(
			ctx, kontext.spiel, kontext.modell, karte, kontext.vermerk, liste)
		befunde = append(befunde, hinweise...)

		gerechnet := gerechnetekarte{name: strconv.Itoa(stelle)}

		if letzte {
			gerechnet.felder = make([]msescript.Eintrag, 0, len(felder))
			for _, feld := range felder {
				gerechnet.felder = append(gerechnet.felder, msescript.Eintrag{
					Name: feld.Name, Wert: msescript.Text(feld.Wert.AlsText()),
				})
			}
		} else {
			gerechnet.abbildung, _ = umgebung.Suche("card")
		}

		runde = append(runde, gerechnet)
	}

	return runde, befunde, true
}

// mitGerechneten setzt die gerechneten Werte in die Karten der Liste.
//
// **Überschrieben wird, was die Karte führt** — und das ist jedes Feld: der
// Weg durch [auszaehlumgebung] und [kartenliste] legt dieselben Namen an, die
// gespeicherten, die erklärten und `notes`. Die Reihenfolge der Liste bleibt,
// wie [kartenliste] sie gebaut hat.
func mitGerechneten(liste msescript.Wert, gerechnet [][]msescript.Eintrag) msescript.Wert {
	if gerechnet == nil {
		return liste
	}

	for i := range liste.Eintraege {
		karte := liste.Eintraege[i].Wert
		karte.Eintraege = append([]msescript.Eintrag(nil), karte.Eintraege...)

		stelle := make(map[string]int, len(karte.Eintraege))
		for j, eintrag := range karte.Eintraege {
			stelle[eintrag.Name] = j
		}

		for _, eintrag := range gerechnet[i] {
			if j, da := stelle[eintrag.Name]; da {
				karte.Eintraege[j].Wert = eintrag.Wert
			}
		}

		liste.Eintraege[i].Wert = karte
	}

	return liste
}
