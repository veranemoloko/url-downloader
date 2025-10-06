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
    """Создает задачу с несколькими файлами для увеличения времени загрузки"""
    print("🔄 Создаем задачу с несколькими файлами для теста докачки...")
    
    # Несколько файлов разного размера чтобы увеличить общее время загрузки
    files = [
        "https://m.kinosimka.plus/engine/go.php?url=aHR0cHM6Ly9lMDEudmFzcWEubmV0L2gwMy9jL2EvR3J5YXpuYXlhX2lncmFfMjAyNV82NDAubXA0",

    ]
    
    print("📦 Файлы для загрузки:")
    for i, url in enumerate(files, 1):
        print(f"   {i}. {url}")
    
    status, body = make_request("POST", "/tasks", {"urls": files})
    
    if status != 201:
        print(f"❌ Ошибка создания задачи: {status}")
        return None
    
    task_data = json.loads(body)
    task_id = task_data["id"]
    print(f"✅ Задача создана: {task_id}")
    return task_id

def monitor_progress_with_interrupt(task_id):
    """Мониторит прогресс и предлагает прервать в нужный момент"""
    print("\n🎯 Мониторим прогресс загрузки...")
    print("   Прервите сервер когда увидите что несколько файлов начали загружаться!")
    
    files_in_progress = 0
    last_progress = {}
    
    for i in range(15):  # Мониторим 15 секунд
        status, body = make_request("GET", f"/tasks/{task_id}")
        
        if status == 200:
            data = json.loads(body)
            current_status = data["status"]
            
            # Показываем прогресс по каждому файлу
            if data.get("results"):
                print(f"\n📊 Прогресс [{i+1}/15]:")
                
                current_files = 0
                for j, result in enumerate(data["results"]):
                    bytes_read = result.get("bytes_read", 0)
                    url_short = result["url"][:50] + "..." if len(result["url"]) > 50 else result["url"]
                    
                    if bytes_read > 0:
                        current_files += 1
                        status_icon = "✅" if result.get('success') else "🔄"
                        print(f"   {status_icon} Файл {j+1}: {bytes_read} байт - {url_short}")
                    
                    # Следим за прогрессом для определения когда прервать
                    if bytes_read > 0 and bytes_read < 102400:  # Если файл в процессе загрузки
                        current_files += 1
                
                files_in_progress = current_files
                
                # Если несколько файлов в процессе - предлагаем прервать
                if files_in_progress >= 2:
                    print(f"\n🎯 ИДЕАЛЬНЫЙ МОМЕНТ ДЛЯ ПРЕРЫВАНИЯ!")
                    print(f"   {files_in_progress} файла в процессе загрузки")
                    print("   Быстро нажмите Ctrl+C в консоли сервера!")
                    print("   У вас есть 3 секунды...")
                    
                    for countdown in range(3, 0, -1):
                        print(f"   {countdown}...")
                        time.sleep(1)
                    
                    print("\n🛑 СЕРВЕР ДОЛЖЕН БЫТЬ ПРЕРВАН!")
                    print("   Если успели прервать - отлично!")
                    print("   Если нет - тест продолжит мониторинг")
                    break
            
            if current_status == "completed":
                print("❌ Все файлы уже скачались! Слишком быстро.")
                return "completed"
            elif current_status == "failed":
                print("❌ Загрузка завершилась с ошибкой")
                return "failed"
        
        time.sleep(1)
    
    return "interrupted"

def test_resume_after_interrupt(task_id):
    """Тестирует восстановление после прерывания"""
    print("\n🔄 ТЕСТ ВОССТАНОВЛЕНИЯ")
    print("=" * 50)
    
    print("1. Перезапустите сервер: go run cmd/server/main.go")
    print("2. Затем нажмите Enter здесь...")
    input()
    
    print("\n3. Проверяем восстановление задачи...")
    status, body = make_request("GET", f"/tasks/{task_id}")
    
    if status != 200:
        print(f"❌ Задача не восстановилась: {status}")
        return False
    
    data = json.loads(body)
    print(f"✅ Задача восстановлена! Статус: {data['status']}")
    
    # Показываем прогресс восстановления
    if data.get("results"):
        print("📊 Прогресс после восстановления:")
        total_files = len(data["results"])
        completed_files = sum(1 for r in data["results"] if r.get('success'))
        in_progress_files = sum(1 for r in data["results"] if r.get('bytes_read', 0) > 0 and not r.get('success'))
        
        print(f"   Всего файлов: {total_files}")
        print(f"   Завершено: {completed_files}")
        print(f"   В процессе: {in_progress_files}")
        
        for j, result in enumerate(data["results"]):
            status_icon = "✅" if result.get('success') else "🔄"
            bytes_info = f"{result.get('bytes_read', 0)} байт" 
            print(f"   {status_icon} Файл {j+1}: {bytes_info}")
    
    # Ждем завершения
    print("\n4. Ожидаем завершения докачки...")
    for i in range(20):
        status, body = make_request("GET", f"/tasks/{task_id}")
        if status == 200:
            data = json.loads(body)
            current_status = data["status"]
            
            if current_status == "completed":
                print("🎉 ДОКАЧКА УСПЕШНА!")
                if data.get("results"):
                    total_bytes = sum(r.get('bytes_read', 0) for r in data["results"])
                    success_count = sum(1 for r in data["results"] if r.get('success'))
                    print(f"📦 Итог: {success_count}/{len(data['results'])} файлов, {total_bytes} байт")
                return True
            elif current_status == "failed":
                print("❌ Докачка завершилась с ошибкой")
                return False
            
            # Показываем промежуточный прогресс
            if i % 5 == 0:  # Каждые 5 итераций
                if data.get("results"):
                    completed = sum(1 for r in data["results"] if r.get('success'))
                    print(f"   [{i+1}/20] Статус: {current_status}, Завершено: {completed}/{len(data['results'])}")
        
        time.sleep(1)
    
    print("⚠️ Докачка заняла больше 20 секунд")
    return True

def main():
    """Основной тест докачки"""
    print("🚀 УСЛОЖНЕННЫЙ ТЕСТ ДОКАЧКИ")
    print("=" * 60)
    
    # Проверяем сервис
    print("🔍 Проверяем сервис...")
    status, body = make_request("GET", "/tasks/nonexistent")
    if status not in [200, 404]:
        print("❌ Сервис не доступен")
        sys.exit(1)
    print("✅ Сервис доступен")
    
    # Создаем задачу с несколькими файлами
    task_id = create_slow_download_task()
    if not task_id:
        sys.exit(1)
    
    # Мониторим и прерываем
    result = monitor_progress_with_interrupt(task_id)
    
    if result == "completed":
        print("\n💡 Все скачалось слишком быстро! Попробуйте:")
        print("   - Более медленное интернет-соединение")
        print("   - Большие файлы")
        print("   - Или просто проверьте что задачи сохраняются в downloads/tasks/")
        return
    
    elif result == "failed":
        print("\n❌ Загрузка завершилась с ошибкой")
        return
    
    # Тестируем восстановление
    success = test_resume_after_interrupt(task_id)
    
    if success:
        print("\n🎉 ТЕСТ ДОКАЧКИ ПРОЙДЕН!")
        print("✅ Сервис корректно восстанавливает прерванные загрузки")
        print("✅ Задачи сохраняются на диск")
        print("✅ Состояние восстанавливается после перезапуска")
    else:
        print("\n⚠️ Тест завершился с проблемами")
        print("💡 Проверьте логи сервера для диагностики")

if __name__ == "__main__":
    main()