DROP TABLE IF EXISTS "accounts", "entries", "transfers" CASCADE;

CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "owner" varchar(255) NOT NULL,
  "balance" bigint NOT NULL,
  "currency" varchar(10) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now()
);



CREATE TABLE "entries" (
  "id" bigserial PRIMARY KEY,
  "account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  CONSTRAINT "entries_account_id_fk" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE
);

CREATE TABLE "transfers" (
  "id" bigserial PRIMARY KEY,
  "from_account_id" bigint NOT NULL,
  "to_account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  CONSTRAINT "transfers_from_account_fk" FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE,
  CONSTRAINT "transfers_to_account_fk" FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE
);


CREATE INDEX ON "accounts" ("owner");

CREATE INDEX ON "entries" ("account_id");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");