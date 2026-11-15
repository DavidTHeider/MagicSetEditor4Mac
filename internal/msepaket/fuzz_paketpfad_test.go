package msepaket_test

// **Wie ein Verweis in ein Paket gelesen wird** — QA-06b
// (#1967),
// Fläche 4 aus `docs/FUZZZIELE.md`: [msepaket.ZerlegePaketpfad],
// [msepaket.ZerlegeEigenpfad] und [msepaket.ZerlegeSetpaketverweis].
//
// # Warum die drei zusammen
//
// **Sie beantworten dieselbe Frage** — welche Datei in welchem Paket ist
// gemeint — für drei Schreibweisen, und jede kommt aus einer fremden Datei:
// `include file:` und `image:` aus einer Vorlage, der Setpaketverweis aus
// einem Bildwert eines Sets. Was hier durchgeht, öffnet der Paketverwalter.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **ein angenommener Verweis, dessen Datei kein gültiger Weg im Paket
//     ist** — geprüft mit [fs.ValidPath], dem Maß, an dem ein [fs.FS] öffnet,
//     und nicht mit dem Leser, der geprüft wird
//   - ein angenommener Paketname mit Schrägstrich, als `.`/`..` oder ohne
//     `.mse-` — und **ein anderes Paket oder eine andere Datei als die, die
//     im Verweis stehen**
//   - **Zerlegen∘Zusammensetzen ist nicht die Identität**: was gelesen wurde,
//     muss, wieder geschrieben, dasselbe ergeben
//   - **ein Eigenverweis und ein voller Verweis auf dieselbe Datei, die
//     Verschiedenes ergeben** — `include file: script` in `magic.mse-game`
//     und `/magic.mse-game/script` meinen dasselbe
//   - ein Setpaketverweis, der beim Zerlegen einen Ort oder eine Datei
//     verliert, die ohne `!` geschrieben wurde
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzPaketpfad -fuzztime 2m

