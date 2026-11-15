package app

// **Redigieren: was aus einem Text heraus muss, bevor er das Haus verlässt** —
// Werkstück O10 von Phase 5.2
// (#1434).
//
// # Die Zusage steht zuerst in der **Auswahl**, nicht hier
//
// Ein Diagnosepaket enthält kein Geheimnis, weil es die Stellen, an denen
// Geheimnisse liegen, **gar nicht liest**: nicht die Tokentafel, nicht den
// Schlüsselbund, nicht die Kennwörter, nicht die TOTP-Geheimnisse, nicht die
// Antworten fremder Anbieter, nicht die Sammlung, nicht die Kartentexte.
// `domain.Diagnoseteile()` zählt auf, was hineinkommt — und das ist die
// eigentliche Absicherung.
//
// **Mit einer Quelle, die doch freien Text trägt: der Absturzspur** (F04,
// #2343). Ihr Kopf nennt den Wert, den der Code mit `panic` warf; er geht
// nicht mit, nur was das Laufzeitsystem schreibt (`absturzspurgestalt.go`,
// Zweitblick über Phase 1 von 0.9.0, Befund 7) — **als Liste dessen, was
// bleibt**, seit dem Nachlauf darüber (Befund 5). **Das Protokoll ist
// davon nicht erfasst**: eine gefangene Panik steht dort mit ihrem Wert unter
// `grund`, und es gilt die Gestaltregel unten (#2399).
//
// # Was hier steht, ist die zweite Reihe
//
// **Für den einen Teil mit freiem Text: das Protokoll.** Seine Merkmale sind
// angemeldet ([erlaubteMerkmale]) und seine Meldungen sind Zeichenketten­
// literale — aber `grund` trägt, was der Unterbau sagt, und der kann ein
// Token in einer Fehlermeldung nennen. Dagegen hilft keine Liste erlaubter
// Namen, sondern nur die **Gestalt**.
//
// # Was das kostet, gehört dazu
//
// Eine lange Zeichenfolge aus Buchstaben **und** Ziffern wird geschwärzt —
// auch wenn sie eine Prüfsumme ist und kein Geheimnis. Ein SHA-256 in einer
// Protokollzeile ist danach nicht mehr lesbar. **Der Handel ist bewusst**: ein
// verlorener Hash kostet eine Nachfrage, ein durchgereichtes Token kostet den
// Zugang.

import (
	"regexp"
)

// Geschwaerzt steht anstelle dessen, was nicht hinausgeht.
const Geschwaerzt = "«geschwärzt»"

// nachweiszeile trifft eine Kopfzeile mit einem Nachweis — samt allem, was
// dahinter steht.
//
// **Das `Bearer` gehört in die Vorsilbe und nicht in den Wert.** Der erste
// Anlauf las `Authorization: Bearer …` als „Schlüsselwort plus ein Wort" und
// schwärzte damit **das Wort `Bearer`** — der Nachweis dahinter blieb stehen.
// Gefangen hat es der Lauf, nicht das Lesen.
var nachweiszeile = regexp.MustCompile(
	`(?i)(authorization\s*[:=]\s*(?:bearer\s+)?|bearer\s+)[^\s"']+`)

// postadresse trifft eine E-Mail-Adresse — OPS04
// (#2350):
// sie ist eine persönliche Angabe, und ein Fehler aus dem Unterbau, der ein
// Konto nennt, trüge sie im `grund`. **Den drei künstlichen Fällen der
// Abnahme fiel sie auf**; Token und Schlüssel fing die Zufallsfolge, sie nicht.
// **Einen Namen in freiem Text fängt keine Gestalt** — dort bleibt die
// Auswahl die Absicherung: das Paket liest die Nutzer nicht.
var postadresse = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)

// zufallsfolge trifft eine lange Folge aus dem Alphabet, das Marken, Token,
// Prüfsummen und Base32-Geheimnisse benutzen.
//
// **Zweiunddreißig Zeichen**, und sie muss Buchstaben **und** Ziffern
// enthalten: ein deutscher Satz erreicht die Länge nicht ohne Leerzeichen, ein
// Dateiname selten, ein Token immer. Die Prüfung auf beides hält Wörter wie
// `Bestandsverschluesselung` heraus.
var zufallsfolge = regexp.MustCompile(`[A-Za-z0-9+/=_-]{32,}`)

// hatZifferUndBuchstabe sagt, ob beides vorkommt.
func hatZifferUndBuchstabe(text string) bool {
	ziffer, buchstabe := false, false

	for _, zeichen := range text {
		switch {
		case zeichen >= '0' && zeichen <= '9':
			ziffer = true
		case zeichen >= 'a' && zeichen <= 'z' || zeichen >= 'A' && zeichen <= 'Z':
			buchstabe = true
		}
	}

	return ziffer && buchstabe
}

// istHexfolge sagt, ob ein Treffer nur aus Hexziffern besteht.
//
// **Eine Marke ist eine Hexfolge, keine Mischung** (SEC02, #2355): 64 Zeichen
// aus `0`–`9` und `a`–`f`, und eine davon kann aus Ziffern allein bestehen
// oder aus Buchstaben allein. Die Prüfung auf beides ließ sie dann durch.
// Kein Wort besteht aus zweiunddreißig Buchstaben zwischen `a` und `f`, und
// keine Zahl, die ein Mensch lesen soll, aus zweiunddreißig Ziffern.
func istHexfolge(text string) bool {
	for _, zeichen := range text {
		ziffer := zeichen >= '0' && zeichen <= '9'
		buchstabe := zeichen >= 'a' && zeichen <= 'f' || zeichen >= 'A' && zeichen <= 'F'

		if !ziffer && !buchstabe {
			return false
		}
	}

	return true
}

// redigiere nimmt aus einem Text, was nicht hinausgehen soll.
//
// **Nach [schwaerze] und nicht statt dessen**: der Benutzerpfad wird zur
// Tilde, damit die Datei- und Verzeichnisnamen lesbar bleiben — sie sind die
// Auskunft, um derentwillen es die Diagnose gibt.
func redigiere(text string) string {
	gekuerzt := nachweiszeile.ReplaceAllString(schwaerze(text), "${1}"+Geschwaerzt)
	gekuerzt = postadresse.ReplaceAllString(gekuerzt, Geschwaerzt)

	return zufallsfolge.ReplaceAllStringFunc(gekuerzt, func(treffer string) string {
		if !hatZifferUndBuchstabe(treffer) && !istHexfolge(treffer) {
			return treffer
		}

		return Geschwaerzt
	})
}
