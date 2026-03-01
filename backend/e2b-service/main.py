#!/usr/bin/env python3
"""
E2B Code Executor Microservice for ClaraVerse
Provides REST API for executing Python code in E2B sandboxes
"""

import os
import base64
from typing import List, Optional, Dict, Any
from fastapi import FastAPI, HTTPException, UploadFile, File, Form, Request
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from e2b_code_interpreter import Sandbox
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# E2B API key: optional at startup, can be provided per-request via X-E2B-API-Key header
E2B_API_KEY = os.getenv("E2B_API_KEY", "")

app = FastAPI(
    title="E2B Code Executor Service",
    description="Microservice for executing Python code in isolated E2B sandboxes",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, restrict to backend service
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Request/Response Models
class ExecuteRequest(BaseModel):
    code: str
    timeout: Optional[int] = 30  # seconds


class PlotResult(BaseModel):
    format: str  # "png", "svg", etc.
    data: str  # base64 encoded


class ExecuteResponse(BaseModel):
    success: bool
    stdout: str
    stderr: str
    error: Optional[str] = None
    plots: List[PlotResult] = []
    execution_time: Optional[float] = None


class FileUploadRequest(BaseModel):
    code: str
    timeout: Optional[int] = 30


# Advanced execution models (with dependencies and output files)
class AdvancedExecuteRequest(BaseModel):
    code: str
    timeout: Optional[int] = 30
    dependencies: List[str] = []  # pip packages to install
    output_files: List[str] = []  # files to retrieve after execution


class FileResult(BaseModel):
    filename: str
    data: str  # base64 encoded
    size: int


class AdvancedExecuteResponse(BaseModel):
    success: bool
    stdout: str
    stderr: str
    error: Optional[str] = None
    plots: List[PlotResult] = []
    files: List[FileResult] = []
    execution_time: Optional[float] = None
    install_output: str = ""


def get_api_key(request: Request) -> str:
    """Get E2B API key from request header or fall back to env var."""
    key = request.headers.get("X-E2B-API-Key", "").strip()
    if key:
        return key
    if E2B_API_KEY:
        return E2B_API_KEY
    return ""


def create_sandbox(api_key: str):
    """Create an E2B sandbox with the given API key."""
    if not api_key:
        raise HTTPException(
            status_code=400,
            detail="E2B API key not configured. Set it in Admin > Code Execution or via E2B_API_KEY env var."
        )
    return Sandbox.create(api_key=api_key)


# Health check endpoint
@app.get("/health")
async def health_check(request: Request):
    """Health check endpoint"""
    key = get_api_key(request)
    return {
        "status": "healthy",
        "service": "e2b-executor",
        "e2b_api_key_set": bool(key)
    }


# Execute Python code endpoint
@app.post("/execute", response_model=ExecuteResponse)
async def execute_code(request: ExecuteRequest, raw_request: Request):
    """
    Execute Python code in an E2B sandbox

    Returns:
        - stdout: Standard output
        - stderr: Standard error
        - plots: List of generated plots (base64 encoded)
        - error: Error message if execution failed
    """
    logger.info(f"Executing code (length: {len(request.code)} chars)")
    api_key = get_api_key(raw_request)

    try:
        # Create sandbox
        with create_sandbox(api_key) as sandbox:
            # Run code
            execution = sandbox.run_code(request.code)

            # Collect stdout
            stdout = ""
            if execution.logs.stdout:
                stdout = "\n".join(execution.logs.stdout)

            # Collect stderr
            stderr = ""
            if execution.logs.stderr:
                stderr = "\n".join(execution.logs.stderr)

            # Check for execution errors
            error_msg = None
            if execution.error:
                error_msg = str(execution.error)
                logger.warning(f"Execution error: {error_msg}")

            # Collect plots and text results
            plots = []
            result_texts = []
            for i, result in enumerate(execution.results):
                if hasattr(result, 'png') and result.png:
                    plots.append(PlotResult(
                        format="png",
                        data=result.png  # Already base64 encoded
                    ))
                    logger.info(f"Found plot {i}: {len(result.png)} bytes (base64)")
                elif hasattr(result, 'text') and result.text:
                    result_texts.append(result.text)
                    logger.info(f"Found text result {i}: {result.text[:100]}...")

            # Append result texts to stdout (captures last expression like Jupyter)
            if result_texts:
                result_output = "\n".join(result_texts)
                if stdout:
                    stdout = stdout + "\n" + result_output
                else:
                    stdout = result_output

            response = ExecuteResponse(
                success=error_msg is None,
                stdout=stdout,
                stderr=stderr,
                error=error_msg,
                plots=plots
            )

            logger.info(f"Execution completed: success={response.success}, plots={len(plots)}")
            return response

    except Exception as e:
        logger.error(f"Sandbox execution failed: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Sandbox execution failed: {str(e)}"
        )


# Execute with file upload endpoint
@app.post("/execute-with-files", response_model=ExecuteResponse)
async def execute_with_files(
    raw_request: Request,
    code: str = Form(...),
    files: List[UploadFile] = File(...),
    timeout: int = Form(30)
):
    """
    Execute Python code with uploaded files

    Files are uploaded to the sandbox and can be accessed by filename in the code
    """
    logger.info(f"Executing code with {len(files)} files")
    api_key = get_api_key(raw_request)

    try:
        # Create sandbox
        with create_sandbox(api_key) as sandbox:
            # Upload files to sandbox
            for file in files:
                content = await file.read()
                sandbox.files.write(file.filename, content)
                logger.info(f"Uploaded file: {file.filename} ({len(content)} bytes)")

            # Run code
            execution = sandbox.run_code(code)

            # Collect stdout
            stdout = ""
            if execution.logs.stdout:
                stdout = "\n".join(execution.logs.stdout)

            # Collect stderr
            stderr = ""
            if execution.logs.stderr:
                stderr = "\n".join(execution.logs.stderr)

            # Check for errors
            error_msg = None
            if execution.error:
                error_msg = str(execution.error)
                logger.warning(f"Execution error: {error_msg}")

            # Collect plots
            plots = []
            for i, result in enumerate(execution.results):
                if hasattr(result, 'png') and result.png:
                    plots.append(PlotResult(
                        format="png",
                        data=result.png
                    ))
                    logger.info(f"Found plot {i}")

            response = ExecuteResponse(
                success=error_msg is None,
                stdout=stdout,
                stderr=stderr,
                error=error_msg,
                plots=plots
            )

            logger.info(f"Execution with files completed: success={response.success}")
            return response

    except Exception as e:
        logger.error(f"Sandbox execution failed: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Sandbox execution failed: {str(e)}"
        )


# Execute with dependencies and output file retrieval
@app.post("/execute-advanced", response_model=AdvancedExecuteResponse)
async def execute_advanced(request: AdvancedExecuteRequest, raw_request: Request):
    """
    Execute Python code with pip dependencies and output file retrieval.

    - Install pip packages before running code
    - Run user code (max 30 seconds)
    - Auto-detect and retrieve ALL generated files (plus any explicitly specified)
    """
    import time

    logger.info(f"Advanced execution: code={len(request.code)} chars, deps={request.dependencies}, output_files={request.output_files}")
    api_key = get_api_key(raw_request)

    try:
        with create_sandbox(api_key) as sandbox:
            start_time = time.time()
            install_output = ""

            # 1. Install dependencies (if any)
            if request.dependencies:
                deps_str = " ".join(request.dependencies)
                logger.info(f"Installing dependencies: {deps_str}")
                try:
                    result = sandbox.commands.run(f"pip install -q {deps_str}", timeout=60)
                    install_output = (result.stdout or "") + (result.stderr or "")
                    logger.info(f"Dependencies installed: {install_output[:200]}")
                except Exception as e:
                    logger.error(f"Dependency installation failed: {e}")
                    return AdvancedExecuteResponse(
                        success=False,
                        stdout="",
                        stderr="",
                        error=f"Failed to install dependencies: {str(e)}",
                        plots=[],
                        files=[],
                        execution_time=time.time() - start_time,
                        install_output=str(e)
                    )

            # 2. List files BEFORE execution to detect new files later
            files_before = set()
            try:
                result = sandbox.commands.run("find /home/user -maxdepth 2 -type f 2>/dev/null || ls -la /home/user", timeout=10)
                if result.stdout:
                    for line in result.stdout.strip().split('\n'):
                        line = line.strip()
                        if line and not line.startswith('total'):
                            # Handle both find output (full paths) and ls output
                            if line.startswith('/'):
                                files_before.add(line)
                            else:
                                # ls -la format: permissions links owner group size date name
                                parts = line.split()
                                if len(parts) >= 9:
                                    files_before.add(parts[-1])
                logger.info(f"Files before execution: {len(files_before)}")
            except Exception as e:
                logger.warning(f"Could not list files before execution: {e}")

            # 3. Run user code
            execution = sandbox.run_code(request.code)

            # Collect stdout
            stdout = ""
            if execution.logs.stdout:
                stdout = "\n".join(execution.logs.stdout)

            # Collect stderr
            stderr = ""
            if execution.logs.stderr:
                stderr = "\n".join(execution.logs.stderr)

            # Check for errors
            error_msg = None
            if execution.error:
                error_msg = str(execution.error)
                logger.warning(f"Execution error: {error_msg}")

            # Collect plots and text results from execution.results
            # E2B results contain last expression value (like Jupyter)
            plots = []
            result_texts = []
            for i, result in enumerate(execution.results):
                if hasattr(result, 'png') and result.png:
                    plots.append(PlotResult(
                        format="png",
                        data=result.png
                    ))
                    logger.info(f"Found plot {i}")
                elif hasattr(result, 'text') and result.text:
                    # Capture text output from last expression (like Jupyter Out[])
                    result_texts.append(result.text)
                    logger.info(f"Found text result {i}: {result.text[:100]}...")

            # Append result texts to stdout if no explicit print was used
            if result_texts:
                result_output = "\n".join(result_texts)
                if stdout:
                    stdout = stdout + "\n" + result_output
                else:
                    stdout = result_output

            # 4. List files AFTER execution to detect new files
            files_after = set()
            new_files = []
            try:
                result = sandbox.commands.run("find /home/user -maxdepth 2 -type f 2>/dev/null || ls -la /home/user", timeout=10)
                if result.stdout:
                    for line in result.stdout.strip().split('\n'):
                        line = line.strip()
                        if line and not line.startswith('total'):
                            if line.startswith('/'):
                                files_after.add(line)
                            else:
                                parts = line.split()
                                if len(parts) >= 9:
                                    files_after.add(parts[-1])
                # Find new files (created during execution)
                new_files = list(files_after - files_before)
                # Filter out common unwanted files
                excluded_patterns = ['.pyc', '__pycache__', '.ipynb_checkpoints', '.cache']
                new_files = [f for f in new_files if not any(p in f for p in excluded_patterns)]
                logger.info(f"Files after execution: {len(files_after)}, new files detected: {new_files}")
            except Exception as e:
                logger.warning(f"Could not list files after execution: {e}")

            # 5. Collect output files (auto-detected + explicitly requested)
            files = []
            collected_filenames = set()

            # First collect explicitly requested files
            for filepath in request.output_files:
                try:
                    content = sandbox.files.read(filepath)
                    # Handle both string and bytes
                    if isinstance(content, str):
                        content = content.encode('utf-8')
                    filename = os.path.basename(filepath)
                    files.append(FileResult(
                        filename=filename,
                        data=base64.b64encode(content).decode('utf-8'),
                        size=len(content)
                    ))
                    collected_filenames.add(filename)
                    logger.info(f"Retrieved requested file: {filepath} ({len(content)} bytes)")
                except Exception as e:
                    logger.warning(f"Could not retrieve file {filepath}: {e}")

            # Then collect auto-detected new files (if not already collected)
            for filepath in new_files:
                filename = os.path.basename(filepath)
                if filename in collected_filenames:
                    continue  # Already collected
                try:
                    # Try both the full path and just the filename
                    content = None
                    for try_path in [filepath, f"/home/user/{filename}", filename]:
                        try:
                            content = sandbox.files.read(try_path)
                            break
                        except:
                            continue

                    if content is not None:
                        if isinstance(content, str):
                            content = content.encode('utf-8')
                        files.append(FileResult(
                            filename=filename,
                            data=base64.b64encode(content).decode('utf-8'),
                            size=len(content)
                        ))
                        collected_filenames.add(filename)
                        logger.info(f"Retrieved auto-detected file: {filename} ({len(content)} bytes)")
                except Exception as e:
                    logger.warning(f"Could not retrieve auto-detected file {filepath}: {e}")

            execution_time = time.time() - start_time

            response = AdvancedExecuteResponse(
                success=error_msg is None,
                stdout=stdout,
                stderr=stderr,
                error=error_msg,
                plots=plots,
                files=files,
                execution_time=execution_time,
                install_output=install_output
            )

            logger.info(f"Advanced execution completed: success={response.success}, plots={len(plots)}, files={len(files)}, time={execution_time:.2f}s")
            return response

    except Exception as e:
        logger.error(f"Advanced sandbox execution failed: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Sandbox execution failed: {str(e)}"
        )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8001,
        log_level="info"
    )
