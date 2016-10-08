FROM python

WORKDIR /var/www

# キャッシュ効率を上げるためにrequirements.txtだけ先にcopyしてインストールする
COPY ./requirements.freeze /var/www

RUN pip install -r requirements.freeze

COPY . /var/www

CMD ["python", "app.py"]
