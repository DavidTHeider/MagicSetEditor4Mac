// **Wen dieses Profil kennt** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Begegnung nichts erlaubt.** Ein unbestätigtes Gegenüber hat
// keinen Knopf, mit dem man ihm etwas freigäbe — erst der Vergleich des
// Fingerabdrucks macht es prüfbar, und erst danach gibt es die Zwecke.
//
// **Und dass der Vergleich ein Feld ist und kein Klick.** Was ein Mensch
// verglichen hat, geht an die Anwendung; sie entscheidet, ob es stimmt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Peerkontakte } from './Peerkontakte'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { dieZwecke, einKontakt, eineListe, zeige } from './peerkontaktprobe'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

// **Zugeklappt wird nicht gefragt.**
//
// **Die Gegenprobe zum Aufklappen** (§10): ein Kasten, der beim Zeichnen
// fragte, ginge bei jedem Seitenwechsel an den Bestand.
it('fragt nichts, solange niemand aufschlägt', () => {
  const gerufen = vi.fn()

  belegeApi({ Peerkontakte: gerufen })

  render(
    <Sprachraum sprache="de">
      <Peerkontakte />
    </Sprachraum>,
  )

  expect(gerufen).not.toHaveBeenCalled()
  expect(
    screen.getByRole('button', { name: eigene.kontakteZeigen }),
  ).toBeInTheDocument()
})

// **Ein unbestätigtes Gegenüber hat keine Zwecke.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **„Pairing allein gestattet keinen Import oder Handel."** Am Fenster heißt
// das: die Knöpfe stehen gar nicht erst da. Ein ausgegrauter Knopf wäre eine
// Einladung, die die Anwendung dann ablehnt.
it('zeigt einem unbestätigten Gegenüber keine Zwecke', async () => {
  belegeApi({ Peerkontakte: () => Promise.resolve(eineListe()) })

  await zeige()

  expect(
    await screen.findByText(eigene.kontakteStandUnbestaetigt),
  ).toBeInTheDocument()

  expect(screen.queryByRole('switch')).not.toBeInTheDocument()

  // **Und der Name steht mit seinem Vorbehalt da.**
  expect(screen.getByText(/Fremder Rechner/)).toBeInTheDocument()
  expect(
    screen.getByText(new RegExp(eigene.kontakteNenntHinweis)),
  ).toBeInTheDocument()
})

// **Der verglichene Abdruck geht an die Anwendung.**
it('reicht den verglichenen Abdruck durch', async () => {
  const gerufen = vi.fn(() =>
    Promise.resolve(eineListe([einKontakt({ stand: 'bestaetigt' })])),
  )

  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe()),
    PeerkontaktBestaetigen: gerufen,
  })

  const hand = await zeige()

  await hand.type(
    await screen.findByLabelText(eigene.kontakteVergleichen),
    'abcd-ef01',
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.kontakteBestaetigen }),
  )

  expect(gerufen).toHaveBeenCalledWith('bgegenueber', 'abcd-ef01')
})

// **Ein bestätigtes Gegenüber hat Zwecke — und die Paarung ist keiner.**
//
// **Sie ist das Kennenlernen selbst**: ein Knopf dafür wäre eine Freigabe für
// etwas, das ohnehin jeder darf, der nicht gesperrt ist.
it('zeigt einem bestätigten Gegenüber die Zwecke ohne die Paarung', async () => {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['teilen'] })]),
      ),
  })

  await zeige()

  await screen.findByText(eigene.kontakteStandBestaetigt)

  expect(screen.queryByText(eigene.zweckPaarung)).not.toBeInTheDocument()
  expect(screen.getByText(eigene.zweckHandel)).toBeInTheDocument()

  // **Der gegebene Zweck steht eingeschaltet da**, die übrigen aus — ein
  // Schalter je Zweck seit UXR-16 (#1924, R7), wo vorher ein Knopf das Wort
  // trug, das ihn umkehrt.
  expect(
    screen
      .getAllByRole('switch')
      .map((schalter) => [
        schalter.closest('label')?.textContent,
        (schalter as HTMLInputElement).checked,
      ]),
  ).toEqual([
    [eigene.zweckTeilen, true],
    [eigene.zweckHandel, false],
    [eigene.zweckPartie, false],
  ])
})

// **Ein Nachladefehler ist kein gescheiterter Schritt** (§7, Regel 3;
// #1960).
//
// Die Sperre ist geschrieben, nur die Liste danach war in Go nicht zu lesen.
// Die Seite nimmt den Schritt als geschehen — sie meldet ihn weiter und lädt
// selbst nach —, und was dann nicht zu lesen ist, steht **an der Liste**.
it('nimmt eine Änderung mit Nachladefehler als geschehen', async () => {
  const klemmt = {
    code: 'unbekannt' as const,
    schluessel: '',
    meldung: 'Die Liste klemmt',
    feld: '',
  }
  const lesen = vi
    .fn()
    .mockResolvedValueOnce(eineListe())
    .mockResolvedValue(eineListe([], klemmt))
  const sperren = vi.fn(() =>
    Promise.resolve({
      kontakte: [],
      zwecke: dieZwecke,
      nachladefehler: klemmt,
      fehler: null,
    }),
  )
  const geaendert = vi.fn()

  belegeApi({ Peerkontakte: lesen, PeerkontaktSperren: sperren })

  render(
    <Sprachraum sprache="de">
      <Peerkontakte eingebettet aufGeaendert={geaendert} />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await hand.click(
    await screen.findByRole('button', { name: eigene.kontakteSperren }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent('Die Liste klemmt')
  expect(screen.getAllByRole('alert')).toHaveLength(1)
  expect(sperren).toHaveBeenCalledWith('bgegenueber')
  expect(geaendert).toHaveBeenCalledTimes(1)
  expect(lesen).toHaveBeenCalledTimes(2)
})
