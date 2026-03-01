#!/usr/bin/env python3

"""
Comprehensive Benchmark Test Suite for Workflow Generator

This script tests the workflow generator with increasingly complex scenarios
to evaluate quality, schema compliance, and handling of edge cases.

Usage:
    python3 benchmark_test.py              # Run all benchmark tests
    python3 benchmark_test.py --category complexity    # Run specific category
    python3 benchmark_test.py --quick      # Run quick subset
"""

import json
import requests
import time
import argparse
from typing import Dict, List, Any, Tuple
from datetime import datetime
import os

# Configuration
API_BASE_URL = "http://localhost:3001"
USE_DEV_ENDPOINTS = True
DEV_USER_ID = "dev-test-user"

# Available models
AVAILABLE_MODELS = [
    {"id": "Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8-TEE", "name": "Qwen3 Coder 480B (Recommended)"},
    {"id": "Qwen/Qwen3-235B-A22B", "name": "Qwen3 235B (Advanced reasoning)"},
    {"id": "deepseek-ai/DeepSeek-V3.2-TEE", "name": "DeepSeek V3.2"},
    {"id": "zai-org/GLM-4.7-TEE", "name": "GLM 4.7"},
    {"id": "Qwen/Qwen3-VL-235B-A22B-Instruct", "name": "Qwen3 VL 235B"},
    {"id": "MiniMaxAI/MiniMax-M2.1-TEE", "name": "MiniMax M2.1"},
]

SELECTED_MODEL = None  # Will be set by CLI arg or auto

# ANSI color codes
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    MAGENTA = '\033[0;35m'
    NC = '\033[0m'

