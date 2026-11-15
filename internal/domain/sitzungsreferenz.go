package domain

// **Der sitzungsgebundene Name einer Bestandszeile** — Werkstück `T3` von
// Phase 5.6 (#1468).
//
// # Wozu es ihn gibt
//
// **Damit ein Angebot auf eine Zeile zeigen kann, ohne sie zu benennen.** Ein
// Gegenüber muss sagen können „das Stück aus deiner Liste"; es darf dabei
// nicht erfahren, welche Zeile des Bestands das ist. Das Issue sagt es
// wörtlich: eine **opake, sitzungsgebundene Referenz statt der lokalen
// Datenbank-ID**.
//
// # Warum eine fortlaufende Nummer nicht genügt
//
// **Weil sie über zwei Sichten hinweg etwas sagt.** Wer heute „Nummer 3" und
// morgen wieder „Nummer 3" hört, weiß nicht nur, dass es dieselbe Karte ist —
// er weiß auch, dass die Liste dazwischen nicht kürzer geworden ist. Über
// viele Sichten hinweg ergibt das den Umfang und die Bewegung eines Bestands,
// den niemand freigegeben hat.
//
// # Warum ein HMAC und keine Tabelle
//
// **Weil eine Tabelle irgendwo liegen müsste.** Sie bräuchte einen Ort, ein
// Aufräumen und eine Antwort auf die Frage, was nach einem Neustart aus ihr
// wird — und die falsche Antwort wäre ein Name, der auf eine fremde Zeile
// zeigt. Nachrechnen kennt diesen halben Stand nicht.
//
// **Die Sitzungskennung ist der Schlüssel.** Sie ist innerhalb eines
// Gesprächs dieselbe und danach nie wieder; damit ist der Name genau so lange
// gültig wie das Gespräch.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
)

// referenzlaenge sind die Zeichen, die ein Name trägt.
//
// **Achtzehn Zeichen Base64 sind hundertacht Bit** — genug, dass niemand einen
// zweiten Namen errät, und kurz genug, dass er in eine Zeile passt. Ein voller
// SHA-256 wäre dreiundvierzig Zeichen und sagte nichts mehr.
const referenzlaenge = 18

// Sitzungsreferenz gibt den Namen, unter dem eine Zeile in einer Sitzung
// erscheint.
func Sitzungsreferenz(sitzung string, exemplarID int64) string {
	unterschrift := hmac.New(sha256.New, []byte(sitzung))

	// **Der Fehler kann hier nicht auftreten**: `hash.Hash.Write` gibt nach
	// seiner eigenen Zusage nie einen (§3) — ein `if` darum wäre ein Zweig,
	// den kein Lauf betritt.
	_, _ = unterschrift.Write([]byte(strconv.FormatInt(exemplarID, 10)))

	return base64.RawURLEncoding.EncodeToString(
		unterschrift.Sum(nil))[:referenzlaenge]
}

// ZeileZuReferenz sucht die Zeile, die in dieser Sitzung so heißt.
//
// **Durch Nachrechnen und nicht durch Nachschlagen.** Die Liste liegt beim
// Aufrufer ohnehin vor — sie ist die Sammlung, aus der die Sicht entstand.
//
// **Ein Vergleich in fester Zeit**: der Name kommt von außen, und ein
// Vergleich, der beim ersten verschiedenen Zeichen abbricht, sagt etwas
// darüber, wie weit jemand richtig geraten hat.
func ZeileZuReferenz(
	sitzung, referenz string, exemplare []*Exemplar,
) (Exemplar, bool) {
	for _, exemplar := range exemplare {
		name := Sitzungsreferenz(sitzung, exemplar.ID)
		if hmac.Equal([]byte(name), []byte(referenz)) {
			return *exemplar, true
		}
	}

	return Exemplar{}, false
}
