// **Wem eine Antwort gehört** — die Befunde 4 (Zweitblick) und 2 (Nachlauf)
// über Phase 5.4.
//
// **Herausgelöst aus `umschlag.test.tsx`**, als die Datei über die
// Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort, **was** der
// Haken tut; hier, **zu welchem Stand** seine Antworten gehören (§7, Regel 2).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { useUmschlag } from './useUmschlag'

afterEach(() => {
  raeumeApi()
})

function rahmen({ children }: { children: React.ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function zeige() {
  return renderHook(() => useUmschlag(), { wrapper: rahmen })
}

const leer = {
  objekte: [],
  urheber: '',
  fingerabdruck: '',
  pfad: '',
  abgebrochen: false,
  fehler: null,
}

const einUmschlag = {
  ...leer,
  pfad: '/tmp/u.msepackage',
  urheber: 'k51q',
  fingerabdruck: 'AB CD',
  objekte: [
    { art: 'set', bezeichnung: 'set.mse-set', groesse: 12, braucht: [] },
  ],
}

describe('eine Antwort gehört zu dem Stand, aus dem sie stammt', () => {
  // **Befund 4 des Zweitblicks über Phase 5.4.** Wählen und Übernehmen
  // hatten zwei Vorgangsschlüssel und hielten einander nicht auf: wer
  // während einer laufenden Übernahme einen zweiten Umschlag wählte, sah ihn
  // kurz — und dann nahm die verspätete Antwort der Übernahme ihn wieder
  // weg (§7, Regel 2).
  it('die späte Übernahme löscht die neue Vorschau nicht', async () => {
    let gibFrei: (antwort: unknown) => void = () => undefined

    const uebernahme = new Promise((erfuellen) => {
      gibFrei = erfuellen
    })

    const zweiter = { ...einUmschlag, pfad: '/tmp/zwei.msepackage' }

    belegeApi({
      UmschlagUebernehmen: vi.fn().mockReturnValue(uebernahme),
      UmschlagWaehlen: vi.fn().mockResolvedValue(zweiter),
    })

    const { result } = zeige()

    act(() => {
      result.current.uebernehmen('/tmp/eins.msepackage')
    })

    // **Und während sie läuft, wird ein zweiter gewählt** — am Knopf geht
    // das seit dieser Runde nicht mehr; hier wird der Weg darunter
    // gemessen, denn er ist die Zusage, und der Knopf ist die Bequemlichkeit.
    await act(async () => {
      result.current.waehlen()

      await Promise.resolve()
    })

    await waitFor(() => {
      expect(result.current.angesehen?.pfad).toBe('/tmp/zwei.msepackage')
    })

    await act(async () => {
      gibFrei({ ...leer, setID: '9' })

      await Promise.resolve()
    })

    // **Die Gegenprobe steckt in der Erwartung**: vor der Korrektur stand
    // hier `undefined`, weil `setzeAngesehen(undefined)` aus dem überholten
    // Lauf kam.
    expect(result.current.angesehen?.pfad).toBe('/tmp/zwei.msepackage')

    // **Und ihr Ergebnis trägt sie auch nicht ein** — sie gehört zu einem
    // Stand, den niemand mehr ansieht.
    expect(result.current.uebernommen).toBeUndefined()
  })

  it('ein später Fehler der Übernahme meldet sich nicht über dem neuen', async () => {
    let scheitere: (grund: unknown) => void = () => undefined

    const uebernahme = new Promise((_, verwerfen) => {
      scheitere = verwerfen
    })

    belegeApi({
      UmschlagUebernehmen: vi.fn().mockReturnValue(uebernahme),
      UmschlagWaehlen: vi.fn().mockResolvedValue(einUmschlag),
    })

    const { result } = zeige()

    act(() => {
      result.current.uebernehmen('/tmp/eins.msepackage')
    })

    await act(async () => {
      result.current.waehlen()

      await Promise.resolve()
    })

    await waitFor(() => {
      expect(result.current.angesehen).toBeDefined()
    })

    await act(async () => {
      scheitere(new Error('kaputt'))

      await Promise.resolve()
    })

    // **Die unauffälligere Hälfte** (siehe den Kopf von `useLauf`): eine
    // Meldung zu einem Umschlag, den niemand mehr ansieht, sieht aus wie
    // eine über den, der dasteht.
    expect(result.current.fehler).toBeUndefined()
  })
})

describe('zwei Sets gehen nebeneinander hinaus', () => {
  // **Befund 2 des Nachlaufs zum Zweitblick über Phase 5.4.** `schnueren`
  // hat einen Vorgangsschlüssel **je Set** — zwei dürfen nebeneinander
  // laufen —, schreibt aber in dieselbe sichtbare Fläche. Die späte Antwort
  // von A stand über B (§7, Regel 2).
  it('die späte Antwort des ersten steht nicht über dem zweiten', async () => {
    let gibFrei: (antwort: unknown) => void = () => undefined

    const erster = new Promise((erfuellen) => {
      gibFrei = erfuellen
    })

    const zweiter = { ...leer, objekte: einUmschlag.objekte, pfad: '/tmp/b' }

    belegeApi({
      UmschlagSchnueren: vi
        .fn()
        .mockReturnValueOnce(erster)
        .mockResolvedValueOnce(zweiter),
    })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('A')
    })

    await act(async () => {
      result.current.schnueren('B')

      await Promise.resolve()
    })

    await waitFor(() => {
      expect(result.current.fuer).toBe('B')
    })

    await act(async () => {
      gibFrei({ ...leer, objekte: einUmschlag.objekte, pfad: '/tmp/a' })

      await Promise.resolve()
    })

    // **Die Gegenprobe steckt in der Erwartung**: ohne den Zähler stünde
    // hier `/tmp/a` — die Bestätigung von A über dem Set B.
    expect(result.current.geschnuert?.pfad).toBe('/tmp/b')
    expect(result.current.fuer).toBe('B')
  })

  it('ein später Fehler des ersten meldet sich nicht über dem zweiten', async () => {
    let scheitere: (grund: unknown) => void = () => undefined

    const erster = new Promise((_, verwerfen) => {
      scheitere = verwerfen
    })

    belegeApi({
      UmschlagSchnueren: vi
        .fn()
        .mockReturnValueOnce(erster)
        .mockResolvedValueOnce({ ...leer, objekte: einUmschlag.objekte }),
    })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('A')
    })

    await act(async () => {
      result.current.schnueren('B')

      await Promise.resolve()
    })

    await waitFor(() => {
      expect(result.current.fuer).toBe('B')
    })

    await act(async () => {
      scheitere(new Error('kaputt'))

      await Promise.resolve()
    })

    expect(result.current.fehler).toBeUndefined()
  })

  it('wer aufräumt, bekommt nichts zurück', async () => {
    let gibFrei: (antwort: unknown) => void = () => undefined

    belegeApi({
      UmschlagSchnueren: vi.fn().mockReturnValue(
        new Promise((erfuellen) => {
          gibFrei = erfuellen
        }),
      ),
    })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('A')
    })

    act(() => {
      result.current.raeumen()
    })

    await act(async () => {
      gibFrei({ ...leer, objekte: einUmschlag.objekte, pfad: '/tmp/a' })

      await Promise.resolve()
    })

    // **Gewartet wird auf das Ende des Vorgangs und nicht auf eine
    // Mikroaufgabe.** Der erste Anlauf sicherte direkt nach `gibFrei` zu —
    // und war grün, **bevor** die Antwort überhaupt eingetragen war. Eine
    // Zusicherung, die schon vor der Antwort zutrifft, misst nichts (§3);
    // die Gegenprobe hat es gezeigt: sie blieb grün, als der Schutz fiel.
    await waitFor(() => {
      expect(result.current.laeuft('schnueren-A')).toBe(false)
    })

    expect(result.current.geschnuert).toBeUndefined()
    expect(result.current.fuer).toBeUndefined()
  })
})
