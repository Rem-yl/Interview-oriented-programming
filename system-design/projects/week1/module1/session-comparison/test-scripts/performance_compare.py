#!/usr/bin/env python3
"""
三种会话管理方案性能对比测试

测试指标:
1. 延迟 (Latency): P50, P95, P99, 平均延迟
2. 吞吐量 (Throughput): QPS (每秒请求数)
3. 并发性能: 不同并发数下的表现
4. 内存占用: 服务器内存使用情况

运行前确保:
1. 三种方案的服务器都已启动
   - Sticky Session: 8081, 8082, 8083
   - Redis Session: 8091, 8092, 8093
   - JWT Token: 8010, 8011, 8012

2. Redis 已启动 (如果测试 Redis Session)
   docker run -d --name redis -p 6379:6379 redis:alpine

运行方式:
    # 完整测试
    python performance_compare.py

    # 只测试延迟
    python performance_compare.py --test latency

    # 只测试吞吐量
    python performance_compare.py --test throughput

    # 指定测试方案
    python performance_compare.py --schemes sticky redis jwt

    # 自定义参数
    python performance_compare.py --requests 1000 --concurrency 100
"""

import requests
import time
import statistics
import argparse
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import Dict, List, Tuple
import json


# ==============================================================================
# 配置
# ==============================================================================

SCHEMES_CONFIG = {
    'sticky': {
        'name': 'Sticky Session',
        'login_url': 'http://localhost:8081/login',
        'profile_url': 'http://localhost:8081/profile',
        'auth_type': 'cookie',
        'cookie_name': 'session_id'
    },
    'redis': {
        'name': 'Redis Session',
        'login_url': 'http://localhost:8091/login',
        'profile_url': 'http://localhost:8091/profile',
        'auth_type': 'cookie',
        'cookie_name': 'sessionID'
    },
    'jwt': {
        'name': 'JWT Token',
        'login_url': 'http://localhost:8010/login',
        'profile_url': 'http://localhost:8010/profile',
        'auth_type': 'token',
        'header_name': 'Authorization'
    }
}


# ==============================================================================
# 辅助函数
# ==============================================================================

def setup_session(scheme_config: Dict) -> Tuple[requests.Session, Dict]:
    """
    设置测试会话（登录并获取认证凭据）

    Returns:
        (session, auth_data): session 对象和认证数据
    """
    session = requests.Session()

    # 登录
    try:
        resp = session.post(
            scheme_config['login_url'],
            json={'username': 'test_user', 'password': '123456'},
            timeout=5
        )

        if resp.status_code != 200:
            raise Exception(f"登录失败: {resp.status_code}")

        auth_data = {}

        if scheme_config['auth_type'] == 'cookie':
            # Cookie 认证
            cookie_name = scheme_config['cookie_name']
            if cookie_name in resp.cookies:
                auth_data['cookies'] = {cookie_name: resp.cookies[cookie_name]}
            else:
                raise Exception(f"未找到 Cookie: {cookie_name}")

        elif scheme_config['auth_type'] == 'token':
            # JWT Token 认证
            data = resp.json()
            if 'token' in data:
                auth_data['headers'] = {'Authorization': f"Bearer {data['token']}"}
            else:
                raise Exception("未找到 token")

        return session, auth_data

    except requests.exceptions.ConnectionError:
        print(f"❌ 无法连接到 {scheme_config['login_url']}")
        print(f"   请确保服务器已启动")
        return None, None
    except Exception as e:
        print(f"❌ 设置失败: {e}")
        return None, None


def check_server_availability(scheme_config: Dict) -> bool:
    """检查服务器是否可用"""
    try:
        resp = requests.get(scheme_config['profile_url'], timeout=2)
        return True
    except:
        return False


# ==============================================================================
# 测试 1: 延迟测试
# ==============================================================================

