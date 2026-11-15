// **Welche Nutzlast ein Spielbefehl braucht** — Werkstück zu
// #1623.
//
// # Wofür es diese Tabelle gibt
//
// **Weil der Kern achtunddreißig Befehle kennt und das Fenster elf davon
// stellen konnte.** Gezählt am 2026-09-17: `objekt.wirken`, `objekt.land`,
// beide Kampfbefehle, alle acht Handlungen mit Schlüsselwort, Leben, Gift,
// Mana und drei aus G1 selbst hatten keine Bedienung. §16 war dem Wortlaut
// nach erfüllt — die Bindung `PartieBefehl` nimmt jede Art —, und dem Sinn
// nach nicht: an drei Flächen kann ein Mensch die Art **tippen**, an der
// vierten konnte niemand sie **anklicken**.
//
// # Warum eine Tabelle und nicht achtunddreißig Bauteile
//
// **Weil die Unterschiede in der Nutzlast liegen und nicht im Aussehen.**
// Ein Befehl nennt eine Zahl, eine Zone, ein zweites Objekt oder ein Paar;
// daraus folgt das Bedienelement. Achtunddreißig handgeschriebene Formen
// liefen mit dem Kern auseinander, sobald einer dazukommt — und genau das
// ist geschehen.
//
// **Und ein Wächter hält sie vollständig**:
// `internal/werkzeuge/befehlsformen_test.go` liest diese Datei gegen
// `service.Sandkastenbefehle()`.

import { istGeordnet } from './zonenordnung'

// Feldart sagt, womit ein Wert eingegeben wird.
export type Feldart =
  | 'zahl'
  | 'text'
  | 'zone'
  | 'stelle'
  | 'sitz'
  | 'manaart'
  | 'ruf'
  | 'objektwahl'
  | 'zielwahl'
  | 'ausloeser'
  | 'spielstein'
  | 'angriffe'
  | 'bloecke'

// Befehlsfeld ist ein Schlüssel der Nutzlast samt seiner Eingabeart.
export interface Befehlsfeld {
  // name ist der Schlüssel, unter dem der Kern ihn liest.
  name: string

  art: Feldart

  // vorgabe steht im Feld, bevor jemand etwas eingibt.
  //
  // **Leer heißt: keine.** Eine Vorgabe ist eine Aussage darüber, was
  // gewöhnlich gemeint ist — `spieler.leben` mit `-1` trifft öfter zu als
  // mit `0`, und `0` täte gar nichts.
  vorgabe?: string

  // nurWenn sagt, wann dieses Feld überhaupt eine Frage ist.
  //
  // **Weil manche Felder von einem anderen abhängen** — Befund **F04**
  // (#1748):
  // die Stelle beim Schieben bedeutet nur etwas, wenn die **Zielzone**
  // geordnet ist. Stünde sie immer da, mit der Vorgabe `oben` und ohne
  // neutralen Wert, wäre der Auslieferungszustand des Formulars ein Fehler —
  // die Vorgabe der Zone ist `hand`.
  //
  // **Ein Feld, das nicht gefragt wird, geht auch nicht hinaus.** Das ist
  // nicht bloß Kosmetik, sondern die zweite Hälfte: `nutzlastAus` übergeht
  // es, sonst schickte die Vorgabe weiter eine Stelle, die niemand sieht
  // (S05.NAT-02 — wer von der Bibliothek auf die Hand umstellt, nimmt die
  // vorher gewählte Stelle aus dem Auftrag).
  //
  // **Es bekommt einen Nachschlag und keine Wertetabelle**, und das ist
  // nicht Geschmack: `sichtbareFelder` legt vorher die Vorgaben unter, also
  // hat **jedes** Feld einen Wert. Eine Tabelle zwänge jeden Nachschlag zu
  // einem `?? ''` für einen Fall, den es nicht gibt — ein Zweig, den kein
  // Lauf betritt, und die Abdeckung fiel dafür auf 99,98 %.
  nurWenn?: (wert: (name: string) => string) => boolean
}

