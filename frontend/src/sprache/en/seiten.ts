// The texts of the individual pages.
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
    kopf: 'File information',
    kopfHinweis:
      'What Magic Set Editor noted about this set when it was saved. It is carried along and not changed.',
    mseFassung: 'Written with MSE',
    spielfassung: 'Game package version',
    stilfassung: 'Stylesheet version',
    versionskontrolle: 'Version control',
    apprenticecode: 'Apprentice code',
    titel: 'Set info',
    leerTitel: 'No set fields',
    leer: 'The game package of this set declares no `set field:` blocks. Choose a game that brings some.',
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
    titel: 'Images',
    erklaerung:
      'Every card is drawn and written on its own. The run happens here in the window: it shows where it stands and can be stopped at any time.',
    ordnerWaehlen: 'Choose a directory …',
    keinOrdner: 'No directory chosen yet.',
    gemerkt: 'Last chosen — the same folder for every set.',
    nichtGemerkt: 'The folder applies to this run, but it was not remembered:',
    vorlage: 'Name template',
    konflikt: 'If a file is already there',
    konfliktHilfe:
      'Two cards with the same name produce the same filename. “Number afresh” is the Magic Set Editor 2 default: it counts up within a run and overwrites whatever an earlier run left behind.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Keep the existing one',
      ueberschreiben: 'Overwrite',
      nummerieren: 'Number',
      nummerierenneu: 'Number afresh',
    },
    vorlageHilfe:
      '{nummer} becomes the three-digit card number, {name} the card name. Equal names get a number appended.',
    format: 'Format',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Quality',
    gueteFast: 'Fast — for a quick look',
    gueteMedium: 'By Magic Set Editor 2’s rule',
    gueteHigh: 'High — no shimmer on fine lines',
    gueteUltra: 'Ultra — for printing at full resolution',
    breite: 'Width in pixels',
    breiteHilfe:
      'No dpi: no template states how wide a card is in millimetres. 750 pixels on a 63 mm card are 300 dpi.',
    ausgeben: 'Write images',
    abbrechen: 'Stop',
    laeuft: 'written of',
    fertig: 'images written to',
    abgebrochen: 'Stopped.',
    uebersprungen:
      '{anzahl} cards skipped — the file was already there, and “Keep the existing one” is selected:',
    befunde:
      'Not complete when drawing. The images are there; these parts are missing from them:',
  },
  druckbogen: {
    titel: 'Print sheet',
    erklaerung:
      'The cards keep their size: 63 × 88 mm, exactly as they are cut. The crop marks sit at the paper edge, not on the card.',
    papier: 'Paper',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Bleed in millimetres',
    anschnittHilfe:
      'The bleed is what the image extends beyond the cut line — it keeps a cut card from showing a white edge. Cards are laid out with twice the bleed between them.',
    aufloesung: 'Resolution',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'Crop marks',
    legen: 'Lay out print sheet …',
    fertig: 'on',
    seiten: 'pages:',
    befunde:
      'Not everything fitted. The sheet is there; these places need a look:',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Bulk data from Scryfall',
      erklaerung:
        'A file with all card data. Once it is here, the Scryfall features work without a network — with data up to a day old.',
      fehlt: 'Not fetched yet.',
      vorhanden: 'Present, {mb} MB.',
      angeboten: 'Scryfall offers {mb} MB, as of {stand}.',
      geholt: 'Fetched {zeit}.',
      holen: 'Fetch bulk data',
      laedt: 'Fetching …',
    },

    titel: 'Packages',
    bezugUeberschrift: 'Fetch a template',
    bezugErklaerung:
      'MSE templates live in repositories. What is fetched is a state — a release or a commit — never a branch: a branch moves, and the same entry would yield a different package tomorrow.',
    bezugAdresse: 'Repository address',
    bezugStand: 'State (release or commit)',
    bezugUnterverzeichnis: 'Subdirectory (package root)',
    bezugName: 'Name in the package directory',
    bezugPruefen: 'Show where it comes from …',
    bezugStarten: 'Fetch',
    bezugAbbrechen: 'Do not fetch',
    bezugGeholt: 'Fetched to',
    bezugPruefsumme: 'Checksum',
    bezugUebergangen: 'Not unpacked — executables do not come along:',
    bezugUnvollstaendig:
      'Address and state are missing. Without both there is nothing to fetch.',
    bezugHerkunft: 'This will fetch {quelle}, state {stand}.',
    bezugLizenz:
      'These contents are not from MagicSetEditor4Mac and are not under our licence. Card frames, symbols and fonts may be protected; using them is your own responsibility. Executables are not unpacked.',
    mse2Ueberschrift: 'Take over from an MSE2 installation',
    mse2Erklaerung:
      'Anyone using MSE2 already has the templates on their machine. The usual places are listed below — none of them is read before you choose it.',
    mse2Uebernehmen: 'Take from here',
    mse2VonHier: 'Take from here — {ort}',
    mse2Waehlen: 'Choose a directory …',
    mse2KeinVorschlag:
      'No MSE2 installation found in the usual places. Choose the directory with the templates yourself — in MSE2 it is called “data”.',
    installerErklaerung:
      'Or a .mse-installer file — the way MSE2 shares templates. Anything already installed stays as it is.',
    installerWaehlen: 'Read installer …',
    mse2Ergebnis: '{anzahl} templates taken over.',
    mse2Uebergangen:
      'Already present and therefore skipped — nothing was overwritten:',
    abhaengigUeberschrift: 'Dependencies',
    abhaengigAllesGut: 'Every dependency is satisfied.',
    abhaengigFehlt:
      '{paket} needs {braucht} — not installed. Fetch it, or take it from an MSE2 installation.',
    abhaengigZuAlt:
      '{paket} needs {braucht} in version {verlangt}; installed is {da}. Without the newer one the template quietly draws something else.',
    abhaengigZyklus: '{paket} and {braucht} need each other in a circle.',
    abhaengigUnlesbar:
      '{paket} needs {braucht}: the versions cannot be compared — required {verlangt}, installed {da}.',
    quellenErklaerung:
      'Four collections, one click each — address and revision are filled in; nothing is fetched until you agree. The templates’ fonts come along; without them no card renders correctly. None of the four carries a licence: the contents belong to their rights holders, not to us.',
    quellenGroesse: 'about {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15: 'The usual M15 frames. The quickest route to a rendered card.',
      fullNonMagic: 'Templates for games other than Magic.',
      m15: 'The public M15 templates — 324 packages. The right start for most people.',
      fullMagic:
        'Nearly every public Magic template — 526 packages. This collection is this project’s yardstick.',
    },
    katalogUeberschrift: 'Installed packages',
    katalogLeer:
      'No package installed. Without a game and a style the preview falls back to a stand-in.',
    katalogName: 'Name',
    katalogArt: 'Kind',
    katalogVersion: 'Version',
    katalogHandlung: 'Action',
    katalogBefunde: 'While reading the package directories:',
    anlegen: 'Create a package',
    anlegenArt: 'Kind of new package',
    anlegenArtSpiel: 'Game (.mse-game)',
    anlegenArtStil: 'Style (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Export template (.mse-export-template)',
    anlegenName: 'Short name',
    anlegenNameHinweis:
      'Letters, digits, hyphens and underscores — it becomes the directory name',
    anlegenVollname: 'Full name',
    anlegenVollnameHinweis: 'Empty means: same as the short name',
    anlegenSpiel: 'Game of the new package',
    anlegenSpielHinweis:
      'A style lays out the fields of its game, an export template writes them out',
    anlegenSpielLeer: 'Please choose',
    anlegenKnopf: 'Create',
    anlegenGelungen: 'Created: {name}',
    loeschen: 'Delete',
    loeschenAn: 'Delete — {name}',
    handlungenAn: 'Actions — {name}',
    handlungenKurz: '…',
    katalogSuche: 'Search',
    auswahlLoesen: 'Show all packages',
    leiste: {
      bezeichnung: 'Package figures',
      installiert: 'Installed',
      befund: 'With findings',
      nichtGelesen: 'not read',
      auswahl: '{pakete} with findings',
    },
    katalogSucheHinweis: 'Name, type, origin or group',
    katalogArtAlle: 'All types',
    katalogOrdnung: 'Sort by',
    katalogOrdnungGruppen: 'Groups',
    katalogOrdnungName: 'Name',
    katalogOrdnungVersion: 'Version',
    katalogOrdnungHerkunft: 'Origin',
    katalogGefunden: '{gefunden} of {gesamt} packages',
    katalogNichtsGefunden:
      'No package matches this search. Try another spelling, or clear the type filter.',
    katalogLeerTitel: 'No package yet',
    katalogNichtsTitel: 'No match',
    katalogOhneGruppe: 'Ungrouped',
    katalogGruppeAn: 'Group {name}',
    katalogGruppe: 'Group',
    katalogOhneVersion: 'no version',
    katalogZustandFehlt: 'Dependency missing',
    katalogZustandZuAlt: 'Dependency too old',
    katalogZustandUnlesbar: 'Versions not comparable',
    katalogZustandZyklus: 'Dependency cycle',
    katalogEinVorbehalt: 'one finding',
    katalogVorbehalte: '{anzahl} findings',
    umschlagTitel: 'A template from an envelope',
    umschlagErklaerung:
      'An envelope carries the **files** of a template; you pick the name with its extension, for instance `mygame.mse-game` or `mystyle.mse-style`. The template then lives here under that name. If the package already exists, nothing is overwritten: you get a refusal and pick another name.',
    umschlagName: 'What the package should be called',
    umschlagOhneNamen: 'First say what the package should be called.',
    umschlagAnnehmen: 'Accept template',
    umschlagAngenommen: 'The template is now in your package folder.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Replace',
    erklaerung:
      'Replaces text in this set\u2019s text fields. Preview first, then apply \u2014 and one undo reverses the whole move.',
    suchtext: 'Find',
    ersatz: 'Replace with',
    grossKlein: 'Match case',
    ganzesWort: 'Whole words only',
    vorschauen: 'Show what changes',
    ersetzen: 'Replace now ({treffer})',
    nichtsGefunden: 'Nothing found. Try another word?',
    zusammenfassung: 'Matches: {treffer} \u00b7 cards affected: {karten}',
    getan: 'Replaced \u2014 matches: {treffer} \u00b7 cards: {karten}',
    ruecknahme: 'Press \u2318Z to undo the whole move.',
    stellen: 'Affected places',
    kartenfeld: 'Card {nummer} \u00b7 {feld}',
    weitere:
      'Further places are not listed \u2014 they are replaced all the same.',
  },

  statistik: {
    worueber: 'Count by',
    belegungWahl: 'Coverage per field',
    gruppeSpielpaket: 'By the game package',
    gruppeFelder: 'By field value',
    verteilungVon:
      'Distribution of “{feld}” — share of the cards that carry it',
    nachDimension: 'Statistics by “{dimension}”',
    titel: 'Statistics',
    karten: 'cards',
    felder: 'distinct fields',
    leere: 'cards without a single value',
    belegung: 'Coverage per field — share of all cards in the set',
    belegungLeer: 'This set carries no field value yet.',
    verteilungLeer: 'This field carries no value in this set.',
    anteil: 'Share',
    dimensionenErklaerung:
      'A game package says for itself what can be counted — colour category, ' +
      'mana value, rarity. Each dimension computes its value per card and ' +
      'therefore takes a moment.',
    dimensionLeer: 'No card carries a value along this dimension.',
    gruppeHinweis: 'Choose a bar to see its cards in the card list.',
    gruppeGewaehlt: 'Chosen: “{gruppe}” · cards: {karten}',
    gruppeZeigen: 'Show cards',
    gruppeAufheben: 'Clear choice',
    gruppeLeer: 'No card carries this value.',
    filterBezeichnung: 'Filter from the statistics',
    filterKopf: 'From the statistics: “{quelle}” = “{gruppe}”',
    filterZurStatistik: 'Back to the statistics',
    filterAufheben: 'Clear filter',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Style',
    vorschauTitel: 'Preview',
    vorschauLeer: 'This set has no card yet — create one to see the style.',
    eigeneFarbe: 'custom colour',
    vorgabe: 'Default',
  },
  suche: {
    titel: 'Search',
    eingabe: 'What are you looking for?',
    recherche: 'Research …',
    feld: 'Only in this field',
    alleFelder: 'All fields',
    set: 'Only in this set',
    alleSets: 'All sets',
    leer: 'Nothing searched yet. Type something to find cards.',
    nichts: 'Nothing found. Try another word?',
    gekuerzt:
      'There are more hits than shown here. Narrow the search to see them.',
    treffer: 'Hits',
  },
  ausgeben,
  hilfe: {
    titel: 'Help',
    erklaerung:
      'What the terms this application works with mean — and what to do on first launch.',
    glossar: 'Terms',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'On this page',
    seitenuebersicht: 'What each page does',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
