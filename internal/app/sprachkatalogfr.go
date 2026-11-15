package app

// **Die Kataloge für Französisch** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteFr = Menuetexte{
	Ablage:         "Fichier",
	Neu:            "Nouveau",
	Oeffnen:        "Ouvrir…",
	Sichern:        "Enregistrer",
	SichernUnter:   "Enregistrer sous…",
	Schliessen:     "Fermer",
	Drucken:        "Imprimer…",
	BestandSichern: "Sauvegarder la base de données…",
	Einstellungen:  "Réglages…",
	Zuletzt:        "Ouvrir récent",
	Hilfe:          "Aide",
	HilfeOeffnen:   "Aide de MagicSetEditor4Mac",

	Set:              "Set",
	Rueckgaengig:     "Annuler",
	AllesAuswaehlen:  "Cocher toutes les cartes",
	Wiederherstellen: "Rétablir",
	Suchen:           "Rechercher…",
	Ersetzen:         "Remplacer…",

	Ausgeben:      "Exporter",
	Exportvorlage: "Avec un modèle d’exportation…",
	Bilder:        "En tant qu’images…",
	Druckbogen:    "Planche d’impression…",

	Ueber: "À propos de MagicSetEditor4Mac",

	Darstellung:            "Présentation",
	SeitenleisteEinblenden: "Afficher la barre latérale",
	SeitenleisteAusblenden: "Masquer la barre latérale",
	NavigatorEinblenden:    "Afficher le panneau de navigation",
	NavigatorAusblenden:    "Masquer le panneau de navigation",
	InspectorEinblenden:    "Afficher l’inspecteur",
	InspectorAusblenden:    "Masquer l’inspecteur",
	SeiteZurueck:           "Précédent",
	SeiteVor:               "Suivant",

	DarstellungZuruecksetzen: "Réinitialiser l’apparence",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "Services",
	Ausblenden:       "Masquer MagicSetEditor4Mac",
	AndereAusblenden: "Masquer les autres",
	AlleEinblenden:   "Tout afficher",
	Beenden:          "Quitter MagicSetEditor4Mac",

	Bearbeiten:      "Édition",
	Widerrufen:      "Annuler",
	Wiederholen:     "Rétablir",
	Ausschneiden:    "Couper",
	Kopieren:        "Copier",
	Einsetzen:       "Coller",
	Loeschen:        "Supprimer",
	AllesMarkieren:  "Tout sélectionner",
	Sprachausgabe:   "Parole",
	SprechenStarten: "Commencer à parler",
	SprechenStoppen: "Arrêter de parler",

	Fenster:       "Fenêtre",
	Minimieren:    "Réduire",
	Zoomen:        "Zoom",
	AlleNachVorne: "Tout ramener au premier plan",
}

