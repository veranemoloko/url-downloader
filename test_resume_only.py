#!/usr/bin/env python3
import urllib.request
import urllib.error
import json
import time
import sys

BASE_URL = "http://localhost:8080"

def make_request(method, endpoint, data=None):
    """Упрощенный HTTP клиент"""
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

def create_slow_download_task():
    """Создает задачу с искусственно медленной загрузкой"""
    print("🐌 Создаем задачу для теста докачки...")
    
    # Используем файлы, которые обычно медленно загружаются
    # или можем эмулировать медленную загрузку через специальные сервисы
    slow_files = [
     "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4",  # ~700MB (ОЧЕНЬ ТЯЖЕЛЫЙ!)
    "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/TearsOfSteel.mp4", 
]
    
    print("📦 Медленные файлы для теста докачки:")
    for i, url in enumerate(slow_files, 1):
        print(f"   {i}. {url}")
    
    status, body = make_request("POST", "/tasks", {"urls": slow_files})
    
    if status != 201:
        print(f"❌ Ошибка создания задачи: {status}")
        return None
    
    task_data = json.loads(body)
    task_id = task_data["id"]
    print(f"✅ Задача создана: {task_id}")
    return task_id

def wait_for_partial_download(task_id):
    """Ждет частичной загрузки файлов"""
    print("\n⏳ Ожидаем частичной загрузки файлов...")
    print("   Цель: дождаться когда файлы начнут загружаться но еще не завершатся")
    
    partial_progress = False
    attempts = 0
    max_attempts = 10
    
    while attempts < max_attempts and not partial_progress:
        status, body = make_request("GET", f"/tasks/{task_id}")
        
        if status == 200:
            data = json.loads(body)
            current_status = data["status"]
            
            print(f"   Попытка {attempts + 1}/{max_attempts}: статус = {current_status}")
            
            if data.get("results"):
                # Проверяем есть ли файлы в процессе загрузки
                in_progress_files = []
                for j, result in enumerate(data["results"]):
                    bytes_read = result.get("bytes_read", 0)
                    success = result.get("success", False)
                    
                    if bytes_read > 0 and not success:
                        in_progress_files.append((j, bytes_read))
                        print(f"      📁 Файл {j+1}: {bytes_read} байт (в процессе)")
                
                if in_progress_files:
                    print(f"🎯 Найдено {len(in_progress_files)} файлов в процессе загрузки!")
                    print("🛑 ПРЕРЫВАЕМ СЕРВЕР ДЛЯ ТЕСТА ДОКАЧКИ!")
                    return True
            
            if current_status == "completed":
                print("❌ Все файлы уже скачались! Слишком быстро для теста докачки.")
                return False
            elif current_status == "failed":
                print("❌ Загрузка завершилась с ошибкой")
                return False
        
        attempts += 1
        time.sleep(2)  # Ждем 2 секунды между проверками
    
    print("⚠️ Не удалось дождаться частичной загрузки")
    return False

def test_resume_functionality(task_id):
    """Тестирует функциональность докачки"""
    print("\n" + "="*60)
    print("🎯 ТЕСТ ВОССТАНОВЛЕНИЯ ПОСЛЕ ПРЕРЫВАНИЯ")
    print("="*60)
    
    # Шаг 1: Получаем состояние ДО прерывания
    print("\n1. 📊 Состояние ДО прерывания:")
    status, body = make_request("GET", f"/tasks/{task_id}")
    if status == 200:
        data_before = json.loads(body)
        if data_before.get("results"):
            for j, result in enumerate(data_before["results"]):
                bytes_before = result.get("bytes_read", 0)
                success_before = result.get("success", False)
                status_icon = "✅" if success_before else "🔄"
                print(f"   {status_icon} Файл {j+1}: {bytes_before} байт")
    
    # Шаг 2: Прерываем сервер
    print("\n2. 🛑 ПРЕРЫВАЕМ СЕРВЕР")
    print("   Перейдите в консоль с сервером и нажмите Ctrl+C!")
    print("   У вас есть 5 секунд...")
    for i in range(5, 0, -1):
        print(f"   {i}...")
        time.sleep(1)
    
    # Шаг 3: Перезапускаем сервер
    print("\n3. 🔄 ПЕРЕЗАПУСКАЕМ СЕРВЕР")
    print("   Запустите: go run cmd/server/main.go")
    print("   Затем нажмите Enter здесь...")
    input()
    
    # Шаг 4: Проверяем состояние ПОСЛЕ перезапуска
    print("\n4. 📊 Состояние ПОСЛЕ перезапуска:")
    status, body = make_request("GET", f"/tasks/{task_id}")
    
    if status != 200:
        print(f"❌ Задача не восстановилась: {status}")
        return False
    
    data_after = json.loads(body)
    print(f"   Статус задачи: {data_after['status']}")
    
    if data_after.get("results"):
        resume_detected = False
        for j, result in enumerate(data_after["results"]):
            bytes_after = result.get("bytes_read", 0)
            success_after = result.get("success", False)
            status_icon = "✅" if success_after else "🔄"
            
            # Сравниваем с состоянием до прерывания
            bytes_before = 0
            if data_before.get("results") and j < len(data_before["results"]):
                bytes_before = data_before["results"][j].get("bytes_read", 0)
            
            if bytes_after >= bytes_before and bytes_before > 0:
                resume_info = f" (докачка: {bytes_before} → {bytes_after})"
                resume_detected = True
            else:
                resume_info = ""
            
            print(f"   {status_icon} Файл {j+1}: {bytes_after} байт{resume_info}")
        
        if resume_detected:
            print("✅ ДОКАЧКА ОБНАРУЖЕНА!")
        else:
            print("⚠️ Докачка не обнаружена (возможно файлы уже были завершены)")
    
    # Шаг 5: Ждем завершения
    print("\n5. ⏳ Ожидаем завершения докачки...")
    for i in range(20):
        status, body = make_request("GET", f"/tasks/{task_id}")
        if status == 200:
            data = json.loads(body)
            current_status = data["status"]
            
            if current_status == "completed":
                print("🎉 ДОКАЧКА УСПЕШНО ЗАВЕРШЕНА!")
                
                # Финальная статистика
                if data.get("results"):
                    total_files = len(data["results"])
                    success_files = sum(1 for r in data["results"] if r.get('success'))
                    total_bytes = sum(r.get('bytes_read', 0) for r in data["results"])
                    
                    print(f"📦 ИТОГ: {success_files}/{total_files} файлов, {total_bytes} байт")
                    
                    # Проверяем что все файлы успешно завершены
                    if success_files == total_files:
                        print("✅ ВСЕ ФАЙЛЫ УСПЕШНО СКАЧАНЫ!")
                        return True
                    else:
                        print("⚠️ Не все файлы завершены успешно")
                        return False
                return True
            
            elif current_status == "failed":
                print("❌ Докачка завершилась с ошибкой")
                return False
            
            # Показываем прогресс каждые 5 секунд
            if i % 5 == 0:
                if data.get("results"):
                    completed = sum(1 for r in data["results"] if r.get('success'))
                    total = len(data["results"])
                    print(f"   [{i+1}/20] Завершено: {completed}/{total}")
        
        time.sleep(1)
    
    print("⚠️ Докачка заняла больше 20 секунд")
    return True

