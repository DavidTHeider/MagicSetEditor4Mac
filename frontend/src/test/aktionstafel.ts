// **Die Tafeln zum Aktionskatalog** — UX-10.AK-01 und AK-04
// (#1682).
//
// Gelesen von `aktionskatalog.test.ts`. Sie stehen für sich, weil sie die
// **Entscheidungen** sind und der Lauf nur ihre Probe: wer eine löschende
// Bindung anlegt, trägt hier ein, was vor ihr steht — und warum.
//
// # Die Regel hinter `loeschende`
//
// **Gefragt wird, wo sich der Verlust nicht in einem Schritt an derselben
// Stelle beheben lässt**: ein Geheimnis, ein Token, ein zweiter Faktor, ein
// Nutzer oder Profil, ein Set mit seinen Karten, ein Protokoll, alle eigenen
// Regeln auf einmal. **Nicht gefragt wird**, wo die Sache ein einzelner
// kurzer Eintrag ist, den man an derselben Stelle neu anlegt, ein erzeugtes
// Ding, das derselbe Knopf neu erzeugt, oder wo die Domäne das Löschen auf
// Entwürfe beschränkt. Eine Frage zu oft kostet auch: wer ohne Grund gefragt
// wird, klickt irgendwann alles weg (`Partienverwaltung.tsx`).
//
// **Bis zum 2026-09-25 fragten acht nicht**, die es hätten tun sollen —
// Zugang, Anbietergeheimnis, Token, zweiter Faktor, Nutzer, Profil,
// Partieprotokoll und das Zurücksetzen der Autoersetzungen. Der Durchgang
// vom 2026-09-21 hatte sie als „in Ordnung" gezählt, ohne dass ein Lauf die
// Zählung hielt. Jetzt hält sie einer.
import type { Aktion } from '../aktionen'

export type Tafeleintrag =
  // rueckfrage: eine `Entscheidung` oder Ja und Nein in der Zeile, in `ort`.
  | { bestaetigung: 'rueckfrage'; ort: string }
  // namen: bestätigt wird mit dem getippten Namen — `feld` ist der Schlüssel
  // der Beschriftung dieses Feldes, und er steht in `ort`.
  | { bestaetigung: 'namen'; ort: string; feld: string }
  // ruecknahme: ⌘Z holt es zurück — `beleg` ist Datei und Marke, an der
  // der Verlauf das festhält.
  | {
      bestaetigung: 'ruecknahme'
      ort: string
      beleg: readonly [string, string]
    }
  // keine: mit Grund.
  | { bestaetigung: 'keine'; ort: string; grund: string }

// loeschendeMuster erkennt eine löschende Bindung an ihrem Namen.
export const loeschendeMuster =
  /(Loeschen|Entfernen|Widerrufen|Vergessen|Zuruecksetzen)$/

