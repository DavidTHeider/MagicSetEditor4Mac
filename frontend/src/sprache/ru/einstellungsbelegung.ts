// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: 'Назначение ⌘0–⌘9',
  belegungHinweis:
    'Какую страницу открывает цифра вместе с ⌘. По умолчанию это разделы боковой панели; вариант «Как до 0.6.0» возвращает назначение до версии 0.5.0, таблица которого приведена в журнале изменений.',
  belegungFuer: 'Страница для {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'Как до 0.6.0',
  belegungVorgabe: 'Вернуть значение по умолчанию',
  kuerzel: {
    palette: 'Открыть палитру команд',
    suche: 'Искать в наборе',
    zurueck: 'Отменить изменение',
    vor: 'Повторить изменение',
    sichern: 'Сохранить набор',
    sichernUnter: 'Сохранить набор как …',
    oeffnen: 'Открыть набор …',
    neu: 'Создать новый набор',
    schliessen: 'Закрыть набор',
    drucken: 'Перейти к листу печати',
    navigator: 'Показать или скрыть навигатор',
    inspector: 'Показать или скрыть инспектор',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: 'Предыдущая страница',
    seiteVor: 'Следующая страница',
  },
} as const
