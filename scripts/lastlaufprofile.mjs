// **Die Profile des Lastlaufs** — QA-08.AK-01
// (#1739).
//
// Je Profil **eine Frage**: Lesen, Schreiben, Rendern, Einfuhr, lange
// Aufträge, MCP über stdio und über HTTP getrennt, gemischte Klienten, zwei
// Identitäten nebeneinander, eine Überlast mit der Erholung danach und
// langsame Klienten. Jedes bekommt `k` — die laufenden Dienste, die Kennungen
// aus der Saat, Dauer und Zahl der Klienten — und gibt seine Proben je
// Gruppe zurück, dazu einen Satz, wenn es etwas zu sagen gibt.
//
// **Jede Probe trägt, was ihr Ruf zusagt** (`wirkung`, Z3): das Bild, das
// entstehen, das Feld, das sich ändern, die Liste, die so lang sein soll.
// `lastlaufwirkung.mjs` prüft es nach dem Profil an der Stichprobe — ein
// Status 2xx allein ist keine Wirkung.
//
// **Geschlossene Schleifen**: ein Klient schickt den nächsten Ruf, sobald der
// vorige zurück ist, und **wartet auf eine Drosselung nicht** — außer, wo das
// Profil ausdrücklich einen höflichen Klienten meint. So zeigt die Spalte
// „gedrosselt", wie viel ein ungeduldiger Aufrufer abgewiesen bekommt.

import { readFileSync } from 'node:fs'
import http from 'node:http'
import { performance } from 'node:perf_hooks'
import { setTimeout as schlafen } from 'node:timers/promises'

import { langsamerKlient, mcpUeberHTTP, ruf } from './lastlaufnetz.mjs'
import { hoeflich } from './lastlaufwirkung.mjs'

// dauerlauf fährt `klienten` Schleifen für `sekunden` und sammelt die Proben.
async function dauerlauf(klienten, sekunden, einRuf) {
  const ende = performance.now() + sekunden * 1000
  const proben = []

  await Promise.all(Array.from({ length: klienten }, async (_, klient) => {
    for (let lauf = 0; performance.now() < ende; lauf += 1) {
      proben.push(await einRuf(klient, lauf))
    }
  }))

  return proben
}

const karte = (k, klient, lauf) => k.karten[(klient + lauf) % k.karten.length]

/** zusage hängt an eine Probe, was ihr Ruf zusagt. */
const zusage = (probe, wirkung) => ({ ...probe, wirkung })

async function restLesen(k, klient, lauf) {
  if (lauf % 2 === 0) {
    return zusage(await ruf(k.rest, 'GET', `/sets/${k.setID}/karten`), { art: 'karten', anzahl: k.karten.length })
  }

  return zusage(await ruf(k.rest, 'GET', '/sets'), { art: 'sets' })
}

async function restSchreiben(k, klient, lauf) {
  const kennung = karte(k, klient, lauf)
  const wert = `Last ${klient}-${lauf}`
  const probe = await ruf(k.rest, 'PUT', `/sets/${k.setID}/karten/${kennung}/felder/name`, { rumpf: { wert } })

  return zusage(probe, { art: 'feld', setID: k.setID, karte: String(kennung), feld: 'name', wert })
}

async function restRendern(k, klient, lauf) {
  const ziel = `${k.arbeit}/bilder/${k.fall}-${klient}-${lauf}.png`

  return zusage(await ruf(k.rest, 'POST', `/karten/${karte(k, klient, lauf)}/bild`, { rumpf: { ziel } }), { art: 'bild', ziel })
}

async function mcpLesen(k) {
  return zusage(await mcpUeberHTTP(k.mcp, 'lesen.karten', { setID: Number(k.setID) }), { art: 'mcp-karten', anzahl: k.karten.length })
}

/** mcpStdioLesen liest die Karten über einen Prozess und trägt seine Zusage ein. */
async function mcpStdioLesen(k, prozess) {
  return zusage(await prozess.ruf('lesen.karten', { setID: Number(k.setID) }), { art: 'mcp-karten', anzahl: k.karten.length })
}

