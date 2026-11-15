package app

// **Was in einer Protokollzeile stehen darf** —
// #857.
//
// # Warum eine Liste und nicht ein Filter über den Inhalt
//
// Ein Filter müsste einen Kartennamen von einem Dateinamen unterscheiden, und
// das kann er nicht: beides sind Zeichenketten, die ein Mensch getippt hat.
// Was sich entscheiden lässt, ist die **Frage**, die ein Merkmal beantwortet —
// und die steht in seinem Namen.
//
// # Warum sie fehlschließt
//
// Ein Merkmal mit unbekanntem Namen kommt **ohne seinen Wert** ins Protokoll.
// Damit kostet eine vergessene Anmeldung eine unlesbare Zeile und nicht einen
// ausgeplauderten Kartentext. Der umgekehrte Nullwert — „unbekannt heißt
// erlaubt" — wäre eine Zusage, die beim ersten neuen Aufruf still bricht.

// Nichterklaert steht anstelle eines Wertes, dessen Merkmal niemand angemeldet
// hat.
const Nichterklaert = "«nicht erklärt»"

// erlaubteMerkmale sind die Namen, die eine Protokollzeile führen darf —
// **jeder mit der Frage, die er beantwortet**.
//
// Der Text daneben ist kein Kommentar, sondern die Zusage: wer ein Merkmal
// anmeldet, sagt, was darunter stehen wird. Ein `ort`, unter dem ein
// Kartenname steht, ist ein Fehler des Aufrufers — und den fängt diese Liste
// nicht (siehe [geschwaerzt]).
var erlaubteMerkmale = map[string]string{
	"ort":   "ein Pfad im Dateisystem — das Benutzerverzeichnis wird geschwärzt",
	"grund": "die Meldung eines Fehlers aus dem Unterbau",
	"panik": "der Wert einer gefangenen Panik — **ganz, denn er bleibt am Rechner**; " +
		"was davon in die Diagnose und das Diagnosepaket geht, hat die Gestalt, die das " +
		"Laufzeitsystem schreibt (#2399, Vermerk 0193)",
	"datei":  "der Name einer Datei ohne ihr Verzeichnis",
	"stelle": "eine Stelle im Quelltext oder in einem Format",
	"code":   "ein Fehlercode dieser Anwendung",
	"art":    "die Art eines Befundes oder eines Vorgangs",
	"anzahl": "eine Zählung",
	"dienst": "der Name eines Dienstes, den das Fenster gestartet hat — `mse-rest` oder `mse-mcp`",
	"zeile": "eine Protokollzeile eines solchen Dienstes, **dort schon geschwärzt** und mit seiner " +
		"Vorgangskennung (Nachlauf über Phase 6.4, N3)",
	"anweisung": "die Anweisung einer Schutzrichtlinie, gegen die die Seite " +
		"verstoßen hat — `script-src-elem` und ihresgleichen (SEC-17, #2164)",
	"abgewiesen": "was die Schutzrichtlinie abgewiesen hat — `inline`, `eval`, " +
		"`data`, `blob` oder der **Ursprung** einer Adresse; ihr Pfad kommt nicht " +
		"hierher, die Hülle schneidet ihn ab (SEC-17)",
	"stand": "eine Fassung, eine Migrationsstufe, eine Version",
	"dauer": "eine Zeitspanne",
	"gegenueber": "der **Fingerabdruck** eines Gegenübers und nicht seine " +
		"Kennung (#1634) — er ist die Form, die ein Mensch vergleicht, und " +
		"er trägt den Schlüssel nicht",
	"ergebnis": "was eine Zustellung bewirkt hat — Vorgang und Revision, " +
		"nicht ihr Inhalt (#1634)",
	"kennung": "die Vorgangskennung dieses Rufs — acht Hexziffern, gewürfelt " +
		"an der Grenze, an der er ankam (#1433). Sie steht in jeder Zeile " +
		"dieses Vorgangs und in seiner Antwort; **[geschwaerzt] setzt sie**, " +
		"kein Aufrufer",
	"posten": "die Kennung eines Kostenpostens — `ki-<karte>-<zeit>` oder " +
		"`probe/<anbieter>/<modell>` (#1542). Sie steht da, damit jemand den " +
		"Posten findet, der offen geblieben ist; ein Kartenname oder ein " +
		"Prompt steht darin nicht",
	"partie": "die Kennung einer Partie — 32 Hexziffern, gewürfelt beim " +
		"Anlegen (G5, #1462). Sie steht da, damit sich eine Gastsitzung " +
		"einem Tisch zuordnen lässt; **ein Sitzname steht nicht daneben**, " +
		"denn den tippt ein Mensch, und was er hineinschreibt, entscheidet " +
		"er — ein Protokoll ist kein Ort für einen selbstgewählten Namen",
	"fehlerkennung": "die Kennung eines Fehlers, den das Fenster beim " +
		"Zeichnen gefangen hat — zweimal vier Hexziffern, gewürfelt im Fenster " +
		"und dort angezeigt, damit sich die Zeile im Diagnosepaket finden lässt " +
		"(F01, #2340). **Nicht `kennung`**: die ist die Vorgangskennung und " +
		"setzt [geschwaerzt]",
	"bereich": "der Schlüssel eines Arbeitsbereichs im Fenster — `sets`, " +
		"`sammlung` und ihresgleichen, vorher auf Kleinbuchstaben und Ziffern " +
		"geprüft (F01, #2340)",
	"vorgang": "der Name dessen, was gerade lief — oder der Satz, den die " +
		"Oberfläche stattdessen bekommt, wenn `domain.MitGrund` den " +
		"technischen Grund daraus genommen hat (#1352)",
}

// merkmalErlaubt sagt, ob ein Merkmal seinen Wert behalten darf.
func merkmalErlaubt(name string) bool {
	_, angemeldet := erlaubteMerkmale[name]

	return angemeldet
}
