FROM perl:5.24.0

RUN cpanm Carton

COPY cpanfile cpanfile.snapshot ./local /root/
RUN carton install --deployment

COPY . /root

CMD carton exec -- plackup -s Starlet -E production
