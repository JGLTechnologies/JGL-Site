# --IMPORTS--

import asyncio
import os
import aiohttp
from fastapi.staticfiles import StaticFiles
from fastapi import FastAPI, Request, Response, Form
from starlette.exceptions import HTTPException as StarletteHTTPException
from fastapi.templating import Jinja2Templates
from dpys import utils
import aiosqlite
import uvicorn
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_ipaddr
from slowapi.errors import RateLimitExceeded
from fastapi.responses import *
from functools import partial
from aiotools.AIObuiltins import aio_round

# --GLOBAL VARIABLES / INITIALIZERS--

limiter = Limiter(key_func=get_ipaddr)
os.chdir("/var/www/html")
app = FastAPI(docs_url=None, redoc_url=None)
api = FastAPI(redoc_url=None, description="The rate limit is 5 requests per second. This is not per page. The IP info api has a rate limit of 1 request per second. When we upgrade our server we will allow people to make more requests. Also if you reach over 200 requests in 10 seconds your IP will be banned for 1 minute.")
api.state.limiter = limiter
api.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)
templates = Jinja2Templates(directory="web files")

# --MAIN WEBSITE CODE--


@app.get("/")
@app.get("/home")
@limiter.limit("5/second")
async def home(request: Request):
    context = {"request": request, "file": "home.html"}
    return await asyncio.get_event_loop().run_in_executor(None, templates.TemplateResponse, "base.html", context)


@app.get("/contact")
@limiter.limit("5/second")
async def contact(request: Request):
    context = {"request": request, "file": "contact.html"}
    return await asyncio.get_event_loop().run_in_executor(None, templates.TemplateResponse, "base.html", context)


@app.get("/freelance")
@limiter.limit("5/second")
async def freelance(request: Request, response: Response):
    context = {"request": request, "file": "freelance.html"}
    return await asyncio.get_event_loop().run_in_executor(None, templates.TemplateResponse, "base.html", context)


@app.get("/discord")
@limiter.limit("5/second")
async def discord(request: Request):
    return RedirectResponse("https://discord.gg/TUUbzTa3B7")


@app.get("/favicon.ico")
@limiter.limit("5/second")
async def ico(request: Request):
    return RedirectResponse(
        "https://raw.githubusercontent.com/Nebulizer1213/JGL-Plugins/main/favicon.ico")


@app.get("/dpys/donate")
@limiter.limit("5/second")
async def dpys_donate(request: Request):
    return RedirectResponse(
        "https://www.paypal.com/donate?business=4RE48WGW7R5YS&no_recurring=0&item_name=DPYS+is+a+python+library+with+a+goal+to+make+bot+development+easy+for+beginners.+We+would+appreciate+if+you+could+donate.+&currency_code=USD")


@app.get("/bot/donate")
@limiter.limit("5/second")
async def bot_donate(request: Request):
    return RedirectResponse(
        "https://www.paypal.com/donate/?business=4RE48WGW7R5YS&no_recurring=0&item_name=The+JGL+Bot+is+a+free+Discord+bot.+We+need+money+to+keep+it+running.+We+would+appreciate+if+you+donated+to+the+bot.&currency_code=USD")


@app.get("/bot")
@limiter.limit("5/second")
async def bot(request: Request):
    return HTMLResponse(
        "JGL Bot documentation is coming soon!<br><a href='/bot/donate'>Donation link</a>")


@app.get("/dpys")
@limiter.limit("5/second")
async def dpys(request: Request):
    return RedirectResponse("https://sites.google.com/view/dpys")


@app.get("/dpys/src")
@limiter.limit("5/second")
async def dpys_src(request: Request):
    return RedirectResponse("https://github.com/Nebulizer1213/dpys")


@app.get("/dpys/pypi")
@limiter.limit("5/second")
async def dpys_src(request: Request):
    return RedirectResponse("https://pypi.org/project/dpys")


@app.get("/src")
@limiter.limit("5/second")
async def src(request: Request):
    return RedirectResponse("https://github.com/Nebulizer1213/jgl-site")


