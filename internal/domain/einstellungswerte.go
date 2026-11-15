// **Das Vokabular der Einstellungen** — die Werte, die eine Einstellung
// annehmen kann, und die Bereiche, in denen sie steht.
//
// **Abgetrennt am 2026-09-05** (§2), als `einstellung.go` mit den drei
// Vorgaben aus Werkstück 9 von Phase 4.12
// (#1351)
// über die Zeilengrenze ging.
//
// **Der Schnitt folgt der Frage**: nebenan steht, **welche** Einstellungen es
// gibt und welche Regel für sie gilt; hier, **womit** sie gefüllt werden. Die
// Aufzählung selbst bleibt drüben — sie ist die Zeile, und eine halbe Zeile
// gibt es nicht.

package domain

// Die vier Anzeigeschriften.
//
// **Jede liegt auf dem Zielrechner** — nachgesehen am 2026-09-03 auf
// macOS 26. Das ist die Bedingung: die Anwendung arbeitet ohne Netz (§16,
// Regel 5), und eine Schrift, die manchen fehlt, ließe die Rückfallkette
// entscheiden, wie das Fenster aussieht.
//
// **Die Namen sind deutsch**, weil es eigene Begriffe dieses Programms sind
// ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)):
// „Pergament" ist keine Schrift, sondern die Anmutung, für die eine gewählt
// wurde.
const (
	// SchriftartPergament ist `Hoefler Text` — ruhig und lesbar, die Vorgabe.
	SchriftartPergament = "pergament"

	// SchriftartFantasy ist `Luminari`.
	SchriftartFantasy = "fantasy"

	// SchriftartAntike ist `Herculanum` — römische Kapitalis.
	SchriftartAntike = "antike"

	// SchriftartKalligrafie ist `Trattatello` — von Hand geschrieben.
	SchriftartKalligrafie = "kalligrafie"

	// **Drei Textschriften** — UX-09.AK-02
	// (#1681).
	// Sie liegen als Systemschriften auf jedem Mac (`ui-serif`, `ui-rounded`,
	// `system-ui`) und tragen, anders als die drei darüber, auch einen
	// Absatz; wer die ganze Oberfläche umstellt, bekommt nur sie.

	// SchriftartBuch ist `New York`.
	SchriftartBuch = "buch"

	// SchriftartRund ist `SF Pro Rounded`.
	SchriftartRund = "rund"

	// SchriftartSerifenlos ist `SF Pro`, die Systemschrift.
	SchriftartSerifenlos = "serifenlos"
)

// schriftarten ist die Reihe der sieben, wie das Fenster sie zeigt — erst
// die Textschriften, dann die Schmuckschriften.
//
// **Eine Reihe für beide Ebenen**: Profil und Mensch nennen dieselben, und
// zwei Listen liefen beim nächsten Wert auseinander.
func schriftarten() []string {
	return []string{
		SchriftartPergament, SchriftartBuch, SchriftartRund,
		SchriftartSerifenlos, SchriftartFantasy, SchriftartAntike,
		SchriftartKalligrafie,
	}
}

// Die vier Farbwelten — UX-09.AK-01
// (#1681).
// **Jede hat ein helles und ein dunkles Bild**, und jedes ist gerechnet
// (`frontend/src/themen.test.ts`, `docs/KONTRASTE.md`).
const (
	// FarbweltPergament ist, was es bisher gab — die Vorgabe.
	FarbweltPergament = "pergament"

	// FarbweltKlar ist neutral und ohne Faser.
	FarbweltKlar = "klar"

	// FarbweltSchiefer ist kühles Blaugrau.
	FarbweltSchiefer = "schiefer"

	// FarbweltMoos ist gedämpftes Grün.
	FarbweltMoos = "moos"
)

// Die zwei Reichweiten der Anzeigeschrift.
const (
	SchriftbereichUeberschriften = "ueberschriften"
	SchriftbereichGanz           = "ganz"
)

// Die vier Stellungen des Fensterbilds — **aus ist die Vorgabe**, und
// „aus" ist zugleich die Rücksetzung, die sofort erreichbar ist.
const (
	HintergrundAus       = "aus"
	HintergrundFuellen   = "fuellen"
	HintergrundEinpassen = "einpassen"
	HintergrundKacheln   = "kacheln"
)

// Die zwei Stellungen der Einführung.
//
// **„An" ist die Vorgabe**: wer die Anwendung zum ersten Mal öffnet, ist
// genau der, für den es sie gibt. Wer sie kennt, schaltet sie im Pop-up
// selbst oder auf der Einstellungsseite ab — an **beiden** Stellen, weil
// niemand die Einstellungen sucht, um etwas loszuwerden, das gerade vor ihm
// steht.
const (
	ErsteSchritteAn  = "an"
	ErsteSchritteAus = "aus"
)

