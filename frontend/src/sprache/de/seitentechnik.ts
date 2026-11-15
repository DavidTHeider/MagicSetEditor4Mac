// Die Texte der **Technikseite** — Betriebszustand, Wirkungszähler und die
// drei Werkzeuge ohne Fenster.
//
// Aus `seiten.ts` herausgelöst, als die Datei mit dem Wirkungszähler
// (Werkstück O9, #1433) über 400 Zeilen wuchs (§2). **Der Schnitt folgt der
// Frage**: dort stehen die Texte aller Seiten, hier die einer, die eine
// eigene Sprache hat — die des Betriebs.

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: 'Betriebszustand',
  gesundheitErklaerung:
    'Ob dieser Prozess lebt und ob er arbeiten kann. Ein fremder Dienst, der nicht antwortet, macht ihn nicht unbereit — das Bearbeiten geht ohne Netz weiter.',
  lebt: 'Der Prozess antwortet.',
  bereit: 'Er kann arbeiten.',
  nichtBereit: 'Er kann gerade nicht arbeiten.',
  grundBestandZu: 'Es ist kein Bestand geöffnet.',
  grundWanderungLaeuft: 'Der Bestand wandert gerade auf ein neues Schema.',
  grundBestandUnlesbar: 'Der Bestand antwortet nicht.',
  fremdeDienste: 'Fremde Dienste',
  erreichbar: 'erreichbar',
  nichtErreichbar: 'nicht erreichbar',
  metriken: 'Gewirkt seit dem Start',
  metrikenErklaerung:
    'Was dieser Prozess nach außen gewirkt hat — je Fläche, Wirkungsklasse und Ausgang. Gezählt wird, was am Riegel vorbeikam; reines Lesen steht nicht darin.',
  metrikenGrenze:
    'Höchstens {grenze} Zeilen — sie wachsen nicht mit der Zahl der Karten, Sets oder Profile.',
  metrikenLeer: 'Bisher nichts nach außen gewirkt.',
  spalteFlaeche: 'Fläche',
  spalteWirkung: 'Wirkung',
  spalteAusgang: 'Ausgang',
  spalteAnzahl: 'Anzahl',
  flaechen: {
    fenster: 'Fenster',
    kommandozeile: 'Kommandozeile',
    mcp: 'MCP',
    dienst: 'Den Dienst für Skripte starten',
  },
  wirkungen: {
    lokalLesen: 'lokal lesen',
    lokalSchreiben: 'lokal schreiben',
    externSenden: 'nach außen senden',
    externSchreiben: 'außen ändern',
    verbindlich: 'verbindlich',
  },
  ausgaenge: { gewaehrt: 'gewährt', abgewiesen: 'abgewiesen' },
  paket: {
    ueberschrift: 'Diagnosepaket',
    erklaerung:
      'Dieselbe Auskunft in Teilen, als Datei mit Prüfsummen — für einen Bericht, der mehr braucht als einen Text. Es entsteht dort, wo du hinzeigst, und geht von selbst nirgendwohin.',
    inhalt:
      'Darin: Fassung, Plattform, Bestand, Pakete, Prozesszahlen, Protokoll und Absturzspuren — keine Karten, Bilder, Sammlung oder Zugangsdaten.',
    zusammenstellen: 'Zusammenstellen',
    sichern: 'Paket sichern',
    loeschen: 'Paket löschen',
    groesse: '{bytes} von höchstens {grenze}.',
    zeitraum: 'Die Protokollzeilen reichen von {von} bis {bis}.',
    ausgelassen: 'Ausgelassen — {teil}:',
    teile: {
      fassung: 'Fassung und Plattform',
      bestand: 'Ort und Stand des Bestands',
      pakete: 'Eingebundene Vorlagenpakete',
      metriken: 'Zahlen dieses Prozesses',
      protokoll: 'Protokollzeilen',
    },
    spalteTeil: 'Teil',
    spalteDatei: 'Datei',
    spalteZeilen: 'Zeilen',
    spalteBytes: 'Größe',
    spalteMitnehmen: 'Mitnehmen',
    mitnehmen: '{teil} ins Paket nehmen',
  },
  diagnose: {
    ueberschrift: 'Diagnose für einen Fehlerbericht',
    erklaerung:
      'Wenn etwas nicht geht, hilft dieser Text beim Suchen. Er wird hier angezeigt — weitergeben tust du ihn selbst.',
    inhalt:
      'Darin stehen Fassung, Rechner, Bestand, Protokoll und Vorlagenpakete — keine Kartentexte, keine Feldwerte, keine Setinhalte. Im Einzelnen: Fassung, Rechnerart, der Bau, aus dem das Programm stammt, der Bindungsvertrag zwischen Fenster und Programm, der Ort des Bestands, der Schemastand, wie viele verwaiste Zeilen im Bestand liegen, der Ort des Protokolls, deine Vorlagenpakete und die letzten 200 Protokollzeilen. Dein Benutzerverzeichnis erscheint als „~“.',
    anzeigen: 'Diagnose anzeigen',
    kopieren: 'Kopieren',
    kopiert: 'Kopiert',
    kopierfehler:
      'Kopieren ging nicht — dem Fenster fehlt die Erlaubnis für die Zwischenablage. Markiere den Text und kopiere ihn von Hand.',
    sichern: 'Als Datei sichern',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'In den Zeilen suchen',
    zeilenGezeigt: '{gezeigt} von {alle} Zeilen',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster“ war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Technik und Diagnose',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Netz und Erreichbarkeit',

  ueberschrift: 'Werkzeuge ohne Fenster',
  erklaerung:
    'Diese Anwendung bringt drei Programme mit, die ohne Fenster arbeiten. mse für die Kommandozeile, mse-mcp für ein Sprachmodell, mse-rest für einen Dienst für Skripte; sie liegen im Programmbündel, in einen Suchpfad kommen sie erst, wenn du sie unten einrichtest, und die Zeilen hier tragen den vollen Ort.',
  keineShell:
    'Kopiert wird, nicht ausgeführt — was hier steht, gehört in ein Terminal deiner Wahl.',
  ort: 'Sie liegen unter {ort}.',
  vorhanden: 'Vorhanden: {namen}.',
  keineWerkzeuge:
    'Dort liegt keines der drei. Das ist der Normalfall, solange die Anwendung aus dem Quelltext läuft — der Freigabeweg legt sie ins Bündel.',
  kopieren: 'Befehl kopieren',
  kopiert: 'Kopiert.',
  kopierfehler:
    'Kopieren ging nicht — dem Fenster fehlt die Erlaubnis für die Zwischenablage. Markiere den Text und kopiere ihn von Hand.',
  befehle: {
    hilfe: 'Was die Kommandozeile kann',
    sets: 'Die Sets im Bestand auflisten',
    verknuepfen: 'mse im Suchpfad verfügbar machen (fragt nach dem Passwort)',
    dienst: 'Den Dienst auf der Rückschleife starten',
    mcp: 'Den MCP-Server starten — für ein Sprachmodell, über stdio',
    karten: 'Die Karten des offenen Sets auflisten',
    statistik: 'Das offene Set auszählen',
  },

  // **Der lokale Suchindex** — Werkstück A15 von Phase 5.3
  // (#1514).
  suchindex: 'Lokaler Suchindex',
  suchindexErklaerung:
    'Kartentexte und Regelstellen, lokal durchsuchbar — im Volltext und über Vektoren, zusammengeführt zu einer Liste. Jeder Treffer nennt, woher er stammt und in welchem Stand die Quelle war.',
  suchindexKeiner:
    'Es ist kein Suchindex aufgebaut. Bis dahin findet die Suche hier nichts.',
  suchindexModell: 'Modell',
  suchindexLaufzeit: 'Laufzeit',
  suchindexVerfahren: 'Verfahren',
  suchindexStaende: 'Stände je Quelle',
  suchindexSemantisch: 'Semantisch suchbar',
  suchindexJa: 'ja',
  suchindexNein: 'nein',
  suchindexFrage: 'Wonach suchen?',
  suchindexSuchen: 'Suchen',
  suchindexLeer: 'Nichts gefunden.',
  suchindexModus: 'Gesucht wurde',
  suchindexModusHybrid: 'im Volltext und über Vektoren',
  suchindexModusLexikalisch: 'nur im Volltext',
  suchindexModusOhneIndex: 'gar nicht — es steht kein Index bereit',
  suchindexAufbau: 'Neu aufbauen',
  suchindexAufbauErklaerung:
    'Baut eine neue Generation neben der laufenden und schaltet erst am Ende um. Das dauert Minuten und braucht ein eingerichtetes Einbettungsmodell.',
  suchindexModellfeld: 'Einbettungsmodell',
  suchindexQuellfeld: 'Sammeldatei',
  suchindexRevisionfeld: 'Stand der Quelle',
  suchindexDelta: 'Nur Geändertes einlesen',
  suchindexAufbauen: 'Aufbauen',
  suchindexAbbrechen: 'Abbrechen',
  suchindexLaeuft: 'Aufbau läuft',
  suchindexGeschrieben: 'Abschnitte geschrieben',
  suchindexFertig: 'Generation aufgebaut',
  // **Konfiguration für Agenten** — A14, #1453.
  agenten: 'Fertig für einen Agenten',
  agentenErklaerung:
    'Claude Code und Codex sprechen MCP. Die Bruchstücke unten tragen den vollen Pfad von mse-mcp auf diesem Rechner — kopieren und einsetzen. Diese Anwendung schreibt in keine dieser Dateien und startet keinen Agenten.',
  agentenOhneGeheimnis:
    'Kein Zugangsschlüssel steht darin. mse-mcp liest sie selbst aus dem Schlüsselbund — was hier stünde, läge im Klartext in einer Datei, die man weitergibt.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Die Kommandozeile einrichten',
  befehlszeileSatz:
    'Legt Verknüpfungen auf mse und mse-mcp in ein Verzeichnis, das dir gehört — ohne Administratorrechte und ohne dein Shellprofil anzufassen. Was dort schon liegt und nicht von hier stammt, bleibt.',
  befehlszeileVerzeichnis: 'Verzeichnis: {verzeichnis}',
  befehlszeileNichtImPfad:
    'Dieses Verzeichnis steht nicht im PATH. Trag diese Zeile selbst in dein Shellprofil ein:',
  befehlszeileStaende: {
    frei: 'noch nicht eingerichtet',
    eigen: 'eingerichtet',
    veraltet: 'veraltet — zeigt auf ein Programm, das es dort nicht mehr gibt',
    fremd: 'dort liegt ein fremder Befehl, er bleibt',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    'Verdeckt von {pfad}: dieser läuft, wenn du den Namen tippst.',
  befehlszeilePruefen: 'Prüfen mit',
  befehlszeileEinrichten: 'Einrichten',
  befehlszeileHerausnehmen: 'Wieder herausnehmen',
  befehlszeileKopierenPfad: 'Die Zeile fürs Shellprofil kopieren',
  befehlszeileKopierenPruefung: 'Den Prüfbefehl kopieren',
  agentenDatei: 'Gehört nach {datei}',
  agentenInhalt: 'Inhalt für {datei}',
  agentenBefehl: 'Oder über die Kommandozeile des Agenten',
  agentenNachgelesen: 'Nachgelesen: {quelle}',
  agentenKopierenDatei: 'Bruchstück kopieren',
  agentenKopierenBefehl: 'Befehl des Agenten kopieren',

  // **Die eigene Peeridentität** — Werkstück N1 von Phase 5.4
  // (#1454).
  peerTitel: 'Womit sich dieses Profil ausweist',
  peerZeigen: 'Identität zeigen',
  peerErklaerung:
    'Ein Schlüsselpaar, mit dem andere erkennen, dass sie wirklich mit dir sprechen. Der private Teil liegt im Schlüsselbund dieses Rechners und geht nirgends hin; nach außen geht nur die Kennung.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'Sie gehört diesem Profil — sie ist weder dein Apple-Account noch der Name dieses Rechners.',
  peerStandOhne: 'Für dieses Profil gibt es noch keine Identität.',
  peerStandBereit: 'Bereit.',
  peerStandVerloren: 'Der Schlüssel fehlt.',
  peerStandFremd: 'Schlüsselbund und Bestand widersprechen sich.',
  peerKennung: 'Kennung',
  peerFingerabdruck: 'Fingerabdruck',
  peerFingerabdruckHinweis:
    'Das ist, was du deinem Gegenüber vorliest — ein Anzeigename ist kein Nachweis.',
  peerSeit: 'Seit {zeitpunkt}',
  peerVorher: 'Vorher: {kennung}',
  peerAnlegen: 'Identität anlegen',
  peerRotieren: 'Identität ablösen',
  peerRotierenFrage:
    'Eine neue Identität anlegen? Jeder, der den bisherigen Fingerabdruck bestätigt hat, muss ihn erneut prüfen.',
  peerRotierenJa: 'Ablösen',
  peerRotierenNein: 'Abbrechen',

  // **Die bekannten Gegenüber** — Werkstück N1 desselben Issues.
  kontakteTitel: 'Wen dieses Profil kennt',
  kontakteErklaerung:
    'Gegenüber, denen dieses Profil schon begegnet ist. Eine Begegnung allein erlaubt nichts: erst der Vergleich des Fingerabdrucks macht jemanden prüfbar, und erst eine Freigabe je Zweck erlaubt etwas.',
  kontakteLeer:
    'Noch niemandem begegnet. Wer sich meldet, steht danach hier — unbestätigt.',
  kontakteZeigen: 'Kontakte zeigen',
  kontakteAufnehmen: 'Kontakt aufnehmen',
  kontakteAufnehmenErklaerung:
    'Solange es keine Verbindung gibt, tauscht man Kennungen von Hand. Wer hier eingetragen wird, ist unbestätigt und darf nichts.',
  kontakteAufnehmenKennung: 'Kennung des Gegenübers',
  kontakteAufnehmenName: 'Wie du ihn nennst',
  kontakteStandUnbestaetigt: 'Unbestätigt',
  kontakteStandBestaetigt: 'Bestätigt',
  kontakteStandGesperrt: 'Gesperrt',
  kontakteAbdruck: 'Fingerabdruck',
  kontakteNennt: 'Nennt sich: {name}',
  kontakteNenntHinweis: 'Ein selbstgewählter Name ist kein Nachweis.',
  kontakteGesehen: 'Zuletzt gesehen: {zeitpunkt}',
  kontakteVergleichen: 'Verglichenen Fingerabdruck eintragen',
  kontakteBestaetigen: 'Bestätigen',
  kontakteSperren: 'Sperren',
  kontakteEntsperren: 'Entsperren',
  kontakteRechte: 'Freigegeben',
  kontakteKeineRechte: 'Nichts freigegeben.',
  zweckPaarung: 'Kennenlernen',
  zweckTeilen: 'Objekte übertragen',
  zweckHandel: 'Karten tauschen',
  zweckPartie: 'Eine Partie spielen',
  kontakteAdresse: 'Erreichbar unter',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'Entweder eine Adresse im selben Netz — oder eine Einladung, die dir dein Gegenüber geschickt hat. Leer heißt: keine.',
  kontakteAdresseMerken: 'Merken',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Notiz',
  kontakteNotizBeispiel: 'tauscht fair, antwortet langsam …',
  kontakteNotizHinweis:
    'Nur für dich. Sie geht über keine Leitung — weder in eine Sicht noch in eine Freigabe.',
  kontakteNotizMerken: 'Notiz merken',
  freigabenTitel: 'Was dieses Gegenüber sieht',
  freigabenErklaerung:
    'Ohne Freigabe sieht es nichts. Mehrere Umfänge gelten nebeneinander — der weitere gewinnt.',
  freigabenLeer: 'Keine Sicht erteilt.',
  freigabenOhneZweck:
    'Dieses Gegenüber darf noch nichts. Erlaube ihm zuerst einen Zweck.',
  freigabenZweck: 'Wofür',
  freigabenUmfang: 'Was',
  freigabenErteilen: 'Sicht erteilen',
  freigabenWiderrufen: 'Widerrufen',
  freigabenVorschau: 'Ansehen, was es sieht',
  freigabenVorschauLeer: 'Es sähe nichts.',
  fernTitel: 'Über fremde Netze erreichbar sein',
  fernErklaerung:
    'Damit erreicht dich ein Gegenüber auch dann, wenn ihr nicht im selben Netz seid — ohne Portweiterleitung und ohne Konto. Du bekommst eine Einladung und schickst sie ihm, auf welchem Weg du willst.',
  fernStill: 'Dieses Profil ist über fremde Netze nicht erreichbar.',
  fernHinweis:
    'Die Verbindung sucht sich den direkten Weg und nimmt ein Relais nur, wenn es anders nicht geht. Über das Relais läuft nur Verschlüsseltes — keine Karten, keine Namen.',
  fernStarten: 'Einladung erzeugen',
  fernBeenden: 'Nicht mehr erreichbar sein',
  fernWeitergeben:
    'Das hier braucht das Gegenüber. Wer sie abfängt, bekommt eine Leitung und keine Identität — geprüft wird beim Verbinden über den Fingerabdruck.',
  fernStirbt:
    'Die Einladung gilt, solange dieses Profil erreichbar ist. Beendest du es, wird sie ungültig — und beim nächsten Mal gibt es eine neue.',
  fernKopieren: 'Einladung kopieren',
  fernKopiert: 'Kopiert.',
  horchenTitel: 'Für andere erreichbar sein',
  horchenErklaerung:
    'Solange dieses Profil horcht, kann ein bekanntes Gegenüber eine ' +
    'Verbindung dorthin aufbauen. Von selbst geschieht das nie — auch ' +
    'nach einem Neustart nicht —, und ein Profilwechsel beendet es.',
  horchenAdresse: 'Adresse, unter der gehorcht wird',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'Nur der Port (:41234) horcht auf jeder Schnittstelle dieses Rechners. ' +
    'Mit 127.0.0.1 davor nur auf diesem Rechner selbst.',
  horchenStarten: 'Erreichbar werden',
  horchenBeenden: 'Nicht mehr erreichbar sein',
  horchenLaeuft: 'Gehorcht wird auf {adresse}',
  horchenNamen: 'Name (gilt auch nach einem Adresswechsel)',
  horchenAdressen: 'Adresse',
  horchenNurHier:
    'Von außen erreicht dich unter dieser Bindung niemand: 127.0.0.1 horcht nur auf diesem Rechner. Gib stattdessen nur den Port an (:41234), damit dieser Mac in seinem Netz erreichbar ist.',
  horchenNurDiesesNetz:
    'Beides gilt in diesem Netz. Eine geliehene Adresse kann sich ändern, ein Name überlebt das meistens — und eine Adresse, die überall gilt, kommt erst mit der Verbindung über fremde Netze.',
  horchenStill: 'Dieses Profil horcht gerade nicht.',
  horchenWeitergeben:
    'Das hier braucht das Gegenüber — zusammen mit der Kennung dieses Profils.',
  sichtbarTitel: 'In der Nähe gefunden werden',
  sichtbarErklaerung:
    'Wer sich ankündigt, sagt jedem im selben Netz, dass es dieses Profil ' +
    'gibt — samt seiner Kennung. Erreichbar sein und gefunden werden sind ' +
    'zwei Entscheidungen, und das hier ist die zweite.',
  sichtbarName: 'Name, unter dem angekündigt wird',
  sichtbarNameHinweis:
    'Ein selbstgewählter Name belegt nichts. Der Rechnername dieses Macs ' +
    'geht nicht hinaus.',
  sichtbarWerden: 'Ankündigen',
  sichtbarBeenden: 'Nicht mehr ankündigen',
  sichtbarLaeuft: 'Dieses Profil kündigt sich an.',
  sichtbarStill: 'Dieses Profil kündigt sich nicht an.',
  sichtbarBrauchtHorchen:
    'Erst erreichbar werden — eine Ankündigung ohne Port wäre eine ' +
    'Einladung an eine Tür, hinter der niemand steht.',
  nachbarnTitel: 'Wer in der Nähe ist',
  nachbarnErklaerung:
    'Gegenüber, die sich gerade im selben Netz ankündigen. Was hier steht, ' +
    'ist ihre Behauptung — geprüft wird sie erst beim Verbinden.',
  nachbarnSuchen: 'Nachsehen',
  nachbarnSucht: 'Wird gesucht …',
  nachbarnLeer: 'Gerade kündigt sich niemand an.',
  nachbarnUnbekannt: 'Unbekannt',
  nachbarnAufnehmen: 'Als Kontakt aufnehmen',
  nachbarnGemerkt: 'Gemerkt wird',
  nachbarnSchonBekannt: 'Steht schon in der Liste',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'Er steht an einer neuen Adresse.',
  nachbarnAdresseBisher: 'Bisher gemerkt',
  nachbarnAdresseKeine: 'keine',
  nachbarnAdresseUebernehmen: 'Adresse übernehmen',
  nachbarnAdresseHinweis:
    'Übernehmen ändert nur, wohin gewählt wird. Vertraut wird ihm nicht mehr als vorher, und sein Schlüssel wird bei jeder Verbindung geprüft.',
  nachbarnVerworfen:
    '{zahl} Antworten trugen keine Peerkennung und stehen nicht in der Liste.',

  // **Wie dieses Programm ins Netz geht** — N5
  // (#1571),
  // eingedampft für UX-15.AK-04
  // (#1771).
  //
  // **Hier steht nur, was kein Feld sagen kann**: woher die Werte kommen und
  // dass sie die Felder darunter übersteuern.
  netzwegTitel: 'Aus der Umgebung dieses Prozesses',
  netzwegErklaerung:
    'Diese Werte kommen aus der Umgebung, in der das Programm gestartet wurde. Sie übersteuern die Felder darunter, und kein Fenster ändert sie.',
  netzwegProxyHTTP: 'Proxy für http',
  netzwegProxyHTTPS: 'Proxy für https',
  netzwegProxyAus: 'Ohne Proxy',
  netzwegAusUmgebung: '{name}: {wert}',

  // **Die zwei schaltbaren Dienste** — N6
  // (#1572).
  dienstRESTTitel: 'Die Schnittstelle für Skripte',
  dienstRESTErklaerung:
    'Ein Webdienst über denselben Bestand — zum Auslesen und Ändern aus einem Skript. Er läuft nur, solange du ihn laufen lässt, und endet mit diesem Fenster.',
  dienstMCPTitel: 'Der Werkzeugserver über HTTP',
  dienstMCPErklaerung:
    'Dieselben Werkzeuge wie über die Standardeingabe, hier über eine Adresse — erreichbar auch von einem anderen Gerät.',
  // **Je Dienst benannt** (#1767): wer Feld für Feld tabbt, hört sonst
  // zweimal „Adresse" und erfährt nicht, welche gemeint ist.
  dienstRESTAdresse: 'Adresse der Skriptschnittstelle',
  dienstRESTPort: 'Port der Skriptschnittstelle',
  dienstMCPAdresse: 'Adresse des Werkzeugservers',
  dienstMCPPort: 'Port des Werkzeugservers',
  dienstMarke: 'Zugang',
  dienstMarkeKopieren: 'Zugang kopieren',
  dienstFingerabdruck: 'Fingerabdruck',
  dienstFingerabdruckKopieren: 'Fingerabdruck kopieren',
  dienstKopiert: 'Kopiert',
  dienstStarten: 'Starten',
  dienstBeenden: 'Beenden',
  dienstErreichbar: 'Erreichbar unter:',
  dienstMitTLS:
    'Verschlüsselt. Das Zeugnis ist selbst ausgestellt und trägt die Kennung dieses Profils — wer sich verbindet, vergleicht ihren Fingerabdruck.',
  dienstOhneTLS:
    'Unverschlüsselt, und das ist hier in Ordnung: die Adresse gilt nur auf diesem Rechner.',
  dienstHinweis:
    '127.0.0.1 heißt: nur dieser Rechner. 0.0.0.0 heißt: jedes Gerät im Netz — dann sind Verschlüsselung und Zugang Pflicht, und beides richtet sich von selbst ein. Port 0 heißt: einen freien suchen.',
}
