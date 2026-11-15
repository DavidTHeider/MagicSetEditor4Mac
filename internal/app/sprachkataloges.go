package app

// **Die Kataloge für Spanisch (Spanien)** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteEs = Menuetexte{
	Ablage:         "Archivo",
	Neu:            "Nuevo",
	Oeffnen:        "Abrir…",
	Sichern:        "Guardar",
	SichernUnter:   "Guardar como…",
	Schliessen:     "Cerrar",
	Drucken:        "Imprimir…",
	BestandSichern: "Hacer copia de seguridad de la base de datos…",
	Einstellungen:  "Ajustes…",
	Zuletzt:        "Abrir recientes",
	Hilfe:          "Ayuda",
	HilfeOeffnen:   "Ayuda de MagicSetEditor4Mac",

	Set:              "Set",
	Rueckgaengig:     "Deshacer",
	AllesAuswaehlen:  "Marcar todas las cartas",
	Wiederherstellen: "Rehacer",
	Suchen:           "Buscar…",
	Ersetzen:         "Reemplazar…",

	Ausgeben:      "Exportar",
	Exportvorlage: "Con una plantilla de exportación…",
	Bilder:        "Como imágenes…",
	Druckbogen:    "Hoja de impresión…",

	Ueber: "Acerca de MagicSetEditor4Mac",

	Darstellung:            "Visualización",
	SeitenleisteEinblenden: "Mostrar la barra lateral",
	SeitenleisteAusblenden: "Ocultar la barra lateral",
	NavigatorEinblenden:    "Mostrar el navegador",
	NavigatorAusblenden:    "Ocultar el navegador",
	InspectorEinblenden:    "Mostrar el inspector",
	InspectorAusblenden:    "Ocultar el inspector",
	SeiteZurueck:           "Atrás",
	SeiteVor:               "Adelante",

	DarstellungZuruecksetzen: "Restablecer apariencia",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "Servicios",
	Ausblenden:       "Ocultar MagicSetEditor4Mac",
	AndereAusblenden: "Ocultar otras",
	AlleEinblenden:   "Mostrar todo",
	Beenden:          "Salir de MagicSetEditor4Mac",

	Bearbeiten:      "Editar",
	Widerrufen:      "Deshacer",
	Wiederholen:     "Rehacer",
	Ausschneiden:    "Cortar",
	Kopieren:        "Copiar",
	Einsetzen:       "Pegar",
	Loeschen:        "Eliminar",
	AllesMarkieren:  "Seleccionar todo",
	Sprachausgabe:   "Locución",
	SprechenStarten: "Empezar locución",
	SprechenStoppen: "Detener locución",

	Fenster:       "Ventana",
	Minimieren:    "Minimizar",
	Zoomen:        "Zoom",
	AlleNachVorne: "Traer todo al frente",
}

