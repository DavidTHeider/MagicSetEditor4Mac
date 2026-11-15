// **Zwei Fenster, ein Tausch** — S12.NAT-04 aus dem Abschluss von Phase 6.11
// (#1672),
// die Handelshälfte zu `nat04`.
//
// Anna und Bert als zwei Programme auf diesem Mac, je mit eigenem Bestand,
// Bund und eigener Identität; beide horchen, jedes kennt die Adresse des
// anderen. **Eingerichtet wird über die Kommandozeile** — Identität,
// Kontakt, Zweck, wie in `handelsdraht_test.go` —, **gehandelt über die
// Fenster**: jeder Zug im Fenster dessen, der ihn geht. **Seit P13 bringt
// ein Zug selbst hinüber, was er gegangen ist** (#2381, Vermerk 0200) —
// „In den Ausgang legen" und „Jetzt übertragen" drückt die Probe nicht mehr,
// sie wartet auf „An … hinausgegangen."
//
// **Was sie fand, bei 100 % Abdeckung** — und was deshalb hier geprüft wird:
// dass ein Fenster **von selbst** zeigt, was über die Leitung kam
// (`Datenstand` zählt das Nebenher, 0017); dass die Gegenseite die Übergabe
// erst mit festgehaltenem Exemplar bestätigen kann und das Fenster den Weg
// nennt — seit P13 hält der Zug ein passendes Exemplar selbst fest, und
// ohne eines hält er an; und dass beide Fenster den Beleg des anderen ohne
// Klick zeigen (`Handelsbeleg`).
//
// **Und seit #2500 der Befund am Beleg**: vor dem Abschluss liegt in Berts
// Bestand ein abgewiesener Gegenbeleg von Anna — gelegt mit `sqlite3`, wie
// ihn die Post festhielte. Sein Fenster nennt ihn nach seinem Klick, und
// sobald Annas geprüfter Beleg kommt, weicht er.
//
// **Was sie nicht belegt**: zwei Geräte in zwei Netzen (#1998), dass ein echter,
// nicht passender Beleg über die Leitung ankommt (das hält
// `service/handelsgegenbelegbefund_test.go`), und dass ein
// Mensch die Stimme des Gegenübers erst einträgt, wenn er sie gesehen hat —
// „Gegenseite bestätigt" drückt hier die Probe, sobald drüben bestätigt ist.
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))

const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")

// zeile liest einen Wert aus der Ausgabe von `mse` (`schluessel<TAB>wert`).
function zeile(ausgabe, schluessel) {
  return ausgabe.split('\n').find((z) => z.startsWith(`${schluessel}\t`))?.split('\t')[1] ?? ''
}

// mitEingabe ruft `mse` mit einer Eingabe auf der Standardeingabe.
function mitEingabe(ort, eingabe, ...argumente) {
  return execFileSync(`${wurzel}.go/bin/mse`, argumente, {
    cwd: wurzel,
    input: eingabe,
    env: kindumgebung(ort),
  }).toString()
}

const vorgang = 'v-nat04'