export const loeschende: Readonly<Record<string, Tafeleintrag>> = {
  AnbietergeheimnisLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Anbietergeheimnisse.tsx',
  },
  // **Die Frage nennt die Sets, die das Symbol tragen** — und dass sie ihre
  // Kopie behalten (DOM-02.AK-02, #1697).
  BibliothekssymbolLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Symbolbibliothek.tsx',
  },
  SprachpaketLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Eigenuebersetzungzeilen.tsx',
  },
  HandelEntwurfLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Handelsaktionen.tsx',
  },
  KontaktLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Adressbuch.tsx',
  },
  AutoersetzungLoeschen: {
    bestaetigung: 'keine',
    ort: 'seiten/Autoersetzungen.tsx',
    grund:
      'Eine Regel sind zwei Wörter, und dieselbe Seite legt sie in einer Zeile wieder an.',
  },
  AutoersetzungenZuruecksetzen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Autoersetzungen.tsx',
  },
  DeckLoeschen: { bestaetigung: 'rueckfrage', ort: 'bausteine/Deckzeile.tsx' },
  DeckleiheWiderrufen: {
    bestaetigung: 'keine',
    ort: 'seiten/Deckleihe.tsx',
    grund:
      'Eine Leihe lässt sich an derselben Stelle neu gewähren; das Deck gehört weiter dem, der es verleiht.',
  },
  DiagnosepaketLoeschen: {
    bestaetigung: 'keine',
    ort: 'seiten/Diagnosepaket.tsx',
    grund:
      'Ein Diagnosepaket ist erzeugt, nicht geschrieben; derselbe Knopf erzeugt es neu.',
  },
  ExemplarLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Sammlungsinspector.tsx',
  },
  FaktorEntfernen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Zweiterfaktor.tsx',
  },
  FormatLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Formatzeile.tsx',
  },
  FreigabeLoeschen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Profilfreigaben.tsx',
    grund:
      'Gelöscht wird eine Erlaubnis, nicht das Geteilte; sie lässt sich an derselben Stelle neu erteilen.',
  },
  KIModellVergessen: {
    bestaetigung: 'keine',
    ort: 'bausteine/KIMessung.tsx',
    grund:
      'Vergessen wird eine Messung, damit sie neu genommen wird — mit derselben Kostenfreigabe wie beim ersten Mal.',
  },
  KarteLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Kartenloeschen.tsx',
  },
  MarkeEntfernen: {
    bestaetigung: 'keine',
    ort: 'seiten/Kartenmarken.tsx',
    grund:
      'Eine Marke ist ein Wort an der Karte; wieder anheften kostet an derselben Stelle einen Klick.',
  },
  MarkeUeberallEntfernen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Markenpflege.tsx',
  },
  MasseLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Massenvorgaenge.tsx',
  },
  NutzerLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Nutzerpflege.tsx',
  },
  NutzerbildEntfernen: {
    bestaetigung: 'keine',
    ort: 'seiten/Nutzerbildwahl.tsx',
    grund:
      'Das Bild lässt sich an derselben Stelle neu wählen; die Datei, aus der es kam, bleibt, wo sie lag.',
  },
  ObjektmarkeEntfernen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Objektmarken.tsx',
    grund:
      'Eine Marke ist ein Wort an einem Objekt; wieder anheften kostet an derselben Stelle einen Klick.',
  },
  PaketEntfernen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Paketplanansicht.tsx',
  },
  // **Die Frage sagt, wo die Datei danach liegt** — in der Sicherung neben
  // dem Paket; im Fenster zurückholen lässt sie sich nicht (FMT-02, #1691).
  PaketdateiEntfernen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Dateiaktionen.tsx',
  },
  PartieLoeschen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Partienverwaltung.tsx',
    grund:
      'Löschen darf nur einen Entwurf — keine Einwilligung, nicht beendet; alles andere weist die Domäne ab (domain.Loeschbar).',
  },
  PartieVergessen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Protokollfreigabe.tsx',
  },
  ProfilLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'seiten/Nutzerpflege.tsx',
  },
  ProfilbadgeEntfernen: {
    bestaetigung: 'keine',
    ort: 'seiten/Profilangaben.tsx',
    grund:
      'Ein Abzeichen wird aus derselben Liste wieder angeheftet, aus der es kam.',
  },
  PunktLoeschen: {
    bestaetigung: 'ruecknahme',
    ort: 'seiten/Symbol.tsx',
    beleg: ['haken/usePunktwege.ts', 'eintragen(neue, {'],
  },
  SammlungsfreigabeWiderrufen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Kontaktfreigaben.tsx',
    grund:
      'Widerrufen wird eine Erlaubnis, nicht die Sammlung; sie lässt sich an derselben Stelle neu erteilen.',
  },
  SchluesselwortLoeschen: {
    bestaetigung: 'ruecknahme',
    ort: 'seiten/Schluesselworteditor.tsx',
    beleg: ['../../internal/domain/verlauf.go', 'HandlungSchluesselwort'],
  },
  SchnittstelleWiderrufen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Schnittstellen.tsx',
    grund:
      'Genommen wird nur der eigene Block, und Einrichten an derselben Stelle legt ihn wieder an; was außerhalb steht, bleibt.',
  },
  // **Alle auf einmal, aus demselben Grund** — G08 (#2272): genommen werden
  // nur die eigenen Blöcke, und *Alles einrichten* daneben legt sie wieder an.
  SchnittstellenAlleWiderrufen: {
    bestaetigung: 'keine',
    ort: 'bausteine/Schnittstellen.tsx',
    grund:
      'Genommen werden nur die eigenen Blöcke in jedem Ziel, und Alles einrichten an derselben Stelle legt sie wieder an; was außerhalb steht, bleibt.',
  },
  SetLoeschen: {
    bestaetigung: 'namen',
    ort: 'seiten/Setloeschen.tsx',
    feld: 'loeschenFeld',
  },
  StiloptionZuruecksetzen: {
    bestaetigung: 'keine',
    ort: 'seiten/Stiloptionen.tsx',
    grund:
      'Zurück geht es auf die Vorgabe des Stils; der vorige Wert ist in derselben Auswahl wieder einzustellen.',
  },
  FensterbildEntfernen: {
    bestaetigung: 'keine',
    ort: 'gestaltungsraum.tsx',
    grund:
      'Danach liegt kein eigenes Bild mehr hinter der Oberfläche; es lässt sich neu wählen, und „Aus" zeigt es ohnehin nicht.',
  },
  TischbildEntfernen: {
    bestaetigung: 'keine',
    ort: 'seiten/Partietisch.tsx',
    grund:
      'Danach gilt wieder eine der mitgelieferten Tischflächen; das eigene Bild lässt sich neu wählen.',
  },
  TokenAlleWiderrufen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Zugriffstoken.tsx',
  },
  TokenLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Tokenzeile.tsx',
  },
  TurnierLoeschen: {
    bestaetigung: 'rueckfrage',
    ort: 'bausteine/Turnierwerkbank.tsx',
  },
  ZugangLoeschen: { bestaetigung: 'rueckfrage', ort: 'bausteine/Zugaenge.tsx' },
  Zuruecksetzen: {
    bestaetigung: 'namen',
    ort: 'seiten/Ruecksetzung.tsx',
    feld: 'nameTippen',
  },
}

