package app

// **Die Naht zum Bildweg** —
// #1270.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"

// Bildwegpfad ist der Anfang jeder Bildadresse — [vorschauweg.Pfad].
//
// **Hier noch einmal**, damit die Tests der Fassade ihn nennen können, ohne
// dass jede von ihnen das Blatt einführt.
const Bildwegpfad = vorschauweg.Pfad

// Bildfaecher ist die Zahl der wartenden Bilder — [vorschauweg.Faecher].
const Bildfaecher = vorschauweg.Faecher

// Spielbildbudget ist, wie viele Bytes der Spieltisch an Texturen halten darf.
//
// **128 MB, und die Zahl ist ein Messstartwert** (G2,
// #1459).
// Eine Karte der Tischgüte ist 512 Bildpunkte breit und wiegt gepackt
// überschlägig ein Viertel Megabyte; ein Tisch mit vier Sitzen zeigt keine
// sechzig **gleichzeitig**, hält aber die Bibliotheken bereit. Gepinnt wird
// die Zahl, wenn am echten Tisch gemessen wurde — bis dahin sagt dieser Satz,
// dass sie es nicht ist.
//
// **Und das Issue verlangt die Ehrlichkeit ausdrücklich**: „GPU-Texturspeicher
// wird als Schätzung ausgewiesen, wenn kein belastbarer Zähler verfügbar ist."
// Was hier zählt, sind **Bytes im Arbeitsspeicher dieses Prozesses** — was die
// Grafikkarte daraus macht, weiß dieses Fach nicht.
const Spielbildbudget = 128 << 20

// Bildweg gibt den Weg, den `main.go` an die Assetleitung hängt.
//
// **Eine Funktion und keine Methode**, obwohl sie auf `a` zugreift: Wails
// bindet jede exportierte **Methode** von [App] ins Fenster. Eine Bindung, die
// einen Weg zurückgibt, wäre in JavaScript nicht darstellbar — und sie zählte
// in `docs/FAEHIGKEITEN.md` und `docs/SCHREIBWEGE.md` als Fähigkeit mit, die
// keine ist.
//
// **Der Rückgabetyp ist [vorschauweg.Weg] und nicht `http.Handler`.** So kommt
// weder in diese Fassade noch in `main.go` ein `net/http` — der Typ wird
// nirgends genannt, nur benutzt. Warum das nötig ist, steht bei [vorschauweg].
func Bildweg(a *App) *vorschauweg.Weg {
	weg := vorschauweg.Neu(&a.bildfach)

	// **Beide Fächer an einer Leitung** (G2, #1459) — welches gemeint ist,
	// sagt der Vorsatz der Adresse. Ohne Spieltisch gibt es keines, und ein
	// Abruf unter dem Spielpfad endet dann mit 404 statt mit einem Sturz.
	if a.spielbilder != nil {
		weg = weg.MitTexturen(a.spielbilder)
	}

	// **Und das eigene Tischbild** (G3,
	// #1460).
	// Es hängt an keiner Bedingung: die Ablage antwortet mit „nichts da", wo
	// keines liegt, und das ist derselbe 404 wie ohne Fach.
	//
	// **Und das eigene Fensterbild** (UX-09,
	// #1681)
	// — aus demselben Grund ohne Bedingung.
	return weg.MitTischbildern(Tischbilder(a)).MitFensterbildern(Fensterbilder(a))
}

// vorschauspeicher gibt das Fach der Vorschauen als Zwischenspeicher — OPS-08
// (#1725).
func (a *App) vorschauspeicher() ZwischenspeicherDTO {
	zahlen := a.bildfach.Zahlen()

	return ZwischenspeicherDTO{
		Name: SpeicherVorschau, Ort: OrtArbeitsspeicher, Treffer: zahlen.Treffer,
		Fehlgriffe: zahlen.Fehlgriffe, Eintraege: zahlen.Eintraege, Bytes: zahlen.Bytes,
		GrenzeEintraege: zahlen.GrenzeEintraege,
	}
}
