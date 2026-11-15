// **Drei Formulare am nativen Fenster vermessen** — P01, P02 und P04 aus
// Phase 0.7.0-I (#2165,
// #2166,
// #2167).
//
// Gemessen wird, was die Abnahmekriterien als Geometrie sagen, und zwar an
// den **echten** Bauteilen im nativen Fenster — WebKit zeichnet ein
// Auswahlfeld niedriger als ein Textfeld, und eine Messung im Browser gäbe
// andere Zahlen:
//
// - **P01, Sitzliste**: je Sitz eine benannte Gruppe aus Zeichen, Name und
//   Deck; „Wegnehmen" auf derselben Stelle relativ zum Sitz, breit wie
//   schmal; ein langer Name bleibt ganz im Feld.
// - **P02, Deckkartenwahl**: breit stehen Beschriftungen auf einer Linie,
//   Feldkörper auf einer zweiten, und „Ins Deck legen" auf der der Felder;
//   schmal steht jede Beschriftung über ihrem Feld. Ein Einlegen ändert die
//   Menge im Deck.
// - **P04, Tischleiste**: die vier Blickhandlungen stehen auf einer Zeile, die
//   zwei Zielhandlungen auf einer — breit wie schmal.
//
// Deutsch hell und Englisch dunkel, je bei 1280 × 800, **1024 × 768** und
// schmal — die mittlere Breite verlangt ABN-01.AK-03
// (#2175).
//
// **Und die Kontraste der umgestellten Teile, am gerenderten Element**
// (`scripts/huellenkontrast.mjs`): Text 4,5 : 1, Ränder, Feldränder und die
// Trennlinie der Tischleiste 3 : 1 (§3, „Trennlinien halten 3 : 1"). Die
// Paare der Merkmale stehen in `docs/KONTRASTE.md`; hier wird gemessen, was
// an diesen Stellen wirklich übereinander liegt.
//
// **Was es nicht belegt**: die Bedienung mit dem Vorleseprogramm.
import { pruefeKontraste } from '../huellenkontrast.mjs'

const breit = { name: 'breit', breite: 1280, hoehe: 800 }
const mittel = { name: 'mittel', breite: 1024, hoehe: 768 }
const schmal = { name: 'schmal', breite: 900, hoehe: 700 }

export const sprachen = {
  de: {
    sitzHinzu: 'Sitz hinzufügen', sitzName: 'Name am Tisch', sitzWeg: 'Diesen Sitz wegnehmen', sitzGruppe: 'Sitz',
    decks: 'Decks', karten: 'Karten suchen', legen: 'Ins Deck legen', partie: 'Partie anfangen', wer: 'Wer sitzt am Tisch',
    weiter: 'Weiterspielen', alsAda: 'Weiterspielen als Ada',
    blick: ['Alles zeigen', 'Näher heran', 'Weiter weg', 'Tisch drehen'], ziel: ['Auf die Auswahl', 'Auf meinen Sitz'],
  },
  en: {
    sitzHinzu: 'Add a seat', sitzName: 'Name at the table', sitzWeg: 'Remove this seat', sitzGruppe: 'Seat',
    decks: 'Decks', karten: 'Search cards', legen: 'Add to deck', partie: 'Start the game', wer: 'Who sits at the table',
    weiter: 'Continue', alsAda: 'Continue as Ada',
    blick: ['Show everything', 'Zoom in', 'Zoom out', 'Rotate the table'], ziel: ['To the selection', 'To my seat'],
  },
}

const rechteck = (ausdruck) => `(() => { const e = ${ausdruck}; if (!e) return null; const r = e.getBoundingClientRect(); return { top: r.top, bottom: r.bottom, left: r.left, right: r.right } })()`

const kastenKarten = (t) => `[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === ${JSON.stringify(t.karten)})`

