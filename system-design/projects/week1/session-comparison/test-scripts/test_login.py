import requests

# 测试：登录
resp = requests.post('http://localhost:8080/login',
                     json={'username': 'alice', 'password': '123456'})

print("=== 测试 1.1: 登录响应 ===")
print(f"状态码: {resp.status_code}")
print(f"响应体: {resp.json()}")
print(f"\n=== 关键点：Cookie ===")
print(f"Cookies: {resp.cookies}")
print(f"是否有 session_id Cookie: {'session_id' in resp.cookies}")

if 'session_id' in resp.cookies:
    print(f"✅ 测试通过：Session ID = {resp.cookies['session_id']}")
else:
    print(f"❌ 测试失败：没有返回 session_id Cookie")
    print(f"\n💡 你需要做什么？")
    print(f"   在 loginHandler 中添加：")
    print(f"   c.SetCookie('session_id', sessionID, 3600, '/', '', false, true)")
