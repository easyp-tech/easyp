# Автодополнение

## Автодополнение для zsh

1. Добавьте следующую строку в ваш файл запуска `~/.zshrc`:

```bash
source <(easyp completion zsh)
```

2. Перезапустите оболочку или выполните:

```bash
source ~/.zshrc
```

## Автодополнение для Bash

1. Установите [bash-completion](https://github.com/scop/bash-completion#installation) и подключите его в `~/.bashrc`.
2. Добавьте следующую строку в файл запуска `~/.bashrc`:

```bash
source <(easyp completion bash)
```

3. Перезапустите оболочку или выполните:

```bash
source ~/.bashrc
```

## Примечания

- Команда `easyp completion <shell>` выводит скрипт автодополнения для соответствующей оболочки.
- Если используете другую оболочку (например, `fish`), можно сгенерировать вывод и сохранить вручную:
  ```bash
  easyp completion fish > ~/.config/fish/completions/easyp.fish
  ```
- При обновлении версии EasyP рекомендуется заново перезагрузить оболочку, чтобы подтянуть возможные изменения в автодополнении.