// Die Texte der einzelnen Seiten.
//
// **Ein `titel` je Seite ist Pflicht**: `navigation.ts` liest ihn, und
// `navigation.test.ts` weist eine Seite ohne Namen ab. Ein leerer Eintrag in
// der Navigation fiele sonst erst im Betrieb auf — als etwas, das niemand
// anklickt.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'Angaben der Datei',
    kopfHinweis:
      'Was Magic Set Editor beim Speichern über dieses Set notiert hat. Sie werden mitgeführt und nicht geändert.',
    mseFassung: 'Geschrieben mit MSE',
    spielfassung: 'Fassung des Spielpakets',
    stilfassung: 'Fassung des Stilpakets',
    versionskontrolle: 'Versionskontrolle',
    apprenticecode: 'Apprentice-Code',
    titel: 'Setangaben',
    leerTitel: 'Keine Angabenfelder',
    leer: 'Das Spielpaket dieses Sets erklärt keine `set field:`-Blöcke. Wähle ein Spiel, das welche mitbringt.',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: 'Bilder',
    erklaerung:
      'Jede Karte wird einzeln gezeichnet und geschrieben. Der Lauf läuft hier im Fenster: er zeigt, wo er steht, und lässt sich jederzeit anhalten.',
    ordnerWaehlen: 'Ordner wählen …',
    keinOrdner: 'Noch kein Ordner gewählt.',
    gemerkt: 'Zuletzt gewählt — derselbe Ordner für alle Sets.',
    nichtGemerkt: 'Für diesen Lauf gilt der Ordner, gemerkt wurde er nicht:',
    vorlage: 'Namensvorlage',
    konflikt: 'Wenn eine Datei schon da ist',
    konfliktHilfe:
      'Zwei Karten desselben Namens ergeben denselben Dateinamen. „Neu nummerieren“ ist die Vorgabe von Magic Set Editor 2: es zählt innerhalb eines Laufs hoch und überschreibt, was ein früherer hinterlassen hat.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Vorhandene behalten',
      ueberschreiben: 'Überschreiben',
      nummerieren: 'Nummerieren',
      nummerierenneu: 'Neu nummerieren',
    },
    vorlageHilfe:
      '{nummer} wird zur dreistelligen Kartennummer, {name} zum Kartennamen. Gleiche Namen bekommen eine Zahl angehängt.',
    format: 'Format',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Güte',
    gueteFast: 'Schnell — für eine rasche Ansicht',
    gueteMedium: 'Nach der Regel von Magic Set Editor 2',
    gueteHigh: 'Hoch — ohne Flimmern an feinen Linien',
    gueteUltra: 'Höchst — für den Druck in voller Auflösung',
    breite: 'Breite in Bildpunkten',
    breiteHilfe:
      'Keine dpi-Angabe: wie groß eine Karte in Millimetern ist, sagt keine Vorlage. 750 Bildpunkte auf eine 63 mm breite Karte sind 300 dpi.',
    ausgeben: 'Bilder ausgeben',
    abbrechen: 'Anhalten',
    laeuft: 'geschrieben von',
    fertig: 'Bilder geschrieben nach',
    abgebrochen: 'Angehalten.',
    uebersprungen:
      '{anzahl} Karten übersprungen — die Datei war schon da, und „Vorhandene behalten“ ist gewählt:',
    befunde:
      'Beim Zeichnen nicht vollständig. Die Bilder sind da; diese Stellen fehlen darin:',
  },
  druckbogen: {
    titel: 'Druckbogen',
    erklaerung:
      'Die Karten sitzen maßhaltig: 63 × 88 mm, so wie sie geschnitten werden. Die Schnittmarken stehen am Papierrand, nicht auf der Karte.',
    papier: 'Papier',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Anschnitt in Millimetern',
    anschnittHilfe:
      'Der Anschnitt ist, was das Bild über die Schnittkante hinausreicht — er sorgt dafür, dass eine geschnittene Karte keinen weißen Rand hat. Die Karten liegen dafür mit dem doppelten Anschnitt Abstand.',
    aufloesung: 'Auflösung',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'Schnittmarken',
    legen: 'Druckbogen legen …',
    fertig: 'auf',
    seiten: 'Seiten:',
    befunde:
      'Beim Legen nicht überall gepasst. Der Bogen ist da; diese Stellen sind zu prüfen:',
  },
  // Der **technische Bereich** (#1164) — Befehle zum Kopieren.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Sammeldaten von Scryfall',
      erklaerung:
        'Eine Datei mit allen Kartenangaben. Liegt sie hier, arbeiten die Scryfall-Wege auch ohne Netz — mit einem Stand von bis zu einem Tag.',
      fehlt: 'Noch nicht geholt.',
      vorhanden: 'Vorhanden, {mb} MB.',
      angeboten: 'Scryfall bietet {mb} MB an, Stand {stand}.',
      geholt: 'Geholt {zeit}.',
      holen: 'Sammeldaten holen',
      laedt: 'Wird geholt …',
    },

    titel: 'Pakete',
    bezugUeberschrift: 'Vorlage beziehen',
    bezugErklaerung:
      'MSE-Vorlagen liegen als Repositorien. Bezogen wird ein Stand — eine Freigabe oder ein Commit —, nicht ein Zweig: ein Zweig ändert sich, und dieselbe Angabe ergäbe morgen ein anderes Paket.',
    bezugAdresse: 'Adresse des Repositoriums',
    bezugStand: 'Stand (Freigabe oder Commit)',
    bezugUnterverzeichnis: 'Unterverzeichnis (Paketwurzel)',
    bezugName: 'Name im Paketverzeichnis',
    bezugPruefen: 'Herkunft anzeigen …',
    bezugStarten: 'Beziehen',
    bezugAbbrechen: 'Nicht beziehen',
    bezugGeholt: 'Bezogen nach',
    bezugPruefsumme: 'Prüfsumme',
    bezugUebergangen:
      'Nicht ausgepackt — ausführbare Dateien kommen nicht mit:',
    bezugUnvollstaendig:
      'Adresse und Stand fehlen. Ohne beides gibt es nichts zu beziehen.',
    bezugHerkunft: 'Bezogen wird {quelle}, Stand {stand}.',
    bezugLizenz:
      'Diese Inhalte stammen nicht von MagicSetEditor4Mac und stehen nicht unter unserer Lizenz. Kartenrahmen, Symbole und Schriften können geschützt sein; wer sie verwendet, tut das in eigener Verantwortung. Ausführbare Dateien werden nicht ausgepackt.',
    mse2Ueberschrift: 'Aus einer MSE2-Installation übernehmen',
    mse2Erklaerung:
      'Wer MSE2 benutzt, hat seine Vorlagen schon auf dem Rechner. Die üblichen Orte stehen unten — gelesen wird keiner davon, bevor du ihn wählst.',
    mse2Uebernehmen: 'Von hier übernehmen',
    mse2VonHier: 'Von hier übernehmen — {ort}',
    mse2Waehlen: 'Verzeichnis wählen …',
    mse2KeinVorschlag:
      'Keine MSE2-Installation an den üblichen Orten gefunden. Wähle das Verzeichnis mit den Vorlagen selbst — bei MSE2 heißt es „data“.',
    installerErklaerung:
      'Oder eine Datei auf .mse-installer — so gibt MSE2 Vorlagen weiter. Was schon eingebunden ist, bleibt, wie es ist.',
    installerWaehlen: 'Installer einlesen …',
    mse2Ergebnis: '{anzahl} Vorlagen übernommen.',
    mse2Uebergangen:
      'Schon vorhanden und deshalb übergangen — nichts wurde überschrieben:',
    abhaengigUeberschrift: 'Abhängigkeiten',
    abhaengigAllesGut: 'Jede Abhängigkeit ist erfüllt.',
    abhaengigFehlt:
      '{paket} braucht {braucht} — nicht installiert. Beziehen oder aus einer MSE2-Installation übernehmen.',
    abhaengigZuAlt:
      '{paket} braucht {braucht} in der Fassung {verlangt}; installiert ist {da}. Ohne die neuere zeichnet die Vorlage still etwas anderes.',
    abhaengigZyklus: '{paket} und {braucht} brauchen einander im Kreis.',
    abhaengigUnlesbar:
      '{paket} braucht {braucht}: die Versionen lassen sich nicht vergleichen — verlangt {verlangt}, installiert {da}.',
    quellenErklaerung:
      'Vier Sammlungen zum Anklicken — Adresse und Stand werden eingetragen, geholt wird erst nach deiner Zustimmung. Die Schriften der Vorlagen kommen mit; ohne sie zeichnet keine Karte richtig. Keine der vier trägt eine Lizenz: die Inhalte gehören ihren Rechteinhabern, nicht uns.',
    quellenGroesse: 'rund {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15:
        'Die üblichen M15-Rahmen. Am schnellsten zu einer gezeichneten Karte.',
      fullNonMagic: 'Vorlagen für andere Spiele als Magic.',
      m15: 'Die öffentlichen M15-Vorlagen — 324 Pakete. Für die meisten der richtige Anfang.',
      fullMagic:
        'Fast alle öffentlichen Magic-Vorlagen — 526 Pakete. Diese Sammlung ist der Maßstab dieses Projekts.',
    },
    katalogUeberschrift: 'Installierte Pakete',
    katalogLeer:
      'Kein Paket installiert. Ohne Spiel und Stil zeigt die Vorschau einen Behelfsplan.',
    katalogName: 'Name',
    katalogArt: 'Art',
    katalogVersion: 'Version',
    katalogHandlung: 'Handlung',
    katalogBefunde: 'Beim Lesen der Paketverzeichnisse:',
    anlegen: 'Neues Paket anlegen',
    anlegenArt: 'Art des neuen Pakets',
    anlegenArtSpiel: 'Spiel (.mse-game)',
    anlegenArtStil: 'Stil (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Exportvorlage (.mse-export-template)',
    anlegenName: 'Kurzname',
    anlegenNameHinweis:
      'Buchstaben, Ziffern, Bindestriche und Unterstriche — er wird zum Verzeichnisnamen',
    anlegenVollname: 'Vollständiger Name',
    anlegenVollnameHinweis: 'Leer heißt: wie der Kurzname',
    anlegenSpiel: 'Spiel des neuen Pakets',
    anlegenSpielHinweis:
      'Ein Stil legt die Felder seines Spiels aus, eine Exportvorlage gibt sie aus',
    anlegenSpielLeer: 'Bitte wählen',
    anlegenKnopf: 'Anlegen',
    anlegenGelungen: 'Angelegt: {name}',
    // **„Löschen“ und nicht „Entfernen“** — UX-10.AK-04.
    //
    // `Paketverwalter.Entfernen` **räumt das Verzeichnis**; der Kommentar
    // daneben sagt es selbst („vergleicht und löscht“), und die Rückfrage
    // darunter ebenfalls. Der Knopf hieß trotzdem „Entfernen“ — das
    // schwächere Wort für die stärkere Wirkung.
    loeschen: 'Löschen',
    loeschenAn: 'Löschen — {name}',
    // **Das Menü je Zeile und die Rückfrage davor** — Werkstück U2 von
    // Phase 5.1 (#1413). Die Frage sagt, was geschieht: die Dateien gehen
    // von der Platte, und ein Set ohne sein Paket zeichnet den Behelfsplan.
    handlungenAn: 'Handlungen — {name}',
    // **Das Zeichen auf dem Knopf** — der Name steht in `handlungenAn`.
    handlungenKurz: '…',
    katalogSuche: 'Suchen',
    // **Die Auswahl aus der Aufmerksamkeit lösen** — AP-8 (#2365).
    auswahlLoesen: 'Alle Pakete zeigen',
    // **Die Leiste über den Paketen** — AP-9, Teil 2 (#2366).
    leiste: {
      bezeichnung: 'Kennzahlen der Pakete',
      installiert: 'Installiert',
      befund: 'Mit Befund',
      nichtGelesen: 'nicht gelesen',
      auswahl: '{pakete} mit Befund',
    },
    katalogSucheHinweis: 'Name, Art, Herkunft oder Gruppe',
    katalogArtAlle: 'Alle Arten',
    katalogOrdnung: 'Ordnen nach',
    katalogOrdnungGruppen: 'Gruppen',
    katalogOrdnungName: 'Name',
    katalogOrdnungVersion: 'Version',
    katalogOrdnungHerkunft: 'Herkunft',
    katalogGefunden: '{gefunden} von {gesamt} Paketen',
    katalogNichtsGefunden:
      'Kein Paket passt zu dieser Suche. Andere Schreibweise, oder den Artfilter zurücksetzen.',
    katalogLeerTitel: 'Noch kein Paket',
    katalogNichtsTitel: 'Kein Treffer',
    katalogOhneGruppe: 'Ohne Gruppe',
    katalogGruppeAn: 'Gruppe {name}',
    katalogGruppe: 'Gruppe',
    katalogOhneVersion: 'ohne Version',
    katalogZustandFehlt: 'Abhängigkeit fehlt',
    katalogZustandZuAlt: 'Abhängigkeit zu alt',
    katalogZustandUnlesbar: 'Version unvergleichbar',
    katalogZustandZyklus: 'Kreis von Abhängigkeiten',
    katalogEinVorbehalt: 'ein Vorbehalt',
    katalogVorbehalte: '{anzahl} Vorbehalte',
    umschlagTitel: 'Eine Vorlage aus einem Umschlag',
    umschlagErklaerung:
      'Ein Umschlag trägt die **Dateien** einer Vorlage; den Namen samt Endung wählst du, etwa `meinspiel.mse-game` oder `meinstil.mse-style`. Unter diesem Namen liegt die Vorlage dann bei dir. Gibt es das Paket schon, wird nichts überschrieben: du bekommst eine Absage und wählst einen anderen Namen.',
    umschlagName: 'Wie das Paket heißen soll',
    umschlagOhneNamen: 'Trag zuerst ein, wie das Paket heißen soll.',
    umschlagAnnehmen: 'Vorlage annehmen',
    umschlagAngenommen: 'Die Vorlage liegt jetzt in deinem Paketverzeichnis.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Ersetzen',
    erklaerung:
      'Ersetzt in den Textfeldern dieses Sets. Erst zeigen, dann tun — und zurückgenommen wird alles in einem Zug.',
    suchtext: 'Suchen nach',
    ersatz: 'Ersetzen durch',
    grossKlein: 'Groß und klein unterscheiden',
    ganzesWort: 'Nur ganze Wörter',
    vorschauen: 'Zeigen, was sich ändert',
    ersetzen: 'Jetzt ersetzen ({treffer})',
    nichtsGefunden: 'Nichts gefunden. Ein anderes Wort probieren?',
    // **Gezählt und nicht gebeugt.** „1 Treffer in 1 Karten“ liest sich
    // falsch, und einen Numerus hat dieser Katalog nicht. Eine Aufzählung
    // mit Doppelpunkt trägt jede Zahl, ohne dass jemand eine Beugung
    // erfinden muss.
    zusammenfassung: 'Treffer: {treffer} · betroffene Karten: {karten}',
    getan: 'Ersetzt — Treffer: {treffer} · Karten: {karten}',
    ruecknahme: 'Mit ⌘Z wird der ganze Zug zurückgenommen.',
    stellen: 'Betroffene Stellen',
    kartenfeld: 'Karte {nummer} · {feld}',
    weitere:
      'Weitere Stellen sind nicht aufgeführt — ersetzt werden sie trotzdem.',
  },

  statistik: {
    worueber: 'Worüber',
    belegungWahl: 'Belegung je Feld',
    gruppeSpielpaket: 'Nach dem Spielpaket',
    gruppeFelder: 'Nach Feldwert',
    verteilungVon:
      'Verteilung von »{feld}« — Anteil an den Karten, die es tragen',
    nachDimension: 'Auszählung nach »{dimension}«',
    titel: 'Statistik',
    karten: 'Karten',
    felder: 'verschiedene Felder',
    leere: 'Karten ohne einen einzigen Wert',
    belegung: 'Belegung je Feld — Anteil an allen Karten des Sets',
    belegungLeer: 'Dieses Set trägt noch keinen Feldwert.',
    verteilungLeer: 'Dieses Feld trägt in diesem Set keinen Wert.',
    anteil: 'Anteil',
    dimensionenErklaerung:
      'Ein Spielpaket sagt selbst, wonach sich auszählen lässt — Farbgruppe, ' +
      'Manawert, Seltenheit. Jede Dimension rechnet ihren Wert je Karte und ' +
      'braucht deshalb einen Augenblick.',
    dimensionLeer: 'Nach dieser Dimension trägt keine Karte einen Wert.',
    gruppeHinweis:
      'Einen Balken wählen, um seine Karten in der Kartenliste zu sehen.',
    gruppeGewaehlt: 'Gewählt: »{gruppe}« · Karten: {karten}',
    gruppeZeigen: 'Karten anzeigen',
    gruppeAufheben: 'Wahl aufheben',
    gruppeLeer: 'Keine Karte trägt diesen Wert.',
    filterBezeichnung: 'Filter aus der Statistik',
    filterKopf: 'Aus der Statistik: »{quelle}« = »{gruppe}«',
    filterZurStatistik: 'Zur Statistik',
    filterAufheben: 'Filter aufheben',
  },
  // Die Lasche „Stil“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Stil',
    vorschauTitel: 'Vorschau',
    vorschauLeer:
      'Dieses Set hat noch keine Karte — leg eine an, um den Stil zu sehen.',
    eigeneFarbe: 'eigene Farbe',

    // **Der Vorgabeeintrag einer Farboption** (#764). MSE beschriftet ihn mit
    // `default name:`; das steht im ganzen Pack kein einziges Mal, und ohne
    // Angabe ist es das übersetzte Wort — hier also unseres.
    vorgabe: 'Vorgabe',
  },
  suche: {
    titel: 'Suche',
    eingabe: 'Wonach suchen?',
    recherche: 'Recherche …',
    feld: 'Nur in diesem Feld',
    alleFelder: 'Alle Felder',
    set: 'Nur in diesem Set',
    alleSets: 'Alle Sets',
    leer: 'Noch nichts gesucht. Tippe etwas, um Karten zu finden.',
    nichts: 'Nichts gefunden. Ein anderes Wort probieren?',
    gekuerzt: 'Es gibt mehr Treffer, als hier stehen. Enger suchen zeigt sie.',
    treffer: 'Treffer',
  },
  ausgeben,
  hilfe: {
    titel: 'Hilfe',
    erklaerung:
      'Was die Begriffe bedeuten, mit denen diese Anwendung arbeitet — und was beim ersten Start zu tun ist.',
    glossar: 'Begriffe',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'Auf dieser Seite',

    seitenuebersicht: 'Was jede Seite tut',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
