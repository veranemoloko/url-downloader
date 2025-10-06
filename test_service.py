#!/usr/bin/env python3
import urllib.request
import urllib.error
import json
import time
import sys

BASE_URL = "http://localhost:8080"

def make_request(method, endpoint, data=None):
    url = f"{BASE_URL}{endpoint}"
    if data:
        data = json.dumps(data).encode('utf-8')
    req = urllib.request.Request(url, data=data, method=method, headers={'Content-Type': 'application/json'})
    try:
        with urllib.request.urlopen(req) as response:
            return response.getcode(), response.read().decode('utf-8')
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode('utf-8')
    except Exception as e:
        return None, str(e)

def test_service_health():
    print("🐱 Checking if service is alive...")
    status, body = make_request("GET", "/tasks/nonexistent")
    if status in [200, 404]:
        print("🐈 Service is alive and responding!")
        return True
    print(f"❌ Service not responding: {status} - {body}")
    return False

def test_invalid_urls():
    print("\n🐾 Checking URL validation...")
    invalid_urls = ["not-a-url", "ftp://example.com/file.txt"]
    status, body = make_request("POST", "/tasks", {"urls": invalid_urls})
    if status == 400:
        print("🐱 URL validation works!")
        return True
    print(f"❌ Expected 400, got: {status}")
    return False

def test_create_task():
    print("\n🐈 Creating a test task...")
    urls = ["https://httpbin.org/bytes/1024", "https://httpbin.org/bytes/2048"]
    status, body = make_request("POST", "/tasks", {"urls": urls})
    if status == 201:
        data = json.loads(body)
        print(f"🐾 Task created! ID={data['id']}")
        return data['id']
    print(f"❌ Task creation failed: {status}")
    return None

def monitor_task(task_id, timeout=30):
    print(f"\n🐱 Monitoring task progress (max {timeout}s)...")
    for i in range(timeout):
        status, body = make_request("GET", f"/tasks/{task_id}")
        if status == 200:
            data = json.loads(body)
            current_status = data['status']
            print(f"  [{i+1}/{timeout}] Status: {current_status}")
            if current_status == "completed":
                print("🐈 Task completed successfully!")
                if 'results' in data:
                    for r in data['results']:
                        ok = "🐾" if r.get('success') else "❌"
                        print(f"   {ok} {r['url']} - {r.get('bytes_read', 0)} bytes")
                return True
            elif current_status == "failed":
                print("❌ Task failed")
                return False
        time.sleep(1)
    print("⚠️ Task did not finish in time")
    return False

def main():
    print("🐱 Starting Download Service Test!")
    if not test_service_health() or not test_invalid_urls():
        sys.exit(1)
    task_id = test_create_task()
    if not task_id:
        sys.exit(1)
    time.sleep(2)
    monitor_task(task_id)
    print("\n🐈 All tests finished!")

if __name__ == "__main__":
    main()
