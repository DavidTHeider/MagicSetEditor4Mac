// **Welche Befehle das Fenster entgegennimmt.**
//
// Aus `Anwendung.tsx` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort steht, **was** die Anwendung zeichnet,
// hier, **welche Befehle** sie anmeldet.
//
// **Ein `useMemo` mit einer langen Abhängigkeitsliste**, und das ist der
// eigentliche Grund für die eigene Datei: die Liste ist so lang wie die
// Handlungen selbst, und sie zweimal untereinander zu lesen war der
// unangenehmste Teil von `Anwendung.tsx`.
//
// # Und der Umschlag gehört dazu
//
// **Jeder Befehl schließt die Einführung** (#1325). Sie ist ein Fenster mit
// `aria-modal`, die Menüleiste dagegen ist **nativ** — Wails zeichnet sie,
// nicht der DOM. ⌘O wirkte damit hinter dem Fenster, während es davor stehen
// blieb; zwei Läufe in `menueablage.test.tsx` haben es angestrichen.
//
// **Ein Umschlag um alle und keine Prüfung je Befehl**: welcher von ihnen
// etwas anzeigt, ist eine Frage, die sich hier nicht stellen soll — wer zum
// Menü greift, hat die Einführung gelesen oder will sie nicht.
import { useMemo } from 'react'

import type { Fensterhandlungen } from '../fenster'

interface Handlungsteile {
  einstellungenOeffnen: Fensterhandlungen['einstellungen']
  sichernJetzt: Fensterhandlungen['sichern']
  sichernUnter: Fensterhandlungen['sichernUnter']
  bestandSichern: Fensterhandlungen['bestandSichern']
  setOeffnen: Fensterhandlungen['setOeffnen']
  setNeu: Fensterhandlungen['setNeu']
  waehleSet: Fensterhandlungen['setZuletzt']
  setSchliessen: Fensterhandlungen['setSchliessen']
  drucken: Fensterhandlungen['drucken']
  hilfe: Fensterhandlungen['hilfe']
  rueckgaengig: Fensterhandlungen['rueckgaengig']
  wiederherstellen: Fensterhandlungen['wiederherstellen']
  seite: Fensterhandlungen['seite']
  ueber: Fensterhandlungen['ueber']
  seitenleiste: Fensterhandlungen['seitenleiste']
  seiteZurueck: Fensterhandlungen['seiteZurueck']
  seiteVor: Fensterhandlungen['seiteVor']
  dateiGeoeffnet: Fensterhandlungen['dateiGeoeffnet']
  dateiFehler: Fensterhandlungen['dateiFehler']
}

export function useFensterhandlungen(
  teile: Handlungsteile,
  vorJedemBefehl: () => void,
): Fensterhandlungen {
  const {
    einstellungenOeffnen,
    sichernJetzt,
    sichernUnter,
    bestandSichern,
    setOeffnen,
    setNeu,
    waehleSet,
    setSchliessen,
    drucken,
    hilfe,
    rueckgaengig,
    wiederherstellen,
    seite,
    ueber,
    seitenleiste,
    seiteZurueck,
    seiteVor,
    dateiGeoeffnet,
    dateiFehler,
  } = teile

  const roh = useMemo(
    () => ({
      einstellungen: einstellungenOeffnen,
      sichern: sichernJetzt,
      sichernUnter,
      bestandSichern,
      setOeffnen,
      setNeu,
      setZuletzt: waehleSet,
      setSchliessen,
      drucken,
      hilfe,
      rueckgaengig,
      wiederherstellen,
      seite,
      ueber,
      seitenleiste,
      seiteZurueck,
      seiteVor,
      dateiGeoeffnet,
      dateiFehler,
    }),
    [
      einstellungenOeffnen,
      sichernJetzt,
      sichernUnter,
      bestandSichern,
      setOeffnen,
      setNeu,
      waehleSet,
      setSchliessen,
      drucken,
      hilfe,
      rueckgaengig,
      wiederherstellen,
      seite,
      ueber,
      seitenleiste,
      seiteZurueck,
      seiteVor,
      dateiGeoeffnet,
      dateiFehler,
    ],
  )

  // **Der Umschlag geht über `Object.entries` und nicht Feld für Feld.**
  // Sechzehn Namen zweimal untereinander wären sechzehn Gelegenheiten, einen
  // zu vergessen — und ein vergessener fiele erst auf, wenn jemand ihn im
  // Menü benutzt, während die Einführung davor steht.
  //
  // **Die Zusicherung ist nötig**: `Object.entries` verliert die Namen, und
  // `Object.fromEntries` gibt `{ [k: string]: … }` zurück.
  return useMemo(
    () =>
      Object.fromEntries(
        Object.entries(roh).map(([name, handlung]) => [
          name,
          (...werte: unknown[]) => {
            vorJedemBefehl()

            return (handlung as (...werte: unknown[]) => unknown)(...werte)
          },
        ]),
      ) as unknown as Fensterhandlungen,
    [roh, vorJedemBefehl],
  )
}