/** grenzenLesen liest die Grenzen — die Zusage ist eine Antwort, die JSON ist. */
async function grenzenLesen(k, optionen = {}) {
  return zusage(await ruf(k.rest, 'GET', '/grenzen', optionen), { art: 'json' })
}

/**
 * einfuhrProfil liest die MTG-JSON-Datei der Saat. **Die Zusage gilt dem Lauf**:
 * dieselbe Quelle ein zweites Mal gelesen legt nichts an (`schonOffen`), und
 * das ist richtig — dass aus der Datei ein Set mit ihren Karten entstanden ist,
 * zeigt der Bestand danach. `vorher` sind die Sets, die schon da waren.
 */
async function einfuhrProfil(k) {
  const einfuhr = await einfuhrVorbereiten(k)

  return dauerlauf(k.klienten, k.dauer, einfuhr)
}

// einfuhrVorbereiten räumt ein geerbtes Set weg und gibt den Ruf der Einfuhr
// samt seiner Zusage.
async function einfuhrVorbereiten(k) {
  const datei = JSON.parse(readFileSync(k.mtgjson, 'utf8')).data
  const lies = async () => JSON.parse((await hoeflich(() => ruf(k.rest, 'GET', '/sets'))).text).sets

  // **Eine Variante erbt kein Set der vorigen** (Lastlauf vom 2026-10-05):
  // beide lesen dieselbe Quelle in denselben Bestand, und die zweite erkannte
  // das Set nur noch — kein Ruf legte etwas an, und die Prüfung fand kein
  // neues. Ein gleichnamiges Set wird deshalb vorher entfernt; dann misst
  // jede Variante dasselbe: eine Einfuhr und die Erkennungen danach.
  for (const geerbt of (await lies()).filter((s) => s.name === datei.name)) {
    // **Mit dem heutigen Namen als Bestätigung** — ohne ihn weist `mse-rest`
    // das Löschen ab (`internal/rest/setloeschen.go`).
    const geloescht = await hoeflich(() => ruf(k.rest, 'DELETE', `/sets/${geerbt.id}?name=${encodeURIComponent(geerbt.name)}`))
    if (geloescht.status !== 200) throw new Error(`das geerbte Set ${geerbt.id} ließ sich nicht entfernen (${geloescht.status})`)
  }

  const sets = await lies()
  const wirkung = { art: 'einfuhr', vorher: sets.map((s) => String(s.id)), name: datei.name, karten: datei.cards.length }

  return async () => zusage(await ruf(k.rest, 'POST', '/mtgjson', { rumpf: { pfad: k.mtgjson, spiel: 'magic' } }), wirkung)
}

/**
 * rohpfadprobe fährt jeden Weg, der eine Wirkung zusagt, **einmal und ohne
 * Last** — QA-LAST01 (#2092): Lesen, ein Bild, eine Einfuhr mit neuem Set und
 * ein Auftrag bis zum fertigen Bogen. Der Lastlauf misst erst, wenn jeder
 * davon wirkt; sonst misst er, wie schnell der Dienst ablehnt.
 */
export async function rohpfadprobe(k) {
  const einfuhr = await einfuhrVorbereiten(k)

  return {
    gruppen: {
      Lesen: [await restLesen(k, 0, 0)],
      Rendern: [await restRendern(k, 0, 0)],
      Einfuhr: [await einfuhr()],
      Auftrag: [await auftrag(k, 0, 0)],
    },
  }
}

/**
 * einfuhrBefund sagt, was die Rufe der Einfuhr bewirkten: **nur der erste
 * legt ein Set an**, die übrigen erkennen die Quelle als schon offen — und
 * eine Abweisung oder Drosselung ist keines von beiden. Bis zum 2026-10-06
 * hieß jeder Ruf, der kein Set anlegte, eine Erkennung; im Lauf für 0.7.0
 * waren es lauter 400 (QA-LAST01, #2092).
 */