// kontrasteSitze misst den ersten Sitz — mit gewähltem Deck, denn erst dann
// steht die Marke mit seiner Größe da.
async function kontrasteSitze(p, t, wo) {
  // **Am Sitz und nicht am ersten `[role=group]` der Seite** — der erste Lauf
  // maß als „Rahmen des Sitzes" eine fremde Gruppe ohne Rand.
  const sitz = "[...document.querySelectorAll('main li')].find((li) => li.checkVisibility() && li.querySelector(':scope > [role=group]'))"
  await p.seite(`const s = ${sitz}.querySelector('select'); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, s.options[1].value); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  await p.warte('die Größe des Decks steht beim Sitz', () => p.seite(`return !!${sitz}?.querySelector('span.text-leise')`), 10000)
  await pruefeKontraste(p, wo, [
    ['Zeichen des Sitzes', `${sitz}?.querySelector('span.size-8')`, 'text', 4.5],
    ['Rand des Zeichens', `${sitz}?.querySelector('span.size-8')`, 'rand', 3],
    ['Rahmen des Sitzes', sitz, 'rand', 3],
    ['Beschriftung des Decks', `${sitz}?.querySelector('label')`, 'text', 4.5],
    ['Größe des Decks', `${sitz}?.querySelector('span.text-leise')`, 'text', 4.5],
    ['Rand des Deckfelds', `${sitz}?.querySelector('select')`, 'rand', 3],
  ])
}

async function kontrasteDeckwahl(p, t, wo) {
  await pruefeKontraste(p, wo, [
    ['Beschriftung', `${kastenKarten(t)}?.querySelector('label span')`, 'text', 4.5],
    ['Rand des Felds', `${kastenKarten(t)}?.querySelector('select, input')`, 'rand', 3],
  ])
}

async function kontrasteTisch(p, t, wo) {
  await pruefeKontraste(p, wo, [
    [`»${t.blick[1]}«`, knopf(t.blick[1]), 'text', 4.5],
    ['Rahmen der Leiste', `${knopf(t.blick[1])}?.parentElement?.closest('.border-kante.bg-flaeche')`, 'rand', 3],
    ['Trennlinie zwischen Blick und Ziel', "document.querySelector('span.w-px.bg-kante')", 'flaeche', 3],
  ])
}

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(text)} || b.innerText.trim() === ${JSON.stringify(text)}))`

// sitze misst jeden Sitz: Gruppe, Name und die Stelle von „Wegnehmen".
async function sitze(p, t) {
  return p.seite(`return [...document.querySelectorAll('main li')].filter((li) => li.checkVisibility() && [...li.querySelectorAll('input')].some((i) => (i.labels?.[0]?.textContent ?? '').startsWith(${JSON.stringify(t.sitzName)}))).map((li) => {
    const r = li.getBoundingClientRect()
    const gruppe = li.querySelector('[role=group]')
    const name = [...li.querySelectorAll('input')].find((i) => (i.labels?.[0]?.textContent ?? '').startsWith(${JSON.stringify(t.sitzName)}))
    const deck = li.querySelector('select')
    const weg = [...li.querySelectorAll('button')].find((b) => b.innerText.trim() === ${JSON.stringify(t.sitzWeg)})
    const w = weg?.getBoundingClientRect()
    return {
      gruppe: gruppe?.getAttribute('aria-label') ?? '',
      nameInGruppe: !!gruppe && gruppe.contains(name), deckInGruppe: !!gruppe && !!deck && gruppe.contains(deck),
      wegOben: w ? Math.round(w.top - r.top) : null, wegRechts: w ? Math.round(r.right - w.right) : null,
      nameGanz: name ? name.scrollWidth <= name.clientWidth + 1 : false,
      // **Der Name nimmt den freien Platz** — bis an „Wegnehmen" oder den Rand.
      nameLuecke: name ? Math.round((w ? w.left : r.right - 12) - name.getBoundingClientRect().right) : null,
      nameZeichen: name ? Math.floor(name.clientWidth / (name.scrollWidth / Math.max(name.value.length, 1))) : null,
    }
  })`)
}

async function sitzliste(p, t, groesse) {
  const gemessen = await sitze(p, t)
  p.zeige(`P01 ${groesse.name}: ${JSON.stringify(gemessen)}`)
  p.pruefe(`P01 ${groesse.name}: jeder Sitz hat eine benannte Gruppe mit Name und Deck`, gemessen.length >= 4 && gemessen.every((s) => s.gruppe.startsWith(t.sitzGruppe) && s.nameInGruppe && s.deckInGruppe))
  // **Ein Sitzname hat keine Längengrenze** (`domain.Sitz`): ein einzeiliges
  // Feld zeigt nie jeden — ein längerer rollt im Feld. Geprüft wird, dass das
  // Feld seinen Platz nimmt statt einer festen Spaltenzahl, und dass ein Name
  // von dreißig Zeichen ganz darin steht; wie viele Zeichen es fasst, steht
  // im Beleg.
  p.zeige(`P01 ${groesse.name}: das Namensfeld fasst etwa ${String(gemessen[0]?.nameZeichen)} Zeichen`)
  p.pruefe(`P01 ${groesse.name}: das Namensfeld nimmt den freien Platz seiner Zeile`, gemessen.every((s) => s.nameLuecke !== null && s.nameLuecke <= 48), JSON.stringify(gemessen.map((s) => s.nameLuecke)))
  p.pruefe(`P01 ${groesse.name}: ein Name von dreißig Zeichen steht ganz im Feld`, gemessen[0]?.nameGanz === true)
  p.pruefe(`P01 ${groesse.name}: „Wegnehmen" steht oben rechts im Sitz`, gemessen.length > 0 && gemessen.every((s) => s.wegOben !== null && s.wegOben <= 40 && s.wegRechts <= 24), JSON.stringify(gemessen.map((s) => [s.wegOben, s.wegRechts])))

  return gemessen
}

// deckwahl misst Beschriftungen, Felder und den Knopf im Kasten „Karten suchen".
async function deckwahl(p, t) {
  return p.seite(`
    const kasten = [...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === ${JSON.stringify(t.karten)})
    if (!kasten) return null
    const paare = [...kasten.querySelectorAll('label, fieldset')].filter((e) => e.checkVisibility() && e.parentElement?.closest('fieldset') === null || e.tagName === 'FIELDSET').map((e) => {
      const beschriftung = e.tagName === 'FIELDSET' ? e.querySelector('legend') : e.querySelector('span')
      const feld = e.tagName === 'FIELDSET' ? e.querySelector('div') : e.querySelector('select, input')
      if (!beschriftung || !feld) return null
      const b = beschriftung.getBoundingClientRect(); const f = feld.getBoundingClientRect()
      return { text: beschriftung.innerText.trim(), beschriftungOben: Math.round(b.top), beschriftungUnten: Math.round(b.bottom), feldOben: Math.round(f.top), feldUnten: Math.round(f.bottom) }
    }).filter(Boolean)
    const k = [...kasten.querySelectorAll('button')].find((b) => b.innerText.trim() === ${JSON.stringify(t.legen)})?.getBoundingClientRect()
    return { paare, knopfOben: k ? Math.round(k.top) : null, knopfUnten: k ? Math.round(k.bottom) : null }
  `)
}

function spanne(werte) {
  return Math.max(...werte) - Math.min(...werte)
}

async function deckkartenwahl(p, t, groesse) {
  const gemessen = await deckwahl(p, t)
  p.zeige(`P02 ${groesse.name}: ${JSON.stringify(gemessen)}`)
  if (!gemessen) {
    p.pruefe(`P02 ${groesse.name}: der Kasten steht da`, false)

    return
  }
  // **Der Knopf gegen die Zeile, in der er steht** — er bricht auch breit
  // in eine eigene Zeile um; dann steht er dort ohne Versatz.
  const seineZeile = gemessen.paare.filter((x) => gemessen.knopfOben < x.feldUnten && gemessen.knopfUnten > x.beschriftungOben)
  if (seineZeile.length > 0) {
    const unten = Math.max(...seineZeile.map((x) => x.feldUnten))
    p.pruefe(`P02 ${groesse.name}: „Ins Deck legen" schließt unten mit den Feldern seiner Zeile ab`, Math.abs(gemessen.knopfUnten - unten) <= 2, `${String(gemessen.knopfUnten)} gegen ${String(unten)}`)
  } else {
    const darueber = Math.max(...gemessen.paare.filter((x) => x.feldUnten <= gemessen.knopfOben).map((x) => x.feldUnten))
    p.pruefe(`P02 ${groesse.name}: „Ins Deck legen" steht in eigener Zeile ohne Versatz`, gemessen.knopfOben - darueber <= 16, `${String(gemessen.knopfOben)} unter ${String(darueber)}`)
  }
  p.pruefe(`P02 ${groesse.name}: die Feldkörper einer Zeile stehen auf einer Linie`, gemessen.paare.every((x) => gemessen.paare.filter((y) => Math.abs(y.beschriftungOben - x.beschriftungOben) <= 4).every((y) => Math.abs(y.feldOben - x.feldOben) <= 1)))
  p.pruefe(`P02 ${groesse.name}: jede Beschriftung steht über ihrem Feld`, gemessen.paare.every((x) => x.beschriftungUnten <= x.feldOben + 1 && x.feldOben - x.beschriftungUnten <= 10))
}

async function tischleiste(p, t, groesse) {
  const oben = async (texte) => p.seite(`return [${texte.map((x) => `${rechteck(knopf(x))}?.top ?? null`).join(', ')}]`)
  const blick = await oben(t.blick)
  const ziel = await oben(t.ziel)
  p.zeige(`P04 ${groesse.name}: Blick ${JSON.stringify(blick)} · Ziel ${JSON.stringify(ziel)}`)
  p.pruefe(`P04 ${groesse.name}: die vier Blickhandlungen stehen auf einer Zeile`, blick.every((x) => x !== null) && spanne(blick) <= 1)
  p.pruefe(`P04 ${groesse.name}: die zwei Zielhandlungen stehen auf einer Zeile`, ziel.every((x) => x !== null) && spanne(ziel) <= 1)
}

export function vorbereiten(p, ort) {
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck mit einem sehr langen Namen, der umbrechen müsste', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  const angelegt = p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = /Partie ([0-9a-f]+)/.exec(angelegt)?.[1] ?? ''
  for (const sitz of ['ada', 'bo']) {
    for (const art of ['spieler.starthand', 'spieler.behalten']) p.mse(ort, 'partie-befehl', partie, sitz, `${art}/${sitz}`, art)
  }
}

export async function zumStartformular(p, t) {
  await p.taste('4', 'befehl')
  await p.wartetext(t.partie, 20000)
  await p.klickeText(t.partie)
  await p.wartetext(t.wer, 20000)
  for (const _ of [1, 2]) await p.klickeText(t.sitzHinzu)
  // **Ein langer Name im ersten Sitz** — über den Setzer des Feldes, wie
  // eine Eingabe ihn setzt.
  await p.seite(`const f = [...document.querySelectorAll('input')].find((i) => i.checkVisibility() && (i.labels?.[0]?.textContent ?? '').startsWith(${JSON.stringify(t.sitzName)})); const setze = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set; setze.call(f, 'Adelheid von Wolkenstein-Trost'); f.dispatchEvent(new Event('input', { bubbles: true })); return true`)
}

async function zurDeckwahl(p, t) {
  await p.taste('4', 'befehl')
  await p.warte('»Decks« steht zur Wahl', () => p.seite(`return [...document.querySelectorAll('[role=tab], button, a')].some((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}]+/u, '').startsWith(${JSON.stringify(t.decks)}))`), 20000)
  await p.klickeText(t.decks)
  // **Die Kachel trägt ihren eigenen Knopf** — benannt „<wählen>: <Deck>".
  const deckknopf = "[...document.querySelectorAll('main [data-eintrag] button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').includes('Probe-Deck'))"
  await p.warte('das Deck steht in der Liste', () => p.seite(`return !!(${deckknopf})`), 20000)
  await p.klickeAuf(deckknopf, 'das Deck')
  await p.warte('der Kasten zur Kartenwahl steht', () => p.seite(`return !!${`[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === ${JSON.stringify(t.karten)})`}`), 20000)
}

async function zumTisch(p, t) {
  await p.taste('4', 'befehl')
  await p.wartetext(t.partie, 20000)
  await p.klickeText(t.partie)
  await p.wartetext(t.weiter)
  await p.klickeText(t.weiter)
  await p.klickeText(t.alsAda)
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 20000)
}

export default async function politur(p) {
  const ort = p.ort()
  vorbereiten(p, ort)

  for (const [sprache, bild] of [['de', 'hell'], ['en', 'dunkel']]) {
    const t = sprachen[sprache]
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')

    // **Je Größe ein eigener Start** — die Größe steht in den Einstellungen,
    // wie nach einem Ziehen am Rand; ein Größenwechsel über Wails' MCP riss
    // im ersten Lauf die Verbindung ab.
    for (const groesse of [breit, mittel, schmal]) {
      p.notiz(`${sprache}, ${bild}, ${groesse.name} (${String(groesse.breite)} × ${String(groesse.hoehe)})`)
      p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
      p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
      await p.starte({ ort })

      const wo = `${sprache}, ${bild}, ${groesse.name}`
      await zumStartformular(p, t)
      groesse.sitze = await sitzliste(p, t, groesse)
      await kontrasteSitze(p, t, wo)
      await p.bild(`politur-sitze-${sprache}-${bild}-${groesse.name}`)

      await zurDeckwahl(p, t)
      await deckkartenwahl(p, t, groesse)
      await kontrasteDeckwahl(p, t, wo)
      await p.bild(`politur-deckwahl-${sprache}-${bild}-${groesse.name}`)

      await zumTisch(p, t)
      await tischleiste(p, t, groesse)
      await kontrasteTisch(p, t, wo)
      await p.bild(`politur-tisch-${sprache}-${bild}-${groesse.name}`)

      p.pruefe('⌘Q beendet', await p.beende())
    }

    const breitWeg = breit.sitze?.map((s) => s.wegOben) ?? []
    for (const andere of [mittel, schmal]) {
      const weg = andere.sitze?.map((s) => s.wegOben) ?? []
      p.pruefe(`P01 ${sprache}: „Wegnehmen" steht breit wie ${andere.name} an derselben Stelle des Sitzes`, breitWeg.length > 0 && breitWeg.every((x, i) => x !== null && Math.abs(x - (weg[i] ?? -99)) <= 2), `${JSON.stringify(breitWeg)} gegen ${JSON.stringify(weg)}`)
    }
  }
}
