# EXODON

A Go package for syncing Starling transactions to Postgres.

## Usage

Install:
```
go install github.com/thisdotrob/exodon
```

Set the following:
`STARLING_TOKEN`
`EXODON_PG_HOST`
`EXODON_PG_PORT`
`EXODON_PG_USER`
`EXODON_PG_PASSWORD`
`EXODON_PG_DBNAME`

Run it:
```
~/go/bin/exodon
```

## Running as a cron job

Install as above.

Edit the environment variables in the `exodon_cron` file.

Symlink it:
```
sudo ln -s $(pwd)/exodon_cron /usr/local/bin/exodon_cron
```

Add something like this to the crontab (`crontab -e`):
```
# m    h dom mon dow   command
  0,30 * *   *   *     /usr/local/bin/exodon_cron
```