class Test:

    @app.get("/test/bmi")
    @limiter.limit("5/second")
    async def bmi_main(request: Request):
        context = {"request": request, "file": "test/bmi/index.html"}
        return await asyncio.get_event_loop().run_in_executor(None, templates.TemplateResponse, "test/bmi/styles.html", context)

    @app.get("/test/bmi/calc")
    @limiter.limit("5/second")
    async def bmi_calc(weight, heightft, heightin, request: Request, response: Response):
        if await utils.var_can_be_type(weight, float) and await utils.var_can_be_type(heightft, float):
            if heightin == "":
                heightin = 0
            else:
                if await utils.var_can_be_type(heightin, float):
                    heightin = float(heightin)
                else:
                    return await asyncio.get_event_loop().run_in_executor(None, partial(templates.TemplateResponse, "test/bmi/invalid.html", {"request": request}, status_code=400))
            # bmi = 703*(weight(lbs)/height(in)**2)
            bmi = float(weight) / \
                (((float(heightft) * 12) + heightin)**2) * 703
            if bmi > 24.9:
                new_weight = (
                    24.9 / 703 * ((float(heightft) * 12) + heightin)**2)
                weight = float(weight) - new_weight
                if weight >= 1:
                    context = {
                        "request": request,
                        "bmi": await aio_round(bmi, 2),
                        "weight": f"You need to loose {await aio_round(weight, 2)} pounds to be healthy."}
                else:
                    context = {
                        "request": request, "bmi": await aio_round(bmi, 2), "weight": ""}
            elif bmi < 18.5:
                new_weight = (
                    18.5 / 703 * ((float(heightft) * 12) + heightin)**2)
                weight = new_weight - float(weight)
                if weight >= 1:
                    context = {
                        "request": request,
                        "bmi": await aio_round(
                            bmi,
                            2),
                        "weight": f"You need to gain {await aio_round(weight, 2)} pounds to be healthy."}
                else:
                    context = {
                        "request": request, "bmi": await aio_round(bmi, 2), "weight": ""}
            else:
                context = {
                    "request": request, "bmi": await aio_round(bmi, 2), "weight": ""}

        else:
            return await asyncio.get_event_loop().run_in_executor(None, partial(templates.TemplateResponse,
                                                                  "test/bmi/invalid.html", {"request": request}, status_code=400))
        return await asyncio.get_event_loop().run_in_executor(None, templates.TemplateResponse, "test/bmi/bmi.html", context)


