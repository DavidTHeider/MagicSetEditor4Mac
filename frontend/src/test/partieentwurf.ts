// **Der kleinste Partieentwurf, der durchgeht** — geteilt von zwei Läufen.
//
// **Eine eigene Datei**, weil `partie.test.tsx` und
// `partieleerlauf.test.tsx` denselben Entwurf schicken. Einer, der zweimal
// dastünde, liefe beim ersten neuen Feld auseinander (§10).

// einEntwurf ist der kleinste, der durchgeht.
export const einEntwurf = {
  spiel: 'magic',
  regelwerk: { kennung: '', fassung: 0, strenge: '', abgewaehlt: [] },
  wirt: '',
  sitze: [
    {
      kennung: 'sitz1',
      name: 'Anna',
      deckID: '1',
      leiheID: '',
      profilID: '',
    },
    {
      kennung: 'sitz2',
      name: 'Bert',
      deckID: '2',
      leiheID: '',
      profilID: '',
    },
  ],
}
