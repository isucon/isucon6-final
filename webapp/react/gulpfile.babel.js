import gulp from 'gulp';
import babel from 'gulp-babel';
import webpack from 'webpack-stream';
import webpackConfig from './webpack.config.js';
import {spawn} from 'child_process';
import del from 'del';

let server;
function restartServer() {
  if (typeof server !== 'undefined') {
    server.kill('SIGTERM');
  }
  console.log('starting server');
  server = spawn('node', ['build/server.js'], {
    env: process.env,
    silent: false,
  });
  server.stdout.on('data', (buf) => process.stdout.write(buf));
  server.stderr.on('data', (buf) => process.stderr.write(buf));
}

gulp.task('default', ['build']);

gulp.task('build', ['server', 'browser']);

gulp.task('clean', () => {
  return del(['build']);
});

gulp.task('watch', ['server', 'browser'], () => {
  const watcher = gulp.watch('**/*.jsx', ['browser', 'server']);
  watcher.on('change', restartServer);
  watcher.on('start', restartServer);
  restartServer();
});

gulp.task('start', ['server', 'browser'], () => {
  return new Promise((resolve, reject) => {
    restartServer();
    server.on('exit', resolve);
  });
});

gulp.task('server', () => {
  return gulp.src(['**/*.jsx'])
    .pipe(babel())
    .on('error', swallowError)
    .pipe(gulp.dest('build'));
});

gulp.task('browser', () => {
  return gulp.src(['browser.js'])
    .pipe(webpack(webpackConfig))
    .on('error', swallowError)
    .pipe(gulp.dest('build'));
});

process.on('exit', () => {
  if (server) {
    server.kill('SIGTERM');
  }
});

// http://stackoverflow.com/questions/23971388/prevent-errors-from-breaking-crashing-gulp-watch
function swallowError (error) {
  // If you want details of the error in the console
  console.log(error.toString())

  this.emit('end')
}

