// **Wo ein Gegenüber erreichbar ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum ein eigener Baustein
//
// **Weil `Peerkontaktzeile.tsx` mit ihm über die Zeilengrenze aus §2 ginge** —
// und weil er eine andere Frage beantwortet: dort steht, **wem** vertraut wird,
// hier, **wohin** gewählt wird.
//
// # Warum das Feld auch an einem unbestätigten Gegenüber steht
//
// **Weil eine Adresse nichts erlaubt.** Sie sagt, wohin gewählt wird, und
// nicht, was danach gilt. Wer sie an die Bestätigung bände, könnte ein frisch
// eingetragenes Gegenüber nicht einmal anrufen, um seinen Abdruck zu
// vergleichen — und genau dafür gibt es den Zweck `paarung`.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { kennungseingabe } from './kennungseingabe'

interface Kontaktadresseneigenschaften {
  kennung: string
  adresse: string
  laeuft: boolean
  setzen: (kennung: string, adresse: string) => void
}

export function Kontaktadresse({
  kennung,
  adresse,
  laeuft,
  setzen,
}: Kontaktadresseneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const feld = useId()

  // **Der Entwurf beginnt beim eingetragenen Wert und bleibt dann hier.**
  //
  // **Geräumt wird er über den `key` des Aufrufers** (§7, Regel 1) und nicht
  // über einen Effekt: nach einem geglückten Schreiben kommt die ganze Liste
  // zurück, und ein neuer `key` baut dieses Feld mit dem neuen Wert auf.
  // Ein `useEffect`, der `setState` ruft, erzeugt eine zweite Zeichnung für
  // dasselbe Ergebnis — React sagt es selbst, und der Linter hält es fest.
  const [entwurf, setzeEntwurf] = useState(adresse)

  return (
    <div className="flex flex-wrap items-end gap-2">
      <label className="flex grow flex-col gap-1 text-sm" htmlFor={feld}>
        {eigene.kontakteAdresse}
        <input
          {...kennungseingabe}
          id={feld}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          placeholder={eigene.kontakteAdresseBeispiel}
          value={entwurf}
          onChange={(ereignis) => {
            setzeEntwurf(ereignis.target.value)
          }}
        />
      </label>

      {/* **Zwei Wege, ein Feld** (N3, #1456): eine Adresse im eigenen Netz
          oder eine Einladung über ein fremdes. Ohne diesen Satz müsste
          jemand raten, ob die Einladung hierher gehört. */}
      <Erklaerungsabsatz
        text={eigene.kontakteAdresseHinweis}
        klasse="w-full text-sm text-leise"
      />

      {/* **Kein Knopf zum Löschen daneben.** Ein leeres Feld und „Merken"
          löscht sie — dieselbe Angabe an derselben Stelle, und niemand muss
          raten, welcher der beiden Knöpfe der richtige ist. */}
      <Knopf
        gewicht="leise"
        angehalten={laeuft || entwurf === adresse}
        aufKlick={() => {
          setzen(kennung, entwurf)
        }}
      >
        {eigene.kontakteAdresseMerken}
      </Knopf>
    </div>
  )
}
