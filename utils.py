from fastapi import Request, Response
from fastapi.responses import PlainTextResponse
from slowapi.errors import RateLimitExceeded


def handler(request: Request, exc: RateLimitExceeded) -> PlainTextResponse:
    response = PlainTextResponse(
        "Too many requests", status_code=429
    )
    response = request.app.state.limiter._inject_headers(
        response, request.state.view_rate_limit
    )
    return response


def var_can_be_type(var, type) -> bool:
    try:
        type(var)
    except ValueError:
        return False
    return True
