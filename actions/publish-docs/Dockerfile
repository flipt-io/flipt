FROM python:2.7

LABEL "maintainer"="mark.aaron.phelps@gmail.com"
LABEL "repository"="https://github.com/markphelps/flipt"

LABEL "com.github.actions.name"="Publish Docs"
LABEL "com.github.actions.description"="Publish mkdocs to gh pages"
LABEL "com.github.actions.icon"="file-text"
LABEL "com.github.actions.color"="purple"

RUN pip install mkdocs mkdocs-material pygments

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
