package domain

// **Die Schlüssel der Symbolbibliothek** — DOM-02
// (#1697).
// Eine eigene Datei, weil `fehlerschluessel.go` an der Zeilengrenze aus §2
// stand; angemeldet werden sie mit allen übrigen in `fehlerschluesselwege.go`.

const (
	// SchluesselSymbolZuVieleTeile, …ZuVielePunkte und …ZuTief heißen: das
	// Symbol überschreitet eine Grenze aus `msepaket` (AK-04). Werte:
	// `anzahl` und `grenze`.
	SchluesselSymbolZuVieleTeile  Schluessel = "symbolZuVieleTeile"
	SchluesselSymbolZuVielePunkte Schluessel = "symbolZuVielePunkte"
	SchluesselSymbolZuTief        Schluessel = "symbolZuTief"

	// SchluesselSymbolZuGross heißt: die Datei ist größer als eine
	// `.mse-symbol` sein darf. Wert: `grenze`.
	SchluesselSymbolZuGross Schluessel = "symbolZuGross"

	// SchluesselSymbolGespeichertZuGross heißt: das Symbol wäre gespeichert
	// größer als eine `.mse-symbol` sein darf — es wurde gezeichnet,
	// abgelegt oder nachgezeichnet, nicht als Datei gewählt. Wert: `grenze`
	// (#1973).
	SchluesselSymbolGespeichertZuGross Schluessel = "symbolGespeichertZuGross"

	// SchluesselSetOhneSymbol heißt: das Set trägt noch keines, das sich
	// aufnehmen ließe.
	SchluesselSetOhneSymbol Schluessel = "setOhneSymbol"

	// SchluesselSymbolVerwendet heißt: Sets tragen dieses Symbol, und
	// gelöscht wird nur mit Bestätigung. Ein Konflikt. Wert: `sets`.
	SchluesselSymbolVerwendet Schluessel = "symbolVerwendet"

	// SchluesselSymbolOhneVorige heißt: dieses Symbol hat keine vorige
	// Fassung, die sich zurückholen ließe (AK-02).
	SchluesselSymbolOhneVorige Schluessel = "symbolOhneVorige"

	// SchluesselFremdeEigeneKarten heißt: ein Deck aus einem Umschlag trug
	// Plätze auf **eigene** Karten des Absenders; ihre Kennung gilt nur in
	// seinem Bestand, und sie kamen nicht mit (FMT-04, #1693). Wert: `karten`.
	SchluesselFremdeEigeneKarten Schluessel = "fremdeEigeneKarten"

	// SchluesselNichtGelegtePlaetze heißt: ein Deck aus einem Umschlag ist
	// angenommen, aber diese Plätze ließen sich hier nicht legen (#1995).
	// Wert: `plaetze`.
	SchluesselNichtGelegtePlaetze Schluessel = "nichtGelegtePlaetze"

	// SchluesselFremdeKartenUndNichtGelegte heißt beides zugleich. Werte:
	// `karten`, `plaetze`.
	SchluesselFremdeKartenUndNichtGelegte Schluessel = "fremdeKartenUndNichtGelegte"
)
