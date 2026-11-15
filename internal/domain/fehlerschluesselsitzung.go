package domain

// **Die Schlüssel der lokalen Sitzung** — ID-02
// (#1701).
//
// **Eigene und nicht `validierung`**: mit ihm zeigte die Oberfläche nur „Die
// Eingabe ist nicht gültig" — und niemand erführe, ob das Kennwort falsch,
// zu kurz oder gar keines da ist (§6). Eine eigene Datei, damit
// `fehlerschluessel.go` unter der Zeilengrenze aus §2 bleibt.
const (
	// SchluesselKennwortFalsch heißt: dieses Kennwort gehört nicht zu diesem
	// Nutzer.
	SchluesselKennwortFalsch Schluessel = "kennwortFalsch"

	// SchluesselKennwortLaenge heißt: ein neues Kennwort ist zu kurz oder zu
	// lang; die Grenzen reisen als Werte mit.
	SchluesselKennwortLaenge Schluessel = "kennwortLaenge"

	// SchluesselSperreOhneRueckweg heißt: gesperrt ließe sich nicht mehr
	// entsperren — kein Kennwort, keine lokale Entsperrung.
	SchluesselSperreOhneRueckweg Schluessel = "sperreOhneRueckweg"

	// SchluesselLokaleEntsperrungFehlt heißt: Touch ID ist für diesen Nutzer
	// auf diesem Rechner nicht eingerichtet.
	SchluesselLokaleEntsperrungFehlt Schluessel = "lokaleEntsperrungFehlt"

	// SchluesselCodeFalsch heißt: dieser Code steht auf keinem Zettel dieses
	// Nutzers — oder ist schon verbraucht.
	SchluesselCodeFalsch Schluessel = "codeFalsch"
)