# Benchmark Test Categories
BENCHMARK_TESTS = {
    "simple": {
        "name": "Simple Workflows (1-2 blocks)",
        "tests": [
            {
                "name": "Single Search",
                "prompt": "Create a workflow that searches for AI news",
                "expected_blocks": 2,
                "expected_structured": 1,
                "complexity": 1
            },
            {
                "name": "Text Generation",
                "prompt": "Write a blog post about renewable energy",
                "expected_blocks": 2,
                "expected_structured": 0,
                "complexity": 1
            }
        ]
    },
    "moderate": {
        "name": "Moderate Workflows (3-5 blocks)",
        "tests": [
            {
                "name": "Search and Summarize",
                "prompt": "Search for AI news, extract the top 3 articles with titles and URLs, then write a summary",
                "expected_blocks": 4,
                "expected_structured": 1,
                "complexity": 2
            },
            {
                "name": "Data Extraction Pipeline",
                "prompt": "Search for Python tutorials, extract titles, URLs, and difficulty levels, then filter only beginner tutorials",
                "expected_blocks": 4,
                "expected_structured": 2,
                "complexity": 2
            },
            {
                "name": "Multi-Tool Workflow",
                "prompt": "Get current time, search for today's tech news, and extract article summaries with timestamps",
                "expected_blocks": 4,
                "expected_structured": 1,
                "complexity": 2
            }
        ]
    },
    "complex": {
        "name": "Complex Workflows (6-10 blocks)",
        "tests": [
            {
                "name": "Research & Report Generator",
                "prompt": "Create a comprehensive research workflow: 1) Search for AI breakthroughs in 2026, 2) Extract key findings with sources, 3) Search for expert opinions, 4) Combine all data into structured format, 5) Generate executive summary, 6) Create detailed report",
                "expected_blocks": 8,
                "expected_structured": 3,
                "complexity": 3
            },
            {
                "name": "Multi-Source Aggregator",
                "prompt": "Build a news aggregator that: searches GitHub for trending repos, searches web for tech news, searches for AI research papers, extracts structured data from each source, combines all results, ranks by relevance, and generates a daily digest",
                "expected_blocks": 10,
                "expected_structured": 4,
                "complexity": 3
            },
            {
                "name": "Conditional Branching",
                "prompt": "Create a weather alert system: get current weather, check if temperature is above 30C OR if it's raining, if hot then search for heat safety tips, if raining then search for indoor activities, combine results and send summary",
                "expected_blocks": 7,
                "expected_structured": 2,
                "complexity": 3
            }
        ]
    },
    "edge_cases": {
        "name": "Edge Cases & Schema Challenges",
        "tests": [
            {
                "name": "Array Root Schema",
                "prompt": "Return a simple array of product IDs directly, not wrapped in an object",
                "expected_blocks": 2,
                "expected_structured": 1,
                "complexity": 2,
                "expects_array_schema": True
            },
            {
                "name": "Nested Arrays",
                "prompt": "Search for restaurants and return a list where each restaurant has an array of reviews, each review has an array of tags",
                "expected_blocks": 3,
                "expected_structured": 1,
                "complexity": 2
            },
            {
                "name": "Mixed Data Types",
                "prompt": "Create a workflow that returns structured data with strings, numbers, booleans, arrays, and nested objects for a product catalog",
                "expected_blocks": 3,
                "expected_structured": 1,
                "complexity": 2
            },
            {
                "name": "Large Schema",
                "prompt": "Extract detailed user profiles with 15+ fields including: name, email, age, address (street, city, state, zip), preferences (array), social media links (object), purchase history (array of objects with date, amount, items), subscription status (boolean), and account metadata",
                "expected_blocks": 3,
                "expected_structured": 1,
                "complexity": 2
            },
            {
                "name": "Multiple Structured Outputs",
                "prompt": "Search for tech companies, extract structured company data (name, founded, employees), then for each company search for recent news and extract structured article data (title, date, summary), then aggregate everything",
                "expected_blocks": 5,
                "expected_structured": 3,
                "complexity": 3
            }
        ]
    },
    "performance": {
        "name": "Performance & Scale Tests",
        "tests": [
            {
                "name": "10-Block Chain",
                "prompt": "Create a 10-step workflow: 1) search topic A, 2) extract data from A, 3) search topic B, 4) extract data from B, 5) search topic C, 6) extract data from C, 7) merge A+B, 8) merge result with C, 9) analyze trends, 10) generate final report",
                "expected_blocks": 12,
                "expected_structured": 5,
                "complexity": 4
            },
            {
                "name": "Parallel Processing",
                "prompt": "Create parallel workflows that simultaneously: search news, search GitHub, search YouTube, search Twitter, search Reddit, then combine all results",
                "expected_blocks": 7,
                "expected_structured": 5,
                "complexity": 3
            }
        ]
    },
    "real_world": {
        "name": "Real-World Use Cases",
        "tests": [
            {
                "name": "Content Curation Bot",
                "prompt": "Build a content curation bot that searches for AI/ML articles, filters by quality indicators (source credibility, engagement), extracts key points, generates summaries, categorizes by topic, and creates a newsletter format",
                "expected_blocks": 8,
                "expected_structured": 3,
                "complexity": 3
            },
            {
                "name": "Market Research Tool",
                "prompt": "Create a market research workflow: search for competitor products, extract pricing and features, search for customer reviews, analyze sentiment, compare with our product, identify gaps, generate competitive analysis report",
                "expected_blocks": 9,
                "expected_structured": 4,
                "complexity": 3
            },
            {
                "name": "Automated Report Generator",
                "prompt": "Build an automated weekly report: get current date, search for industry news from past 7 days, extract key events with dates, search for market data, calculate trends, generate charts description, write executive summary, compile full report with sections",
                "expected_blocks": 10,
                "expected_structured": 3,
                "complexity": 3
            },
            {
                "name": "E-commerce Analytics",
                "prompt": "Create product analytics workflow: search for product reviews, extract ratings and feedback, calculate average rating, identify common complaints, search for competitor products, compare features and prices, generate improvement recommendations",
                "expected_blocks": 9,
                "expected_structured": 4,
                "complexity": 3
            }
        ]
    }
}