export default async function nat04handel(p) {
  const bert = new p.constructor('nat04handel-bert')
  bert.zeilen = p.zeilen

  try {
    await zweiFenster(p, bert)
  } finally {
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}

async function zweiFenster(anna, bert) {
  anna.notiz('Zwei Bestände, einander bekannt — über die Kommandozeile')
  const annaOrt = anna.ort()
  const bertOrt = bert.ort()
  const wer = {}
  for (const [name, p, ort] of [['Anna', anna, annaOrt], ['Bert', bert, bertOrt]]) {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    const angelegt = p.mse(ort, 'peer-anlegen')
    wer[name] = { kennung: zeile(angelegt, 'kennung'), abdruck: zeile(angelegt, 'fingerabdruck') }
  }
  for (const [hier, ort, dort, name] of [[anna, annaOrt, wer.Bert, 'Bert'], [bert, bertOrt, wer.Anna, 'Anna']]) {
    hier.mse(ort, 'kontakt-aufnehmen', dort.kennung, name)
    hier.mse(ort, 'kontakt-bestaetigen', dort.kennung, dort.abdruck)
    hier.mse(ort, 'zweck-erlauben', dort.kennung, 'handel')
  }
  anna.pruefe('zwei Identitäten', wer.Anna.kennung !== '' && wer.Bert.kennung !== '' && wer.Anna.kennung !== wer.Bert.kennung)

  // **Bert eröffnet einen Tausch und bietet ihn an** — Anna gibt ihm eine
  // Karte. Das Eröffnen ist nicht, was hier geprüft wird.
  const positionen = JSON.stringify([
    { von: wer.Anna.kennung, printing: 'bolt-m10', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
  ])
  mitEingabe(bertOrt, positionen, 'handel', 'eroeffnen', vorgang, 'tausch', wer.Anna.kennung)
  bert.mse(bertOrt, 'handel', 'anbieten', vorgang)

  // **Anna legt die Karte, die sie hergibt, erst an, als der Zug danach
  // fragt** (`ohneDeckung`) — mit ihr hält der Zug sie selbst fest (P13).
  // Zwei passende Zeilen gibt es nicht: die Sammlung führt je Druck, Zustand,
  // Sprache und Ausführung eine.
  //
  // **Daneben ein Exemplar desselben Drucks in anderem Zustand** (#2492):
  // es darf weder der Zug nehmen noch die Wahl unter „Zusagen“ anbieten —
  // bis dahin ging es weg, und drüben kam die abgesprochene Zeile an.
  anna.mse(annaOrt, 'exemplar', 'bolt-m10', '1', 'excellent', 'de', 'nonfoil')

  anna.notiz('Zwei Programme, die einander erreichen')
  await anna.starte({ ort: annaOrt })
  await bert.starte({ ort: bertOrt })
  for (const [hier, dort, name] of [[anna, bert, 'Anna'], [bert, anna, 'Bert']]) {
    const horch = await bindung(hier, "b.HorchStarten('127.0.0.1:0')")
    const port = horch?.stand?.adresse?.split(':').pop()
    hier.pruefe(`${name} horcht`, Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
    const andere = name === 'Anna' ? wer.Anna : wer.Bert
    await bindung(dort, `b.PeerkontaktAdresse(${JSON.stringify(andere.kennung)}, '127.0.0.1:${port}')`)
  }

  // **Gehandelt wird im Fenster** — wer dran ist, sagt der Stand seines
  // Bestands, und das Fenster muss es von selbst zeigen.
  const parteien = [
    { name: 'Bert', p: bert, ort: bertOrt, gegenueber: 'Anna', gibt: false },
    { name: 'Anna', p: anna, ort: annaOrt, gegenueber: 'Bert', gibt: true },
  ]
  for (const partei of parteien) {
    await partei.p.aktiviere()
    await partei.p.taste('5', 'befehl')
  }

  let runden = 0
  let still = 0
  const gegangen = []
  while (runden < 16 && still < 2) {
    runden += 1
    let bewegt = false
    for (const partei of parteien) {
      const st = stand(partei)
      if (st.lage !== 'dran' || !st.naechster) continue
      bewegt = true
      anna.notiz(`${partei.name}: ${st.naechster} (Stand ${st.zustand})`)
      await partei.p.aktiviere()
      await waehle(partei)
      // **Von selbst, ohne neu zu laden** — das Fenster fragt alle zwei
      // Sekunden nach dem Datenstand (`useBestandsstand`); die Probe gibt
      // ihm drei Abstände.
      const zeigt = await bisZu(6000, () => partei.p.seite(`return (document.querySelector('[data-handelsassistent]')?.innerText ?? '').includes('Du bist dran')`))
      partei.p.pruefe(`${partei.name} sieht von selbst, dass sie dran ist`, zeigt)
      if (st.naechster === 'uebergeben' && partei.gibt && !partei.gehalten) {
        await ohneDeckung(partei)
        partei.p.mse(partei.ort, 'exemplar', 'bolt-m10', '1', 'nearMint', 'de', 'nonfoil')
        partei.gehalten = true
      }
      await gehe(partei, st.naechster)
      try {
        await partei.p.warte(`${partei.name} hat »${st.naechster}« gestellt`, () => {
          const neu = stand(partei)

          return neu.zustand !== st.zustand || neu.naechster !== st.naechster || neu.lage !== st.lage
        }, 15000)
      } catch (grund) {
        partei.p.zeige(`Meldungen ${partei.name}: ${JSON.stringify(await meldungen(partei.p))}`)
        partei.p.zeige(`Raum ${partei.name}: ${JSON.stringify(await partei.p.seite("return (document.querySelector('[data-handelsassistent]')?.innerText ?? '').slice(0, 1500)"))}`)
        await partei.p.bild(`${partei.name.toLowerCase()}-${st.naechster}-klemmt`)
        throw grund
      }
      gegangen.push(`${partei.name}:${st.naechster}`)
      await hinaus(partei)
      const offen = await meldungen(partei.p)
      if (offen.length > 0) partei.p.zeige(`Meldungen ${partei.name}: ${JSON.stringify(offen)}`)
    }
    if (!bewegt) {
      still += 1
      await new Promise((los) => setTimeout(los, 3000))
    } else {
      still = 0
    }
  }
  anna.zeige(`Gegangen: ${gegangen.join(' → ')}`)

  // **Was der Vertrag dem Menschen lässt** (T10, #2026): eine fremde
  // Abschlussstimme wird nicht nachgespielt — sie verlangt einen Nachweis an
  // **diesem** Fenster. Wer gesehen hat, dass drüben bestätigt ist, sagt es
  // mit „Gegenseite bestätigt"; danach geht der Beleg hinüber.
  anna.notiz('Der Abschluss: jede Seite bestätigt die andere')
  // **Ein abgewiesener Gegenbeleg in Berts Bestand** (#2500) — an seinem
  // eigenen Beleg, wie die Post ihn festhielte.
  execFileSync('sqlite3', ['-cmd', '.timeout 5000', bertOrt.bestand, `INSERT INTO handelsgegenbelegbefunde (profil_id, vorgang, von, abdruck, mangel, grund)
SELECT profil_id, vorgang, '${wer.Anna.kennung}', 'abdruck-probe', 'zeilen', 'Probe: die Zeilen weichen ab' FROM handelsbelege WHERE vorgang = '${vorgang}'`])
  const befundsatz = 'passt aber nicht zu deinem — die verbuchten Karten weichen ab.'
  for (const partei of parteien) {
    const st = stand(partei)
    partei.p.zeige(`vorher ${partei.name}: ${st.zustand} · ${st.lage} · ${st.naechster || '—'}`)
    if (st.zustand !== 'abgleichOffen') continue
    await partei.p.aktiviere()
    await partei.p.klickeAuf(schritt(4), 'Schritt 4 »Verbuchen«')
    const abschnitt = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Einen Handel verbuchen')"
    await partei.p.klickeAuf(`[...((${abschnitt})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.trim() === 'Gegenseite bestätigt')`, '»Gegenseite bestätigt«')
    await partei.p.warte(`${partei.name} hat die Gegenseite bestätigt`, () => stand(partei).zustand !== st.zustand, 15000)
    if (partei.name === 'Bert') {
      const genannt = await bisZu(8000, async () => ((await partei.p.seite(`return (${abschnitt})?.innerText ?? ''`)) ?? '').includes(befundsatz))
      partei.p.pruefe('Berts Fenster nennt den abgewiesenen Beleg von Anna mit Mangel und nächstem Schritt', genannt && ((await partei.p.seite(`return (${abschnitt})?.innerText ?? ''`)) ?? '').includes('Ein Beleg von Anna kam'), await partei.p.seite(`return ((${abschnitt})?.innerText ?? '').split('\\n').slice(-2).join(' · ')`))
      await partei.p.bild('bert-befund')
    }
    await hinaus(partei)
  }
  for (const partei of parteien) {
    const st = stand(partei)
    partei.p.pruefe(`${partei.name} ist beidseitig verbucht`, st.zustand === 'beidseitigVerbucht', `${st.zustand} · ${st.lage}`)
    await partei.p.aktiviere()
    await partei.p.klickeAuf(schritt(4), 'Schritt 4 »Verbuchen«')
    // **Von selbst, ohne einen Klick** — wer zuerst bestätigt hat, sah den
    // Beleg des anderen bis dahin nie (`Handelsbeleg`).
    const abschnitt = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Einen Handel verbuchen')"
    const beide = await bisZu(6000, async () => ((await partei.p.seite(`return (${abschnitt})?.innerText ?? ''`)) ?? '').includes('Beide haben unterschrieben'))
    partei.p.pruefe(`das Fenster ${partei.name} sagt von selbst, dass beide unterschrieben haben`, beide, await partei.p.seite(`return ((${abschnitt})?.innerText ?? '').split('\\n').slice(-3).join(' · ')`))
    partei.p.pruefe('und kein abgewiesener Beleg steht mehr da', !((await partei.p.seite(`return (${abschnitt})?.innerText ?? ''`)) ?? '').includes(befundsatz))
    const verbuchen = await partei.p.seite(`return [...((${abschnitt})?.querySelectorAll('button') ?? [])].some((b) => b.innerText.trim() === 'Verbuchen')`)
    partei.p.pruefe('und bietet nichts mehr zu verbuchen an', !verbuchen)
    await partei.p.bild(`${partei.name.toLowerCase()}-abschluss`)
  }
  // **Und im Bestand**: die Karte ging von Anna zu Bert.
  const sammlung = (p, ort) => p.mse(ort, '--json', 'besitz', 'bolt-m10').replace(/\s+/g, ' ').slice(0, 300)
  anna.zeige(`Annas Besitz: ${sammlung(anna, annaOrt)}`)
  bert.zeige(`Berts Besitz: ${sammlung(bert, bertOrt)}`)
  anna.pruefe('das Exemplar in anderem Zustand blieb bei Anna', sammlung(anna, annaOrt).includes('excellent'))
  bert.pruefe('bei Bert kam das abgesprochene an', sammlung(bert, bertOrt).includes('nearMint') && !sammlung(bert, bertOrt).includes('excellent'))
}

// bisZu fragt, bis `bedingung` wahr ist oder die Frist um — **ohne**
// abzubrechen: das Ergebnis ist ein Befund, kein Halt.
async function bisZu(frist, bedingung) {
  const ende = Date.now() + frist
  for (;;) {
    if (await bedingung()) return true
    if (Date.now() > ende) return false
    await new Promise((los) => setTimeout(los, 150))
  }
}

// ohneDeckung drückt den Zug der Übergabe, bevor ein Exemplar da ist — der
// Zug hält an, nennt den Weg, und der Stand bleibt.
async function ohneDeckung(partei) {
  const vorher = stand(partei)
  await partei.p.klickeAuf(uebergabeknopf, 'der Zug der Übergabe')
  const nennt = await bisZu(8000, async () => (await meldungen(partei.p)).some((m) => m.includes('Wähle es unter „Zusagen“')))
  partei.p.pruefe(`ohne passendes Exemplar sagt das Fenster ${partei.name}, was zuerst zu tun ist`, nennt, JSON.stringify(await meldungen(partei.p)))
  const nachher = stand(partei)
  partei.p.pruefe('und der Stand bleibt', nachher.zustand === vorher.zustand && nachher.lage === vorher.lage, `${nachher.zustand} · ${nachher.lage}`)
  // **Und „Zusagen“ bietet das Exemplar in anderem Zustand nicht an** (#2492).
  const text = await partei.p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")
  partei.p.pruefe(`unter „Zusagen“ steht bei ${partei.name} keine unpassende Zeile zur Wahl`, text.includes('Keine Zeile deiner Sammlung passt zu Zustand, Sprache und Ausführung dieser Zeile.') && !text.includes('Excellent'), text.split('\n').filter((z) => z.includes('Zeile') || z.includes('Excellent')).join(' · ').slice(0, 400))
  await partei.p.bild(`${partei.name.toLowerCase()}-ohne-deckung`)
}

// stand liest den Stand des Tauschs im Bestand einer Partei.
function stand(partei) {
  try {
    return JSON.parse(partei.p.mse(partei.ort, 'handel', 'stand', vorgang, '--json'))
  } catch {
    return { zustand: 'unbekannt', lage: '', naechster: '' }
  }
}

const zeileVon = (name) => `[...document.querySelectorAll('main li button')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(name)}))`

// waehle schlägt den Tausch in der Liste auf.
async function waehle(partei) {
  // **Von selbst in der Liste** — ein Tausch, der über die Leitung kam,
  // soll dastehen, ohne dass jemand die Seite neu öffnet.
  const ende = Date.now() + 5000
  let da = false
  while (!da && Date.now() < ende) {
    da = await partei.p.seite(`return !!(${zeileVon(partei.gegenueber)})`)
    if (!da) await new Promise((los) => setTimeout(los, 250))
  }
  partei.p.pruefe(`der Tausch mit ${partei.gegenueber} steht bei ${partei.name} von selbst in der Liste`, da)
  if (!da) {
    await partei.p.taste('1', 'befehl')
    await partei.p.taste('5', 'befehl')
  }
  await partei.p.klickeAuf(zeileVon(partei.gegenueber), `den Tausch mit ${partei.gegenueber}`)
  await partei.p.warte('der Tausch ist gewählt', () => partei.p.seite(`return (${zeileVon(partei.gegenueber)})?.getAttribute('aria-current') === 'true'`), 20000)
}

const schritt = (n) => `document.querySelector('[data-handelsassistent] nav li:nth-child(${n}) button')`

// uebergabeknopf ist der Zug der Übergabe — er heißt „Übergeben und
// verbuchen", wo die eigene Stimme die letzte ist, sonst „Übergabe
// bestätigen" (P13).
const uebergabeknopf = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && ['Übergeben und verbuchen', 'Übergabe bestätigen'].includes(b.innerText.trim()))"

// gehe stellt den nächsten Zug — am Knopf, der sagt, was er tut.
async function gehe(partei, naechster) {
  if (naechster === 'uebergeben') {
    await partei.p.klickeAuf(uebergabeknopf, 'der Zug der Übergabe')

    return
  }
  const knoepfe = { zustimmen: 'Zustimmen', vorbereiten: 'Vorbereiten und übergeben' }
  const buchung = { verbuchen: 'Verbuchen', abgleichen: 'Abgleich melden', bestaetigen: 'Abschluss bestätigen' }
  if (buchung[naechster]) {
    await partei.p.klickeAuf(schritt(4), 'Schritt 4 »Verbuchen«')
    const abschnitt = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Einen Handel verbuchen')"
    await partei.p.klickeAuf(`[...((${abschnitt})?.querySelectorAll('button') ?? [])].find((b) => b.innerText.trim() === ${JSON.stringify(buchung[naechster])})`, `»${buchung[naechster]}«`)

    return
  }
  await partei.p.klickeText(knoepfe[naechster] ?? naechster)
}

// hinaus wartet, bis der Zug sagt, dass er hinübergebracht hat — **auf die
// Wirkung am Fenster** (§3); ob es drüben ankam, sagt der Stand drüben.
async function hinaus(partei) {
  const satz = `An ${partei.gegenueber} hinausgegangen.`
  const gesagt = await bisZu(15000, () => partei.p.seite(`return [...document.querySelectorAll('[role=status]')].some((e) => e.checkVisibility() && e.innerText.includes(${JSON.stringify(satz)}))`))
  partei.p.pruefe(`${partei.name}: »${satz}«`, gesagt, JSON.stringify(await meldungen(partei.p)))
}