def test_latency(scheme_name: str, scheme_config: Dict, num_requests: int = 100) -> Dict:
    """
    测试延迟

    Args:
        scheme_name: 方案名称
        scheme_config: 方案配置
        num_requests: 请求次数

    Returns:
        延迟统计数据 (P50, P95, P99, 平均值, 最小值, 最大值)
    """
    print(f"\n{'='*70}")
    print(f"测试延迟: {scheme_config['name']}")
    print(f"{'='*70}")

    # 设置 Session
    session, auth_data = setup_session(scheme_config)
    if not session:
        return None

    print(f"发送 {num_requests} 个请求...")

    latencies = []
    success_count = 0
    error_count = 0

    for i in range(num_requests):
        try:
            start = time.time()

            if auth_data.get('cookies'):
                resp = session.get(
                    scheme_config['profile_url'],
                    cookies=auth_data['cookies'],
                    timeout=5
                )
            else:
                resp = session.get(
                    scheme_config['profile_url'],
                    headers=auth_data['headers'],
                    timeout=5
                )

            latency = (time.time() - start) * 1000  # 转换为毫秒

            if resp.status_code == 200:
                latencies.append(latency)
                success_count += 1
            else:
                error_count += 1

        except Exception as e:
            error_count += 1

        # 进度显示
        if (i + 1) % 20 == 0:
            print(f"  进度: {i+1}/{num_requests} ({(i+1)*100//num_requests}%)")

    if not latencies:
        print("❌ 所有请求都失败了")
        return None

    # 计算统计数据
    latencies.sort()

    results = {
        'scheme': scheme_config['name'],
        'total_requests': num_requests,
        'success_count': success_count,
        'error_count': error_count,
        'success_rate': success_count / num_requests * 100,
        'min': min(latencies),
        'max': max(latencies),
        'avg': statistics.mean(latencies),
        'p50': statistics.median(latencies),
        'p95': latencies[int(len(latencies) * 0.95)],
        'p99': latencies[int(len(latencies) * 0.99)]
    }

    # 打印结果
    print(f"\n结果:")
    print(f"  总请求数: {results['total_requests']}")
    print(f"  成功: {results['success_count']}, 失败: {results['error_count']}")
    print(f"  成功率: {results['success_rate']:.2f}%")
    print(f"\n  延迟统计:")
    print(f"    最小值:  {results['min']:.2f} ms")
    print(f"    平均值:  {results['avg']:.2f} ms")
    print(f"    P50:     {results['p50']:.2f} ms")
    print(f"    P95:     {results['p95']:.2f} ms")
    print(f"    P99:     {results['p99']:.2f} ms")
    print(f"    最大值:  {results['max']:.2f} ms")

    return results


# ==============================================================================
# 测试 2: 吞吐量测试
# ==============================================================================

