FROM perl:5.24.0

RUN cpanm Carton

COPY ./local /root
RUN carton install

COPY . /root

CMD carton exec -- plackup -s Starlet -E production
