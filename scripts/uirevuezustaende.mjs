// **Der Vertrag des Dichtemaßes in befüllten Zuständen** — AP-13
// (#2339),
// Befund B-33 aus dem Audit zur Produktionsreife 0.9.0.
//
// # Wofür
//
// `uirevue` misst jede Seite mit dem Probeset der Saat — und dort hält R2.
// Im Alltag hält es nicht: ein offenes Deck und ein Set mit einem reichen
// Stil zeigten im Audit je rund 35 Bedienelemente im Bild gegen 20. **Ein
// Maß, das nur den leeren Zustand sieht, misst den Prospekt.** Deshalb
// nennt dieser Vertrag die Zustände, die gemessen sein müssen, und je
// Zustand den leeren Gegenpart — getrennt, damit kein Mittelwert beide
// verwischt.
//
// # Was er nicht hält
//
// **Ob die Zahlen stimmen.** Er hält, dass jeder Zustand in jeder Größe und
// jedem Erscheinungsbild **erreicht** und gemessen ist; dass die Seite die
// richtige war, zeigen die Bilder des Laufs.

// groessen sind die Fenster: das Maß der Abzüge aus dem Audit und das
// Mindestmaß des Fensters (`internal/fenster/bildschirm.go`).
export const groessen = [
  { name: '1440x838', breite: 1440, hoehe: 838 },
  { name: '1024x700', breite: 1024, hoehe: 700 },
]

export const erscheinungen = ['hell', 'dunkel']

// pflichtzustaende sind die Blicke aus B-33, je befüllt und leer — **mit der
// Seite, auf der sie stehen** (Befund 6 des Zweitblicks über Phase 0): ohne
// sie bestand ein Leerzustand auch auf jeder anderen Seite.
export const pflichtzustaende = [
  { name: 'karten-gross', befuellt: true, seite: 'karten', was: 'ein Set mit 3.788 künstlichen Karten im Stil M15, eine Karte gewählt' },
  { name: 'karten-leer', befuellt: false, seite: 'karten', was: 'dasselbe Spiel und derselbe Stil, ein Set ohne Karten' },
  { name: 'deck-offen', befuellt: true, seite: 'decks', was: 'ein Deck mit 60 Karten aufgeschlagen' },
  { name: 'decks-leer', befuellt: false, seite: 'decks', was: 'Decks ohne ein einziges Deck' },
  { name: 'sammlung-voll', befuellt: true, seite: 'sammlung', was: 'eine Sammlung mit 1.000 Exemplaren' },
  { name: 'sammlung-leer', befuellt: false, seite: 'sammlung', was: 'eine Sammlung ohne Exemplar' },
]

// pruefeZustaende nennt, was einem Satz fehlt — leer heißt vollständig.
export function pruefeZustaende(satz) {
  const fehler = []

  if (!/^[0-9a-f]{12,}$/.test(satz?.bauabdruck ?? '')) {
    fehler.push('der Satz nennt keinen Bauabdruck — ohne ihn ist er keinem Bündel zuzuordnen')
  }

  const zeilen = new Map((satz?.zeilen ?? []).map((z) => [`${z.groesse}/${z.erscheinung}/${z.zustand}`, z]))

  for (const groesse of groessen) {
    for (const erscheinung of erscheinungen) {
      for (const zustand of pflichtzustaende) {
        const k = `${groesse.name}/${erscheinung}/${zustand.name}`
        const zeile = zeilen.get(k)

        if (zeile === undefined) fehler.push(`${k}: fehlt`)
        else if (zeile.erreicht !== true) fehler.push(`${k}: nicht erreicht`)
        else if (zeile.seite !== zustand.seite) fehler.push(`${k}: auf »${String(zeile.seite)}« statt »${zustand.seite}«`)
        else if (!Number.isInteger(zeile.bedienImBild) || !Number.isInteger(zeile.hauptImBild)) fehler.push(`${k}: nicht gemessen`)
      }
    }
  }

  return fehler
}

// schluesselVon ist die Kennung einer Zeile: Größe, Bild und Zustand — oder,
// in `uirevue` über alle Seiten, Größe und Seite.
const schluesselVon = (z) => [z.groesse, z.erscheinung, z.zustand ?? z.ziel].filter((teil) => teil !== undefined).join('/')

// hauptgewichtBefunde nennt jede Sicht mit mehr als einem gefüllten
// Hauptknopf im Bild und jeden gewählten Umschalter, der gefüllt ist — AP-3
// (#2360),
// B-08. Leer heißt: R1 hält je Blick. **Auch für den leeren Zustand**: der
// Entwurf `Zustaende09` sagt „je Blick“, nicht „je befülltem Blick“.
//
// Was es nicht hält: ob der eine Knopf der richtige ist. Das sieht ein Mensch
// am Bild des Laufs.
export function hauptgewichtBefunde(satz) {
  const befunde = []

  for (const z of (satz?.zeilen ?? []).filter((x) => x.erreicht === true)) {
    const k = schluesselVon(z)

    if (!Array.isArray(z.hauptNamen) || !Array.isArray(z.gewaehltGefuellt)) {
      befunde.push(`${k}: ohne Namen gemessen`)
      continue
    }

    if (z.hauptImBild > 1) befunde.push(`${k}: ${String(z.hauptImBild)} gefüllt im Bild (${z.hauptNamen.join(', ')})`)
    if (z.gewaehltGefuellt.length > 0) befunde.push(`${k}: gewählt und gefüllt (${z.gewaehltGefuellt.join(', ')})`)
  }

  return befunde
}

// ohneZeit nimmt einem Namen Datum und Uhrzeit — der Eintrag auf *Start*
// nennt, wann ein Set zuletzt offen war, und das ist in jedem Lauf eine andere
// Minute. Was danach gleich heißt, ist dieselbe Handlung.
const ohneZeit = (name) => name.replace(/\d{1,2}\.\d{1,2}\.\d{4}(, \d{1,2}:\d{2})?/g, '‹Zeit›')

// verloreneHandlungen nennt je Zeile, was vorher an Handlungen dastand und
// nachher fehlt — AP-3: **null verlorene Handlungen**. Verglichen wird über
// Rolle und Namen, ob im Bild oder nicht; ein Knopf, der sein Gewicht
// abgibt, bleibt dieselbe Handlung.
//
// Was es nicht hält: ob eine Handlung **dieselbe Wirkung** behält, und eine
// Handlung, die nur umbenannt wurde, zählt als verloren — beides mit Absicht
// laut statt still. **Nur Datum und Uhrzeit im Namen zählen nicht**
// (`ohneZeit`).
export function verloreneHandlungen(vorher, nachher) {
  const danach = new Map((nachher?.zeilen ?? []).map((z) => [schluesselVon(z), z]))
  const befunde = []

  for (const z of (vorher?.zeilen ?? []).filter((x) => x.erreicht === true)) {
    const k = schluesselVon(z)
    const n = danach.get(k)

    if (!Array.isArray(z.handlungen)) {
      befunde.push(`${k}: vorher ohne Handlungen gemessen`)
    } else if (n?.erreicht !== true) {
      befunde.push(`${k}: nachher nicht erreicht`)
    } else if (!Array.isArray(n.handlungen)) {
      befunde.push(`${k}: nachher ohne Handlungen gemessen`)
    } else {
      const bleibt = new Set(n.handlungen.map(ohneZeit))
      for (const h of z.handlungen.filter((x) => !bleibt.has(ohneZeit(x)))) befunde.push(`${k}: ${h} fehlt`)
    }
  }

  return befunde
}
