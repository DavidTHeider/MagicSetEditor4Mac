package app

// **Die Sätze der beiden Dialogsprachen** — aus `menuetexte.go` herausgelöst,
// als die Datei mit der verdeckten Diagnose (Werkstück O9, #1433) über die
// Zeilengrenze aus §2 wuchs.
//
// **Der Schnitt folgt der Frage**: dort — heute `dialogtexte.go` — steht,
// *was* ein Dialogtext ist und wie man ihn bekommt; hier stehen die Sätze
// selbst.

var dialogkataloge = map[string]Dialogtexte{
	"de": {
		BildOeffnen: "Bild öffnen",
		BildFilter:  "Bilder (*.png, *.jpg, *.gif)",

		PaketdateiEinfuegen: "Datei ins Paket legen",
		PaketdateiFilter:    "Bilder und Schriften (*.png, *.jpg, *.gif, *.ttf, *.otf)",

		PaketWeitergeben:  "Paket weitergeben als",
		InstallerEinlesen: "Installer einlesen",
		InstallerFilter:   "MSE-Installer (*.mse-installer)",

		TischabzugSpeichern: "Tischbild sichern unter",
		TischabzugFilter:    "PNG-Bilder (*.png)",
		TischabzugVorschlag: "tisch.png",

		TischabzugPDFSpeichern: "Tischbild als PDF sichern unter",
		TischabzugPDFFilter:    "PDF-Dateien (*.pdf)",
		TischabzugPDFVorschlag: "tisch.pdf",

		PartieSpeichern: "Partie ausführen nach",
		PartieOeffnen:   "Partie einlesen",
		PartieFilter:    "MSE4Mac-Partien (*.msepartie)",
		PartieVorschlag: "partie.msepartie",

		TurnierSpeichern: "Turnier ausgeben nach",
		TurnierOeffnen:   "Turnier einlesen",
		TurnierFilter:    "MSE4Mac-Turniere (*.mseturnier)",
		TurnierVorschlag: "turnier.mseturnier",
		BerichtSpeichern: "Turnierbericht sichern unter",
		BerichtFilter:    "Markdown (*.md)",
		BerichtVorschlag: "turnierbericht.md",

		SymbolAusgeben:      "Symbol ausgeben nach",
		SymbolEinlesen:      "Symbol in die Bibliothek einlesen",
		SymbolFilter:        "MSE-Symbole (*.mse-symbol)",
		SymbolVorschlag:     "symbol.mse-symbol",
		SprachpaketAusgeben: "Übersetzung ausgeben nach",
		SprachpaketEinlesen: "Übersetzung einlesen",
		SprachpaketFilter:   "MSE4Mac-Übersetzungen (*.json)",

		SetOeffnen:   "Set öffnen",
		SetFilter:    "MSE-Sets (*.mse-set)",
		SetSpeichern: "Set sichern unter",
		SetVorschlag: "set.mse-set",

		UmschlagSchnueren: "Set als Umschlag ablegen",
		UmschlagOeffnen:   "Umschlag ansehen",
		UmschlagFilter:    "MSE4Mac-Umschlag (*.msepackage)",
		UmschlagVorschlag: "umschlag.msepackage",
		Uebernommen:       "übernommen",

		LeiheAblegen:   "Umschlag der Leihe ablegen",
		LeiheVorschlag: "leihe.msepackage",

		BestandSpeichern: "Bestand sichern",
		BestandFilter:    "MSE4Mac-Sicherung (*.msesicherung)",
		BestandVorschlag: "mse-sicherung.msesicherung",
		BestandOeffnen:   "Sicherung zum Zurückspielen wählen",
		KontakteOeffnen:  "Kontakte als vCard wählen",
		KontakteFilter:   "vCard (*.vcf)",

		DiagnoseSpeichern: "Diagnose sichern",
		DiagnoseFilter:    "Textdatei (*.txt)",
		DiagnoseVorschlag: "mse-diagnose.txt",

		DiagnoseBestand:      "Bestand",
		DiagnoseProtokollort: "Protokoll",
		DiagnoseKanaele:      "Kanäle",
		DiagnoseKanalDatei:   "Datei",
		DiagnoseKanalSystem:  "Unified Log",
		DiagnoseKanalAusgabe: "Standardfehlerausgabe",
		DiagnoseKanalAn:      "an",
		DiagnoseKanalAus:     "aus",
		DiagnoseKanalFehlt:   "fehlt",
		DiagnoseStufe:        "Stufe",
		DiagnoseAusfuehrlich: "Das ausführliche Protokoll ist an — es hält höchstens 1 MB " +
			"je Datei und eine Vorfassung. Stelle es unter Einstellungen › Diagnose zurück, " +
			"wenn der Fehler gefunden ist.",
		DiagnosepaketSpeichern: "Diagnosepaket sichern",
		DiagnosepaketVorschlag: "diagnose.msediagnose",
		DiagnosepaketFilter:    "Diagnosepaket (*.msediagnose)",
		DiagnoseTeilUnbekannt: "Dieser Teil ist unbekannt — die Anwendung " +
			"kennt seinen Namen, aber nicht seinen Inhalt.",
		DiagnoseProtokollVerdeckt: "Die Protokollzeilen stehen hier nicht: " +
			"sie gehören diesem Prozess und nicht diesem Profil. Wer sie " +
			"braucht, öffnet sein eigenes Profil oder liest die Datei am " +
			"oben genannten Ort.",
		DiagnoseSchema:          "Schema",
		DiagnoseWaisen:          "Verwaiste Zeilen",
		DiagnoseBau:             "Bau",
		DiagnoseVeraendert:      "Arbeitsbaum verändert",
		DiagnoseBindungsvertrag: "Bindungsvertrag",
		DiagnoseMigrationen:     "%d Migrationen, zuletzt %s",
		DiagnoseKeineSchema:     "keine Migration verbucht",
		DiagnoseNichtLesbar:     "nicht zu lesen",
		DiagnosePakete:          "Pakete",
		DiagnoseWeitere:         "… und %d weitere",
		DiagnoseKeinBestand:     "kein Bestand offen",
		DiagnoseKeinProtokoll:   "Kein Protokoll",
		DiagnoseKeineVerwaltung: "keine Paketverwaltung eingerichtet",

		MTGJSONDatei:       "MTG JSON öffnen",
		MTGJSONFilter:      "MTG JSON (*.json)",
		MTGJSONVerzeichnis: "Verzeichnis mit MTG JSON öffnen",
		MTGJSONSchreiben:   "Als MTG JSON sichern",
		MTGJSONVorschlag:   "set.json",

		PersonendatenSichern: "Meine Daten sichern",
		JSONFilter:           "JSON (*.json)",

		EinstellungenSichern:   "Einstellungen sichern",
		EinstellungenOeffnen:   "Einstellungen einlesen",
		EinstellungenVorschlag: "einstellungen.json",

		VorlageSchreiben: "Über eine Exportvorlage sichern",
		VorlageFilter:    "Ausgabe der Vorlage",

		Bildordner:     "Verzeichnis für die Bilder wählen",
		BogenSchreiben: "Druckbogen sichern",
		BogenVorschlag: "druckbogen.pdf",
		PDFFilter:      "PDF (*.pdf)",

		PaketWaehlen: "Vorlagenpaket wählen",
		MSE2Waehlen:  "Verzeichnis der MSE2-Vorlagen wählen",

		Planordner: "Ordner für die wöchentliche Sicherung wählen",
	},
	"en": {
		BildOeffnen: "Open image",
		BildFilter:  "Images (*.png, *.jpg, *.gif)",

		PaketdateiEinfuegen: "Add file to package",
		PaketdateiFilter:    "Images and fonts (*.png, *.jpg, *.gif, *.ttf, *.otf)",

		PaketWeitergeben:  "Share package as",
		InstallerEinlesen: "Read installer",
		InstallerFilter:   "MSE installers (*.mse-installer)",

		TischabzugSpeichern: "Save table image as",
		TischabzugFilter:    "PNG images (*.png)",
		TischabzugVorschlag: "table.png",

		TischabzugPDFSpeichern: "Save table image as PDF",
		TischabzugPDFFilter:    "PDF files (*.pdf)",
		TischabzugPDFVorschlag: "table.pdf",

		PartieSpeichern: "Export game to",
		PartieOeffnen:   "Import game",
		PartieFilter:    "MSE4Mac games (*.msepartie)",
		PartieVorschlag: "game.msepartie",

		TurnierSpeichern: "Export tournament to",
		TurnierOeffnen:   "Import tournament",
		TurnierFilter:    "MSE4Mac tournaments (*.mseturnier)",
		TurnierVorschlag: "tournament.mseturnier",
		BerichtSpeichern: "Save tournament report as",
		BerichtFilter:    "Markdown (*.md)",
		BerichtVorschlag: "tournament-report.md",

		SymbolAusgeben:      "Export symbol to",
		SymbolEinlesen:      "Import symbol into the library",
		SymbolFilter:        "MSE symbols (*.mse-symbol)",
		SymbolVorschlag:     "symbol.mse-symbol",
		SprachpaketAusgeben: "Export translation to",
		SprachpaketEinlesen: "Import translation",
		SprachpaketFilter:   "MSE4Mac translations (*.json)",

		SetOeffnen:   "Open set",
		SetFilter:    "MSE sets (*.mse-set)",
		SetSpeichern: "Save set as",
		SetVorschlag: "set.mse-set",

		UmschlagSchnueren: "Save set as an envelope",
		UmschlagOeffnen:   "Inspect envelope",
		UmschlagFilter:    "MSE4Mac envelope (*.msepackage)",
		UmschlagVorschlag: "envelope.msepackage",
		Uebernommen:       "imported",

		LeiheAblegen:   "Save the loan envelope",
		LeiheVorschlag: "loan.msepackage",

		BestandSpeichern: "Back up database",
		BestandFilter:    "MSE4Mac backup (*.msesicherung)",
		BestandVorschlag: "mse-backup.msesicherung",
		BestandOeffnen:   "Choose a backup to restore",
		KontakteOeffnen:  "Choose contacts as vCard",
		KontakteFilter:   "vCard (*.vcf)",

		DiagnoseSpeichern: "Save diagnostics",
		DiagnoseFilter:    "Text file (*.txt)",
		DiagnoseVorschlag: "mse-diagnostics.txt",

		DiagnoseBestand:      "Database",
		DiagnoseProtokollort: "Log",
		DiagnoseKanaele:      "Channels",
		DiagnoseKanalDatei:   "file",
		DiagnoseKanalSystem:  "Unified Log",
		DiagnoseKanalAusgabe: "standard error",
		DiagnoseKanalAn:      "on",
		DiagnoseKanalAus:     "off",
		DiagnoseKanalFehlt:   "missing",
		DiagnoseStufe:        "level",
		DiagnoseAusfuehrlich: "The detailed log is on — it keeps at most 1 MB per file " +
			"and one earlier file. Set it back under Settings › Diagnostics once the " +
			"problem is found.",
		DiagnosepaketSpeichern: "Save diagnostic bundle",
		DiagnosepaketVorschlag: "diagnostics.msediagnose",
		DiagnosepaketFilter:    "Diagnostic bundle (*.msediagnose)",
		DiagnoseTeilUnbekannt: "This part is unknown — the application knows " +
			"its name but not its content.",
		DiagnoseProtokollVerdeckt: "The log lines are not shown here: they " +
			"belong to this process, not to this profile. Open your own " +
			"profile, or read the file at the location given above.",
		DiagnoseSchema:          "Schema",
		DiagnoseWaisen:          "Orphaned rows",
		DiagnoseBau:             "Build",
		DiagnoseVeraendert:      "working tree modified",
		DiagnoseBindungsvertrag: "Binding contract",
		DiagnoseMigrationen:     "%d migrations, latest %s",
		DiagnoseKeineSchema:     "no migration recorded",
		DiagnoseNichtLesbar:     "cannot be read",
		DiagnosePakete:          "Packages",
		DiagnoseWeitere:         "… and %d more",
		DiagnoseKeinBestand:     "no database open",
		DiagnoseKeinProtokoll:   "No log",
		DiagnoseKeineVerwaltung: "no package management set up",

		MTGJSONDatei:       "Open MTG JSON",
		MTGJSONFilter:      "MTG JSON (*.json)",
		MTGJSONVerzeichnis: "Open a directory of MTG JSON",
		MTGJSONSchreiben:   "Save as MTG JSON",
		MTGJSONVorschlag:   "set.json",

		PersonendatenSichern: "Save my data",
		JSONFilter:           "JSON (*.json)",

		EinstellungenSichern:   "Save settings",
		EinstellungenOeffnen:   "Import settings",
		EinstellungenVorschlag: "einstellungen.json",

		VorlageSchreiben: "Export via template",
		VorlageFilter:    "Template output",

		Bildordner:     "Choose a directory for the images",
		BogenSchreiben: "Save print sheet",
		BogenVorschlag: "printsheet.pdf",
		PDFFilter:      "PDF (*.pdf)",

		PaketWaehlen: "Choose a template package",
		MSE2Waehlen:  "Choose the directory of the MSE2 templates",

		Planordner: "Choose a folder for the weekly backup",
	},
	"es": dialogtexteEs,
	"fr": dialogtexteFr,
	"it": dialogtexteIt,
	"ja": dialogtexteJa,
	"ru": dialogtexteRu,
	"zh": dialogtexteZh,
}