// Gruppe sagt, worauf ein Befehl zielt — und damit, wo er im Fenster steht.
//
// **Nach Gegenstand und nicht nach Werkstück.** Wer eine Karte in der Hand
// hat, sucht „was kann ich mit dieser Karte tun" und nicht „was kam mit
// M7".
export type Gruppe = 'karte' | 'zone' | 'sitz' | 'kampf' | 'zug' | 'beginn'

// Zustand ist eine Bedingung am gewählten Objekt.
//
// **Warum das Fenster sie kennt und nicht der Kern**: `Sandkasten.Erlaubt`
// rechnet je **Sitz** (CR 117.1) und kennt kein gewähltes Objekt. „Tappen"
// an einer getappten Karte wäre trotzdem ein Knopf, der immer abgewiesen
// wird — und eine Einladung, die man zurücknimmt, ist keine.
//
// **Vier und nicht mehr.** Was darüber hinausgeht — ob eine Kreatur
// angreifen darf, ob ein Land noch gespielt werden kann — hängt an Regeln,
// die der Kern führt; sie hier nachzubauen hieße, sie zweimal zu haben (§2).
//
// **`gehaeuft` kam mit dem Haufen dazu**
// (#1756).
// Es liest wie `haengt` ein **Feld des Objekts** und keine Regel: ob etwas in
// einem Haufen liegt, steht an ihm dran.
export type Zustand = 'enttappt' | 'getappt' | 'haengt' | 'gehaeuft'

// Befehlsform ist, was eine Befehlsart zum Stellen braucht.
export interface Befehlsform {
  art: string
  gruppe: Gruppe

  // nurBei sagt, an welchem Zustand dieser Befehl überhaupt erscheint.
  nurBei?: Zustand

  // felder sind die **zusätzlichen** Schlüssel der Nutzlast.
  //
  // **Ohne `objekt`**: bei der Gruppe `karte` ist das die gewählte Karte,
  // und sie wird eingesetzt statt abgefragt. Wer sie hier führte, zeigte
  // ein Eingabefeld für etwas, das schon dasteht.
  felder: Befehlsfeld[]
}

// nurObjekt ist die Form der Befehle, die nichts außer der Karte brauchen.
function nurObjekt(art: string): Befehlsform {
  return { art, gruppe: 'karte', felder: [] }
}

