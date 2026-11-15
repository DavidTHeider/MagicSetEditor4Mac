// **Die kleinen Bilder der Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenbildAntwort } from '../api'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { useListenbilder } from './useListenbilder'

afterEach(() => {
  raeumeApi()
})

// Probe zeigt, was der Haken zurückgibt.
function Probe({
  setID,
  ids,
  stand = 0,
}: {
  setID: string
  ids: string[]

  // stand ist der Bestandsstand — die Bilder hängen daran (#1215, Befund 5).
  stand?: number
}) {
  const bilder = useListenbilder(setID, ids, stand)

  return <p data-testid="bilder">{JSON.stringify(bilder)}</p>
}

describe('die Listenbilder', () => {
  // **Geholt wird, was zu sehen ist — in einem Ruf.**
  it('fragt die sichtbaren Kennungen auf einmal', async () => {
    const geholt = vi.fn().mockResolvedValue({
      bilder: { '1': 'data:image/png;base64,AAA' },
      fehler: undefined,
    })
    belegeApi({ Kartenbilder: geholt })

    render(<Probe setID="7" ids={['1', '2']} />)

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    // `toHaveBeenCalled` trifft zu, sobald gerufen wurde — die Antwort ist
    // dann noch unterwegs, und was sie bewirkt, steht noch nicht da.
    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('data:image/png')
    })

    expect(geholt).toHaveBeenCalledWith('7', ['1', '2'])
  })

  // **Was schon gefragt ist, wird nicht wieder gefragt.**
  //
  // Auch das, was kein Bild hatte: sonst fragte jedes Rollen erneut nach jeder
  // Karte ohne Illustration — und das sind die meisten.
  it('fragt eine Kennung kein zweites Mal', async () => {
    const geholt = vi.fn().mockResolvedValue({ bilder: {}, fehler: undefined })
    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1', '2']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenCalledTimes(1)
    })

    // **Ein überlappendes Fenster fragt nur das Neue.** Ohne diese Zusage
    // hätte der Zähler oben auch bei einem Haken gehalten, der gar nichts
    // merkt — das Fenster wäre ja dasselbe.
    rerender(<Probe setID="7" ids={['2', '3']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenCalledTimes(2)
    })

    expect(geholt).toHaveBeenLastCalledWith('7', ['3'])
  })

  // **Ein Setwechsel räumt die Tabelle** (§7, Regel 1).
  //
  // Kennungen zählen je Bestand und nicht je Set; ein gemerktes Bild unter der
  // Kennung 1 träfe im nächsten Set eine **andere** Karte.
  it('vergisst die Bilder beim Setwechsel', async () => {
    const geholt = vi.fn().mockResolvedValue({
      bilder: { '1': 'data:image/png;base64,AAA' },
      fehler: undefined,
    })
    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('data:image/png')
    })

    rerender(<Probe setID="8" ids={['1']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenLastCalledWith('8', ['1'])
    })
  })

  // **Ein Fehler färbt die Liste nicht rot** (§8).
  //
  // Die Zeile steht mit ihrem Namen da; ein Hinweis über einer vollständigen
  // Liste sagte etwas Falscheres als gar keiner.
  it('bleibt bei einem Fehler still', async () => {
    const geholt = vi.fn().mockRejectedValue(new Error('das Bild fehlt'))
    belegeApi({ Kartenbilder: geholt })

    render(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenCalled()
    })

    // **Erst muss die Ablehnung angekommen sein.** „Nichts eingetragen" ist
    // vorher keine Aussage, sondern eine Selbstverständlichkeit (§3, #1321).
    await antwortAbwarten()

    expect(screen.getByTestId('bilder')).toHaveTextContent('{}')
  })

  // **Ein leeres Fenster fragt gar nicht.**
  it('fragt nichts, wenn nichts zu sehen ist', async () => {
    const geholt = vi.fn().mockResolvedValue({ bilder: {}, fehler: undefined })
    belegeApi({ Kartenbilder: geholt })

    render(<Probe setID="7" ids={[]} />)

    // **Und der Beleg, dass jemand zuhört**: derselbe Haken mit einer Kennung
    // ruft. Ohne ihn wäre die Verneinung auch bei einem Haken grün, der nie
    // etwas tut (§3).
    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toBeInTheDocument()
    })

    expect(geholt).not.toHaveBeenCalled()

    render(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenCalledOnce()
    })
  })

  // **Ein Fehlschlag macht die Bilder nicht dauerhaft unsichtbar.**
  //
  // # Der Befund
  //
  // Der erste Anlauf trug die Kennungen **vor** dem Ruf in `geholt` ein und
  // ließ sie auch dann stehen, wenn der Stapel scheiterte. Derselbe
  // Bildbereich wurde danach nie wieder angefordert — „ein Fehler färbt die
  // Liste nicht rot" wurde zu „ein vorübergehender Fehler macht die Bilder bis
  // zum Setwechsel dauerhaft unsichtbar". Angestrichen vom Zweitblick über
  // Phase 3.5 als Befund 7: die Verwechslung von **versucht** und
  // **beantwortet**.
  it('fragt nach einem Fehlschlag wieder', async () => {
    const geholt = vi
      .fn()
      .mockRejectedValueOnce(new Error('das Netz hakt'))
      .mockResolvedValue({
        bilder: { '1': 'data:image/png;base64,AAA' },
        fehler: null,
      })

    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(geholt).toHaveBeenCalledTimes(1)
    })

    // **Dasselbe Fenster noch einmal** — wie beim Zurückrollen.
    rerender(<Probe setID="7" ids={[]} />)
    rerender(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('data:image/png')
    })

    expect(geholt).toHaveBeenCalledTimes(2)
  })

  // **Zwei Stapel desselben Sets ergänzen einander.**
  //
  // # Der Befund
  //
  // Der erste Anlauf zählte einen Lauf je Sichtfenster mit: wer weiter rollte,
  // während der vorige Stapel lief, erklärte dessen Antwort für überholt —
  // obwohl beide zum selben Set gehören. Gemerkt waren die Kennungen trotzdem,
  // also blieben die Bilder beim Zurückrollen leer. Angestrichen als Befund 8.
  //
  // **Hier lösen sie sich in Gegenreihenfolge auf**, denn genau das war der
  // Fall, den kein Lauf hielt.
  it('verwirft den ersten Stapel nicht, wenn der zweite früher antwortet', async () => {
    const antworten = new Map<string, (wert: KartenbildAntwort) => void>()

    const geholt = vi.fn((_setID: string, ids: string[]) => {
      return new Promise<KartenbildAntwort>((aufloesen) => {
        antworten.set(ids.join(','), aufloesen)
      })
    })

    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1']} />)

    await waitFor(() => {
      expect(antworten.has('1')).toBe(true)
    })

    rerender(<Probe setID="7" ids={['2']} />)

    await waitFor(() => {
      expect(antworten.has('2')).toBe(true)
    })

    // **Der zweite zuerst.**
    antworten.get('2')?.({ bilder: { '2': 'zwei' }, fehler: null })
    antworten.get('1')?.({ bilder: { '1': 'eins' }, fehler: null })

    await waitFor(() => {
      const gezeigt = screen.getByTestId('bilder').textContent

      expect(gezeigt).toContain('eins')
      expect(gezeigt).toContain('zwei')
    })
  })

  // **Eine späte Antwort des alten Sets fasst nichts mehr an.**
  //
  // # Der Befund
  //
  // Beim Setwechsel ersetzt der Haken `geholt` und `laufend` durch neue
  // Mengen. Die schon laufenden Blöcke hielten aber keine setgebundene
  // Generation und schrieben nach ihrer Antwort in genau **diese neuen**
  // Referenzen. Angestrichen vom Nachlauf über Phase 3.5 als Befund 2.
  //
  // # Warum die Anzeige das nicht zeigt — und was deshalb geprüft wird
  //
  // Der Lesewächter (`gezeigt.fuer === setID`) verdeckt den Schaden: die
  // Bilder des alten Sets erscheinen ohnehin nicht. **Der Schaden steht in
  // der Buchführung**: die alte Antwort markiert Kennungen im **neuen** Set
  // als geholt, und die werden danach nie wieder gefragt — die Karte bleibt
  // für immer ohne Bild.
  //
  // **Deshalb fragt das alte Set nach zwei Kennungen und das neue nach
  // einer.** Die zweite ist die, die der Fehler verschluckt.
  it('lässt eine späte Antwort des alten Sets die Buchführung nicht anfassen', async () => {
    const antworten = new Map<string, (wert: KartenbildAntwort) => void>()

    const geholt = vi.fn((setID: string, ids: string[]) => {
      return new Promise<KartenbildAntwort>((aufloesen) => {
        antworten.set(`${setID}:${ids.join(',')}`, aufloesen)
      })
    })

    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1', '2']} />)

    await waitFor(() => {
      expect(antworten.has('7:1,2')).toBe(true)
    })

    // **Das Set wechselt, während die erste Antwort unterwegs ist.**
    rerender(<Probe setID="8" ids={['1']} />)

    await waitFor(() => {
      expect(antworten.has('8:1')).toBe(true)
    })

    // **Und jetzt antwortet das alte Set — für beide Kennungen.**
    antworten.get('7:1,2')?.({
      bilder: { '1': 'alt-eins', '2': 'alt-zwei' },
      fehler: null,
    })

    antworten.get('8:1')?.({ bilder: { '1': 'neu-eins' }, fehler: null })

    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('neu-eins')
    })

    // **Jetzt rollt jemand zu Karte 2 des neuen Sets.**
    //
    // Sie ist nie gefragt worden — die alte Antwort hat sie nur **markiert**.
    rerender(<Probe setID="8" ids={['2']} />)

    await waitFor(() => {
      expect(antworten.has('8:2')).toBe(true)
    })
  })

  // **Ein neuer Bestandsstand macht die gemerkten Bilder ungültig.**
  //
  // # Der Befund
  //
  // Wer ein Kartenbild **einsetzt oder ersetzt**, ändert die Datei, auf die
  // ein gemerktes Listenbild zeigt. `geholt` merkte sich jede Antwort und
  // wurde nur beim **Setwechsel** geräumt — der Kasten zeigte danach weiter
  // das alte Bild oder blieb leer. Angestrichen als Befund 5.
  it('fragt nach einer Bestandsänderung wieder', async () => {
    const geholt = vi
      .fn()
      .mockResolvedValueOnce({ bilder: { '1': 'vorher' }, fehler: null })
      .mockResolvedValue({ bilder: { '1': 'nachher' }, fehler: null })

    belegeApi({ Kartenbilder: geholt })

    const { rerender } = render(<Probe setID="7" ids={['1']} stand={0} />)

    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('vorher')
    })

    rerender(<Probe setID="7" ids={['1']} stand={1} />)

    await waitFor(() => {
      expect(screen.getByTestId('bilder')).toHaveTextContent('nachher')
    })

    expect(geholt).toHaveBeenCalledTimes(2)
  })
})
