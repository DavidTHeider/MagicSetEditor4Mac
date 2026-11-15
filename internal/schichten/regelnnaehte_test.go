package schichten

// **Die Blätter, die ein anderes Blatt kennen** — die benannten Ausnahmen
// aus §2.
//
// **Eine eigene Datei**, weil `regeln_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **wer
// wen kennt** — hier, **welche Blätter einander kennen dürfen und warum**.
//
// **Drei sind es**, und jede hat denselben Grund: das eine ist der **zweite
// Verbraucher** derselben Sache.
//
//   - `svg` kennt `render` — derselbe Zeichenplan.
//   - `fernleitung` kennt `leitung` — derselbe gesicherte Weg über einem
//     Strom, den jemand anderes gebaut hat.
//   - `austausch` kennt `sicherung` — dasselbe ZIP mit Manifest,
//     Prüfsummen, Größengrenzen und Pfadschutz.
//
// **Umgekehrt geht keine von ihnen.** Ein `render`, das von SVG wüsste, ein
// `leitung`, das von tailcat wüsste, ein `sicherung`, das von Objektarten und
// Unterschriften wüsste — das wären genau die Vermischungen, die diese Regel
// verhindert.
var naehte = map[string][]string{
	// **`austausch` kennt `sicherung` — die dritte benannte Ausnahme dieser
	// Art** (N4, #1457).
	//
	// **Derselbe Grund wie bei `fernleitung` zu `leitung` und `svg` zu
	// `render`**: es ist der **zweite Verbraucher** derselben Sache — ein ZIP
	// mit Manifest, Prüfsummen, Größengrenzen und Pfadschutz. Das noch einmal
	// zu schreiben hieße, es noch einmal zu härten, und die zweite Fassung
	// wäre die ungeprüfte.
	//
	// **Umgekehrt geht nicht**: eine Sicherung, die von Objektarten und
	// Unterschriften wüsste, wäre die Vermischung, die diese Regel
	// verhindert. `sicherung` weiß nicht, wer seine Archive schreibt.
	//
	// **`peer` steht dabei**, weil ein Umschlag unterschrieben wird und seine
	// Unterschrift geprüft: [peer.Identitaet.Signiere] und [peer.Pruefe].
	"austausch": {"domain", "peer", "sicherung"},

	// **`fernleitung` kennt `leitung` — die zweite benannte Ausnahme dieser
	// Art** (N3, #1456).
	//
	// **Derselbe Grund wie bei `svg` zu `render`**: es ist der **zweite
	// Verbraucher** desselben Weges. tailcat liefert einen `net.Conn`, und
	// alles darüber — TLS 1.3, das Festnageln auf die erwartete Kennung, die
	// Kanalbindung, die Rahmung — ist dasselbe wie im eigenen Netz.
	//
	// **Umgekehrt geht nicht**: ein `leitung`, das von tailcat wüsste, wäre
	// die Vermischung, die diese Regel verhindert. Die Kante zeigt in eine
	// Richtung, und `leitung.UeberStrom` weiß nicht, wer den Strom gebaut hat.
	//
	// **Und ohne sie gäbe es den Weg zweimal** — einmal geprüft und einmal
	// nicht. Genau das ist die Probe auf die Trennung: trägt sie auch, wenn
	// jemand anderes als ein TCP-Wähler den gesicherten Weg braucht?
	"fernleitung": {"domain", "leitung", "peer"},
}
