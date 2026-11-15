// **Die Naht zum Arbeitsablauf** — einfahren, sichern, zeichnen, den Bogen
// legen, MTG JSON.
//
// **Eine eigene Datei**, seit `naehte.go` mit dem Probelauf (UXR-F7,
// #2015) über
// die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort stehen die
// Nähte in ihrer Gesamtheit, hier die eine, die schreibt — und neben ihr in
// `probelauf.go` die, die nur sagt, was sie schriebe.

package rest

import "context"

// Ausfuhr ist, was die drei Endpunkte des Arbeitsablaufs brauchen.
//
// **Sie ist nicht `mcp.Ausfuhr`, obwohl sie gleich aussieht.** Beide Pakete
// sind Blätter und kennen einander nicht (§2). Die Wiederholung ist der Preis
// dafür, und sie ist der billigere: ein gemeinsames Paket dazwischen wäre eine
// Schicht, die niemand braucht, und ein Import zwischen den beiden machte aus
// zwei Blättern eine Kette.
type Ausfuhr interface {
	Einfahren(ctx context.Context, pfad string) (setID int64, karten int, befunde []string, fehler error)
	Sichern(ctx context.Context, setID int64, ziel string) (befunde []string, fehler error)
	Zeichnen(ctx context.Context, karteID int64, ziel, guete string) (
		breite, hoehe int, befunde []string, fehler error,
	)

	// Bogen legt die Karten eines Sets als Druckbogen (#1291).
	Bogen(ctx context.Context, setID int64, ziel string) (
		seiten, karten int, befunde []string, fehler error,
	)

	// MTGJSONLesen fährt eine MTG-JSON-Quelle in den Bestand ein (#1291).
	MTGJSONLesen(ctx context.Context, pfad, spiel string) (
		angelegt []int64, befunde []string, fehler error,
	)

	// MTGJSONSchreiben schreibt ein Set als MTG JSON (#1291).
	MTGJSONSchreiben(ctx context.Context, setID int64, spiel, ziel string) (
		karten int, befunde []string, fehler error,
	)

	// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt (#1291).
	MTGJSONSpiele() []string
}
