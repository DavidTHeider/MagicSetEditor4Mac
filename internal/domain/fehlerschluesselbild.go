package domain

// **Der Schlüssel eines Bildes, das zu groß ist** — FMT-06b
// (#1974).
// Eine eigene Datei, weil `fehlerschluessel.go` an der Zeilengrenze aus §2
// steht; angemeldet wird er mit allen übrigen in `fehlerschluesselwege.go`.

// SchluesselBildZuGross heißt: ein Bild ist zu groß, um es auszupacken — über
// der Grenze je Seite, der Bildpunkte oder des Speichers (`msepaket`). Bis
// hierher hieß das an jeder Fläche „kein lesbares Bild", und das schickte
// jemanden nach einem Fehler in einer Datei suchen, die in Ordnung ist.
const SchluesselBildZuGross Schluessel = "bildZuGross"