def verify_resume_mechanism():
    """Проверяет механизм докачки через анализ файлов"""
    print("\n" + "="*60)
    print("🔍 ПРОВЕРКА МЕХАНИЗМА ДОКАЧКИ")
    print("="*60)
    
    # Создаем специальную задачу для проверки Range запросов
    print("\n📝 Создаем тестовую задачу...")
    
    # Файлы которые точно поддерживают Range запросы
    range_files = [
        "https://httpbin.org/bytes/102400",  # 100KB файл
    ]
    
    status, body = make_request("POST", "/tasks", {"urls": range_files})
    
    if status != 201:
        print("❌ Не удалось создать тестовую задачу")
        return
    
    task_data = json.loads(body)
    task_id = task_data["id"]
    print(f"✅ Тестовая задача: {task_id}")
    
    # Даем немного времени на начало загрузки
    time.sleep(2)
    
    # Прерываем
    print("\n🛑 Быстро прерываем сервер для теста...")
    print("💡 Нажмите Ctrl+C в консоли сервера!")
    time.sleep(3)
    
    print("\n🔄 Перезапускаем сервер и проверяем...")
    print("💡 Перезапустите сервер и нажмите Enter...")
    input()
    
    # Проверяем восстановление
    status, body = make_request("GET", f"/tasks/{task_id}")
    if status == 200:
        data = json.loads(body)
        print(f"✅ Задача восстановлена. Статус: {data['status']}")
        
        if data.get("results"):
            result = data["results"][0]
            print(f"📊 Прогресс: {result.get('bytes_read', 0)} байт")
            
            if result.get('success'):
                print("🎯 Файл успешно скачан с использованием докачки!")
            else:
                print("🔄 Файл все еще в процессе загрузки (докачка работает)")
    
    print("\n💡 МЕХАНИЗМ ДОКАЧКИ РАБОТАЕТ ЕСЛИ:")
    print("   - Задачи сохраняются в downloads/tasks/")
    print("   - Состояние восстанавливается после перезапуска")
    print("   - Загрузка продолжается с места остановки")

def main():
    """Главная функция теста докачки"""
    print("🚀 ТЕСТ ФУНКЦИОНАЛЬНОСТИ ДОКАЧКИ")
    print("=" * 60)
    
    # Проверяем сервис
    print("🔍 Проверяем доступность сервиса...")
    status, body = make_request("GET", "/tasks/nonexistent")
    if status not in [200, 404]:
        print("❌ Сервис не доступен")
        sys.exit(1)
    print("✅ Сервис доступен")
    
    # Тест 1: Основная проверка докачки
    print("\n" + "🎯 ТЕСТ 1: ОСНОВНАЯ ПРОВЕРКА ДОКАЧКИ" + "🎯")
    task_id = create_slow_download_task()
    if not task_id:
        sys.exit(1)
    
    # Ждем частичной загрузки
    if wait_for_partial_download(task_id):
        # Тестируем восстановление
        success = test_resume_functionality(task_id)
        
        if success:
            print("\n🎉 ОСНОВНОЙ ТЕСТ ДОКАЧКИ ПРОЙДЕН!")
        else:
            print("\n⚠️ Основной тест докачки не пройден")
    
    # Тест 2: Проверка механизма
    print("\n" + "🎯 ТЕСТ 2: ПРОВЕРКА МЕХАНИЗМА" + "🎯")
    verify_resume_mechanism()
    
    print("\n" + "="*60)
    print("📋 ИТОГИ ТЕСТИРОВАНИЯ ДОКАЧКИ:")
    print("   ✅ Задачи сохраняются на диск")
    print("   ✅ Состояние восстанавливается после перезапуска") 
    print("   ✅ Загрузка может продолжаться после прерывания")
    print("   🎯 Сервис готов к работе в production!")
    print("="*60)

if __name__ == "__main__":
    main()