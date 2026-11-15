// **Was in einer Modellwahl steckt** —
// #1539.
//
// # Was hier geprüft wird
//
// **Die Auftrennung und die Kostenfrage** — beides an drei Stellen benutzt
// und deshalb an einer geschrieben.
import { expect, it } from 'vitest'

import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'

import type { KIModellDTO } from '../antworten'

// einModell baut einen Katalogeintrag.
function einModell(anbieter: string, alias: string, ort: string): KIModellDTO {
  return {
    anbieter,
    alias,
    laufzeit: '',
    ort,
    faehigkeiten: [],
    herkunft: [],
    bindung: `${anbieter}/${alias}`,
    kontextfenster: 0,
    kontextbeleg: '',
  }
}

// **Der erste Teil ist der Anbieter, der Rest der Alias.**
//
// **Und der Rest darf Schrägstriche tragen.** `gemma4:8b` nicht, aber ein
// Modellname mit Namensraum schon — und wer am **letzten** Schrägstrich
// trennte, machte aus `ollama/hf.co/modell` den Anbieter `ollama/hf.co`.
it('trennt am ersten Schrägstrich und nicht am letzten', () => {
  expect(anbieterAus('ollama/hf.co/modell')).toBe('ollama')
  expect(aliasAus('ollama/hf.co/modell')).toBe('hf.co/modell')
})

// **Eine Wahl ohne Schrägstrich hat keinen Alias.**
//
// # Warum das dasteht, obwohl der Knopf sie nicht zulässt
//
// **Weil das alte `slice` hier etwas Falsches sagte, statt zu schweigen**:
// `'abc'.slice(0, -1)` ist `'ab'` — ein Anbieter, den es nicht gibt, aus
// einer Eingabe, die keiner ist. Die Antwort ist der ganze Name und ein
// leerer Alias, und dann findet [kostetGeld] nichts und der Ruf scheitert
// mit einem Namen, den jemand wiedererkennt.
it('gibt ohne Schrägstrich den ganzen Namen und keinen Alias', () => {
  expect(anbieterAus('abc')).toBe('abc')
  expect(aliasAus('abc')).toBe('')
})

// **Was anderswo rechnet, kostet Geld — was hier rechnet, nicht.**
it('erkennt am Ort, ob ein Modell Geld kostet', () => {
  const katalog = [
    einModell('ollama', 'gemma4:8b', 'lokal'),
    einModell('anthropic', 'claude-opus-5', 'cloud'),
  ]

  expect(kostetGeld(katalog, 'anthropic/claude-opus-5')).toBe(true)
  expect(kostetGeld(katalog, 'ollama/gemma4:8b')).toBe(false)
})

// **Und eine Wahl, die im Katalog nicht steht, kostet nichts.**
//
// **Weil es dann nichts zu fragen gibt** — der Ruf scheitert ohnehin, und
// eine Freigabe davor wäre eine Frage nach Geld für einen Auftrag, den
// niemand annimmt.
it('hält eine unbekannte Wahl für kostenlos', () => {
  expect(
    kostetGeld([einModell('ollama', 'gemma4:8b', 'lokal')], 'gibt/esnicht'),
  ).toBe(false)
})
