package app

// **Die Kataloge für Italienisch** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteIt = Menuetexte{
	Ablage:         "File",
	Neu:            "Nuovo",
	Oeffnen:        "Apri…",
	Sichern:        "Salva",
	SichernUnter:   "Salva con nome…",
	Schliessen:     "Chiudi",
	Drucken:        "Stampa…",
	BestandSichern: "Esegui backup del database…",
	Einstellungen:  "Impostazioni…",
	Zuletzt:        "Apri recenti",
	Hilfe:          "Aiuto",
	HilfeOeffnen:   "Aiuto di MagicSetEditor4Mac",

	Set:              "Set",
	Rueckgaengig:     "Annulla",
	AllesAuswaehlen:  "Spunta tutte le carte",
	Wiederherstellen: "Ripeti",
	Suchen:           "Trova…",
	Ersetzen:         "Sostituisci…",

	Ausgeben:      "Esporta",
	Exportvorlage: "Con un modello di esportazione…",
	Bilder:        "Come immagini…",
	Druckbogen:    "Foglio di stampa…",

	Ueber: "Informazioni su MagicSetEditor4Mac",

	Darstellung:            "Vista",
	SeitenleisteEinblenden: "Mostra la barra laterale",
	SeitenleisteAusblenden: "Nascondi la barra laterale",
	NavigatorEinblenden:    "Mostra il navigatore",
	NavigatorAusblenden:    "Nascondi il navigatore",
	InspectorEinblenden:    "Mostra l’ispettore",
	InspectorAusblenden:    "Nascondi l’ispettore",
	SeiteZurueck:           "Indietro",
	SeiteVor:               "Avanti",

	DarstellungZuruecksetzen: "Ripristina aspetto",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "Servizi",
	Ausblenden:       "Nascondi MagicSetEditor4Mac",
	AndereAusblenden: "Nascondi altre",
	AlleEinblenden:   "Mostra tutto",
	Beenden:          "Esci da MagicSetEditor4Mac",

	Bearbeiten:      "Modifica",
	Widerrufen:      "Annulla",
	Wiederholen:     "Ripeti",
	Ausschneiden:    "Taglia",
	Kopieren:        "Copia",
	Einsetzen:       "Incolla",
	Loeschen:        "Elimina",
	AllesMarkieren:  "Seleziona tutto",
	Sprachausgabe:   "Voce",
	SprechenStarten: "Inizia lettura",
	SprechenStoppen: "Interrompi lettura",

	Fenster:       "Finestra",
	Minimieren:    "Riduci a icona",
	Zoomen:        "Zoom",
	AlleNachVorne: "Porta tutto in primo piano",
}