// befehlsformen führt jede Art aus `service.Sandkastenbefehle()`.
//
// **Die Reihenfolge ist die des Tisches und nicht die des Kerns**: was man
// oft tut, steht oben.
export const befehlsformen: Befehlsform[] = [
  // ---- Mit der gewählten Karte -------------------------------------------
  { art: 'objekt.tappen', gruppe: 'karte', nurBei: 'enttappt', felder: [] },
  { art: 'objekt.enttappen', gruppe: 'karte', nurBei: 'getappt', felder: [] },
  nurObjekt('objekt.umdrehen'),

  {
    art: 'objekt.schieben',
    gruppe: 'karte',
    felder: [
      { name: 'zone', art: 'zone', vorgabe: 'hand' },

      // **Nur in eine geordnete Zone** — CR 400.5, und Befund F04.
      {
        name: 'stelle',
        art: 'stelle',
        vorgabe: 'oben',
        nurWenn: (wert) => istGeordnet(wert('zone')),
      },
    ],
  },

  // **Eine Geste, die der Kern deutet** — P3 (#2371): ein Land wird
  // gespielt, ein Zauber gewirkt, im Sandkasten geschoben. Der Weg der
  // Tastatur zu dem, was am Tisch das Ziehen tut.
  {
    art: 'objekt.ziehen',
    gruppe: 'karte',
    felder: [
      { name: 'zone', art: 'zone', vorgabe: 'tisch' },
      { name: 'sitz', art: 'sitz' },
    ],
  },

  {
    art: 'objekt.zaehler',
    gruppe: 'karte',
    felder: [
      { name: 'art', art: 'text', vorgabe: '+1/+1' },
      { name: 'menge', art: 'zahl', vorgabe: '1' },
    ],
  },

  {
    art: 'objekt.anhaengen',
    gruppe: 'karte',
    felder: [{ name: 'an', art: 'objektwahl' }],
  },

  { art: 'objekt.loesen', gruppe: 'karte', nurBei: 'haengt', felder: [] },

  // **Der Haufen** — CR 400.5, eine Anordnung und keine Regelbewegung
  // (#1756).
  //
  // **`objektwahl` wie beim Anhängen**: gelegt wird auf ein anderes Objekt,
  // und welche es gibt, weiß die Partie — eine Liste im Code böte eines an,
  // das weggeräumt ist.
  {
    art: 'objekt.haeufen',
    gruppe: 'karte',
    felder: [{ name: 'worauf', art: 'objektwahl' }],
  },

  {
    art: 'objekt.enthaeufen',
    gruppe: 'karte',
    nurBei: 'gehaeuft',
    felder: [],
  },

  {
    art: 'objekt.schaden',
    gruppe: 'karte',
    felder: [{ name: 'um', art: 'zahl', vorgabe: '1' }],
  },

  // **Ein Land tappt für sein Mana** — CR 305.6, die einzige Manafähigkeit,
  // die im Regelwerk steht statt auf der Karte.
  nurObjekt('objekt.mana'),
  nurObjekt('objekt.land'),

  {
    art: 'objekt.wirken',
    gruppe: 'karte',
    felder: [{ name: 'ziele', art: 'zielwahl' }],
  },

  nurObjekt('objekt.zerstoeren'),
  nurObjekt('objekt.opfern'),
  nurObjekt('objekt.exilieren'),
  nurObjekt('objekt.neutralisieren'),

  {
    art: 'objekt.kaempfen',
    gruppe: 'karte',
    felder: [{ name: 'gegen', art: 'objektwahl' }],
  },

  {
    art: 'objekt.beherrschen',
    gruppe: 'karte',
    felder: [{ name: 'wer', art: 'sitz' }],
  },

  nurObjekt('objekt.umwandeln'),

  {
    art: 'objekt.kopieren',
    gruppe: 'karte',
    felder: [{ name: 'vorlage', art: 'objektwahl' }],
  },
  nurObjekt('spieler.legende'),
  nurObjekt('spieler.heimholen'),
  nurObjekt('zug.abwerfen'),

  // ---- Mit einer Zone ----------------------------------------------------
  {
    art: 'zone.blick',
    gruppe: 'zone',
    felder: [{ name: 'zone', art: 'zone', vorgabe: 'bibliothek' }],
  },
  {
    art: 'zone.mischen',
    gruppe: 'zone',
    felder: [{ name: 'zone', art: 'zone', vorgabe: 'bibliothek' }],
  },
  {
    art: 'zone.mahlen',
    gruppe: 'zone',
    felder: [{ name: 'wie', art: 'zahl', vorgabe: '1' }],
  },
  {
    art: 'zone.hellsehen',
    gruppe: 'zone',
    felder: [{ name: 'wie', art: 'zahl', vorgabe: '1' }],
  },
  {
    art: 'zone.ueberwachen',
    gruppe: 'zone',
    felder: [{ name: 'wie', art: 'zahl', vorgabe: '1' }],
  },

  // ---- Der eigene Sitz ---------------------------------------------------
  {
    art: 'spieler.leben',
    gruppe: 'sitz',
    felder: [{ name: 'um', art: 'zahl', vorgabe: '-1' }],
  },
  {
    art: 'spieler.vergiftung',
    gruppe: 'sitz',
    felder: [{ name: 'um', art: 'zahl', vorgabe: '1' }],
  },
  {
    art: 'spieler.mana',
    gruppe: 'sitz',
    felder: [
      { name: 'art', art: 'manaart', vorgabe: 'C' },
      { name: 'um', art: 'zahl', vorgabe: '1' },
    ],
  },
  {
    art: 'spieler.muenzwurf',
    gruppe: 'sitz',
    felder: [{ name: 'ruf', art: 'ruf', vorgabe: '' }],
  },
  {
    art: 'spieler.wuerfeln',
    gruppe: 'sitz',
    felder: [
      { name: 'seiten', art: 'zahl', vorgabe: '6' },
      { name: 'wuerfel', art: 'zahl', vorgabe: '1' },
    ],
  },
  // **Die Kennung wird nicht abgefragt, sondern gebildet.** Sie muss in der
  // Partie eindeutig sein (CR 111.1 kennt sie nicht — sie ist eine Sache
  // dieses Programms), und niemand am Tisch möchte sich eine ausdenken.
  // **Die Bezeichnungen** — CR 725, CR 726, CR 728 (M20,
  // #1621).
  { art: 'spieler.monarch', gruppe: 'sitz', felder: [] },
  { art: 'spieler.initiative', gruppe: 'sitz', felder: [] },
  {
    art: 'spieler.strahlung',
    gruppe: 'sitz',
    felder: [{ name: 'um', art: 'zahl', vorgabe: '1' }],
  },

  {
    art: 'spieler.spielstein',
    gruppe: 'sitz',
    felder: [{ name: 'felder', art: 'spielstein' }],
  },
  { art: 'spieler.emblem', gruppe: 'sitz', felder: [] },

  // **Die Reihung wartender Fähigkeiten** — CR 603.3b (M9,
  // #1611).
  //
  // **Auch hier, obwohl der Tisch sie oben stellt.** Die Tabelle ist der
  // Vertrag, den `internal/werkzeuge/befehlsformen_test.go` hält — sie führt
  // **jede** Art des Kerns. Wo ein Befehl erscheint, entscheidet die Gruppe;
  // die Reihung steht zusätzlich in `Ausloeserreihung` über dem Tisch, weil
  // sie das Spiel anhält (derselbe Grund wie bei `spieler.legende`).
  {
    art: 'spieler.reihung',
    gruppe: 'sitz',
    felder: [{ name: 'kennung', art: 'ausloeser' }],
  },

  // ---- Der Kampf ---------------------------------------------------------
  {
    art: 'kampf.angreifer',
    gruppe: 'kampf',
    felder: [{ name: 'angriffe', art: 'angriffe' }],
  },
  {
    art: 'kampf.blocker',
    gruppe: 'kampf',
    felder: [{ name: 'bloecke', art: 'bloecke' }],
  },

  // ---- Der Zug -----------------------------------------------------------
  //
  // **Diese vier zeigt die `Tischsteuerung`** und nicht die
  // `Handlungsleiste`: sie stehen **über** dem Tisch, bei der Auskunft, wer
  // am Zug ist — dort, wo man liest, dass man dran ist. Sie stehen hier,
  // weil die Tabelle der Vertrag ist, den der Wächter hält; wo sie
  // erscheinen, entscheidet die Gruppe.
  { art: 'zug.passen', gruppe: 'zug', felder: [] },
  { art: 'zug.abkuerzen', gruppe: 'zug', felder: [] },
  { art: 'zug.beenden', gruppe: 'zug', felder: [] },
  { art: 'zug.aufgeben', gruppe: 'zug', felder: [] },

  // **Die Rücknahme des letzten eigenen Zuges** — P8 (#2376), Vermerk 0197.
  // Sie steht neben dem Tisch und hinter ⌘Z (`useTischruecknahme`).
  { art: 'spieler.zuruecknehmen', gruppe: 'zug', felder: [] },

  // ---- Der Beginn --------------------------------------------------------
  //
  // **Diese vier zeigt `Eroeffnung` über dem Tisch** (CR 103.5,
  // #1839)
  // und nicht die Handlungsleiste — sie halten das Spiel an wie die
  // Legendenwahl. **Die Nutzlast des Unterlegens (`objekte`) setzt die Tafel
  // selbst**: sie wählt aus der eigenen Hand, und dafür gibt es hier kein
  // Feld.
  { art: 'spieler.starthand', gruppe: 'beginn', felder: [] },
  { art: 'spieler.mulligan', gruppe: 'beginn', felder: [] },
  { art: 'spieler.unterlegen', gruppe: 'beginn', felder: [] },
  { art: 'spieler.behalten', gruppe: 'beginn', felder: [] },
]
