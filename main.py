# Copyright (c) JGL WEBSITE 2021, JGL Technolgies

# --IMPORTS--

import asyncio
import os
import re
import aiohttp_jinja2
import jinja2
import aiohttp
from aiohttp import web
from dpys import utils

# --GLOBAL VARIABLES / INITIALIZERS--

# os.chdir("/var/www/html")
routes = web.RouteTableDef()
# SECRET_KEY = os.environ.get("SECRET_KEY")

# --MAIN WEBSITE CODE--

@routes.get("/")
@routes.get("/home")
@routes.get("/home/")
@aiohttp_jinja2.template("base.html", status=200)
async def home(request):
    context = {"file":"home.html"}
    return context

@routes.get("/contact")
@routes.get("/contact/")
@aiohttp_jinja2.template("base.html", status=200)
async def contact(request):
    context = {"file":"contact.html"}
    return context

@routes.get("/freelance")
@routes.get("/freelance/")
@aiohttp_jinja2.template("base.html", status=200)
async def freelance(request):
    context = {"file":"freelance.html"}
    return context

@routes.get("/discord")
@routes.get("/discord/")
async def discord(request):
    raise web.HTTPFound("https://discord.gg/TUUbzTa3B7")

@routes.get("/favicon.ico")
async def ico(request):
    return web.Response(text="https://raw.githubusercontent.com/Nebulizer1213/4535435456543/main/favicon.ico")

@routes.get("/dpys/donate")
@routes.get("/dpys/donate/")
async def dpys_donate(request):
    raise web.HTTPFound("https://www.paypal.com/donate?business=4RE48WGW7R5YS&no_recurring=0&item_name=DPYS+is+a+python+library+with+a+goal+to+make+bot+development+easy+for+beginners.+We+would+appreciate+if+you+could+donate.+&currency_code=USD")

@routes.get("/bot/donate")
@routes.get("/bot/donate/")
async def bot_donate(request):
    raise web.HTTPFound("https://www.paypal.com/donate/?business=4RE48WGW7R5YS&no_recurring=0&item_name=The+JGL+Bot+is+a+free+Discord+bot.+We+need+money+to+keep+it+running.+We+would+appreciate+if+you+donated+to+the+bot.&currency_code=USD")

@routes.get("/bot")
@routes.get("/bot/")
async def bot(request):
    return web.Response(text="JGL Bot documentation is coming soon!<br><a href='/bot/donate'>Donation link</a>", status=200, content_type="text/html")

@routes.get("/dpys")
@routes.get("/dpys/")
async def dpys(request):
    raise web.HTTPFound("https://sites.google.com/view/dpys")

class Test:

    @routes.get("/test/bmi")
    @routes.get("/test/bmi/")
    @aiohttp_jinja2.template("test/bmi/styles.html", status=200)
    async def bmi_main(request):
        context = {"file":"test/bmi/index.html"}
        return context

    @routes.get("/test/bmi/calc")
    @routes.get("/test/bmi/calc/")
    @aiohttp_jinja2.template("test/bmi/bmi.html", status=200)
    async def bmi_calc(request):
        # try:
        #     token = request.query["token"]
        #     async with aiohttp.ClientSession() as session:
        #         async with session.get(f"https://www.google.com/recaptcha/api/siteverify?secret={SECRET_KEY}&remoteip={request.remote}&response={token}") as response:
        #             data = await response.json()
        #             if data['score'] <= 0.6:
        #                 return web.Response(text="<script>alert('Your reCPATCHA token has been scored below 0.7 which means you are probably a bot. If you are not, report the issue to us in our discord server. http://jgltechnologies.com/discord'); document.location = '/';</script>", content_type="text/html", status=403)
        # except Exception as e:
        #     return web.Response(text=f"<script>alert('Something went wrong. Error code: {e}. Report the error in our discord server http://jgltechnologies.com/discord'); document.location = '/';</script>", content_type="text/html", status=500)
        if request.query.get('weight') is None or request.query.get('heightin') is None or request.query.get('heightft') is None:
            return web.Response(status=400, text="Invalid params")
        if await utils.var_can_be_type(request.query.get('weight'), float) and await utils.var_can_be_type(request.query.get('heightft'), float):
            if request.query['heightin'] == "":
                heightin = 0
            else:
                if await utils.var_can_be_type(request.query.get('heightin'), float):
                    heightin = float(request.query['heightin'])
                else:
                    return aiohttp_jinja2.render_template("test/bmi/invalid.html", request, context={}, status=400)
            # bmi = 703*(weight(lbs)/height(in)**2)
            bmi = float(request.query['weight'])/(((float(request.query['heightft'])*12) + heightin)**2)*703
            if bmi > 24.9:
                weight = (24.9/703*((float(request.query['heightft'])*12) + heightin)**2)
                weight = float(request.query['weight']) - weight
                if weight >= 1:
                    context = {"bmi":round(bmi, 2), "weight":f"You need to loose {round(weight, 2)} pounds to be healthy."}
                else:
                    context = {"bmi":round(bmi, 2), "weight":""}
            elif bmi < 18.5:
                weight = (18.5/703*((float(request.query['heightft'])*12) + heightin)**2)
                weight = weight - float(request.query['weight'])
                if weight >= 1:
                    context = {"bmi":round(bmi, 2), "weight":f"You need to gain {round(weight, 2)} pounds to be healthy."}
                else:
                    context = {"bmi":round(bmi, 2), "weight":""}
            else:
                context = {"bmi":round(bmi, 2), "weight":""}

        else:
            return aiohttp_jinja2.render_template("test/bmi/invalid.html", request, context={}, status=400)
        return context

PATHS = set({})
for route in routes:
    path = str(route)[14:]
    path = re.sub(" ->.*", "", path)
    PATHS.add(path)

@routes.get("/{tail:.*}")
async def invalid_path(request):
    if request.path not in PATHS:
        raise web.HTTPFound("/")

def startup():
    app = web.Application()
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    aiohttp_jinja2.setup(app, loader=jinja2.FileSystemLoader('web files')) 
    app.add_routes(routes)
    web.run_app(app, port=80)

startup()
