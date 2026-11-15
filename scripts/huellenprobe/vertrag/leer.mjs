// **Die Kohorte der leeren Listen** — C01
// (#2226).
//
// Gerufen **vor** jeder anderen Gruppe, auf einem Bestand, der außer dem
// Saatset nichts trägt: keine Exemplare, keine Kontakte, keine Partien. Eine
// leere Liste muss `[]` sein und nicht `null` — Go gibt für eine Liste ohne
// Anlage `null`, und das Fenster typisiert sie als Liste. Genau so stand
// `gebrauchtVon` in vier Paketantworten (C01, erster Rundgang).
//
// **Eine Kohorte und keine Gruppe**: dieselben Bindungen laufen später noch
// einmal mit Daten — die leere Antwort belegt die Form einer Liste ohne ein
// einziges Element, also **nicht** die Form ihrer Elemente.

const leer = (name) => ({ name, kohorte: 'leer' })

export const leereListen = {
  name: 'Leere Listen (Kohorte)',
  kohorte: true,
  rufe: ['Sammlung', 'SammlungLagerorte', 'SammlungWuensche', 'Portfolio', 'Kontakte', 'Peerkontakte',
    'Nachbarn', 'Partien', 'Deckleihen', 'Auftraege', 'Rueckkehren', 'Zugaenge', 'Weiterarbeiten',
    'Aufmerksamkeit', 'Metriken'].map(leer),
}
