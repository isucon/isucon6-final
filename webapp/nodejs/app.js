const Koa = require('koa');
const app = new Koa();

const convert = require('koa-convert');
const logger = require('koa-logger');
const json = require('koa-json');

app.use(convert(json()));
app.use(convert(logger()));

// logger
app.use(async (ctx, next) => {
  const start = new Date();
  await next();
  const ms = new Date() - start;
  console.log(`[app] ${ctx.method} ${ctx.url} - ${ms}ms`);
});

app.on('error', function(err, ctx){
  console.log(err)
  logger.error('server error', err, ctx);
});

const router = require('koa-router')();
router.post('api/csrf_token', async () => {
});

router.get('api/rooms', async () => {
});

router.post('api/rooms', async () => {
});

router.get('api/rooms/:id', async () => {
});

router.post('api/strokes/rooms/:id', async () => {
});

router.get('api/initialize', async () => {
});

router.use('/', router.routes(), router.allowedMethods());
app.use(router.routes(), router.allowedMethods);

module.exports = app;
