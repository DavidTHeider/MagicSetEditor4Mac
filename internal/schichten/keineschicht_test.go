package schichten

// **Was unter `internal/` keine Schicht ist** — und warum.
//
// Eigene Datei, weil `regeln_test.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **wer wen kennen darf**; hier, **wer
// gar nicht mitspielt**.
//
// **Ein leerer Grund ist keiner** — der Wächter in `schichten_test.go` weist
// ihn ab.

// keineSchicht sind die Pakete unter internal/, die keine Schicht sind, samt
// Grund. Ein leerer Grund ist keiner.
var keineSchicht = map[string]string{
	"flaechen": "reines Abnahmepaket ohne Produktivcode: es baut `cmd/mse` und " +
		"`cmd/mse-mcp` und fährt denselben Arbeitsablauf über beide — die " +
		"Abnahme von Phase T (#886). Zwei `package main` lassen sich nicht " +
		"zusammen einbinden; gefahren wird deshalb, was jemand bekommt",
	"werkzeuge": "reines Wächterpaket ohne Produktivcode; prüft die Versionspinnung, nicht die Fachlichkeit",
	"schichten": "dieses Paket selbst — es prüft die Regel und steht damit außerhalb von ihr",
}