var dialogtexteIt = Dialogtexte{
	BildOeffnen: "Apri immagine",
	BildFilter:  "Immagini (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "Aggiungi file al pacchetto",
	PaketdateiFilter:    "Immagini e font (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "Condividi pacchetto come",
	InstallerEinlesen: "Importa installer",
	InstallerFilter:   "Installer MSE (*.mse-installer)",

	TischabzugSpeichern: "Salva immagine del tavolo come",
	TischabzugFilter:    "Immagini PNG (*.png)",
	TischabzugVorschlag: "tavolo.png",

	TischabzugPDFSpeichern: "Salva immagine del tavolo come PDF",
	TischabzugPDFFilter:    "File PDF (*.pdf)",
	TischabzugPDFVorschlag: "tavolo.pdf",

	PartieSpeichern: "Esporta partita in",
	PartieOeffnen:   "Importa partita",
	PartieFilter:    "Partite MSE4Mac (*.msepartie)",
	PartieVorschlag: "partita.msepartie",

	TurnierSpeichern: "Esporta torneo in",
	TurnierOeffnen:   "Importa torneo",
	TurnierFilter:    "Tornei MSE4Mac (*.mseturnier)",
	TurnierVorschlag: "torneo.mseturnier",
	BerichtSpeichern: "Salva resoconto del torneo come",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "resoconto-torneo.md",

	SymbolAusgeben:      "Esporta simbolo in",
	SymbolEinlesen:      "Importa simbolo nella libreria",
	SymbolFilter:        "Simboli MSE (*.mse-symbol)",
	SymbolVorschlag:     "simbolo.mse-symbol",
	SprachpaketAusgeben: "Esporta traduzione in",
	SprachpaketEinlesen: "Importa traduzione",
	SprachpaketFilter:   "Traduzioni MSE4Mac (*.json)",

	SetOeffnen:   "Apri set",
	SetFilter:    "Set MSE (*.mse-set)",
	SetSpeichern: "Salva set come",
	SetVorschlag: "set.mse-set",

	UmschlagSchnueren: "Salva set come busta",
	UmschlagOeffnen:   "Ispeziona busta",
	UmschlagFilter:    "Busta MSE4Mac (*.msepackage)",
	UmschlagVorschlag: "busta.msepackage",
	Uebernommen:       "importato",

	LeiheAblegen:   "Salva la busta del prestito",
	LeiheVorschlag: "prestito.msepackage",

	BestandSpeichern: "Esegui backup del database",
	BestandFilter:    "Backup MSE4Mac (*.msesicherung)",
	BestandVorschlag: "mse-backup.msesicherung",
	BestandOeffnen:   "Scegli un backup da ripristinare",
	KontakteOeffnen:  "Scegli i contatti come vCard",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "Salva diagnostica",
	DiagnoseFilter:    "File di testo (*.txt)",
	DiagnoseVorschlag: "mse-diagnostica.txt",

	DiagnoseBestand:      "Database",
	DiagnoseProtokollort: "Registro",
	DiagnoseKanaele:      "Canali",
	DiagnoseKanalDatei:   "file",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "standard error",
	DiagnoseKanalAn:      "attivo",
	DiagnoseKanalAus:     "disattivato",
	DiagnoseKanalFehlt:   "mancante",
	DiagnoseStufe:        "livello",
	DiagnoseAusfuehrlich: "Il registro dettagliato è attivo: conserva al massimo 1 MB " +
		"per file e un file precedente. Ripristina l’impostazione in Impostazioni › " +
		"Diagnostica una volta trovato il problema.",
	DiagnosepaketSpeichern: "Salva pacchetto di diagnostica",
	DiagnosepaketVorschlag: "diagnostica.msediagnose",
	DiagnosepaketFilter:    "Pacchetto di diagnostica (*.msediagnose)",
	DiagnoseTeilUnbekannt: "Questa parte è sconosciuta: l’applicazione " +
		"ne conosce il nome, ma non il contenuto.",
	DiagnoseProtokollVerdeckt: "Le righe del registro non compaiono qui: " +
		"appartengono a questo processo e non a questo profilo. Chi ne ha " +
		"bisogno apre il proprio profilo o legge il file nel percorso " +
		"indicato sopra.",
	DiagnoseSchema:          "Schema",
	DiagnoseWaisen:          "Righe orfane",
	DiagnoseBau:             "Build",
	DiagnoseVeraendert:      "albero di lavoro modificato",
	DiagnoseBindungsvertrag: "Contratto dei binding",
	DiagnoseMigrationen:     "%d migrazioni, ultima %s",
	DiagnoseKeineSchema:     "nessuna migrazione registrata",
	DiagnoseNichtLesbar:     "illeggibile",
	DiagnosePakete:          "Pacchetti",
	DiagnoseWeitere:         "… e altri %d",
	DiagnoseKeinBestand:     "nessun database aperto",
	DiagnoseKeinProtokoll:   "Nessun registro",
	DiagnoseKeineVerwaltung: "nessuna gestione dei pacchetti configurata",

	MTGJSONDatei:       "Apri MTG JSON",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "Apri una cartella con MTG JSON",
	MTGJSONSchreiben:   "Salva come MTG JSON",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "Salva i miei dati",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "Salva impostazioni",
	EinstellungenOeffnen:   "Importa impostazioni",
	EinstellungenVorschlag: "impostazioni.json",

	VorlageSchreiben: "Esporta tramite modello",
	VorlageFilter:    "Output del modello",

	Bildordner:     "Scegli una cartella per le immagini",
	BogenSchreiben: "Salva foglio di stampa",
	BogenVorschlag: "foglio-di-stampa.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "Scegli un pacchetto di modelli",
	MSE2Waehlen:  "Scegli la cartella dei modelli MSE2",

	Planordner: "Scegli una cartella per il backup settimanale",
}

var berichttexteIt = domain.Berichttexte{
	Spiel: "Gioco", Format: "Formato", Stand: "Stato", Abgeschlossen: "concluso il",
	Laufend: "in corso", Teilnehmende: "Partecipanti", Ausgestiegen: "ritirato", Runden: "Turni",
	Rangliste: "Classifica", Platz: "Posizione", Name: "Nome", Punkte: "Punti", Bilanz: "Bilancio",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "Turno", Tisch: "Tavolo", SeiteA: "Giocatore A", SeiteB: "Giocatore B", Ergebnis: "Risultato",
	Bye: "Bye", Offen: "non ancora giocato",
}