// Die drei Stufen der Aktualisierung.
//
// **„Aus" ist die Vorgabe, und das ist die Zusage** — §16 Regel 6: nichts wird
// still übernommen. Ein Programm, das beim ersten Start ungefragt ins Netz
// greift, ist eines, das etwas will.
const (
	// AktualisierungAus heißt: nichts. Kein Netzverkehr, keine Frage.
	AktualisierungAus = "aus"

	// AktualisierungMelden heißt: sagen, dass eine neue Fassung vorliegt.
	// Mehr nicht.
	AktualisierungMelden = "melden"

	// AktualisierungHolen heißt: sie zusätzlich ins Downloadverzeichnis
	// laden und um die Installation bitten.
	//
	// **Noch nicht gebaut** (Werkstück 4 von #1100). Die Stufe steht hier,
	// weil die Reihenfolge der Werte die Reihenfolge im Fenster ist und ein
	// späteres Einschieben die gespeicherte Wahl verschöbe.
	AktualisierungHolen = "holen"
)

// Die vier Verhaltensweisen bei einem Namenskonflikt.
//
// **Als Konstanten und nicht als Zeichenketten am Aufrufort**: sie stehen in
// drei Schichten, und ein Tippfehler in einer davon führte auf den Zweig
// „unbekannt", der stillschweigend überschreibt.
const (
	BildkonfliktAltbehalten    = "altbehalten"
	BildkonfliktUeberschreiben = "ueberschreiben"
	BildkonfliktNummerieren    = "nummerieren"
	BildkonfliktNummerierenNeu = "nummerierenneu"
)

// Die Bereiche, in denen eine Einstellung steht.
//
// **Nicht jede Einstellung gehört auf die Einstellungsseite.** Die Wahl beim
// Namenskonflikt gehört zu dem Ausgabelauf, den man gerade startet — MSE hält
// sie ebenso im Ausgabefenster (`images_export_window.cpp`, Fassung f9d9356
// vom 2021-07-06) und nicht in den Programmeinstellungen.
//
// **Der Bereich steht hier und nicht im Fenster**: sonst entschiede die
// Oberfläche, was eine Anwendungseinstellung ist, und die Antwort stünde an
// zwei Stellen.
const (
	BereichAnwendung   = "anwendung"
	BereichBildausfuhr = "bildausfuhr"

	// BereichDienste sind die fremden Kartendienste — Phase 4.8
	// (#1321).
	//
	// **Ein eigener Bereich und nicht `anwendung`**: was hier steht, betrifft
	// nicht die Anwendung, sondern was sie **nach außen** tut, und daneben
	// gehört die Auskunft, welcher Dienst gegen sein Original gelaufen ist
	// und welcher nicht ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)).
	BereichDienste = "dienste"

	// BereichVorgaben ist, womit ein neuer Datensatz anfängt — Werkstück 9
	// (#1351), Punkte 2, 3 und 8.
	//
	// **Ein eigener Bereich, weil es eine andere Frage ist**: die
	// Anwendungseinstellungen sagen, wie das Fenster **aussieht**; diese
	// sagen, was in einem Formular schon dasteht, bevor jemand tippt.
	BereichVorgaben = "vorgaben"

	// BereichArbeitsplatz ist, wie das Fenster steht — Werkstück O3 von
	// Phase 5.2
	// (#1427).
	//
	// **Ein eigener Bereich, weil er anders reist**: was hier steht, hängt an
	// einem Bildschirm und nicht an einer Arbeitsumgebung. Der Austausch
	// nimmt ihn deshalb nur auf Wunsch mit.
	BereichArbeitsplatz = "arbeitsplatz"

	// BereichRichtlinien sind die Schranken — Werkstück O3 von Phase 5.2.
	//
	// **Ein eigener Bereich und nicht `dienste`**: jene sagen, ob dieses
	// Programm einen bestimmten Dienst fragen darf; diese stehen **über**
	// allen und lassen sich von unten nicht aufheben (`richtlinie.go`). Ein
	// Schalter neben einer Schranke sähe aus, als wäre er ihr gleichgestellt.
	BereichRichtlinien = "richtlinien"

	// BereichAnmeldung ist, was ein Anbieter braucht, um überhaupt zu
	// antworten — Nacharbeit zu Werkstück 4 von Phase 4.12 (#1351).
	//
	// **Ein eigener Bereich und nicht `dienste`**: jene sagen, ob dieses
	// Programm einen fremden Dienst **fragen** darf; dieser sagt, unter
	// welcher Kennung es sich dabei ausweist. Wer beides zusammenlegte,
	// stellte einen Schalter neben ein Textfeld, das ohne den Schalter
	// wirkungslos ist — und umgekehrt.
	BereichAnmeldung = "anmeldung"

	// BereichNetz ist, **wie** dieses Programm nach außen geht — Werkstück
	// N5 von Phase 5.4
	// (#1571).
	//
	// **Ein eigener Bereich und nicht `dienste`**: jener sagt, **welchen**
	// Dienst dieses Programm fragen darf und unter welcher Adresse; dieser
	// sagt, **auf welchem Weg** es dorthin kommt — und der gilt für jeden
	// Dienst zugleich, auch für die, die kein Mensch je einstellt. Wer beides
	// zusammenlegte, stellte eine Eigenschaft der Leitung neben eine
	// Eigenschaft der Gegenstelle.
	BereichNetz = "netz"

	// BereichSchalter sind die Bereiche und Funktionen, die sich abschalten
	// lassen — OPS-10
	// (#1908).
	//
	// **Ein eigener Bereich und nicht `anwendung`**: jener sagt, wie das
	// Programm aussieht; dieser, was es überhaupt anbietet.
	BereichSchalter = "schalter"
)
