FROM ruby:2.3.1

RUN gem update -N --system && gem install -N bundler

RUN mkdir -p /app
WORKDIR /app

COPY Gemfile Gemfile.lock /app/
RUN bundle install --jobs 4

CMD bundle exec thin start --port 80 -A rack -R config.ru