def test_throughput(scheme_name: str, scheme_config: Dict,
                   duration: int = 10, concurrency: int = 50) -> Dict:
    """
    测试吞吐量 (QPS)

    Args:
        scheme_name: 方案名称
        scheme_config: 方案配置
        duration: 测试持续时间（秒）
        concurrency: 并发数

    Returns:
        吞吐量统计数据 (QPS, 总请求数, 成功/失败数)
    """
    print(f"\n{'='*70}")
    print(f"测试吞吐量: {scheme_config['name']}")
    print(f"{'='*70}")
    print(f"并发数: {concurrency}, 持续时间: {duration} 秒")

    # 设置 Session（每个线程一个）
    sessions = []
    auth_datas = []

    for i in range(concurrency):
        session, auth_data = setup_session(scheme_config)
        if not session:
            print(f"❌ 设置第 {i+1} 个会话失败")
            return None
        sessions.append(session)
        auth_datas.append(auth_data)

    print(f"✅ 设置了 {len(sessions)} 个会话")

    # 统计变量
    request_count = 0
    success_count = 0
    error_count = 0

    def make_request(worker_id: int):
        """单个请求"""
        nonlocal request_count, success_count, error_count

        session = sessions[worker_id % len(sessions)]
        auth_data = auth_datas[worker_id % len(auth_datas)]

        try:
            if auth_data.get('cookies'):
                resp = session.get(
                    scheme_config['profile_url'],
                    cookies=auth_data['cookies'],
                    timeout=5
                )
            else:
                resp = session.get(
                    scheme_config['profile_url'],
                    headers=auth_data['headers'],
                    timeout=5
                )

            request_count += 1

            if resp.status_code == 200:
                success_count += 1
            else:
                error_count += 1

        except Exception:
            request_count += 1
            error_count += 1

    # 执行测试
    start_time = time.time()
    end_time = start_time + duration

    print(f"\n开始压测...")

    with ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = []
        worker_id = 0
        last_print_time = 0

        while time.time() < end_time:
            # 提交任务
            future = executor.submit(make_request, worker_id)
            futures.append(future)
            worker_id += 1

            # 每秒显示一次进度
            current_time = time.time()
            elapsed = current_time - start_time

            if current_time - last_print_time >= 2.0 and elapsed > 0:
                current_qps = request_count / elapsed if elapsed > 0 else 0
                print(f"  进度: {int(elapsed)}/{duration} 秒, 当前 QPS: {current_qps:.0f}, 已提交: {len(futures)}")
                last_print_time = current_time

            # 控制提交速率，避免创建过多任务
            if len(futures) > concurrency * 100:  # 限制队列大小
                # 等待一些任务完成
                for future in as_completed(futures[:concurrency], timeout=1):
                    pass
                futures = futures[concurrency:]

        # 等待所有任务完成
        print(f"  等待所有任务完成...")
        for future in as_completed(futures, timeout=30):
            pass

    elapsed_time = time.time() - start_time
    qps = request_count / elapsed_time if elapsed_time > 0 else 0

    results = {
        'scheme': scheme_config['name'],
        'duration': elapsed_time,
        'concurrency': concurrency,
        'total_requests': request_count,
        'success_count': success_count,
        'error_count': error_count,
        'success_rate': success_count / request_count * 100 if request_count > 0 else 0,
        'qps': qps
    }

    # 打印结果
    print(f"\n结果:")
    print(f"  持续时间: {results['duration']:.2f} 秒")
    print(f"  总请求数: {results['total_requests']}")
    print(f"  成功: {results['success_count']}, 失败: {results['error_count']}")
    print(f"  成功率: {results['success_rate']:.2f}%")
    print(f"  QPS: {results['qps']:.0f} 请求/秒")

    return results


# ==============================================================================
# 测试 3: 并发扩展性测试
# ==============================================================================

def test_concurrency_scalability(scheme_name: str, scheme_config: Dict,
                                 concurrency_levels: List[int] = [10, 50, 100, 200]) -> Dict:
    """
    测试不同并发数下的性能

    Args:
        scheme_name: 方案名称
        scheme_config: 方案配置
        concurrency_levels: 并发数列表

    Returns:
        各并发级别的性能数据
    """
    print(f"\n{'='*70}")
    print(f"测试并发扩展性: {scheme_config['name']}")
    print(f"{'='*70}")

    results = []

    for concurrency in concurrency_levels:
        print(f"\n测试并发数: {concurrency}")
        result = test_throughput(scheme_name, scheme_config, duration=5, concurrency=concurrency)

        if result:
            results.append({
                'concurrency': concurrency,
                'qps': result['qps'],
                'success_rate': result['success_rate']
            })

    # 打印汇总
    print(f"\n{'='*70}")
    print(f"并发扩展性汇总: {scheme_config['name']}")
    print(f"{'='*70}")
    print(f"{'并发数':<12} {'QPS':<15} {'成功率':<15}")
    print("-" * 45)

    for r in results:
        print(f"{r['concurrency']:<12} {r['qps']:<15.0f} {r['success_rate']:<15.2f}%")

    return results


# ==============================================================================
# 对比测试
# ==============================================================================