export function einfuhrBefund(proben) {
  const angelegte = (p) => (JSON.parse(p.text || '{}').angelegt ?? []).length
  const angelegt = proben.filter((p) => p.status === 201 && angelegte(p) > 0).length
  const erkannt = proben.filter((p) => p.status >= 200 && p.status < 300 && angelegte(p) === 0).length
  const gedrosselt = proben.filter((p) => p.status === 429).length
  const abgewiesen = proben.filter((p) => p.status >= 300 && p.status !== 429)
  const status = Object.entries(Object.groupBy(abgewiesen, (p) => String(p.status)))
    .map(([wert, liste]) => `${wert} × ${String(liste.length)}`).join(', ')

  return `${String(proben.length)} Rufe: ${String(angelegt)} legte${angelegt === 1 ? '' : 'n'} ein Set an, ` +
    `${String(erkannt)} erkannten die Quelle als schon offen, ${String(abgewiesen.length)} abgewiesen` +
    `${abgewiesen.length > 0 ? ` (${status})` : ''}, ${String(gedrosselt)} gedrosselt`
}

// auftrag stößt einen Bogen als Auftrag an und wartet auf sein Ende — die
// Probe ist der **ganze** Auftrag, vom Annehmen bis zum Stand danach.
async function auftrag(k, klient, lauf) {
  const beginn = performance.now()
  const ziel = `${k.arbeit}/boegen/${k.fall}-${klient}-${lauf}.pdf`
  const angenommen = await ruf(k.rest, 'POST', `/sets/${k.setID}/bogen`, {
    rumpf: { ziel }, koepfe: { Prefer: 'respond-async' },
  })

  // **Ein voller Auftragsplatz wird abgewartet** — `Retry-After` sagt fünf
  // Sekunden; wer sofort wieder fragt, misst nur die Absage.
  if (angenommen.status === 429) await schlafen(angenommen.warten * 1000)
  if (angenommen.status !== 202) return angenommen

  // **Höflich gefragt**: wer nach dem Stand fragt, zahlt wie jeder Leser —
  // eine Drosselung dabei wird abgewartet und nicht als Ende des Auftrags
  // gezählt.
  const ort = angenommen.koepfe.location
  for (;;) {
    const stand = await ruf(k.rest, 'GET', ort)
    if (stand.status === 429) {
      await schlafen(stand.warten * 1000)
      continue
    }

    if (stand.status !== 200) return { ...stand, ms: performance.now() - beginn }

    const gelesen = JSON.parse(stand.text).auftrag
    if (gelesen.stand !== 'laeuft') {
      // **Nur ein fertiger Auftrag sagt eine Datei zu** — ein abgebrochener
      // oder gescheiterter hat keinen Bogen versprochen.
      const probe = { status: gelesen.status, ms: performance.now() - beginn, inhalt: gelesen.stand }

      return gelesen.stand === 'fertig' ? zusage(probe, { art: 'bogen', ziel }) : probe
    }

    await schlafen(50)
  }
}

// abbruch stößt einen Auftrag an, bricht ihn gleich wieder ab und wartet, bis
// er nicht mehr läuft — **höflich**: eine Drosselung wird abgewartet.
async function abbruch(k) {
  const anstossen = () => ruf(k.rest, 'POST', `/sets/${k.setID}/bogen`, {
    rumpf: { ziel: `${k.arbeit}/boegen/${k.fall}-abbruch.pdf` }, koepfe: { Prefer: 'respond-async' },
  })
  let angenommen = await anstossen()
  for (let versuch = 0; angenommen.status === 429 && versuch < 3; versuch += 1) {
    await schlafen(angenommen.warten * 1000)
    angenommen = await anstossen()
  }

  if (angenommen.status !== 202) return `der Auftrag zum Abbrechen kam nicht an (${angenommen.status})`

  const ort = angenommen.koepfe.location
  const beginn = performance.now()
  const abgebrochen = await ruf(k.rest, 'DELETE', ort)
  let stand = 'laeuft'

  while (stand === 'laeuft' && performance.now() - beginn < 10_000) {
    const antwort = await ruf(k.rest, 'GET', ort)
    if (antwort.status === 200) stand = JSON.parse(antwort.text).auftrag?.stand
    await schlafen(antwort.status === 429 ? antwort.warten * 1000 : 20)
  }

  return `ein Auftrag, gleich nach dem Annehmen abgebrochen: \`DELETE\` → ${abgebrochen.status}, ` +
    `Stand \`${stand}\` nach ${Math.round(performance.now() - beginn)} ms`
}

