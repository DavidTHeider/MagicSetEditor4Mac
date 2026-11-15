//go:build fremdpakete

// Der Satz der Demokarten
// (#476).
//
// Er steht in einer eigenen Datei, weil `demokarten_test.go` mit ihm über 400
// Zeilen ginge (§2). Welche Achsen er decken muss und wer das prüft, steht in
// `demokartenachsen_test.go`.
package app_test

// demokarten sind die Karten des Satzes, von Hand geschrieben.
//
// **Eigener Inhalt** (§11): kein Name, kein Regeltext und keine Illustration
// stammt aus einem fremden Paket.
//
// **Jede trägt ihren Fuß** — Illustrator, Copyright, Setcode. Echte Karten
// haben diese Zeile, und sie war bis
// #382
// schwarz auf schwarz gesetzt: gezeichnet und unsichtbar.
//
// # Die ersten sieben sind historisch, der Rest ist systematisch
//
// Karte 1 bis 7 sind entstanden, weil ein Befund sie verlangt hat. Was danach
// kommt, deckt die Achsen aus `demokartenachsen_test.go` — Rahmenfarbe,
// Vorlage, Textform —, und jede neue Karte deckt **mehrere** Achsenwerte
// zugleich. Sonst wären es fünfunddreißig statt einundzwanzig, und jede
// zusätzliche kostet einen Planbau und eine Rasterung.
func demokarten() []demokarte {
	return mitFuss([]demokarte{
		{
			name: "Kreatur", stil: "magic-m15", bilder: 6, texte: 8,
			felder: map[string]string{
				"name": "Wächter der Werkbank", "casting cost": "2W",
				"super type": "Creature", "sub type": "Human Artificer",
				"rule text":   "Vigilance\n<sym>T</sym>: Add <sym>W</sym>.",
				"flavor text": "<i-flavor>Wer misst, gewinnt.</i-flavor>",
				"power":       "2", "toughness": "3",
				"card color": "white", "rarity": "uncommon", "card number": "1",
			},
		},
		{
			name: "Spontanzauber", stil: "magic-m15", bilder: 6, texte: 7,
			felder: map[string]string{
				"name": "Kurzschluss", "casting cost": "1U",
				"super type": "Instant",
				"rule text":  "Counter target spell unless its controller pays <sym>2</sym>.",
				"card color": "blue", "rarity": "common", "card number": "2",
			},
		},
		{
			name: "Mehrfarbig", stil: "magic-m15", bilder: 6, texte: 8,
			felder: map[string]string{
				"name": "Zwietracht und Zwiespalt", "casting cost": "1BR",
				"super type": "Legendary Creature", "sub type": "Demon Rogue",
				"rule text": "Menace\nWhenever this creature attacks, it deals 1 damage to each opponent.",
				"power":     "3", "toughness": "2",
				"card color": "black, red, multicolor", "rarity": "rare", "card number": "3",
			},
		},
		{
			name: "Land", stil: "magic-m15", bilder: 6, texte: 6,
			felder: map[string]string{
				"name": "Verwitterte Schmiede", "super type": "Land",
				"rule text":  "<sym>T</sym>: Add <sym>C</sym>.",
				"card color": "land", "rarity": "common", "card number": "4",
			},
		},
		{
			// **Ein anderer Stil**, damit dieser Lauf nicht eine einzige
			// Vorlage prüft und „die Karten stimmen" darüber sagt.
			//
			// # Eine Saga speichert ihre Kapitel einzeln
			//
			// Hier stand bis zum 2026-08-18 ein `rule text` — und dieselbe
			// Zeichenkette erschien danach zweimal auf der Karte
			// (#482).
			// **Das war kein Fehler im Zeichnen.** `magic.mse-game/script`
			// gibt den Regeltext bei einer Saga an **zwei** Felder weiter:
			//
			//	chapter_text_filter_default := { … if card.rule_text != ""
			//	    then card.rule_text else "<i-auto>(…)</i-auto>" … }
			//	text_filter_default := { if transfer_levels() then
			//	    (if margin_code == "lv1" then card.rule_text …) }
			//
			// Zeilen 3174–3195, und `transfer_levels := {true}` in Zeile 369.
			// Wer einer Saga einen `rule text` gibt, bekommt ihn im
			// Kapitelkasten **und** in Stufe 1 — bei MSE genauso.
			//
			// **Eine echte Saga führt keinen `rule text`.** `Cleric Class` aus
			// dem kanonischen Set `White` speichert `chapter text`,
			// `level_1_text`, `level_2_text` und `level_3_text` einzeln und
			// `rule text` gar nicht. Diese Vorrichtung war großzügiger als der
			// Betrieb (§3) — sie speicherte ein Feld, das MSE hier nie
			// schreibt.
			name: "Zeitalter", stil: "magic-m15-saga", bilder: 4, texte: 6,
			felder: map[string]string{
				"name": "Weg der langen Schatten", "casting cost": "2B",
				"super type": "Enchantment", "sub type": "Saga",
				"chapter text": "<i-auto>(Sobald diese Saga ins Spiel kommt und " +
					"nach deinem Ziehsegment, lege eine Überlieferungsmarke darauf.)</i-auto>",
				"level 1 text": "Jeder Gegner wirft eine Karte ab.",
				"level 2 text": "Bringe eine Kreatur aus deinem Friedhof zurück.",
				"card color":   "black", "rarity": "uncommon", "card number": "5",
			},
		},
		{
			// **Mit eigener Illustration**, und das ist die sechste Sorte:
			// jede echte Karte trägt eine, und bis #380 blieb das Kunstfenster
			// leer. Der Nachweis steht unten — die gemalte Ecke muss im Bild
			// ankommen.
			name: "Illustriert", stil: "magic-m15", bilder: 7, texte: 8,
			kunst: "kunst.png",
			felder: map[string]string{
				"name": "Leuchtturm am Nordriff", "casting cost": "3G",
				"super type": "Creature", "sub type": "Elemental",
				"rule text": "Reach\nWhenever this creature enters, draw a card.",
				"power":     "4", "toughness": "4",
				"card color": "green", "rarity": "rare", "card number": "6",
				"image": "kunst.png",
			},
		},
		{
			// **Ein Artefakt mit Farbindikator** — die siebte Sorte.
			//
			// Der Punkt vor der Typzeile erscheint **nicht** von selbst:
			// `has_identity_general` zeigt ihn nur, wenn die Identität der
			// Rahmenfarbe widerspricht, und sein Vorgabewert ist die
			// Rahmenfarbe selbst (`magic.mse-game/card_fields`, Zeile 771:
			// `default: stylesheet; card.card_color`). Ein farbloses Artefakt
			// mit ausdrücklichem `indicator: white` ist der Fall, in dem er
			// gehört (#390).
			name: "Artefakt", stil: "magic-m15", bilder: 6, texte: 7,
			indikator: true,
			felder: map[string]string{
				"name": "Werkbank der Erfinderin", "casting cost": "2",
				"super type": "Artifact", "rule text": "<sym>T</sym>: Add <sym>C</sym>.",
				"card color": "artifact", "indicator": "white",
				"rarity": "uncommon", "card number": "7",
			},
		},
		{
			// **Farblos** — die `empty choice` des Feldes `card color`
			// (`magic.mse-game/card_fields`, Zeile 112). Ein Eldrazi-artiges
			// Wesen ohne Farbe, und dazu die Textform **einzeilig**.
			name: "Farblos", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Wandler ohne Namen", "casting cost": "6",
				"super type": "Creature", "sub type": "Eldrazi",
				"rule text": "Devoid",
				"power":     "5", "toughness": "5",
				"card color": "colorless", "rarity": "rare", "card number": "8",
			},
		},
		{
			// **Hybrid**, und dazu die Textform **Aufzählung**: `<bullet>` ist
			// die Auszeichnung, an der ein Umbruch anders hängt als in einem
			// Absatz (#431).
			name: "Hybrid", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Zwiegesicht der Küste", "casting cost": "W/U W/U",
				"super type": "Creature", "sub type": "Merfolk Cleric",
				"rule text": "<bullet>Fliegend</bullet>\n<bullet>Wachsamkeit</bullet>",
				"power":     "2", "toughness": "2",
				"card color": "white, blue, hybrid", "rarity": "uncommon",
				"card number": "9",
			},
		},
		{
			// **Waagerecht geteilt.** `horizontal`, `vertical` und `radial`
			// sind `type: radio` und schließen einander aus — deshalb drei
			// Karten und nicht eine.
			name: "Waagerecht", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Bruch im Horizont", "casting cost": "2RG",
				"super type": "Sorcery",
				"rule text":  "Füge einem Ziel deiner Wahl 3 Schadenspunkte zu.",
				"card color": "red, green, multicolor, horizontal",
				"rarity":     "rare", "card number": "10",
			},
		},
		{
			name: "Senkrecht", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Geteilte Tiefe", "casting cost": "1UB",
				"super type": "Instant",
				"rule text":  "Ziehe zwei Karten, dann wirf eine Karte ab.",
				"card color": "blue, black, multicolor, vertical",
				"rarity":     "uncommon", "card number": "11",
			},
		},
		{
			name: "Strahlenförmig", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Rat der drei Häuser", "casting cost": "WUB",
				"super type": "Legendary Creature", "sub type": "Human Advisor",
				"rule text": "Immer wenn diese Kreatur angreift, ziehe eine Karte.",
				"power":     "3", "toughness": "4",
				"card color": "white, blue, black, multicolor, radial",
				"rarity":     "mythic rare", "card number": "12",
			},
		},
		{
			// **Überlagert** ist nur wählbar, wenn genau zwei Farben stehen
			// **und** `hybrid` gewählt ist (`card_fields`, Zeile 136).
			name: "Überlagert", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Wurzel und Licht", "casting cost": "G/W G/W",
				"super type": "Enchantment",
				"rule text":  "Kreaturen, die du kontrollierst, erhalten +1/+1.",
				"card color": "green, white, hybrid, overlay",
				"rarity":     "rare", "card number": "13",
			},
		},
		{
			// **Umgekehrt** ist wählbar, sobald zwei Farben stehen und
			// `overlay` **nicht** gewählt ist (`card_fields`, Zeile 139).
			name: "Umgekehrt", stil: "magic-m15", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Kehrseite des Hains", "casting cost": "2BG",
				"super type": "Artifact Creature", "sub type": "Horror",
				"rule text": "Todesberührung",
				"power":     "4", "toughness": "1",
				"card color": "black, green, multicolor, reversed",
				"rarity":     "rare", "card number": "14",
			},
		},
		{
			// **Planeswalker** — eine eigene Vorlage mit Loyalitätszeilen.
			// **Die Mitte einer weißen Planeswalkerkarte ist weiß** — leeres
			// Kunstfenster, leerer Textkasten, und Weiß hat R = G = B.
			// Gemessen: 19,4 % farbig.
			name: "Planeswalker", stil: "magic-m15-planeswalker",
			bilder: 1, texte: 1, buntMindestens: 15, absatzband: true,
			felder: map[string]string{
				"name": "Ilva, Hüterin der Uhr", "casting cost": "3WU",
				"super type": "Legendary Planeswalker", "sub type": "Ilva",
				// **Die Fähigkeitstexte stehen in `rule text`, `rule text 2`
				// und `rule text 3`** — `text`, `text 2`, `text 3` sind
				// daraus gerechnet und tragen `save value: false`
				// (`magic.mse-game/card_fields`, Zeilen 1339–1393).
				"loyalty cost 1": "+1", "loyalty cost 2": "-2",
				"loyalty cost 3": "-7",
				"rule text": "Ziehe eine Karte.\n" +
					"Schicke eine Kreatur ins Exil.\n" +
					"Du erhältst ein Emblem.",
				"loyalty":    "4",
				"card color": "white, blue, multicolor", "rarity": "mythic rare",
				"card number": "15",
			},
		},
		{
			// **Spielstein.** Er trägt keine Manakosten und keine Seltenheit —
			// die Vorlage rechnet damit.
			name: "Spielstein", stil: "magic-m15-token", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Wolf", "super type": "Token Creature", "sub type": "Wolf",
				"power": "2", "toughness": "2",
				"card color": "green", "card number": "16",
			},
		},
		{
			// **Stufenkarte** — drei Textkästen übereinander.
			name: "Stufenkarte", stil: "magic-m15-leveler", bilder: 1, texte: 1,
			mischart: "multiply",
			felder: map[string]string{
				"name": "Lehrling der Schmiede", "casting cost": "1R",
				"super type": "Creature", "sub type": "Human Artificer",
				"rule text": "Stufe auf <sym>2</sym>",
				"power":     "1", "toughness": "2",

				// **Die Stufen sind gesetzt, und das ist der Punkt.** Der
				// Feldstil `lvlup multiply 0` trägt
				// `image: {set_combine(input: …, combine: "multiply")}` und
				// `visible: {card.level_0 != ""}` (Zeilen 659–667 des Stils).
				// Ohne einen Stufenwert steht die Ebene nicht im Plan — und
				// dann prüft **keine** Demokarte, ob eine Mischart ankommt
				// (#470).
				"level 0": "1", "level 1": "2", "level 2": "3",
				"level 0 text": "Erstschlag", "level 1 text": "Wachsamkeit",
				"level 2 text": "Doppelschlag",
				"card color":   "red", "rarity": "rare", "card number": "17",
			},
		},
		{
			// **Wendekarte** — zwei Hälften, die zweite auf dem Kopf.
			name: "Wendekarte", stil: "magic-m15-flip", bilder: 1, texte: 1,
			felder: map[string]string{
				"name": "Späher am Pass", "casting cost": "1G",
				"super type": "Creature", "sub type": "Human Scout",
				"rule text": "Immer wenn diese Kreatur Schaden zufügt, wende sie.",
				"power":     "1", "toughness": "1",
				"card color": "green", "rarity": "uncommon", "card number": "18",
			},
		},
		{
			// **Textlos**, und dazu die Textform **leer**: ein Regeltext, den
			// es nicht gibt, ist eine eigene Klasse — die Messstufe rechnet
			// dann über null Zeilen.
			// **Eine textlose Vorlage hat keinen Textkasten**; was farbig
			// ist, ist allein der Rahmen. Gemessen: 13,5 %.
			// **Und sie zeichnet keine Typzeile** — im Bild steht keine, und
			// der Stil führt für `super type` keinen Feldstil.
			name: "Textlos", stil: "magic-m15-textless",
			bilder: 1, texte: 1, buntMindestens: 10,
			ohneFelder: []string{"super type"},
			felder: map[string]string{
				"name": "Stiller Wächter", "casting cost": "2W",
				"super type": "Creature", "sub type": "Spirit",
				"power": "3", "toughness": "3",
				"card color": "white", "rarity": "common", "card number": "19",
			},
		},
		{
			// **Textloses Land** — die Vorlage der Standardländer.
			// **Bei einer Vollbildvorlage ist die ganze Karte das
			// Kunstfenster** — ohne Illustration also schwarz. Gemessen:
			// 76,5 % dunkel.
			name: "Textloses Land", stil: "magic-m15-textless-land",
			bilder: 1, texte: 1, buntMindestens: 10, dunkelHoechstens: 85,
			felder: map[string]string{
				"name": "Ebene", "super type": "Basic Land", "sub type": "Plains",
				"card color": "land", "rarity": "common", "card number": "20",
			},
		},
		{
			// **Verschmelzbare Teilung** — die Vorlage, deren `init script:`
			// bis #468
			// nicht ausführbar war. Sie steht hier, damit das nicht ein
			// zweites Mal unbemerkt passiert.
			name: "Teilung", stil: "magic-m15-split-fusable", bilder: 1, texte: 1,

			// **Beide Hälften sind Spontanzauber**, und die Typzeile steht
			// deshalb zweimal auf der Karte — aus **zwei** Feldern,
			// `super type` und `super type 2`. Das ist die Karte und kein
			// Mangel; benannt wird es, damit
			// [pruefeKeinDoppelterText] scharf bleiben kann
			// (#482).
			doppelt: []string{"Instant —"},
			felder: map[string]string{
				"name": "Feuer", "casting cost": "1R",
				"super type": "Instant",
				"rule text":  "Füge einem Ziel deiner Wahl 2 Schadenspunkte zu.",
				"name 2":     "Eis", "casting cost 2": "1U",
				"super type 2": "Instant",
				"rule text 2":  "Neutralisiere einen Zauberspruch deiner Wahl.",
				"card color":   "red", "card color 2": "blue",
				"rarity": "uncommon", "card number": "21",
			},
		},
	})
}