def compare_all_schemes(test_type: str, schemes: List[str], **kwargs):
    """
    对比所有方案

    Args:
        test_type: 测试类型 ('latency', 'throughput', 'concurrency')
        schemes: 要测试的方案列表
        **kwargs: 传递给测试函数的参数
    """
    print(f"\n{'#'*70}")
    print(f"# 会话管理方案性能对比测试")
    print(f"# 测试类型: {test_type}")
    print(f"# 测试方案: {', '.join([SCHEMES_CONFIG[s]['name'] for s in schemes])}")
    print(f"{'#'*70}")

    results = {}

    for scheme in schemes:
        if scheme not in SCHEMES_CONFIG:
            print(f"⚠️  未知方案: {scheme}")
            continue

        config = SCHEMES_CONFIG[scheme]

        # 检查服务器可用性
        if not check_server_availability(config):
            print(f"\n⚠️  {config['name']} 服务器不可用，跳过测试")
            print(f"   请确保服务器已启动: {config['login_url']}")
            continue

        # 执行测试
        if test_type == 'latency':
            result = test_latency(scheme, config, **kwargs)
        elif test_type == 'throughput':
            result = test_throughput(scheme, config, **kwargs)
        elif test_type == 'concurrency':
            result = test_concurrency_scalability(scheme, config, **kwargs)
        else:
            print(f"未知测试类型: {test_type}")
            return

        if result:
            results[scheme] = result

    # 打印对比汇总
    print_comparison_summary(test_type, results)


def print_comparison_summary(test_type: str, results: Dict):
    """打印对比汇总表"""
    if not results:
        print("\n没有测试结果")
        return

    print(f"\n{'='*70}")
    print(f"对比汇总: {test_type.upper()}")
    print(f"{'='*70}")

    if test_type == 'latency':
        # 延迟对比表
        print(f"\n{'方案':<20} {'P50 (ms)':<12} {'P95 (ms)':<12} {'P99 (ms)':<12} {'平均 (ms)':<12}")
        print("-" * 70)

        for scheme, data in results.items():
            print(f"{data['scheme']:<20} {data['p50']:<12.2f} {data['p95']:<12.2f} "
                  f"{data['p99']:<12.2f} {data['avg']:<12.2f}")

    elif test_type == 'throughput':
        # 吞吐量对比表
        print(f"\n{'方案':<20} {'QPS':<15} {'成功率':<15} {'并发数':<10}")
        print("-" * 65)

        for scheme, data in results.items():
            print(f"{data['scheme']:<20} {data['qps']:<15.0f} "
                  f"{data['success_rate']:<15.2f}% {data['concurrency']:<10}")

    # 推荐
    print(f"\n推荐:")

    if test_type == 'latency':
        # 找出 P50 最低的
        best = min(results.items(), key=lambda x: x[1]['p50'])
        print(f"  延迟最低: {best[1]['scheme']} (P50: {best[1]['p50']:.2f} ms)")

    elif test_type == 'throughput':
        # 找出 QPS 最高的
        best = max(results.items(), key=lambda x: x[1]['qps'])
        print(f"  吞吐量最高: {best[1]['scheme']} (QPS: {best[1]['qps']:.0f})")


# ==============================================================================
# 主函数
# ==============================================================================

def main():
    parser = argparse.ArgumentParser(description='会话管理方案性能对比测试')

    parser.add_argument(
        '--test',
        choices=['latency', 'throughput', 'concurrency', 'all'],
        default='all',
        help='测试类型 (默认: all)'
    )

    parser.add_argument(
        '--schemes',
        nargs='+',
        choices=['sticky', 'redis', 'jwt'],
        default=['sticky', 'redis', 'jwt'],
        help='要测试的方案 (默认: 所有)'
    )

    parser.add_argument(
        '--requests',
        type=int,
        default=100,
        help='延迟测试的请求数 (默认: 100)'
    )

    parser.add_argument(
        '--duration',
        type=int,
        default=10,
        help='吞吐量测试的持续时间（秒）(默认: 10)'
    )

    parser.add_argument(
        '--concurrency',
        type=int,
        default=50,
        help='吞吐量测试的并发数 (默认: 50)'
    )

    args = parser.parse_args()

    # 执行测试
    if args.test in ['latency', 'all']:
        compare_all_schemes('latency', args.schemes, num_requests=args.requests)

    if args.test in ['throughput', 'all']:
        compare_all_schemes('throughput', args.schemes,
                          duration=args.duration, concurrency=args.concurrency)

    if args.test == 'concurrency':
        compare_all_schemes('concurrency', args.schemes)

    print(f"\n{'='*70}")
    print("测试完成！")
    print(f"{'='*70}\n")


if __name__ == '__main__':
    main()
