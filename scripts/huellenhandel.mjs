// **Die Griffe am Handelsbereich für die Hüllenprobe** — gezogen aus
// `huellenprobe/ausgangswerte09handeln.mjs` (P0,
// #2336),
// damit das Drehbuch unter der Zeilengrenze aus §2 bleibt und die Messungen
// der Phase 8 dieselben Griffe benutzen statt eigener.
//
// **Nicht unter `huellenprobe/`**: dort ist jede Datei ein Drehbuch.

// zeile liest einen Wert aus der Ausgabe von `mse` (`schluessel<TAB>wert`).
export const zeile = (ausgabe, schluessel) => ausgabe.split('\n').find((z) => z.startsWith(`${schluessel}\t`))?.split('\t')[1] ?? ''

// bindung ruft eine Bindung des Fensters — Einrichtung, kein Messwert.
export const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

// knopf findet einen sichtbaren Knopf nach Beschriftung oder Text.
export const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && ((b.getAttribute('aria-label') ?? '') === ${JSON.stringify(text)} || b.innerText.replace(/\\s+/g, ' ').trim() === ${JSON.stringify(text)}))`

// feld findet ein sichtbares Eingabefeld nach seiner Beschriftung.
export const feld = (name) => `[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && ((e.getAttribute('aria-label') ?? '') === ${JSON.stringify(name)} || (e.labels?.[0]?.innerText ?? '').trim() === ${JSON.stringify(name)}))`

// einfuegen setzt einen Text in ein Feld, wie es ein Einfügen täte.
export const einfuegen = (ausdruck, text) => `const e = ${ausdruck}; e.focus(); const setze = Object.getOwnPropertyDescriptor(e instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype, 'value').set; setze.call(e, ${JSON.stringify(text)}); e.dispatchEvent(new Event('input', { bubbles: true })); return true`

// handel liest die Handel eines Bestands über `mse`.
export const handel = (p, ort) => JSON.parse(p.mse(ort, 'handel', 'liste', '--json'))

// gekoppelt macht zwei Bestände einander bekannt — bestätigt, mit Erlaubnis
// zu handeln. Über die Kommandozeile, wie `nat04handel`: das Aufnehmen hat
// seinen eigenen Beleg (`g01fan`).
export function gekoppelt(p, orte) {
  const wer = {}
  for (const [name, ort] of Object.entries(orte)) {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
    const angelegt = p.mse(ort, 'peer-anlegen')
    wer[name] = { kennung: zeile(angelegt, 'kennung'), abdruck: zeile(angelegt, 'fingerabdruck') }
  }
  for (const [name, ort] of Object.entries(orte)) {
    for (const [anderer, dort] of Object.entries(wer)) {
      if (anderer === name) continue
      p.mse(ort, 'kontakt-aufnehmen', dort.kennung, anderer)
      p.mse(ort, 'kontakt-bestaetigen', dort.kennung, dort.abdruck)
      p.mse(ort, 'zweck-erlauben', dort.kennung, 'handel')
    }
  }

  return wer
}

// preisstelle ist eine Gegenstelle in diesem Prozess, die für jeden Druck
// einen Preis nennt — in der Gestalt, die `trade02` für Scryfall benutzt.
// Die Namen sind erfunden; aus dem Netz kommt nichts.
export async function preisstelle(createServer) {
  const stelle = createServer((anfrage, antwort) => {
    const id = anfrage.url.split('?')[0].split('/').pop()
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(JSON.stringify({
      object: 'card', id, name: `Probe ${id}`, lang: 'de', layout: 'normal', mana_cost: '{R}',
      type_line: 'Instant', oracle_text: '', rarity: 'common', set: 'p09', set_name: 'Probesatz',
      collector_number: '1', artist: 'Selbst Gemalt', scryfall_uri: 'https://beispiel.invalid/karte/1',
      prices: { eur: '1.50', eur_foil: '4.25', usd: '1.80', usd_foil: '5.10', tix: '0.03' },
    }))
  })
  await new Promise((fertig) => { stelle.listen(0, '127.0.0.1', fertig) })

  return { stelle, adresse: `http://127.0.0.1:${String(stelle.address().port)}` }
}

// vorgangsstand liest den Stand eines Handels in einem Bestand.
export function vorgangsstand(p, ort, vorgang) {
  try {
    return JSON.parse(p.mse(ort, 'handel', 'stand', vorgang, '--json'))
  } catch {
    return { zustand: 'unbekannt', lage: '', naechster: '' }
  }
}

// schritt ist der Knopf eines Schritts im Handelsassistenten.
export const schritt = (n) => `document.querySelector('[data-handelsassistent] nav li:nth-child(${String(n)}) button')`
const buchung = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Einen Handel verbuchen')"
const imAbschnitt = (abschnitt, text) => `[...((${abschnitt})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.trim() === ${JSON.stringify(text)})`
const assistent = "return document.querySelector('[data-handelsassistent]')?.innerText ?? ''"

