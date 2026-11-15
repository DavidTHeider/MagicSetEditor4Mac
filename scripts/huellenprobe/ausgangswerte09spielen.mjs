// **Ausgangswerte Spielen am nativen Fenster** — P0
// (#2336),
// Befund C-22 aus dem Audit zur Produktionsreife 0.9.0.
//
// Die Wege S1 bis S11 aus Audit C (Abschnitt 3.1), mit Maus und — wo die
// Pflichtliste es verlangt — mit Tastatur, gezählt nach `huellenwege.mjs`
// und abgelegt als `docs/belege/ausgangswerte09-spielen.json` (Vertrag:
// `scripts/ausgangswege.mjs`). S12 und S13 brauchen zwei Fenster und stehen
// in `ausgangswerte09netz.mjs`, der Handel in `ausgangswerte09handeln.mjs`.
//
// # Was gezählt wird
//
// **Eine Handlung ist, was ein Mensch einmal tut**: ein Klick, ein Kürzel,
// ein getippter Wert, eine Wahl in einer Liste — und **jedes ⇥**: auf dem
// Tastaturweg ist der Weg zum Ziel Teil der Arbeit. Die Wahl in einer
// Auswahlliste setzt die Probe über die Seite und zählt sie als **eine**
// Handlung, auch auf dem Tastaturweg. **Ein Ziel, das ⇥ in 200 Drücken nicht
// erreicht, hat keinen Tastaturweg** und steht als `FEHLT` mit dieser Zahl;
// die Tabstellen eines erreichten stehen in der Zeile.
//
// # Der Bestand
//
// **Selbst geschrieben und künstlich**: `testdaten/mtgjson/tisch/TSP.json`
// — zwei Grundländer, drei Kreaturen, ein Spontanzauber, alle Namen
// erfunden —, eingelesen über `mse mtgjson-lesen`, dazu zwei Decks. Was vor
// einem Weg angelegt wird, damit es ihn gibt — eine Partie, ihr Beginn, eine
// Kreatur auf dem Spielfeld —, entsteht über `mse` und zählt nicht.
//
// # Was es nicht belegt
//
// **Menschenzeit, Fehlgriffe, Lesen und Suchen** — die Probe kennt den Weg.
// Den Weg mit dem Vorleseprogramm. Und zwei Geräte.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { ziehe } from '../huellenrufbuch.mjs'
import { anfangen, waehlen, waehlenMitTasten, namensfeldLeeren, sichtwert, begonnen, deckwahl, kartenknopf, knopf, namensfeld, kennungen, heisst, hinzu, bisZug, meldungen, schliessen, tastenweg, sichtwahl, stand, tabZu, zaehle, zonenfeld, zugpunkte, zumTisch } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/ausgangswerte09-spielen.json')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// Die Karten des Sets in der Reihenfolge, in der `mse karten 1` sie nennt.
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }

export default async function ausgangswerte09spielen(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  // **Deck 1** trägt jede Karte genau so oft, dass die Starthand von sieben
  // Karten alle Sorten hält; **Deck 2** ist das Gegenüber.
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [[karte.waldboden, 2], [karte.moosbaer, 1], [karte.glutlaeufer, 1], [karte.funkenregen, 1], [karte.rindenwall, 2]]) p.mse(ort, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [[karte.bergpfad, 16], [karte.glutlaeufer, 4], [karte.rindenwall, 4]]) p.mse(ort, 'deck', 'legen', '2', `--karte=${k}`, String(n))

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz('S1 — eine Partie anfangen, aus dem Bereich, mit der Maus')
  let vorher = JSON.parse(p.mse(ort, 'partien', '--json')).partien.length
  await z.beginne('S1', 'bereich', 'maus')
  await z.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await z.text('Partie anfangen')
  await p.wartetext('Wer sitzt am Tisch', 20000)
  for (const [n, name, deck] of [[1, 'Ada', '1'], [2, 'Bo', '2']]) {
    await z.klickAuf(namensfeld(n), `das Feld »Name am Tisch ${String(n)}«`)
    await p.seite(namensfeldLeeren(n))
    await z.tippe(name)
    await z.wahl(deckwahl(n - 1, deck))
  }
  await z.klickAuf(anfangen, '»Partie anfangen« unter der Anlage')
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 30000)
  await z.ende({ wirkung: JSON.parse(p.mse(ort, 'partien', '--json')).partien.length === vorher + 1 })
  await p.bild('s1-maus-tisch')

  p.notiz('S1 — eine Partie anfangen, aus dem Bereich, mit der Tastatur')
  await schliessen(p)
  await p.taste('1', 'befehl')
  await p.ruhe()
  vorher = JSON.parse(p.mse(ort, 'partien', '--json')).partien.length
  await z.beginne('S1', 'bereich', 'tastatur')
  await z.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  let erreicht = await tabZu(p, z, knopf('Partie anfangen'))
  if (erreicht) {
    await z.taste('Enter')
    await p.wartetext('Wer sitzt am Tisch', 20000)
    for (const [n, name, deck] of [[1, 'Ada', '1'], [2, 'Bo', '2']]) {
      erreicht = erreicht && await tabZu(p, z, namensfeld(n))
      if (erreicht) await p.seite(namensfeldLeeren(n)).then(() => z.tippe(name))
      erreicht = erreicht && await tabZu(p, z, `[...document.querySelectorAll('section[aria-label="Wer sitzt am Tisch"] select')][${String(n - 1)}]`)
      if (erreicht) await z.wahl(deckwahl(n - 1, deck))
    }
    erreicht = erreicht && await tabZu(p, z, anfangen)
    if (erreicht) {
      await z.taste('Enter')
      await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 30000).catch(() => false)
    }
  }
  if (erreicht) await z.ende({ wirkung: JSON.parse(p.mse(ort, 'partien', '--json')).partien.length === vorher + 1 })
  else z.ohneGang('S1', 'bereich', 'tastatur', 'FEHLT', 'ein Ziel des Wegs ist in 200 ⇥ nicht zu erreichen')

  p.notiz('S1 — eine Partie anfangen, aus dem Deck, mit der Maus')
  await schliessen(p)
  await p.taste('1', 'befehl')
  await p.ruhe()
  vorher = JSON.parse(p.mse(ort, 'partien', '--json')).partien.length
  const vorDeck = kennungen(p, ort)
  await z.beginne('S1', 'deck', 'maus')
  await z.taste('4', 'befehl')
  await p.warte('das Deck steht zur Wahl', () => p.seite(`return !!${knopf('Öffnen')} || [...document.querySelectorAll('button')].some((b) => (b.getAttribute('aria-label') ?? '') === 'Öffnen: Waldschritt')`), 20000)
  await z.klickAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Öffnen: Waldschritt')", '»Öffnen: Waldschritt«')
  await p.warte('»Mit diesem Deck spielen« steht', () => p.seite(`return !!${knopf('Mit diesem Deck spielen')}`), 20000)
  await z.text('Mit diesem Deck spielen')
  await p.wartetext('Wer sitzt am Tisch', 20000)
  for (const [n, name] of [[1, 'Ada'], [2, 'Bo']]) {
    await z.klickAuf(namensfeld(n), `das Feld »Name am Tisch ${String(n)}«`)
    await p.seite(namensfeldLeeren(n))
    await z.tippe(name)
  }
  await z.wahl(deckwahl(1, '2'))
  await z.klickAuf(anfangen, '»Partie anfangen« unter der Anlage')
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 30000)
  const ausDeck = hinzu(p, ort, vorDeck)
  await z.ende({ wirkung: JSON.parse(p.mse(ort, 'partien', '--json')).partien.length === vorher + 1 })

  // **Die Sitze einer im Fenster angelegten Partie heißen `sitz1` und
  // `sitz2`** (`sitzkennung` in `Sitzliste.tsx`), nicht wie die über `mse`.
  p.notiz('S2 — der Beginn im Hot-Seat bis Zug 1, mit der Maus')
  await p.warte('der Tisch trägt seine Zonenmarken', () => p.seite(`return !!${zonenfeld('hand', 'sitz1')}`), 30000)
  await z.beginne('S2', 'tisch', 'maus')
  for (const sitz of ['sitz1', 'sitz2']) {
    // **Seit P6 (#2374) folgt die Sicht dem Beginn, gezogen ist beim Anfangen** — bedingte Griffe.
    const fremd = sitz === 'sitz2' && (await p.seite(sichtwert)) !== 'sitz2'; z.bedingt('Sicht wählen', fremd); if (fremd) await z.wahl(sichtwahl('sitz2'))
    const ziehen = await p.seite(`return !!${knopf('Starthand ziehen')}`); z.bedingt('Starthand ziehen', ziehen)
    if (ziehen) await z.text('Starthand ziehen')
    await p.warte('»Hand behalten« steht', () => p.seite(`return !!${knopf('Hand behalten')}`), 15000)
    await z.text('Hand behalten')
    await p.warte('behalten', () => stand(p, ort, ausDeck, 'sitz1').sitze.find((s) => s.kennung === sitz)?.eroeffnung === 'behalten', 15000)
  }
  await z.wahl(sichtwahl('sitz1'))
  await z.ende({ wirkung: stand(p, ort, ausDeck, 'sitz1').zug.nummer === 1 && !(await p.seite(`return !!${knopf('Starthand ziehen')}`)) })
  await p.bild('s2-maus-beginn')

  p.notiz('S2 — der Beginn im Hot-Seat, mit der Tastatur')
  const ohneBeginn = begonnen(p, ort, '1', '2', true, ['Eda', 'Finn'])
  await schliessen(p)
  await zumTisch(p, 'Eda · Finn', 'Eda')
  await z.beginne('S2', 'tisch', 'tastatur')
  erreicht = true
  for (const sitz of ['ada', 'bo']) {
    if (sitz === 'bo') {
      erreicht = erreicht && await tabZu(p, z, "[...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Sicht für'))?.querySelector('select')")
      if (erreicht) await z.wahl(sichtwahl('bo'))
    }
    erreicht = erreicht && await tabZu(p, z, knopf('Starthand ziehen'))
    if (erreicht) await z.taste('Enter')
    erreicht = erreicht && await p.warte('»Hand behalten« steht', () => p.seite(`return !!${knopf('Hand behalten')}`), 15000).then(() => true, () => false)
    erreicht = erreicht && await tabZu(p, z, knopf('Hand behalten'))
    if (erreicht) await z.taste('Enter')
  }
  // **Beide Sitze behalten, Zug 1 begonnen** (Befund 4 des Zweitblicks über
  // Phase 0): „nicht mehr ausstehend“ hieß schon nach dem Ziehen grün.
  if (erreicht) await z.ende({ wirkung: stand(p, ort, ohneBeginn).sitze.every((s) => s.eroeffnung === 'behalten') && stand(p, ort, ohneBeginn).zug.nummer === 1 })
  else z.ohneGang('S2', 'tisch', 'tastatur', 'FEHLT', 'ein Knopf des Beginns ist in 200 ⇥ nicht zu erreichen')

  // **Für S3 bis S11 eine Partie mit gespieltem Beginn**: Adas Starthand
  // hält alle sieben Karten ihres Decks.
  const partie = begonnen(p, ort, '1', '2', false, ['Ida', 'Jo'])
  // **Nach jeder Vorbereitung über `mse` wird der Tisch neu aufgeschlagen**:
  // ein offener Tisch liest einen fremden Schreiber nicht nach, und der
  // nächste Befehl prallte als Konflikt ab (Befund dieses Laufs, behoben mit
  // #2388 — `tischfremd09`; das Aufschlagen bleibt, damit die Zählung
  // dieselbe ist).
  const neuAufschlagen = async () => {
    await schliessen(p)
    await zumTisch(p, 'Ida · Jo', 'Ida')
  }
  p.pruefe('die Partie steht in Idas erster Hauptphase', bisZug(p, ort, partie, (zug) => zug.phase === 'hauptVorKampf'))
  await schliessen(p)
  await zumTisch(p, 'Ida · Jo', 'Ida')
  await p.warte('Adas Hand steht', () => p.seite(`return !!${kartenknopf('Waldboden', 'Hand')}`), 20000)
  await p.bild('s3-tisch-vorher')

  p.notiz('S3 — ein Land aus der Hand spielen, mit der Maus über »Als Land spielen«')
  let land = zaehle(p, ort, partie, 'Waldboden', 'tisch')
  await z.beginne('S3', 'tisch', 'maus')
  await z.klickAuf(kartenknopf('Waldboden', 'Hand'), 'Waldboden auf der Hand', waehlen)
  await p.warte('»Als Land spielen« steht', () => p.seite(`return !!${knopf('Als Land spielen')}`), 10000)
  await z.text('Als Land spielen')
  await p.warte('das Land liegt', () => zaehle(p, ort, partie, 'Waldboden', 'tisch') > land, 10000).catch(() => false)
  await z.ende({ wirkung: zaehle(p, ort, partie, 'Waldboden', 'tisch') === land + 1 })

  p.notiz('S3 — ein Land aus der Hand ziehen, mit der Maus (der freie Weg)')
  land = zaehle(p, ort, partie, 'Waldboden', 'tisch')
  await z.beginne('S3', 'tisch-ziehen', 'maus')
  const [von, nach] = await zugpunkte(p, kartenknopf('Waldboden', 'Hand'), zonenfeld('tisch', 'ada'))
  if (von && nach) {
    await ziehe(p, [von, { x: (von.x + nach.x) / 2, y: (von.y + nach.y) / 2 }, nach])
    await z.wahl('return true')
    await p.bild('s3-nach-ziehen')
    await p.warte('das Land liegt', () => zaehle(p, ort, partie, 'Waldboden', 'tisch') > land, 10000).catch(() => false)
    p.zeige(`Meldungen: ${JSON.stringify(await meldungen(p))}`)
    await z.ende({ wirkung: zaehle(p, ort, partie, 'Waldboden', 'tisch') === land + 1 })
  } else {
    z.ohneGang('S3', 'tisch-ziehen', 'maus', 'NOT_RUN', 'Karte oder Zonenmarke nicht gefunden')
  }

  p.notiz('S5 — tappen, mit der Maus')
  await z.beginne('S5', 'tisch', 'maus')
  await z.klickAuf(kartenknopf('Waldboden', 'Schlachtfeld'), 'Waldboden auf dem Spielfeld', waehlen)
  await p.warte('»Tappen« steht', () => p.seite(`return !!${knopf('Tappen')}`), 10000)
  await z.text('Tappen')
  await p.warte('getappt', () => stand(p, ort, partie).objekte.some((o) => heisst(o, 'Waldboden') && o.zone === 'tisch' && o.getappt), 10000).catch(() => false)
  await z.ende({ wirkung: stand(p, ort, partie).objekte.some((o) => heisst(o, 'Waldboden') && o.zone === 'tisch' && o.getappt) })

  p.notiz('S6 — eine Karte auf den Friedhof, mit der Maus über »Schieben nach …«')
  let begraben = zaehle(p, ort, partie, 'Rindenwall', 'friedhof')
  await z.beginne('S6', 'tisch', 'maus')
  await z.klickAuf(kartenknopf('Rindenwall', 'Hand'), 'Rindenwall auf der Hand', waehlen)
  await p.warte('»Schieben nach« steht', () => p.seite(`return !!${knopf('Schieben nach')}`), 10000)
  await z.text('Schieben nach')
  await p.warte('das Formular steht', () => p.seite("return !!document.querySelector('[role=group][aria-label=\"Schieben nach\"] select')"), 10000)
  await z.wahl("const ziel = document.querySelector('[role=group][aria-label=\"Schieben nach\"] select'); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(ziel, 'friedhof'); ziel.dispatchEvent(new Event('change', { bubbles: true })); return true")
  await z.text('Stellen')
  await p.warte('im Friedhof', () => zaehle(p, ort, partie, 'Rindenwall', 'friedhof') > begraben, 10000).catch(() => false)
  await z.ende({ wirkung: zaehle(p, ort, partie, 'Rindenwall', 'friedhof') === begraben + 1 })

  p.notiz('S6 — eine Karte auf den Friedhof ziehen')
  begraben = zaehle(p, ort, partie, 'Rindenwall', 'friedhof')
  await z.beginne('S6', 'tisch-ziehen', 'maus')
  const [karteZwei, friedhof] = await zugpunkte(p, kartenknopf('Rindenwall', 'Hand'), zonenfeld('friedhof', 'ada'))
  if (karteZwei && friedhof) {
    await ziehe(p, [karteZwei, { x: (karteZwei.x + friedhof.x) / 2, y: (karteZwei.y + friedhof.y) / 2 }, friedhof])
    await z.wahl('return true')
    await p.warte('im Friedhof', () => zaehle(p, ort, partie, 'Rindenwall', 'friedhof') > begraben, 10000).catch(() => false)
    await z.ende({ wirkung: zaehle(p, ort, partie, 'Rindenwall', 'friedhof') === begraben + 1 })
  } else {
    z.ohneGang('S6', 'tisch-ziehen', 'maus', 'NOT_RUN', 'Karte oder Zonenmarke nicht gefunden')
  }

  p.notiz('S7 — eigenes Leben um 3 senken')
  const leben = (sitz) => stand(p, ort, partie).sitze.find((s) => s.kennung === sitz).leben
  let lebenVorher = leben('ada')
  const lebensfeld = "document.querySelector('[role=group][aria-label=\"Leben\"] input')"
  // **Geleert wie mit ⌘A**: ein Mensch markiert den Vorschlag und tippt
  // darüber — die Probe leert ihn, und das zählt nicht als eigene Handlung.
  const leeren = `const e = ${lebensfeld}; e.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`
  await z.beginne('S7', 'tisch', 'maus')
  await z.text('Leben')
  await p.warte('das Formular steht', () => p.seite(`return !!${lebensfeld}`), 10000)
  await z.klickAuf(lebensfeld, 'das Feld »um«')
  await p.seite(leeren)
  await z.tippe('-3')
  await z.text('Stellen')
  await p.warte('Leben gesunken', () => leben('ada') < lebenVorher, 10000).catch(() => false)
  await z.ende({ wirkung: leben('ada') === lebenVorher - 3 })

  p.notiz('S8 — das Leben des Gegenübers im Hot-Seat')
  lebenVorher = leben('bo')
  await z.beginne('S8', 'tisch', 'maus')
  await z.wahl(sichtwahl('bo'))
  await schlafe(500)
  await z.text('Leben')
  await p.warte('das Formular steht', () => p.seite(`return !!${lebensfeld}`), 10000)
  await z.klickAuf(lebensfeld, 'das Feld »um«')
  await p.seite(leeren)
  await z.tippe('-3')
  await z.text('Stellen')
  await p.warte('Bos Leben gesunken', () => leben('bo') < lebenVorher, 10000).catch(() => false)
  await z.wahl(sichtwahl('ada'))
  await z.ende({ wirkung: leben('bo') === lebenVorher - 3 })
  await schlafe(500)

  p.notiz('S4 — einen Zauber ohne Ziel wirken und verrechnen lassen')
  // **Das Mana liegt bereit**, über `mse` und ungezählt: Funkenregen kostet
  // R, und Idas Land gibt G (CR 601.2h, Fassung 2026-09-25). Ein Mensch
  // tappte dafür ein Gebirge — das Deck hat keins.
  p.mse(ort, 'partie-befehl', partie, 'ada', 'mana-s4', 'spieler.mana', JSON.stringify({ art: 'R', um: 1 }))
  await neuAufschlagen()
  const zauberVorher = zaehle(p, ort, partie, 'Funkenregen', 'friedhof')
  await z.beginne('S4', 'tisch', 'maus')
  await z.klickAuf(kartenknopf('Funkenregen', 'Hand'), 'Funkenregen auf der Hand', waehlen)
  await p.warte('»Wirken« steht', () => p.seite(`return !!${knopf('Wirken')}`), 10000)
  await z.text('Wirken')
  await p.warte('»Stellen« steht', () => p.seite(`return !!${knopf('Stellen')}`), 10000)
  await z.text('Stellen')
  await schlafe(600)
  p.zeige(`nach »Stellen«: ${JSON.stringify(await meldungen(p))} · Stapel: ${String(zaehle(p, ort, partie, 'Funkenregen', 'stapel'))}`)
  await p.bild('s4-nach-stellen')
  for (const sitz of ['ada', 'bo']) {
    if (sitz === 'bo') await z.wahl(sichtwahl('bo'))
    // **Ohne »Passen« scheitert der Weg**, statt eine Handlung still
    // auszulassen (Befund 3 des Nachlaufs über Phase 0).
    await p.warte('»Passen« steht', () => p.seite(`return !!${knopf('Passen')}`), 10000)
    await z.text('Passen')
    await schlafe(400)
  }
  await z.wahl(sichtwahl('ada'))
  await p.warte('verrechnet', () => zaehle(p, ort, partie, 'Funkenregen', 'friedhof') > zauberVorher, 10000).catch(() => false)
  p.zeige(`Meldungen: ${JSON.stringify(await meldungen(p))} · Zug: ${JSON.stringify(stand(p, ort, partie).zug).slice(0, 200)}`)
  await z.ende({ wirkung: zaehle(p, ort, partie, 'Funkenregen', 'friedhof') === zauberVorher + 1 })
  await p.bild('s4-nach-verrechnen')

  // **Für S9 zwei Kreaturen, schon in Idas erstem Zug** — eine Kreatur greift
  // erst an, wenn sie seit Beginn des Zugs unter Kontrolle steht (CR 302.6,
  // Fassung 2026-09-25); gelegt wird über `mse`, und es zählt nicht.
  const schiebe = (name, zone, aus = 'hand') => {
    const ding = stand(p, ort, partie).objekte.find((o) => heisst(o, name) && o.zone === aus)
    if (ding) p.mse(ort, 'partie-befehl', partie, 'ada', `${zone}-${ding.kennung}`, 'objekt.schieben', JSON.stringify({ objekt: ding.kennung, zone }))
  }
  for (const name of ['Moosbär', 'Glutläufer']) schiebe(name, 'tisch')
  // **Erst wenn das Fenster sie zeigt**: ein Befehl gegen den alten Stand
  // wird als Konflikt abgewiesen (»rechnet gegen die Fassung …«).
  await neuAufschlagen()
  await p.ruhe()

  p.notiz('S10 — den Zug weiterschalten')
  // **Die Wirkung ist ein anderer Zugstand**: »Zug beenden« springt in den
  // Aufräumschritt (CR 724.1, Fassung 2026-09-25), und der nächste Zug
  // beginnt erst danach.
  const zugVorher = JSON.stringify(stand(p, ort, partie).zug)
  await z.beginne('S10', 'tisch', 'maus')
  await p.warte('»Zug beenden« steht', () => p.seite(`return !!${knopf('Zug beenden')}`), 10000).catch(() => false)
  await z.text('Zug beenden')
  await p.warte('ein anderer Zugstand', () => JSON.stringify(stand(p, ort, partie).zug) !== zugVorher, 10000).catch(() => false)
  p.zeige(`Meldungen: ${JSON.stringify(await meldungen(p))} · Zug: ${JSON.stringify(stand(p, ort, partie).zug).slice(0, 200)}`)
  await z.ende({ wirkung: JSON.stringify(stand(p, ort, partie).zug) !== zugVorher })

  // **S9 in Idas drittem Zug** — Jos Zug läuft über `mse` durch. Idas
  // Bibliothek ist nach der Starthand leer, und wer aus einer leeren zieht,
  // verliert (CR 704.5b, Fassung 2026-09-25); deshalb geht vorher eine
  // Karte aus dem Friedhof zurück.
  schiebe('Rindenwall', 'bibliothek', 'friedhof')
  p.pruefe('Idas dritter Zug steht in der Hauptphase', bisZug(p, ort, partie, (zug) => zug.aktiv === 'ada' && zug.nummer >= 3 && zug.phase === 'hauptVorKampf', 80))

  // **Die Tastaturwege S3, S5 und S6 ebenfalls hier** — in Idas erstem Zug
  // ist das Land schon gespielt (CR 305.2, Fassung 2026-09-25). Ein
  // Waldboden geht über `mse` zurück auf die Hand, ungezählt.
  schiebe('Waldboden', 'hand', 'tisch')
  await neuAufschlagen()
  const getappt = () => stand(p, ort, partie).objekte.filter((o) => heisst(o, 'Waldboden') && o.zone === 'tisch' && o.getappt).length
  const friedhofVorher = zaehle(p, ort, partie, 'Rindenwall', 'friedhof')
  const landVorher = zaehle(p, ort, partie, 'Waldboden', 'tisch')
  const zonenliste = "document.querySelector('[role=group][aria-label=\"Schieben nach\"] select')"
  const zielwahl = { wo: zonenliste, wahl: `const ziel = ${zonenliste}; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(ziel, 'friedhof'); ziel.dispatchEvent(new Event('change', { bubbles: true })); return true` }
  p.notiz('S3, S5 und S6 — mit der Tastatur')
  await tastenweg(p, z, 'S3', [{ was: 'Waldboden auf der Hand', tab: kartenknopf('Waldboden', 'Hand') }, ...waehlenMitTasten, ['»Als Land spielen«', knopf('Als Land spielen')]], () => zaehle(p, ort, partie, 'Waldboden', 'tisch') === landVorher + 1)
  const getapptVorher = getappt()
  await tastenweg(p, z, 'S5', [{ was: 'Waldboden auf dem Spielfeld', tab: kartenknopf('Waldboden', 'Schlachtfeld') }, ...waehlenMitTasten, ['»Tappen«', knopf('Tappen')]], () => getappt() === getapptVorher + 1)
  await tastenweg(p, z, 'S6', [{ was: 'Rindenwall auf der Hand', tab: kartenknopf('Rindenwall', 'Hand') }, ...waehlenMitTasten, ['»Schieben nach …«', knopf('Schieben nach')], zielwahl, ['»Stellen«', knopf('Stellen')]], () => zaehle(p, ort, partie, 'Rindenwall', 'friedhof') === friedhofVorher + 1)

  p.pruefe('Idas dritter Zug steht bei den Angreifern', bisZug(p, ort, partie, (zug) => zug.aktiv === 'ada' && zug.nummer >= 3 && zug.schritt === 'angreifer', 80))
  await neuAufschlagen()
  p.notiz('S9 — zwei Kreaturen angreifen lassen')
  await z.beginne('S9', 'tisch', 'maus')
  await p.warte('»Angreifer erklären« steht', () => p.seite(`return !!${knopf('Angreifer erklären')}`), 10000).catch(() => false)
  if (await p.seite(`return !!${knopf('Angreifer erklären')}`)) {
    await z.text('Angreifer erklären')
    // **Die Zeilen der beiden Kreaturen** — das Formular nannte jede eigene
    // Bleibende, auch die Länder, und ließ den Kern abweisen (Befund dieses
    // Laufs, behoben mit #2390 — `angreifer09`). Gewählt wird weiter nach
    // dem Namen.
    const zeile = (name) => `[...document.querySelectorAll('[role=group][aria-label="Angreifer erklären"] select')].find((s) => (s.getAttribute('aria-label') ?? '').startsWith(${JSON.stringify(name)}))`
    await p.warte('das Formular steht', () => p.seite(`return !!${zeile('Glutläufer')}`), 10000).catch(() => false)
    for (const name of ['Moosbär', 'Glutläufer']) {
      await z.wahl(`const s = ${zeile(name)}; const o = [...s.options].find((x) => x.value !== ''); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, o?.value ?? ''); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
    }
    await z.text('Stellen')
    await p.warte('angegriffen', () => stand(p, ort, partie).zug.angriffe.length >= 2, 10000).catch(() => false)
    p.zeige(`Meldungen: ${JSON.stringify(await meldungen(p))}`)
    await z.ende({ wirkung: stand(p, ort, partie).zug.angriffe.length >= 2 })
  } else {
    z.ohneGang('S9', 'tisch', 'maus', 'NOT_RUN', 'im Zugstand dieser Probe ist kein Angriff angeboten')
  }

  await z.beginne('S10', 'tisch', 'tastatur')
  if (await tabZu(p, z, knopf('Passen'))) {
    const vorPass = JSON.stringify(stand(p, ort, partie).zug)
    await z.taste('Enter')
    await p.warte('gepasst', () => JSON.stringify(stand(p, ort, partie).zug) !== vorPass, 10000).catch(() => false)
    await z.ende({ wirkung: JSON.stringify(stand(p, ort, partie).zug) !== vorPass })
  } else {
    z.ohneGang('S10', 'tisch', 'tastatur', 'FEHLT', '»Passen« ist in 200 ⇥ nicht zu erreichen')
  }

  p.notiz('S11 — den letzten Zug zurücknehmen')
  z.ohneGang('S11', 'tisch', 'maus', 'FEHLT', 'es gibt keinen Befehl, der einen Spielzug zurücknimmt (P8, #2376)')

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
  p.zeige(`Datensatz: ${ziel}`)
  p.pruefe('⌘Q beendet', await p.beende())
}
