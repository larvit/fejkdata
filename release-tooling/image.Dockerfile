FROM scratch
ARG TARGETPLATFORM
COPY LICENSE DATA-LICENSES.md data/en_US/WORDNET-LICENSE.txt /usr/share/doc/fejkdata/
COPY $TARGETPLATFORM/fejkdata /fejkdata
ENTRYPOINT ["/fejkdata"]
