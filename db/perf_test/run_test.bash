#!/bin/bash

# Параметры
URL="http://localhost:8001"
SCRIPT="get_data.lua"
THREADS=4
CONNECTIONS=100
DURATION="30s"
OUTDIR="results"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
OUTFILE="$OUTDIR/result-$TIMESTAMP.txt"

# Создаем директорию для результатов
mkdir -p "$OUTDIR"

# Запуск теста
echo "🔁 Запуск нагрузки на $URL с помощью $SCRIPT"
wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" -s "$SCRIPT" "$URL" | tee "$OUTFILE"

# Вывод информации
echo -e "\n✅ Результаты сохранены в: $OUTFILE"