class Api:

    class Main:

        @api.post("/contact", include_in_schema=False)
        @limiter.limit("5/second")
        async def contact_api(response: Response, request: Request, name: str = Form(None), email: str = Form(None), message: str = Form(None), token: str = Form(None)):
            async with aiohttp.ClientSession() as session:
                async with session.post("http://jglbotapi.us/contact", json={"ip": request.headers.get("X-Forwarded-For").split(",")[0], "name": name, "email": email, "message": message, "token": token}, headers={"IP-OG": request.headers.get("X-Forwarded-For").split(",")[0]}) as response:
                    return HTMLResponse(await response.read(), status_code=response.status)

        @api.post("/freelance", include_in_schema=False)
        @limiter.limit("5/second")
        async def freelance_api(response: Response, request: Request, name: str = Form(None), email: str = Form(None), message: str = Form(None), token: str = Form(None)):
            async with aiohttp.ClientSession() as session:
                async with session.post("http://jglbotapi.us/freelance", json={"ip": request.headers.get("X-Forwarded-For").split(",")[0], "name": name, "email": email, "message": message, "token": token}, headers={"IP-OG": request.headers.get("X-Forwarded-For").split(",")[0]}) as response:
                    return HTMLResponse(await response.read(), status_code=response.status)

        @api.get("/ip/{ip}", description="Gets info about an ip address")
        @limiter.limit("5/second")
        async def ip_info(request: Request, ip: str):
            async with aiohttp.ClientSession() as session:
                try:
                    async with session.get(f"https://tools.keycdn.com/geo.json?host={ip}", headers={f"User-Agent": "keycdn-tools:http://{ip}"}) as res:
                        data = await res.json()
                        for x in data.get("data").get("geo"):
                            if data.get("data").get("geo").get(x) is None or data.get(
                                    "data").get("geo").get(x) == "":
                                data["data"]["geo"][x] = "Not Found"
                        if res.status == 429:
                            return JSONResponse(data, status_code=429)
                        return JSONResponse(
                            data.get("data").get("geo"), indent=4)
                except:
                    return PlainTextResponse(
                        "Domain/IP not found!", status_code=404)

    class Bot:

        @api.get("/bot/status",
                 description="Checks if the JGL Bot is online or offline")
        @limiter.limit("5/second")
        async def jgl_bot_status(request: Request):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get("http://jglbotapi.us/bot_is_online", timeout=.1) as bot_response:
                        data = await bot_response.json()
                        if data["online"]:
                            response = {"online": True}
            except:
                response = {"online": False}
            return response

        @api.get("/bot/info", description="Gets info for the JGL Bot")
        @limiter.limit("5/second")
        async def get_info_for_jgl_bot(request: Request):
            async with aiohttp.ClientSession() as session:
                try:
                    async with session.get("http://jglbotapi.us/info", timeout=.1) as response:
                        data = await response.json()
                        guilds = data["guilds"]
                        cogs = data["cogs"]
                        shards = data["shards"]
                        size_gb = data["size"]["gb"]
                        size_mb = data["size"]["mb"]
                        size_kb = data["size"]["kb"]
                        ping = data["ping"]
                except:
                    guilds = "Not Found"
                    cogs = "Not Found"
                    shards = "Not Found"
                    size_gb = "Not Found"
                    size_mb = "Not Found"
                    size_kb = "Not Found"
                    ping = "Not Found"
                dict = {
                    "guilds": guilds,
                    "shards": shards,
                    "cogs": cogs,
                    "ping": ping,
                    "size": {
                        "gb": size_gb,
                        "mb": size_mb,
                        "kb": size_kb}}
                return JSONResponse(dict, indent=4)

        @api.get("/dpys", description="Gets info for DPYS")
        @limiter.limit("5/second")
        async def dpys_info(request: Request):
            async with aiohttp.ClientSession() as session:
                async with session.get("https://pypi.org/pypi/dpys/json") as response:
                    data = await response.json()
                    version = data["info"]["version"]
                async with session.get(f"https://raw.githubusercontent.com/Nebulizer1213/DPYS/main/dist/dpys-{version}.tar.gz") as response:
                    file_bytes = str(await response.read())
            response_data = {"version": version, "file_bytes": file_bytes}
            return JSONResponse(response_data, indent=4)

    # The forum api is not finished.
    class Forum:
        @api.get("/forum/login", include_in_schema=False)
        @limiter.limit("5/second")
        async def login(request: Request):
            try:
                username = request.headers["username"]
                passw = request.headers["password"]
            except:
                return "login configured wrong"
            async with aiosqlite.connect("users.db") as db:
                async with db.execute("""
                SELECT username, password
                FROM accounts;
                """) as cur:
                    async for entry in cur:
                        user, password = entry
                        if username == user and passw == password:
                            return {"success": True}
                    return {"success": False}

        @api.post("/forum/createacc", include_in_schema=False)
        @limiter.limit("5/second")
        async def createacc(request: Request):
            async with aiosqlite.connect("users.db") as db:
                try:
                    await db.execute("INSERT INTO accounts (username,password) VALUES (?,?)", (request.headers["username"], request.headers["password"]))
                    await db.commit()
                    return "account created"
                except:
                    return "account already exists"

        @api.post("/forum/sendmsg", include_in_schema=False)
        @limiter.limit("5/second")
        async def sendmsg(request: Request):
            try:
                body = request.headers["body"]
                user = request.headers["username"]
                password = request.headers["password"]
            except:
                return "login configured wrong"
            async with aiosqlite.connect("users.db") as db:
                try:
                    async with db.execute(f"""
                    SELECT password
                    FROM accounts
                    WHERE username='{user}';
                    """) as cur:
                        async for x in cur:
                            if x[0] == password:
                                return f"Message Sent with: {body}"
                            else:
                                return "password is incorrect"
                except:
                    return "username does not exist"

        async def setup():
            async with aiosqlite.connect("users.db") as db:
                await db.execute("""CREATE TABLE IF NOT EXISTS messages(
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    body TEXT,
                    user TEXT
                    )""")

                await db.execute("""CREATE TABLE IF NOT EXISTS accounts(
                    username TEXT PRIMARY KEY,
                    password TEXT
                    )""")
                await db.commit()


@app.exception_handler(StarletteHTTPException)
async def invalid_path(request, exc):
    if exc.status_code == 404:
        return RedirectResponse("/")


@api.exception_handler(StarletteHTTPException)
async def invalid_path(request, exc):
    if exc.status_code == 404:
        return RedirectResponse("/api/docs")


@app.on_event("startup")
async def startup():
    await Api.Forum.setup()


def startup():
    app.mount("/api", api)
    app.mount("/static", StaticFiles(directory="static"), name="static")
    # uvicorn.run(
    #     "main:app",
    #     port=81,
    #     host="0.0.0.0",
    #     reload=True,
    #     workers=4)
    os.system(
        "gunicorn main:app --workers=9 -k uvicorn.workers.UvicornWorker --reload -b 0.0.0.0:81")


if __name__ == "__main__":
    startup()
