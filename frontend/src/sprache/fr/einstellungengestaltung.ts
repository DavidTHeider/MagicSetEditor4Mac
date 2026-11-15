// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: 'Palette de couleurs',
  schriftbereich: 'La police d’affichage s’applique à',
  hintergrund: 'Arrière-plan de la fenêtre',
  gestaltung: 'Couleurs personnalisées',
}

export const gestaltungswerte = {
  'farbwelt.pergament': 'Parchemin — chaleureux, avec du grain',
  'farbwelt.klar': 'Clair — neutre, sans grain',
  'farbwelt.schiefer': 'Ardoise — gris-bleu froid',
  'farbwelt.moos': 'Mousse — vert atténué',
  'schriftbereich.ueberschriften': 'Titres uniquement',
  'schriftbereich.ganz': 'Toute l’interface',
  'hintergrund.aus': 'Désactivé',
  'hintergrund.fuellen': 'Remplir',
  'hintergrund.einpassen': 'Ajuster',
  'hintergrund.kacheln': 'Mosaïque',
  'schriftart.buch': 'Livre — New York',
  'schriftart.rund': 'Arrondie — SF Pro Rounded',
  'schriftart.serifenlos': 'Sans empattement — la police du système',
}

export const gestaltungshinweise = {
  farbwelt:
    'Les couleurs de l’interface, en clair comme en sombre, toutes recalculées. Avec un contraste élevé, l’encre s’applique, quel que soit le choix fait ici.',
  schriftbereich:
    'Toute l’interface n’accepte qu’une police de texte ; une police décorative reste réservée aux titres. Les cartes conservent les polices de leur modèle.',
  hintergrund:
    'Place ta propre image sous le fond, sous un voile qui garde le texte lisible. Tu choisis l’image ci-dessous ; **Désactivé** la retire aussitôt.',
  gestaltung:
    'Des couleurs personnalisées pour les dix propriétés de l’interface, une par ligne. Uniquement des couleurs hexadécimales — ni adresses ni polices. Tout ce qui serait illisible ne s’applique pas.',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'Ton arrière-plan de fenêtre',
  vorschau: 'Aperçu de l’arrière-plan de la fenêtre',
  keinBild:
    'Pas encore d’image personnelle. Choisis-en une — elle apparaît ici et, dès que l’arrière-plan de la fenêtre n’est plus désactivé, derrière l’interface.',
  aus: 'L’image est prête, mais l’arrière-plan de la fenêtre est réglé sur **Désactivé**.',
  waehlen: 'Choisir une image …',
  entfernen: 'Retirer l’image',
  merkmale: 'Couleurs personnalisées',
  merkmaleHilfe:
    'Une entrée par ligne, par exemple « --mse-marke: #4e2f16; ». Sont autorisés : {merkmale}.',
  uebernehmen: 'Appliquer',
  zeilenfehler:
    'Ligne {zeile} : une entrée telle que « --mse-grund: #1e1b16; » est attendue — l’une des propriétés autorisées et une couleur hexadécimale.',
  nichtLesbar: 'Non appliqué — ce serait illisible :',
  paar: '{vorn} sur {hinten} : {verhaeltnis} : 1, {schwelle} : 1 requis',
  paarnamen: {
    'status:warnung': 'Avertissements',
    'status:erfolg': 'Messages de réussite',
    'fensterbild:dunkel': '--mse-grund sur une image de fenêtre sombre',
    'fensterbild:hell': '--mse-grund sur une image de fenêtre claire',
  } as Record<string, string>,
  angewandt: 'Appliqué.',
  ruhtBeiKontrast:
    'Enregistré — les couleurs personnalisées ne s’appliquent pas tant que le contraste élevé est activé ; elles reviennent dès qu’il est désactivé.',
  schmuckNurOben:
    'La police d’affichage choisie est décorative — elle reste réservée aux titres.',
  zuruecksetzenGescheitert: 'L’apparence n’a pas pu être réinitialisée :',
  zurVorgabe: 'Rétablir les couleurs par défaut',
}
