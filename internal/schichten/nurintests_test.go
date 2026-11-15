// **Die Kanten, die nur ein Test führen darf**
// (#1107).
//
// **Eine eigene Datei**, weil `composition_test.go` mit dem KI-Eintrag aus
// A7 (#1446)
// über die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort
// steht, **welche Datei** eine erlaubte Kante führen darf, hier, welche
// Kante **gar nicht** in den Produktivcode gehört.
package schichten

// nurInTests erlaubt eine Kante **allein aus Testdateien**.
//
// # Wofür sie da ist
//
// Ein Test darf das Erzeugnis aufmachen, das die geprüfte Schicht schreibt.
// `internal/app` schreibt seit dem 2026-08-28 keine Sicherung mehr selbst
// (#1175) — aber die Abnahme der Fassade prüft, dass am gewählten Ort wirklich
// ein Archiv steht, dass die Bilder darin liegen und dass eine Sicherung aus
// der Zukunft abgewiesen wird. Ohne diesen Eintrag ließe sich das nur noch
// behaupten.
//
// # Was sie **nicht** erlaubt
//
// **Produktivcode.** Eine `.go`-Datei ohne `_test`-Endung in `app`, die
// `internal/sicherung` importiert, bleibt ein Verstoß — und genau das ist der
// Kern von #1175. Die Gegenprobe dazu steht in `gegenproben_test.go`.
//
// **Und Fachlichkeit im Test.** Wer den Rückspielweg in eine Testdatei in
// `app` schriebe, bekäme von diesem Wächter nichts zu hören. Er bekäme es
// vorher auch nicht; was ihn hält, ist das Review (§10).
var nurInTests = map[string][]string{
	"app": {"sicherung"},

	// **`service` darf `leitung` nur in Tests kennen** — Werkstück N2 von
	// Phase 5.4
	// (#1455).
	//
	// # Warum ein Test sie braucht
	//
	// **Weil ein Handschlag zwei Seiten hat.** Seit `Oeffne` ihn führt,
	// kommt eine Sitzung nur zustande, wenn am anderen Ende wirklich jemand
	// antwortet; ein Doppel, das das nachspielte, prüfte eine Sitzung, die es
	// so nie gibt. `leitung.Schleife` verbindet zwei Verwaltungen im selben
	// Prozess — ohne Netz und ohne zweiten Rechner.
	//
	// # Warum der Produktivcode sie **nicht** kennen darf
	//
	// **Weil `service` keinen Transport kennt.** Es stellt [Waehler] und
	// [Verbindung] auf; **welcher** Transport sie erfüllt, entscheidet der
	// Composition Root. Eine Kante von hier nach `leitung` wäre die Stelle,
	// an der die Fachschicht sich einen aussucht — und N3 (tailcat,
	// [0112](../../docs/entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md))
	// wäre dann kein Blatttausch mehr.
	//
	// **Für `app` galt das bis zum 2026-09-14 auch**, und hier stand: „der
	// Composition Root **darf** sie im Produktivcode kennen, sobald N2 den
	// Weg „Adresse von Hand" baut. Bis dahin steht sie hier, und der
	// Eintrag ist dann zu verschieben statt zu ergänzen." Genau das ist
	// geschehen — `app/peeradresswaehler.go` und `app/peerhorchen.go` bauen
	// die Leitung, und `leitung` steht jetzt in [erlaubtInApp].
	//
	// **Für `service` bleibt es**, und der Unterschied ist der ganze Punkt
	// dieser Liste: der eine sucht einen Transport aus, der andere darf es
	// nicht.
	"service": {"leitung"},
}