// kuerzelstellen sind die Dateien, in denen ein [Aktionsknopf] sein Kürzel
// ankündigt — mit den Läufen, die es drücken: je Lauf die Datei und eine
// Marke, die in ihr steht. **Eine Kette hat mehrere Glieder**: ⌘Z im
// Symboleditor geht vom Fenster über den Rahmen zur offenen Zeichnung, und
// jedes Glied hat seinen eigenen Lauf.
//
// **Nicht angekündigt wird**, wo das Kürzel etwas anderes trifft als der
// Knopf: ⌘S sichert das offene Set und nicht die Zeile der Settabelle, ⌘N
// öffnet das Formular für ein neues Set und legt keines an, ⌘Z nimmt den
// letzten Zug zurück und nicht einen bestimmten Vorgang im Verlauf. **Der
// Kopfknopf *Neues Set* der Setliste tut, was ⌘N tut** (UXR-04, #1912), und
// kündigt es deshalb an — *Set anlegen* im Assistenten nicht.
export const kuerzelstellen: Readonly<
  Record<
    string,
    { aktion: Aktion; laeufe: readonly (readonly [string, string])[] }
  >
> = {
  // **Seit UXR-06 (#1914) ein Eintrag im Menü „Diese Karte"** — vorher der
  // Knopf in `Kartenablage.tsx`.
  'seiten/Kartenhandlungen.tsx': {
    aktion: 'kopieren',
    laeufe: [['seiten/kartenablage.test.tsx', "['c', 'AblageKopieren']"]],
  },
  'seiten/Sets.tsx': {
    aktion: 'anlegen',
    laeufe: [['menueablage.test.tsx', "ausloesen(horcher, 'menue:set-neu')"]],
  },
  'seiten/Symbolwerkzeugleiste.tsx': {
    aktion: 'zuruecknehmen',
    laeufe: [
      ['rahmen/tastenkuerzel.test.tsx', "{ key: 'z', metaKey: true }"],
      ['anwendungruecknahme.test.tsx', 'trifft die offene Zeichnung'],
    ],
  },
}

// objektkuerzelstellen sind die Läufe, die ein Kürzel der Objekthandlungen
// drücken — AP-7 (#2364), nachgetragen nach Befund 4 des Zweitblicks über
// Phase 6 von 0.9.0: `kuerzel` an einem Eintrag ist dieselbe Zusage wie
// `mitKuerzel`, und eine Zusage braucht einen Lauf. Je Kürzel der Test, der
// es im Fenster nachstellt, und die Probe, die es am nativen Fenster drückt.
export const objektkuerzelstellen: Readonly<
  Record<string, readonly (readonly [string, string])[]>
> = {
  'befehl+return': [
    ['objekthandlungstasten.test.tsx', 'metaKey: true'],
    [
      '../../scripts/huellenprobe/ap5kontext09.mjs',
      "p.taste('Enter', 'befehl')",
    ],
  ],
  'befehl+d': [
    ['objekthandlungstasten.test.tsx', "key: 'd', metaKey: true"],
    ['../../scripts/huellenprobe/ap5kontext09.mjs', "p.taste('d', 'befehl')"],
  ],
  'befehl+backspace': [
    ['objekthandlungstasten.test.tsx', "key: 'Backspace'"],
    [
      '../../scripts/huellenprobe/ap5kontext09.mjs',
      "p.taste('Backspace', 'befehl')",
    ],
  ],
  'bild auf': [
    ['haken/useKartenblaettern.test.tsx', "taste('PageUp')"],
    ['../../scripts/huellenprobe/ap5kontext09.mjs', "p.taste('PageUp')"],
  ],
  'bild ab': [
    ['haken/useKartenblaettern.test.tsx', "taste('PageDown')"],
    ['../../scripts/huellenprobe/ap5kontext09.mjs', "p.taste('PageDown')"],
  ],
}
