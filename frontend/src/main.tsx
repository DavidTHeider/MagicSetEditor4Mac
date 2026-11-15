// Einhängepunkt der Oberfläche.
//
// Hier steht nur noch der React-Aufruf. Alles, was sich unter jsdom prüfen
// lässt, steht in `einhaengen.ts` — die dateiweite Ausnahme in `.covignore`
// deckt damit wirklich nur, was ohne Fenster nicht auszuführen ist.
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Call, Events, Window } from '@wailsio/runtime'

import { App } from './App'
import { anbinden } from './anbindung'
import { Wurzelgrenze } from './rahmen/Wurzelgrenze'
import { wurzelSuchen } from './einhaengen'
import { installiereTastaturrollen } from './tastaturrollen'
import './stil.css'
import './schriftreihen.css'
import './themen.css'
import './barrierefreiheit.css'

// **Vor dem ersten Ruf** — Wails 3 legt nichts mehr am Fenster ab, und die
// Fassade sucht dort (ENG-02, `anbindung.ts`).
anbinden(Call, Events, globalThis, Window)

// **Pfeiltasten rollen jeden Rollbereich** — WebKit tut es nicht selbst
// (`tastaturrollen.ts`).
installiereTastaturrollen(document)

createRoot(wurzelSuchen(document)).render(
  <StrictMode>
    <Wurzelgrenze>
      <App />
    </Wurzelgrenze>
  </StrictMode>,
)