import (
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein angenommener Verweis nennt ein Paket und einen gültigen Weg darin —
// und kommt zurück, wie er ging.**
func FuzzPaketpfad(f *testing.F) {
	for _, saat := range [][2]string{
		{"/magic-modules.mse-include/scripts", "magic.mse-game"},
		{"/magic-m15.mse-style/watermarks/watermarks", "magic.mse-game"},
		{"/a.mse-include/../b.mse-include/x", "a.mse-include"},
		{"//a.mse-x//b/", "a.mse-x"},
		{" /a.mse-x/b ", ".."},
		{"/a.mse-x/./b", "a/b.mse-x"},
		{"script", "magic.mse-game"},
		{"watermarks/watermarks", "magic-m15-extra.mse-style"},
		{"setpaket:/Users/x/Mein Set!.mse-set!image1", "probe.mse-set"},
		{"setpaket:ohne-trenner", ""},
		{"/", ""},
		{"", ""},
	} {
		f.Add(saat[0], saat[1])
	}

	f.Fuzz(func(t *testing.T, verweis, paket string) {
		pruefePaketpfad(t, verweis)
		pruefeEigenpfad(t, paket, verweis)
		pruefeSetpaketverweis(t, verweis, paket)
	})
}

// pruefePaketpfad hält [msepaket.ZerlegePaketpfad].
func pruefePaketpfad(t *testing.T, verweis string) {
	t.Helper()

	pfad, fehler := msepaket.ZerlegePaketpfad(verweis)
	nochmal, zweiter := msepaket.ZerlegePaketpfad(verweis)

	if (fehler == nil) != (zweiter == nil) || pfad != nochmal {
		t.Fatalf("%q zweimal verschieden: %v / %v", verweis, fehler, zweiter)
	}

	if fehler != nil {
		return
	}

	if !fs.ValidPath(pfad.Datei) || pfad.Datei == "." {
		t.Errorf("%q gibt die Datei %q — kein Weg in einem Paket", verweis, pfad.Datei)
	}

	if strings.Contains(pfad.Paket, "/") || pfad.Paket == "." || pfad.Paket == ".." ||
		!strings.Contains(pfad.Paket, ".mse-") {
		t.Errorf("%q gibt das Paket %q", verweis, pfad.Paket)
	}

	// **Zerlegt wird, was dasteht**: das Paket ist das erste nichtleere
	// Glied, die Datei sind die übrigen. Ein Verweis, der über `..` ein
	// anderes Paket erreicht als das genannte, wäre gültig und täte etwas
	// anderes, als er sagt.
	glieder := slices.DeleteFunc(strings.Split(strings.TrimSpace(verweis), "/"),
		func(glied string) bool { return glied == "" })
	if len(glieder) < 2 || pfad.Paket != glieder[0] || pfad.Datei != strings.Join(glieder[1:], "/") {
		t.Errorf("%q gibt %+v — dort steht etwas anderes", verweis, pfad)
	}

	if zurueck, rueckfehler := msepaket.ZerlegePaketpfad("/" + pfad.Paket + "/" + pfad.Datei); rueckfehler != nil || zurueck != pfad {
		t.Errorf("%q → %+v → %+v (%v): Zerlegen∘Zusammensetzen ist nicht die Identität",
			verweis, pfad, zurueck, rueckfehler)
	}
}

// pruefeEigenpfad hält [msepaket.ZerlegeEigenpfad] — und gegen den vollen
// Verweis auf dieselbe Datei.
func pruefeEigenpfad(t *testing.T, paket, verweis string) {
	t.Helper()

	pfad, fehler := msepaket.ZerlegeEigenpfad(paket, verweis)
	nochmal, zweiter := msepaket.ZerlegeEigenpfad(paket, verweis)

	if (fehler == nil) != (zweiter == nil) || pfad != nochmal {
		t.Fatalf("%q in %q zweimal verschieden: %v / %v", verweis, paket, fehler, zweiter)
	}

	if fehler != nil {
		return
	}

	if pfad.Paket != paket {
		t.Errorf("%q in %q nennt das Paket %q", verweis, paket, pfad.Paket)
	}

	if !fs.ValidPath(pfad.Datei) || pfad.Datei == "." {
		t.Errorf("%q in %q gibt die Datei %q — kein Weg im Paket", verweis, paket, pfad.Datei)
	}

	if zurueck, rueckfehler := msepaket.ZerlegeEigenpfad(paket, pfad.Datei); rueckfehler != nil || zurueck != pfad {
		t.Errorf("%q → %+v → %+v (%v): Zerlegen∘Zusammensetzen ist nicht die Identität",
			verweis, pfad, zurueck, rueckfehler)
	}

	// **Nur, wo das Paket selbst ein voller Verweis sein kann** — sonst
	// gibt es den Vergleich nicht.
	if voll, vollfehler := msepaket.ZerlegePaketpfad("/" + paket + "/x"); vollfehler != nil || voll.Paket != paket {
		return
	}

	if voll, vollfehler := msepaket.ZerlegePaketpfad("/" + paket + "/" + pfad.Datei); vollfehler != nil || voll != pfad {
		t.Errorf("%q in %q gibt %+v, der volle Verweis %+v (%v)", verweis, paket, pfad, voll, vollfehler)
	}
}

// pruefeSetpaketverweis hält [msepaket.ZerlegeSetpaketverweis] in beide
// Richtungen: den Fuzzwert als Verweis und den gebauten aus Ort und Datei.
func pruefeSetpaketverweis(t *testing.T, verweis, datei string) {
	t.Helper()

	ort, gelesen, ist := msepaket.ZerlegeSetpaketverweis(verweis)

	if ist != strings.HasPrefix(strings.TrimSpace(verweis), "setpaket:") {
		t.Errorf("%q: ist ein Setpaketverweis = %v", verweis, ist)
	}

	if ist && strings.Contains(gelesen, "!") {
		t.Errorf("%q gibt die Datei %q — getrennt wird am letzten `!`", verweis, gelesen)
	}

	// **Gebaut aus dem Fuzzwert als Ort und einer Datei ohne `!`** — so,
	// wie ein Set seine Bildwerte schreibt. **Leerraum am Ende bleibt
	// erlaubt**: der Zerteiler schneidet einen Wert nur links, wie MSE2
	// (`trim_left`, `src/util/io/reader.cpp`, Zeile 224, Fassung f9d9356 vom
	// 2021-07-06), und `image: image1 ` nennt dort die Datei `image1 `.
	if strings.Contains(datei, "!") {
		return
	}

	gebaut := msepaket.SetpaketVerweis(ort, datei)

	zurueckOrt, zurueckDatei, zurueckIst := msepaket.ZerlegeSetpaketverweis(gebaut)
	if !zurueckIst || zurueckOrt != ort || zurueckDatei != datei {
		t.Errorf("Ort %q und Datei %q → %q → %q, %q", ort, datei, gebaut, zurueckOrt, zurueckDatei)
	}
}