var dialogtexteFr = Dialogtexte{
	BildOeffnen: "Ouvrir une image",
	BildFilter:  "Images (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "Ajouter un fichier au paquet",
	PaketdateiFilter:    "Images et polices (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "Partager le paquet sous",
	InstallerEinlesen: "Importer l’installeur",
	InstallerFilter:   "Installeurs MSE (*.mse-installer)",

	TischabzugSpeichern: "Enregistrer l’image de la table sous",
	TischabzugFilter:    "Images PNG (*.png)",
	TischabzugVorschlag: "table.png",

	TischabzugPDFSpeichern: "Enregistrer l’image de la table au format PDF",
	TischabzugPDFFilter:    "Fichiers PDF (*.pdf)",
	TischabzugPDFVorschlag: "table.pdf",

	PartieSpeichern: "Exporter la partie vers",
	PartieOeffnen:   "Importer une partie",
	PartieFilter:    "Parties MSE4Mac (*.msepartie)",
	PartieVorschlag: "partie.msepartie",

	TurnierSpeichern: "Exporter le tournoi vers",
	TurnierOeffnen:   "Importer un tournoi",
	TurnierFilter:    "Tournois MSE4Mac (*.mseturnier)",
	TurnierVorschlag: "tournoi.mseturnier",
	BerichtSpeichern: "Enregistrer le rapport du tournoi sous",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "rapport-tournoi.md",

	SymbolAusgeben:      "Exporter le symbole vers",
	SymbolEinlesen:      "Importer le symbole dans la bibliothèque",
	SymbolFilter:        "Symboles MSE (*.mse-symbol)",
	SymbolVorschlag:     "symbole.mse-symbol",
	SprachpaketAusgeben: "Exporter la traduction vers",
	SprachpaketEinlesen: "Importer une traduction",
	SprachpaketFilter:   "Traductions MSE4Mac (*.json)",

	SetOeffnen:   "Ouvrir un set",
	SetFilter:    "Sets MSE (*.mse-set)",
	SetSpeichern: "Enregistrer le set sous",
	SetVorschlag: "set.mse-set",

	UmschlagSchnueren: "Enregistrer le set sous forme d’enveloppe",
	UmschlagOeffnen:   "Examiner l’enveloppe",
	UmschlagFilter:    "Enveloppe MSE4Mac (*.msepackage)",
	UmschlagVorschlag: "enveloppe.msepackage",
	Uebernommen:       "importé",

	LeiheAblegen:   "Enregistrer l’enveloppe du prêt",
	LeiheVorschlag: "pret.msepackage",

	BestandSpeichern: "Sauvegarder la base de données",
	BestandFilter:    "Sauvegarde MSE4Mac (*.msesicherung)",
	BestandVorschlag: "mse-sauvegarde.msesicherung",
	BestandOeffnen:   "Choisir une sauvegarde à restaurer",
	KontakteOeffnen:  "Choisir les contacts au format vCard",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "Enregistrer le diagnostic",
	DiagnoseFilter:    "Fichier texte (*.txt)",
	DiagnoseVorschlag: "mse-diagnostic.txt",

	DiagnoseBestand:      "Base de données",
	DiagnoseProtokollort: "Journal",
	DiagnoseKanaele:      "Canaux",
	DiagnoseKanalDatei:   "fichier",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "erreur standard",
	DiagnoseKanalAn:      "activé",
	DiagnoseKanalAus:     "désactivé",
	DiagnoseKanalFehlt:   "absent",
	DiagnoseStufe:        "niveau",
	DiagnoseAusfuehrlich: "Le journal détaillé est activé : il conserve au plus 1 Mo " +
		"par fichier et un fichier précédent. Réinitialise-le sous Réglages › " +
		"Diagnostic une fois l’erreur trouvée.",
	DiagnosepaketSpeichern: "Enregistrer le paquet de diagnostic",
	DiagnosepaketVorschlag: "diagnostic.msediagnose",
	DiagnosepaketFilter:    "Paquet de diagnostic (*.msediagnose)",
	DiagnoseTeilUnbekannt: "Cette partie est inconnue : l’application " +
		"en connaît le nom, mais pas le contenu.",
	DiagnoseProtokollVerdeckt: "Les lignes du journal ne figurent pas ici : " +
		"elles appartiennent à ce processus et non à ce profil. Qui en a " +
		"besoin ouvre son propre profil ou lit le fichier à l’emplacement " +
		"indiqué ci-dessus.",
	DiagnoseSchema:          "Schéma",
	DiagnoseWaisen:          "Lignes orphelines",
	DiagnoseBau:             "Build",
	DiagnoseVeraendert:      "arbre de travail modifié",
	DiagnoseBindungsvertrag: "Contrat de liaison",
	DiagnoseMigrationen:     "%d migrations, dernière : %s",
	DiagnoseKeineSchema:     "aucune migration enregistrée",
	DiagnoseNichtLesbar:     "illisible",
	DiagnosePakete:          "Paquets",
	DiagnoseWeitere:         "… et %d autres",
	DiagnoseKeinBestand:     "aucune base de données ouverte",
	DiagnoseKeinProtokoll:   "Aucun journal",
	DiagnoseKeineVerwaltung: "aucune gestion des paquets configurée",

	MTGJSONDatei:       "Ouvrir MTG JSON",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "Ouvrir un dossier contenant du MTG JSON",
	MTGJSONSchreiben:   "Enregistrer au format MTG JSON",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "Enregistrer mes données",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "Enregistrer les réglages",
	EinstellungenOeffnen:   "Importer les réglages",
	EinstellungenVorschlag: "reglages.json",

	VorlageSchreiben: "Exporter via un modèle",
	VorlageFilter:    "Sortie du modèle",

	Bildordner:     "Choisir un dossier pour les images",
	BogenSchreiben: "Enregistrer la planche d’impression",
	BogenVorschlag: "planche-impression.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "Choisir un paquet de modèles",
	MSE2Waehlen:  "Choisir le dossier des modèles MSE2",

	Planordner: "Choisir un dossier pour la sauvegarde hebdomadaire",
}

var berichttexteFr = domain.Berichttexte{
	Spiel: "Jeu", Format: "Format", Stand: "État", Abgeschlossen: "terminé le",
	Laufend: "en cours", Teilnehmende: "Participants", Ausgestiegen: "a abandonné", Runden: "Rondes",
	Rangliste: "Classement", Platz: "Place", Name: "Nom", Punkte: "Points", Bilanz: "Bilan",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "Ronde", Tisch: "Table", SeiteA: "Joueur A", SeiteB: "Joueur B", Ergebnis: "Résultat",
	Bye: "Bye", Offen: "pas encore jouée",
}
