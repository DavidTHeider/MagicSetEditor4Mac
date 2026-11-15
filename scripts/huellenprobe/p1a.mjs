// **Die Abwahl beim Anfangen trägt bis in die Partie** — P1a, Befund C-01
// aus dem Audit zur Produktionsreife 0.9.0
// (#2335).
//
// Am nativen Fenster, mit einem eigenen künstlichen Bestand:
//
// 1. **Mit Abwahl**: Spielen → »Partie anfangen«, zwei Namen und zwei
//    Decks, »Wir wählen selbst«, die Gruppe »Ablauf einer Partie« auf,
//    »Zugstruktur selbst übernehmen« an, anfangen. Danach steht im Bestand
//    die Abwahl **und** die eingefrorene Matrix nennt die Zugstruktur
//    `manuell`; die Kommandozeile liest dieselbe Abwahl.
// 2. **Die Gegenprobe**: derselbe Weg ohne Abwahl — die Matrix nennt die
//    Zugstruktur `automatisch`. Ohne sie bewiese Teil 1 nur eine Matrix, die
//    immer `manuell` sagt (§10).
//
// **Gelesen wird der Bestand selbst** (`sqlite3` über `bestand.db`) und nicht
// eine Antwort, die die Wahl wiederholt: der Befund war genau eine Vorschau,
// die etwas anderes sagte als die gespeicherte Partie.
//
// **Was es nicht belegt**: die Bedienung mit dem Vorleseprogramm, und
// Kommandozeile, MCP und Dienst — die fährt `internal/flaechen/partieabwahl_test.go`.
import { execFileSync } from 'node:child_process'

const sitze = "section[aria-label=\"Wer sitzt am Tisch\"]"

// eingefroren liest die zuletzt angelegte Partie aus dem Bestand.
function eingefroren(ort) {
  const roh = execFileSync('sqlite3', [ort.bestand, 'SELECT zustand FROM partien ORDER BY rowid DESC LIMIT 1']).toString()
  const partie = JSON.parse(roh)
  const eintrag = (partie.Profil?.Matrix?.Eintraege ?? []).find((e) => e.Bereich === 'zugstruktur')

  return { id: partie.ID, abgewaehlt: partie.Regelwerk?.Abgewaehlt ?? [], grad: eintrag?.Grad ?? 'fehlt' }
}

// anfangen geht den Weg bis zum Tisch — mit oder ohne Abwahl.
async function anfangen(p, mitAbwahl) {
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  await p.wartetext('Wer sitzt am Tisch', 20000)

  for (const [nummer, name] of [[1, 'Ada'], [2, 'Bo']]) {
    const feld = `[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && (x.labels?.[0]?.textContent ?? '').trim() === 'Name am Tisch ${String(nummer)}')`
    await p.klickeAuf(feld, `das Feld »Name am Tisch ${String(nummer)}«`)
    await p.seite(`(${feld}).focus(); return true`)
    await p.tippeInsFeld(name)
  }
  for (const sitz of [0, 1]) {
    await p.seite(`const e = [...document.querySelectorAll('${sitze} select')][${String(sitz)}]; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(e, '1'); e.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  }

  if (mitAbwahl) {
    await p.wartetext('Wir wählen selbst', 20000)
    await p.klickeText('Wir wählen selbst')
    await p.wartetext('Ablauf einer Partie', 20000)
    await p.klickeText('Ablauf einer Partie')
    await p.wartetext('Zugstruktur selbst übernehmen', 20000)
    await p.klickeText('Zugstruktur selbst übernehmen')
    const angehakt = await p.warte('»Zugstruktur selbst übernehmen« ist angehakt', () => p.seite("return [...document.querySelectorAll('label')].find((l) => l.innerText.trim() === 'Zugstruktur selbst übernehmen')?.querySelector('input')?.checked === true"), 10000)
    p.pruefe('die Abwahl steht an der Tafel', Boolean(angehakt))
    await p.bild('p1a-tafel-mit-abwahl')
  }

  await p.klickeAuf("[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen').at(-1)", '»Partie anfangen« unter der Anlage')
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 30000)
}

export default async function p1a(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')

  for (const [bild, mitAbwahl] of [['hell', true], ['dunkel', false]]) {
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    p.notiz(mitAbwahl ? `Mit Abwahl der Zugstruktur (${bild})` : `Gegenprobe ohne Abwahl (${bild})`)
    await p.starte({ ort })
    await anfangen(p, mitAbwahl)
    await p.bild(mitAbwahl ? 'p1a-tisch-mit-abwahl' : 'p1a-tisch-ohne-abwahl')

    const stand = eingefroren(ort)
    p.zeige(`eingefroren: ${JSON.stringify(stand)}`)

    if (mitAbwahl) {
      p.pruefe('im Bestand steht die Abwahl »zugstruktur«', stand.abgewaehlt.length === 1 && stand.abgewaehlt[0] === 'zugstruktur', JSON.stringify(stand.abgewaehlt))
      p.pruefe('die eingefrorene Matrix nennt die Zugstruktur manuell', stand.grad === 'manuell', stand.grad)

      // **Und dieselbe Partie an einer anderen Fläche** — die Kommandozeile
      // liest, was das Fenster angelegt hat.
      const gelesen = JSON.parse(p.mse(ort, 'partie-zeigen', stand.id, 'sitz1', '--json'))
      p.pruefe('die Kommandozeile liest dieselbe Abwahl', JSON.stringify(gelesen.partie?.regelwerk?.abgewaehlt) === '["zugstruktur"]', JSON.stringify(gelesen.partie?.regelwerk))
    } else {
      p.pruefe('ohne Wahl steht keine Abwahl im Bestand', stand.abgewaehlt.length === 0, JSON.stringify(stand.abgewaehlt))
      p.pruefe('ohne Wahl nennt die Matrix die Zugstruktur automatisch', stand.grad === 'automatisch', stand.grad)
    }

    p.pruefe('⌘Q beendet', await p.beende())
  }
}
