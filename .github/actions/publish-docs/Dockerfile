FROM python:2.7

LABEL "maintainer"="mark.aaron.phelps@gmail.com"
LABEL "repository"="https://github.com/markphelps/flipt"

RUN pip install mkdocs mkdocs-material pygments

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
