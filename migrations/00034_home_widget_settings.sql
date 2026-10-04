-- TASK-011 п.6: настройки видимости блоков на главной («Что показывать на главной»).
ALTER TABLE user_settings
    ADD COLUMN IF NOT EXISTS home_widgets JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN user_settings.home_widgets IS
  'Map widget_key -> bool. Missing keys mean default (visible). Known keys: tasks, habits, finance, notes, calendar, reminders, health, debts, analytics.';
