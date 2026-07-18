#!/bin/sh
set -e

# Entrypoint скрипт для приложения Crypto Pay
# Выполняет предварительные проверки перед запуском приложения

echo "=== Crypto Pay Application Entrypoint ==="

# Функция для проверки доступности PostgreSQL
wait_for_postgres() {
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
        echo "Warning: DB_HOST or DB_PORT not set, skipping PostgreSQL check"
        return 0
    fi

    echo "Waiting for PostgreSQL to be ready..."
    max_attempts=30
    attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
            echo "PostgreSQL is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        echo "Attempt $attempt/$max_attempts: PostgreSQL not ready, waiting 1 second..."
        sleep 1
    done

    echo "Error: PostgreSQL is not available after $max_attempts attempts"
    return 1
}

# Функция для проверки доступности NFS volumes
check_storage_volumes() {
    echo "Checking storage volumes..."
    
    # Определяем, нужно ли игнорировать ошибки монтирования
    # В development/local окружении NFS может быть недоступен локально
    ignore_errors=false
    if [ "$APP_ENV" = "development" ] || [ "$APP_ENV" = "local" ]; then
        ignore_errors=true
        echo "APP_ENV is '$APP_ENV' - NFS volume errors will be ignored"
    fi
    
    return 0
}

# Функция для проверки критических переменных окружения
check_required_env() {
    echo "Checking required environment variables..."
    
    required_vars="DB_PASSWORD JWT_SECRET CRYPT_KEY"
    missing_vars=""

    for var in $required_vars; do
        eval value=\$$var
        if [ -z "$value" ]; then
            missing_vars="$missing_vars $var"
        fi
    done

    if [ -n "$missing_vars" ]; then
        echo "Error: Missing required environment variables:$missing_vars"
        return 1
    fi

    echo "✓ All required environment variables are set"
    return 0
}

# Функция для проверки длины JWT_SECRET
check_jwt_secret_length() {
    if [ -n "$JWT_SECRET" ] && [ ${#JWT_SECRET} -lt 32 ]; then
        echo "Warning: JWT_SECRET should be at least 32 characters long"
    fi
}

# Функция для логирования информации о запуске
log_startup_info() {
    echo "=== Startup Information ==="
    echo "APP_ENV: ${APP_ENV:-not set}"
    echo "APP_PORT: ${APP_PORT:-8080}"
    echo "DB_HOST: ${DB_HOST:-not set}"
    echo "DB_PORT: ${DB_PORT:-5432}"
    echo "DB_NAME: ${DB_NAME:-crypto-pay}"
    echo "Working directory: $(pwd)"
    echo "=========================="
}

# Основная логика запуска
main() {
    # Логируем информацию о запуске
    log_startup_info

    # Проверяем критичные переменные окружения
    if ! check_required_env; then
        echo "Error: Environment validation failed"
        exit 1
    fi

    # Проверяем длину JWT_SECRET
    check_jwt_secret_length

    # Ожидаем готовности PostgreSQL (только если запускается приложение, не мигратор)
    if [ "$1" = "./crypto-pay" ] || [ -z "$1" ]; then
        wait_for_postgres || {
            echo "Warning: PostgreSQL check failed, but continuing..."
        }
    fi

    
    echo "=== Starting application ==="
    
    # Выполняем переданную команду (по умолчанию запуск приложения)
    exec "$@"
}

# Запускаем основную логику
main "$@"

