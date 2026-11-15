// Die Texte der **Symbolbibliothek** — DOM-02
// (#1697).

export const symbolbibliothek = {
  ueberschrift: 'Bibliothèque de symboles',
  erklaerung:
    'Symboles de ce profil qui peuvent être placés dans n’importe quel set. Une copie est placée : le set la conserve à l’enregistrement, même si le symbole est supprimé ici par la suite.',

  name: 'Nom du symbole de set',
  aufnehmen: 'Ajouter le symbole du set',
  einlesen: 'Importer un fichier de symbole …',
  suche: 'Rechercher des symboles',

  leer: 'Aucun symbole pour l’instant. Ajoute le symbole de ce set ou importe un fichier .mse-symbol — il pourra ensuite être placé dans tout autre set.',
  keinTreffer:
    'Aucun symbole ne correspond à la recherche et au filtre d’étiquettes. Modifie l’un des deux.',
  liste: 'Symboles de la bibliothèque',
  vorschau: 'Aperçu de {name}',
  verwendet: 'Utilisé par {sets}',
  unverwendet: 'Aucun set ne l’utilise.',

  einsetzen: 'Placer dans ce set',
  eingesetzt: '{name} est maintenant le symbole de ce set.',
  duplizieren: 'Dupliquer',
  kopieName: '{name} (copie)',
  umbenennen: 'Renommer',
  neuerName: 'Nouveau nom pour {name}',
  namenUebernehmen: 'Appliquer le nom',
  umbenennenAbbrechen: 'Annuler',
  ausgeben: 'Exporter …',
  ausgegeben: 'Exporté vers {pfad}',
  eingelesen: '{name} figure maintenant dans la bibliothèque.',
  aufgenommen: '{name} figure maintenant dans la bibliothèque.',

  loeschen: 'Supprimer',
  loeschenFrage: 'Supprimer {name} de la bibliothèque ?',
  loeschenFolgen:
    'Ces sets l’utilisent et conservent leur copie : {sets}. Seule l’entrée de la bibliothèque est perdue — et avec elle la trace du fait qu’ils utilisent le même symbole.',
  loeschenOhneFolgen:
    'Aucun set ne l’utilise. La suppression est irréversible.',
  loeschenJa: 'Supprimer',
  loeschenNein: 'Annuler',

  schriften: 'Polices de symboles',
  schriftenErklaerung:
    'Les polices de symboles sont des paquets de cet ordinateur — symboles de mana et glyphes similaires qu’un style place dans des champs de texte. Elles s’installent sous « Paquets » et ne sont affichées ici que pour consultation.',
  keineSchriften:
    'Aucune police de symboles installée. Il est possible d’en récupérer une sous « Paquets ».',
  schriftMitFassung: '{name} · version {fassung}',
  schriftenListe: 'Polices de symboles installées',
  schriftenZeigen: 'Polices de symboles installées ({zahl})',

  bearbeiten: 'Modifier le dessin',
  bearbeitungsueberschrift: '{name} de la bibliothèque',
  bearbeitungserklaerung:
    'Tu modifies un symbole de la bibliothèque. L’enregistrement crée une nouvelle version ; la version actuelle est conservée comme précédente. Les sets gardent leur copie tant que tu ne la leur appliques pas.',
  zumSetsymbol: 'Retour au symbole du set',
  inBibliothekSichern: 'Enregistrer dans la bibliothèque',
  inBibliothekGesichert: 'Nouvelle version de {name} enregistrée.',
  veraltet: 'Ces sets utilisent encore la version précédente : {sets}.',
  keineVeraltet: 'Aucun set n’utilise la version précédente.',
  nachziehen: 'Appliquer à ces sets',
  nachgezogen: 'Appliqué à {sets}. Leurs cartes sont redessinées.',
  zuruecknehmen: 'Restaurer la version précédente',
  zurueckgeholt: 'La version précédente de {name} est de retour.',

  fassungen: 'Dans les variantes de ce set',
  fassungenZu: 'Fermer les variantes',
  fassungsliste: 'Variantes de {name}',
  keineFassungen:
    'Le style de ce set ne nomme aucune variante pour le symbole de set. Choisis un style avec des variantes pour voir le symbole dans celles-ci.',
  fassungsbild: '{name} en {fassung}, {groesse} points',
}
