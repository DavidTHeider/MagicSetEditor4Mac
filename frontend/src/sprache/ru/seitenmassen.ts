// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'Предпросмотр',
  vorschauErklaerung:
    'Предпросмотр показывает по каждой карте, что изменилось бы. Запись происходит только после нажатия «Применить» — и только если до этого никто ничего не записал.',
  karte: 'Карта',
  alt: 'Было',
  neu: 'Станет',
  keinWertBisher: 'нет значения',
  bleibtGleich: 'без изменений — уже так',
  zusammenfassung:
    'Карт будет изменено: {geaendert}, останутся прежними: {gleich}.',
  nichtsZuTun: 'У всех выбранных карт уже такое значение. Менять нечего.',
  konflikt:
    'Эти карты изменились с момента предпросмотра. Ничего не записано — получи предпросмотр заново:',
  inzwischenGeaendert: 'изменено с тех пор',
  angewendet: 'Карт изменено: {anzahl}.',
} as const
