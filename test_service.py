#!/usr/bin/env python3
import urllib.request
import urllib.error
import json
import time
import sys

BASE_URL = "http://localhost:8080"

def make_request(method, endpoint, data=None):
    """Упрощенный HTTP клиент без внешних зависимостей"""
    url = f"{BASE_URL}{endpoint}"
    
    if data:
        data = json.dumps(data).encode('utf-8')
    
    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={'Content-Type': 'application/json'}
    )
    
    try:
        with urllib.request.urlopen(req) as response:
            return response.getcode(), response.read().decode('utf-8')
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode('utf-8')
    except Exception as e:
        return None, str(e)

def test_service_health():
    """Проверяем что сервис запущен"""
    print("🔍 Проверяем доступность сервиса...")
    status, body = make_request("GET", "/tasks/nonexistent")
    
    if status in [200, 404]:
        print("✅ Сервис запущен и отвечает")
        return True
    else:
        print(f"❌ Сервис не доступен: {status} - {body}")
        return False

def test_create_task():
    """Тест создания задачи"""
    print("\n📝 Тестируем создание задачи...")
    
    test_urls = ["https://httpbin.org/robots.txt"]
    
    status, body = make_request("POST", "/tasks", {"urls": test_urls})
    
    if status == 201:
        data = json.loads(body)
        print(f"✅ Задача создана: ID={data['id']}")
        return data['id']
    else:
        print(f"❌ Ошибка создания задачи: {status} - {body}")
        return None

def test_get_task_status(task_id):
    """Тест получения статуса задачи"""
    print(f"\n📊 Проверяем статус задачи {task_id}...")
    
    status, body = make_request("GET", f"/tasks/{task_id}")
    
    if status == 200:
        data = json.loads(body)
        print(f"✅ Статус задачи: {data['status']}")
        return data
    else:
        print(f"❌ Ошибка получения статуса: {status} - {body}")
        return None

def test_invalid_urls():
    """Тест невалидных URL"""
    print("\n🚫 Тестируем валидацию URL...")
    
    invalid_urls = ["not-a-url", "ftp://example.com/file.txt"]
    
    status, body = make_request("POST", "/tasks", {"urls": invalid_urls})
    
    if status == 400:
        print("✅ Валидация URL работает корректно")
        return True
    else:
        print(f"❌ Ожидалась ошибка 400, получили: {status}")
        return False

def monitor_task_progress(task_id, timeout=30):
    """Мониторинг прогресса задачи"""
    print(f"\n🔄 Отслеживаем прогресс задачи (таймаут: {timeout}сек)...")
    
    for i in range(timeout):
        status, body = make_request("GET", f"/tasks/{task_id}")
        
        if status == 200:
            data = json.loads(body)
            current_status = data['status']
            print(f"  [{i+1}/{timeout}] Статус: {current_status}")
            
            if current_status == "completed":
                print("✅ Задача успешно завершена!")
                if 'results' in data:
                    for result in data['results']:
                        success = "✅" if result.get('success') else "❌"
                        print(f"   {success} {result['url']} - {result.get('bytes_read', 0)} bytes")
                return True
            elif current_status == "failed":
                print("❌ Задача завершилась с ошибкой")
                if 'results' in data:
                    for result in data['results']:
                        if 'error' in result:
                            print(f"   Ошибка: {result['error']}")
                return False
        
        time.sleep(1)
    
    print(f"❌ Задача не завершилась за {timeout} секунд")
    return False

def main():
    """Основная функция тестирования"""
    print("🚀 Запуск тестов Download Service...")
    
    # Тест 1: Проверка доступности
    if not test_service_health():
        print("\n💡 Совет: Убедитесь что сервер запущен: go run cmd/server/main.go")
        sys.exit(1)
    
    # Тест 2: Валидация URL
    if not test_invalid_urls():
        sys.exit(1)
    
    # Тест 3: Создание и отслеживание задачи
    task_id = test_create_task()
    if not task_id:
        sys.exit(1)
    
    # Даем время на старт обработки
    time.sleep(2)
    
    # Тест 4: Проверка статуса
    task_data = test_get_task_status(task_id)
    if not task_data:
        sys.exit(1)
    
    # Тест 5: Мониторинг до завершения
    if monitor_task_progress(task_id):
        print("\n🎉 Все тесты прошли успешно!")
    else:
        print("\n⚠️  Задача не завершилась успешно, но основные функции работают")

if __name__ == "__main__":
    main()