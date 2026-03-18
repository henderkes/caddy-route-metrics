FROM dunglas/frankenphp:builder AS builder

COPY --from=caddy:builder /usr/bin/xcaddy /usr/bin/xcaddy

COPY caddy/ /src/caddy/

RUN CGO_ENABLED=1 \
    XCADDY_GO_BUILD_FLAGS="-ldflags='-w -s' -tags=nobadger,nomysql,nopgx" \
    CGO_CFLAGS="$(php-config --includes)" \
    CGO_LDFLAGS="$(php-config --ldflags) $(php-config --libs)" \
    xcaddy build \
        --output /usr/local/bin/frankenphp \
        --with github.com/dunglas/frankenphp/caddy \
        --with github.com/dunglas/mercure/caddy \
        --with github.com/dunglas/vulcain/caddy \
        --with github.com/dunglas/caddy-cbrotli \
        --with github.com/henderkes/caddy-route-metrics/caddy=/src/caddy

# ---

FROM dunglas/frankenphp:latest

COPY --from=builder /usr/local/bin/frankenphp /usr/local/bin/frankenphp
COPY Caddyfile.example /etc/frankenphp/Caddyfile

EXPOSE 80 443

ENTRYPOINT ["frankenphp"]
CMD ["run", "--config", "/etc/frankenphp/Caddyfile"]

# ---

FROM dunglas/frankenphp:latest AS test

RUN apt-get update && apt-get install -y --no-install-recommends unzip && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/frankenphp /usr/local/bin/frankenphp
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY test-symfony/ /app/
WORKDIR /app
RUN composer install --no-dev --optimize-autoloader

COPY Caddyfile.example /etc/frankenphp/Caddyfile

EXPOSE 80

ENTRYPOINT ["frankenphp"]
CMD ["run", "--config", "/etc/frankenphp/Caddyfile"]
