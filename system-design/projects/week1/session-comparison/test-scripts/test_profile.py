import requests

# Step 1: 登录
session = requests.Session()  # Session 对象会自动管理 Cookie
resp = session.post('http://localhost:8080/login',
                    json={'username': 'alice', 'password': '123456'})

print("=== Step 1: 登录 ===")
print(f"状态码: {resp.status_code}")
print(f"Cookies: {session.cookies}")

# Step 2: 访问 /profile
print("\n=== Step 2: 访问 /profile ===")
try:
    resp = session.get('http://localhost:8080/profile')
    print(f"状态码: {resp.status_code}")
    print(f"响应体: {resp.json()}")

    if resp.status_code == 200:
        print("✅ 测试通过：能够获取用户信息")
    else:
        print("❌ 测试失败：无法获取用户信息")
except Exception as e:
    print(f"❌ 测试失败：接口不存在 - {e}")
    print("\n💡 你需要做什么？")
    print("   添加 GET /profile 接口")
