// **Vorrichtungen für die Läufe über den Bereich Spielen** — Werkstück G1
// von Phase 5.5
// (#1458).
//
// **Eine Datei für die Vorrichtungen**, wie `peerkontaktprobe.ts` und aus
// demselben Grund: was in vier Läufen gleich ist, gehört an eine Stelle —
// sonst weicht die fünfte Kopie beim ersten neuen Feld ab.
import { befehlsformen } from '../spiel/befehlsformen'
import type {
  BefehlAntwort,
  Objekthandlung,
  Partie,
  PartieAntwort,
  Sitz,
  Spielobjekt,
  Zug,
} from '../apipartie'
import type { DeckDTO } from '../apidecks'
import type { Partielauf } from '../haken/usePartie'
import type { Texturlauf } from '../haken/useSpieltexturen'

// zweiDecks sind die zwei, ohne die es keine Partie gibt.
export function zweiDecks(): DeckDTO[] {
  return [einDeck('1', 'Aggro'), einDeck('2', 'Kontrolle')]
}

// einDeck gibt ein Deck, das für den Anfang reicht.
export function einDeck(id: string, name: string): DeckDTO {
  return {
    id,
    fassung: '1',
    name,
    spiel: 'magic',
    notiz: '',
    formatID: '',
    plaetze: [],
    hauptdeck: 60,
    sideboard: 0,
    herkunft: { art: '', quelle: '', quellProfil: '', uebernommen: '' },
  }
}

// einObjekt gibt eine Karte am Tisch.
export function einObjekt(teile: Partial<Spielobjekt> = {}): Spielobjekt {
  return {
    kennung: 'sitz1/k1/0',
    art: 'karte',
    beherrscher: 'sitz1',
    eigner: 'sitz1',
    zone: 'bibliothek',
    karte: '',
    name: '',
    karteID: '',
    bild: '',
    ordnung: 0,
    getappt: false,
    verdeckt: false,
    geheim: true,
    haengtAn: '',
    haufen: '',
    zaehler: {},

    // **`null` und nicht ein leerer Verbund** — die Vorgabe ist ein
    // geheimes Objekt, und das trägt keine Eigenschaften (M2, #1604).
    eigenschaften: null,
    ...teile,
  }
}

// einSitz gibt einen Platz am Tisch — mit **allen** Feldern.
//
// **Eine Stelle und nicht zehn** (§3): `Sitz` hat seit G9 sieben Felder, und
// jede von Hand gebaute Kopie wich beim nächsten ab. Wer eines ergänzt,
// ergänzt es hier.
export function einSitz(abweichend: Partial<Sitz> = {}): Sitz {
  return {
    kennung: 'sitz1',
    name: 'Anna',
    deckabdruck: 'abc',
    karten: 3,
    wartet: false,
    eingeladen: '',
    profilID: '',
    leben: 20,
    vergiftung: 0,
    manavorrat: '',
    strahlung: 0,
    commander: '',
    commandersteuer: 0,
    commanderschaden: {},
    leergezogen: false,
    eroeffnung: '',
    mulligans: 0,
    unterlegen: 0,
    ...abweichend,
  }
}

// einZug gibt den Zugstand, den eine Sandkastenpartie hat: keinen.
//
// **Null in `nummer` heißt: es geht nicht zugweise zu** (G9, #1592), und das
// ist der Stand jeder Partie aus G1 bis G8.
export function einZug(abweichend: Partial<Zug> = {}): Zug {
  return {
    nummer: 0,
    aktiv: '',
    starter: '',
    phase: '',
    schritt: '',
    prioritaet: '',
    abwurf: 0,
    gepasst: [],
    angriffe: [],
    bloecke: [],
    ...abweichend,
  }
}

// einePartie gibt einen Tisch mit zwei Sitzen.
export function einePartie(teile: Partial<Partie> = {}): Partie {
  return {
    id: 'p1',
    fassung: '1',
    spiel: 'magic',
    stand: 'laeuft',
    fuer: 'sitz1',
    wirt: 'sitz1',
    regelwerk: {
      kennung: 'sandkasten',
      fassung: 1,
      strenge: 'sandkasten',
      abgewaehlt: [],
    },
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna', deckabdruck: 'abc' }),
      einSitz({ kennung: 'sitz2', name: 'Bert', deckabdruck: 'def' }),
    ],
    zuschauer: [],
    objekte: [einObjekt()],
    mischpflichten: [],
    legendenwahlen: [],
    ausgeloest: [],
    ausloeserreihung: '',
    vorrat: [],
    zug: einZug(),
    ausgang: { sieger: '', unentschieden: false, ausgeschieden: [] },
    ...teile,
  }
}

