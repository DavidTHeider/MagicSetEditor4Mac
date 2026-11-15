// **Die Schranken des Handels** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungennetz.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// handelsnamen sind die Beschriftungen der fünf Schranken.
export const handelsnamen = {
  'richtlinie.handel': 'Handeln erlauben',
  'richtlinie.handel.markt': 'Handel bei einem fremden Anbieter erlauben',
  'richtlinie.handel.menge': 'Höchstmenge je Position',
  'richtlinie.handel.wert':
    'Höchstwert eines Angebots in der kleinsten Einheit (bei EUR: Cent)',
  'richtlinie.handel.frist': 'Längste Zusagefrist in Minuten',
}

// handelshilfe ist der erklärende Satz je Schranke.
export const handelshilfe = {
  'richtlinie.handel':
    'Schaltet Teilen, Deckleihe, Tausch und Markt zugleich ab. **Eine Schranke, keine Vorliebe**: was hier aus ist, lässt sich in keinem Profil wieder anschalten.',
  'richtlinie.handel.markt':
    'Vorgabe ist **aus**. Der Markt greift auf einen fremden Anbieter und auf Geld — alles andere bleibt unter zwei Menschen.',
  'richtlinie.handel.menge':
    'Wie viele Stücke eine einzelne Zeile eines Angebots höchstens nennen darf. **Leer heißt: keine Grenze.**',
  'richtlinie.handel.wert':
    'Was ein Angebot zusammen höchstens wert sein darf. Gilt nur, wenn die Bewertung **Vertragsinhalt** ist — eine bloße Marktansicht begrenzt sie nicht. **Leer heißt: keine Grenze.**',
  'richtlinie.handel.frist':
    'Wie lange eine Zusage Exemplare festhalten darf. **Leer heißt: keine Grenze.**',
}

// handelswerte sind die Stellungen der zwei Schalter.
export const handelswerte = {
  'richtlinie.handel.aus': 'Aus',
  'richtlinie.handel.an': 'An',
  'richtlinie.handel.markt.aus': 'Aus',
  'richtlinie.handel.markt.an': 'An',
}