def print_header():
    print(f"{Colors.MAGENTA}╔════════════════════════════════════════════════════════════════╗{Colors.NC}")
    print(f"{Colors.MAGENTA}║{Colors.NC}  🏁 ClaraVerse Workflow Generator Benchmark Suite          {Colors.MAGENTA}║{Colors.NC}")
    print(f"{Colors.MAGENTA}╚════════════════════════════════════════════════════════════════╝{Colors.NC}")
    print()

def print_section(title: str):
    print(f"\n{Colors.YELLOW}▶ {title}{Colors.NC}")
    print("━" * 64)

def select_model() -> str:
    """Interactive model selection."""
    global SELECTED_MODEL

    print(f"\n{Colors.YELLOW}▶ Model Selection{Colors.NC}")
    print("━" * 64)
    print(f"{Colors.CYAN}Available models:{Colors.NC}\n")

    for i, model in enumerate(AVAILABLE_MODELS, 1):
        print(f"  {Colors.YELLOW}{i}.{Colors.NC} {model['name']}")
        print(f"     {Colors.BLUE}{model['id']}{Colors.NC}")
        print()

    print(f"  {Colors.YELLOW}0.{Colors.NC} Auto (let the generator choose)")
    print()

    while True:
        try:
            choice = input(f"{Colors.CYAN}Select a model (0-{len(AVAILABLE_MODELS)}): {Colors.NC}").strip()
            choice_num = int(choice)

            if choice_num == 0:
                SELECTED_MODEL = None
                print(f"{Colors.GREEN}✅ Using auto model selection{Colors.NC}")
                return None
            elif 1 <= choice_num <= len(AVAILABLE_MODELS):
                SELECTED_MODEL = AVAILABLE_MODELS[choice_num - 1]['id']
                print(f"{Colors.GREEN}✅ Selected: {AVAILABLE_MODELS[choice_num - 1]['name']}{Colors.NC}")
                return SELECTED_MODEL
            else:
                print(f"{Colors.RED}Invalid choice. Please enter 0-{len(AVAILABLE_MODELS)}{Colors.NC}")
        except ValueError:
            print(f"{Colors.RED}Invalid input. Please enter a number{Colors.NC}")
        except KeyboardInterrupt:
            print(f"\n{Colors.YELLOW}Using auto selection{Colors.NC}")
            return None

def check_backend() -> bool:
    """Check if backend is reachable."""
    try:
        response = requests.get(f"{API_BASE_URL}/health", timeout=5)
        return response.status_code == 200
    except:
        return False

def create_test_agent(name: str) -> str:
    """Create a test agent and return its ID."""
    try:
        response = requests.post(
            f"{API_BASE_URL}/api/dev-test/agents" if USE_DEV_ENDPOINTS else f"{API_BASE_URL}/api/agents",
            headers={"Content-Type": "application/json"},
            json={"name": name, "description": "Benchmark test agent"},
            timeout=10
        )
        if response.status_code in [200, 201]:
            return response.json().get('id')
    except:
        pass
    return None

