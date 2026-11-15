// **Der Fenstertitel nennt das offene Set** — AP-7
// (#2364),
// Befund B-06. Wie ein Dokumentfenster unter macOS: der Name des Dokuments
// zuerst, die Anwendung danach; ohne offenes Set nur sie. Das Fenster- und
// das Dockmenü nennen damit, woran gearbeitet wird.
import { useEffect } from 'react'

import { laufzeit } from '../fenster'
import { useTexte } from '../sprache/kontext'

export function useFenstertitel(setname: string | undefined): void {
  const anwendung = useTexte().allgemein.anwendungsname

  useEffect(() => {
    laufzeit()?.TitelSetzen?.(
      setname === undefined || setname === ''
        ? anwendung
        : `${setname} — ${anwendung}`,
    )
  }, [setname, anwendung])
}
