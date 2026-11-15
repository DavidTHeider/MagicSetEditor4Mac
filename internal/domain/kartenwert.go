package domain

// **Stärke, Widerstandskraft, Loyalität und Verteidigung** — Werkstück M2 der
// Regelmaschine
// (#1604).
//
// # Warum ein eigener Typ und nicht ein `int`
//
// **Weil es drei Antworten gibt und nicht zwei.** CR 208.3 (Fassung
// 2026-08-19): eine Nichtkreatur hat **keine** Stärke — nicht null. CR 208.2:
// manche Karten tragen statt einer Zahl einen **Stern**, und was er bedeutet,
// sagt eine Fähigkeit (CR 604.3), die es hier nicht gibt (`M8`). Ein `int`
// könnte das eine nicht vom anderen unterscheiden, und ein `*int` sagte
// „fehlt" für beides.

import (
	"strconv"
	"strings"
)

// Kartenwert ist eine Zahl auf einer Karte, die auch fehlen oder ein Stern
// sein darf.
type Kartenwert struct {
	// Steht sagt, ob überhaupt etwas dasteht — CR 208.3.
	Steht bool

	// Stern sagt, dass statt einer Zahl ein `*` dasteht — CR 208.2.
	//
	// **Die Zahl daneben ist dann die gedruckte Null**, und was wirklich
	// gilt, sagt eine Fähigkeit der Karte (CR 604.3, `M8`). Ein `1+*` trägt
	// die Eins und den Stern.
	Stern bool

	Zahl int
}

// KartenwertAus liest einen Wert, wie er auf der Karte steht.
//
// **Leer heißt: es steht keiner da** (CR 208.3) — und das ist der Normalfall,
// denn die meisten Karten sind keine Kreaturen.
func KartenwertAus(roh string) Kartenwert {
	text := strings.TrimSpace(roh)
	if text == "" {
		return Kartenwert{}
	}

	wert := Kartenwert{Steht: true}

	ohneStern := strings.ReplaceAll(text, "*", "")
	if ohneStern != text {
		wert.Stern = true
	}

	// **Das Pluszeichen fällt weg, das Minus bleibt**: `1+*` ist eine Eins
	// mit Stern, `-1/-1` ist eine Minuseins.
	ohneStern = strings.TrimSuffix(strings.TrimSpace(ohneStern), "+")

	if zahl, fehler := strconv.Atoi(strings.TrimSpace(ohneStern)); fehler == nil {
		wert.Zahl = zahl
	}

	return wert
}
