// Atlas project configuration for slips-core.
//
// We keep the existing migration directory in golang-migrate format
// (e.g. 001_init.up.sql / 001_init.down.sql).
//
// Provide runtime values via CLI flags:
//   atlas ... --env local --var url=... --var dev_url=...

variable "url" {
  type = string
}

variable "dev_url" {
  type = string
}

env "local" {
  url = var.url
  dev = var.dev_url

  migration {
    dir = "file://migrations?format=golang-migrate"
  }
}
