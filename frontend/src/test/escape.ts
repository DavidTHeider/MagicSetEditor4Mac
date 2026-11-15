// **Wer Escape verbraucht, sagt es** — die Abmachung, auf die sich die
// `Schublade` verlässt (UX-08,
// #1680):
// sie schließt nur bei einem Escape, das niemand darin mit `preventDefault`
// für sich genommen hat.
//
// merkeEscape hängt einen Horcher ans Fenster, **in der Capture-Phase** und
// damit vor jedem anderen, und gibt eine Funktion zurück, die ihn wieder
// abnimmt und sagt, ob das zuletzt gedrückte Escape verbraucht wurde — oder
// `undefined`, wenn gar keines ankam. Das Ereignis behält `defaultPrevented`
// über seine Zustellung hinaus; deshalb genügt es, es festzuhalten.
export function merkeEscape(): () => boolean | undefined {
  let gemerkt: KeyboardEvent | undefined

  const merken = (ereignis: KeyboardEvent) => {
    if (ereignis.key === 'Escape') {
      gemerkt = ereignis
    }
  }

  window.addEventListener('keydown', merken, true)

  return () => {
    window.removeEventListener('keydown', merken, true)

    return gemerkt?.defaultPrevented
  }
}
