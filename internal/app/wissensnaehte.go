package app

// **Die Nähte um das, was dieses Programm über Karten weiß** — der lokale
// Suchindex, die Recherche darüber und das Lesen eines Kartenfotos.
//
// # Warum sie zusammenstehen
//
// **Weil sie eine Frage teilen und eine Stelle**: alle drei enden bei
// denselben Quellen. Der Index von A15 trägt sie, die Recherche von A16 sucht
// darin, und die Bilderkennung von A13.3 führt ihre gelesenen Stellen an
// **dieselbe** Drucksuche, die eine getippte Kartenliste geht (A2). Ein Foto
// kann damit keinen Druck nennen, den die Suche nicht auch gefunden hätte.
//
// **Eine eigene Datei und ein eingebetteter Verbund**, seit `anwendung.go`
// mit dem Bilderkenner über die Zeilengrenze aus §2 ging — dieselbe Form wie
// `bildnaehte`. Eingebettet heißt: `a.recherche` bleibt `a.recherche`; der
// Schnitt ist einer der **Datei**, nicht der Naht.

type wissensnaehte struct {
	// retrieval führt den lokalen Suchindex — Werkstück A15 von Phase 5.3
	// (#1514).
	retrieval Retrievalfuehrer

	// indexabbruch hält den Abbruch des Aufbaus, der gerade läuft — eigens
	// hier und nicht nur im Führer, damit ein Abschalten ihn erreicht, bevor
	// der Führer seinen eigenen eingetragen hat (Zweitblick über Phase
	// 6.12, Z2).
	indexabbruch abbruchplatz

	// recherche beantwortet Fragen gegen den Wissensbestand — Werkstück A16
	// von Phase 5.3
	// (#1515).
	recherche Recherchefuehrer

	// bilderkenner macht aus gelesenen Stellen Kandidaten — Werkstück A13.3
	// (#1452).
	//
	// **Er fragt kein Modell.** Das tut die Naht; er nimmt entgegen, was sie
	// gelesen hat, und sucht damit über **A2**.
	bilderkenner Bilddeuter
}