export const profile = {
  'rest-lesen': async (k) => ({ gruppen: { REST: await dauerlauf(k.klienten, k.dauer, (c, l) => restLesen(k, c, l)) } }),

  'rest-schreiben': async (k) => ({ gruppen: { REST: await dauerlauf(k.klienten, k.dauer, (c, l) => restSchreiben(k, c, l)) } }),

  'rest-rendern': async (k) => ({ gruppen: { REST: await dauerlauf(k.klienten, k.dauer, (c, l) => restRendern(k, c, l)) } }),

  'rest-einfuhr': async (k) => {
    const proben = await einfuhrProfil(k)

    return { gruppen: { REST: proben }, befund: einfuhrBefund(proben) }
  },

  'rest-auftraege': async (k) => {
    const proben = await dauerlauf(k.klienten, k.dauer, (c, l) => auftrag(k, c, l))

    return { gruppen: { 'REST-Aufträge': proben }, befund: await abbruch(k) }
  },

  'mcp-stdio': async (k) => {
    const prozesse = await Promise.all(Array.from({ length: Math.min(k.klienten, 4) }, () => k.stdio()))
    const proben = await dauerlauf(prozesse.length, k.dauer, (c) => mcpStdioLesen(k, prozesse[c]))
    // **Die Verteilung höflich abgefragt** — ein leeres Guthaben wird
    // abgewartet, sonst fehlte die Zahl gerade dort, wo gedrosselt wurde.
    let metriken = await prozesse[0].ruf('lesen.metriken', {})
    for (let versuch = 0; metriken.gedrosselt && versuch < 3; versuch += 1) {
      await schlafen(metriken.warten * 1000)
      metriken = await prozesse[0].ruf('lesen.metriken', {})
    }
    await Promise.all(prozesse.map((p) => p.ende()))

    return {
      gruppen: { 'MCP-stdio': proben }, stdioMetriken: metriken,
      befund: `${prozesse.length} Prozesse, je einer ein eigener Principal und ein eigenes Guthaben`,
    }
  },

  'mcp-http': async (k) => ({ gruppen: { 'MCP-HTTP': await dauerlauf(k.klienten, k.dauer, () => mcpLesen(k)) } }),

  gemischt: async (k) => {
    const stdio = await k.stdio()
    const [lesen, schreiben, rendern, mcp, mcpStdio] = await Promise.all([
      dauerlauf(3, k.dauer, (c, l) => restLesen(k, c, l)),
      dauerlauf(2, k.dauer, (c, l) => restSchreiben(k, c, l)),
      dauerlauf(1, k.dauer, (c, l) => restRendern(k, c, l)),
      dauerlauf(2, k.dauer, () => mcpLesen(k)),
      dauerlauf(1, k.dauer, () => mcpStdioLesen(k, stdio)),
    ])
    await stdio.ende()

    return { gruppen: { 'REST lesen': lesen, 'REST schreiben': schreiben, 'REST rendern': rendern, 'MCP-HTTP': mcp, 'MCP-stdio': mcpStdio } }
  },

  // **Zwei Identitäten**: A flutet mit sechs Klienten, B fragt alle 250 ms
  // einmal. Fair ist, wenn B davon nichts merkt.
  fairness: async (k) => {
    const [flut, gemessen] = await Promise.all([
      dauerlauf(6, k.dauer, async () => zusage(await ruf(k.rest, 'GET', '/sets', { nachweis: k.token.a }), { art: 'sets' })),
      dauerlauf(1, k.dauer, async () => {
        const probe = zusage(await ruf(k.rest, 'GET', '/sets', { nachweis: k.token.b }), { art: 'sets' })
        await schlafen(250)

        return probe
      }),
    ])

    return { gruppen: { 'A flutet': flut, 'B fragt': gemessen } }
  },

  // **Überlast und Erholung**: erst ein höflicher Klient allein, dann
  // viermal so viele ungeduldige wie sonst, dann wieder der höfliche. Erholt
  // ist, was wieder gelingt und so schnell ist wie vorher.
  ueberlast: async (k) => {
    // **Ein höflicher Klient** fragt zehnmal je Sekunde und wartet eine
    // Drosselung ab — vorher als Maßstab, nachher als der, der sich erholt.
    let erholt = null
    let ende = 0
    const hoeflich = (sekunden, merken) => dauerlauf(1, sekunden, async (c, l) => {
      const probe = await restLesen(k, c, l)
      if (merken && erholt === null && probe.status === 200) erholt = performance.now() - ende
      await schlafen(probe.status === 429 ? probe.warten * 1000 : 100)

      return probe
    })

    const vorher = await hoeflich(2, false)
    const flut = await dauerlauf(k.klienten * 4, k.dauer, (c, l) => restLesen(k, c, l))
    ende = performance.now()
    const nachher = await hoeflich(3, true)

    return {
      gruppen: { 'vorher (höflich)': vorher, Überlast: flut, 'nachher (höflich)': nachher },
      dauern: { 'vorher (höflich)': 2, Überlast: k.dauer, 'nachher (höflich)': 3 },
      befund: `erster gelungener Ruf ${Math.round(erholt ?? Number.NaN)} ms nach dem Ende der Überlast`,
    }
  },

  // **Langsame Klienten**: zwanzig Verbindungen, die einen halben Kopf
  // schicken und dann schweigen — mehr als die sechzehn Plätze des Dienstes.
  // Daneben fragt je Dienst ein Klient alle 200 ms, zwölf Sekunden lang: so
  // lange, bis die Lesefrist von zehn Sekunden die Schweigenden hinauswirft.
  langsam: async (k) => {
    const schliessen = Array.from({ length: 20 }, () => [langsamerKlient(k.rest), langsamerKlient(k.mcp)]).flat()
    await schlafen(200)

    const [rest, mcp] = await Promise.all([
      alle200ms(12, () => grenzenLesen(k, { frist: 15_000 })),
      alle200ms(12, () => mcpUeberHTTP(k.mcp, 'lesen.grenzen', {})),
    ])
    schliessen.forEach((zu) => zu())

    return { gruppen: { 'REST hinter 20 halben Köpfen': rest, 'MCP-HTTP hinter 20 halben Köpfen': mcp }, sekunden: 12 }
  },

  // **Ruhende Verbindungen**: ein Klient mit sechzehn Verbindungen im Pool
  // fragt je einmal und lässt sie offen — wie jeder Pool mit `keepAlive`.
  // Die Lesefrist greift dann nicht, denn ein Kopf kam ja; was greift, ist
  // der Leerlauf von zwei Minuten. Gefragt wird zwölf Sekunden lang, mit
  // einer Frist von fünfzehn je Ruf.
  ruhend: async (k) => {
    const pool = new http.Agent({ keepAlive: true, maxSockets: 16 })
    const fremd = { ...k.rest, agent: pool }
    await Promise.all(Array.from({ length: 16 }, () => ruf(fremd, 'GET', '/grenzen')))

    const rest = await alle200ms(12, () => grenzenLesen(k, { frist: 15_000 }))
    pool.destroy()
    const danach = await alle200ms(2, () => grenzenLesen(k, { frist: 15_000 }))

    return {
      gruppen: { 'REST neben 16 ruhenden': rest, 'REST, nachdem sie gingen': danach },
      dauern: { 'REST neben 16 ruhenden': 12, 'REST, nachdem sie gingen': 2 },
    }
  },
}

// alle200ms fragt einmal und wartet dann 200 ms — `sekunden` lang.
function alle200ms(sekunden, einRuf) {
  return dauerlauf(1, sekunden, async () => {
    const probe = await einRuf()
    await schlafen(200)

    return probe
  })
}

// Die Reihenfolge der Profile im Beleg — und welche in einer Variante fehlen.
export const reihenfolge = Object.keys(profile)