// eineAntwort verpackt eine Partie so, wie die Naht sie gibt.
export function eineAntwort(
  teile: Partial<Partie> = {},
  erlaubt: string[] = ['objekt.tappen', 'objekt.schieben'],
): PartieAntwort {
  return {
    partie: einePartie(teile),
    erlaubt,
    handlungen: [],
    befunde: [],
    fehler: null,
  }
}

// einBefehlsausgang verpackt, was ein Spielbefehl hinterlässt.
export function einBefehlsausgang(
  privat: Record<string, unknown> = {},
): BefehlAntwort {
  return {
    fassung: '2',
    ereignisse: [
      {
        folge: '1',
        art: 'objekt.tappen',
        oeffentlich: { sitz: 'sitz1' },
        privat,
        kette: 'abcd',
      },
    ],
    fehler: null,
  }
}

// nichts ist der Weg, den ein Lauf nicht überschreibt.
//
// **Einer und nicht fünf.** Fünf gleiche Pfeilfunktionen sind fünf Funktionen,
// und wer nur eine davon drückt, lässt vier ungemessen — die Abdeckung des
// Frontends verlangt sie alle (§3). Die Frage, die sie beantworten, ist
// dieselbe: „tu nichts".
const nichts = () => undefined

// einLauf gibt einen Partielauf, dessen Wege sich zählen lassen.
//
// volleAuskunft gibt einem Objekt jede Kartenhandlung, aber keine Primär- und
// keine Zweithandlung — für Läufe über die Leiste, die prüfen, **was ein
// Knopf stellt**, und nicht, was der Kern anbietet (P4, #2372). Ein Klick
// wählt dann, wie vor P4.
export function volleAuskunft(kennung: string): Objekthandlung {
  return {
    objekt: kennung,
    handlungen: befehlsformen
      .filter((form) => form.gruppe === 'karte')
      .map((form) => form.art),
    primaer: '',
    zweit: '',
  }
}

// **Ein Doppel und kein zweiter Haken** (§3): was `usePartie` tut, misst
// `haken/partie.test.tsx`; hier geht es darum, **was** eine Schaltfläche
// ruft.
export function einLauf(teile: Partial<Partielauf> = {}): Partielauf {
  return {
    anfangen: nichts,
    ansehen: nichts,
    befehlen: nichts,
    wechseln: nichts,
    raeumen: nichts,
    laeuft: () => false,
    partie: einePartie(),
    wechselt: false,

    // **Alles erlaubt**, weil die Handlungsleiste sich daraus ergibt
    // (#1623):
    // ein leeres `erlaubt` zeigte keinen einzigen Knopf, und ein Lauf über
    // eine Schaltfläche prüfte dann die Vorrichtung statt das Bauteil.
    //
    // **Aus der Tabelle und nicht aus einer Aufzählung daneben** — die
    // zweite liefe beim nächsten Befehl auseinander.
    erlaubt: befehlsformen.map((form) => form.art),
    handlungen: [],
    wartend: 0,
    befunde: [],
    ereignisse: [],
    ereignisluecke: false,
    fehler: undefined,
    fehlerVorgang: undefined,
    ...teile,
  }
}

// element gibt einen Knoten als HTMLElement — oder scheitert laut.
//
// **Laut und nicht still**: ein `as HTMLElement` über einem `undefined`
// führte zu einem Fehler drei Zeilen weiter, der von etwas anderem spräche.
//
// **Der Satz steht im Wurf und ist trotzdem keiner für die Oberfläche** —
// deshalb die Ausnahme: dieser Helfer läuft nur in Läufen, und ein Fenster,
// in dem er erschiene, gibt es nicht.
export function element(knoten: Element | null | undefined): HTMLElement {
  if (!(knoten instanceof HTMLElement)) {
    // keinfehlertext: eine Vorrichtung für Läufe, nie im Fenster
    throw new Error('kein Element')
  }

  return knoten
}

// ruhenderLauf gibt einen Texturlauf, der nichts holt.
//
// **Diese Läufe prüfen Befehle und kein Bild.** Ein echter Lauf zöge die
// Wails-Fassade in jeden von ihnen — und was er belegte, stünde in keiner
// Zusicherung.
export function ruhenderLauf(): Texturlauf {
  return {
    belege: nichts,
    gib: nichts,
    adresse: () => '',
    zweiseitig: () => false,
    vorladen: nichts,
    raeume: nichts,
    befunde: [],
    fehler: undefined,
  }
}
