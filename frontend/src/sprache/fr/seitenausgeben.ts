// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: 'Exporter',
  erklaerung:
    'Cinq façons de sortir un set de cette application — et à quoi sert chacune.',
  wohin: 'Vers où ?',
  keinZiel: 'Aucune destination choisie',
  nichtGemerkt:
    'La cible s’applique, mais elle n’a pas été mémorisée — elle ne sera pas présélectionnée la prochaine fois :',
  keinZielErklaerung:
    'Choisis sous « Vers où ? » sous quelle forme le set doit sortir — ce qu’il faut régler apparaîtra ici.',
  ohneKarten:
    'Les images et la planche d’impression n’auraient rien à montrer. Les cartes naissent sous « Cartes » — créées, collées ou importées.',
  ziele: {
    setdatei: {
      name: 'Fichier de set',
      kurz: 'Le set lui-même, pour le rouvrir — un fichier `.mse-set` que Magic Set Editor 2 lit aussi.',
      lang: 'Le raccourci est ⌘S. **Ne sont pas inclus** : les decks, les exemplaires, les prix, les étiquettes et l’origine des valeurs individuelles — ils restent dans la bibliothèque, car MSE2 abandonne à l’enregistrement ce qu’il ne connaît pas.',
    },
    bilder: {
      kurz: 'Un fichier image par carte, à réutiliser ailleurs — un forum, un tableur, un programme de jeu.',
    },
    druckbogen: {
      kurz: 'Les cartes côte à côte sur du papier, avec des traits de coupe.',
    },
    vorlage: {
      kurz: 'Du texte au lieu d’une image : une liste, un tableau, un fichier pour un autre programme.',
      lang: 'Les modèles disponibles dépendent du paquet installé.',
    },
    mtgjson: {
      kurz: 'Le format de cartes de mtgjson.com — pour les outils qui le lisent.',
      lang: 'Pas pour continuer à travailler dans cette application : ce que MTG JSON ne peut pas contenir est indiqué après l’export.',
    },
  },
  vorlagenEingrenzen: 'Filtrer les modèles',
  weitereOptionen: 'Autres options ({zahl})',
  // **„Ce que cela donnera"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'Ce que cela donnera',
    rechnet: 'Calcul en cours…',
    hinweis:
      'Rien n’est écrit tant que tu n’appuies pas sur le bouton ci-dessous.',
    bogen: {
      jeSeite: '{zahl} par page · {seiten}',
      masse: '{karten} · papier {breite} × {hoehe} mm · {dpi} dpi',
      letzte: 'Sur la dernière page : {karten}.',
    },
    vorlage: {
      datei: 'Fichier : {name} ({groesse})',
      karten: 'Évalué sur {karten}.',
      beigaben: 'Fichiers à côté : {zahl} (dossier « {ordner} », {groesse})',
      mindestens: 'au moins {groesse}',
      befunde: 'Relevé lors de l’évaluation :',
    },
  },
}
