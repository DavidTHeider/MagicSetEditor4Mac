// **Die Verwaltung einer Partie als Ganzem, am Haken** — PLAY-03.AK-05
// (#1711).
//
// # Wogegen hier geprüft wird
//
//   - **Jeder der vier Wege meldet, was daraus wurde** — und lädt danach
//     nach (§7, Regel 3).
//   - **Ein Fehler lädt nicht nach** — sonst überschriebe ein Lesefehler
//     die Meldung, die sagt, was zu tun ist.
//   - **Ein Abbruch ist kein Ergebnis** — wer den Dialog wegklickt, hat sich
//     entschieden (§6).
//   - **Ein zweiter Klick, während der erste läuft, tut nichts** (§7,
//     Regel 4).
import { render, screen, waitFor } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Einlesekennung, usePartieverwaltung } from './usePartieverwaltung'
import { Knopf } from '../bausteine/Knopf'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { PartieverwaltungAntwort } from '../apipartieverwaltung'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// eineAntwort gibt den Ausgang, den die Naht schickt.
function eineAntwort(
  abweichend: Partial<PartieverwaltungAntwort> = {},
): PartieverwaltungAntwort {
  return {
    kennung: 'p2',
    pfad: '',
    inhalt: '',
    abgebrochen: false,
    befunde: [],
    fehler: null,
    ...abweichend,
  }
}

// Probe zeichnet vier Knöpfe über dem Haken — einen je Weg.
function Probe({ nachladen }: { nachladen: () => void }) {
  const lauf = usePartieverwaltung(nachladen)

  return (
    <div>
      <Knopf
        aufKlick={() => {
          lauf.loeschen('p1')
        }}
      >
        löschen
      </Knopf>

      <Knopf
        aufKlick={() => {
          lauf.kopieren('p1')
        }}
      >
        kopieren
      </Knopf>

      <Knopf
        aufKlick={() => {
          lauf.ausfuehren('p1')
        }}
      >
        ausführen
      </Knopf>

      <Knopf aufKlick={lauf.einlesen}>einlesen</Knopf>

      {lauf.meldung !== undefined && (
        <p data-testid="meldung">
          {lauf.meldung.kennung} · {lauf.meldung.satz} ·{' '}
          {String(lauf.meldung.istFehler)}
        </p>
      )}

      <p data-testid="laeuft">{String(lauf.laeuft('p1'))}</p>
    </div>
  )
}

// zeige zeichnet die Probe und gibt die Hand.
function zeige(nachladen = vi.fn()) {
  const { unmount } = render(
    <Sprachraum sprache="de">
      <Probe nachladen={nachladen} />
    </Sprachraum>,
  )

  return { hand: nutzerereignis.setup(), nachladen, unmount }
}

// **Jeder Weg meldet seinen Satz und lädt danach nach.**
it('meldet jeden der vier Wege und lädt danach nach', async () => {
  const faelle = [
    { knopf: 'löschen', satz: eigene.partienGeloescht },
    { knopf: 'kopieren', satz: eigene.partienDupliziert },
    { knopf: 'ausführen', satz: 'Abgelegt nach /tmp/p.msepartie.' },
    { knopf: 'einlesen', satz: eigene.partienEingelesen },
  ] as const

  for (const fall of faelle) {
    belegeApi({
      PartieLoeschen: () => Promise.resolve(eineAntwort()),
      PartieKopieren: () => Promise.resolve(eineAntwort()),
      PartieAusfuhrSpeichern: () =>
        Promise.resolve(eineAntwort({ pfad: '/tmp/p.msepartie' })),
      PartieEinfuhrOeffnen: () => Promise.resolve(eineAntwort()),
    })

    const { hand, nachladen, unmount } = zeige()

    await hand.click(screen.getByRole('button', { name: fall.knopf }))

    await waitFor(() => {
      expect(screen.getByTestId('meldung')).toHaveTextContent(fall.satz)
    })

    expect(nachladen).toHaveBeenCalledTimes(1)

    // **Abgeräumt zwischen den Fällen**: zwei Proben nebeneinander hätten
    // jeden Knopf zweimal, und `getByRole` fände beide.
    unmount()
    raeumeApi()
  }
})

// **Die Meldung gehört zu der Kennung, die sie ausgelöst hat** — und beim
// Einlesen zu keiner Zeile.
it('hängt die Meldung an die Kennung ihres Weges', async () => {
  belegeApi({ PartieEinfuhrOeffnen: () => Promise.resolve(eineAntwort()) })

  const { hand } = zeige()

  await hand.click(screen.getByRole('button', { name: 'einlesen' }))

  await waitFor(() => {
    expect(screen.getByTestId('meldung')).toHaveTextContent(Einlesekennung)
  })
})

// **Ein Befund geht mit** — er macht die Partie nicht unbenutzbar, und er
// wird trotzdem genannt (§8).
it('nennt die Befunde der eingelesenen Datei', async () => {
  belegeApi({
    PartieEinfuhrOeffnen: () =>
      Promise.resolve(eineAntwort({ befunde: ['Das Feld »x« ist fremd.'] })),
  })

  const { hand } = zeige()

  await hand.click(screen.getByRole('button', { name: 'einlesen' }))

  await waitFor(() => {
    expect(screen.getByTestId('meldung')).toHaveTextContent('»x«')
  })
})

// **Ein Abbruch ist kein Ergebnis** — keine Meldung, kein Nachladen.
it('meldet nichts, wenn der Dialog weggeklickt wird', async () => {
  belegeApi({
    PartieAusfuhrSpeichern: () =>
      Promise.resolve(eineAntwort({ abgebrochen: true })),
  })

  const { hand, nachladen } = zeige()

  await hand.click(screen.getByRole('button', { name: 'ausführen' }))
  await antwortAbwarten()

  expect(screen.queryByTestId('meldung')).not.toBeInTheDocument()
  expect(nachladen).not.toHaveBeenCalled()
})

// **Ein Fehler wird gemeldet und nicht nachgeladen** — §7, Regel 3.
it('meldet den Fehler und lädt nicht nach', async () => {
  belegeApi({
    PartieLoeschen: () =>
      Promise.resolve(
        eineAntwort({
          fehler: {
            code: 'konflikt',
            meldung: 'Die Partie ist beendet.',
            schluessel: '',
          },
        }),
      ),
  })

  const { hand, nachladen } = zeige()

  await hand.click(screen.getByRole('button', { name: 'löschen' }))

  await waitFor(() => {
    expect(screen.getByTestId('meldung')).toHaveTextContent('beendet')
  })

  expect(screen.getByTestId('meldung')).toHaveTextContent('true')
  expect(nachladen).not.toHaveBeenCalled()
})

// **Ein zweiter Klick, während der erste läuft, tut nichts** — §7, Regel 4.
it('lässt einen zweiten Ruf nicht an denselben Datensatz', async () => {
  let loesen: (antwort: PartieverwaltungAntwort) => void = () => undefined

  const gerufen = vi.fn(
    () =>
      new Promise<PartieverwaltungAntwort>((fertig) => {
        loesen = fertig
      }),
  )

  belegeApi({ PartieLoeschen: gerufen })

  const { hand } = zeige()

  await hand.click(screen.getByRole('button', { name: 'löschen' }))

  await waitFor(() => {
    expect(screen.getByTestId('laeuft')).toHaveTextContent('true')
  })

  await hand.click(screen.getByRole('button', { name: 'löschen' }))

  expect(gerufen).toHaveBeenCalledTimes(1)

  loesen(eineAntwort())

  await waitFor(() => {
    expect(screen.getByTestId('laeuft')).toHaveTextContent('false')
  })
})