def generate_workflow(agent_id: str, prompt: str) -> Dict:
    """Generate a workflow for the given prompt."""
    try:
        # Build request body
        request_body = {
            "user_message": prompt,
            "conversation_id": f"benchmark-{int(time.time())}"
        }

        # Add model_id if one was selected
        if SELECTED_MODEL:
            request_body["model_id"] = SELECTED_MODEL

        response = requests.post(
            f"{API_BASE_URL}/api/dev-test/agents/{agent_id}/generate-workflow" if USE_DEV_ENDPOINTS else f"{API_BASE_URL}/api/agents/{agent_id}/generate-workflow",
            headers={"Content-Type": "application/json"},
            json=request_body,
            timeout=120  # Allow up to 2 minutes for complex workflows
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {"error": f"Status {response.status_code}: {response.text}"}
    except Exception as e:
        return {"error": str(e)}

def cleanup_agent(agent_id: str):
    """Delete the test agent."""
    try:
        if USE_DEV_ENDPOINTS:
            requests.delete(f"{API_BASE_URL}/api/dev-test/agents/{agent_id}", timeout=5)
        else:
            requests.delete(f"{API_BASE_URL}/api/agents/{agent_id}", timeout=5)
    except:
        pass

def analyze_workflow_quality(workflow_data: Dict, test_config: Dict) -> Dict:
    """Comprehensive quality analysis with scoring."""

    if not workflow_data or 'workflow' not in workflow_data:
        return {
            "score": 0,
            "errors": ["Invalid workflow data"],
            "metrics": {}
        }

    workflow = workflow_data['workflow']
    blocks = workflow.get('blocks', [])

    # Basic metrics
    metrics = {
        "total_blocks": len(blocks),
        "llm_blocks": len([b for b in blocks if b.get('type') == 'llm_inference']),
        "variable_blocks": len([b for b in blocks if b.get('type') == 'variable']),
        "structured_blocks": len([b for b in blocks if b.get('type') == 'llm_inference' and b.get('config', {}).get('outputFormat') == 'json']),
        "connections": len(workflow.get('connections', [])),
    }

    # Schema analysis
    schema_issues = []
    complete_required_arrays = True
    missing_additional_props = 0
    has_array_schema = False

    for i, block in enumerate(blocks):
        if block.get('type') == 'llm_inference':
            config = block.get('config', {})
            if config.get('outputFormat') == 'json':
                schema = config.get('outputSchema', {})
                schema_type = schema.get('type', 'object')

                if schema_type == 'array':
                    has_array_schema = True
                    items = schema.get('items', {})
                    if items.get('type') == 'object':
                        props = items.get('properties', {})
                        required = items.get('required', [])
                        for prop in props.keys():
                            if prop not in required:
                                complete_required_arrays = False
                                schema_issues.append(f"Block {i}: Array item property '{prop}' not in required")

                elif schema_type == 'object':
                    props = schema.get('properties', {})
                    required = schema.get('required', [])
                    for prop in props.keys():
                        if prop not in required:
                            complete_required_arrays = False
                            schema_issues.append(f"Block {i}: Property '{prop}' not in required")

                    if schema.get('additionalProperties') != False:
                        missing_additional_props += 1

    # Calculate score
    score = 100

    # Block count accuracy (20 points)
    expected_blocks = test_config.get('expected_blocks', 0)
    if expected_blocks > 0:
        block_diff = abs(metrics['total_blocks'] - expected_blocks)
        if block_diff == 0:
            score += 0  # Perfect
        elif block_diff <= 2:
            score -= 5
        else:
            score -= 15

    # Structured output usage (20 points)
    expected_structured = test_config.get('expected_structured', 0)
    if expected_structured > 0:
        structured_diff = abs(metrics['structured_blocks'] - expected_structured)
        if structured_diff == 0:
            score += 0
        elif structured_diff <= 1:
            score -= 5
        else:
            score -= 15

    # Schema compliance (40 points)
    if not complete_required_arrays:
        score -= 20

    if missing_additional_props > 0:
        score -= min(20, missing_additional_props * 5)

    # Array schema handling (20 points bonus for edge cases)
    if test_config.get('expects_array_schema', False):
        if has_array_schema:
            score += 10
        else:
            score -= 20
            schema_issues.append("Expected array schema at root level but got object")

    score = max(0, min(100, score))

    return {
        "score": score,
        "metrics": metrics,
        "schema_issues": schema_issues,
        "complete_required_arrays": complete_required_arrays,
        "missing_additional_props": missing_additional_props,
        "has_array_schema": has_array_schema
    }

def run_benchmark_test(category: str, test: Dict, test_num: int, total: int) -> Tuple[Dict, float]:
    """Run a single benchmark test and return results."""

    print(f"\n{Colors.CYAN}{'═' * 64}{Colors.NC}")
    print(f"{Colors.CYAN}Test {test_num}/{total}: {test['name']}{Colors.NC}")
    print(f"{Colors.CYAN}Complexity: {'⭐' * test['complexity']}{Colors.NC}")
    print(f"{Colors.CYAN}{'═' * 64}{Colors.NC}")
    print(f"\n{Colors.BLUE}Prompt:{Colors.NC} {test['prompt'][:100]}...")

    # Create agent
    agent_id = create_test_agent(f"Benchmark {test['name']}")
    if not agent_id:
        return {"error": "Failed to create agent", "score": 0}, 0.0

    # Generate workflow (timed)
    start_time = time.time()
    workflow_data = generate_workflow(agent_id, test['prompt'])
    generation_time = time.time() - start_time

    # Clean up
    cleanup_agent(agent_id)

    # Check for errors
    if 'error' in workflow_data:
        print(f"{Colors.RED}❌ Generation failed: {workflow_data['error'][:100]}{Colors.NC}")
        return {"error": workflow_data['error'], "score": 0}, generation_time

    # Analyze quality
    analysis = analyze_workflow_quality(workflow_data, test)

    # Print results
    print(f"\n{Colors.YELLOW}Results:{Colors.NC}")
    print(f"  ⏱️  Generation time: {generation_time:.2f}s")
    print(f"  📊 Blocks: {analysis['metrics']['total_blocks']} (expected: {test.get('expected_blocks', '?')})")
    print(f"  🔧 Structured outputs: {analysis['metrics']['structured_blocks']} (expected: {test.get('expected_structured', '?')})")

    if analysis['schema_issues']:
        print(f"  {Colors.YELLOW}⚠️  Schema issues:{Colors.NC}")
        for issue in analysis['schema_issues'][:3]:
            print(f"     • {issue}")
        if len(analysis['schema_issues']) > 3:
            print(f"     • ... and {len(analysis['schema_issues']) - 3} more")

    score = analysis['score']
    if score >= 90:
        print(f"  {Colors.GREEN}🏆 Score: {score}/100 (Excellent){Colors.NC}")
    elif score >= 70:
        print(f"  {Colors.YELLOW}⭐ Score: {score}/100 (Good){Colors.NC}")
    else:
        print(f"  {Colors.RED}⚠️  Score: {score}/100 (Needs Improvement){Colors.NC}")

    return analysis, generation_time

def run_benchmark_suite(categories: List[str] = None, quick: bool = False):
    """Run the full benchmark suite."""

    print_header()

    if not check_backend():
        print(f"{Colors.RED}❌ Backend not reachable at {API_BASE_URL}{Colors.NC}")
        return

    print(f"{Colors.GREEN}✅ Backend is running{Colors.NC}")

    # Let user select model (if not already set via CLI)
    if SELECTED_MODEL is None:
        select_model()

    # Select categories
    if categories:
        test_categories = {k: v for k, v in BENCHMARK_TESTS.items() if k in categories}
    else:
        test_categories = BENCHMARK_TESTS

    # Quick mode: only run first test from each category
    if quick:
        print(f"{Colors.YELLOW}🚀 Quick mode: Running first test from each category{Colors.NC}")
        for cat in test_categories.values():
            cat['tests'] = cat['tests'][:1]

    # Count total tests
    total_tests = sum(len(cat['tests']) for cat in test_categories.values())

    print(f"\n{Colors.BLUE}📋 Running {total_tests} benchmark tests across {len(test_categories)} categories{Colors.NC}")

    # Results tracking
    all_results = {}
    test_counter = 0

    # Create output directory
    os.makedirs("benchmark-results", exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    # Run tests
    for cat_key, category in test_categories.items():
        print(f"\n{Colors.MAGENTA}{'═' * 64}{Colors.NC}")
        print(f"{Colors.MAGENTA}Category: {category['name']}{Colors.NC}")
        print(f"{Colors.MAGENTA}{'═' * 64}{Colors.NC}")

        cat_results = []

        for test in category['tests']:
            test_counter += 1
            result, gen_time = run_benchmark_test(cat_key, test, test_counter, total_tests)

            cat_results.append({
                "name": test['name'],
                "prompt": test['prompt'],
                "complexity": test['complexity'],
                "generation_time": gen_time,
                "score": result.get('score', 0),
                "metrics": result.get('metrics', {}),
                "issues": result.get('schema_issues', [])
            })

            time.sleep(2)  # Avoid rate limiting

        all_results[cat_key] = {
            "category_name": category['name'],
            "results": cat_results
        }

    # Generate summary
    print(f"\n{Colors.MAGENTA}{'═' * 64}{Colors.NC}")
    print(f"{Colors.MAGENTA}Benchmark Summary{Colors.NC}")
    print(f"{Colors.MAGENTA}{'═' * 64}{Colors.NC}\n")

    total_score = 0
    total_time = 0
    test_count = 0

    for cat_key, cat_data in all_results.items():
        cat_scores = [r['score'] for r in cat_data['results'] if 'score' in r]
        cat_times = [r['generation_time'] for r in cat_data['results'] if 'generation_time' in r]

        if cat_scores:
            avg_score = sum(cat_scores) / len(cat_scores)
            avg_time = sum(cat_times) / len(cat_times)

            print(f"{Colors.CYAN}{cat_data['category_name']}:{Colors.NC}")
            print(f"  Tests: {len(cat_data['results'])}")
            print(f"  Avg Score: {avg_score:.1f}/100")
            print(f"  Avg Time: {avg_time:.2f}s")
            print()

            total_score += sum(cat_scores)
            total_time += sum(cat_times)
            test_count += len(cat_scores)

    if test_count > 0:
        overall_avg = total_score / test_count
        overall_time = total_time / test_count

        print(f"{Colors.BLUE}Overall Results:{Colors.NC}")
        print(f"  Total tests: {test_count}")
        print(f"  Average score: {overall_avg:.1f}/100")
        print(f"  Average generation time: {overall_time:.2f}s")
        print()

        if overall_avg >= 90:
            print(f"{Colors.GREEN}🏆 Overall Grade: EXCELLENT{Colors.NC}")
            print("   Workflow generator is production-ready!")
        elif overall_avg >= 75:
            print(f"{Colors.YELLOW}⭐ Overall Grade: GOOD{Colors.NC}")
            print("   Workflow generator performs well with minor issues.")
        elif overall_avg >= 60:
            print(f"{Colors.YELLOW}📊 Overall Grade: FAIR{Colors.NC}")
            print("   Workflow generator needs improvements in some areas.")
        else:
            print(f"{Colors.RED}⚠️  Overall Grade: NEEDS WORK{Colors.NC}")
            print("   Significant improvements needed.")

    # Save results
    output_file = f"benchmark-results/benchmark_{timestamp}.json"
    with open(output_file, 'w') as f:
        json.dump({
            "timestamp": timestamp,
            "total_tests": test_count,
            "average_score": overall_avg if test_count > 0 else 0,
            "average_time": overall_time if test_count > 0 else 0,
            "categories": all_results
        }, f, indent=2)

    print(f"\n{Colors.GREEN}✅ Detailed results saved to: {output_file}{Colors.NC}")

def main():
    global SELECTED_MODEL

    parser = argparse.ArgumentParser(description="Benchmark the workflow generator")
    parser.add_argument("--category", type=str, help="Run specific category (simple, moderate, complex, edge_cases, performance, real_world)")
    parser.add_argument("--quick", action="store_true", help="Quick mode: Run only first test from each category")
    parser.add_argument("--model", type=str, help="Model ID to use (e.g., Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8-TEE)")

    args = parser.parse_args()

    # Set model if provided
    if args.model:
        SELECTED_MODEL = args.model
        print(f"{Colors.CYAN}Using model: {args.model}{Colors.NC}\n")

    categories = [args.category] if args.category else None
    run_benchmark_suite(categories=categories, quick=args.quick)

if __name__ == "__main__":
    main()
