// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: 'Où en est cet échange',
  phasen: {
    gegenueber: 'Avec qui',
    angebot: 'Ce que vous échangez',
    zusagen: 'Réservations',
    verbuchen: 'Enregistrement',
  },
  erledigt: 'fait',
  jetzt: 'maintenant',
  kommt: 'à venir',
  schritte: 'Ce que tu peux faire maintenant',
  keinSchritt: 'Pour l’instant, il n’y a rien à faire de ta part ici.',
  neu: 'Un nouvel échange : rassemble ce qui va et vient, puis lance-le.',
  abgleich: 'Ce que vous pourriez échanger',
  unterwegs: 'Ce qui est en route',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'ton correspondant',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'Historique',
  verlaufWechsel: 'Révision {revision} : {von} → {nach} · {zeit}',
  verlaufNeu: 'Révision {revision} : {nach} · {zeit}',
  verlaufLeer: 'Aucun changement pour l’instant.',
  verlaufOhneHandel:
    'Choisis un échange — tu verras ici comment il en est arrivé là.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Ajouter une ligne : {seite}',
  zeileKurz: '+ Ligne',
  keinPreis: 'sans prix',
  differenzDir: '⇆ {betrag} en ta faveur',
  differenzGegenueber: '⇆ {betrag} en faveur de ton correspondant',
  ausgeglichen: '⇆ Équilibré',
  differenzTeil: '{zahl} sans prix — seul ce qui en a un est compté.',
  preisHinweis:
    'Prix du {datum}, issus de {quelle}. Ils sont indicatifs, pas contraignants.',
  keinePreise: 'Aucun prix n’est encore disponible pour ces cartes.',
  zeileVerwerfen: 'Annuler',
}
