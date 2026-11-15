package service

// **Die Namen der Handlungen** — Werkstücke M7 und M15 der Regelmaschine
// (#1609,
// #1617).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Befehle, die den Zug oder einen Spieler meinen**; hier
// stehen die, die eine **Regel aus CR 701, 705 oder 706** benennen. Und
// `zugbefehle.go` stand mit beidem bei 408 Zeilen (§2).

const (
	// Die **Handlungen mit Schlüsselwort**, die ohne Kartentext auskommen —
	// CR 701, Werkstück M7
	// (#1609).
	//
	// # Warum sie als Befehle dastehen
	//
	// **Weil ein Mensch sie am Tisch tut.** Was eine Karte sagt, führt
	// weiterhin niemand aus (CR 608.2c) — wer einen Blitz wirkt, zerstört
	// danach von Hand, so wie er heute eine Karte umlegt. Damit sind sie
	// **die Bausteine**, aus denen `M8` später die Kartentexte
	// zusammensetzt, und sie sind schon da, wenn es so weit ist.
	//
	// **Alle acht brauchen Priorität**, wie jede Handlung am Tisch
	// (CR 117.1): sie bewegen Karten zwischen Zonen.
	BefehlZerstoeren     = "objekt.zerstoeren"
	BefehlOpfern         = "objekt.opfern"
	BefehlExilieren      = "objekt.exilieren"
	BefehlNeutralisieren = "objekt.neutralisieren"
	BefehlKaempfen       = "objekt.kaempfen"
	BefehlMahlen         = "zone.mahlen"
	BefehlHellsehen      = "zone.hellsehen"
	BefehlUeberwachen    = "zone.ueberwachen"

	// BefehlMuenzwurf wirft eine Münze — CR 705, Werkstück M15
	// (#1617).
	//
	// **Er ändert nichts am Tisch.** Was ein Wurf bewirkt, steht auf einer
	// Karte (`M8`); was hier entsteht, ist sein **Ergebnis im
	// Ereignisstrom** — und der hängt an der Hashkette (G1), also lässt es
	// sich nachträglich nicht ändern.
	BefehlMuenzwurf = "spieler.muenzwurf"

	// BefehlWuerfeln würfelt — CR 706 (M15).
	//
	// **Die Seitenzahl kommt mit**: CR 706.1 sagt, dass die Anweisung sie
	// nennt. Ein sechsseitiger Würfel ist kein Naturgesetz.
	BefehlWuerfeln = "spieler.wuerfeln"
)
