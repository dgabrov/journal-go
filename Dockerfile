FROM amd64/alpine
EXPOSE 3001
WORKDIR /app
COPY journal_backend .
CMD ["/app/journal_backend"]