// bisZu fragt, bis `bedingung` wahr ist oder die Frist um — ohne Abbruch.
export async function bisZu(frist, bedingung) {
  const ende = Date.now() + frist
  for (;;) {
    if (await bedingung()) return true
    if (Date.now() > ende) return false
    await new Promise((los) => setTimeout(los, 150))
  }
}

// waehle schlägt den Handel mit dem Gegenüber in der Liste auf, gezählt.
export async function waehle(partei) {
  const eintrag = `[...document.querySelectorAll('main li button')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(partei.gegenueber)}))`
  await bisZu(8000, () => partei.p.seite(`return !!(${eintrag})`))
  const gewaehlt = (await partei.p.seite(`return (${eintrag})?.getAttribute('aria-current')`)) === 'true'
  partei.z.bedingt(`${partei.name}: den Handel mit ${partei.gegenueber} wählen — er war ${gewaehlt ? 'schon' : 'nicht'} gewählt`, !gewaehlt)
  if (gewaehlt) return
  await partei.z.klickAuf(eintrag, `den Handel mit ${partei.gegenueber}`)
  await partei.p.warte('der Handel ist gewählt', () => partei.p.seite(`return (${eintrag})?.getAttribute('aria-current') === 'true'`), 20000)
}

// halte nimmt in Schritt 3 das eigene Exemplar: Zeile wählen, »Festhalten«.
export async function halte(partei, vorgang) {
  const feldsatz = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Welche Exemplare weggehen' && s.checkVisibility())?.querySelector('fieldset')"
  // **Nur wenn Schritt 3 nicht schon offen ist** — ein Klick auf den Schritt,
  // der dasteht, ist keine Handlung eines Menschen.
  const offen = Boolean(await partei.p.seite(`return !!(${feldsatz})`))
  partei.z.bedingt(`${partei.name}: Schritt 3 öffnen — er war ${offen ? 'schon' : 'nicht'} offen`, !offen)
  if (!offen) await partei.z.klickAuf(schritt(3), 'Schritt 3 »Zusagen«')
  await bisZu(5000, () => partei.p.seite(`return !!(${feldsatz})`))
  await partei.z.klickAuf(`(${feldsatz})?.querySelector('label')`, 'das eigene Exemplar')
  await partei.z.klickAuf(`[...((${feldsatz})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.trim() === 'Festhalten')`, '»Festhalten«')
  await partei.p.warte(`${partei.name} hat festgehalten`, () => partei.p.mse(partei.ort, '--json', 'handelszusagen', vorgang).includes('"exemplarId"'), 10000)
}

// gehe stellt den nächsten Schritt — am Knopf, der so heißt wie er.
export async function gehe(partei, naechster) {
  const knoepfe = { zustimmen: 'Zustimmen', vorbereiten: 'Vorbereiten', uebergeben: 'Übergabe bestätigen' }
  const gebucht = { verbuchen: 'Verbuchen', abgleichen: 'Abgleich melden', bestaetigen: 'Abschluss bestätigen' }
  if (gebucht[naechster]) {
    await partei.z.klickAuf(schritt(4), 'Schritt 4 »Verbuchen«')
    await partei.z.klickAuf(imAbschnitt(buchung, gebucht[naechster]), `»${gebucht[naechster]}«`)

    return
  }
  await partei.z.text(knoepfe[naechster] ?? naechster)
}

// bestaetigeGegenseite sagt »Gegenseite bestätigt« — die Stimme, die der
// Vertrag einem Menschen an diesem Fenster lässt (T10, #2026).
export async function bestaetigeGegenseite(partei) {
  await partei.z.klickAuf(schritt(4), 'Schritt 4 »Verbuchen«')
  await partei.z.klickAuf(imAbschnitt(buchung, 'Gegenseite bestätigt'), '»Gegenseite bestätigt«')
}

// uebertrage legt den Stand in den Ausgang und bringt ihn hinaus, gezählt:
// ohne diese zwei Klicks erfährt das Gegenüber nichts.
export async function uebertrage(partei) {
  await partei.z.klickAuf(schritt(4), 'Schritt 4 »Verbuchen«')
  await partei.p.wartetext('In den Ausgang legen', 10000)
  const vorher = ((await partei.p.seite(assistent)).match(/[0-9a-f]{32}/g) ?? []).length
  await partei.z.text('In den Ausgang legen')
  const gelegt = await bisZu(10000, async () => ((await partei.p.seite(assistent)).match(/[0-9a-f]{32}/g) ?? []).length > vorher)
  if (!gelegt) return
  await partei.z.text('Jetzt übertragen')
  await partei.p.warte(`${partei.name} hat übertragen`, async () => (await partei.p.seite(assistent)).includes('Hinausgegangen sind'), 15000)
}
