package agent

import (
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
)

func buildSystemPrompt(tools *Registry, memories []ai.MemorySnippet, language string) string {
	var b strings.Builder
	b.WriteString("Ты — личный ассистент LifeOS в Telegram. Отвечай кратко, по-человечески.\n")
	b.WriteString("Вход может быть текстом или расшифровкой голоса/кружочка — учитывай возможные огрехи STT.\n")
	b.WriteString("Ты можешь задавать уточняющие вопросы, если не хватает данных.\n")
	b.WriteString("Для любых побочных эффектов (создать задачу, записать расход, сохранить память и т.п.) ")
	b.WriteString("ты ОБЯЗАН вызвать соответствующий инструмент. Никогда не утверждай, что действие выполнено, ")
	b.WriteString("пока не получил результат инструмента.\n")
	b.WriteString("Учитывай личные воспоминания пользователя, если они есть.\n")
	b.WriteString("Правила выбора действия:\n")
	b.WriteString("- Приветствие, «как дела», благодарность, болтовня без действия → type=reply (тепло и коротко).\n")
	b.WriteString("- Вопрос «что умеешь?» → type=reply со списком возможностей без вызова tools.\n")
	b.WriteString("- Расход/доход с суммой → finance.expense / finance.income (не задача).\n")
	b.WriteString("- Напоминание → reminder.create (нужны message и time_text; иначе ask).\n")
	b.WriteString("- Привычка → habit.create / habit.track.\n")
	b.WriteString("- Список дел на сегодня → task.list_today.\n")
	b.WriteString("- Создать задачу → task.create; title = конкретная формулировка из запроса пользователя ")
	b.WriteString("(например «купить хлеб»), НИКОГДА не подставляй заглушки вроде «разобрать входящие/почту», «задача», «todo».\n")
	b.WriteString("- Долги → finance.list_debts / create_debt / pay_debt; кэшфлоу → finance.cash_flow; план → finance.list_plan / create_planned.\n")
	b.WriteString("- Заметки → note.create / note.list / note.search.\n")
	b.WriteString("- Календарь → calendar.create / calendar.list_today.\n")
	b.WriteString("- Проекты → project.create / project.list; сферы → sphere.list / sphere.create.\n")
	b.WriteString("- Вес/шаги/сон → health.*; контакты/навыки → career.*.\n")
	b.WriteString("- Перегруз дня → plan.triage; доступность → plan.set_availability.\n")
	b.WriteString("- Отмена/перенос задачи → task.cancel / task.reschedule; всё на завтра → task.reschedule_all.\n")
	b.WriteString("- Проект: archive / tasks / progress; заметку удалить → note.delete.\n")
	b.WriteString("- Настройки обзоров/тихих часов → settings.morning_review / evening_review / quiet_hours.\n")
	b.WriteString("- Приоритеты → query.priorities; аналитика → analytics.summary.\n")
	b.WriteString("- Если неясно, что сделать — type=ask, а не угадывай tool.\n")
	b.WriteString("- Светская болтовня без действия → type=reply (не создавай пустые задачи).\n")
	b.WriteString("- Команды вроде /start обрабатывает бот отдельно — не пытайся их эмулировать tools.\n")
	if language != "" {
		b.WriteString("Язык ответа: ")
		b.WriteString(language)
		b.WriteString(".\n")
	} else {
		b.WriteString("Язык ответа: русский, если пользователь не пишет на другом языке.\n")
	}
	b.WriteString("\nВыводи ТОЛЬКО JSON одного из видов:\n")
	b.WriteString(`{"type":"ask","text":"..."}` + " — уточняющий вопрос\n")
	b.WriteString(`{"type":"reply","text":"..."}` + " — обычный ответ пользователю\n")
	b.WriteString(`{"type":"tool","tool":"name","args":{...}}` + " — вызов инструмента\n")
	b.WriteString("Поле text опционально для type=tool.\n")

	if tools != nil {
		list := tools.List()
		if len(list) > 0 {
			b.WriteString("\nДоступные инструменты:\n")
			for _, t := range list {
				b.WriteString("- ")
				b.WriteString(t.Name)
				if t.Description != "" {
					b.WriteString(": ")
					b.WriteString(t.Description)
				}
				if t.Parameters != "" {
					b.WriteString(" | params: ")
					b.WriteString(t.Parameters)
				}
				b.WriteString("\n")
			}
		}
	}

	if len(memories) > 0 {
		b.WriteString("\nЛичная память:\n")
		for _, m := range memories {
			b.WriteString(fmt.Sprintf("- [%s] %s = %s\n", m.Kind, m.Key, m.Value))
		}
	}

	return b.String()
}

func serializeUserBlob(req ai.DialogueRequest, history []ai.DialogueTurn) string {
	var b strings.Builder
	if len(history) > 0 {
		b.WriteString("История диалога:\n")
		for _, turn := range history {
			switch turn.Role {
			case "user":
				b.WriteString("user: ")
				b.WriteString(turn.Content)
				b.WriteString("\n")
			case "assistant":
				b.WriteString("assistant: ")
				b.WriteString(turn.Content)
				b.WriteString("\n")
			case "tool":
				b.WriteString("tool ")
				b.WriteString(turn.ToolName)
				b.WriteString(" args=")
				b.WriteString(fmt.Sprint(turn.ToolArgs))
				b.WriteString(" result=")
				b.WriteString(turn.ToolResult)
				b.WriteString("\n")
			default:
				b.WriteString(turn.Role)
				b.WriteString(": ")
				b.WriteString(turn.Content)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("Сообщение пользователя:\n")
	b.WriteString(req.Text)
	return b.String()
}