var dialogtexteEs = Dialogtexte{
	BildOeffnen: "Abrir imagen",
	BildFilter:  "Imágenes (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "Añadir archivo al paquete",
	PaketdateiFilter:    "Imágenes y fuentes (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "Compartir paquete como",
	InstallerEinlesen: "Importar instalador",
	InstallerFilter:   "Instaladores MSE (*.mse-installer)",

	TischabzugSpeichern: "Guardar imagen de la mesa como",
	TischabzugFilter:    "Imágenes PNG (*.png)",
	TischabzugVorschlag: "mesa.png",

	TischabzugPDFSpeichern: "Guardar imagen de la mesa como PDF",
	TischabzugPDFFilter:    "Archivos PDF (*.pdf)",
	TischabzugPDFVorschlag: "mesa.pdf",

	PartieSpeichern: "Exportar partida a",
	PartieOeffnen:   "Importar partida",
	PartieFilter:    "Partidas de MSE4Mac (*.msepartie)",
	PartieVorschlag: "partida.msepartie",

	TurnierSpeichern: "Exportar torneo a",
	TurnierOeffnen:   "Importar torneo",
	TurnierFilter:    "Torneos de MSE4Mac (*.mseturnier)",
	TurnierVorschlag: "torneo.mseturnier",
	BerichtSpeichern: "Guardar informe del torneo como",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "informe-torneo.md",

	SymbolAusgeben:      "Exportar símbolo a",
	SymbolEinlesen:      "Importar símbolo a la biblioteca",
	SymbolFilter:        "Símbolos MSE (*.mse-symbol)",
	SymbolVorschlag:     "simbolo.mse-symbol",
	SprachpaketAusgeben: "Exportar traducción a",
	SprachpaketEinlesen: "Importar traducción",
	SprachpaketFilter:   "Traducciones de MSE4Mac (*.json)",

	SetOeffnen:   "Abrir set",
	SetFilter:    "Sets MSE (*.mse-set)",
	SetSpeichern: "Guardar set como",
	SetVorschlag: "set.mse-set",

	UmschlagSchnueren: "Guardar set como sobre",
	UmschlagOeffnen:   "Inspeccionar sobre",
	UmschlagFilter:    "Sobre de MSE4Mac (*.msepackage)",
	UmschlagVorschlag: "sobre.msepackage",
	Uebernommen:       "importado",

	LeiheAblegen:   "Guardar el sobre del préstamo",
	LeiheVorschlag: "prestamo.msepackage",

	BestandSpeichern: "Hacer copia de seguridad de la base de datos",
	BestandFilter:    "Copia de seguridad de MSE4Mac (*.msesicherung)",
	BestandVorschlag: "mse-copia-de-seguridad.msesicherung",
	BestandOeffnen:   "Elegir una copia de seguridad para restaurar",
	KontakteOeffnen:  "Elegir los contactos como vCard",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "Guardar diagnóstico",
	DiagnoseFilter:    "Archivo de texto (*.txt)",
	DiagnoseVorschlag: "mse-diagnostico.txt",

	DiagnoseBestand:      "Base de datos",
	DiagnoseProtokollort: "Registro",
	DiagnoseKanaele:      "Canales",
	DiagnoseKanalDatei:   "archivo",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "salida de error estándar",
	DiagnoseKanalAn:      "activado",
	DiagnoseKanalAus:     "desactivado",
	DiagnoseKanalFehlt:   "falta",
	DiagnoseStufe:        "nivel",
	DiagnoseAusfuehrlich: "El registro detallado está activado: conserva como máximo 1 MB " +
		"por archivo y un archivo anterior. Restablécelo en Ajustes › Diagnóstico " +
		"cuando hayas encontrado el error.",
	DiagnosepaketSpeichern: "Guardar paquete de diagnóstico",
	DiagnosepaketVorschlag: "diagnostico.msediagnose",
	DiagnosepaketFilter:    "Paquete de diagnóstico (*.msediagnose)",
	DiagnoseTeilUnbekannt: "Esta parte es desconocida: la aplicación " +
		"conoce su nombre, pero no su contenido.",
	DiagnoseProtokollVerdeckt: "Las líneas del registro no aparecen aquí: " +
		"pertenecen a este proceso y no a este perfil. Quien las " +
		"necesite abre su propio perfil o lee el archivo en la " +
		"ubicación indicada arriba.",
	DiagnoseSchema:          "Esquema",
	DiagnoseWaisen:          "Filas huérfanas",
	DiagnoseBau:             "Compilación",
	DiagnoseVeraendert:      "árbol de trabajo modificado",
	DiagnoseBindungsvertrag: "Contrato de enlaces",
	DiagnoseMigrationen:     "%d migraciones, la última %s",
	DiagnoseKeineSchema:     "ninguna migración registrada",
	DiagnoseNichtLesbar:     "no se puede leer",
	DiagnosePakete:          "Paquetes",
	DiagnoseWeitere:         "… y %d más",
	DiagnoseKeinBestand:     "ninguna base de datos abierta",
	DiagnoseKeinProtokoll:   "Ningún registro",
	DiagnoseKeineVerwaltung: "gestión de paquetes no configurada",

	MTGJSONDatei:       "Abrir MTG JSON",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "Abrir una carpeta con MTG JSON",
	MTGJSONSchreiben:   "Guardar como MTG JSON",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "Guardar mis datos",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "Guardar ajustes",
	EinstellungenOeffnen:   "Importar ajustes",
	EinstellungenVorschlag: "ajustes.json",

	VorlageSchreiben: "Exportar mediante plantilla",
	VorlageFilter:    "Salida de la plantilla",

	Bildordner:     "Elegir una carpeta para las imágenes",
	BogenSchreiben: "Guardar hoja de impresión",
	BogenVorschlag: "hoja-de-impresion.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "Elegir un paquete de plantillas",
	MSE2Waehlen:  "Elegir la carpeta de las plantillas de MSE2",

	Planordner: "Elegir una carpeta para la copia de seguridad semanal",
}

var berichttexteEs = domain.Berichttexte{
	Spiel: "Juego", Format: "Formato", Stand: "Estado", Abgeschlossen: "concluido el",
	Laufend: "en curso", Teilnehmende: "Participantes", Ausgestiegen: "de baja", Runden: "Rondas",
	Rangliste: "Clasificación", Platz: "Puesto", Name: "Nombre", Punkte: "Puntos", Bilanz: "Balance",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "Ronda", Tisch: "Mesa", SeiteA: "Jugador A", SeiteB: "Jugador B", Ergebnis: "Resultado",
	Bye: "Bye", Offen: "aún sin jugar",
}
